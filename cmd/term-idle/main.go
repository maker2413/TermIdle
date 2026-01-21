package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/maker2413/term-idle/internal/api"
	"github.com/maker2413/term-idle/internal/config"
	"github.com/maker2413/term-idle/internal/db"
	gamepkg "github.com/maker2413/term-idle/internal/game"
	"github.com/maker2413/term-idle/internal/ui"
)

func main() {
	var configPath = flag.String("config", "", "Path to configuration file")
	var migrate = flag.Bool("migrate", false, "Run database migrations and exit")
	var serverMode = flag.Bool("server", false, "Run in API server mode")
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

	// Check if we should run in server mode
	if *serverMode {
		runServerMode(cfg, database)
		return
	}

	// Run in normal TUI mode
	runTUIMode(cfg, database, username)
}

func runServerMode(cfg *config.Config, database *db.SQLiteDB) {
	// Parse port from string to int
	port := 8080 // default
	if cfg.Server.Port != "" {
		if p, err := strconv.Atoi(cfg.Server.Port); err == nil {
			port = p
		}
	}

	// Create API server configuration
	apiConfig := &api.Config{
		Port: port,
		Host: cfg.Server.Host,
	}

	// Create and start API server
	server := api.NewServer(database, apiConfig)

	// Set up graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting Term Idle API Server")
	log.Printf("Address: %s:%d", apiConfig.Host, apiConfig.Port)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	log.Println("API server started successfully")
	log.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	<-done
	log.Println("Shutting down API server...")
	log.Println("Server stopped")
}

func runTUIMode(cfg *config.Config, database *db.SQLiteDB, username *string) {
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
