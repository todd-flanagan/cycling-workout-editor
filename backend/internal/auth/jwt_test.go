package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	mgr := NewJWTManager("test-secret-key", 24*time.Hour)

	token, err := mgr.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("Subject: got %q, want %q", claims.Subject, "user-123")
	}
}

func TestValidateTokenExpired(t *testing.T) {
	mgr := NewJWTManager("test-secret-key", -1*time.Hour)

	token, err := mgr.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = mgr.ValidateToken(token)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	mgr := NewJWTManager("test-secret-key", 24*time.Hour)

	_, err := mgr.ValidateToken("this-is-not-a-valid-token")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	mgr1 := NewJWTManager("secret-one", 24*time.Hour)
	mgr2 := NewJWTManager("secret-two", 24*time.Hour)

	token, err := mgr1.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = mgr2.ValidateToken(token)
	if err == nil {
		t.Error("expected error when validating with wrong secret, got nil")
	}
}

func TestValidateTokenEmptyString(t *testing.T) {
	mgr := NewJWTManager("test-secret-key", 24*time.Hour)

	_, err := mgr.ValidateToken("")
	if err == nil {
		t.Error("expected error for empty token, got nil")
	}
}

func TestTokenContainsIssuedAt(t *testing.T) {
	mgr := NewJWTManager("test-secret-key", 24*time.Hour)

	token, err := mgr.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.IssuedAt == nil {
		t.Error("IssuedAt should be set")
	}
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}

	// Verify expiry is roughly 24 hours from now.
	diff := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	if diff < 23*time.Hour || diff > 25*time.Hour {
		t.Errorf("expected ~24h between iat and exp, got %v", diff)
	}
}

func TestValidateTokenWrongSigningMethod(t *testing.T) {
	// Create a token with a different signing method (none).
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// Use "none" signing method to test rejection.
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	mgr := NewJWTManager("test-secret-key", 24*time.Hour)
	_, err := mgr.ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error for wrong signing method, got nil")
	}
}

func TestGenerateTokenDifferentUsers(t *testing.T) {
	mgr := NewJWTManager("test-secret-key", 24*time.Hour)

	token1, err := mgr.GenerateToken("user-1")
	if err != nil {
		t.Fatalf("GenerateToken user-1 failed: %v", err)
	}

	token2, err := mgr.GenerateToken("user-2")
	if err != nil {
		t.Fatalf("GenerateToken user-2 failed: %v", err)
	}

	if token1 == token2 {
		t.Error("tokens for different users should be different")
	}

	claims1, _ := mgr.ValidateToken(token1)
	claims2, _ := mgr.ValidateToken(token2)

	if claims1.Subject != "user-1" {
		t.Errorf("claims1 subject: got %q, want %q", claims1.Subject, "user-1")
	}
	if claims2.Subject != "user-2" {
		t.Errorf("claims2 subject: got %q, want %q", claims2.Subject, "user-2")
	}
}
