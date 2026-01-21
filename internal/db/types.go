package db

import "time"

// GameState represents the game state stored in the database
type GameState struct {
	PlayerID            string    `json:"player_id"`
	CurrentLevel        int       `json:"current_level"`
	Keystrokes          float64   `json:"keystrokes"`
	Words               int       `json:"words"`
	Programs            int       `json:"programs"`
	AIAutomations       int       `json:"ai_automations"`
	StoryProgress       int       `json:"story_progress"`
	ProductionRate      float64   `json:"production_rate"`
	KeystrokesPerSecond float64   `json:"keystrokes_per_second"`
	LastSave            time.Time `json:"last_save"`
	LastUpdate          time.Time `json:"last_update"`
}
