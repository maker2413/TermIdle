# Terminal Idle Game Implementation Plan

**Status:** Planning\
**Version:** 1.0\
**Last Updated:** 2026-01-19

### Recent Progress

**2026-01-21:** Phase 3 SSH Server Implementation ✅ IMPLEMENTED
- Completed robust SSH server using wish framework with proper middleware chain
- Implemented SSH key authentication with validation and security features
- Created session management with concurrent player support and proper cleanup
- Added graceful shutdown handling and comprehensive logging
- Integrated game initialization and Bubbletea TUI over SSH connections
- All SSH server components tested with comprehensive unit test coverage
- Added configuration management and host key generation
- Verified SSH server builds and runs correctly with make build-ssh

**2026-01-21:** Phase 4 Main Application Entry Point ✅ IMPLEMENTED
- Completed cmd/term-idle/main.go with proper error handling and graceful shutdown
- Added tea.WithAltScreen() support for better terminal experience
- Verified application builds and runs successfully with make build && make run
- All existing unit tests pass and application is fully functional
- Main entry point creates player sessions and initializes Bubbletea TUI correctly

**2026-01-21:** Phase 2 Game Mechanics ✅ IMPLEMENTED
- Completed upgrade system with 5 upgrade types (typing_speed, vocabulary, programming, ai_efficiency, story_progress)
- Implemented dynamic cost calculation with exponential scaling (1.5x - 2.5x multipliers)
- Created upgrade validation system with level requirements and max levels
- Integrated upgrade system with UI for interactive purchase interface
- Added comprehensive upgrade effects on production calculations
- Implemented story progression upgrades for narrative advancement
- Added 16 comprehensive unit tests for upgrade functionality
- Enhanced UI with upgrade shop navigation (arrow keys) and purchase (enter key)
- Verified upgrade system integration with existing game mechanics

**2026-01-20:** Phase 1 Foundation ✅ IMPLEMENTED
- Completed project directory structure (cmd/, internal/, pkg/, etc.)
- Initialized Go module with all required dependencies (Bubbletea, lipgloss, etc.)
- Set up Makefile with build, test, and lint commands
- Implemented core game state with resource system and production calculations
- Created basic Bubbletea UI with tabs (Game, Upgrades, Story, Stats)
- Added comprehensive unit and integration tests
- Verified application builds and runs successfully

**2026-01-19:** Initial project planning ✅ DOCUMENTATION
- Created comprehensive implementation plan for terminal-based idle game
- Defined project structure and technical architecture
- Outlined 10-phase development approach with Go + Bubbletea + SSH
- Planned Monkey evolution story from random typing to AI programming
- Designed competitive leaderboard system
- Total estimated effort: 13 weeks

### Project Vision

A terminal-based idle game served over SSH using Bubbletea TUI, where players guide a Monkey's journey from randomly hitting keys to becoming an AI programmer. The game combines traditional idle mechanics with a compelling programming evolution story, featuring leaderboards for competitive progression.

---

## Quick Reference

| Component | Technology | Purpose | Status |
|-----------|------------|---------|---------|
| Game Server | Go + Bubbletea | Core game logic and TUI | Planning |
| SSH Server | wish (charmbracelet/wish) | Remote terminal access | Planning |
| Database | SQLite | Player progress, leaderboards | Planning |
| Authentication | SSH keys + usernames | Player identification | Planning |
| API | HTTP/JSON | Leaderboards, player data | Planning |

---

## Phase 1: Core Foundation

**Goal:** Establish basic game structure and infrastructure.

**Status:** Planning

**Estimated Effort:** 1 week

### 1.1 Project Structure

**Status:** ✅ COMPLETED

**Completed:** 2026-01-20

**Implementation:**
```
TermIdle/
├── cmd/
│   ├── term-idle/          # Main game server
│   │   └── main.go
│   └── ssh-server/         # SSH gateway server
│       └── main.go
├── internal/
│   ├── game/              # Core game logic
│   │   ├── state.go       # Game state management
│   │   ├── upgrades.go    # Upgrade system
│   │   ├── production.go  # Resource generation
│   │   └── story.go       # Story progression
│   ├── ui/                # Bubbletea components
│   │   ├── model.go       # Main UI model
│   │   ├── view.go        # Rendering logic
│   │   ├── update.go      # Event handling
│   │   └── components/    # Reusable UI components
│   ├── ssh/               # SSH server handling
│   │   ├── server.go      # SSH server setup
│   │   ├── session.go     # Player session management
│   │   └── handler.go     # Command processing
│   ├── db/                # Database layer
│   │   ├── sqlite.go      # SQLite implementation
│   │   ├── migrations/    # Database migrations
│   │   └── models.go      # Data models
│   └── api/               # HTTP API
│       ├── server.go      # HTTP server
│       ├── handlers/      # API endpoints
│       └── middleware/    # Auth, logging, etc.
├── pkg/                   # Public packages
│   ├── client/            # Optional client library
│   └── protocol/          # Communication protocol
├── web/                   # Web dashboard (optional)
│   └── leaderboard/       # Leaderboard interface
├── configs/               # Configuration files
├── scripts/               # Build/deployment scripts
├── docs/                  # Documentation
└── tests/                 # Integration tests
```

**Checklist:**
- [x] Create basic project directory structure
- [x] Initialize Go module
- [x] Set up Makefile with build commands
- [x] Create placeholder files for main components

### 1.2 Core Dependencies

**Status:** ✅ COMPLETED

**Completed:** 2026-01-20

**Required Go modules:**
```go
// go.mod
module github.com/maker2413/term-idle

require (
    github.com/charmbracelet/bubbletea v0.26.6
    github.com/charmbracelet/lipgloss v0.10.0
    github.com/charmbracelet/wish v0.6.0
    github.com/charmbracelet/log v0.4.0
    github.com/gorilla/mux v1.8.1
    github.com/mattn/go-sqlite3 v1.14.22
    github.com/google/uuid v1.6.0
    golang.org/x/crypto v0.19.0
    github.com/stretchr/testify v1.8.4
)
```

**Checklist:**
- [x] Initialize go.mod with required dependencies
- [x] Set up go.sum verification
- [x] Create vendor directory if needed
- [x] Verify all dependencies compile

### 1.3 Database Schema

**Status:** Planning

**Tables to create:**
```sql
-- Players
CREATE TABLE players (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    ssh_key TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_active DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Game State
CREATE TABLE game_states (
    player_id TEXT PRIMARY KEY,
    current_level INTEGER DEFAULT 1,
    keystrokes INTEGER DEFAULT 0,
    words INTEGER DEFAULT 0,
    programs INTEGER DEFAULT 0,
    ai_automations INTEGER DEFAULT 0,
    story_progress INTEGER DEFAULT 0,
    last_save DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (player_id) REFERENCES players(id)
);

-- Upgrades
CREATE TABLE player_upgrades (
    player_id TEXT,
    upgrade_id TEXT,
    level INTEGER DEFAULT 0,
    purchased_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (player_id, upgrade_id),
    FOREIGN KEY (player_id) REFERENCES players(id)
);

-- Leaderboards
CREATE TABLE leaderboard_entries (
    player_id TEXT,
    keystrokes_per_second REAL DEFAULT 0,
    total_keystrokes INTEGER DEFAULT 0,
    level INTEGER DEFAULT 1,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (player_id) REFERENCES players(id)
);

-- Story Events
CREATE TABLE story_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trigger_level INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    upgrade_unlock TEXT
);
```

**Checklist:**
- [ ] Create database migration file
- [ ] Implement schema creation scripts
- [ ] Add indexes for performance
- [ ] Create data access layer
- [ ] Add foreign key constraints

---

## Phase 2: Game Mechanics

**Goal:** Implement core idle game mechanics.

**Status:** Planning

**Estimated Effort:** 2 weeks

**Dependencies:** Phase 1

### 2.1 Resource System

**Status:** ✅ COMPLETED

**Completed:** 2026-01-21

**Primary Resources:**
- **Keystrokes**: Base currency, generated automatically
- **Words**: Formed from keystrokes, worth more
- **Programs**: Created from words, high value
- **AI Automations**: Ultimate resource, massive production

**Production Formula:**
```go
func calculateProduction(state *GameState) float64 {
    baseProduction := state.KeystrokesPerSecond
    wordBonus := float64(state.Words) * 1.5
    programBonus := float64(state.Programs) * 10.0
    aiBonus := float64(state.AIAutomations) * 100.0
    
    return baseProduction + wordBonus + programBonus + aiBonus + upgradeBonus
}
```

**Checklist:**
- [x] Implement resource types and constants
- [x] Create production calculation functions
- [x] Add resource conversion mechanics
- [x] Implement upgrade cost calculations
- [x] Add unit tests for production formulas

### 2.2 Upgrade System

**Status:** ✅ COMPLETED

**Completed:** 2026-01-21

**Upgrade Categories:**

**Production Upgrades (Implemented):**
- Faster Typing Speed (+keystrokes/sec) - 50 levels, 1.5x cost, 1.2x effect
- Better Vocabulary (+word conversion) - 25 levels, 1.8x cost, 1.3x effect
- Programming Skills (+program value) - 20 levels, 2.0x cost, 1.5x effect
- AI Efficiency (+automation output) - 15 levels, 2.5x cost, 1.8x effect

**Story Progression Upgrades (Implemented):**
- Story Insight - 10 levels, 1.3x cost, unlocks narrative content

**Future Special Upgrades (Planned):**
- Code Review (production boost for 60s)
- Coffee Rush (temporary speed boost)
- Stack Overflow Help (unlock special upgrades)

**Checklist:**
- [x] Define upgrade types and categories
- [x] Implement upgrade purchase logic
- [x] Create upgrade effect calculations
- [ ] Add temporary upgrade system (future)
- [x] Implement upgrade progression validation

### 2.3 Story Integration

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-21

**Story Progression System:**
- Implemented story upgrade system for narrative advancement
- Basic story progression tracking via Story Insight upgrades
- Framework for story event triggering by level

**Story Triggers (Framework):**
```go
type StoryEvent struct {
    TriggerLevel   int
    Title         string
    Content       string
    UpgradeUnlock string
}

var storyEvents = []StoryEvent{
    {1, "The First Key", "Our monkey randomly hits the keyboard. Amazing!", "better_typing"},
    {10, "Letter Recognition", "The monkey starts recognizing patterns!", "vocabulary_boost"},
    {25, "First Word", "\"Hello\" appears on screen. Progress!", "word_formation"},
    {50, "Basic Programming", "Simple loops and functions emerge.", "programming_basics"},
    {100, "AI Assistant", "The monkey creates its first AI helper!", "ai_automation"},
    // ... more story beats
}
```

**Checklist:**
- [x] Define story event structures
- [x] Implement story trigger detection (basic level-based)
- [x] Create story content management (basic)
- [x] Add story progression tracking
- [x] Implement story-unlock system (upgrade-based)

---

## Phase 3: SSH Server Implementation

**Goal:** Create robust SSH access using wish.

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

**Estimated Effort:** 1 week

**Dependencies:** Phase 1

### 3.1 SSH Server Setup

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

```go
// internal/ssh/server.go
func StartSSHServer(config *Config) error {
    s, err := wish.NewServer(
        wish.WithAddress(fmt.Sprintf(":%d", config.SSHPort)),
        wish.WithHostKeyPEM([]byte(config.HostKey)),
        wish.WithMiddleware(
            authMiddleware,
            sessionMiddleware,
            gameMiddleware,
        ),
    )
    if err != nil {
        return fmt.Errorf("failed to create SSH server: %w", err)
    }

    log.Infof("Starting SSH server on port %d", config.SSHPort)
    return s.ListenAndServe()
}
```

**Checklist:**
- [x] Set up wish SSH server
- [x] Configure host key management
- [x] Implement middleware chain
- [x] Add graceful shutdown handling
- [x] Add logging and monitoring

### 3.2 Player Authentication

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

```go
// internal/ssh/auth.go
func authMiddleware(next wish.Middleware) wish.Middleware {
    return func(sess ssh.Session) {
        username := sess.User()
        pk := sess.PublicKey()
        
        // Verify user exists and public key matches
        player, err := db.GetPlayerByUsername(username)
        if err != nil || !keysMatch(pk, player.SSHKey) {
            log.Warnf("Failed auth attempt for user %s", username)
            sess.Exit(1)
            return
        }
        
        // Set player context
        sess.Context().SetValue("player_id", player.ID)
        next(sess)
    }
}
```

**Checklist:**
- [x] Implement SSH key authentication
- [x] Add player verification logic
- [x] Create user registration flow
- [x] Add session context management
- [x] Implement auth failure logging

### 3.3 Game Session Management

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

```go
// internal/ssh/session.go
func sessionMiddleware(next wish.Middleware) wish.Middleware {
    return func(sess ssh.Session) {
        playerID := sess.Context().Value("player_id").(string)
        
        // Load game state
        gameState, err := db.LoadGameState(playerID)
        if err != nil {
            gameState = NewGameState(playerID)
        }
        
        // Create Bubbletea program
        p := tea.NewProgram(NewModel(gameState), tea.WithInput(sess), tea.WithOutput(sess))
        
        // Run game in goroutine
        go func() {
            _, err := p.Run()
            if err != nil {
                log.Errorf("Game session error: %v", err)
            }
        }()
        
        next(sess)
    }
}
```

**Checklist:**
- [x] Implement session state loading
- [x] Create Bubbletea program integration
- [x] Add session cleanup handling
- [x] Implement reconnection support
- [x] Add session monitoring

---

## Phase 4: Bubbletea UI Implementation

**Goal:** Create responsive terminal UI.

**Status:** Planning

**Estimated Effort:** 2 weeks

**Dependencies:** Phase 1, 2

### 4.1 Main Model Structure

**Status:** ✅ COMPLETED

**Completed:** 2026-01-21

**Estimated Effort:** 1 week

**Dependencies:** Phase 2, 4

### 5.1 Production Ticker

**Status:** Planning

```go
// internal/game/production.go
func (g *Game) StartProductionTicker(ctx context.Context) tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return ProductionTickMsg{Time: t}
    })
}

func (m Model) updateProduction(msg ProductionTickMsg) (Model, tea.Cmd) {
    secondsSinceUpdate := msg.Time.Sub(m.lastUpdate).Seconds()
    
    // Calculate production
    production := m.gameState.CalculateProduction()
    keystrokesEarned := production * secondsSinceUpdate
    
    // Update resources
    m.gameState.Keystrokes += keystrokesEarned
    m.gameState.TryUpgradeResources() // Check for word/program formation
    
    // Check story triggers
    if story := m.gameState.CheckStoryTrigger(); story != nil {
        m.notifications = append(m.notifications, story.Title)
    }
    
    m.lastUpdate = msg.Time
    return m, nil
}
```

**Checklist:**
- [ ] Implement production ticker system
- [ ] Add resource update logic
- [ ] Create story trigger detection
- [ ] Add notification handling
- [ ] Optimize performance for concurrent players

### 5.2 Input Handling

```go
// internal/ui/update.go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyEnter, tea.KeySpace:
            cmd = m.handleAction()
        case tea.KeyTab:
            m.tabs = m.tabs.Update(msg)
        case tea.KeyCtrlC:
            return m, tea.Quit
        }
    case ProductionTickMsg:
        return m.updateProduction(msg)
    }
    
    m.tabs, cmd = m.tabs.Update(msg)
    return m, cmd
}
```

---

## Phase 6: Leaderboards and Competition

**Goal:** Add competitive elements with leaderboards.

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

**Estimated Effort:** 1 week

**Dependencies:** Phase 1, 5

### 6.1 Leaderboard API

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

**Implementation:**
- Completed HTTP API server with Gorilla Mux routing
- Implemented GET /api/leaderboard endpoint with limit and sorting
- Implemented GET /api/leaderboard/player/{id} for player rank lookup  
- Implemented GET /api/players/{id} and GET /api/players/username/{name} for player data
- Implemented POST /api/players/{id}/leaderboard for updating player stats
- Added CORS middleware and logging middleware
- Included health check endpoint at /api/health
- Built comprehensive error handling with proper JSON responses

**Checklist:**
- [x] Create HTTP API server structure
- [x] Implement leaderboard retrieval endpoint
- [x] Add player rank lookup endpoint
- [x] Create player data endpoints
- [x] Add leaderboard update endpoint
- [x] Implement CORS and logging middleware
- [x] Add error handling and validation
- [x] Create API response types

### 6.2 Leaderboard Display

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

**Implementation:**
- Enhanced Stats tab to display top 10 leaderboard entries
- Added real-time leaderboard refresh with [R] key
- Integrated medal emojis (🥇🥈🥉) for top 3 players
- Display player production rate and level in leaderboard
- Added player-specific stats section above leaderboard
- Implemented leaderboard service layer for data management
- Added periodic leaderboard updates every 30 seconds
- Created leaderboard entry formatting and display logic

**Checklist:**
- [x] Implement leaderboard service layer
- [x] Integrate leaderboard into UI stats tab
- [x] Add leaderboard refresh functionality
- [x] Display top players with ranks and stats
- [x] Add player position and surrounding players
- [x] Implement periodic leaderboard updates
- [x] Add visual elements (medals, formatting)

### 6.3 Database Layer

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

**Implementation:**
- Completed SQLite database implementation with full schema
- Implemented players table with authentication and profile data
- Created game_states table for current game progress
- Built leaderboard_entries table with competitive rankings
- Added proper indexes for performance optimization
- Implemented database interface for dependency injection
- Created data models for players, game states, and leaderboards
- Added comprehensive database operations (CRUD)

**Checklist:**
- [x] Create database schema with migrations
- [x] Implement player data operations
- [x] Add game state persistence
- [x] Build leaderboard storage and retrieval
- [x] Add database indexes for performance
- [x] Create database interface for testing
- [x] Implement connection management and cleanup

### 6.4 Testing Coverage

**Status:** ✅ IMPLEMENTED

**Completed:** 2026-01-20

**Implementation:**
- Created comprehensive unit tests for database layer
- Built integration tests for API endpoints
- Added mock database for API testing
- Tested all database operations (players, game states, leaderboards)
- Verified API responses and error handling
- Added persistence tests for database durability
- Created leaderboard ranking and sorting tests

**Checklist:**
- [x] Write database unit tests
- [x] Create API integration tests
- [x] Add mock implementations for testing
- [x] Test database persistence and operations
- [x] Verify API endpoints and responses
- [x] Test error handling and edge cases

### 6.1 Leaderboard API

```go
// internal/api/leaderboard.go
type LeaderboardEntry struct {
    Username          string  `json:"username"`
    KeystrokesPerSec  float64 `json:"keystrokes_per_sec"`
    TotalKeystrokes   int64   `json:"total_keystrokes"`
    Level             int     `json:"level"`
    Rank              int     `json:"rank"`
}

func (s *Server) getLeaderboard(w http.ResponseWriter, r *http.Request) {
    limit := 50
    if l := r.URL.Query().Get("limit"); l != "" {
        if parsed, err := strconv.Atoi(l); err == nil {
            limit = parsed
        }
    }
    
    entries, err := s.db.GetLeaderboard(limit)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(entries)
}
```

### 6.2 Leaderboard Display

```go
// internal/ui/leaderboard.go
func (m Model) leaderboardView() string {
    entries, _ := m.getLeaderboard()
    
    var builder strings.Builder
    builder.WriteString(titleStyle.Render("🏆 Leaderboard"))
    builder.WriteString("\n\n")
    
    for i, entry := range entries[:10] { // Top 10
        rank := fmt.Sprintf("%d.", i+1)
        if i < 3 {
            rank = medalEmojis[i] + " " + rank
        }
        
        line := fmt.Sprintf("%-4s %-20s %12.1f/s %10d Lvl:%d",
            rank,
            entry.Username,
            entry.KeystrokesPerSec,
            entry.TotalKeystrokes,
            entry.Level,
        )
        
        builder.WriteString(line + "\n")
    }
    
    return builder.String()
}
```

---

## Phase 7: Story Content System

**Goal:** Create engaging narrative that integrates with gameplay.

**Status:** ✅ IMPLEMENTED

**Estimated Effort:** 2 weeks

**Dependencies:** Phase 2, 4

### 7.1 Story Database

```go
// internal/game/story.go
type StoryChapter struct {
    ID           int    `json:"id"`
    Title        string `json:"title"`
    Content      string `json:"content"`
    TriggerLevel int    `json:"trigger_level"`
    Unlocks      []string `json:"unlocks"`
}

var storyChapters = []StoryChapter{
    {
        ID:           1,
        Title:        "The Beginning",
        Content:      "In a digital jungle, a young monkey discovers a keyboard. Random keystrokes echo through the void...",
        TriggerLevel: 1,
        Unlocks:      []string{"random_typing"},
    },
    {
        ID:           2,
        Title:        "Pattern Recognition",
        Content:      "After countless random taps, the monkey starts seeing patterns. Letters form, then words emerge...",
        TriggerLevel: 10,
        Unlocks:      []string{"letter_recognition", "basic_vocabulary"},
    },
    // ... more chapters
}
```

### 7.2 Story Display System

```go
// internal/ui/story.go
func (m Model) storyView() string {
    currentChapter := m.gameState.GetCurrentChapter()
    
    return lipgloss.JoinVertical(
        lipgloss.Left,
        storyTitleStyle.Render(currentChapter.Title),
        "",
        storyContentStyle.Render(currentChapter.Content),
        "",
        storyProgressStyle.Render(fmt.Sprintf("Progress: %d/%d", 
            m.gameState.StoryProgress, len(storyChapters))),
    )
}
```

---

## Phase 8: Configuration and Deployment

**Goal:** Create production-ready deployment.

**Status:** Planning

**Estimated Effort:** 1 week

**Dependencies:** All phases

### 8.1 Configuration

```yaml
# config.yaml
server:
  port: 8080
  host: "0.0.0.0"

ssh:
  port: 2222
  host_key_file: "/etc/term-idle/host_key"
  max_sessions: 100

database:
  path: "./data/term-idle.db"
  
game:
  save_interval: 30s
  production_tick: 1s
  max_players: 1000

logging:
  level: "info"
  file: "./logs/term-idle.log"
```

### 8.2 Deployment Script

```bash
#!/bin/bash
# scripts/deploy.sh

set -e

echo "Building TermIdle..."
go build -o bin/term-idle cmd/term-idle/main.go
go build -o bin/ssh-server cmd/ssh-server/main.go

echo "Running migrations..."
./bin/term-idle migrate

echo "Starting services..."
./bin/ssh-server &
./bin/term-idle &

echo "Deployment complete!"
echo "SSH: ssh username@localhost -p 2222"
echo "API: http://localhost:8080"
```

---

## Phase 9: Testing and Quality Assurance

**Goal:** Ensure robust, bug-free experience.

**Status:** Planning

**Estimated Effort:** 1 week

**Dependencies:** All phases

### 9.1 Unit Tests

```go
// internal/game/production_test.go
func TestCalculateProduction(t *testing.T) {
    state := &GameState{
        KeystrokesPerSecond: 1.0,
        Words:              5,
        Programs:           2,
        AIAutomations:      1,
    }
    
    expected := 1.0 + 5*1.5 + 2*10.0 + 1*100.0
    actual := calculateProduction(state)
    
    assert.Equal(t, expected, actual)
}
```

### 9.2 Integration Tests

```go
// tests/ssh_test.go
func TestSSHConnection(t *testing.T) {
    // Test SSH server startup
    // Test authentication
    // Test game session creation
    // Test production updates
}
```

### 9.3 Load Testing

```bash
# scripts/load_test.sh
#!/bin/bash

# Simulate 100 concurrent SSH connections
for i in {1..100}; do
    (
        echo "Simulating player $i"
        ssh -i test_keys/player$i test@localhost -p 2222 -o StrictHostKeyChecking=no &
    ) &
done

wait
echo "Load test complete"
```

---

## Phase 10: Monitoring and Analytics

**Goal:** Track game performance and player behavior.

**Status:** Planning

**Estimated Effort:** 1 week

**Dependencies:** Phase 8

### 10.1 Metrics Collection

```go
// internal/metrics/metrics.go
var (
    playersConnected = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "term_idle_players_connected",
        Help: "Number of currently connected players",
    })
    
    totalKeystrokes = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "term_idle_keystrokes_total",
        Help: "Total keystrokes across all players",
    })
)
```

### 10.2 Analytics Dashboard

**Key metrics to track:**
- Daily active players
- Average session duration
- Peak concurrent players
- Progress distribution
- Most popular upgrades
- Story completion rates

---

## Implementation Timeline

| Phase | Description | Status | Estimated Effort |
|-------|-------------|--------|------------------|
| 1. Foundation | Project structure, dependencies, database schema | ✅ COMPLETED | 1 week |
| 2. Game Mechanics | Resource system, upgrades, story integration | ✅ COMPLETED | 2 weeks |
| 3. SSH Server | Authentication, session management, wish integration | Planning | 1 week |
| 4. Bubbletea UI | Terminal UI components, responsive design | ✅ COMPLETED | 2 weeks |
| 5. Game Loop | Real-time updates, production ticker, event handling | ✅ COMPLETED (Basic) | 1 week |
| 6. Leaderboards | Competitive ranking, API endpoints, display | ✅ IMPLEMENTED | 1 week |
| 7. Story System | Narrative content, triggers, progression | ✅ IMPLEMENTED | 2 weeks |
| 8. Deployment | Configuration, scripts, production setup | Planning | 1 week |
| 9. Testing | Unit tests, integration tests, load testing | ✅ COMPLETED (Unit) | 1 week |
| 10. Monitoring | Metrics collection, analytics, dashboards | Planning | 1 week |

**Total estimated effort:** 13 weeks

---

## Build Order

```
Phase 1: Foundation (no deps)
    ↓
Phase 2: Game Mechanics (depends on Phase 1)
Phase 3: SSH Server (depends on Phase 1)
    ↓
Phase 4: Bubbletea UI (depends on Phase 1, 2)
Phase 5: Game Loop (depends on Phase 2, 4)
    ↓
Phase 6: Leaderboards (depends on Phase 1, 5)
Phase 7: Story System (depends on Phase 2, 4)
    ↓
Phase 8: Deployment (depends on all phases)
Phase 9: Testing (depends on all phases)
Phase 10: Monitoring (depends on Phase 8)
```

---

## Success Metrics

**Technical Metrics:**
- < 100ms response time for all inputs
- Support 100+ concurrent players
- 99.9% uptime
- Zero data loss

**Player Metrics:**
- > 50% story completion rate
- > 10 minute average session time
- > 70% player retention after first day
- Active leaderboard participation

---

## Technical Challenges

### SSH Session Management
- Handling multiple concurrent connections
- Maintaining state across reconnections
- Resource cleanup on disconnection

### Real-time Updates
- Efficient production calculations
- Balancing server load with many players
- Preventing cheating/exploits

### Database Performance
- Optimizing for high-frequency updates
- Handling concurrent player saves
- Leaderboard query optimization

### UI Responsiveness
- Terminal size compatibility
- Cross-platform rendering
- Accessibility considerations

---

## Success Metrics

**Technical Metrics:**
- < 100ms response time for all inputs
- Support 100+ concurrent players
- 99.9% uptime
- Zero data loss

**Player Metrics:**
- > 50% story completion rate
- > 10 minute average session time
- > 70% player retention after first day
- Active leaderboard participation

---

## Future Enhancements

### Short Term (Post-MVP)
- Player profiles and customization
- Achievement system
- Special events and challenges
- Mobile SSH client

### Long Term
- Guild/team play
- Mini-games within main game
- Custom story creation tools
- Plugin system for community content

---

## Technical Challenges

### SSH Session Management
- Handling multiple concurrent connections
- Maintaining state across reconnections
- Resource cleanup on disconnection

### Real-time Updates
- Efficient production calculations
- Balancing server load with many players
- Preventing cheating/exploits

### Database Performance
- Optimizing for high-frequency updates
- Handling concurrent player saves
- Leaderboard query optimization

### UI Responsiveness
- Terminal size compatibility
- Cross-platform rendering
- Accessibility considerations

---

## Future Enhancements

### Short Term (Post-MVP)
- Player profiles and customization
- Achievement system
- Special events and challenges
- Mobile SSH client

### Long Term
- Guild/team play
- Mini-games within main game
- Custom story creation tools
- Plugin system for community content

---

## Conclusion

This implementation plan provides a comprehensive roadmap for building a terminal-based idle game that combines engaging gameplay with a compelling narrative. The technical architecture is designed for scalability, performance, and maintainability while providing an authentic terminal experience.

The SSH-based delivery model creates a unique, nostalgic gaming experience that stands out from web-based idle games, while the story-driven progression keeps players engaged beyond simple number growth.

By following this phased approach, we can build a robust, feature-rich game that delivers on both the technical requirements and player experience goals.</arg_value>
</tool_call>

---
