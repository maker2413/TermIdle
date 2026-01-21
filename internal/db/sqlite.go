package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database interface for data access operations
type Database interface {
	// Player operations
	CreatePlayer(player *Player) error
	GetPlayer(id string) (*Player, error)
	GetPlayerByUsername(username string) (*Player, error)
	UpdatePlayer(player *Player) error

	// Game state operations
	SaveGameState(state *GameState) error
	LoadGameState(playerID string) (*GameState, error)

	// Leaderboard operations
	GetLeaderboard(limit int) ([]*LeaderboardEntry, error)
	UpdateLeaderboard(entry *LeaderboardEntry) error
	GetPlayerRank(playerID string) (int, error)

	// Database operations
	Close() error
	Migrate() error
}

// Player represents a player in the system
type Player struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	SSHKey     string    `json:"ssh_key"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	PlayerID         string    `json:"player_id"`
	Username         string    `json:"username"`
	KeystrokesPerSec float64   `json:"keystrokes_per_sec"`
	TotalKeystrokes  float64   `json:"total_keystrokes"`
	Words            int       `json:"words"`
	Programs         int       `json:"programs"`
	AIAutomations    int       `json:"ai_automations"`
	Level            int       `json:"level"`
	Rank             int       `json:"rank"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SQLiteDB implements the Database interface using SQLite
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB creates a new SQLite database instance
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqliteDB := &SQLiteDB{db: db}

	return sqliteDB, nil
}

// NewSQLiteDBWithMigration creates a new SQLite database instance and runs migrations
func NewSQLiteDBWithMigration(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqliteDB := &SQLiteDB{db: db}

	// Initialize database schema
	if err := sqliteDB.Migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return sqliteDB, nil
}

// Migrate creates the necessary database tables
func (s *SQLiteDB) Migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		ssh_key TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_active DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS game_states (
		player_id TEXT PRIMARY KEY,
		current_level INTEGER DEFAULT 1,
		keystrokes REAL DEFAULT 0.0,
		words INTEGER DEFAULT 0,
		programs INTEGER DEFAULT 0,
		ai_automations INTEGER DEFAULT 0,
		story_progress INTEGER DEFAULT 0,
		production_rate REAL DEFAULT 1.0,
		keystrokes_per_second REAL DEFAULT 1.0,
		last_save DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_update DATETIME DEFAULT CURRENT_TIMESTAMP,
		notifications TEXT DEFAULT '[]',
		upgrades TEXT DEFAULT '{}',
		FOREIGN KEY (player_id) REFERENCES players(id)
	);

	CREATE TABLE IF NOT EXISTS leaderboard_entries (
		player_id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		keystrokes_per_second REAL DEFAULT 0.0,
		total_keystrokes REAL DEFAULT 0.0,
		words INTEGER DEFAULT 0,
		programs INTEGER DEFAULT 0,
		ai_automations INTEGER DEFAULT 0,
		level INTEGER DEFAULT 1,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (player_id) REFERENCES players(id)
	);

	CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
	CREATE INDEX IF NOT EXISTS idx_leaderboard_keystrokes_per_sec ON leaderboard_entries(keystrokes_per_second DESC);
	CREATE INDEX IF NOT EXISTS idx_leaderboard_total_keystrokes ON leaderboard_entries(total_keystrokes DESC);
	CREATE INDEX IF NOT EXISTS idx_leaderboard_level ON leaderboard_entries(level DESC);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// CreatePlayer creates a new player
func (s *SQLiteDB) CreatePlayer(player *Player) error {
	query := `
	INSERT INTO players (id, username, ssh_key, created_at, last_active)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, player.ID, player.Username, player.SSHKey, player.CreatedAt, player.LastActive)
	if err != nil {
		return fmt.Errorf("failed to create player: %w", err)
	}

	return nil
}

// GetPlayer retrieves a player by ID
func (s *SQLiteDB) GetPlayer(id string) (*Player, error) {
	query := `
	SELECT id, username, ssh_key, created_at, last_active
	FROM players
	WHERE id = ?
	`

	player := &Player{}
	err := s.db.QueryRow(query, id).Scan(
		&player.ID, &player.Username, &player.SSHKey,
		&player.CreatedAt, &player.LastActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return player, nil
}

// GetPlayerByUsername retrieves a player by username
func (s *SQLiteDB) GetPlayerByUsername(username string) (*Player, error) {
	query := `
	SELECT id, username, ssh_key, created_at, last_active
	FROM players
	WHERE username = ?
	`

	player := &Player{}
	err := s.db.QueryRow(query, username).Scan(
		&player.ID, &player.Username, &player.SSHKey,
		&player.CreatedAt, &player.LastActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found: %s", username)
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return player, nil
}

// UpdatePlayer updates an existing player
func (s *SQLiteDB) UpdatePlayer(player *Player) error {
	query := `
	UPDATE players
	SET username = ?, ssh_key = ?, last_active = ?
	WHERE id = ?
	`

	_, err := s.db.Exec(query, player.Username, player.SSHKey, player.LastActive, player.ID)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	return nil
}

// SaveGameState saves the game state for a player
func (s *SQLiteDB) SaveGameState(state *GameState) error {
	query := `
	INSERT OR REPLACE INTO game_states (
		player_id, current_level, keystrokes, words, programs, ai_automations,
		story_progress, production_rate, keystrokes_per_second, last_save,
		last_update, notifications, upgrades
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		state.PlayerID, state.CurrentLevel, state.Keystrokes, state.Words,
		state.Programs, state.AIAutomations, state.StoryProgress,
		state.ProductionRate, state.KeystrokesPerSecond, state.LastSave,
		state.LastUpdate, "[]", "{}",
	)

	if err != nil {
		return fmt.Errorf("failed to save game state: %w", err)
	}

	return nil
}

// LoadGameState loads the game state for a player
func (s *SQLiteDB) LoadGameState(playerID string) (*GameState, error) {
	query := `
	SELECT player_id, current_level, keystrokes, words, programs, ai_automations,
	       story_progress, production_rate, keystrokes_per_second, last_save,
	       last_update
	FROM game_states
	WHERE player_id = ?
	`

	state := &GameState{}
	err := s.db.QueryRow(query, playerID).Scan(
		&state.PlayerID, &state.CurrentLevel, &state.Keystrokes,
		&state.Words, &state.Programs, &state.AIAutomations,
		&state.StoryProgress, &state.ProductionRate, &state.KeystrokesPerSecond,
		&state.LastSave, &state.LastUpdate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game state not found for player: %s", playerID)
		}
		return nil, fmt.Errorf("failed to load game state: %w", err)
	}

	// Initialize state without complex structures that are handled by game layer

	return state, nil
}

// GetLeaderboard retrieves the top entries from the leaderboard
func (s *SQLiteDB) GetLeaderboard(limit int) ([]*LeaderboardEntry, error) {
	query := `
	SELECT le.player_id, le.username, le.keystrokes_per_second, le.total_keystrokes,
	       le.words, le.programs, le.ai_automations, le.level,
	       RANK() OVER (ORDER BY le.total_keystrokes DESC) as rank,
	       le.updated_at
	FROM leaderboard_entries le
	ORDER BY le.total_keystrokes DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []*LeaderboardEntry
	for rows.Next() {
		entry := &LeaderboardEntry{}
		err := rows.Scan(
			&entry.PlayerID, &entry.Username, &entry.KeystrokesPerSec,
			&entry.TotalKeystrokes, &entry.Words, &entry.Programs,
			&entry.AIAutomations, &entry.Level, &entry.Rank, &entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating leaderboard rows: %w", err)
	}

	return entries, nil
}

// UpdateLeaderboard updates a player's leaderboard entry
func (s *SQLiteDB) UpdateLeaderboard(entry *LeaderboardEntry) error {
	query := `
	INSERT OR REPLACE INTO leaderboard_entries (
		player_id, username, keystrokes_per_second, total_keystrokes,
		words, programs, ai_automations, level, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		entry.PlayerID, entry.Username, entry.KeystrokesPerSec,
		entry.TotalKeystrokes, entry.Words, entry.Programs,
		entry.AIAutomations, entry.Level, entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update leaderboard: %w", err)
	}

	return nil
}

// GetPlayerRank gets a player's rank on the leaderboard
func (s *SQLiteDB) GetPlayerRank(playerID string) (int, error) {
	query := `
	SELECT COUNT(*) + 1 as rank
	FROM leaderboard_entries le
	WHERE le.total_keystrokes > (
		SELECT total_keystrokes FROM leaderboard_entries WHERE player_id = ?
	)
	`

	var rank int
	err := s.db.QueryRow(query, playerID).Scan(&rank)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("player not found on leaderboard: %s", playerID)
		}
		return 0, fmt.Errorf("failed to get player rank: %w", err)
	}

	return rank, nil
}
