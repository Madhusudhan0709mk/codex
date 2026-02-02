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

type InterviewRequest struct {
	ID          string `json:"id"`
	RecruiterID string `json:"recruiter_id"`
	CandidateID string `json:"candidate_id"`
	Status      string `json:"status"`
	ExpiresAt   string `json:"expires_at"`
}

type RequestStore struct {
	mu       sync.RWMutex
	requests map[string]InterviewRequest
}

func NewRequestStore() *RequestStore {
	return &RequestStore{requests: make(map[string]InterviewRequest)}
}

func (s *RequestStore) Create(req InterviewRequest) InterviewRequest {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests[req.ID] = req
	return req
}

func (s *RequestStore) Get(id string) (InterviewRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	request, ok := s.requests[id]
	return request, ok
}

func (s *RequestStore) Update(id, status string) (InterviewRequest, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	request, ok := s.requests[id]
	if !ok {
		return InterviewRequest{}, false
	}
	request.Status = status
	s.requests[id] = request
	return request, true
}

type RequestCreate struct {
	RecruiterID   string `json:"recruiter_id"`
	CandidateID   string `json:"candidate_id"`
	ExpiresInDays int    `json:"expires_in_days"`
}

type RequestRespond struct {
	Status string `json:"status"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()
	store := NewRequestStore()
	chatURL := getEnv("CHAT_URL", "")
	client := &http.Client{Timeout: 3 * time.Second}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/requests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req RequestCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		expiresIn := req.ExpiresInDays
		if expiresIn <= 0 {
			expiresIn = 7
		}
		request := InterviewRequest{
			ID:          newID("req"),
			RecruiterID: req.RecruiterID,
			CandidateID: req.CandidateID,
			Status:      "pending",
			ExpiresAt:   time.Now().AddDate(0, 0, expiresIn).UTC().Format(time.RFC3339),
		}
		respondJSON(w, http.StatusCreated, store.Create(request))
	})

	mux.HandleFunc("/requests/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/requests/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		id := parts[0]
		if len(parts) == 1 {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			request, ok := store.Get(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			respondJSON(w, http.StatusOK, request)
			return
		}

		if len(parts) == 2 && parts[1] == "respond" {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var req RequestRespond
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			status := strings.ToLower(req.Status)
			if status != "confirmed" && status != "rejected" && status != "no_response" {
				http.Error(w, "invalid status", http.StatusBadRequest)
				return
			}
			request, ok := store.Update(id, status)
			if !ok {
				http.NotFound(w, r)
				return
			}
			if status == "confirmed" {
				openChatSession(client, chatURL, request)
			}
			respondJSON(w, http.StatusOK, request)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "recruiter-workflow"
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

func openChatSession(client *http.Client, chatURL string, request InterviewRequest) {
	if chatURL == "" {
		return
	}
	payload := map[string]string{
		"candidate_id": request.CandidateID,
		"recruiter_id": request.RecruiterID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("chat payload error: %v", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(chatURL, "/")+"/sessions", bytes.NewReader(body))
	if err != nil {
		log.Printf("chat request error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("chat call failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("chat call status %d", resp.StatusCode)
	}
}
