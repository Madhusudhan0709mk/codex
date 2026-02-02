package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
)

type CandidateIndex struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Skills          []string `json:"skills"`
	ReadinessStatus string   `json:"readiness_status"`
}

type IndexStore struct {
	mu    sync.RWMutex
	items map[string]CandidateIndex
}

func NewIndexStore() *IndexStore {
	return &IndexStore{items: make(map[string]CandidateIndex)}
}

func (s *IndexStore) Upsert(candidate CandidateIndex) {
	s.mu.Lock()
	defer s.mu.Unlock()
	candidate.ReadinessStatus = strings.ToLower(candidate.ReadinessStatus)
	s.items[candidate.ID] = candidate
}

func (s *IndexStore) Search(request SearchRequest) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	skills := make(map[string]struct{})
	for _, skill := range request.Skills {
		skills[strings.ToLower(skill)] = struct{}{}
	}

	results := make([]SearchResult, 0)
	for _, candidate := range s.items {
		if request.ReadinessStatus != "" && strings.ToLower(candidate.ReadinessStatus) != strings.ToLower(request.ReadinessStatus) {
			continue
		}
		score := 0
		for _, skill := range candidate.Skills {
			if _, ok := skills[strings.ToLower(skill)]; ok {
				score++
			}
		}

		if request.MinimumScore > 0 && score < request.MinimumScore {
			continue
		}

		results = append(results, SearchResult{Candidate: candidate, Score: score})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	return results
}

type SearchRequest struct {
	Skills         []string `json:"skills"`
	ReadinessStatus string   `json:"readiness_status"`
	MinimumScore   int      `json:"minimum_score"`
}

type SearchResult struct {
	Candidate CandidateIndex `json:"candidate"`
	Score     int            `json:"score"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()
	store := NewIndexStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var candidate CandidateIndex
		if err := json.NewDecoder(r.Body).Decode(&candidate); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if candidate.ID == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		store.Upsert(candidate)
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		respondJSON(w, http.StatusOK, store.Search(req))
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "recruiter-search"
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
