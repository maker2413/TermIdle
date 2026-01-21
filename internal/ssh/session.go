package ssh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/gliderlabs/ssh"
	"github.com/maker2413/term-idle/internal/game"
	"github.com/maker2413/term-idle/internal/ui"
)

// Session represents an active SSH game session
type Session struct {
	ID         string
	PlayerID   string
	Username   string
	GameState  *game.GameState
	Program    *tea.Program
	SSHSession ssh.Session
	LastActive time.Time
	CreatedAt  time.Time
	logger     *log.Logger
	cancel     context.CancelFunc
	ctx        context.Context
	mu         sync.RWMutex
}

// NewSession creates a new game session
func NewSession(sessionID, playerID, username string, sshSession ssh.Session, logger *log.Logger) *Session {
	ctx, cancel := context.WithCancel(context.Background())

	return &Session{
		ID:         sessionID,
		PlayerID:   playerID,
		Username:   username,
		SSHSession: sshSession,
		LastActive: time.Now(),
		CreatedAt:  time.Now(),
		logger:     logger,
		cancel:     cancel,
		ctx:        ctx,
	}
}

// InitializeGame sets up the game state and Bubbletea program
func (s *Session) InitializeGame() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create game state
	s.GameState = game.NewGameState(s.PlayerID)

	// Create UI model
	model := ui.NewModel(s.GameState)

	// Create Bubbletea program with SSH I/O
	s.Program = tea.NewProgram(
		model,
		tea.WithInput(s.SSHSession),
		tea.WithOutput(s.SSHSession),
		tea.WithAltScreen(),
	)

	s.logger.Infof("Game initialized for session %s (player: %s)", s.ID, s.PlayerID)
	return nil
}

// Start begins the game session
func (s *Session) Start() error {
	if s.Program == nil {
		return fmt.Errorf("game not initialized")
	}

	s.logger.Infof("Starting game session %s for player %s", s.ID, s.Username)

	// Run the game in a goroutine
	go func() {
		defer s.cancel()

		if _, err := s.Program.Run(); err != nil {
			s.logger.Errorf("Game session %s error: %v", s.ID, err)
		}

		s.logger.Infof("Game session %s ended for player %s", s.ID, s.Username)
	}()

	return nil
}

// Stop terminates the game session
func (s *Session) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	if s.Program != nil {
		s.Program.Quit()
	}

	s.logger.Infof("Game session %s stopped", s.ID)
}

// UpdateLastActive updates the session's last active time
func (s *Session) UpdateLastActive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActive = time.Now()
}

// IsActive checks if the session is still active
func (s *Session) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	select {
	case <-s.ctx.Done():
		return false
	default:
		return true
	}
}

// GetDuration returns how long the session has been active
func (s *Session) GetDuration() time.Duration {
	return time.Since(s.CreatedAt)
}

// GetIdleTime returns how long since the session was last active
func (s *Session) GetIdleTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.Since(s.LastActive)
}

// Cleanup performs cleanup when the session ends
func (s *Session) Cleanup() {
	s.Stop()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Additional cleanup if needed
	s.logger.Infof("Session %s cleaned up", s.ID)
}

// SendNotification sends a notification to the player (placeholder for future implementation)
func (s *Session) SendNotification(message string) error {
	s.logger.Infof("Notification sent to session %s: %s", s.ID, message)
	// TODO: Implement notification system
	return nil
}

// GetSessionInfo returns session information for logging/monitoring
func (s *Session) GetSessionInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"id":          s.ID,
		"player_id":   s.PlayerID,
		"username":    s.Username,
		"created_at":  s.CreatedAt,
		"last_active": s.LastActive,
		"duration":    s.GetDuration().String(),
		"idle_time":   s.GetIdleTime().String(),
		"is_active":   s.IsActive(),
	}
}
