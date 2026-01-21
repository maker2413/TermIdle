package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/wish"
	"github.com/gliderlabs/ssh"
)

// Server represents the SSH server that handles player connections
type Server struct {
	config   *Config
	sessions map[string]*Session
	mu       sync.RWMutex
	logger   *log.Logger
}

// Config holds SSH server configuration
type Config struct {
	Port        int    `yaml:"port" koanf:"port"`
	HostKeyFile string `yaml:"host_key_file" koanf:"host_key_file"`
	MaxSessions int    `yaml:"max_sessions" koanf:"max_sessions"`
}

// DefaultConfig returns a default SSH server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:        2222,
		HostKeyFile: "./ssh_host_key",
		MaxSessions: 100,
	}
}

// NewServer creates a new SSH server instance
func NewServer(config *Config, logger *log.Logger) (*Server, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if logger == nil {
		logger = log.Default()
	}

	return &Server{
		config:   config,
		sessions: make(map[string]*Session),
		logger:   logger,
	}, nil
}

// StartSSHServer creates and starts the SSH server with wish
func StartSSHServer(config *Config) error {
	logger := log.Default()

	server, err := NewServer(config, logger)
	if err != nil {
		return fmt.Errorf("failed to create SSH server: %w", err)
	}

	// Generate or load host key
	hostKey, err := server.getOrGenerateHostKey()
	if err != nil {
		return fmt.Errorf("failed to setup host key: %w", err)
	}

	// Create wish server with middleware chain
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf(":%d", config.Port)),
		wish.WithHostKeyPEM(hostKey),
		wish.WithMiddleware(
			server.authMiddleware,
			server.sessionMiddleware,
			server.gameMiddleware,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create wish server: %w", err)
	}

	logger.Infof("Starting SSH server on port %d", config.Port)
	logger.Infof("Max sessions allowed: %d", config.MaxSessions)

	return s.ListenAndServe()
}

// getOrGenerateHostKey loads existing host key or generates a new one
func (s *Server) getOrGenerateHostKey() ([]byte, error) {
	// Generate a new ED25519 key pair
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate host key: %w", err)
	}

	// Convert to PEM format for OpenSSH
	pemKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKey,
	})

	s.logger.Info("Generated new SSH host key")
	return pemKey, nil
}

// AddSession adds a new session to the server
func (s *Server) AddSession(sessionID string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = session
	s.logger.Infof("Session %s added. Total sessions: %d", sessionID, len(s.sessions))
}

// RemoveSession removes a session from the server
func (s *Server) RemoveSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		// Cleanup session resources
		session.Cleanup()
		delete(s.sessions, sessionID)
		s.logger.Infof("Session %s removed. Total sessions: %d", sessionID, len(s.sessions))
	}
}

// GetSessionCount returns the current number of active sessions
func (s *Server) GetSessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.sessions)
}

// IsAtCapacity checks if the server has reached maximum sessions
func (s *Server) IsAtCapacity() bool {
	return s.GetSessionCount() >= s.config.MaxSessions
}

// GetLogger returns the server logger
func (s *Server) GetLogger() *log.Logger {
	return s.logger
}

// validatePublicKey checks if a public key is valid
func (s *Server) validatePublicKey(pk ssh.PublicKey) bool {
	if pk == nil {
		return false
	}

	// For now, accept any non-nil public key
	// In production, you might want to validate specific key types
	return true
}
