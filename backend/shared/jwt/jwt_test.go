package jwt_test

import (
	"os"
	"testing"

	"shared/jwt"
)

func TestGenerateAndVerifyAccessToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	token, err := jwt.GenerateAccessToken("user-1", "test@example.com")
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	claims, err := jwt.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("verify access token: %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "test@example.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestGenerateAndVerifyRefreshToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	token, tokenID, err := jwt.GenerateRefreshToken("user-1", "test@example.com")
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}
	if tokenID == "" {
		t.Fatal("expected token id")
	}

	claims, err := jwt.VerifyRefreshToken(token)
	if err != nil {
		t.Fatalf("verify refresh token: %v", err)
	}
	if claims.TokenID != tokenID {
		t.Fatalf("token id mismatch")
	}
}
