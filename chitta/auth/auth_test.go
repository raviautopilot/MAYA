package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}
	if hash == "testpassword" {
		t.Fatal("hash should not equal plaintext")
	}
}

func TestCheckPassword_Valid(t *testing.T) {
	hash, _ := HashPassword("mypassword")
	if err := CheckPassword(hash, "mypassword"); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

func TestCheckPassword_Invalid(t *testing.T) {
	hash, _ := HashPassword("mypassword")
	if err := CheckPassword(hash, "wrongpassword"); err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestGenerateAndValidateJWT(t *testing.T) {
	secret := "test-secret-key"
	email := "test@example.com"

	token, err := GenerateJWT(email, secret, 1)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	if claims.Email != email {
		t.Errorf("expected email '%s', got '%s'", email, claims.Email)
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	token, _ := GenerateJWT("test@example.com", "correct-secret", 1)
	_, err := ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	claims := Claims{
		Email: "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))

	_, err := ValidateJWT(signed, secret)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateJWT_InvalidFormat(t *testing.T) {
	_, err := ValidateJWT("not.a.valid.token", "secret")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestGenerateJWT_DefaultExpiry(t *testing.T) {
	token, err := GenerateJWT("test@example.com", "secret", 0)
	if err != nil {
		t.Fatalf("GenerateJWT with 0 expiry failed: %v", err)
	}
	claims, err := ValidateJWT(token, "secret")
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	if claims.Email != "test@example.com" {
		t.Error("unexpected email in claims")
	}
}
