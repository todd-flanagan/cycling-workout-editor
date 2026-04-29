package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tflanaga/cycling-workout-editor/backend/internal/auth"
)

func newTestJWTManager() *auth.JWTManager {
	return auth.NewJWTManager("test-middleware-secret", 24*time.Hour)
}

func TestJWTAuthValidToken(t *testing.T) {
	mgr := newTestJWTManager()
	token, err := mgr.GenerateToken("user-42")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	var capturedUserID string
	handler := JWTAuth(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if capturedUserID != "user-42" {
		t.Errorf("user ID: got %q, want %q", capturedUserID, "user-42")
	}
}

func TestJWTAuthMissingHeader(t *testing.T) {
	mgr := newTestJWTManager()

	handler := JWTAuth(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with missing auth header")
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}

	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["error"] != "missing authorization header" {
		t.Errorf("error message: got %q", body["error"])
	}
}

func TestJWTAuthInvalidHeaderFormat(t *testing.T) {
	mgr := newTestJWTManager()

	handler := JWTAuth(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	// Test with wrong prefix.
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestJWTAuthInvalidToken(t *testing.T) {
	mgr := newTestJWTManager()

	handler := JWTAuth(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid token")
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}

	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["error"] != "invalid or expired token" {
		t.Errorf("error message: got %q", body["error"])
	}
}

func TestJWTAuthExpiredToken(t *testing.T) {
	// Create a manager with already-expired tokens.
	expiredMgr := auth.NewJWTManager("test-middleware-secret", -1*time.Hour)
	token, err := expiredMgr.GenerateToken("user-42")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	mgr := newTestJWTManager()
	handler := JWTAuth(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with expired token")
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestJWTAuthBearerCaseInsensitive(t *testing.T) {
	mgr := newTestJWTManager()
	token, err := mgr.GenerateToken("user-42")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	var capturedUserID string
	handler := JWTAuth(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// Use lowercase "bearer".
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if capturedUserID != "user-42" {
		t.Errorf("user ID: got %q, want %q", capturedUserID, "user-42")
	}
}

func TestJWTAuthOnlyBearer(t *testing.T) {
	mgr := newTestJWTManager()

	handler := JWTAuth(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	// Just "Bearer" with no token.
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-99")
	uid := GetUserID(ctx)
	if uid != "user-99" {
		t.Errorf("GetUserID: got %q, want %q", uid, "user-99")
	}
}

func TestGetUserIDFromEmptyContext(t *testing.T) {
	uid := GetUserID(context.Background())
	if uid != "" {
		t.Errorf("GetUserID from empty context: got %q, want empty string", uid)
	}
}

func TestJWTAuthResponseContentType(t *testing.T) {
	mgr := newTestJWTManager()

	handler := JWTAuth(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
	}
}
