package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type ChatMessage struct {
	SenderID string `json:"sender_id"`
	Text     string `json:"text"`
	SentAt   string `json:"sent_at"`
}

type ChatSession struct {
	ID          string        `json:"id"`
	CandidateID string        `json:"candidate_id"`
	RecruiterID string        `json:"recruiter_id"`
	Messages    []ChatMessage `json:"messages"`
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]ChatSession
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]ChatSession)}
}

func (s *SessionStore) Create(session ChatSession) ChatSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return session
}

func (s *SessionStore) Get(id string) (ChatSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	return session, ok
}

func (s *SessionStore) AddMessage(id string, message ChatMessage) (ChatSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return ChatSession{}, false
	}
	session.Messages = append(session.Messages, message)
	s.sessions[id] = session
	return session, true
}

type SessionRequest struct {
	CandidateID string `json:"candidate_id"`
	RecruiterID string `json:"recruiter_id"`
}

type MessageRequest struct {
	SenderID string `json:"sender_id"`
	Text     string `json:"text"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()
	store := NewSessionStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req SessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		session := ChatSession{ID: newID("chat"), CandidateID: req.CandidateID, RecruiterID: req.RecruiterID}
		respondJSON(w, http.StatusCreated, store.Create(session))
	})

	mux.HandleFunc("/sessions/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/sessions/")
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
			session, ok := store.Get(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			respondJSON(w, http.StatusOK, session)
			return
		}
		if len(parts) == 2 && parts[1] == "messages" {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var req MessageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			message := ChatMessage{SenderID: req.SenderID, Text: req.Text, SentAt: time.Now().UTC().Format(time.RFC3339)}
			session, ok := store.AddMessage(id, message)
			if !ok {
				http.NotFound(w, r)
				return
			}
			respondJSON(w, http.StatusOK, session)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "chat"
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

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
