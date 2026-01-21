package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_RejectsNilKey(t *testing.T) {
	server, err := NewServer(nil, nil)
	require.NoError(t, err)

	// Test that the server rejects nil key
	valid := server.validatePublicKey(nil)
	assert.False(t, valid)
}
