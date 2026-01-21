package ui

import (
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/maker2413/term-idle/internal/game"
	"github.com/stretchr/testify/assert"
)

func TestUpgradeIntegration(t *testing.T) {
	gs := game.NewGameState("test")
	gs.Keystrokes = 1000.0
	gs.CurrentLevel = 5

	model := NewModel(gs)

	// Switch to upgrades tab
	model.switchTab("next")

	// Test that upgrade manager is initialized
	assert.NotNil(t, model.upgradeManager)

	// Test available upgrades
	availableUpgrades := model.getAvailableUpgrades()
	assert.Greater(t, len(availableUpgrades), 0)

	// Test upgrade selection - only test if we have upgrades
	if len(availableUpgrades) > 0 {
		model.handleUpKey()
		model.handleDownKey()
		// Selection should stay valid within bounds
		assert.GreaterOrEqual(t, model.selectedUpgrade, 0)
		assert.Less(t, model.selectedUpgrade, len(availableUpgrades))
	}
}

func TestHandleUpgradePurchase(t *testing.T) {
	gs := game.NewGameState("test")
	gs.Keystrokes = 100.0
	gs.CurrentLevel = 1

	model := NewModel(gs)
	model.activeTab = "upgrades"

	// Get first available upgrade (should be typing_speed)
	availableUpgrades := model.getAvailableUpgrades()
	assert.Greater(t, len(availableUpgrades), 0)

	// Select and purchase upgrade
	model.selectedUpgrade = 0
	model.handleUpgradePurchase()

	// Check upgrade was purchased
	upgrade, exists := gs.Upgrades["typing_speed"]
	assert.True(t, exists)
	assert.Equal(t, 1, upgrade.Level)

	// Check keystrokes were deducted
	assert.Less(t, gs.Keystrokes, 100.0)
}

func TestUpgradePurchaseInsufficientFunds(t *testing.T) {
	gs := game.NewGameState("test")
	gs.Keystrokes = 10.0 // Not enough for any upgrade
	gs.CurrentLevel = 1

	model := NewModel(gs)
	model.activeTab = "upgrades"

	// Try to purchase upgrade
	model.selectedUpgrade = 0
	model.handleUpgradePurchase()

	// Check upgrade was not purchased
	_, exists := gs.Upgrades["typing_speed"]
	assert.False(t, exists)

	// Check keystrokes were not deducted
	assert.Equal(t, 10.0, gs.Keystrokes)

	// Check notification
	assert.Contains(t, gs.Notifications[0], "cannot purchase")
}

func TestRenderUpgradesView(t *testing.T) {
	gs := game.NewGameState("test")
	gs.Keystrokes = 1000.0
	gs.CurrentLevel = 10

	model := NewModel(gs)
	model.activeTab = "upgrades"
	model.width = 80
	model.height = 24

	// Test render
	view := model.renderUpgradesView()
	assert.Contains(t, view, "🛠️ Upgrades")
	assert.Contains(t, view, "Available Upgrades")
}

func TestUpdateWithUpgradeKeys(t *testing.T) {
	gs := game.NewGameState("test")
	gs.Keystrokes = 1000.0
	gs.CurrentLevel = 5

	model := NewModel(gs)
	model.activeTab = "upgrades"

	// Test up key
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(upMsg)
	assert.NotNil(t, newModel)

	// Test down key
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = model.Update(downMsg)
	assert.NotNil(t, newModel)

	// Test enter key for purchase
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = model.Update(enterMsg)
	assert.NotNil(t, newModel)
}

func TestProductionWithUpgrades(t *testing.T) {
	gs := game.NewGameState("test")
	model := NewModel(gs)

	// Add some upgrades manually
	gs.Upgrades["typing_speed"] = &game.Upgrade{
		Type:   "typing_speed",
		Level:  2,
		Effect: 0.6, // Should be calculated from upgrade manager
	}

	// Test production calculation
	production := model.gameState.CalculateProductionWithUpgradeManager(model.upgradeManager)
	assert.Greater(t, production, gs.KeystrokesPerSecond)
}

// Helper methods to simulate key handling
func (m *Model) handleUpKey() {
	availableUpgrades := m.getAvailableUpgrades()
	if len(availableUpgrades) == 0 {
		return
	}
	m.selectedUpgrade--
	if m.selectedUpgrade < 0 {
		m.selectedUpgrade = len(availableUpgrades) - 1
	}
}

func (m *Model) handleDownKey() {
	availableUpgrades := m.getAvailableUpgrades()
	if len(availableUpgrades) == 0 {
		return
	}
	m.selectedUpgrade = (m.selectedUpgrade + 1) % len(availableUpgrades)
}
