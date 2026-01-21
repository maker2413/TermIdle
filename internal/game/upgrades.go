package game

import (
	"fmt"
	"math"
	"time"
)

// UpgradeType represents different categories of upgrades
type UpgradeType string

const (
	UpgradeTypingSpeed   UpgradeType = "typing_speed"
	UpgradeVocabulary    UpgradeType = "vocabulary"
	UpgradeProgramming   UpgradeType = "programming"
	UpgradeAIEfficiency  UpgradeType = "ai_efficiency"
	UpgradeStoryProgress UpgradeType = "story_progress"
)

// UpgradeDefinition defines the properties of an upgrade type
type UpgradeDefinition struct {
	ID               UpgradeType
	Name             string
	Description      string
	BaseCost         float64
	CostMultiplier   float64
	BaseEffect       float64
	EffectMultiplier float64
	MaxLevel         int
	RequiredLevel    int
	Category         string
}

// UpgradeManager handles upgrade definitions and logic
type UpgradeManager struct {
	definitions map[UpgradeType]*UpgradeDefinition
}

// NewUpgradeManager creates a new upgrade manager with predefined upgrades
func NewUpgradeManager() *UpgradeManager {
	um := &UpgradeManager{
		definitions: make(map[UpgradeType]*UpgradeDefinition),
	}

	// Initialize upgrade definitions
	um.initializeUpgrades()
	return um
}

// initializeUpgrades sets up all available upgrade definitions
func (um *UpgradeManager) initializeUpgrades() {
	upgrades := []*UpgradeDefinition{
		{
			ID:               UpgradeTypingSpeed,
			Name:             "Faster Typing",
			Description:      "Increase keystroke generation speed",
			BaseCost:         50.0,
			CostMultiplier:   1.5,
			BaseEffect:       0.5,
			EffectMultiplier: 1.2,
			MaxLevel:         50,
			RequiredLevel:    1,
			Category:         "production",
		},
		{
			ID:               UpgradeVocabulary,
			Name:             "Better Vocabulary",
			Description:      "Improve word formation rate",
			BaseCost:         200.0,
			CostMultiplier:   1.8,
			BaseEffect:       1.0,
			EffectMultiplier: 1.3,
			MaxLevel:         25,
			RequiredLevel:    5,
			Category:         "production",
		},
		{
			ID:               UpgradeProgramming,
			Name:             "Programming Skills",
			Description:      "Increase program creation efficiency",
			BaseCost:         1000.0,
			CostMultiplier:   2.0,
			BaseEffect:       5.0,
			EffectMultiplier: 1.5,
			MaxLevel:         20,
			RequiredLevel:    10,
			Category:         "production",
		},
		{
			ID:               UpgradeAIEfficiency,
			Name:             "AI Efficiency",
			Description:      "Boost automation production",
			BaseCost:         5000.0,
			CostMultiplier:   2.5,
			BaseEffect:       20.0,
			EffectMultiplier: 1.8,
			MaxLevel:         15,
			RequiredLevel:    20,
			Category:         "production",
		},
		{
			ID:               UpgradeStoryProgress,
			Name:             "Story Insight",
			Description:      "Unlock story chapters faster",
			BaseCost:         100.0,
			CostMultiplier:   1.3,
			BaseEffect:       0.0,
			EffectMultiplier: 1.0,
			MaxLevel:         10,
			RequiredLevel:    1,
			Category:         "story",
		},
	}

	for _, upgrade := range upgrades {
		um.definitions[upgrade.ID] = upgrade
	}
}

// GetDefinition returns the definition for an upgrade type
func (um *UpgradeManager) GetDefinition(upgradeType UpgradeType) (*UpgradeDefinition, error) {
	def, exists := um.definitions[upgradeType]
	if !exists {
		return nil, fmt.Errorf("upgrade type %s not found", upgradeType)
	}
	return def, nil
}

// GetAllDefinitions returns all upgrade definitions
func (um *UpgradeManager) GetAllDefinitions() map[UpgradeType]*UpgradeDefinition {
	return um.definitions
}

// GetAvailableUpgrades returns upgrades available at the given level
func (um *UpgradeManager) GetAvailableUpgrades(level int) []*UpgradeDefinition {
	var available []*UpgradeDefinition
	for _, def := range um.definitions {
		if def.RequiredLevel <= level {
			available = append(available, def)
		}
	}
	return available
}

// CalculateCost calculates the cost for an upgrade at a specific level
func (um *UpgradeManager) CalculateCost(upgradeType UpgradeType, level int) (float64, error) {
	def, err := um.GetDefinition(upgradeType)
	if err != nil {
		return 0, err
	}

	if level > def.MaxLevel {
		return math.Inf(1), nil // Return infinity for max level
	}

	cost := def.BaseCost * math.Pow(def.CostMultiplier, float64(level-1))
	return cost, nil
}

// CalculateEffect calculates the effect value for an upgrade at a specific level
func (um *UpgradeManager) CalculateEffect(upgradeType UpgradeType, level int) (float64, error) {
	def, err := um.GetDefinition(upgradeType)
	if err != nil {
		return 0, err
	}

	if level == 0 {
		return 0, nil
	}

	effect := def.BaseEffect * math.Pow(def.EffectMultiplier, float64(level-1))
	return effect, nil
}

// CanPurchase checks if a player can purchase an upgrade
func (um *UpgradeManager) CanPurchase(gs *GameState, upgradeType UpgradeType) (bool, error) {
	def, err := um.GetDefinition(upgradeType)
	if err != nil {
		return false, err
	}

	// Check level requirement
	if gs.CurrentLevel < def.RequiredLevel {
		return false, nil
	}

	// Check if already at max level
	currentUpgrade, exists := gs.Upgrades[string(upgradeType)]
	if exists && currentUpgrade.Level >= def.MaxLevel {
		return false, nil
	}

	// Calculate cost for next level
	nextLevel := 1
	if exists {
		nextLevel = currentUpgrade.Level + 1
	}

	cost, err := um.CalculateCost(upgradeType, nextLevel)
	if err != nil {
		return false, err
	}

	return gs.CanAfford(cost), nil
}

// PurchaseUpgrade processes an upgrade purchase
func (um *UpgradeManager) PurchaseUpgrade(gs *GameState, upgradeType UpgradeType) error {
	canPurchase, err := um.CanPurchase(gs, upgradeType)
	if err != nil {
		return err
	}

	if !canPurchase {
		return fmt.Errorf("cannot purchase upgrade %s", upgradeType)
	}

	def, err := um.GetDefinition(upgradeType)
	if err != nil {
		return err
	}

	// Get current upgrade or create new one
	currentUpgrade, exists := gs.Upgrades[string(upgradeType)]
	newLevel := 1
	if exists {
		newLevel = currentUpgrade.Level + 1
	}

	// Calculate and deduct cost
	cost, err := um.CalculateCost(upgradeType, newLevel)
	if err != nil {
		return err
	}

	gs.SpendResources(cost)

	// Calculate effect
	effect, err := um.CalculateEffect(upgradeType, newLevel)
	if err != nil {
		return err
	}

	// Create or update upgrade
	upgrade := &Upgrade{
		ID:          string(upgradeType),
		Type:        string(upgradeType),
		Level:       newLevel,
		Cost:        cost,
		Effect:      effect,
		Description: fmt.Sprintf("%s (Level %d)", def.Name, newLevel),
		PurchasedAt: time.Now(),
	}

	gs.Upgrades[string(upgradeType)] = upgrade

	// Apply special effects for story upgrades
	if upgradeType == UpgradeStoryProgress {
		gs.StoryProgress += newLevel
	}

	// Add notification
	gs.AddNotification(fmt.Sprintf("🎉 Purchased %s!", def.Name))

	// Update production rate
	gs.UpdateProduction()

	return nil
}

// GetUpgradeBonus calculates the total bonus from all upgrades for production
func (um *UpgradeManager) GetUpgradeBonus(gs *GameState) float64 {
	totalBonus := 0.0

	for _, upgrade := range gs.Upgrades {
		upgradeType := UpgradeType(upgrade.Type)
		def, err := um.GetDefinition(upgradeType)
		if err != nil {
			continue
		}

		// Only production upgrades affect the production rate
		if def.Category == "production" {
			totalBonus += upgrade.Effect
		}
	}

	return totalBonus
}
