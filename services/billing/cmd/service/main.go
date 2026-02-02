package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Plan struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type Subscription struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	PlanID    string `json:"plan_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type SubscriptionStore struct {
	mu            sync.RWMutex
	subscriptions map[string]Subscription
}

func NewSubscriptionStore() *SubscriptionStore {
	return &SubscriptionStore{subscriptions: make(map[string]Subscription)}
}

func (s *SubscriptionStore) Create(sub Subscription) Subscription {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscriptions[sub.ID] = sub
	return sub
}

type SubscribeRequest struct {
	UserID string `json:"user_id"`
	PlanID string `json:"plan_id"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

var plans = []Plan{
	{ID: "starter", Name: "Starter", Price: 0},
	{ID: "pro", Name: "Pro", Price: 4999},
	{ID: "enterprise", Name: "Enterprise", Price: 19999},
}

func main() {
	serviceName := getServiceName()
	store := NewSubscriptionStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/plans", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		respondJSON(w, http.StatusOK, plans)
	})

	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req SubscribeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		subscription := Subscription{
			ID:        newID("sub"),
			UserID:    req.UserID,
			PlanID:    req.PlanID,
			Status:    "active",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		respondJSON(w, http.StatusCreated, store.Create(subscription))
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "billing"
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
