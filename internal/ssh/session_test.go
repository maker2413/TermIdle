package ssh

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSession(t *testing.T) {
	sessionID := "test_session_123"
	playerID := "player_456"
	username := "testuser"

	// Create a mock SSH session (we can't easily create a real one in tests)
	// So we'll test the parts that don't require a real SSH session
	server, _ := NewServer(nil, nil)
	logger := server.logger

	session := NewSession(sessionID, playerID, username, nil, logger)

	require.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, playerID, session.PlayerID)
	assert.Equal(t, username, session.Username)
	assert.NotNil(t, session.logger)
	assert.NotNil(t, session.cancel)
	assert.NotNil(t, session.ctx)
}

func TestSessionIsActive(t *testing.T) {
	session := &Session{
		cancel: func() {},
		ctx:    context.Background(), // Use a proper context
	}

	// Test that session is active
	assert.True(t, session.IsActive())
}

func TestSessionUpdateLastActive(t *testing.T) {
	session := &Session{}
	lastActive := session.LastActive

	// Wait a bit to ensure different timestamp
	time.Sleep(1 * time.Millisecond)

	session.UpdateLastActive()
	assert.True(t, session.LastActive.After(lastActive))
}

func TestSessionGetDuration(t *testing.T) {
	now := time.Now()
	session := &Session{CreatedAt: now}

	// Duration should be very small (almost zero)
	duration := session.GetDuration()
	assert.True(t, duration >= 0)
	assert.True(t, duration < 100*time.Millisecond)
}

func TestSessionGetIdleTime(t *testing.T) {
	now := time.Now()
	session := &Session{LastActive: now}

	// Idle time should be very small (almost zero)
	idleTime := session.GetIdleTime()
	assert.True(t, idleTime >= 0)
	assert.True(t, idleTime < 100*time.Millisecond)
}

func TestSessionGetSessionInfo(t *testing.T) {
	session := NewSession("test_session", "player_123", "testuser", nil, nil)

	sessionInfo := session.GetSessionInfo()

	assert.NotNil(t, sessionInfo)
	assert.Equal(t, "test_session", sessionInfo["id"])
	assert.Equal(t, "player_123", sessionInfo["player_id"])
	assert.Equal(t, "testuser", sessionInfo["username"])
	assert.Contains(t, sessionInfo, "created_at")
	assert.Contains(t, sessionInfo, "last_active")
	assert.Contains(t, sessionInfo, "duration")
	assert.Contains(t, sessionInfo, "idle_time")
	assert.Contains(t, sessionInfo, "is_active")
}
