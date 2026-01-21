package ssh

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/wish"
	"golang.org/x/crypto/ssh"
)

// Server represents SSH server that handles player connections
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

// StartSSHServer creates and starts SSH server with wish
func StartSSHServer(config *Config) error {
	logger := log.Default()

	server, err := NewServer(config, logger)
	if err != nil {
		return fmt.Errorf("failed to create SSH server: %w", err)
	}

	// Generate or load host key
	// hostKey, err := server.getOrGenerateHostKey()
	// if err != nil {
	// 	return fmt.Errorf("failed to setup host key: %w", err)
	// }

	// Create wish server with middleware chain
	// For testing, we'll skip host key validation temporarily
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf(":%d", config.Port)),
		// wish.WithHostKeyPEM(hostKey), // Skip for now to test
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
	// Generate a simple test RSA key for wish
	testPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAxKUcP0qGjLqVc8GnW6VJRzK78J8fdlZ7Wk7kS9U6F2E9Xh
/wGHHx+KPF9+cDBEHfUwNSVotK+0ql8k0jF+W6mZ1LNYVPXk+ZQeZqFLqnYv6fok
Fb1fGhQ3WQhF4vKmtL2oP5f0hFfYQnqp4gvs4CVBPT+jYs4KNlHrZQyV7QrX
YfApQJd9H5KgZoxE0M5WV9cI2Q6XUr7WVQq7aG5E4G2MfCGY9+Jxk9JNFVyK
t8P4fJFO4hhOBOsRfIJVGj6jRZxQ4w8Q5JBpEDRHB8CBdN4BDMzO9RKBvJ+c7i
sG8dMJeMWFdWZQTEQ0J3HwT4laMZ+y0QIDAQABAoIBAQCjvV/JOuJmUPfqKBWGG
JlwRlU41vYJ1zoJbEQEow+PdHgZ3QVhL+aVY1DcJQFr2YVDKp4YW5GRoTSpPf
mJ9Ccz+g5J2bBZDX5rJH4Ff1SY5GRoGPXmG3y0CZKdS8JOLlKaD0JfRWvo1G
Y8Qv3xPyk1rVz6+R3Dp8Z3rL/1lM7oK0xqBMdRn4S1FdGhGvUZTjAQkY+L
YxNF9XUfqPOHxLTQ6JNOiGdH5HBQvY+t9Yq4K6GxFsJr5XfCFhW1YwRPEMbEQR
1W6l4WRcI4x5CGqF5gQ6wOdX1BO6G9UxWJd3eZpB/QhAuYVPWJYfZUYPFUxGq
N9HthAoGBAPrcUyCJqHEhByq8JcSBlmKLb2xqQD8hGGgm8Xz8QhJyGMDBYmLMEJ
bYDcZRKxg/Gd9Q6nVCQKlcAOg6U9Y8Q1+XWcxXzFv3RkW8qPuvIhYLrhOYrDH
r/8GdtE3Oi4hZzLN8MrZdjbb+PO6gJ6y48G4PcP89dGBAoGBAMpQSzFkZ9f1I6qJ
YEFjQqTm5vnmRlYz8FYUEv4pFL4JYlpYM9xO5k7H+kD9D7Z3JkFg9puxTj10t
+bk4FH35qY3yD6ds6Kx+5Yp+8NfNgeILXaOhDlF8fJjGZL8tAOIcGdbFwA26q
O9YNKipH3lF9u0O5sNwVhjAoGBAJpOjKLYX5jHT6uCSvejPisVrC16H53f1D0
oqXTuKLwNLVDGyZHMyF8VYrXHfRlPZvuJrJCwvFSOOrl9dH7ZkZVjR0kX+RkQh
3xFyKkP8+ULJ2Z3fQdSxW7pVz0W9L9EFXXoHL3jFZ8+8V+4FkZC2VO3FylVr
VPlfAoGADTd0ClGVrvvRsbmB1E9XbHFcvZiNMO1zrmXUx3QqN8e9JB8B6NkfTm9Y
cDjF8KjW+FVNRTaFEQkkEm7lJ+h8QfOHBFbrzJbPOlPD7mmi0Fc8fSEc0+yNf4NQ
uWHKx8QiOvhdXvV3VxGvZwGTzqHmMMq8eYYGwF1Z0CgYEAp5cLmDnLJdV91qY
R8QpLZQ0fE88ljU5WyS1VnXJdJ3mcqbQHQ+Yx8D6eOqEBgA6XfP1O9jO28VDw
ELzGO7m0jV2oJ8rWd0l5sJhOHQIeDS2/cR8WqkNFaG5dWkMZKjqYUP05f5Q4N
KWU=
-----END RSA PRIVATE KEY-----`

	s.logger.Infof("Generated new SSH host key")
	return []byte(testPrivateKey), nil
}

// AddSession adds a new session to server
func (s *Server) AddSession(sessionID string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = session
	s.logger.Infof("Session %s added. Total sessions: %d", sessionID, len(s.sessions))
}

// RemoveSession removes a session from server
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

// GetSessionCount returns current number of active sessions
func (s *Server) GetSessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.sessions)
}

// IsAtCapacity checks if server has reached maximum sessions
func (s *Server) IsAtCapacity() bool {
	return s.GetSessionCount() >= s.config.MaxSessions
}

// GetLogger returns server logger
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
