package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onlyboxes/onlyboxes/console/internal/persistence"
	"github.com/onlyboxes/onlyboxes/console/internal/testutil/registrytest"
)

func TestResolveDashboardCredentials(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		password         string
		expectUsername   string
		expectPassword   string
		expectRandomUser bool
		expectRandomPass bool
	}{
		{
			name:             "both-random",
			expectRandomUser: true,
			expectRandomPass: true,
		},
		{
			name:             "only-username-configured",
			username:         "admin-fixed",
			expectUsername:   "admin-fixed",
			expectRandomPass: true,
		},
		{
			name:             "only-password-configured",
			password:         "secret-fixed",
			expectPassword:   "secret-fixed",
			expectRandomUser: true,
		},
		{
			name:           "both-configured",
			username:       "admin-fixed",
			password:       "secret-fixed",
			expectUsername: "admin-fixed",
			expectPassword: "secret-fixed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			credentials, err := ResolveDashboardCredentials(tc.username, tc.password)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.expectRandomUser {
				if !strings.HasPrefix(credentials.Username, dashboardUsernamePrefix) {
					t.Fatalf("expected username prefix %q, got %q", dashboardUsernamePrefix, credentials.Username)
				}
				if len(credentials.Username) <= len(dashboardUsernamePrefix) {
					t.Fatalf("expected random username suffix, got %q", credentials.Username)
				}
			} else if credentials.Username != tc.expectUsername {
				t.Fatalf("expected username %q, got %q", tc.expectUsername, credentials.Username)
			}

			if tc.expectRandomPass {
				if strings.TrimSpace(credentials.Password) == "" {
					t.Fatalf("expected random password, got empty")
				}
			} else if credentials.Password != tc.expectPassword {
				t.Fatalf("expected password %q, got %q", tc.expectPassword, credentials.Password)
			}
		})
	}
}

func TestInitializeAdminAccountPersistsOnFirstRun(t *testing.T) {
	ctx := context.Background()
	db := openTestAuthDB(t)
	defer func() {
		_ = db.Close()
	}()

	result, err := InitializeAdminAccount(ctx, db.Queries, "", "")
	if err != nil {
		t.Fatalf("initialize admin account: %v", err)
	}
	if !result.InitializedNow {
		t.Fatalf("expected initialized now")
	}
	if result.EnvIgnored {
		t.Fatalf("expected env ignored=false")
	}
	if !strings.HasPrefix(result.Username, dashboardUsernamePrefix) {
		t.Fatalf("expected generated username prefix %q, got %q", dashboardUsernamePrefix, result.Username)
	}
	if strings.TrimSpace(result.PasswordPlaintext) == "" {
		t.Fatalf("expected plaintext password in first init result")
	}
	if strings.TrimSpace(result.AccountID) == "" {
		t.Fatalf("expected account_id in init result")
	}

	stored, err := db.Queries.GetAccountByID(ctx, result.AccountID)
	if err != nil {
		t.Fatalf("load persisted account: %v", err)
	}
	if stored.Username != result.Username {
		t.Fatalf("unexpected username: %q", stored.Username)
	}
	if stored.IsAdmin != 1 {
		t.Fatalf("expected is_admin=1, got %d", stored.IsAdmin)
	}
	if !strings.EqualFold(stored.HashAlgo, dashboardPasswordHashAlgo) {
		t.Fatalf("unexpected hash algo: %q", stored.HashAlgo)
	}
	if !compareDashboardPassword(stored.PasswordHash, result.PasswordPlaintext) {
		t.Fatalf("expected stored hash to match initialized password")
	}
}

func TestInitializeAdminAccountLoadsPersistedAndIgnoresEnv(t *testing.T) {
	ctx := context.Background()
	db := openTestAuthDB(t)
	defer func() {
		_ = db.Close()
	}()

	first, err := InitializeAdminAccount(ctx, db.Queries, "admin-first", "password-first")
	if err != nil {
		t.Fatalf("first initialize admin account: %v", err)
	}
	if !first.InitializedNow {
		t.Fatalf("expected first initialization")
	}

	second, err := InitializeAdminAccount(ctx, db.Queries, "admin-second", "password-second")
	if err != nil {
		t.Fatalf("second initialize admin account: %v", err)
	}
	if second.InitializedNow {
		t.Fatalf("expected loading persisted admin account")
	}
	if !second.EnvIgnored {
		t.Fatalf("expected env ignored=true when admin account exists")
	}
	if second.AccountID != first.AccountID {
		t.Fatalf("expected persisted account_id %q, got %q", first.AccountID, second.AccountID)
	}
	if second.Username != first.Username {
		t.Fatalf("expected persisted username %q, got %q", first.Username, second.Username)
	}
	if second.PasswordPlaintext != "" {
		t.Fatalf("expected empty plaintext password for loaded account")
	}
}

func TestInitializeAdminAccountRetriesOnAccountIDConflict(t *testing.T) {
	ctx := context.Background()
	db := openTestAuthDB(t)
	defer func() {
		_ = db.Close()
	}()

	seedTestAccount(db.Queries, "acc-conflict", "seed-user", "seed-pass", false)

	previousGenerator := accountIDGenerator
	sequence := []string{"acc-conflict", "acc-retry-success"}
	generateIdx := 0
	accountIDGenerator = func() (string, error) {
		if generateIdx >= len(sequence) {
			return "", errors.New("account id sequence exhausted")
		}
		value := sequence[generateIdx]
		generateIdx++
		return value, nil
	}
	t.Cleanup(func() {
		accountIDGenerator = previousGenerator
	})

	result, err := InitializeAdminAccount(ctx, db.Queries, "admin-retry", "password-retry")
	if err != nil {
		t.Fatalf("initialize admin account with account_id conflict retry: %v", err)
	}
	if !result.InitializedNow {
		t.Fatalf("expected initialized now")
	}
	if result.AccountID != "acc-retry-success" {
		t.Fatalf("expected retried account_id acc-retry-success, got %q", result.AccountID)
	}
}

func TestInitializeAdminAccountReturnsConflictOnUsernameKeyCollision(t *testing.T) {
	ctx := context.Background()
	db := openTestAuthDB(t)
	defer func() {
		_ = db.Close()
	}()

	seedTestAccount(db.Queries, "acc-existing", "admin-dup", "seed-pass", false)

	_, err := InitializeAdminAccount(ctx, db.Queries, "ADMIN-dup", "password-new")
	if !errors.Is(err, errAccountRegistrationConflict) {
		t.Fatalf("expected errAccountRegistrationConflict, got %v", err)
	}
}

func TestConsoleAuthLoginLogoutLifecycle(t *testing.T) {
	handler := NewWorkerHandler(registrytest.NewStore(t), 15*time.Second, nil, nil, nil, "")
	auth := newTestConsoleAuth(t)
	router := NewRouter(handler, auth, newTestMCPAuth())

	failedReq := httptest.NewRequest(http.MethodPost, "/api/v1/console/login", strings.NewReader(`{"username":"wrong","password":"wrong"}`))
	failedReq.Header.Set("Content-Type", "application/json")
	failedRec := httptest.NewRecorder()
	router.ServeHTTP(failedRec, failedReq)
	if failedRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid login, got %d", failedRec.Code)
	}

	sessionCookie := loginSessionCookie(t, router)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/workers", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for authenticated admin list request, got %d body=%s", listRec.Code, listRec.Body.String())
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/console/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutRec := httptest.NewRecorder()
	router.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for logout, got %d", logoutRec.Code)
	}

	listAfterLogoutReq := httptest.NewRequest(http.MethodGet, "/api/v1/workers", nil)
	listAfterLogoutReq.AddCookie(sessionCookie)
	listAfterLogoutRec := httptest.NewRecorder()
	router.ServeHTTP(listAfterLogoutRec, listAfterLogoutReq)
	if listAfterLogoutRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d body=%s", listAfterLogoutRec.Code, listAfterLogoutRec.Body.String())
	}
}

func TestConsoleAuthSessionEndpoint(t *testing.T) {
	handler := NewWorkerHandler(registrytest.NewStore(t), 15*time.Second, nil, nil, nil, "")
	auth := newTestConsoleAuthWithRegistration(t, true)
	router := NewRouter(handler, auth, newTestMCPAuth())
	cookie := loginSessionCookie(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/console/session", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	payload := accountSessionResponse{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode session payload: %v", err)
	}
	if !payload.Authenticated {
		t.Fatalf("expected authenticated=true")
	}
	if payload.Account.AccountID == "" || payload.Account.Username != testDashboardUsername || !payload.Account.IsAdmin {
		t.Fatalf("unexpected session account payload: %#v", payload.Account)
	}
	if !payload.RegistrationEnabled {
		t.Fatalf("expected registration_enabled=true")
	}
}

func TestConsoleAuthSessionExpires(t *testing.T) {
	handler := NewWorkerHandler(registrytest.NewStore(t), 15*time.Second, nil, nil, nil, "")
	auth := newTestConsoleAuth(t)
	now := time.Unix(1_700_000_000, 0)
	auth.nowFn = func() time.Time {
		return now
	}
	router := NewRouter(handler, auth, newTestMCPAuth())
	sessionCookie := loginSessionCookie(t, router)

	now = now.Add(dashboardSessionTTL + time.Second)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workers", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired session, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestConsoleAuthRegisterAndAdminGuard(t *testing.T) {
	handler := NewWorkerHandler(registrytest.NewStore(t), 15*time.Second, nil, nil, nil, "")
	auth := newTestConsoleAuthWithRegistration(t, true)
	router := NewRouter(handler, auth, newTestMCPAuth())
	adminCookie := loginSessionCookie(t, router)

	registerBody := []byte(`{"username":"member-a","password":"member-a-pass"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/console/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerReq.AddCookie(adminCookie)
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", registerRec.Code, registerRec.Body.String())
	}

	nonAdminCookie := loginSessionCookieFor(t, router, "member-a", "member-a-pass")

	workersReq := httptest.NewRequest(http.MethodGet, "/api/v1/workers", nil)
	workersReq.AddCookie(nonAdminCookie)
	workersRec := httptest.NewRecorder()
	router.ServeHTTP(workersRec, workersReq)
	if workersRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin workers access, got %d body=%s", workersRec.Code, workersRec.Body.String())
	}

	nonAdminRegisterReq := httptest.NewRequest(http.MethodPost, "/api/v1/console/register", bytes.NewReader(registerBody))
	nonAdminRegisterReq.Header.Set("Content-Type", "application/json")
	nonAdminRegisterReq.AddCookie(nonAdminCookie)
	nonAdminRegisterRec := httptest.NewRecorder()
	router.ServeHTTP(nonAdminRegisterRec, nonAdminRegisterReq)
	if nonAdminRegisterRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin register, got %d body=%s", nonAdminRegisterRec.Code, nonAdminRegisterRec.Body.String())
	}
}

func TestConsoleAuthRegisterDuplicateUsernameConflict(t *testing.T) {
	handler := NewWorkerHandler(registrytest.NewStore(t), 15*time.Second, nil, nil, nil, "")
	auth := newTestConsoleAuthWithRegistration(t, true)
	router := NewRouter(handler, auth, newTestMCPAuth())
	adminCookie := loginSessionCookie(t, router)

	registerReqA := httptest.NewRequest(http.MethodPost, "/api/v1/console/register", strings.NewReader(`{"username":"member-dup","password":"member-pass"}`))
	registerReqA.Header.Set("Content-Type", "application/json")
	registerReqA.AddCookie(adminCookie)
	registerRecA := httptest.NewRecorder()
	router.ServeHTTP(registerRecA, registerReqA)
	if registerRecA.Code != http.StatusCreated {
		t.Fatalf("expected first register 201, got %d body=%s", registerRecA.Code, registerRecA.Body.String())
	}

	registerReqB := httptest.NewRequest(http.MethodPost, "/api/v1/console/register", strings.NewReader(`{"username":"MEMBER-dup","password":"member-pass-2"}`))
	registerReqB.Header.Set("Content-Type", "application/json")
	registerReqB.AddCookie(adminCookie)
	registerRecB := httptest.NewRecorder()
	router.ServeHTTP(registerRecB, registerReqB)
	if registerRecB.Code != http.StatusConflict {
		t.Fatalf("expected duplicate register 409, got %d body=%s", registerRecB.Code, registerRecB.Body.String())
	}
	if !strings.Contains(registerRecB.Body.String(), errAccountRegistrationConflict.Error()) {
		t.Fatalf("expected duplicate register error message %q, got %s", errAccountRegistrationConflict.Error(), registerRecB.Body.String())
	}
}

func TestConsoleAuthRegisterDisabled(t *testing.T) {
	handler := NewWorkerHandler(registrytest.NewStore(t), 15*time.Second, nil, nil, nil, "")
	auth := newTestConsoleAuthWithRegistration(t, false)
	router := NewRouter(handler, auth, newTestMCPAuth())
	adminCookie := loginSessionCookie(t, router)

	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/console/register", strings.NewReader(`{"username":"member-x","password":"pass"}`))
	registerReq.Header.Set("Content-Type", "application/json")
	registerReq.AddCookie(adminCookie)
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when registration is disabled, got %d body=%s", registerRec.Code, registerRec.Body.String())
	}
}

func loginSessionCookieFor(t *testing.T, router http.Handler, username string, password string) *http.Cookie {
	t.Helper()
	body, err := json.Marshal(loginRequest{Username: username, Password: password})
	if err != nil {
		t.Fatalf("marshal login payload: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/console/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected login success for %q, got %d body=%s", username, rec.Code, rec.Body.String())
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == dashboardSessionCookieName {
			return cookie
		}
	}
	t.Fatalf("expected %s cookie in login response", dashboardSessionCookieName)
	return nil
}

func openTestAuthDB(t *testing.T) *persistence.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dashboard-auth.db")
	db, err := persistence.Open(context.Background(), persistence.Options{
		Path:             path,
		BusyTimeoutMS:    5000,
		HashKey:          "test-hash-key",
		TaskRetentionDay: 30,
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return db
}
