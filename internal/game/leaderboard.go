package game

import (
	"fmt"
	"time"

	"github.com/maker2413/term-idle/internal/db"
)

// LeaderboardService handles leaderboard operations
type LeaderboardService struct {
	db db.Database
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(database db.Database) *LeaderboardService {
	return &LeaderboardService{
		db: database,
	}
}

// UpdatePlayerLeaderboard updates a player's entry on the leaderboard
func (ls *LeaderboardService) UpdatePlayerLeaderboard(state *GameState, username string) error {
	// Convert to database state
	dbState := &db.GameState{
		PlayerID:            state.PlayerID,
		CurrentLevel:        state.CurrentLevel,
		Keystrokes:          state.Keystrokes,
		Words:               state.Words,
		Programs:            state.Programs,
		AIAutomations:       state.AIAutomations,
		StoryProgress:       state.StoryProgress,
		ProductionRate:      state.ProductionRate,
		KeystrokesPerSecond: state.KeystrokesPerSecond,
		LastSave:            state.LastSave,
		LastUpdate:          state.LastUpdate,
	}

	entry := &db.LeaderboardEntry{
		PlayerID:         dbState.PlayerID,
		Username:         username,
		KeystrokesPerSec: dbState.ProductionRate,
		TotalKeystrokes:  dbState.Keystrokes,
		Words:            dbState.Words,
		Programs:         dbState.Programs,
		AIAutomations:    dbState.AIAutomations,
		Level:            dbState.CurrentLevel,
		UpdatedAt:        time.Now(),
	}

	return ls.db.UpdateLeaderboard(entry)
}

// GetLeaderboardEntries retrieves the top entries from the leaderboard
func (ls *LeaderboardService) GetLeaderboardEntries(limit int) ([]*db.LeaderboardEntry, error) {
	return ls.db.GetLeaderboard(limit)
}

// GetPlayerRank gets a player's rank on the leaderboard
func (ls *LeaderboardService) GetPlayerRank(playerID string) (int, error) {
	return ls.db.GetPlayerRank(playerID)
}

// LeaderboardEntry represents a simplified leaderboard entry for display
type LeaderboardEntry struct {
	Rank             int
	Username         string
	KeystrokesPerSec float64
	TotalKeystrokes  float64
	Words            int
	Programs         int
	AIAutomations    int
	Level            int
}

// GetFormattedLeaderboard returns formatted leaderboard entries for display
func (ls *LeaderboardService) GetFormattedLeaderboard(limit int) ([]*LeaderboardEntry, error) {
	entries, err := ls.db.GetLeaderboard(limit)
	if err != nil {
		return nil, err
	}

	formatted := make([]*LeaderboardEntry, len(entries))
	for i, entry := range entries {
		formatted[i] = &LeaderboardEntry{
			Rank:             entry.Rank,
			Username:         entry.Username,
			KeystrokesPerSec: entry.KeystrokesPerSec,
			TotalKeystrokes:  entry.TotalKeystrokes,
			Words:            entry.Words,
			Programs:         entry.Programs,
			AIAutomations:    entry.AIAutomations,
			Level:            entry.Level,
		}
	}

	return formatted, nil
}

// GetPlayerLeaderboardPosition gets a player's position and surrounding players
func (ls *LeaderboardService) GetPlayerLeaderboardPosition(playerID string) ([]*LeaderboardEntry, error) {
	rank, err := ls.db.GetPlayerRank(playerID)
	if err != nil {
		return nil, err
	}

	// Get a wider range around the player (5 above, player, 5 below)
	lowerRank := rank - 5
	if lowerRank < 1 {
		lowerRank = 1
	}

	// Get top 50 entries and find our section
	allEntries, err := ls.db.GetLeaderboard(50)
	if err != nil {
		return nil, err
	}

	// Filter entries around our player's rank
	var result []*LeaderboardEntry
	for _, entry := range allEntries {
		if entry.Rank >= lowerRank && entry.Rank <= rank+5 {
			result = append(result, &LeaderboardEntry{
				Rank:             entry.Rank,
				Username:         entry.Username,
				KeystrokesPerSec: entry.KeystrokesPerSec,
				TotalKeystrokes:  entry.TotalKeystrokes,
				Words:            entry.Words,
				Programs:         entry.Programs,
				AIAutomations:    entry.AIAutomations,
				Level:            entry.Level,
			})
		}
	}

	return result, nil
}

// PeriodicUpdater handles periodic leaderboard updates
type PeriodicUpdater struct {
	service        *LeaderboardService
	gameState      *GameState
	username       string
	stopChan       chan bool
	updateInterval time.Duration
}

// NewPeriodicUpdater creates a new periodic leaderboard updater
func NewPeriodicUpdater(service *LeaderboardService, gameState *GameState, username string) *PeriodicUpdater {
	return &PeriodicUpdater{
		service:        service,
		gameState:      gameState,
		username:       username,
		stopChan:       make(chan bool),
		updateInterval: 30 * time.Second, // Update every 30 seconds
	}
}

// Start begins the periodic updates
func (pu *PeriodicUpdater) Start() {
	go func() {
		ticker := time.NewTicker(pu.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := pu.service.UpdatePlayerLeaderboard(pu.gameState, pu.username)
				if err != nil {
					// Log error but don't stop the updater
					fmt.Printf("Failed to update leaderboard: %v\n", err)
				}
			case <-pu.stopChan:
				return
			}
		}
	}()
}

// Stop stops the periodic updates
func (pu *PeriodicUpdater) Stop() {
	close(pu.stopChan)
}
