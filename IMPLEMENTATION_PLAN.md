# Term Idle Implementation Plan

## Current Objective

Build a working MVP of Term Idle - a terminal-based idle game served over SSH where players guide a monkey's evolution from random typing to AI programming. **Repository is completely empty with no Go infrastructure** - must establish all project foundation from zero following AGENTS.md specifications and architecture.md requirements.

## Gap Analysis: Current State vs Requirements

### What Exists (Current State):
- Planning documents (JTBD.md, architecture.md, AGENTS.md)
- Basic project structure documentation
- GitHub repository with git history

### What's Missing (Critical Gaps):
1. **No Go Module**: `go.mod` file doesn't exist - cannot manage dependencies
2. **No Source Code**: Zero `.go` files in entire repository
3. **No Build Infrastructure**: No Makefile, no build scripts, no CI/CD
4. **No Dependencies**: None of the required Go packages are installed or referenced
5. **No Project Structure**: Empty directories `cmd/`, `internal/`, etc. don't exist
6. **No Configuration**: No config files, environment setup, or database schema
7. **No Testing Infrastructure**: No test files, no testing framework setup
8. **No Deployment**: No Docker, scripts, or production configuration

### Priority Assessment:
**CRITICAL (Blocks Everything):** Go module initialization, basic project structure, Makefile
**HIGH:** Dependencies setup, configuration management, database layer
**MEDIUM:** Game engine, SSH server, UI components
**LOW:** Advanced features, monitoring, optimization

## Implementation Tasks

### Phase 1: Project Foundation (Week 1) - CRITICAL PATH - STARTING FROM EMPTY REPO
- [ ] Initialize Go module: `go mod init github.com/maker2413/TermIdleOpenCode`
- [ ] Add core dependencies to go.mod exactly matching AGENTS.md requirements:
  ```go
  require (
      github.com/charmbracelet/bubbletea v0.26.6
      github.com/charmbracelet/lipgloss v0.10.0
      github.com/charmbracelet/wish v0.6.0
      github.com/charmbracelet/log v0.4.0
      github.com/gorilla/mux v1.8.1
      github.com/mattn/go-sqlite3 v1.14.22
      github.com/google/uuid v1.6.0
      golang.org/x/crypto v0.19.0
      github.com/knadh/koanf v1.5.0
      github.com/stretchr/testify v1.8.4
  )
  ```
- [ ] Create complete project directory structure exactly as defined in specs/architecture.md:
  ```
  cmd/term-idle/main.go
  cmd/ssh-server/main.go  
  internal/game/state.go, upgrades.go, production.go, story.go
  internal/ui/model.go, view.go, update.go, components/header.go, components/tabs.go, components/upgrades/, components/story/, components/stats/
  internal/ssh/server.go, session.go, auth.go, handler.go
  internal/db/sqlite.go, models.go, migrations/
  internal/api/server.go, handlers/leaderboard.go, handlers/players.go, middleware/
  internal/config/config.go, loader.go
  pkg/protocol/
  configs/config.yaml
  Makefile
  ```
- [ ] Set up Makefile EXACTLY as specified in AGENTS.md with all commands working:
  - `make build` - Compiles binary to `./term-idle` 
  - `make run` - Runs application with config from cmd/term-idle/main.go
  - `make lint` - Runs golangci-lint (must install and configure)
  - `make test` - Runs tests with coverage for all packages
- [ ] Create .gitignore following Go best practices (bin/, vendor/, *.exe, coverage.out, debug.log)
- [ ] Create configuration management with koanf supporting YAML files and environment variables matching architecture.md patterns
- [ ] Set up structured logging with charmbracelet/log following AGENTS.md error handling patterns
- [ ] Verify everything compiles: run `go mod tidy`, `make build`, and `make lint` to ensure foundation is solid

### Phase 2: Database Layer (Week 1) - PARALLEL with Phase 1 - EMPTY IMPLEMENTATION
- [ ] Set up SQLite database connection using github.com/mattn/go-sqlite3 with proper driver registration
- [ ] Create migration system in internal/db/migrations/ with version tracking and rollback support
- [ ] Create database schema exactly as defined in specs/architecture.md with all required tables:
  ```sql
  players (id TEXT PRIMARY KEY, username TEXT UNIQUE NOT NULL, ssh_key TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, last_active DATETIME DEFAULT CURRENT_TIMESTAMP)
  game_states (player_id TEXT PRIMARY KEY, current_level INTEGER DEFAULT 1, keystrokes REAL DEFAULT 0, words INTEGER DEFAULT 0, programs INTEGER DEFAULT 0, ai_automations INTEGER DEFAULT 0, story_progress INTEGER DEFAULT 0, last_save DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (player_id) REFERENCES players(id))
  player_upgrades (player_id TEXT, upgrade_id TEXT, level INTEGER DEFAULT 0, purchased_at DATETIME DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (player_id, upgrade_id), FOREIGN KEY (player_id) REFERENCES players(id))
  leaderboard_entries (player_id TEXT, keystrokes_per_second REAL DEFAULT 0, total_keystrokes INTEGER DEFAULT 0, level INTEGER DEFAULT 1, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (player_id) REFERENCES players(id))
  story_events (id INTEGER PRIMARY KEY AUTOINCREMENT, trigger_level INTEGER NOT NULL, title TEXT NOT NULL, content TEXT NOT NULL, upgrade_unlock TEXT)
  ```
- [ ] Create Go data models with proper struct tags, validation, and godoc comments EXACTLY following AGENTS.md conventions:
  ```go
  type Player struct {
      ID        string    `json:"id" db:"id"`
      Username  string    `json:"username" db:"username"`
      SSHKey    string    `json:"-" db:"ssh_key"` // Never export sensitive data
      CreatedAt time.Time `json:"created_at" db:"created_at"`
      LastActive time.Time `json:"last_active" db:"last_active"`
  }
  ```
- [ ] Implement Database interface in internal/db/sqlite.go with CRUD operations, proper error handling using fmt.Errorf for wrapping, and godoc comments
- [ ] Add database indexes for performance: CREATE INDEX idx_players_username ON players(username), CREATE INDEX idx_leaderboard_keystrokes ON leaderboard_entries(keystrokes_per_second DESC)
- [ ] Create seed data for initial story events following the 7 chapter monkey evolution story from specs/architecture.md
- [ ] Add connection pooling, transaction support, proper error handling using patterns from AGENTS.md, and structured logging

### Phase 3: Core Game Engine (Week 2) - DEPENDS on Phase 2 - START FROM EMPTY
- [ ] Implement GameState struct in internal/game/state.go EXACTLY as defined in specs/architecture.md with resource tracking:
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
  ```
- [ ] Create production calculation system in internal/game/production.go following specs/architecture.md formulas:
  ```go
  func calculateProduction(state *GameState) float64 {
      base := state.ProductionRate
      wordBonus := float64(state.Words) * 1.5
      programBonus := float64(state.Programs) * 10.0
      aiBonus := float64(state.AIAutomations) * 100.0
      return base + wordBonus + programBonus + aiBonus
  }
  ```
- [ ] Define upgrade types in internal/game/upgrades.go (typing_speed, vocabulary, programming_skills, ai_efficiency) with exponential cost scaling and factory pattern
- [ ] Implement upgrade effect calculations, validation, and purchase logic with proper error handling using AGENTS.md patterns
- [ ] Create production ticker that runs every second using tea.Tick for real-time updates, integrated with Bubbletea message system
- [ ] Add story trigger detection in internal/game/story.go based on 7 chapter monkey evolution from specs/architecture.md
- [ ] Implement resource conversion mechanics (keystrokes → words → programs → AI automations) with proper validation
- [ ] Add comprehensive unit tests for all game logic following testify patterns and AGENTS.md conventions

### Phase 4: SSH Server Implementation (Week 2) - DEPENDS on Phase 1,2 - EMPTY IMPLEMENTATION
- [ ] Set up wish-based SSH server in internal/ssh/server.go EXACTLY following specs/architecture.md middleware chain:
  ```go
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
  }
  ```
- [ ] Implement SSH key authentication in internal/ssh/auth.go with player database verification using golang.org/x/crypto/ssh and AGENTS.md error handling patterns
- [ ] Create session management in internal/ssh/session.go with concurrent player connection support and proper sync.RWMutex usage following AGENTS.md concurrency guidelines
- [ ] Add graceful shutdown and resource cleanup on disconnect using defer patterns and proper cleanup functions per AGENTS.md
- [ ] Integrate game engine initialization with SSH sessions, creating Bubbletea programs for each connection using tea.NewProgram()
- [ ] Implement session context management and player identification with proper error handling using fmt.Errorf wrapping
- [ ] Add connection rate limiting and security measures to prevent abuse following specs/architecture.md security considerations
- [ ] Create comprehensive error handling for failed authentication and connection issues using charmbracelet/log structured logging
- [ ] Add unit tests for SSH server components using mock interfaces and testify patterns

### Phase 5: Basic Bubbletea UI (Week 3) - DEPENDS on Phase 3 - START FROM EMPTY
- [ ] Create main Model struct EXACTLY as defined in specs/architecture.md integrating game state with Bubbletea:
  ```go
  type Model struct {
      gameEngine    *game.Engine
      tabs          *tabs.Model
      components    map[string]tea.Model
      width, height int
      quiting       bool
  }
  ```
- [ ] Implement tab navigation system (Game, Upgrades, Story, Stats) using component architecture from specs/architecture.md
- [ ] Build header component in internal/ui/components/header.go showing real-time resources and production rates using lipgloss styling
- [ ] Add responsive layout handling with tea.WindowSizeMsg for terminal resizing support
- [ ] Implement keyboard controls following AGENTS.md patterns (tab switching, ctrl+c quit, enter/space actions)
- [ ] Create notification system for game events using EventBus pattern from specs/architecture.md
- [ ] Set up lipgloss styling for consistent appearance with predefined styles per AGENTS.md conventions
- [ ] Add viewport management for content scrolling in story and stats views

### Phase 6: Game Views and Interactions (Week 3) - DEPENDS on Phase 5 - ALL COMPONENTS FROM SCRATCH
- [ ] Create game view in internal/ui/components/ with resource counters and main action buttons using component pattern
- [ ] Implement upgrade shop in internal/ui/components/upgrades/ with cost display and purchase animations following specs/architecture.md component hierarchy
- [ ] Build story view in internal/ui/components/story/ showing current chapter and progress timeline using scrollable content
- [ ] Add stats view in internal/ui/components/stats/ with player metrics and leaderboard integration
- [ ] Create upgrade validation logic and purchase confirmations with proper error handling using AGENTS.md patterns
- [ ] Implement story chapter unlocking and progression display tied to GameState.StoryProgress
- [ ] Add achievement notifications and milestone celebrations using the notification system
- [ ] Create help dialog with controls and game instructions accessible via '?' key

### Phase 7: Game Loop Integration (Week 4) - DEPENDS on Phase 3,6 - FULL INTEGRATION FROM GROUND UP
- [ ] Connect production ticker to real-time UI updates via tea.Msg following specs/architecture.md ProductionTickEvent pattern
- [ ] Integrate upgrade purchases with database persistence using Database interface from Phase 2
- [ ] Implement auto-save functionality every 30 seconds using tea.Tick and proper error handling
- [ ] Add story progression triggers and content unlocks using EventBus pattern from specs/architecture.md
- [ ] Create seamless state synchronization between game logic and UI using observer pattern
- [ ] Handle edge cases (network issues, database errors, state corruption) with proper AGENTS.md error handling
- [ ] Optimize performance for multiple concurrent players using goroutine pools and channel communication
- [ ] Add comprehensive error handling and recovery mechanisms using fmt.Errorf wrapping and structured logging

### Phase 8: HTTP API for Leaderboards (Week 4) - PARALLEL with Phase 7 - START FROM EMPTY
- [ ] Set up HTTP server in internal/api/server.go with gorilla/mux routing exactly as defined in specs/architecture.md
- [ ] Implement GET /api/leaderboard endpoint in internal/api/handlers/leaderboard.go with pagination support:
  ```go
  func (s *Server) getLeaderboard(w http.ResponseWriter, r *http.Request) {
      limit := 50
      if l := r.URL.Query().Get("limit"); l != "" {
          if parsed, err := strconv.Atoi(l); err == nil {
              limit = parsed
          }
      }
      entries, err := s.db.GetLeaderboard(limit)
  }
  ```
- [ ] Add GET /api/players/:id/stats endpoint in internal/api/handlers/players.go for individual progress
- [ ] Create authentication middleware in internal/api/middleware/ for API endpoints with proper security
- [ ] Add proper JSON response formatting and error handling following AGENTS.md conventions
- [ ] Implement CORS support in internal/api/middleware/ for potential web dashboard
- [ ] Add rate limiting and input validation following specs/architecture.md security considerations
- [ ] Create API documentation and examples following OpenAPI standards

### Phase 9: Story Content and Game Balance (Week 5) - PARALLEL development - CREATE FROM NOTHING
- [ ] Write 7 story chapters covering monkey evolution EXACTLY matching specs/architecture.md story progression:
  1. The Beginning (Level 1) - "In a digital jungle, a young monkey discovers a keyboard..."
  2. Pattern Recognition (Level 10) - "After countless random taps, the monkey starts seeing patterns..."
  3. First Word (Level 25) - Letter recognition and word formation story
  4. Basic Programming (Level 50) - Simple loops and functions emerge
  5. Advanced Programming (Level 75) - Complex algorithms and optimization
  6. AI Assistant (Level 100) - "The monkey creates its first AI helper!"
  7. AI Programmer (Level 150) - Complete evolution to AI creation
- [ ] Create 20 balanced upgrades with exponential cost scaling following specs/architecture.md upgrade categories:
  - Monkey Evolution upgrades (story progression)
  - Production upgrades (typing_speed, vocabulary, programming_skills, ai_efficiency)
  - Special upgrades (code_review, coffee_rush, stack_overflow_help)
- [ ] Add achievement milestones (first word, first program, first AI automation) with database tracking
- [ ] Balance production rates for satisfying early-game progression using the production formula from Phase 3
- [ ] Implement help system with controls and game mechanics explanation in story view
- [ ] Add tutorial flow for new players triggered on first connection
- [ ] Create reward system for story progression unlocking new upgrade categories

### Phase 10: Testing and Production Readiness (Week 5) - FINAL PHASE - ESTABLISH INFRASTRUCTURE
- [ ] Add comprehensive unit tests for game logic, production calculations, upgrade validation using testify patterns from AGENTS.md
- [ ] Create integration tests for SSH server, database operations, API endpoints using mock interfaces
- [ ] Implement load testing for 10+ concurrent SSH connections simulating real player behavior
- [ ] Set up golangci-lint with configuration EXACTLY matching AGENTS.md standards and ensure `make lint` passes
- [ ] Add end-to-end tests simulating complete player journey from SSH connection to upgrade purchase
- [ ] Create deployment scripts and Docker configuration following specs/architecture.md deployment patterns
- [ ] Implement monitoring and health check endpoints in HTTP API following Prometheus patterns
- [ ] Add comprehensive error logging and debug capabilities using charmbracelet/log from Phase 1
- [ ] Create user documentation and setup guide including SSH connection instructions and gameplay tutorial
- [ ] Verify all AGENTS.md commands work: `make build`, `make run`, `make lint`, `make test` with proper coverage >80%

## Dependencies and Build Order

**Critical Path for MVP (5 weeks total):**
1. **Week 1:** Project Foundation + Database Layer (parallel)
2. **Week 2:** Core Game Engine + SSH Server (parallel)
3. **Week 3:** Basic UI + Game Views (sequential)
4. **Week 4:** Game Loop Integration + HTTP API (parallel)
5. **Week 5:** Content/Balance + Testing/Production (parallel)

**Key Dependencies:**
- Game Engine depends on Database layer
- SSH Server depends on Database and Config
- UI depends on Game Engine
- Game Loop depends on Game Engine + UI
- API depends on Database layer only

**Parallel Development Opportunities:**
- Phase 1 (Foundation) and Phase 2 (Database) can start simultaneously
- Phase 3 (Game Engine) and Phase 4 (SSH) can run in parallel after Phase 1-2
- Phase 8 (API) can start once Phase 2 (Database) is complete
- Phase 9 (Content) can be written while technical phases progress

## MVP Success Criteria

**Functional Requirements (Week 5 milestone):**
- [ ] Players can SSH connect and see interactive TUI game
- [ ] Core idle loop works (keystrokes generate, upgrades purchasable)
- [ ] Story progression unlocks new chapters and upgrades
- [ ] Basic leaderboard displays competitive rankings
- [ ] Game state persists across SSH sessions
- [ ] Production continues when player is offline

**Technical Requirements:**
- [ ] Supports 10+ concurrent SSH players
- [ ] Response time < 200ms for all UI interactions  
- [ ] Auto-save every 30s prevents data loss
- [ ] Clean disconnect handling with state cleanup
- [ ] Proper error recovery and structured logging following AGENTS.md patterns
- [ ] Zero data corruption under concurrent load using proper database transactions
- [ ] make lint passes without errors
- [ ] make test runs with >80% coverage
- [ ] Memory usage stays < 100MB per player session

**Quality Gates:**
- [ ] All unit tests pass (>80% coverage)
- [ ] Integration tests validate full player journey from SSH connection to upgrade purchase
- [ ] `make lint` (golangci-lint) passes without errors following AGENTS.md configuration
- [ ] Load test handles 10+ concurrent SSH connections
- [ ] Memory usage stays < 100MB per player session
- [ ] Database passes consistency checks under concurrent load
- [ ] Story progression triggers correctly in all test scenarios

## Technical Implementation Notes

**Development Strategy:**
- Each phase results in demonstrable functionality increment
- Prioritize working game loop over comprehensive features
- Focus on simplicity and reliability initially, optimize later
- Test integration throughout to prevent regressions
- Use interfaces to enable parallel development and easier testing

**Key Technical Decisions:**
- SQLite for simplicity, can migrate to PostgreSQL later for scaling
- Wish SSH framework for robust terminal handling
- Bubbletea for responsive TUI with component architecture
- Koanf for flexible configuration management
- Structured logging with charmbracelet/log for debugging

**Risk Mitigation:**
- Database transaction handling for concurrent player saves
- Session cleanup to prevent memory leaks
- Input validation to prevent injection attacks
- Rate limiting on SSH connections and API calls
- Graceful degradation for database connectivity issues