package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/maker2413/term-idle/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDB implements the Database interface for testing
type MockDB struct {
	players     map[string]*db.Player
	leaderboard map[string]*db.LeaderboardEntry
}

func NewMockDB() *MockDB {
	return &MockDB{
		players:     make(map[string]*db.Player),
		leaderboard: make(map[string]*db.LeaderboardEntry),
	}
}

func (m *MockDB) GetLeaderboard(limit int) ([]*db.LeaderboardEntry, error) {
	entries := make([]*db.LeaderboardEntry, 0, len(m.leaderboard))
	for _, entry := range m.leaderboard {
		entries = append(entries, entry)
	}

	// Simple mock ranking by total keystrokes
	for i := range entries {
		entries[i].Rank = i + 1
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

func (m *MockDB) GetPlayerRank(playerID string) (int, error) {
	entry, exists := m.leaderboard[playerID]
	if !exists {
		return 0, assert.AnError
	}
	return entry.Rank, nil
}

func (m *MockDB) GetPlayer(id string) (*db.Player, error) {
	player, exists := m.players[id]
	if !exists {
		return nil, assert.AnError
	}
	return player, nil
}

func (m *MockDB) GetPlayerByUsername(username string) (*db.Player, error) {
	for _, player := range m.players {
		if player.Username == username {
			return player, nil
		}
	}
	return nil, assert.AnError
}

func (m *MockDB) UpdateLeaderboard(entry *db.LeaderboardEntry) error {
	m.leaderboard[entry.PlayerID] = entry
	return nil
}

func TestGetLeaderboard(t *testing.T) {
	mockDB := NewMockDB()

	// Add test data
	entry1 := &db.LeaderboardEntry{
		PlayerID:         "player1",
		Username:         "player1",
		KeystrokesPerSec: 10.0,
		TotalKeystrokes:  1000.0,
		Words:            10,
		Programs:         2,
		AIAutomations:    1,
		Level:            5,
		Rank:             1,
	}

	entry2 := &db.LeaderboardEntry{
		PlayerID:         "player2",
		Username:         "player2",
		KeystrokesPerSec: 8.0,
		TotalKeystrokes:  800.0,
		Words:            8,
		Programs:         1,
		AIAutomations:    0,
		Level:            4,
		Rank:             2,
	}

	_ = mockDB.UpdateLeaderboard(entry1)
	_ = mockDB.UpdateLeaderboard(entry2)

	// Create server with mock DB
	config := &Config{Port: 8080, Host: "localhost"}
	server := NewServer(mockDB, config)

	// Create request
	req := httptest.NewRequest("GET", "/api/leaderboard?limit=10", nil)
	w := httptest.NewRecorder()

	// Use the router directly
	router := mux.NewRouter()
	router.HandleFunc("/api/leaderboard", server.getLeaderboard).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "leaderboard")
	assert.Contains(t, response, "limit")
	assert.Contains(t, response, "total")

	leaderboard := response["leaderboard"].([]interface{})
	assert.Len(t, leaderboard, 2)
}

func TestGetPlayer(t *testing.T) {
	mockDB := NewMockDB()

	// Add test player
	player := &db.Player{
		ID:         "test-player",
		Username:   "testuser",
		SSHKey:     "ssh-rsa test-key",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	mockDB.players["test-player"] = player

	// Create server with mock DB
	config := &Config{Port: 8080, Host: "localhost"}
	server := NewServer(mockDB, config)

	// Create request
	req := httptest.NewRequest("GET", "/api/players/test-player", nil)
	w := httptest.NewRecorder()

	// Use the router directly
	router := mux.NewRouter()
	router.HandleFunc("/api/players/{playerID}", server.getPlayer).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response PlayerAPIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, player.ID, response.ID)
	assert.Equal(t, player.Username, response.Username)
}

func TestGetPlayerNotFound(t *testing.T) {
	mockDB := NewMockDB()

	// Create server with mock DB (empty)
	config := &Config{Port: 8080, Host: "localhost"}
	server := NewServer(mockDB, config)

	// Create request for non-existent player
	req := httptest.NewRequest("GET", "/api/players/non-existent", nil)
	w := httptest.NewRecorder()

	// Use the router directly
	router := mux.NewRouter()
	router.HandleFunc("/api/players/{playerID}", server.getPlayer).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Not Found", response.Error)
}

func TestUpdatePlayerLeaderboard(t *testing.T) {
	mockDB := NewMockDB()

	// Create server with mock DB
	config := &Config{Port: 8080, Host: "localhost"}
	server := NewServer(mockDB, config)

	// Create request body
	entry := LeaderboardAPIResponse{
		PlayerID:         "player1",
		Username:         "player1",
		KeystrokesPerSec: 15.0,
		TotalKeystrokes:  1500.0,
		Words:            15,
		Programs:         3,
		AIAutomations:    2,
		Level:            8,
	}

	body, _ := json.Marshal(entry)
	req := httptest.NewRequest("POST", "/api/players/player1/leaderboard", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Use the router directly
	router := mux.NewRouter()
	router.HandleFunc("/api/players/{playerID}/leaderboard", server.updatePlayerLeaderboard).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the entry was added to mock DB
	assert.Contains(t, mockDB.leaderboard, "player1")
	assert.Equal(t, 15.0, mockDB.leaderboard["player1"].KeystrokesPerSec)
}

func TestHealthCheck(t *testing.T) {
	mockDB := NewMockDB()

	// Create server with mock DB
	config := &Config{Port: 8080, Host: "localhost"}
	server := NewServer(mockDB, config)

	// Create request
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	// Use the router directly
	router := mux.NewRouter()
	router.HandleFunc("/api/health", server.healthCheck).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "term-idle-api", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
}
