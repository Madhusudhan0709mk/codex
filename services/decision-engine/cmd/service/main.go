package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
)

type ScoreRequest struct {
	SkillMatch     float64 `json:"skill_match"`
	Experience     float64 `json:"experience"`
	Education      float64 `json:"education"`
	ReadinessBoost float64 `json:"readiness_boost"`
}

type ScoreResponse struct {
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := getServiceName()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler(serviceName))
	mux.HandleFunc("/readyz", readyHandler)
	mux.HandleFunc("/score", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req ScoreRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		score := (req.SkillMatch * 0.5) + (req.Experience * 0.3) + (req.Education * 0.1) + (req.ReadinessBoost * 0.1)
		score = math.Min(1.0, math.Max(0, score))
		explanation := "Score weighted by skills, experience, education, readiness."
		respondJSON(w, http.StatusOK, ScoreResponse{Score: score, Explanation: explanation})
	})

	startServer(serviceName, mux)
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "decision-engine"
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
