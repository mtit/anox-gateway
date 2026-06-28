package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Watcher struct {
	cfg      Config
	registry *RegistryState
	conn     *websocket.Conn
	mu       sync.Mutex
	closed   bool
}

func NewWatcher(cfg Config, registry *RegistryState) *Watcher {
	return &Watcher{cfg: cfg, registry: registry}
}

func (w *Watcher) Start() error {
	if err := w.connect(); err != nil {
		return err
	}
	log.Printf("[Gateway] watcher connected to %s", w.cfg.WebSocketURL())
	go w.readLoop()
	return nil
}

func (w *Watcher) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = true
	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
	}
}

func (w *Watcher) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(w.cfg.WebSocketURL(), nil)
	if err != nil {
		return err
	}
	if err := conn.WriteJSON(map[string]string{"type": "watch_services"}); err != nil {
		conn.Close()
		return err
	}
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		conn.Close()
		return nil
	}
	if w.conn != nil {
		w.conn.Close()
	}
	w.conn = conn
	w.mu.Unlock()
	return nil
}

func (w *Watcher) readLoop() {
	for {
		conn := w.currentConn()
		if conn == nil {
			if !w.retryConnect() {
				return
			}
			continue
		}
		_, payload, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Gateway] watcher disconnected: %v", err)
			w.dropConn(conn)
			if !w.retryConnect() {
				return
			}
			continue
		}
		w.handleMessage(payload)
	}
}

func (w *Watcher) handleMessage(payload []byte) {
	var envelope struct{ Type string `json:"type"` }
	if err := json.Unmarshal(payload, &envelope); err != nil {
		log.Printf("[Gateway] invalid watcher message: %v", err)
		return
	}
	switch envelope.Type {
	case "services_snapshot":
		var msg servicesSnapshotMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("[Gateway] invalid services snapshot: %v", err)
			return
		}
		w.registry.ReplaceAll(msg.Services)
	case "service_event":
		var msg serviceEventMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("[Gateway] invalid service event: %v", err)
			return
		}
		w.registry.ApplyEvent(msg)
	}
}

func (w *Watcher) retryConnect() bool {
	for {
		w.mu.Lock()
		closed := w.closed
		w.mu.Unlock()
		if closed {
			return false
		}
		err := w.connect()
		if err == nil {
			log.Printf("[Gateway] watcher connected to %s", w.cfg.WebSocketURL())
			return true
		}
		log.Printf("[Gateway] watcher reconnect failed: %v", err)
		time.Sleep(2 * time.Second)
	}
}

func (w *Watcher) currentConn() *websocket.Conn {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn
}

func (w *Watcher) dropConn(conn *websocket.Conn) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn == conn {
		w.conn.Close()
		w.conn = nil
	}
}