package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type AuditEvent struct {
	Actor    string `json:"actor"`
	Action   string `json:"action"`
	Entity   string `json:"entity"`
	Recorded string `json:"recorded"`
}

type AuditStore struct {
	mu     sync.RWMutex
	events []AuditEvent
}

func NewAuditStore() *AuditStore {
	return &AuditStore{events: make([]AuditEvent, 0)}
}

func (s *AuditStore) Add(event AuditEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, event)
}

func (s *AuditStore) List() []AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copyEvents := make([]AuditEvent, len(s.events))
	copy(copyEvents, s.events)
	return copyEvents
}

type AuditRequest struct {
	Actor  string `json:"actor"`
	Action string `json:"action"`
	Entity string `json:"entity"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()
	store := NewAuditStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, http.StatusOK, store.List())
		case http.MethodPost:
			var req AuditRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			store.Add(AuditEvent{Actor: req.Actor, Action: req.Action, Entity: req.Entity, Recorded: time.Now().UTC().Format(time.RFC3339)})
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "audit-log"
	}
	return serviceName
}

func startServer(serviceName string, mux *http.ServeMux) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("%s listening on :%s", serviceName, port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, HealthResponse{Status: "ok", Service: serviceName})
	}
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
