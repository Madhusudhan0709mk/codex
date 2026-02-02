package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

type EventCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type AnalyticsStore struct {
	mu     sync.RWMutex
	counts map[string]int
}

func NewAnalyticsStore() *AnalyticsStore {
	return &AnalyticsStore{counts: make(map[string]int)}
}

func (s *AnalyticsStore) Increment(eventType string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[eventType]++
}

func (s *AnalyticsStore) Summary() []EventCount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]EventCount, 0, len(s.counts))
	for eventType, count := range s.counts {
		results = append(results, EventCount{Type: eventType, Count: count})
	}
	return results
}

type EventRequest struct {
	Type string `json:"type"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()
	store := NewAnalyticsStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		store.Increment(req.Type)
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		respondJSON(w, http.StatusOK, store.Summary())
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "analytics"
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
