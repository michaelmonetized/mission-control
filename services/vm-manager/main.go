package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	// HTTP Routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/launch", launchHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/status", statusHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	log.Printf("VM Manager listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"service": "vm-manager",
		"uptime_seconds": 0, // TODO: track uptime
	})
}

// Launch VM endpoint
func launchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"userId"`
		RepoID   string `json:"repoId"`
		RepoName string `json:"repoName"`
		CloneURL string `json:"cloneUrl"`
		Branch   string `json:"branch"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Launching VM for user=%s, repo=%s", req.UserID, req.RepoName)

	// TODO: Call Fly.io API to create machine
	// TODO: Clone repo and inject Claude API key
	// TODO: Start Claude Code CLI

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vmId": "vm-pending-" + req.RepoID,
		"status": "starting",
		"createdAt": 0,
	})
}

// Stop VM endpoint
func stopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		VMId string `json:"vmId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Stopping VM: %s", req.VMId)

	// TODO: Call Fly.io API to delete machine
	// TODO: Cleanup volumes

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "stopping",
	})
}

// Status endpoint
func statusHandler(w http.ResponseWriter, r *http.Request) {
	vmID := r.URL.Query().Get("vmId")

	if vmID == "" {
		http.Error(w, "vmId required", http.StatusBadRequest)
		return
	}

	log.Printf("Getting status for VM: %s", vmID)

	// TODO: Call Fly.io API to get machine status

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vmId": vmID,
		"status": "running",
		"cpu_usage": 15.5,
		"memory_usage": 512,
		"uptime_seconds": 3600,
	})
}
