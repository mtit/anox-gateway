package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type Config struct {
	AnoxURL   string
	HTTPHost  string
	HTTPPort  string
	JWTSecret string
}

func LoadConfig() Config {
	cfg := Config{
		AnoxURL:   getEnv("ANOX_URL", "127.0.0.1:8848"),
		HTTPHost:  getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort:  getEnv("HTTP_PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", ""),
	}
	cfg.AnoxURL = strings.TrimSpace(cfg.AnoxURL)
	cfg.HTTPHost = strings.TrimSpace(cfg.HTTPHost)
	cfg.HTTPPort = strings.TrimSpace(cfg.HTTPPort)
	cfg.JWTSecret = strings.TrimSpace(cfg.JWTSecret)
	return cfg
}

func (c Config) WebSocketURL() string {
	return fmt.Sprintf("ws://%s/ws", c.AnoxURL)
}

func (c Config) ListenAddress() string {
	return netJoinHostPort(c.HTTPHost, c.HTTPPort)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func netJoinHostPort(host, port string) string {
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		return fmt.Sprintf("[%s]:%s", host, port)
	}
	return fmt.Sprintf("%s:%s", host, port)
}

type AuthState struct {
	mu            sync.RWMutex
	fallback      string
	globalSecret  string
	anoxSecret    string
	currentSecret string
}

func NewAuthState(fallback string) *AuthState {
	fallback = strings.TrimSpace(fallback)
	return &AuthState{fallback: fallback, currentSecret: fallback}
}

func (a *AuthState) Secret() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentSecret
}

func (a *AuthState) UpdateConfig(service string, values map[string]string) {
	secret := strings.TrimSpace(values["jwt_secret"])
	if secret == "" {
		secret = strings.TrimSpace(values["secret"])
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	switch service {
	case "_global":
		a.globalSecret = secret
	case "anox":
		a.anoxSecret = secret
	default:
		return
	}
	a.currentSecret = firstNonEmpty(a.globalSecret, a.anoxSecret, a.fallback)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
