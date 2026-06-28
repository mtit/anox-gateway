package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AnoxURL  string
	HTTPHost string
	HTTPPort string
}

func LoadConfig() Config {
	cfg := Config{
		AnoxURL:  getEnv("ANOX_URL", "127.0.0.1:8848"),
		HTTPHost: getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
	}
	cfg.AnoxURL = strings.TrimSpace(cfg.AnoxURL)
	cfg.HTTPHost = strings.TrimSpace(cfg.HTTPHost)
	cfg.HTTPPort = strings.TrimSpace(cfg.HTTPPort)
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