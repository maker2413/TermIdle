package game

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUpgradeManager(t *testing.T) {
	um := NewUpgradeManager()
	assert.NotNil(t, um)
	assert.Equal(t, 5, len(um.definitions))
}

func TestGetDefinition(t *testing.T) {
	um := NewUpgradeManager()

	// Test existing upgrade
	def, err := um.GetDefinition(UpgradeTypingSpeed)
	assert.NoError(t, err)
	assert.Equal(t, "Faster Typing", def.Name)
	assert.Equal(t, 50.0, def.BaseCost)

	// Test non-existent upgrade
	_, err = um.GetDefinition("non_existent")
	assert.Error(t, err)
}

func TestCalculateCost(t *testing.T) {
	um := NewUpgradeManager()

	// Test cost calculations for different levels
	cost1, err := um.CalculateCost(UpgradeTypingSpeed, 1)
	assert.NoError(t, err)
	assert.Equal(t, 50.0, cost1)

	cost2, err := um.CalculateCost(UpgradeTypingSpeed, 2)
	assert.NoError(t, err)
	assert.Equal(t, 75.0, cost2) // 50 * 1.5

	cost3, err := um.CalculateCost(UpgradeTypingSpeed, 3)
	assert.NoError(t, err)
	assert.Equal(t, 112.5, cost3) // 75 * 1.5

	// Test max level
	costMax, err := um.CalculateCost(UpgradeTypingSpeed, 51)
	assert.NoError(t, err)
	assert.True(t, math.IsInf(costMax, 1))
}

func TestCalculateEffect(t *testing.T) {
	um := NewUpgradeManager()

	// Test effect calculations
	effect1, err := um.CalculateEffect(UpgradeTypingSpeed, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0.5, effect1)

	effect2, err := um.CalculateEffect(UpgradeTypingSpeed, 2)
	assert.NoError(t, err)
	assert.Equal(t, 0.6, effect2) // 0.5 * 1.2

	// Test level 0
	effect0, err := um.CalculateEffect(UpgradeTypingSpeed, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, effect0)
}

func TestCanPurchase(t *testing.T) {
	um := NewUpgradeManager()
	gs := NewGameState("test")

	// Test basic requirements
	can, err := um.CanPurchase(gs, UpgradeTypingSpeed)
	assert.NoError(t, err)
	assert.False(t, can) // Can't afford

	// Add keystrokes
	gs.Keystrokes = 100.0
	can, err = um.CanPurchase(gs, UpgradeTypingSpeed)
	assert.NoError(t, err)
	assert.True(t, can)

	// Test level requirement
	gs.Keystrokes = 10000.0
	can, err = um.CanPurchase(gs, UpgradeProgramming)
	assert.NoError(t, err)
	assert.False(t, can) // Level too low (level 1 vs required 10)

	gs.CurrentLevel = 10
	can, err = um.CanPurchase(gs, UpgradeProgramming)
	assert.NoError(t, err)
	assert.True(t, can)
}

func TestPurchaseUpgrade(t *testing.T) {
	um := NewUpgradeManager()
	gs := NewGameState("test")
	gs.Keystrokes = 100.0

	// Test successful purchase
	err := um.PurchaseUpgrade(gs, UpgradeTypingSpeed)
	assert.NoError(t, err)

	// Check upgrade was created
	upgrade, exists := gs.Upgrades[string(UpgradeTypingSpeed)]
	assert.True(t, exists)
	assert.Equal(t, 1, upgrade.Level)
	assert.Equal(t, 50.0, upgrade.Cost)
	assert.Equal(t, 0.5, upgrade.Effect)

	// Check keystrokes were deducted
	assert.Equal(t, 50.0, gs.Keystrokes)

	// Check notification
	assert.Contains(t, gs.Notifications, "🎉 Purchased Faster Typing!")

	// Test second level purchase
	gs.Keystrokes = 100.0
	err = um.PurchaseUpgrade(gs, UpgradeTypingSpeed)
	assert.NoError(t, err)

	upgrade = gs.Upgrades[string(UpgradeTypingSpeed)]
	assert.Equal(t, 2, upgrade.Level)
	assert.Equal(t, 75.0, upgrade.Cost)
	assert.Equal(t, 25.0, gs.Keystrokes) // 100 - 75
}

func TestGetAvailableUpgrades(t *testing.T) {
	um := NewUpgradeManager()

	// Test level 1
	available := um.GetAvailableUpgrades(1)
	assert.Equal(t, 2, len(available)) // typing_speed and story_progress

	// Test level 10
	available = um.GetAvailableUpgrades(10)
	assert.Equal(t, 4, len(available)) // adds vocabulary and programming

	// Test level 20
	available = um.GetAvailableUpgrades(20)
	assert.Equal(t, 5, len(available)) // adds ai_efficiency
}

func TestGetUpgradeBonus(t *testing.T) {
	um := NewUpgradeManager()
	gs := NewGameState("test")

	// Test no upgrades
	bonus := um.GetUpgradeBonus(gs)
	assert.Equal(t, 0.0, bonus)

	// Add some upgrades
	gs.Upgrades[string(UpgradeTypingSpeed)] = &Upgrade{
		Type:   string(UpgradeTypingSpeed),
		Level:  2,
		Effect: 0.6,
	}
	gs.Upgrades[string(UpgradeVocabulary)] = &Upgrade{
		Type:   string(UpgradeVocabulary),
		Level:  1,
		Effect: 1.0,
	}
	gs.Upgrades[string(UpgradeStoryProgress)] = &Upgrade{
		Type:   string(UpgradeStoryProgress),
		Level:  1,
		Effect: 0.0, // Story upgrades don't affect production
	}

	bonus = um.GetUpgradeBonus(gs)
	assert.Equal(t, 1.6, bonus) // Only production upgrades count
}

func TestPurchaseUpgradeMaxLevel(t *testing.T) {
	um := NewUpgradeManager()
	gs := NewGameState("test")

	// Purchase up to max level
	purchaseCount := 0
	for {
		// Calculate cost for next level
		nextLevel := purchaseCount + 1
		cost, _ := um.CalculateCost(UpgradeTypingSpeed, nextLevel)

		if math.IsInf(cost, 1) {
			break // Max level reached
		}

		gs.Keystrokes = cost + 100 // Ensure we can afford it
		err := um.PurchaseUpgrade(gs, UpgradeTypingSpeed)
		if err != nil {
			break // Can't purchase anymore
		}
		purchaseCount++
	}

	upgrade := gs.Upgrades[string(UpgradeTypingSpeed)]
	assert.Equal(t, 50, upgrade.Level)
}

func TestStoryUpgradeEffect(t *testing.T) {
	um := NewUpgradeManager()
	gs := NewGameState("test")
	gs.Keystrokes = 100.0

	initialProgress := gs.StoryProgress
	assert.Equal(t, 0, initialProgress)

	err := um.PurchaseUpgrade(gs, UpgradeStoryProgress)
	assert.NoError(t, err)

	assert.Equal(t, 1, gs.StoryProgress)

	// Purchase again
	gs.Keystrokes = 200.0
	err = um.PurchaseUpgrade(gs, UpgradeStoryProgress)
	assert.NoError(t, err)

	assert.Equal(t, 3, gs.StoryProgress) // 1 + 2
}
