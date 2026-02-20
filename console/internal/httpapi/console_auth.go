package httpapi

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/onlyboxes/onlyboxes/console/internal/persistence/sqlc"
	"golang.org/x/crypto/bcrypt"
)

const (
	dashboardSessionCookieName      = "onlyboxes_console_session"
	dashboardSessionMaxAgeSec       = 12 * 60 * 60
	dashboardUsernamePrefix         = "admin-"
	dashboardUsernameRandomByteSize = 4
	dashboardPasswordRandomByteSize = 24
	dashboardPasswordHashAlgo       = "bcrypt"
	dashboardPasswordBCryptCost     = 12
	accountIDPrefix                 = "acc_"
	accountIDRandomByteSize         = 16
	maxAccountUsernameRunes         = 64

	requestAccountIDGinKey       = "request_account_id"
	requestAccountUsernameGinKey = "request_account_username"
	requestAccountIsAdminGinKey  = "request_account_is_admin"
)

var (
	dashboardSessionTTL               = 12 * time.Hour
	errAccountUsernameRequired        = errors.New("username is required")
	errAccountUsernameTooLong         = errors.New("username length must be <= 64")
	errAccountPasswordRequired        = errors.New("password is required")
	errAccountRegistrationDisabled    = errors.New("registration is disabled")
	errAccountRegistrationConflict    = errors.New("username already exists")
	errAccountInvalidCredentialRecord = errors.New("invalid account credential record")
	accountIDGenerator                = generateAccountID
)

type DashboardCredentials struct {
	Username string
	Password string
}

type AdminAccountInitResult struct {
	AccountID         string
	Username          string
	PasswordPlaintext string
	InitializedNow    bool
	EnvIgnored        bool
}

type SessionAccount struct {
	AccountID string `json:"account_id"`
	Username  string `json:"username"`
	IsAdmin   bool   `json:"is_admin"`
}

type accountSessionState struct {
	Account   SessionAccount
	ExpiresAt time.Time
}

type accountContextKey struct{}

type consoleAccountContext struct {
	AccountID string
	Username  string
	IsAdmin   bool
}

type ConsoleAuth struct {
	queries             *sqlc.Queries
	registrationEnabled bool

	sessionMu sync.Mutex
	sessions  map[string]accountSessionState
	nowFn     func() time.Time
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type accountSessionResponse struct {
	Authenticated       bool           `json:"authenticated,omitempty"`
	Account             SessionAccount `json:"account"`
	RegistrationEnabled bool           `json:"registration_enabled"`
}

type registerAccountResponse struct {
	Account   SessionAccount `json:"account"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type createAccountInput struct {
	Username string
	Password string
	IsAdmin  bool
	Now      time.Time
}

type createdAccount struct {
	AccountID string
	Username  string
	IsAdmin   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type accountInsertConflict int

const (
	accountInsertConflictUnknown accountInsertConflict = iota
	accountInsertConflictAccountID
	accountInsertConflictUsernameKey
)

func ResolveDashboardCredentials(username string, password string) (DashboardCredentials, error) {
	credentials := DashboardCredentials{
		Username: strings.TrimSpace(username),
		Password: password,
	}

	if credentials.Username == "" {
		suffix, err := randomHex(dashboardUsernameRandomByteSize)
		if err != nil {
			return DashboardCredentials{}, err
		}
		credentials.Username = dashboardUsernamePrefix + suffix
	}
	if credentials.Password == "" {
		secret, err := randomHex(dashboardPasswordRandomByteSize)
		if err != nil {
			return DashboardCredentials{}, err
		}
		credentials.Password = secret
	}

	return credentials, nil
}

func InitializeAdminAccount(
	ctx context.Context,
	queries *sqlc.Queries,
	envUsername string,
	envPassword string,
) (AdminAccountInitResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if queries == nil {
		return AdminAccountInitResult{}, errors.New("account queries is required")
	}

	adminCount, err := queries.CountAdminAccounts(ctx)
	if err != nil {
		return AdminAccountInitResult{}, fmt.Errorf("count admin accounts: %w", err)
	}
	if adminCount > 0 {
		adminAccount, err := queries.GetFirstAdminAccount(ctx)
		if err != nil {
			return AdminAccountInitResult{}, fmt.Errorf("load admin account: %w", err)
		}
		return AdminAccountInitResult{
			AccountID:      strings.TrimSpace(adminAccount.AccountID),
			Username:       strings.TrimSpace(adminAccount.Username),
			InitializedNow: false,
			EnvIgnored:     envUsername != "" || envPassword != "",
		}, nil
	}

	credentials, err := ResolveDashboardCredentials(envUsername, envPassword)
	if err != nil {
		return AdminAccountInitResult{}, err
	}

	created, err := createAccountWithRetry(ctx, queries, createAccountInput{
		Username: credentials.Username,
		Password: credentials.Password,
		IsAdmin:  true,
		Now:      time.Now(),
	})
	if err != nil {
		return AdminAccountInitResult{}, fmt.Errorf("insert admin account: %w", err)
	}

	return AdminAccountInitResult{
		AccountID:         created.AccountID,
		Username:          created.Username,
		PasswordPlaintext: credentials.Password,
		InitializedNow:    true,
		EnvIgnored:        false,
	}, nil
}

func NewConsoleAuth(queries *sqlc.Queries, registrationEnabled bool) *ConsoleAuth {
	if queries == nil {
		panic("console auth requires non-nil queries")
	}
	return &ConsoleAuth{
		queries:             queries,
		registrationEnabled: registrationEnabled,
		sessions:            make(map[string]accountSessionState),
		nowFn:               time.Now,
	}
}

func (a *ConsoleAuth) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	account, ok := a.lookupAccount(c.Request.Context(), req.Username)
	if !ok || !a.verifyPassword(account, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	sessionAccount := SessionAccount{
		AccountID: strings.TrimSpace(account.AccountID),
		Username:  strings.TrimSpace(account.Username),
		IsAdmin:   account.IsAdmin == 1,
	}
	sessionID, expiresAt, err := a.createSession(sessionAccount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	a.setSessionCookie(c, sessionID, expiresAt)
	c.JSON(http.StatusOK, accountSessionResponse{
		Authenticated:       true,
		Account:             sessionAccount,
		RegistrationEnabled: a.registrationEnabled,
	})
}

func (a *ConsoleAuth) Session(c *gin.Context) {
	account, ok := requireSessionAccount(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, accountSessionResponse{
		Authenticated:       true,
		Account:             account,
		RegistrationEnabled: a.registrationEnabled,
	})
}

func (a *ConsoleAuth) Register(c *gin.Context) {
	if !a.registrationEnabled {
		c.JSON(http.StatusForbidden, gin.H{"error": errAccountRegistrationDisabled.Error()})
		return
	}

	account, ok := requireSessionAccount(c)
	if !ok {
		return
	}
	if !account.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var req registerAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	now := time.Now()
	if a.nowFn != nil {
		now = a.nowFn()
	}

	created, err := createAccountWithRetry(c.Request.Context(), a.queries, createAccountInput{
		Username: req.Username,
		Password: req.Password,
		IsAdmin:  false,
		Now:      now,
	})
	if err != nil {
		switch {
		case errors.Is(err, errAccountUsernameRequired),
			errors.Is(err, errAccountUsernameTooLong),
			errors.Is(err, errAccountPasswordRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, errAccountRegistrationConflict):
			c.JSON(http.StatusConflict, gin.H{"error": errAccountRegistrationConflict.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		}
		return
	}

	c.JSON(http.StatusCreated, registerAccountResponse{
		Account: SessionAccount{
			AccountID: created.AccountID,
			Username:  created.Username,
			IsAdmin:   created.IsAdmin,
		},
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	})
}

func (a *ConsoleAuth) Logout(c *gin.Context) {
	if sessionID, err := c.Cookie(dashboardSessionCookieName); err == nil {
		a.deleteSession(sessionID)
	}

	a.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

func (a *ConsoleAuth) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(dashboardSessionCookieName)
		if err != nil || strings.TrimSpace(sessionID) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		sessionState, ok := a.sessionState(sessionID)
		if !ok {
			a.clearSessionCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		setRequestSessionAccount(c, sessionState.Account)
		c.Next()
	}
}

func (a *ConsoleAuth) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		account, ok := requestSessionAccountFromGin(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		if !account.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (a *ConsoleAuth) lookupAccount(ctx context.Context, username string) (sqlc.Account, bool) {
	if a == nil || a.queries == nil {
		return sqlc.Account{}, false
	}
	if ctx == nil {
		ctx = context.Background()
	}
	_, usernameKey, err := normalizeAccountUsername(username)
	if err != nil {
		return sqlc.Account{}, false
	}
	account, err := a.queries.GetAccountByUsernameKey(ctx, usernameKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.Account{}, false
		}
		return sqlc.Account{}, false
	}
	if strings.TrimSpace(account.AccountID) == "" || strings.TrimSpace(account.PasswordHash) == "" {
		return sqlc.Account{}, false
	}
	return account, true
}

func (a *ConsoleAuth) verifyPassword(account sqlc.Account, password string) bool {
	if strings.TrimSpace(password) == "" {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(account.HashAlgo), dashboardPasswordHashAlgo) {
		return false
	}
	return compareDashboardPassword(account.PasswordHash, password)
}

func (a *ConsoleAuth) createSession(account SessionAccount) (string, time.Time, error) {
	sessionID, err := randomHex(32)
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now()
	if a != nil && a.nowFn != nil {
		now = a.nowFn()
	}
	expiresAt := now.Add(dashboardSessionTTL)

	a.sessionMu.Lock()
	a.sessions[sessionID] = accountSessionState{
		Account:   account,
		ExpiresAt: expiresAt,
	}
	a.sessionMu.Unlock()

	return sessionID, expiresAt, nil
}

func (a *ConsoleAuth) sessionState(sessionID string) (accountSessionState, bool) {
	now := time.Now()
	if a != nil && a.nowFn != nil {
		now = a.nowFn()
	}

	a.sessionMu.Lock()
	defer a.sessionMu.Unlock()

	state, ok := a.sessions[sessionID]
	if !ok {
		return accountSessionState{}, false
	}
	if !state.ExpiresAt.After(now) {
		delete(a.sessions, sessionID)
		return accountSessionState{}, false
	}
	return state, true
}

func (a *ConsoleAuth) deleteSession(sessionID string) {
	a.sessionMu.Lock()
	delete(a.sessions, sessionID)
	a.sessionMu.Unlock()
}

func (a *ConsoleAuth) setSessionCookie(c *gin.Context, sessionID string, expiresAt time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     dashboardSessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   dashboardSessionMaxAgeSec,
		Expires:  expiresAt,
		Secure:   requestIsTLS(c.Request),
	})
}

func (a *ConsoleAuth) clearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     dashboardSessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   requestIsTLS(c.Request),
	})
}

func setRequestSessionAccount(c *gin.Context, account SessionAccount) {
	if c == nil {
		return
	}
	accountID := strings.TrimSpace(account.AccountID)
	username := strings.TrimSpace(account.Username)
	if accountID == "" || username == "" {
		return
	}

	c.Set(requestAccountIDGinKey, accountID)
	c.Set(requestAccountUsernameGinKey, username)
	c.Set(requestAccountIsAdminGinKey, account.IsAdmin)

	if c.Request != nil {
		ctxValue := consoleAccountContext{
			AccountID: accountID,
			Username:  username,
			IsAdmin:   account.IsAdmin,
		}
		ctx := context.WithValue(c.Request.Context(), accountContextKey{}, ctxValue)
		c.Request = c.Request.WithContext(ctx)
	}
}

func requestSessionAccountFromContext(ctx context.Context) (SessionAccount, bool) {
	if ctx == nil {
		return SessionAccount{}, false
	}
	value := ctx.Value(accountContextKey{})
	ctxValue, ok := value.(consoleAccountContext)
	if !ok {
		return SessionAccount{}, false
	}
	accountID := strings.TrimSpace(ctxValue.AccountID)
	username := strings.TrimSpace(ctxValue.Username)
	if accountID == "" || username == "" {
		return SessionAccount{}, false
	}
	return SessionAccount{
		AccountID: accountID,
		Username:  username,
		IsAdmin:   ctxValue.IsAdmin,
	}, true
}

func requestSessionAccountFromGin(c *gin.Context) (SessionAccount, bool) {
	if c == nil {
		return SessionAccount{}, false
	}

	accountIDValue, ok := c.Get(requestAccountIDGinKey)
	if ok {
		usernameValue, usernameOK := c.Get(requestAccountUsernameGinKey)
		isAdminValue, isAdminOK := c.Get(requestAccountIsAdminGinKey)
		accountID, accountIDOK := accountIDValue.(string)
		username, usernameTypeOK := usernameValue.(string)
		isAdmin, isAdminTypeOK := isAdminValue.(bool)
		if usernameOK && isAdminOK && accountIDOK && usernameTypeOK && isAdminTypeOK {
			account := SessionAccount{
				AccountID: strings.TrimSpace(accountID),
				Username:  strings.TrimSpace(username),
				IsAdmin:   isAdmin,
			}
			if account.AccountID != "" && account.Username != "" {
				return account, true
			}
		}
	}

	if c.Request != nil {
		return requestSessionAccountFromContext(c.Request.Context())
	}
	return SessionAccount{}, false
}

func requireSessionAccount(c *gin.Context) (SessionAccount, bool) {
	account, ok := requestSessionAccountFromGin(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		c.Abort()
		return SessionAccount{}, false
	}
	return account, true
}

func normalizeAccountUsername(value string) (string, string, error) {
	username := strings.TrimSpace(value)
	if username == "" {
		return "", "", errAccountUsernameRequired
	}
	if utf8.RuneCountInString(username) > maxAccountUsernameRunes {
		return "", "", errAccountUsernameTooLong
	}
	return username, strings.ToLower(username), nil
}

func createAccountWithRetry(
	ctx context.Context,
	queries *sqlc.Queries,
	input createAccountInput,
) (createdAccount, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if queries == nil {
		return createdAccount{}, errors.New("account queries is required")
	}

	normalizedUsername, usernameKey, err := normalizeAccountUsername(input.Username)
	if err != nil {
		return createdAccount{}, err
	}
	if strings.TrimSpace(input.Password) == "" {
		return createdAccount{}, errAccountPasswordRequired
	}

	passwordHash, err := hashDashboardPassword(input.Password)
	if err != nil {
		return createdAccount{}, fmt.Errorf("hash account password: %w", err)
	}

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	nowMS := now.UnixMilli()

	for i := 0; i < 8; i++ {
		accountID, err := accountIDGenerator()
		if err != nil {
			return createdAccount{}, fmt.Errorf("generate account id: %w", err)
		}
		err = queries.InsertAccount(ctx, sqlc.InsertAccountParams{
			AccountID:       accountID,
			Username:        normalizedUsername,
			UsernameKey:     usernameKey,
			PasswordHash:    passwordHash,
			HashAlgo:        dashboardPasswordHashAlgo,
			IsAdmin:         boolToInt64(input.IsAdmin),
			CreatedAtUnixMs: nowMS,
			UpdatedAtUnixMs: nowMS,
		})
		if err == nil {
			createdAt := time.UnixMilli(nowMS)
			return createdAccount{
				AccountID: accountID,
				Username:  normalizedUsername,
				IsAdmin:   input.IsAdmin,
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			}, nil
		}

		if isSQLiteConstraintError(err) {
			conflict, classifyErr := classifyAccountInsertConflict(ctx, queries, accountID, usernameKey)
			if classifyErr != nil {
				return createdAccount{}, classifyErr
			}
			switch conflict {
			case accountInsertConflictAccountID:
				continue
			case accountInsertConflictUsernameKey:
				return createdAccount{}, errAccountRegistrationConflict
			default:
				return createdAccount{}, err
			}
		}
		return createdAccount{}, err
	}

	return createdAccount{}, errors.New("failed to create account")
}

func classifyAccountInsertConflict(
	ctx context.Context,
	queries *sqlc.Queries,
	accountID string,
	usernameKey string,
) (accountInsertConflict, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if queries == nil {
		return accountInsertConflictUnknown, nil
	}

	_, err := queries.GetAccountByID(ctx, strings.TrimSpace(accountID))
	if err == nil {
		return accountInsertConflictAccountID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return accountInsertConflictUnknown, err
	}

	_, err = queries.GetAccountByUsernameKey(ctx, strings.TrimSpace(usernameKey))
	if err == nil {
		return accountInsertConflictUsernameKey, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return accountInsertConflictUnknown, err
	}

	return accountInsertConflictUnknown, nil
}

func generateAccountID() (string, error) {
	randomPart, err := randomHex(accountIDRandomByteSize)
	if err != nil {
		return "", err
	}
	return accountIDPrefix + randomPart, nil
}

func requestIsTLS(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.TLS != nil {
		return true
	}

	forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if forwardedProto == "" {
		return false
	}

	parts := strings.Split(forwardedProto, ",")
	if len(parts) == 0 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(parts[0]), "https")
}

func hashDashboardPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), dashboardPasswordBCryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func compareDashboardPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(strings.TrimSpace(hash)), []byte(password)) == nil
}

func randomHex(byteSize int) (string, error) {
	if byteSize <= 0 {
		return "", errors.New("byteSize must be positive")
	}

	raw := make([]byte, byteSize)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
