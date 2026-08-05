package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/AlRowne/chirpy/internal/database/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	pw := "lolkek"
	hash, err := auth.HashPassword(pw)
	if err != nil {
		t.Errorf("HashPassword returned an error: %v", err)
	}
	if hash == "" {
		t.Error("empty hash")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	pw := "lolkek"
	hash, err := auth.HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword returned an error: %v", err)
	}
	match, err := auth.CheckPasswordHash(pw, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash couldn't check passwords: %v", err)
	}
	if !match {
		t.Error("password didn't match hash, even though it should")
	}
}

func TestCheckWrongPassword(t *testing.T) {
	pw := "lolkek"
	hash, err := auth.HashPassword(pw)
	if err != nil {
		t.Errorf("HashPassword returned an error: %v", err)
	}
	match, err := auth.CheckPasswordHash("wrongpw", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash couldn't check passwords: %v", err)
	}
	if match {
		t.Error("password matched hash, even though it shouldn't")
	}
}
func TestCheckWrongPasswordHash(t *testing.T) {
	pw := "lolkek"
	match, err := auth.CheckPasswordHash(pw, "wronghash")
	if err == nil {
		t.Error("checked a wrong hash, even though it shouldn't")
	}
	if match {
		t.Error("password matched hash, even though it shouldn't")
	}
}

func TestDifferentPWHaveSameHash(t *testing.T) {
	hash1, _ := auth.HashPassword("lolkek")
	hash2, _ := auth.HashPassword("lolkek")

	if hash1 == hash2 {
		t.Error("expected different hashes for same password (different salts)")
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenString, err := auth.MakeJWT(userID, "supersecret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned an error: %v", err)
	}
	if tokenString == "" {
		t.Error("expected a non-empty token string")
	}
}

func TestMakeJWTAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "supersecret"

	tokenString, err := auth.MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned an error: %v", err)
	}

	gotID, err := auth.ValidateJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned an error: %v", err)
	}
	if gotID != userID {
		t.Errorf("expected userID %v, got %v", userID, gotID)
	}
}

func TestValidateJWTMultipleUsers(t *testing.T) {
	secret := "supersecret"
	userIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	for _, userID := range userIDs {
		tokenString, err := auth.MakeJWT(userID, secret, time.Hour)
		if err != nil {
			t.Fatalf("MakeJWT returned an error: %v", err)
		}
		gotID, err := auth.ValidateJWT(tokenString, secret)
		if err != nil {
			t.Fatalf("ValidateJWT returned an error: %v", err)
		}
		if gotID != userID {
			t.Errorf("expected userID %v, got %v", userID, gotID)
		}
	}
}

func TestValidateJWTExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "supersecret"

	tokenString, err := auth.MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned an error: %v", err)
	}

	_, err = auth.ValidateJWT(tokenString, secret)
	if err == nil {
		t.Error("expected an error for an expired token, got nil")
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenString, err := auth.MakeJWT(userID, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned an error: %v", err)
	}

	_, err = auth.ValidateJWT(tokenString, "wrong-secret")
	if err == nil {
		t.Error("expected an error for a wrong secret, got nil")
	}
}

func TestValidateJWTMalformedToken(t *testing.T) {
	_, err := auth.ValidateJWT("this-is-not-a-jwt", "supersecret")
	if err == nil {
		t.Error("expected an error for a malformed token, got nil")
	}
}

func TestValidateJWTEmptyToken(t *testing.T) {
	_, err := auth.ValidateJWT("", "supersecret")
	if err == nil {
		t.Error("expected an error for an empty token, got nil")
	}
}

func TestValidateJWTInvalidSubject(t *testing.T) {
	secret := "supersecret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Subject:   "not-a-uuid",
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("couldn't sign test token: %v", err)
	}

	_, err = auth.ValidateJWT(tokenString, secret)
	if err == nil {
		t.Error("expected an error for a non-UUID subject, got nil")
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name         string
		authorizaton string
		wantToken    string
		wantErr      bool
	}{
		{
			name:         "valid bearer token",
			authorizaton: "Bearer abc.def.ghi",
			wantToken:    "abc.def.ghi",
		},
		{
			name:         "valid bearer token with extra whitespace",
			authorizaton: "  Bearer   abc.def.ghi  ",
			wantToken:    "abc.def.ghi",
		},
		{
			name:    "missing authorization header",
			wantErr: true,
		},
		{
			name:         "wrong authentication scheme",
			authorizaton: "Basic abc.def.ghi",
			wantErr:      true,
		},
		{
			name:         "missing token",
			authorizaton: "Bearer",
			wantErr:      true,
		},
		{
			name:         "invalid bearer prefix",
			authorizaton: "BearerXYZ abc.def.ghi",
			wantErr:      true,
		},
		{
			name:         "too many header values",
			authorizaton: "Bearer abc.def.ghi extra",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.authorizaton != "" {
				headers.Set("Authorization", tt.authorizaton)
			}

			gotToken, err := auth.GetBearerToken(headers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() = %q, want %q", gotToken, tt.wantToken)
			}
		})
	}
}

func TestMakeJWTIssuer(t *testing.T) {
	userID := uuid.New()
	secret := "supersecret"
	tokenString, err := auth.MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned an error: %v", err)
	}

	claims := &jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("couldn't parse token: %v", err)
	}
	if claims.Issuer != "chirpy-access" {
		t.Errorf("expected issuer %q, got %q", "chirpy-access", claims.Issuer)
	}
}
