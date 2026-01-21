package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/maker2413/term-idle/internal/config"
	"github.com/maker2413/term-idle/internal/db"
	gamepkg "github.com/maker2413/term-idle/internal/game"
	"github.com/maker2413/term-idle/internal/ui"
)

func main() {
	var configPath = flag.String("config", "", "Path to configuration file")
	var migrate = flag.Bool("migrate", false, "Run database migrations and exit")
	var username = flag.String("username", "Player", "Username for the player")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	database, err := db.NewSQLiteDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// If migrate flag is set, run migrations and exit
	if *migrate {
		if err := database.Migrate(); err != nil {
			log.Fatalf("Failed to run database migrations: %v", err)
		}
		fmt.Println("Database migrations completed successfully!")
		return
	}

	// Run migrations for normal operation as well
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Generate or get player ID
	playerID := uuid.New().String()

	// Create or load player
	player, err := database.GetPlayerByUsername(*username)
	if err != nil {
		// Create new player if not exists
		player = &db.Player{
			ID:         playerID,
			Username:   *username,
			SSHKey:     "local",
			CreatedAt:  time.Now(),
			LastActive: time.Now(),
		}
		if err := database.CreatePlayer(player); err != nil {
			log.Fatalf("Failed to create player: %v", err)
		}
	} else {
		playerID = player.ID
	}

	// Load or create game state
	dbGameState, err := database.LoadGameState(playerID)
	var gameState *gamepkg.GameState
	if err != nil {
		// Create new game state if not exists
		gameState = gamepkg.NewGameState(playerID)
	} else {
		// Convert db.GameState to game.GameState
		gameState = gamepkg.NewGameState(playerID)
		gameState.CurrentLevel = dbGameState.CurrentLevel
		gameState.Keystrokes = dbGameState.Keystrokes
		gameState.Words = dbGameState.Words
		gameState.Programs = dbGameState.Programs
		gameState.AIAutomations = dbGameState.AIAutomations
		gameState.StoryProgress = dbGameState.StoryProgress
		gameState.ProductionRate = dbGameState.ProductionRate
		gameState.KeystrokesPerSecond = dbGameState.KeystrokesPerSecond
		gameState.LastSave = dbGameState.LastSave
		gameState.LastUpdate = dbGameState.LastUpdate
	}

	// Initialize game components
	leaderboardService := gamepkg.NewLeaderboardService(database)

	// Update game state with current time
	gameState.LastUpdate = time.Now()
	gameState.UpdateProduction()

	// Create UI model with all services (database and leaderboard)
	model := ui.NewModelWithAll(gameState, database, leaderboardService, *username)

	// Initialize and start the Bubbletea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}

	log.Println("Game session ended")
}
