package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

type HTTPServer struct {
	cfg      Config
	registry *RegistryState
}

func NewHTTPServer(cfg Config, registry *RegistryState) *HTTPServer {
	return &HTTPServer{cfg: cfg, registry: registry}
}

func (s *HTTPServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/", s.handleProxy)
	return mux
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *HTTPServer) handleProxy(w http.ResponseWriter, r *http.Request) {
	service, targetPath, ok := splitServicePath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	instance, err := s.registry.Pick(service)
	if err != nil {
		if errors.Is(err, ErrNoAvailableInstance) {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, "gateway error", http.StatusBadGateway)
		return
	}

	target := "http://" + netJoinHostPort(instance.HttpHost, instance.HttpPort)
	proxy := httputil.NewSingleHostReverseProxy(mustParseTarget(target))
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = targetPath
		req.Host = req.URL.Host
		req.Header.Set("X-Anox-Gateway", "anox-gateway")
		req.Header.Set("X-Anox-Service", service)
		req.Header.Set("X-Anox-Instance", instance.ID)
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, proxyErr error) {
		log.Printf("[Gateway] proxy error for %s via %s: %v", service, target, proxyErr)
		http.Error(rw, "bad gateway", http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, r)
}

func splitServicePath(path string) (service, target string, ok bool) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 || parts[0] != "api" {
		return "", "", false
	}
	service = parts[1] + "-service"
	remaining := parts[2:]
	if len(remaining) == 0 {
		target = "/"
	} else {
		target = "/" + strings.Join(remaining, "/")
	}
	return service, target, true
}