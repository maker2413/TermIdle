package db

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteDB_CreatePlayer(t *testing.T) {
	// Create temporary database
	dbPath := ":memory:"
	sqliteDB, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer sqliteDB.Close()

	// Create a player
	player := &Player{
		ID:         "test-player-1",
		Username:   "testuser1",
		SSHKey:     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// Save player
	err = sqliteDB.CreatePlayer(player)
	assert.NoError(t, err)

	// Retrieve player
	retrieved, err := sqliteDB.GetPlayer(player.ID)
	assert.NoError(t, err)
	assert.Equal(t, player.ID, retrieved.ID)
	assert.Equal(t, player.Username, retrieved.Username)
	assert.Equal(t, player.SSHKey, retrieved.SSHKey)
}

func TestSQLiteDB_GetPlayerByUsername(t *testing.T) {
	// Create temporary database
	dbPath := ":memory:"
	sqliteDB, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer sqliteDB.Close()

	// Create a player
	player := &Player{
		ID:         "test-player-2",
		Username:   "testuser2",
		SSHKey:     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// Save player
	err = sqliteDB.CreatePlayer(player)
	require.NoError(t, err)

	// Retrieve by username
	retrieved, err := sqliteDB.GetPlayerByUsername(player.Username)
	assert.NoError(t, err)
	assert.Equal(t, player.Username, retrieved.Username)
	assert.Equal(t, player.ID, retrieved.ID)
}

func TestSQLiteDB_SaveGameState(t *testing.T) {
	// Create temporary database
	dbPath := ":memory:"
	sqliteDB, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer sqliteDB.Close()

	// Create a player first
	player := &Player{
		ID:         "test-player-3",
		Username:   "testuser3",
		SSHKey:     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	err = sqliteDB.CreatePlayer(player)
	require.NoError(t, err)

	// Create game state
	gameState := &GameState{
		PlayerID:            player.ID,
		CurrentLevel:        5,
		Keystrokes:          1234.5,
		Words:               10,
		Programs:            3,
		AIAutomations:       1,
		StoryProgress:       2,
		ProductionRate:      15.5,
		KeystrokesPerSecond: 1.5,
		LastSave:            time.Now(),
		LastUpdate:          time.Now(),
	}

	// Save game state
	err = sqliteDB.SaveGameState(gameState)
	assert.NoError(t, err)

	// Retrieve game state
	retrieved, err := sqliteDB.LoadGameState(player.ID)
	assert.NoError(t, err)
	assert.Equal(t, gameState.PlayerID, retrieved.PlayerID)
	assert.Equal(t, gameState.CurrentLevel, retrieved.CurrentLevel)
	assert.Equal(t, gameState.Keystrokes, retrieved.Keystrokes)
	assert.Equal(t, gameState.Words, retrieved.Words)
	assert.Equal(t, gameState.Programs, retrieved.Programs)
	assert.Equal(t, gameState.AIAutomations, retrieved.AIAutomations)
}

func TestSQLiteDB_LeaderboardOperations(t *testing.T) {
	// Create temporary database
	dbPath := ":memory:"
	sqliteDB, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer sqliteDB.Close()

	// Create players
	player1 := &Player{
		ID:         "test-player-4",
		Username:   "player1",
		SSHKey:     "test-key-1",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	player2 := &Player{
		ID:         "test-player-5",
		Username:   "player2",
		SSHKey:     "test-key-2",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	err = sqliteDB.CreatePlayer(player1)
	require.NoError(t, err)
	err = sqliteDB.CreatePlayer(player2)
	require.NoError(t, err)

	// Create leaderboard entries
	entry1 := &LeaderboardEntry{
		PlayerID:         player1.ID,
		Username:         player1.Username,
		KeystrokesPerSec: 10.5,
		TotalKeystrokes:  5000.0,
		Words:            50,
		Programs:         5,
		AIAutomations:    1,
		Level:            10,
		UpdatedAt:        time.Now(),
	}

	entry2 := &LeaderboardEntry{
		PlayerID:         player2.ID,
		Username:         player2.Username,
		KeystrokesPerSec: 8.0,
		TotalKeystrokes:  3000.0,
		Words:            30,
		Programs:         3,
		AIAutomations:    0,
		Level:            8,
		UpdatedAt:        time.Now(),
	}

	// Save entries
	err = sqliteDB.UpdateLeaderboard(entry1)
	assert.NoError(t, err)
	err = sqliteDB.UpdateLeaderboard(entry2)
	assert.NoError(t, err)

	// Get leaderboard
	entries, err := sqliteDB.GetLeaderboard(10)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	// Should be sorted by total_keystrokes DESC
	assert.Equal(t, player1.Username, entries[0].Username)
	assert.Equal(t, player2.Username, entries[1].Username)
	assert.Equal(t, 1, entries[0].Rank)
	assert.Equal(t, 2, entries[1].Rank)

	// Test player rank
	rank, err := sqliteDB.GetPlayerRank(player1.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, rank)
}

func TestSQLiteDB_Persistence(t *testing.T) {
	// Create temporary file database
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	dbPath := tmpFile.Name()
	sqliteDB, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	sqliteDB.Close()

	// Reopen database
	sqliteDB, err = NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer sqliteDB.Close()

	// Should be able to use it without errors
	player := &Player{
		ID:         "test-player-persistent",
		Username:   "persistent_user",
		SSHKey:     "ssh-rsa test-key",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	err = sqliteDB.CreatePlayer(player)
	assert.NoError(t, err)

	retrieved, err := sqliteDB.GetPlayer(player.ID)
	assert.NoError(t, err)
	assert.Equal(t, player.Username, retrieved.Username)
}
