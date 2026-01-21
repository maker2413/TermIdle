package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maker2413/term-idle/internal/config"
	"github.com/maker2413/term-idle/internal/db"
	gamepkg "github.com/maker2413/term-idle/internal/game"
)

func TestDatabaseIntegration(t *testing.T) {
	// Create temporary database
	dbPath := ":memory:"
	database, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			t.Logf("Error closing database: %v", err)
		}
	}()

	// Test player creation
	playerID := uuid.New().String()
	player := &db.Player{
		ID:         playerID,
		Username:   "testuser",
		SSHKey:     "test_key",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	err = database.CreatePlayer(player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Test game state save/load
	gameState := gamepkg.NewGameState(playerID)
	gameState.CurrentLevel = 5
	gameState.Keystrokes = 1234.5

	// Convert to db.GameState for saving
	dbGameState := &db.GameState{
		PlayerID:            gameState.PlayerID,
		CurrentLevel:        gameState.CurrentLevel,
		Keystrokes:          gameState.Keystrokes,
		Words:               gameState.Words,
		Programs:            gameState.Programs,
		AIAutomations:       gameState.AIAutomations,
		StoryProgress:       gameState.StoryProgress,
		ProductionRate:      gameState.ProductionRate,
		KeystrokesPerSecond: gameState.KeystrokesPerSecond,
		LastSave:            gameState.LastSave,
		LastUpdate:          gameState.LastUpdate,
	}

	err = database.SaveGameState(dbGameState)
	if err != nil {
		t.Fatalf("Failed to save game state: %v", err)
	}

	// Load and verify
	loadedState, err := database.LoadGameState(playerID)
	if err != nil {
		t.Fatalf("Failed to load game state: %v", err)
	}

	if loadedState.CurrentLevel != 5 {
		t.Errorf("Expected level 5, got %d", loadedState.CurrentLevel)
	}

	if loadedState.Keystrokes != 1234.5 {
		t.Errorf("Expected keystrokes 1234.5, got %f", loadedState.Keystrokes)
	}
}

func TestConfigLoading(t *testing.T) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.SSH.Port != 2222 {
		t.Errorf("Expected SSH port 2222, got %d", cfg.SSH.Port)
	}

	if cfg.Database.Path != "./term_idle.db" {
		t.Errorf("Expected database path './term_idle.db', got %s", cfg.Database.Path)
	}
}
