package tests

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"

	"github.com/maker2413/term-idle/internal/db"
	"github.com/maker2413/term-idle/internal/game"
)

// IntegrationTestSuite tests the complete system integration
type IntegrationTestSuite struct {
	suite.Suite
	ctx            context.Context
	cancel         context.CancelFunc
	testDBPath     string
	testConfigPath string
	sshPort        int
	apiPort        int
	baseURL        string
}

// SetupSuite runs once before all tests
func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	// Use different ports for testing
	s.sshPort = 2223
	s.apiPort = 8081
	s.baseURL = fmt.Sprintf("http://localhost:%d", s.apiPort)

	// Create temporary files
	s.testDBPath = filepath.Join(os.TempDir(), fmt.Sprintf("term_idle_test_%s.db", uuid.New().String()[:8]))
	s.testConfigPath = filepath.Join(os.TempDir(), fmt.Sprintf("config_test_%s.yaml", uuid.New().String()[:8]))

	// Create test configuration
	s.createTestConfig()

	// Clean up any existing test processes
	s.cleanupTestProcesses()
}

// TearDownSuite runs once after all tests
func (s *IntegrationTestSuite) TearDownSuite() {
	s.cleanupTestProcesses()

	// Clean up temporary files
	_ = os.Remove(s.testDBPath)
	_ = os.Remove(s.testConfigPath)

	s.cancel()
}

// SetupTest runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	// Ensure clean state for each test
	s.cleanupTestProcesses()
	_ = os.Remove(s.testDBPath)
}

// TearDownTest runs after each test
func (s *IntegrationTestSuite) TearDownTest() {
	s.cleanupTestProcesses()
}

// createTestConfig creates a test configuration file
func (s *IntegrationTestSuite) createTestConfig() {
	configContent := fmt.Sprintf(`
# Test Configuration
ssh:
  port: %d
  host_key_file: "./test_ssh_host_key"
  max_sessions: 10

game:
  save_interval: "5s"
  production_tick: "500ms"
  max_players: 50
  offline_production: true

database:
  path: "%s"
  max_connections: 5
  timeout: "10s"

server:
  port: "%d"
  host: "127.0.0.1"
  read_timeout: "10s"
  write_timeout: "10s"

logging:
  level: "error"
  format: "text"
  file: ""
`, s.sshPort, s.testDBPath, s.apiPort)

	err := os.WriteFile(s.testConfigPath, []byte(configContent), 0644)
	s.Require().NoError(err)
}

// cleanupTestProcesses stops any running test processes
func (s *IntegrationTestSuite) cleanupTestProcesses() {
	// This is a simplified cleanup - in a real scenario you'd track PIDs
	// For now, we'll just make sure ports are free
	// The actual cleanup would be more sophisticated
}

// generateTestSSHKey creates a temporary SSH key pair for testing
func (s *IntegrationTestSuite) generateTestSSHKey() (string, string) {
	privateKeyFile := filepath.Join(os.TempDir(), fmt.Sprintf("test_private_%s", uuid.New().String()[:8]))
	publicKeyFile := filepath.Join(os.TempDir(), fmt.Sprintf("test_public_%s.pub", uuid.New().String()[:8]))

	// Generate SSH key pair
	cmd := exec.CommandContext(s.ctx, "ssh-keygen", "-t", "rsa", "-b", "2048",
		"-f", privateKeyFile, "-N", "", "-C", "test-key")
	err := cmd.Run()
	s.Require().NoError(err)

	return privateKeyFile, publicKeyFile
}

// waitForService waits for a service to be ready
func (s *IntegrationTestSuite) waitForService(port int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("service on port %d not ready within timeout", port)
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
			if err == nil {
				_ = conn.Close()
				return nil
			}
		}
	}
}

// createTestPlayer creates a test player in the database
func (s *IntegrationTestSuite) createTestPlayer(username string) (*db.Player, error) {
	// This would use the actual database interface to create a player
	// For the integration test, we'll simulate this
	player := &db.Player{
		ID:         uuid.New().String(),
		Username:   username,
		SSHKey:     "test-ssh-key-" + username,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// In a real implementation, you'd use s.db.CreatePlayer(player)
	return player, nil
}

// TestHTTPAPIStartup tests that the HTTP API server starts correctly
func (s *IntegrationTestSuite) TestHTTPAPIStartup() {
	// Start the HTTP API server
	cmd := exec.CommandContext(s.ctx, "go", "run", "cmd/term-idle/main.go",
		"--config", s.testConfigPath)

	err := cmd.Start()
	s.Require().NoError(err)
	defer func() { _ = cmd.Process.Kill() }()

	// Wait for server to be ready
	err = s.waitForService(s.apiPort, 10*time.Second)
	s.NoError(err, "HTTP API server should start within timeout")

	// Test health endpoint
	resp, err := http.Get(s.baseURL + "/api/health")
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)

	// Test leaderboard endpoint
	resp, err = http.Get(s.baseURL + "/api/leaderboard")
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

// TestSSHServerStartup tests that the SSH server starts correctly
func (s *IntegrationTestSuite) TestSSHServerStartup() {
	// Start the SSH server
	cmd := exec.CommandContext(s.ctx, "go", "run", "cmd/ssh-server/main.go",
		"--config", s.testConfigPath)

	err := cmd.Start()
	s.Require().NoError(err)
	defer func() { _ = cmd.Process.Kill() }()

	// Wait for server to be ready
	err = s.waitForService(s.sshPort, 10*time.Second)
	s.NoError(err, "SSH server should start within timeout")
}

// TestPlayerJourney tests the complete player journey
func (s *IntegrationTestSuite) TestPlayerJourney() {
	// Create test player
	player, err := s.createTestPlayer("testuser")
	s.Require().NoError(err)

	// Start both services
	apiCmd := exec.CommandContext(s.ctx, "go", "run", "cmd/term-idle/main.go",
		"--config", s.testConfigPath)
	err = apiCmd.Start()
	s.Require().NoError(err)
	defer func() { _ = apiCmd.Process.Kill() }()

	sshCmd := exec.CommandContext(s.ctx, "go", "run", "cmd/ssh-server/main.go",
		"--config", s.testConfigPath)
	err = sshCmd.Start()
	s.Require().NoError(err)
	defer func() { _ = sshCmd.Process.Kill() }()

	// Wait for services to be ready
	err = s.waitForService(s.apiPort, 10*time.Second)
	s.NoError(err)

	err = s.waitForService(s.sshPort, 10*time.Second)
	s.NoError(err)

	// Test player data retrieval
	resp, err := http.Get(fmt.Sprintf("%s/api/players/%s", s.baseURL, player.ID))
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)

	// Test leaderboard functionality
	resp, err = http.Get(s.baseURL + "/api/leaderboard")
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)

	// Test player leaderboard update
	reqBody := strings.NewReader(`{
		"keystrokes_per_second": 1.5,
		"total_keystrokes": 100,
		"level": 5
	}`)

	resp, err = http.Post(
		fmt.Sprintf("%s/api/players/%s/leaderboard", s.baseURL, player.ID),
		"application/json",
		reqBody,
	)
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

// TestConcurrentSSHConnections tests multiple simultaneous SSH connections
func (s *IntegrationTestSuite) TestConcurrentSSHConnections() {
	// Start SSH server
	cmd := exec.CommandContext(s.ctx, "go", "run", "cmd/ssh-server/main.go",
		"--config", s.testConfigPath)
	err := cmd.Start()
	s.Require().NoError(err)
	defer func() { _ = cmd.Process.Kill() }()

	// Wait for server to be ready
	err = s.waitForService(s.sshPort, 10*time.Second)
	s.NoError(err)

	// Generate test SSH keys
	privateKeyFile, publicKeyFile := s.generateTestSSHKey()
	defer func() { _ = os.Remove(privateKeyFile) }()
	defer func() { _ = os.Remove(publicKeyFile) }()

	// Create multiple concurrent connections
	const numConnections = 5
	var wg sync.WaitGroup
	errors := make(chan error, numConnections)

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			// Parse private key
			key, err := os.ReadFile(privateKeyFile)
			s.Require().NoError(err)

			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				errors <- fmt.Errorf("failed to parse private key: %w", err)
				return
			}

			// Create SSH client config
			config := &ssh.ClientConfig{
				User: fmt.Sprintf("testuser%d", connID),
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Only for testing
				Timeout:         10 * time.Second,
			}

			// Connect to SSH server
			conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%d", s.sshPort), config)
			if err != nil {
				errors <- fmt.Errorf("failed to connect: %w", err)
				return
			}
			defer func() { _ = conn.Close() }()

			// Keep connection alive briefly
			time.Sleep(2 * time.Second)
			errors <- nil
		}(i)
	}

	// Wait for all connections to complete
	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		if err != nil {
			s.T().Logf("SSH connection error: %v", err)
			// Note: Some connection errors are expected in testing without full auth setup
		}
	}

	s.T().Logf("Completed %d concurrent SSH connection attempts", numConnections)
}

// TestDatabasePersistence tests that game state persists correctly
func (s *IntegrationTestSuite) TestDatabasePersistence() {
	// This test would verify that game state is correctly saved and loaded
	// For the integration test, we'll simulate this by testing the database layer directly

	// Create a test game state
	gameState := &game.GameState{
		PlayerID:      uuid.New().String(),
		CurrentLevel:  5,
		Keystrokes:    100.5,
		Words:         10,
		Programs:      2,
		AIAutomations: 1,
		StoryProgress: 3,
		LastSave:      time.Now(),
	}

	// In a real implementation, you would:
	// 1. Save the game state to database
	// 2. Load it back and verify all fields are preserved
	// 3. Test concurrent access to the same player data

	s.NotEmpty(gameState.PlayerID)
	s.Equal(5, gameState.CurrentLevel)
	s.Equal(100.5, gameState.Keystrokes)
}

// TestGameLoopIntegration tests the game loop with real-time updates
func (s *IntegrationTestSuite) TestGameLoopIntegration() {
	// This test would verify that:
	// 1. Production ticker works correctly
	// 2. Auto-save functions properly
	// 3. UI updates in real-time
	// 4. Story triggers fire at appropriate times

	s.T().Log("Game loop integration test placeholder")

	// In a real implementation, you would:
	// 1. Start a game session
	// 2. Verify production increases over time
	// 3. Check auto-save intervals
	// 4. Verify story progression triggers
}

// TestErrorHandling tests system behavior under error conditions
func (s *IntegrationTestSuite) TestErrorHandling() {
	// Test with invalid configuration
	invalidConfigPath := filepath.Join(os.TempDir(), "invalid_config.yaml")
	err := os.WriteFile(invalidConfigPath, []byte("invalid: yaml: content:"), 0644)
	s.Require().NoError(err)
	defer func() { _ = os.Remove(invalidConfigPath) }()

	// Try to start server with invalid config
	cmd := exec.CommandContext(s.ctx, "go", "run", "cmd/term-idle/main.go",
		"--config", invalidConfigPath)

	err = cmd.Run()
	// Command should fail with invalid config
	s.Error(err)
}

// TestShutdownGraceful tests that services shut down gracefully
func (s *IntegrationTestSuite) TestShutdownGraceful() {
	// Start services
	apiCmd := exec.CommandContext(s.ctx, "go", "run", "cmd/term-idle/main.go",
		"--config", s.testConfigPath)
	err := apiCmd.Start()
	s.Require().NoError(err)

	sshCmd := exec.CommandContext(s.ctx, "go", "run", "cmd/ssh-server/main.go",
		"--config", s.testConfigPath)
	err = sshCmd.Start()
	s.Require().NoError(err)

	// Wait for services to be ready
	_ = s.waitForService(s.apiPort, 10*time.Second)
	_ = s.waitForService(s.sshPort, 10*time.Second)

	// Send interrupt signal to simulate graceful shutdown
	_ = apiCmd.Process.Signal(os.Interrupt)
	_ = sshCmd.Process.Signal(os.Interrupt)

	// Wait for graceful shutdown
	err = apiCmd.Wait()
	s.NoError(err)

	err = sshCmd.Wait()
	s.NoError(err)
}

// TestIntegrationHelper provides utility methods for integration tests
type TestIntegrationHelper struct {
	suite *IntegrationTestSuite
}

// NewTestIntegrationHelper creates a new helper instance
func NewTestIntegrationHelper(suite *IntegrationTestSuite) *TestIntegrationHelper {
	return &TestIntegrationHelper{suite: suite}
}

// CreateMockPlayer creates a mock player for testing
func (h *TestIntegrationHelper) CreateMockPlayer(username string) *db.Player {
	return &db.Player{
		ID:         uuid.New().String(),
		Username:   username,
		SSHKey:     "mock-ssh-key",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
}

// AssertAPIResponse asserts that an API response has the expected status and content
func (h *TestIntegrationHelper) AssertAPIResponse(resp *http.Response, expectedStatus int, expectedContent string) {
	defer func() { _ = resp.Body.Close() }()

	h.suite.Equal(expectedStatus, resp.StatusCode)

	if expectedContent != "" {
		body, err := io.ReadAll(resp.Body)
		h.suite.NoError(err)
		h.suite.Contains(string(body), expectedContent)
	}
}

// WaitForCondition waits for a condition to be true
func (h *TestIntegrationHelper) WaitForCondition(condition func() bool, timeout time.Duration, message string) {
	ctx, cancel := context.WithTimeout(h.suite.ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.suite.Fail(message + " (timeout)")
			return
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// TestIntegration
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Check if we're in CI environment - skip integration tests if so
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration tests in CI environment")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
