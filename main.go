package main

import (
	"log"
	"net/http"
)

func main() {
	cfg := LoadConfig()
	registry := NewRegistryState()
	watcher := NewWatcher(cfg, registry)

	if err := watcher.Start(); err != nil {
		log.Fatalf("[Gateway] failed to start watcher: %v", err)
	}
	defer watcher.Close()

	server := NewHTTPServer(cfg, registry)
	addr := cfg.ListenAddress()
	log.Printf("[Gateway] listening on %s, anox=%s", addr, cfg.AnoxURL)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("[Gateway] server error: %v", err)
	}
}