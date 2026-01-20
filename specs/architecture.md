## Overview

Term Idle is a terminal-based idle game served over SSH using Bubbletea TUI, where players guide a Monkey's journey from randomly hitting keys to becoming an AI programmer. The game combines traditional idle mechanics with a compelling programming evolution story, featuring leaderboards for competitive progression.

The system is designed around three core principles:

1. **Simplicity** - Minimal dependencies with clear separation of concerns
2. **Scalability** - Support for 100+ concurrent players with efficient resource management  
3. **Authenticity** - True terminal experience using SSH and TUI interfaces

## High-Level Architecture

```
┌─────────────────┐    SSH Connection   ┌─────────────────┐    HTTP/JSON    ┌─────────────────┐
│   Player        │ ──────────────────▶ │  SSH Server     │ ──────────────▶ │   Web API       │
│   (SSH Client)  │  (wish + Bubbletea) │                 │                 │   (Leaderboards)│
└─────────────────┘                     └─────────────────┘                 └─────────────────┘
                                                │
                                                ▼
                                        ┌─────────────────┐
                                        │   Game Engine   │
                                        │   (Bubbletea)   │
                                        └─────────────────┘
                                                │
                                                ▼
                                        ┌─────────────────┐
                                        │   Database      │
                                        │   (SQLite)      │
                                        └─────────────────┘
```

## Package Structure

```
TermIdle/
├── cmd/
│   ├── term-idle/          # Main game server binary
│   │   └── main.go
│   └── ssh-server/         # SSH gateway server binary  
│       └── main.go
├── internal/
│   ├── game/              # Core game logic
│   │   ├── state.go       # Game state management
│   │   ├── upgrades.go    # Upgrade system
│   │   ├── production.go  # Resource generation
│   │   └── story.go       # Story progression
│   ├── ui/                # Bubbletea UI components
│   │   ├── model.go       # Main UI model
│   │   ├── view.go        # Rendering logic
│   │   ├── update.go      # Event handling
│   │   └── components/    # Reusable UI components
│   │       ├── header.go  # Header with resources
│   │       ├── tabs.go    # Navigation tabs
│   │       ├── upgrades/  # Upgrade shop UI
│   │       ├── story/     # Story display UI
│   │       └── stats/     # Statistics UI
│   ├── ssh/               # SSH server handling
│   │   ├── server.go      # SSH server setup
│   │   ├── session.go     # Player session management
│   │   ├── auth.go        # Authentication middleware
│   │   └── handler.go     # Command processing
│   ├── db/                # Database layer
│   │   ├── sqlite.go      # SQLite implementation
│   │   ├── migrations/    # Database migrations
│   │   └── models.go      # Data models
│   ├── api/               # HTTP API for leaderboards
│   │   ├── server.go      # HTTP server
│   │   ├── handlers/      # API endpoints
│   │   │   ├── leaderboard.go
│   │   │   └── players.go
│   │   └── middleware/    # Auth, logging, CORS
│   └── config/            # Configuration management
│       ├── config.go      # Configuration structs
│       └── loader.go      # YAML/ENV loading
├── pkg/                   # Public packages
│   ├── client/            # Optional client library
│   └── protocol/          # Communication protocol definitions
├── configs/               # Configuration files
│   ├── config.yaml
│   └── config.example.yaml
├── scripts/               # Build/deployment scripts
├── docs/                  # Documentation
└── tests/                 # Integration tests
```

## Component Architecture

### Game Engine (internal/game/)

Core game logic separated from UI and networking:

```go
type GameState struct {
    PlayerID           string
    CurrentLevel       int
    Keystrokes         float64
    Words              int
    Programs           int
    AIAutomations      int
    StoryProgress      int
    Upgrades           map[string]*Upgrade
    ProductionRate     float64
    LastSave           time.Time
    Notifications      []string
}

type GameEngine struct {
    state  *GameState
    db     Database
    config *Config
}
```

**Key responsibilities:**
- Resource calculation and production
- Upgrade validation and application
- Story progression tracking
- Auto-save functionality
- Event generation (story triggers, achievements)

### SSH Server (internal/ssh/)

Wish-based SSH server with authentication and session management:

```go
type Server struct {
    config   *Config
    db       Database
    sessions map[string]*Session
    mu       sync.RWMutex
}

type Session struct {
    PlayerID    string
    GameEngine  *GameEngine
    Program     *tea.Program
    LastActive  time.Time
}
```

**Middleware chain:**
1. **Authentication** - SSH key verification against database
2. **Session Creation** - Load game state, create Bubbletea program
3. **Game Execution** - Run TUI interface over SSH connection
4. **Cleanup** - Save state, close resources on disconnect

### UI System (internal/ui/)

Bubbletea-based terminal UI with component architecture:

```go
type Model struct {
    gameEngine    *game.Engine
    tabs          *tabs.Model
    components    map[string]tea.Model
    width, height int
    quiting       bool
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKey(msg)
    case tea.WindowSizeMsg:
        return m.handleResize(msg)
    case game.ProductionTickMsg:
        return m.handleProduction(msg)
    case game.StoryTriggerMsg:
        return m.handleStory(msg)
    }
    // Delegate to active tab/component
    return m.updateComponents(msg)
}
```

**Component hierarchy:**
- **Header** - Resource display, production rates
- **Tab Navigation** - Game, Upgrades, Story, Stats
- **Game View** - Main action buttons, progress bars
- **Upgrade Shop** - Purchase interface with animations
- **Story View** - Scrollable narrative content
- **Stats View** - Player statistics and leaderboards

### Database Layer (internal/db/)

SQLite with migration support and optimized queries:

```go
type Database interface {
    GetPlayer(id string) (*Player, error)
    SavePlayer(player *Player) error
    GetGameState(playerID string) (*game.State, error)
    SaveGameState(state *game.State) error
    GetLeaderboard(limit int) ([]*LeaderboardEntry, error)
    UpdateLeaderboard(entry *LeaderboardEntry) error
}

type SQLiteDB struct {
    db *sql.DB
}
```

**Schema design:**
- **players** - Authentication and profile data
- **game_states** - Current game progress (one-to-one with players)
- **upgrades** - Player purchase history
- **leaderboards** - Competitive rankings
- **story_events** - Narrative triggers and content

### HTTP API (internal/api/)

REST API for leaderboards and external integration:

```go
type Server struct {
    config *Config
    db     Database
    router *mux.Router
}

// GET /api/leaderboard?limit=50
func (s *Server) getLeaderboard(w http.ResponseWriter, r *http.Request)

// GET /api/players/:id/stats  
func (s *Server) getPlayerStats(w http.ResponseWriter, r *http.Request)

// POST /api/players/:id/leaderboard
func (s *Server) updateLeaderboard(w http.ResponseWriter, r *http.Request)
```

## Design Patterns Used

### Component Pattern (Bubbletea UI)

Each UI element is a separate Bubbletea model:

```go
type ResourceDisplay struct {
    state  *game.State
    width  int
    style  lipgloss.Style
}

func (r ResourceDisplay) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Only handle resize, state is read-only
    if msg, ok := msg.(tea.WindowSizeMsg); ok {
        r.width = msg.Width
    }
    return r, nil
}

func (r ResourceDisplay) View() string {
    // Render current resource levels
    return r.style.Render(fmt.Sprintf("Keystrokes: %.1f/s", r.state.ProductionRate))
}
```

### Observer Pattern (Game Events)

Game engine emits events that UI components can subscribe to:

```go
type GameEvent interface {
    Type() string
    Timestamp() time.Time
}

type ProductionTickEvent struct {
    time     time.Time
    produced float64
}

type StoryTriggerEvent struct {
    time   time.Time
    story  *StoryChapter
    level  int
}

type EventBus interface {
    Subscribe(eventType string, handler EventHandler)
    Publish(event GameEvent)
}
```

### Repository Pattern (Database)

Abstract database operations behind interfaces:

```go
type PlayerRepository interface {
    FindByID(id string) (*Player, error)
    FindByUsername(username string) (*Player, error)
    Save(player *Player) error
    Delete(id string) error
}

type GameStateRepository interface {
    Load(playerID string) (*game.State, error)
    Save(state *game.State) error
    UpdateProduction(playerID string, rate float64) error
}
```

### Factory Pattern (Upgrades)

Dynamic upgrade creation based on configuration:

```go
type UpgradeFactory interface {
    CreateUpgrade(upgradeType string, level int) (*Upgrade, error)
    GetAvailableUpgrades(level int) []*UpgradeDefinition
}

type BaseUpgradeFactory struct {
    definitions map[string]*UpgradeDefinition
}

func (f *BaseUpgradeFactory) CreateUpgrade(upgradeType string, level int) (*Upgrade, error) {
    def, exists := f.definitions[upgradeType]
    if !exists {
        return nil, fmt.Errorf("unknown upgrade type: %s", upgradeType)
    }
    
    return &Upgrade{
        Type:        upgradeType,
        Level:       level,
        Cost:        f.calculateCost(def, level),
        Effect:      f.calculateEffect(def, level),
        Description: fmt.Sprintf(def.DescriptionTemplate, level),
    }, nil
}
```

### Strategy Pattern (Authentication)

Multiple authentication strategies:

```go
type AuthProvider interface {
    Authenticate(username, key string) (*Player, error)
    Register(username, key string) (*Player, error)
}

type SSHKeyAuthProvider struct {
    db Database
}

type TestAuthProvider struct {
    // For development/testing
    allowedUsers map[string]string
}
```

## Data Flow

### Player Connection Flow

```
1. SSH Connection → wish.Server
2. Authentication Middleware → SSHKeyAuthProvider.Authenticate()
3. Session Creation → Database.LoadGameState()
4. Bubbletea Program → tea.NewProgram()
5. UI Rendering → Component.Update()/View()
6. Game Loop → ProductionTicker → EventBus.Publish()
7. Disconnection → Session.Cleanup() → Database.SaveGameState()
```

### Production Update Flow

```
1. Timer (every 1s) → GameEngine.ProductionTick()
2. Calculate production → UpgradeSystem.CalculateProduction()
3. Update resources → GameState.Keystrokes += production
4. Check story triggers → StorySystem.CheckTriggers()
5. Publish events → EventBus.Publish(ProductionTickEvent)
6. Auto-save → Database.SaveGameState()
7. Update leaderboard → Database.UpdateLeaderboard()
```

### Upgrade Purchase Flow

```
1. Player Action → UI.UpgradeButtonClicked()
2. Validate cost → GameState.CanAfford(upgrade)
3. Deduct resources → GameState.SpendResources(cost)
4. Apply effects → UpgradeSystem.ApplyUpgrade()
5. Save state → Database.SaveGameState()
6. Update UI → Component.Refresh()
7. Notify player → EventBus.Publish(UpgradePurchasedEvent)
```

## Extension Points

### Adding New Upgrade Types

1. Define upgrade behavior:

```go
type AutomationUpgrade struct {
    BaseUpgrade
    Efficiency float64
}

func (a *AutomationUpgrade) ApplyEffect(state *game.State) {
    state.ProductionRate += a.Efficiency * float64(state.AIAutomations)
}
```

2. Register in factory:

```go
factory.RegisterUpgradeType("automation_efficiency", &AutomationUpgradeFactory{})
```

3. Add UI components:

```go
type AutomationUpgradeComponent struct {
    upgrades []*AutomationUpgrade
    // ... UI fields
}
```

### Adding New Story Chapters

1. Define chapter in database:

```sql
INSERT INTO story_events (trigger_level, title, content, upgrade_unlock) 
VALUES (50, 'AI Assistant', 'The monkey creates its first AI helper!', 'ai_automation');
```

2. Create story event handler:

```go
type AIStoryHandler struct{}

func (h *AIStoryHandler) Handle(state *game.State, event *StoryEvent) error {
    state.UnlockUpgrade("ai_automation")
    state.AddNotification("🤖 AI automation unlocked!")
    return nil
}
```

### Adding New Leaderboard Categories

1. Extend database schema:

```sql
ALTER TABLE leaderboard_entries 
ADD COLUMN words_per_second REAL DEFAULT 0,
ADD COLUMN programs_completed INTEGER DEFAULT 0;
```

2. Update API endpoints:

```go
type LeaderboardEntry struct {
    // ... existing fields
    WordsPerSecond   float64 `json:"words_per_second"`
    ProgramsCompleted int64  `json:"programs_completed"`
}
```

3. Add UI display options:

```go
type LeaderboardView struct {
    mode LeaderboardMode // keystrokes, words, programs
}

type LeaderboardMode int

const (
    ModeKeystrokes LeaderboardMode = iota
    ModeWords
    ModePrograms
)
```

## Performance Considerations

### Database Optimizations

- **Indexing**: Player lookups by username, leaderboard queries by score
- **Connection Pooling**: Reuse database connections across requests
- **Batch Updates**: Update leaderboards in batches every 30 seconds
- **Read Replicas**: Separate read-only DB for leaderboard queries (future)

### Memory Management

- **Session Limits**: Cap concurrent sessions per player
- **Resource Cleanup**: Automatic cleanup of disconnected sessions
- **Event Buffering**: Limit queued events per session
- **UI Component Caching**: Cache rendered components where possible

### Concurrency

- **Goroutine Pools**: Limit concurrent game calculations
- **Mutex Granularity**: Fine-grained locking in game state
- **Channel Communication**: Non-blocking updates between components
- **Rate Limiting**: Prevent abuse of API endpoints

## Security Considerations

### SSH Security

- **Key Authentication**: Only SSH public key authentication
- **Rate Limiting**: Limit connection attempts per IP
- **Session Isolation**: Separate goroutines per player
- **Input Validation**: Sanitize all user inputs

### Data Protection

- **SQL Injection**: Parameterized queries throughout
- **Input Validation**: Validate all configuration and user data
- **Secret Management**: Never log sensitive player data
- **Access Control**: Players can only access their own data

### Resource Limits

- **Session Duration**: Automatic timeout for idle connections
- **Memory Limits**: Caps on per-session memory usage
- **Database Limits**: Query timeouts and connection limits
- **File System**: No direct file system access for players

## Deployment Architecture

### Production Setup

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   SSH Server    │    │   Game Server   │
│   (nginx/HAProxy)│◀──▶│   (port 2222)   │    │   (port 8080)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                └──────────┬─────────────┘
                                           ▼
                                    ┌─────────────────┐
                                    │   Database      │
                                    │   (SQLite)      │
                                    └─────────────────┘
```

### Scaling Strategy

1. **Horizontal**: Multiple SSH servers behind load balancer
2. **Database**: Shared SQLite file with proper locking (or migrate to PostgreSQL)
3. **Caching**: Redis for session state and leaderboards
4. **Monitoring**: Metrics collection and alerting

### Development Setup

```
docker-compose.yml:
- term-idle-ssh (port 2222)
- term-idle-api (port 8080) 
- sqlite (volume mounted)
- redis (optional, for caching)
```

## Technology Stack

### Core Dependencies

- **Go 1.25+** - Core language and runtime
- **github.com/charmbracelet/bubbletea** - TUI framework
- **github.com/charmbracelet/lipgloss** - Terminal styling
- **github.com/charmbracelet/wish** - SSH server framework
- **github.com/gorilla/mux** - HTTP routing
- **github.com/mattn/go-sqlite3** - SQLite driver
- **github.com/knadh/koanf** - Configuration management

### Development Tools

- **github.com/stretchr/testify** - Testing framework
- **github.com/golangci/golangci-lint** - Linting
- **github.com/prometheus/client_golang** - Metrics collection
- **github.com/charmbracelet/log** - Structured logging

### Build and Deployment

- **Makefile** - Build automation
- **Docker** - Containerization
- **GitHub Actions** - CI/CD pipeline
- **SQLite** - Embedded database

## Future Extensibility

The architecture supports several future enhancements:

### Multi-Game Support
- Abstract game engine interface
- Multiple game modes in same server
- Shared authentication and leaderboards

### Plugin System
- Dynamic upgrade loading
- Custom story modules
- Third-party UI components

### Real-time Features
- WebSocket support for web clients
- Live leaderboards
- Multi-player competitions

### Analytics Expansion
- Player behavior tracking
- A/B testing framework
- Advanced reporting dashboard
