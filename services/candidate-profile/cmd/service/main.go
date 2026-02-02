package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Candidate struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Skills          []string `json:"skills"`
	ReadinessStatus string   `json:"readiness_status"`
	UpdatedAt       string   `json:"updated_at"`
}

type CandidateStore struct {
	mu         sync.RWMutex
	candidates map[string]Candidate
}

func NewCandidateStore() *CandidateStore {
	return &CandidateStore{candidates: make(map[string]Candidate)}
}

func (s *CandidateStore) List() []Candidate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]Candidate, 0, len(s.candidates))
	for _, candidate := range s.candidates {
		results = append(results, candidate)
	}

	return results
}

func (s *CandidateStore) Get(id string) (Candidate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	candidate, ok := s.candidates[id]
	return candidate, ok
}

func (s *CandidateStore) Upsert(candidate Candidate) Candidate {
	s.mu.Lock()
	defer s.mu.Unlock()

	candidate.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.candidates[candidate.ID] = candidate
	return candidate
}

type CandidateRequest struct {
	Name            string   `json:"name"`
	Skills          []string `json:"skills"`
	ReadinessStatus string   `json:"readiness_status"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()
	store := NewCandidateStore()
	searchURL := getEnv("SEARCH_URL", "")
	client := &http.Client{Timeout: 3 * time.Second}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/candidates", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, http.StatusOK, store.List())
			case http.MethodPost:
				var req CandidateRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "invalid payload", http.StatusBadRequest)
					return
				}
				candidate := Candidate{
					ID:              newID("cand"),
					Name:            req.Name,
					Skills:          req.Skills,
					ReadinessStatus: normalizeReadiness(req.ReadinessStatus),
				}
				created := store.Upsert(candidate)
				indexCandidate(client, searchURL, created)
				respondJSON(w, http.StatusCreated, created)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})

	mux.HandleFunc("/candidates/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/candidates/")
		if id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			candidate, ok := store.Get(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			respondJSON(w, http.StatusOK, candidate)
			case http.MethodPut:
				var req CandidateRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "invalid payload", http.StatusBadRequest)
					return
				}
				candidate := Candidate{
					ID:              id,
					Name:            req.Name,
					Skills:          req.Skills,
					ReadinessStatus: normalizeReadiness(req.ReadinessStatus),
				}
				updated := store.Upsert(candidate)
				indexCandidate(client, searchURL, updated)
				respondJSON(w, http.StatusOK, updated)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "candidate-profile"
	}
	return serviceName
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
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

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func normalizeReadiness(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "verified", "interview-ready", "ready":
		return "verified"
	case "unverified", "not-ready", "not interview-ready":
		return "unverified"
	default:
		return "unverified"
	}
}

func indexCandidate(client *http.Client, searchURL string, candidate Candidate) {
	if searchURL == "" {
		return
	}
	payload := map[string]any{
		"id":               candidate.ID,
		"name":             candidate.Name,
		"skills":           candidate.Skills,
		"readiness_status": candidate.ReadinessStatus,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("index payload error: %v", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(searchURL, "/")+"/index", bytes.NewReader(body))
	if err != nil {
		log.Printf("index request error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("index call failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("index call status %d", resp.StatusCode)
	}
}
