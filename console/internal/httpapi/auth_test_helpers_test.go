package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onlyboxes/onlyboxes/console/internal/persistence"
	"github.com/onlyboxes/onlyboxes/console/internal/persistence/sqlc"
)

const (
	testDashboardUsername = "admin-test"
	testDashboardPassword = "password-test"
	testMCPToken          = "mcp-token-test"
	testMCPTokenB         = "mcp-token-test-b"
)

var testDashboardAccountID = mustGenerateTestAccountID()

func mustGenerateTestAccountID() string {
	accountID, err := generateAccountID()
	if err != nil {
		panic(err)
	}
	return accountID
}

func newTestConsoleAuth(t *testing.T) *ConsoleAuth {
	return newTestConsoleAuthWithRegistration(t, false)
}

func newTestConsoleAuthWithRegistration(t *testing.T, registrationEnabled bool) *ConsoleAuth {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	path := fmt.Sprintf("file:onlyboxes-consoleauth-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := persistence.Open(ctx, persistence.Options{
		Path:             path,
		BusyTimeoutMS:    5000,
		HashKey:          "test-hash-key",
		TaskRetentionDay: 30,
	})
	if err != nil {
		t.Fatalf("open test console auth db: %v", err)
	}
	seedTestAccount(db.Queries, testDashboardAccountID, testDashboardUsername, testDashboardPassword, true)
	return NewConsoleAuth(db.Queries, registrationEnabled)
}

func newTestMCPAuth() *MCPAuth {
	auth := newBareTestMCPAuth()
	tokenA := testMCPToken
	tokenB := testMCPTokenB
	if _, _, err := auth.createToken(context.Background(), testDashboardAccountID, "token-a", &tokenA); err != nil {
		panic(err)
	}
	if _, _, err := auth.createToken(context.Background(), testDashboardAccountID, "token-b", &tokenB); err != nil {
		panic(err)
	}
	return auth
}

func newBareTestMCPAuth() *MCPAuth {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	path := fmt.Sprintf("file:onlyboxes-mcpauth-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := persistence.Open(ctx, persistence.Options{
		Path:             path,
		BusyTimeoutMS:    5000,
		HashKey:          "test-hash-key",
		TaskRetentionDay: 30,
	})
	if err != nil {
		panic(err)
	}
	seedTestAccount(db.Queries, testDashboardAccountID, testDashboardUsername, testDashboardPassword, true)
	return NewMCPAuthWithPersistence(db)
}

func setMCPTokenHeader(req *http.Request) {
	if req == nil {
		return
	}
	req.Header.Set(trustedTokenHeader, testMCPToken)
}

func seedTestAccount(queries *sqlc.Queries, accountID string, username string, password string, isAdmin bool) {
	if queries == nil {
		panic("nil queries")
	}
	passwordHash, err := hashDashboardPassword(password)
	if err != nil {
		panic(err)
	}
	nowMS := time.Now().UnixMilli()
	if err := queries.InsertAccount(context.Background(), sqlc.InsertAccountParams{
		AccountID:       strings.TrimSpace(accountID),
		Username:        strings.TrimSpace(username),
		UsernameKey:     strings.ToLower(strings.TrimSpace(username)),
		PasswordHash:    passwordHash,
		HashAlgo:        dashboardPasswordHashAlgo,
		IsAdmin:         boolToInt64(isAdmin),
		CreatedAtUnixMs: nowMS,
		UpdatedAtUnixMs: nowMS,
	}); err != nil {
		panic(err)
	}
}

func loginSessionCookie(t *testing.T, router http.Handler) *http.Cookie {
	t.Helper()

	body, err := json.Marshal(loginRequest{
		Username: testDashboardUsername,
		Password: testDashboardPassword,
	})
	if err != nil {
		t.Fatalf("failed to marshal login request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/console/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected login success, got %d body=%s", rec.Code, rec.Body.String())
	}

	resp := rec.Result()
	defer resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == dashboardSessionCookieName {
			return cookie
		}
	}
	t.Fatalf("expected %s cookie in login response", dashboardSessionCookieName)
	return nil
}
