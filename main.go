package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Metric represents our telemetry data model
type Metric struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`
	CPULoad   float64   `json:"cpu_load"`
	MemUsedMB int       `json:"mem_used_mb"`
	Timestamp time.Time `json:"timestamp"`
}

// Global in-memory storage with a Mutex for thread-safety (Goroutine protection)
var (
	metricsStore = []Metric{}
	storeMutex   sync.Mutex
)

// GET /metrics - Retrieve all stored metrics
func getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	storeMutex.Lock()
	defer storeMutex.Unlock()

	json.NewEncoder(w).Encode(metricsStore)
}

// POST /metrics - Create a new metric entry
func createMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newMetric Metric
	// Decode incoming JSON request body into our struct
	err := json.NewDecoder(r.Body).Decode(&newMetric)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Assign server-side metadata
	newMetric.Timestamp = time.Now()
	if newMetric.ID == "" {
		newMetric.ID = fmt.Sprintf("metric-%d", time.Now().UnixNano())
	}

	// Thread-safe append to slice
	storeMutex.Lock()
	metricsStore = append(metricsStore, newMetric)
	storeMutex.Unlock()

	// Return 201 Created and the created record as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newMetric)
}

func main() {
	// Seed initial data
	metricsStore = append(metricsStore, Metric{
		ID:        "metric-1",
		Host:      "prod-server-01",
		CPULoad:   42.5,
		MemUsedMB: 2048,
		Timestamp: time.Now(),
	})

	// Define API routes
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getMetricsHandler(w, r)
		case http.MethodPost:
			createMetricHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := ":8080"
	fmt.Printf("Server listening on http://localhost%s...\n", port)

	// Start HTTP server
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
