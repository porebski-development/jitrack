package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"jitrack/internal/db"
	"jitrack/internal/models"
)

// JitrackServer represents the HTTP server for Chrome extension
type JitrackServer struct {
	port     int
	token    string
	server   *http.Server
	db       *db.Database
	onImport func(models.Task)
}

// NewServer creates a new HTTP server
func NewServer(database *db.Database, onImport func(models.Task)) *JitrackServer {
	return &JitrackServer{
		port:     0, // Will be determined at startup
		token:    generateToken(),
		db:       database,
		onImport: onImport,
	}
}

// Start starts the HTTP server on an available port (8765-8769)
func (s *JitrackServer) Start() (int, error) {
	mux := http.NewServeMux()
	
	// Health check
	mux.HandleFunc("/health", s.handleHealth)
	
	// Config endpoint
	mux.HandleFunc("/api/config", s.handleConfig)
	
	// Import endpoint
	mux.HandleFunc("/api/import", s.handleImport)

	// Try ports 8765-8769
	for port := 8765; port <= 8769; port++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		s.server = &http.Server{
			Addr:    addr,
			Handler: s.corsMiddleware(mux),
		}
		
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			continue // Port in use, try next
		}
		listener.Close()
		
		s.port = port
		go func() {
			if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("Server error: %v\n", err)
			}
		}()
		
		return port, nil
	}
	
	return 0, fmt.Errorf("no available ports in range 8765-8769")
}

// Stop stops the HTTP server
func (s *JitrackServer) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

// GetPort returns the server port
func (s *JitrackServer) GetPort() int {
	return s.port
}

// GetToken returns the auth token
func (s *JitrackServer) GetToken() string {
	return s.token
}

// generateToken creates a random 32-character hex token
func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// corsMiddleware adds CORS headers
func (s *JitrackServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Jitrack-Token")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// handleHealth handles health check requests
func (s *JitrackServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleConfig returns the server config (token and port)
func (s *JitrackServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": s.token,
		"port":  s.port,
	})
}

// ImportRequest represents a task import request from Chrome extension
type ImportRequest struct {
	IssueKey       string   `json:"issue_key"`
	IssueKeyAlt    string   `json:"issueKey"`
	Summary        string   `json:"summary"`
	Project        string   `json:"project"`
	URL            string   `json:"url"`
	EstimatedHours *float64 `json:"estimated_hours"`
	EstHoursAlt    *float64 `json:"estimatedHours"`
}

// handleImport handles task import from Chrome extension
func (s *JitrackServer) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Verify auth token
	token := r.Header.Get("X-Jitrack-Token")
	if token != s.token {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid token",
		})
		return
	}
	
	// Parse request body
	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}
	
	// Handle both snake_case and camelCase
	issueKey := req.IssueKey
	if issueKey == "" {
		issueKey = req.IssueKeyAlt
	}
	
	estimatedHours := req.EstimatedHours
	if estimatedHours == nil {
		estimatedHours = req.EstHoursAlt
	}
	
	if issueKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Missing issue_key",
		})
		return
	}
	
	// Create task
	task := models.Task{
		IssueKey:       issueKey,
		Summary:        req.Summary,
		Project:        req.Project,
		URL:            req.URL,
		EstimatedHours: estimatedHours,
		IsCompleted:    false,
		ImportedAt:     time.Now(),
		LastUpdated:    time.Now(),
	}
	
	// Save to database
	if err := s.db.SaveTask(task); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	// Emit event to frontend
	if s.onImport != nil {
		s.onImport(task)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Imported %s", issueKey),
	})
}
