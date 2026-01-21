package game

import (
	"fmt"
	"time"
)

// GameState represents the current state of a player's game
type GameState struct {
	PlayerID            string              `json:"player_id"`
	CurrentLevel        int                 `json:"current_level"`
	Keystrokes          float64             `json:"keystrokes"`
	Words               int                 `json:"words"`
	Programs            int                 `json:"programs"`
	AIAutomations       int                 `json:"ai_automations"`
	StoryProgress       int                 `json:"story_progress"`
	ProductionRate      float64             `json:"production_rate"`
	KeystrokesPerSecond float64             `json:"keystrokes_per_second"`
	LastSave            time.Time           `json:"last_save"`
	LastUpdate          time.Time           `json:"last_update"`
	Notifications       []string            `json:"notifications"`
	Upgrades            map[string]*Upgrade `json:"upgrades"`
	StoryManager        *StoryManager       `json:"-"`
}

// Upgrade represents a purchased upgrade
type Upgrade struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Level       int       `json:"level"`
	Cost        float64   `json:"cost"`
	Effect      float64   `json:"effect"`
	Description string    `json:"description"`
	PurchasedAt time.Time `json:"purchased_at"`
}

// StoryChapter represents a story event
type StoryChapter struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	TriggerLevel int      `json:"trigger_level"`
	Unlocks      []string `json:"unlocks"`
	IsRead       bool     `json:"is_read"`
}

// Constants for the game
const (
	BaseKeystrokesPerSecond = 1.0
	WordFormationCost       = 100.0
	ProgramFormationCost    = 1000.0
	AIAutomationCost        = 10000.0
)

// NewGameState creates a new game state for a player
func NewGameState(playerID string) *GameState {
	return &GameState{
		PlayerID:            playerID,
		CurrentLevel:        1,
		Keystrokes:          0.0,
		Words:               0,
		Programs:            0,
		AIAutomations:       0,
		StoryProgress:       0,
		ProductionRate:      BaseKeystrokesPerSecond,
		KeystrokesPerSecond: BaseKeystrokesPerSecond,
		LastSave:            time.Now(),
		LastUpdate:          time.Now(),
		Notifications:       make([]string, 0),
		Upgrades:            make(map[string]*Upgrade),
		StoryManager:        NewStoryManager(),
	}
}

// CalculateProduction calculates the current production rate based on upgrades and resources
func (gs *GameState) CalculateProduction() float64 {
	return gs.CalculateProductionWithUpgradeManager(nil)
}

// CalculateProductionWithUpgradeManager calculates production using an upgrade manager
func (gs *GameState) CalculateProductionWithUpgradeManager(um *UpgradeManager) float64 {
	baseProduction := gs.KeystrokesPerSecond
	wordBonus := float64(gs.Words) * 1.5
	programBonus := float64(gs.Programs) * 10.0
	aiBonus := float64(gs.AIAutomations) * 100.0

	// Add upgrade bonuses
	upgradeBonus := 0.0
	if um != nil {
		upgradeBonus = um.GetUpgradeBonus(gs)
	} else {
		// Fallback for backward compatibility
		for _, upgrade := range gs.Upgrades {
			upgradeBonus += upgrade.Effect
		}
	}

	return baseProduction + wordBonus + programBonus + aiBonus + upgradeBonus
}

// UpdateProduction updates the production rate
func (gs *GameState) UpdateProduction() {
	gs.ProductionRate = gs.CalculateProduction()
}

// AddNotification adds a notification to the game state
func (gs *GameState) AddNotification(message string) {
	gs.Notifications = append(gs.Notifications, message)
	// Keep only last 10 notifications
	if len(gs.Notifications) > 10 {
		gs.Notifications = gs.Notifications[len(gs.Notifications)-10:]
	}
}

// CanAfford checks if the player can afford something
func (gs *GameState) CanAfford(cost float64) bool {
	return gs.Keystrokes >= cost
}

// SpendResources spends keystrokes
func (gs *GameState) SpendResources(amount float64) {
	if gs.CanAfford(amount) {
		gs.Keystrokes -= amount
	}
}

// UpdateProduction updates the resources based on time passed
func (gs *GameState) UpdateResources(currentTime time.Time) {
	if gs.LastUpdate.IsZero() {
		gs.LastUpdate = currentTime
		return
	}

	timeDiff := currentTime.Sub(gs.LastUpdate).Seconds()
	production := gs.CalculateProduction()
	keystrokesEarned := production * timeDiff

	gs.Keystrokes += keystrokesEarned
	gs.LastUpdate = currentTime

	// Check for automatic resource formation
	gs.TryFormResources()
}

// TryFormResources attempts to automatically form higher-tier resources
func (gs *GameState) TryFormResources() {
	// Form words from keystrokes
	for gs.Keystrokes >= WordFormationCost {
		gs.Keystrokes -= WordFormationCost
		gs.Words++
		gs.AddNotification("✨ Formed a word!")
	}

	// Form programs from words
	for gs.Words >= 10 {
		gs.Words -= 10
		gs.Programs++
		gs.AddNotification("🚀 Created a program!")
	}

	// Form AI automations from programs
	for gs.Programs >= 5 {
		gs.Programs -= 5
		gs.AIAutomations++
		gs.AddNotification("🤖 Built an AI automation!")
	}
}

// CheckStoryTriggers checks for new story chapters and returns any newly unlocked ones
func (gs *GameState) CheckStoryTriggers() []StoryChapter {
	if gs.StoryManager == nil {
		return nil
	}

	newChapters := gs.StoryManager.CheckTriggers(gs)
	for _, chapter := range newChapters {
		gs.AddNotification(fmt.Sprintf("📖 New Story Chapter Unlocked: %s", chapter.Title))
	}

	return newChapters
}
