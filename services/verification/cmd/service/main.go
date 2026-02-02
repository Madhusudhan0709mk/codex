package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Verification struct {
	CandidateID string `json:"candidate_id"`
	Status      string `json:"status"`
	UpdatedAt   string `json:"updated_at"`
}

type VerificationStore struct {
	mu            sync.RWMutex
	verifications map[string]Verification
}

func NewVerificationStore() *VerificationStore {
	return &VerificationStore{verifications: make(map[string]Verification)}
}

func (s *VerificationStore) Upsert(ver Verification) Verification {
	s.mu.Lock()
	defer s.mu.Unlock()

	ver.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.verifications[ver.CandidateID] = ver
	return ver
}

func (s *VerificationStore) Get(candidateID string) (Verification, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ver, ok := s.verifications[candidateID]
	return ver, ok
}

type VerificationRequest struct {
	CandidateID string `json:"candidate_id"`
	Status      string `json:"status"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()
	store := NewVerificationStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req VerificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		status := strings.ToLower(req.Status)
		if status != "verified" && status != "unverified" {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
		ver := store.Upsert(Verification{CandidateID: req.CandidateID, Status: status})
		respondJSON(w, http.StatusOK, ver)
	})

	mux.HandleFunc("/verifications/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		candidateID := strings.TrimPrefix(r.URL.Path, "/verifications/")
		ver, ok := store.Get(candidateID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		respondJSON(w, http.StatusOK, ver)
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "verification"
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
