package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 2222, config.Port)
	assert.Equal(t, "./ssh_host_key", config.HostKeyFile)
	assert.Equal(t, 100, config.MaxSessions)
}

func TestNewServer(t *testing.T) {
	config := &Config{
		Port:        2223,
		HostKeyFile: "/tmp/test_key",
		MaxSessions: 50,
	}

	server, err := NewServer(config, nil)
	require.NoError(t, err)
	require.NotNil(t, server)

	assert.Equal(t, config, server.config)
	assert.NotNil(t, server.sessions)
	assert.NotNil(t, server.logger)
}

func TestNewServerWithDefaults(t *testing.T) {
	server, err := NewServer(nil, nil)
	require.NoError(t, err)
	require.NotNil(t, server)

	assert.Equal(t, DefaultConfig(), server.config)
}

func TestServerSessionManagement(t *testing.T) {
	server, err := NewServer(nil, nil)
	require.NoError(t, err)

	// Initially should have no sessions
	assert.Equal(t, 0, server.GetSessionCount())
	assert.False(t, server.IsAtCapacity())

	// Add a session with proper logger
	session := NewSession("test_session", "test_player", "testuser", nil, server.logger)
	server.AddSession("test_session", session)
	assert.Equal(t, 1, server.GetSessionCount())

	// Remove session
	server.RemoveSession("test_session")
	assert.Equal(t, 0, server.GetSessionCount())
}

func TestValidatePublicKey(t *testing.T) {
	server, err := NewServer(nil, nil)
	require.NoError(t, err)

	// Test with nil key
	assert.False(t, server.validatePublicKey(nil))
}

func TestIsAtCapacity(t *testing.T) {
	config := &Config{MaxSessions: 1}
	server, err := NewServer(config, nil)
	require.NoError(t, err)

	assert.False(t, server.IsAtCapacity())

	// Add a session to reach capacity
	session := NewSession("test_session", "test_player", "testuser", nil, server.logger)
	server.AddSession("test_session", session)
	assert.True(t, server.IsAtCapacity())
}
