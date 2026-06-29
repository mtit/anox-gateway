package main

import (
	"context"
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
	currentSecret string
	ready         chan struct{}
	readyOnce     sync.Once
}

func NewAuthState(fallback string) *AuthState {
	fallback = strings.TrimSpace(fallback)
	auth := &AuthState{
		fallback:      fallback,
		currentSecret: fallback,
		ready:         make(chan struct{}),
	}
	if fallback != "" {
		auth.readyOnce.Do(func() { close(auth.ready) })
	}
	return auth
}

func (a *AuthState) Secret() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentSecret
}

func (a *AuthState) Source() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.globalSecret != "" {
		return "_global.jwt_secret"
	}
	if a.fallback != "" {
		return "JWT_SECRET"
	}
	return ""
}

func (a *AuthState) UpdateConfig(service string, values map[string]string) {
	if service != "_global" {
		return
	}
	secret := strings.TrimSpace(values["jwt_secret"])
	a.mu.Lock()
	defer a.mu.Unlock()
	a.globalSecret = secret
	a.currentSecret = firstNonEmpty(a.globalSecret, a.fallback)
	if a.currentSecret != "" {
		a.readyOnce.Do(func() { close(a.ready) })
	}
}

func (a *AuthState) Wait(ctx context.Context) error {
	if a.Secret() != "" {
		return nil
	}
	select {
	case <-a.ready:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("missing jwt secret from anox config: %w", ctx.Err())
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
