package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const anonymousUserID = "0"

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type jwtClaims struct {
	UserID    any   `json:"user_id"`
	IssuedAt  int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
}

func userIDFromAuthorization(authHeader, secret string, now time.Time) string {
	if secret == "" {
		return anonymousUserID
	}

	token := strings.TrimSpace(authHeader)
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return anonymousUserID
	}

	userID, ok := verifyJWT(token, secret, now)
	if !ok {
		return anonymousUserID
	}
	return userID
}

func verifyJWT(token, secret string, now time.Time) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", false
	}

	var header jwtHeader
	if err := decodeJWTPart(parts[0], &header); err != nil {
		return "", false
	}
	if header.Algorithm != "HS256" {
		return "", false
	}

	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return "", false
	}

	var claims jwtClaims
	if err := decodeJWTPart(parts[1], &claims); err != nil {
		return "", false
	}
	if claims.IssuedAt <= 0 || claims.ExpiresAt <= now.Unix() || claims.IssuedAt > now.Add(time.Minute).Unix() {
		return "", false
	}

	userID := formatUserID(claims.UserID)
	if userID == "" {
		return "", false
	}
	return userID, true
}

func decodeJWTPart(part string, dst any) error {
	raw, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}

func formatUserID(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return fmt.Sprintf("%v", v)
	default:
		return ""
	}
}
