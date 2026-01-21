package game

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewGameState(t *testing.T) {
	playerID := uuid.New().String()
	state := NewGameState(playerID)

	if state.PlayerID != playerID {
		t.Errorf("Expected PlayerID %s, got %s", playerID, state.PlayerID)
	}

	if state.CurrentLevel != 1 {
		t.Errorf("Expected CurrentLevel 1, got %d", state.CurrentLevel)
	}

	if state.Keystrokes != 0 {
		t.Errorf("Expected Keystrokes 0, got %f", state.Keystrokes)
	}

	if state.ProductionRate != BaseKeystrokesPerSecond {
		t.Errorf("Expected ProductionRate %f, got %f", BaseKeystrokesPerSecond, state.ProductionRate)
	}
}

func TestCalculateProduction(t *testing.T) {
	state := NewGameState("test-player")

	// Base production should be 1.0
	baseProduction := state.CalculateProduction()
	if baseProduction != BaseKeystrokesPerSecond {
		t.Errorf("Expected base production %f, got %f", BaseKeystrokesPerSecond, baseProduction)
	}

	// Add some resources and check bonuses
	state.Words = 5
	state.Programs = 2
	state.AIAutomations = 1

	expected := BaseKeystrokesPerSecond + 5*1.5 + 2*10.0 + 1*100.0
	actual := state.CalculateProduction()

	if actual != expected {
		t.Errorf("Expected production %f, got %f", expected, actual)
	}
}

func TestCanAfford(t *testing.T) {
	state := NewGameState("test-player")

	// Should not be able to afford anything at start
	if state.CanAfford(100) {
		t.Error("Expected to not afford 100 keystrokes at start")
	}

	// Add keystrokes and test
	state.Keystrokes = 150
	if !state.CanAfford(100) {
		t.Error("Expected to afford 100 keystrokes after adding 150")
	}

	if state.CanAfford(200) {
		t.Error("Expected to not afford 200 keystrokes with only 150")
	}
}

func TestSpendResources(t *testing.T) {
	state := NewGameState("test-player")
	state.Keystrokes = 100

	state.SpendResources(50)
	if state.Keystrokes != 50 {
		t.Errorf("Expected 50 keystrokes after spending 50, got %f", state.Keystrokes)
	}

	// Should not go negative
	state.SpendResources(100)
	if state.Keystrokes != 50 {
		t.Errorf("Expected keystrokes to remain 50 when trying to spend more than available, got %f", state.Keystrokes)
	}
}

func TestUpdateResources(t *testing.T) {
	state := NewGameState("test-player")
	startTime := time.Now()

	// Update after 1 second
	futureTime := startTime.Add(time.Second)
	state.UpdateResources(futureTime)

	expected := BaseKeystrokesPerSecond * 1.0
	if state.Keystrokes < expected-0.0001 || state.Keystrokes > expected+0.0001 {
		t.Errorf("Expected approximately %f keystrokes after 1 second, got %f", expected, state.Keystrokes)
	}
}

func TestTryFormResources(t *testing.T) {
	state := NewGameState("test-player")

	// Test word formation
	state.Keystrokes = WordFormationCost*2 + 50
	state.TryFormResources()

	if state.Words != 2 {
		t.Errorf("Expected 2 words formed, got %d", state.Words)
	}

	if state.Keystrokes < 50 {
		t.Errorf("Expected at least 50 keystrokes remaining, got %f", state.Keystrokes)
	}

	// Test program formation
	state.Words = 15
	state.TryFormResources()

	if state.Programs != 1 {
		t.Errorf("Expected 1 program formed, got %d", state.Programs)
	}

	if state.Words != 5 {
		t.Errorf("Expected 5 words remaining after forming program, got %d", state.Words)
	}

	// Test AI automation formation
	state.Programs = 7
	state.TryFormResources()

	if state.AIAutomations != 1 {
		t.Errorf("Expected 1 AI automation formed, got %d", state.AIAutomations)
	}

	if state.Programs != 2 {
		t.Errorf("Expected 2 programs remaining after forming AI, got %d", state.Programs)
	}
}

func TestAddNotification(t *testing.T) {
	state := NewGameState("test-player")

	state.AddNotification("Test notification")

	if len(state.Notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(state.Notifications))
	}

	if state.Notifications[0] != "Test notification" {
		t.Errorf("Expected 'Test notification', got '%s'", state.Notifications[0])
	}

	// Test notification limit
	for i := 0; i < 15; i++ {
		state.AddNotification(fmt.Sprintf("Notification %d", i))
	}

	if len(state.Notifications) > 10 {
		t.Errorf("Expected max 10 notifications, got %d", len(state.Notifications))
	}
}
