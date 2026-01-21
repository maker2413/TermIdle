package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maker2413/term-idle/internal/game"
)

// Model represents the main UI model
type Model struct {
	gameState          *game.GameState
	upgradeManager     *game.UpgradeManager
	leaderboardService *game.LeaderboardService
	width, height      int
	quitting           bool
	lastUpdate         time.Time
	activeTab          string
	selectedUpgrade    int
	leaderboardEntries []*game.LeaderboardEntry
	playerUsername     string
}

// GetGameState returns a copy of the current game state (for testing)
func (m Model) GetGameState() game.GameState {
	return *m.gameState
}

// Messages for tea updates
type ProductionTickMsg time.Time
type GameStateUpdateMsg *game.GameState

// Styles for the UI
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1)

	resourceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#EEEBFF")).
			Padding(0, 1).
			Bold(true)

	notificationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F25D94")).
				Italic(true)
)

// NewModel creates a new UI model with the given game state
func NewModel(gameState *game.GameState) Model {
	return Model{
		gameState:          gameState,
		upgradeManager:     game.NewUpgradeManager(),
		lastUpdate:         time.Now(),
		activeTab:          "game",
		quitting:           false,
		selectedUpgrade:    0,
		leaderboardEntries: make([]*game.LeaderboardEntry, 0),
		playerUsername:     "Player", // Will be set from SSH session
	}
}

// NewModelWithLeaderboard creates a new UI model with leaderboard support
func NewModelWithLeaderboard(gameState *game.GameState, leaderboardService *game.LeaderboardService, username string) Model {
	model := NewModel(gameState)
	model.leaderboardService = leaderboardService
	model.playerUsername = username
	return model
}

// UpdateLeaderboard refreshes the leaderboard data
func (m *Model) UpdateLeaderboard() error {
	if m.leaderboardService == nil {
		return nil
	}

	entries, err := m.leaderboardService.GetFormattedLeaderboard(10)
	if err != nil {
		return err
	}

	m.leaderboardEntries = entries
	return nil
}

// UpdatePlayerLeaderboard updates the current player's leaderboard entry
func (m *Model) UpdatePlayerLeaderboard() error {
	if m.leaderboardService == nil {
		return nil
	}

	return m.leaderboardService.UpdatePlayerLeaderboard(m.gameState, m.playerUsername)
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return ProductionTickMsg(t)
	})
}

// Update handles updates to the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyTab, tea.KeyRight:
			m.switchTab("next")
		case tea.KeyLeft:
			m.switchTab("prev")
		case tea.KeyEnter:
			if m.activeTab == "upgrades" {
				m.handleUpgradePurchase()
			} else {
				m.handleAction()
			}
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				switch msg.Runes[0] {
				case 'r', 'R':
					if m.activeTab == "stats" {
						m.UpdateLeaderboard()
					}
				}
			}
		case tea.KeyUp:
			if m.activeTab == "upgrades" {
				m.selectedUpgrade--
				if m.selectedUpgrade < 0 {
					m.selectedUpgrade = len(m.getAvailableUpgrades()) - 1
				}
			}
		case tea.KeyDown:
			if m.activeTab == "upgrades" {
				availableUpgrades := m.getAvailableUpgrades()
				m.selectedUpgrade = (m.selectedUpgrade + 1) % len(availableUpgrades)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case ProductionTickMsg:
		currentTime := time.Time(msg)
		m.gameState.UpdateResources(currentTime)
		m.gameState.UpdateProduction()
		m.lastUpdate = currentTime

		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return ProductionTickMsg(t)
		})
	}

	return m, nil
}

// View renders the model
func (m Model) View() string {
	if m.quitting {
		return "Thanks for playing Term Idle!\n"
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	switch m.activeTab {
	case "upgrades":
		return m.renderUpgradesView()
	case "story":
		return m.renderStoryView()
	case "stats":
		return m.renderStatsView()
	default:
		return m.renderGameView()
	}
}

// renderGameView renders the main game view
func (m Model) renderGameView() string {
	var content string

	// Title
	title := titleStyle.Render("🐒 Term Idle")
	content += title + "\n\n"

	// Header with resources
	header := m.renderHeader()
	content += header + "\n\n"

	// Main game area
	gameArea := m.renderGameArea()
	content += gameArea + "\n\n"

	// Notifications
	if len(m.gameState.Notifications) > 0 {
		notifications := m.renderNotifications()
		content += notifications + "\n\n"
	}

	// Help text
	help := m.renderHelp()
	content += help

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderHeader renders the resource display header
func (m Model) renderHeader() string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		headerStyle.Render("⌨️ Keystrokes"),
		headerStyle.Render(fmt.Sprintf(" %.1f", m.gameState.Keystrokes)),
		" ",
		headerStyle.Render("📝 Words"),
		headerStyle.Render(fmt.Sprintf(" %d", m.gameState.Words)),
		" ",
		headerStyle.Render("💻 Programs"),
		headerStyle.Render(fmt.Sprintf(" %d", m.gameState.Programs)),
		" ",
		headerStyle.Render("🤖 AI"),
		headerStyle.Render(fmt.Sprintf(" %d", m.gameState.AIAutomations)),
	)
}

// renderGameArea renders the main game interaction area
func (m Model) renderGameArea() string {
	production := m.gameState.CalculateProductionWithUpgradeManager(m.upgradeManager)

	content := resourceStyle.Render(fmt.Sprintf("Production: %.1f keystrokes/second", production))
	content += "\n\n"

	content += lipgloss.NewStyle().Render("🎮 Current Level: " + fmt.Sprintf("%d", m.gameState.CurrentLevel))
	content += "\n\n"

	content += "Press [Enter] to generate keystrokes manually\n"
	content += "Press [Tab] to switch tabs\n"
	content += "Press [Ctrl+C] to quit\n"

	return content
}

// renderNotifications renders recent notifications
func (m Model) renderNotifications() string {
	if len(m.gameState.Notifications) == 0 {
		return ""
	}

	content := "📢 Notifications:\n"
	for i, notification := range m.gameState.Notifications {
		if i >= 3 { // Show only last 3
			break
		}
		content += notificationStyle.Render("  "+notification) + "\n"
	}

	return content
}

// renderHelp renders help text
func (m Model) renderHelp() string {
	tabs := []string{"Game", "Upgrades", "Story", "Stats"}
	var tabLine string

	for i, tab := range tabs {
		if i > 0 {
			tabLine += " "
		}

		if tab == m.activeTab {
			tabLine += activeTabStyle.Render(tab)
		} else {
			tabLine += tabStyle.Render(tab)
		}
	}

	return tabLine
}

// renderUpgradesView renders the upgrades shop
func (m Model) renderUpgradesView() string {
	title := titleStyle.Render("🛠️ Upgrades")
	content := title + "\n\n"

	availableUpgrades := m.getAvailableUpgrades()
	if len(availableUpgrades) == 0 {
		content += "No upgrades available yet. Level up to unlock more!\n\n"
	} else {
		content += "Available Upgrades:\n\n"

		for i, def := range availableUpgrades {
			upgradeType := game.UpgradeType(def.ID)
			canPurchase, _ := m.upgradeManager.CanPurchase(m.gameState, upgradeType)

			// Get current level
			currentLevel := 0
			if currentUpgrade, exists := m.gameState.Upgrades[string(upgradeType)]; exists {
				currentLevel = currentUpgrade.Level
			}

			// Calculate next level cost
			nextLevel := currentLevel + 1
			cost, _ := m.upgradeManager.CalculateCost(upgradeType, nextLevel)

			// Format upgrade line
			line := ""
			if i == m.selectedUpgrade {
				line += "▶ "
			} else {
				line += "  "
			}

			line += fmt.Sprintf("%s (Lv. %d→%d)", def.Name, currentLevel, nextLevel)
			line += fmt.Sprintf(" - Cost: %.1f keystrokes", cost)

			if currentLevel >= def.MaxLevel {
				line += " [MAX]"
			} else if !canPurchase {
				line += " [❌]"
			} else {
				line += " [✅]"
			}

			if i == m.selectedUpgrade {
				line = activeTabStyle.Render(line)
			}

			content += line + "\n"
			content += fmt.Sprintf("      %s\n", def.Description)

			// Calculate and show effect
			if currentLevel < def.MaxLevel {
				effect, _ := m.upgradeManager.CalculateEffect(upgradeType, nextLevel)
				if effect > 0 {
					content += fmt.Sprintf("      Effect: +%.1f production\n", effect)
				}
			}

			content += "\n"
		}
	}

	content += "\n"
	content += "Use [↑↓] to select, [Enter] to purchase\n"
	content += m.renderHelp()

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderStoryView renders the story view
func (m Model) renderStoryView() string {
	title := titleStyle.Render("📖 Story")
	content := title + "\n\n"

	if m.gameState.CurrentLevel == 1 {
		content += "A monkey sits at a keyboard, randomly hitting keys...\n\n"
		content += "Level up to unlock more of the story!\n"
	} else {
		content += "The monkey continues its journey...\n\n"
		content += "More story content coming soon...\n"
	}

	content += "\n" + m.renderHelp()

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderStatsView renders the statistics view
func (m Model) renderStatsView() string {
	title := titleStyle.Render("📊 Statistics & Leaderboard")
	content := title + "\n\n"

	// Player stats
	content += "Your Stats:\n"
	content += fmt.Sprintf("  Level: %d\n", m.gameState.CurrentLevel)
	content += fmt.Sprintf("  Total Keystrokes: %.1f\n", m.gameState.Keystrokes)
	content += fmt.Sprintf("  Words Formed: %d\n", m.gameState.Words)
	content += fmt.Sprintf("  Programs Created: %d\n", m.gameState.Programs)
	content += fmt.Sprintf("  AI Automations: %d\n", m.gameState.AIAutomations)
	content += fmt.Sprintf("  Production Rate: %.1f/sec\n", m.gameState.ProductionRate)
	content += "\n"

	// Leaderboard
	if len(m.leaderboardEntries) > 0 {
		content += "🏆 Top Players:\n\n"
		for _, entry := range m.leaderboardEntries {
			rankDisplay := fmt.Sprintf("%d.", entry.Rank)
			if entry.Rank == 1 {
				rankDisplay = "🥇"
			} else if entry.Rank == 2 {
				rankDisplay = "🥈"
			} else if entry.Rank == 3 {
				rankDisplay = "🥉"
			}

			content += fmt.Sprintf("  %s %-20s %12.1f/s Lvl:%d\n",
				rankDisplay, entry.Username, entry.KeystrokesPerSec, entry.Level)
		}
	} else {
		content += "🏆 Leaderboard (Press R to refresh):\n"
		content += "  Loading or no data available...\n"
	}

	content += "\n"
	content += "Press [R] to refresh leaderboard\n"
	content += m.renderHelp()

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// switchTab switches to the next or previous tab
func (m *Model) switchTab(direction string) {
	tabs := []string{"game", "upgrades", "story", "stats"}

	currentIndex := 0
	for i, tab := range tabs {
		if tab == m.activeTab {
			currentIndex = i
			break
		}
	}

	if direction == "next" {
		currentIndex = (currentIndex + 1) % len(tabs)
	} else {
		currentIndex = (currentIndex - 1 + len(tabs)) % len(tabs)
	}

	m.activeTab = tabs[currentIndex]
}

// getAvailableUpgrades returns available upgrades for the current level
func (m Model) getAvailableUpgrades() []*game.UpgradeDefinition {
	return m.upgradeManager.GetAvailableUpgrades(m.gameState.CurrentLevel)
}

// handleUpgradePurchase handles the purchase of the selected upgrade
func (m *Model) handleUpgradePurchase() {
	availableUpgrades := m.getAvailableUpgrades()
	if m.selectedUpgrade >= len(availableUpgrades) {
		return
	}

	selectedDef := availableUpgrades[m.selectedUpgrade]
	upgradeType := game.UpgradeType(selectedDef.ID)

	err := m.upgradeManager.PurchaseUpgrade(m.gameState, upgradeType)
	if err != nil {
		m.gameState.AddNotification(fmt.Sprintf("❌ %s", err.Error()))
	}
}

// handleAction handles main game actions
func (m *Model) handleAction() {
	// Manual keystroke generation
	keystrokes := m.gameState.KeystrokesPerSecond * 10 // 10 seconds worth
	m.gameState.Keystrokes += keystrokes
	m.gameState.TryFormResources()

	// Check for level up
	newLevel := int(m.gameState.Keystrokes/1000) + 1
	if newLevel > m.gameState.CurrentLevel {
		m.gameState.CurrentLevel = newLevel
		m.gameState.AddNotification(fmt.Sprintf("🎉 Leveled up to %d!", newLevel))
	}
}
