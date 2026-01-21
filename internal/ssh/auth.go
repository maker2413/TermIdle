package ssh

import (
	"fmt"

	"github.com/gliderlabs/ssh"
)

// authMiddleware handles SSH key authentication
func (s *Server) authMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		username := sess.User()
		pk := sess.PublicKey()

		s.logger.Infof("Authentication attempt for user %s", username)

		// Validate public key format
		if !s.validatePublicKey(pk) {
			s.logger.Warnf("Invalid SSH key for user %s", username)
			_ = sess.Exit(1)
			return
		}

		// For now, accept any valid SSH key (simplified for demo)
		// In production, you would verify against a database
		playerID := fmt.Sprintf("player_%s", username)

		s.logger.Infof("Authentication successful for user %s (player: %s)", username, playerID)

		// Set player context for downstream middleware
		sess.Context().SetValue("player_id", playerID)
		sess.Context().SetValue("username", username)

		next(sess)
	}
}

// sessionMiddleware creates and manages player sessions
func (s *Server) sessionMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		playerID := sess.Context().Value("player_id").(string)
		username := sess.Context().Value("username").(string)

		// Check server capacity
		if s.IsAtCapacity() {
			s.logger.Warnf("Server at capacity. Rejecting connection from %s", username)
			_ = sess.Exit(1)
			return
		}

		// Generate unique session ID
		sessionID := fmt.Sprintf("%s_%d", playerID, sess.Context().Value("session_id"))

		// Create new session
		session := NewSession(sessionID, playerID, username, sess, s.logger)

		// Add to server session map
		s.AddSession(sessionID, session)

		// Set session context
		sess.Context().SetValue("session", session)
		sess.Context().SetValue("session_id", sessionID)

		s.logger.Infof("Session %s created for user %s", sessionID, username)

		// Handle session cleanup on exit
		defer func() {
			s.RemoveSession(sessionID)
		}()

		next(sess)
	}
}

// gameMiddleware initializes and runs the game for authenticated sessions
func (s *Server) gameMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		session := sess.Context().Value("session").(*Session)

		// Initialize game state
		if err := session.InitializeGame(); err != nil {
			s.logger.Errorf("Failed to initialize game for session %s: %v", session.ID, err)
			_ = sess.Exit(1)
			return
		}

		// Start game
		if err := session.Start(); err != nil {
			s.logger.Errorf("Failed to start game for session %s: %v", session.ID, err)
			_ = sess.Exit(1)
			return
		}

		// Start the game
		if err := session.Start(); err != nil {
			s.logger.Errorf("Failed to start game for session %s: %v", session.ID, err)
			_ = sess.Exit(1)
			return
		}

		s.logger.Infof("Game started for session %s", session.ID)

		// The game runs in the background, so we just wait for the session to end
		next(sess)
	}
}
