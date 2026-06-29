package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"
)

func TestUserIDFromAuthorization(t *testing.T) {
	now := time.Unix(1700000000, 0)
	token := testJWT(t, `{"alg":"HS256","typ":"JWT"}`, `{"user_id":123,"iat":1700000000,"exp":1700003600}`, "secret")

	if got := userIDFromAuthorization("Bearer "+token, "secret", now); got != "123" {
		t.Fatalf("userIDFromAuthorization() = %q, want 123", got)
	}
	if got := userIDFromAuthorization("Bearer "+token, "other-secret", now); got != anonymousUserID {
		t.Fatalf("wrong secret user id = %q, want anonymous", got)
	}
	if got := userIDFromAuthorization("Bearer "+token, "secret", now.Add(2*time.Hour)); got != anonymousUserID {
		t.Fatalf("expired token user id = %q, want anonymous", got)
	}
}

func TestVerifyJWTRejectsUnsupportedAlgorithm(t *testing.T) {
	now := time.Unix(1700000000, 0)
	token := testJWT(t, `{"alg":"HS384","typ":"JWT"}`, `{"user_id":"u1","iat":1700000000,"exp":1700003600}`, "secret")

	if _, _, ok := verifyJWT(token, "secret", now); ok {
		t.Fatal("expected unsupported algorithm to be rejected")
	}
}

func testJWT(t *testing.T, header, payload, secret string) string {
	t.Helper()

	encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	signingInput := encodedHeader + "." + encodedPayload

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature
}
