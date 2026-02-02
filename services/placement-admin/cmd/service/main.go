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

type Student struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	College         string `json:"college"`
	PlacementStatus string `json:"placement_status"`
}

type StudentStore struct {
	mu       sync.RWMutex
	students map[string]Student
}

func NewStudentStore() *StudentStore {
	return &StudentStore{students: make(map[string]Student)}
}

func (s *StudentStore) Create(student Student) Student {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.students[student.ID] = student
	return student
}

func (s *StudentStore) Get(id string) (Student, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	student, ok := s.students[id]
	return student, ok
}

func (s *StudentStore) List() []Student {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]Student, 0, len(s.students))
	for _, student := range s.students {
		results = append(results, student)
	}
	return results
}

type StudentRequest struct {
	Name            string `json:"name"`
	College         string `json:"college"`
	PlacementStatus string `json:"placement_status"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()
	store := NewStudentStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)

	mux.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, http.StatusOK, store.List())
		case http.MethodPost:
			var req StudentRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			student := Student{
				ID:              newID("student"),
				Name:            req.Name,
				College:         req.College,
				PlacementStatus: strings.ToLower(req.PlacementStatus),
			}
			respondJSON(w, http.StatusCreated, store.Create(student))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/students/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/students/")
		student, ok := store.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		respondJSON(w, http.StatusOK, student)
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "placement-admin"
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
