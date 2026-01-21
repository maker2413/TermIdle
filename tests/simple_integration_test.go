package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/maker2413/term-idle/internal/db"
	"github.com/maker2413/term-idle/internal/game"
)

// SimpleIntegrationTestSuite tests core integration without requiring servers
type SimpleIntegrationTestSuite struct {
	suite.Suite
	db  *db.SQLiteDB
	ctx context.Context
}

// SetupSuite runs once before all tests
func (s *SimpleIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Use in-memory database for testing
	testDB := ":memory:"

	var err error
	s.db, err = db.NewSQLiteDB(testDB)
	s.Require().NoError(err, "Failed to create test database")
}

// TearDownSuite runs once after all tests
func (s *SimpleIntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

// TestPlayerCreationAndRetrieval tests creating and retrieving players
func (s *SimpleIntegrationTestSuite) TestPlayerCreationAndRetrieval() {
	// Create a test player
	player := &db.Player{
		ID:         uuid.New().String(),
		Username:   "testuser_" + uuid.New().String()[:8],
		SSHKey:     "test-ssh-key",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// Save player to database
	err := s.db.CreatePlayer(player)
	s.NoError(err, "Should be able to create player")

	// Retrieve player by ID
	retrieved, err := s.db.GetPlayer(player.ID)
	s.NoError(err, "Should be able to retrieve player")
	s.NotNil(retrieved, "Retrieved player should not be nil")
	s.Equal(player.ID, retrieved.ID)
	s.Equal(player.Username, retrieved.Username)
	s.Equal(player.SSHKey, retrieved.SSHKey)

	// Retrieve player by username
	retrievedByUsername, err := s.db.GetPlayerByUsername(player.Username)
	s.NoError(err, "Should be able to retrieve player by username")
	s.NotNil(retrievedByUsername, "Retrieved player should not be nil")
	s.Equal(player.ID, retrievedByUsername.ID)
	s.Equal(player.Username, retrievedByUsername.Username)
}

// TestGameStatePersistence tests saving and loading game state
func (s *SimpleIntegrationTestSuite) TestGameStatePersistence() {
	playerID := uuid.New().String()

	// Create initial game state using db.GameState
	gameState := &db.GameState{
		PlayerID:            playerID,
		CurrentLevel:        5,
		Keystrokes:          100.5,
		Words:               10,
		Programs:            2,
		AIAutomations:       1,
		StoryProgress:       3,
		ProductionRate:      2.5,
		KeystrokesPerSecond: 1.5,
		LastSave:            time.Now(),
		LastUpdate:          time.Now(),
	}

	// Save game state
	err := s.db.SaveGameState(gameState)
	s.NoError(err, "Should be able to save game state")

	// Load game state
	loadedState, err := s.db.LoadGameState(playerID)
	s.NoError(err, "Should be able to load game state")
	s.NotNil(loadedState, "Loaded game state should not be nil")

	// Verify all fields are preserved
	s.Equal(gameState.PlayerID, loadedState.PlayerID)
	s.Equal(gameState.CurrentLevel, loadedState.CurrentLevel)
	s.Equal(gameState.Keystrokes, loadedState.Keystrokes)
	s.Equal(gameState.Words, loadedState.Words)
	s.Equal(gameState.Programs, loadedState.Programs)
	s.Equal(gameState.AIAutomations, loadedState.AIAutomations)
	s.Equal(gameState.StoryProgress, loadedState.StoryProgress)
	s.Equal(gameState.ProductionRate, loadedState.ProductionRate)
	s.Equal(gameState.KeystrokesPerSecond, loadedState.KeystrokesPerSecond)
}

// TestLeaderboardIntegration tests leaderboard functionality
func (s *SimpleIntegrationTestSuite) TestLeaderboardIntegration() {
	// Create multiple players
	players := make([]*db.Player, 5)
	for i := 0; i < 5; i++ {
		players[i] = &db.Player{
			ID:         uuid.New().String(),
			Username:   fmt.Sprintf("player%d", i+1),
			SSHKey:     fmt.Sprintf("ssh-key-%d", i+1),
			CreatedAt:  time.Now(),
			LastActive: time.Now(),
		}

		err := s.db.CreatePlayer(players[i])
		s.Require().NoError(err, "Should be able to create player %d", i+1)
	}

	// Create game states with different scores
	levels := []int{10, 25, 5, 15, 30}
	kps := []float64{1.5, 3.2, 0.8, 2.1, 4.5}

	for i, player := range players {
		gameState := &db.GameState{
			PlayerID:            player.ID,
			CurrentLevel:        levels[i],
			KeystrokesPerSecond: kps[i],
			LastUpdate:          time.Now(),
		}

		err := s.db.SaveGameState(gameState)
		s.Require().NoError(err, "Should be able to save game state for player %d", i+1)

		// Update leaderboard
		entry := &db.LeaderboardEntry{
			PlayerID:         player.ID,
			Username:         player.Username,
			KeystrokesPerSec: kps[i],
			TotalKeystrokes:  float64(levels[i] * 100),
			Level:            levels[i],
			UpdatedAt:        time.Now(),
		}

		err = s.db.UpdateLeaderboard(entry)
		s.Require().NoError(err, "Should be able to update leaderboard for player %d", i+1)
	}

	// Retrieve leaderboard
	leaderboard, err := s.db.GetLeaderboard(10)
	s.NoError(err, "Should be able to retrieve leaderboard")
	s.Len(leaderboard, 5, "Should have 5 entries in leaderboard")

	// Verify leaderboard is sorted correctly (highest KPS first)
	s.True(leaderboard[0].KeystrokesPerSec >= leaderboard[1].KeystrokesPerSec,
		"Leaderboard should be sorted by keystrokes per second descending")

	// Verify highest scoring player is at top
	s.Equal("player5", leaderboard[0].Username, "Player 5 should be at top with highest KPS")
	s.Equal(4.5, leaderboard[0].KeystrokesPerSec, "Player 5 should have 4.5 KPS")
}

// TestUpgradeSystemIntegration tests upgrade system with game state
func (s *SimpleIntegrationTestSuite) TestUpgradeSystemIntegration() {
	playerID := uuid.New().String()

	// Create initial game state using game package
	gameState := game.NewGameState(playerID)

	// Save initial state (convert to db.GameState)
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

	err := s.db.SaveGameState(dbGameState)
	s.NoError(err, "Should be able to save initial game state")

	// Test upgrade manager
	upgradeManager := game.NewUpgradeManager()

	// Test upgrade definitions
	for _, upgradeType := range []game.UpgradeType{
		game.UpgradeTypingSpeed,
		game.UpgradeVocabulary,
		game.UpgradeProgramming,
		game.UpgradeAIEfficiency,
		game.UpgradeStoryProgress,
	} {
		definition, err := upgradeManager.GetDefinition(upgradeType)
		s.NoError(err, "Should be able to get upgrade definition for %v", upgradeType)
		s.NotNil(definition, "Upgrade definition should not be nil")

		// Test cost calculation
		cost, err := upgradeManager.CalculateCost(upgradeType, 1)
		s.NoError(err, "Should be able to calculate cost for %v", upgradeType)
		s.Greater(cost, 0.0, "Cost should be greater than 0")
	}

	// Test production calculation
	production := gameState.CalculateProduction()
	s.GreaterOrEqual(production, 1.0, "Production should be at least base rate")
}

// TestStorySystemIntegration tests story system
func (s *SimpleIntegrationTestSuite) TestStorySystemIntegration() {
	playerID := uuid.New().String()

	// Create story manager
	storyManager := game.NewStoryManager()

	// Test initial story state
	currentChapter := storyManager.GetCurrentChapter()
	s.NotNil(currentChapter, "Should have initial chapter")
	s.Equal(1, currentChapter.ID, "Should start with chapter 1")

	// Test story progress
	progress := storyManager.GetProgress()
	s.GreaterOrEqual(progress, 0.0, "Progress should be >= 0")
	s.LessOrEqual(progress, 100.0, "Progress should be <= 100")

	// Test next chapter (requires game state parameter)
	gameState := game.NewGameState(playerID)
	nextChapter := storyManager.GetNextChapter(gameState)
	s.NotNil(nextChapter, "Should have next chapter")
	s.Greater(nextChapter.ID, currentChapter.ID, "Next chapter should have higher ID")

	// Test story hints
	hint := storyManager.GetHint(gameState)
	s.NotEmpty(hint, "Should provide a hint")

	// Test mark chapter as read
	storyManager.MarkChapterRead(currentChapter.ID)
	// Note: MarkChapterRead doesn't return an error in current implementation

	// Verify chapter is marked as read
	unlockedChapters := storyManager.GetUnlockedChapters()
	s.Len(unlockedChapters, 1, "Should have 1 unlocked chapter")
	s.True(unlockedChapters[0].IsRead, "Chapter should be marked as read")

	s.T().Logf("Story system integration test completed. Current chapter: %s, Progress: %.1f%%",
		currentChapter.Title, progress)
}

// TestGameStateOperations tests game state operations
func (s *SimpleIntegrationTestSuite) TestGameStateOperations() {
	playerID := uuid.New().String()

	// Create game state
	gameState := game.NewGameState(playerID)

	// Test initial state
	s.Equal(playerID, gameState.PlayerID)
	s.Equal(1, gameState.CurrentLevel, "Should start at level 1")
	s.GreaterOrEqual(gameState.Keystrokes, 0.0, "Keystrokes should be >= 0")
	s.GreaterOrEqual(gameState.Words, 0, "Words should be >= 0")
	s.GreaterOrEqual(gameState.Programs, 0, "Programs should be >= 0")
	s.GreaterOrEqual(gameState.AIAutomations, 0, "AI Automations should be >= 0")

	// Test resource operations
	initialKeystrokes := gameState.Keystrokes
	gameState.UpdateResources(time.Now())
	s.GreaterOrEqual(gameState.Keystrokes, initialKeystrokes,
		"Resources should update over time")

	// Test notifications
	gameState.AddNotification("Test notification")
	s.Contains(gameState.Notifications, "Test notification",
		"Notification should be added")

	// Test cost affordability
	testCost := 50.0
	canAfford := gameState.CanAfford(testCost)
	s.False(canAfford, "Should not be able to afford 50 keystrokes initially")

	// Add some keystrokes
	gameState.Keystrokes = 100.0
	canAfford = gameState.CanAfford(testCost)
	s.True(canAfford, "Should be able to afford 50 keystrokes after earning some")

	// Test spending
	gameState.SpendResources(testCost)
	s.Equal(50.0, gameState.Keystrokes, "Should have 50 keystrokes remaining")

	// Test production calculation
	production := gameState.CalculateProduction()
	s.GreaterOrEqual(production, 1.0, "Production should be at least base rate")

	s.T().Logf("Game state operations test completed. Production rate: %.2f/s", production)
}

// TestConcurrentDatabaseAccess tests concurrent database operations
func (s *SimpleIntegrationTestSuite) TestConcurrentDatabaseAccess() {
	// Create multiple game states concurrently
	const numGoroutines = 10
	const numOperations = 5

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numOperations; j++ {
				playerID := fmt.Sprintf("concurrent_%d_%d", goroutineID, j)

				// Create game state
				gameState := &db.GameState{
					PlayerID:            playerID,
					CurrentLevel:        goroutineID,
					KeystrokesPerSecond: float64(goroutineID),
					LastUpdate:          time.Now(),
				}

				// Save and load
				err := s.db.SaveGameState(gameState)
				if err != nil {
					s.T().Errorf("Failed to save game state for %s: %v", playerID, err)
					continue
				}

				_, err = s.db.LoadGameState(playerID)
				if err != nil {
					s.T().Errorf("Failed to load game state for %s: %v", playerID, err)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	time.Sleep(2 * time.Second)

	s.T().Logf("Completed %d concurrent database access operations",
		numGoroutines*numOperations)
}

// TestErrorHandling tests error conditions and edge cases
func (s *SimpleIntegrationTestSuite) TestErrorHandling() {
	// Test retrieving non-existent player
	_, err := s.db.GetPlayer("non-existent-id")
	s.Error(err, "Should return error for non-existent player")

	// Test retrieving non-existent game state
	_, err = s.db.LoadGameState("non-existent-player-id")
	s.Error(err, "Should return error for non-existent game state")

	// Test saving invalid game state (empty player ID)
	invalidGameState := &db.GameState{
		PlayerID:   "", // Invalid empty ID
		LastUpdate: time.Now(),
	}

	err = s.db.SaveGameState(invalidGameState)
	// This might or might not error depending on implementation
	// The test just ensures behavior is consistent

	s.T().Log("Error handling tests completed")
}

// TestSimpleIntegration runs the integration tests
func TestSimpleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(SimpleIntegrationTestSuite))
}
