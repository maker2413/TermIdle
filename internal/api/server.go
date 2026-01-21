package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/mux"
	"github.com/maker2413/term-idle/internal/db"
)

// Server represents the HTTP API server
type Server struct {
	db     Database
	router *mux.Router
	config *Config
}

// Database interface for the API server (narrower than the full db.Database)
type Database interface {
	GetLeaderboard(limit int) ([]*db.LeaderboardEntry, error)
	GetPlayerRank(playerID string) (int, error)
	GetPlayer(id string) (*db.Player, error)
	GetPlayerByUsername(username string) (*db.Player, error)
	UpdateLeaderboard(entry *db.LeaderboardEntry) error
}

// Config represents the API server configuration
type Config struct {
	Port int    `koanf:"port"`
	Host string `koanf:"host"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NewServer creates a new API server
func NewServer(database Database, config *Config) *Server {
	server := &Server{
		db:     database,
		router: mux.NewRouter(),
		config: config,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api").Subrouter()

	// Leaderboard endpoints
	api.HandleFunc("/leaderboard", s.getLeaderboard).Methods("GET")
	api.HandleFunc("/leaderboard/player/{playerID}", s.getPlayerRank).Methods("GET")

	// Player endpoints
	api.HandleFunc("/players/{playerID}", s.getPlayer).Methods("GET")
	api.HandleFunc("/players/username/{username}", s.getPlayerByUsername).Methods("GET")
	api.HandleFunc("/players/{playerID}/leaderboard", s.updatePlayerLeaderboard).Methods("POST")

	// Health check
	api.HandleFunc("/health", s.healthCheck).Methods("GET")

	// Add middleware
	api.Use(s.corsMiddleware)
	api.Use(s.loggingMiddleware)
}

// LeaderboardAPIResponse represents the API response structure
type LeaderboardAPIResponse struct {
	PlayerID         string  `json:"player_id"`
	Username         string  `json:"username"`
	KeystrokesPerSec float64 `json:"keystrokes_per_sec"`
	TotalKeystrokes  float64 `json:"total_keystrokes"`
	Words            int     `json:"words"`
	Programs         int     `json:"programs"`
	AIAutomations    int     `json:"ai_automations"`
	Level            int     `json:"level"`
	Rank             int     `json:"rank"`
	UpdatedAt        string  `json:"updated_at"`
}

// PlayerAPIResponse represents the player API response structure
type PlayerAPIResponse struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	CreatedAt  string `json:"created_at"`
	LastActive string `json:"last_active"`
}

// getLeaderboard handles GET /api/leaderboard
func (s *Server) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter with default of 50
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	entries, err := s.db.GetLeaderboard(limit)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "failed to get leaderboard", err)
		return
	}

	// Convert to API format
	apiEntries := make([]*LeaderboardAPIResponse, len(entries))
	for i, entry := range entries {
		apiEntries[i] = &LeaderboardAPIResponse{
			PlayerID:         entry.PlayerID,
			Username:         entry.Username,
			KeystrokesPerSec: entry.KeystrokesPerSec,
			TotalKeystrokes:  entry.TotalKeystrokes,
			Words:            entry.Words,
			Programs:         entry.Programs,
			AIAutomations:    entry.AIAutomations,
			Level:            entry.Level,
			Rank:             entry.Rank,
			UpdatedAt:        entry.UpdatedAt.Format(time.RFC3339),
		}
	}

	response := map[string]interface{}{
		"leaderboard": apiEntries,
		"limit":       limit,
		"total":       len(apiEntries),
	}

	s.writeJSONResponse(w, http.StatusOK, response)
}

// getPlayerRank handles GET /api/leaderboard/player/{playerID}
func (s *Server) getPlayerRank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerID := vars["playerID"]

	if playerID == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "player ID is required", nil)
		return
	}

	rank, err := s.db.GetPlayerRank(playerID)
	if err != nil {
		s.writeErrorResponse(w, http.StatusNotFound, "player not found", err)
		return
	}

	response := map[string]interface{}{
		"player_id": playerID,
		"rank":      rank,
	}

	s.writeJSONResponse(w, http.StatusOK, response)
}

// getPlayer handles GET /api/players/{playerID}
func (s *Server) getPlayer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerID := vars["playerID"]

	if playerID == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "player ID is required", nil)
		return
	}

	player, err := s.db.GetPlayer(playerID)
	if err != nil {
		s.writeErrorResponse(w, http.StatusNotFound, "player not found", err)
		return
	}

	apiPlayer := &PlayerAPIResponse{
		ID:         player.ID,
		Username:   player.Username,
		CreatedAt:  player.CreatedAt.Format(time.RFC3339),
		LastActive: player.LastActive.Format(time.RFC3339),
	}

	s.writeJSONResponse(w, http.StatusOK, apiPlayer)
}

// getPlayerByUsername handles GET /api/players/username/{username}
func (s *Server) getPlayerByUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	if username == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "username is required", nil)
		return
	}

	player, err := s.db.GetPlayerByUsername(username)
	if err != nil {
		s.writeErrorResponse(w, http.StatusNotFound, "player not found", err)
		return
	}

	apiPlayer := &PlayerAPIResponse{
		ID:         player.ID,
		Username:   player.Username,
		CreatedAt:  player.CreatedAt.Format(time.RFC3339),
		LastActive: player.LastActive.Format(time.RFC3339),
	}

	s.writeJSONResponse(w, http.StatusOK, apiPlayer)
}

// updatePlayerLeaderboard handles POST /api/players/{playerID}/leaderboard
func (s *Server) updatePlayerLeaderboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerID := vars["playerID"]

	if playerID == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "player ID is required", nil)
		return
	}

	var entry LeaderboardAPIResponse
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "invalid JSON", err)
		return
	}

	// Validate entry
	if entry.PlayerID != playerID {
		s.writeErrorResponse(w, http.StatusBadRequest, "player ID mismatch", nil)
		return
	}

	if entry.Username == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "username is required", nil)
		return
	}

	// Convert to database format
	dbEntry := &db.LeaderboardEntry{
		PlayerID:         entry.PlayerID,
		Username:         entry.Username,
		KeystrokesPerSec: entry.KeystrokesPerSec,
		TotalKeystrokes:  entry.TotalKeystrokes,
		Words:            entry.Words,
		Programs:         entry.Programs,
		AIAutomations:    entry.AIAutomations,
		Level:            entry.Level,
		UpdatedAt:        time.Now(),
	}

	if err := s.db.UpdateLeaderboard(dbEntry); err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "failed to update leaderboard", err)
		return
	}

	response := map[string]interface{}{
		"message": "leaderboard entry updated successfully",
		"entry":   entry,
	}

	s.writeJSONResponse(w, http.StatusOK, response)
}

// healthCheck handles GET /api/health
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "term-idle-api",
		"version": "1.0.0",
	}

	s.writeJSONResponse(w, http.StatusOK, response)
}

// Start starts the API server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Infof("Starting API server on %s", addr)

	return http.ListenAndServe(addr, s.router)
}

// writeJSONResponse writes a JSON response
func (s *Server) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Errorf("Failed to encode JSON response: %v", err)
	}
}

// writeErrorResponse writes an error response
func (s *Server) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	if err != nil {
		log.Errorf("API error: %s: %v", message, err)
	}

	s.writeJSONResponse(w, statusCode, errorResp)
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("API request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
