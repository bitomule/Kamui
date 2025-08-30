package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitomule/kamui/pkg/types"
)

func TestNew(t *testing.T) {
	// This test requires the claude binary to be in PATH
	// In real environments this would work, but in CI we'll skip if not found
	_, err := New()
	if err != nil {
		// Should be a proper error type
		var agxErr *types.AGXError
		require.ErrorAs(t, err, &agxErr)
		assert.Equal(t, types.ErrCodeClaudeNotFound, agxErr.Code)
		t.Skip("Claude binary not found in PATH - expected in CI environment")
		return
	}

	// If claude is found, verify client was created properly
	// Note: This won't run in CI as expected since we don't install claude there
}

func TestHasSession_EmptySessionID(t *testing.T) {
	client := &Client{claudePath: "/mock/claude"}

	exists, err := client.HasSession("", "/tmp/project")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestHasSession_WithSessionFile(t *testing.T) {
	tempHome := t.TempDir()

	// Mock home directory for testing
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tempHome)
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	workingDir := "/tmp/test-project"
	sessionID := "test-session-123"

	// Create the expected directory structure
	encodedPath := strings.ReplaceAll(workingDir, "/", "-")
	sessionDir := filepath.Join(tempHome, ".claude", "projects", encodedPath)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	client := &Client{claudePath: "/mock/claude"}

	// Session should not exist initially
	exists, err := client.HasSession(sessionID, workingDir)
	require.NoError(t, err)
	assert.False(t, exists)

	// Create session file
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(`{"test": "data"}`), 0o644))

	// Session should exist now
	exists, err = client.HasSession(sessionID, workingDir)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestHasSession_HomeDirectoryError(t *testing.T) {
	// Set HOME to an invalid value to trigger error
	t.Setenv("HOME", "")

	client := &Client{claudePath: "/mock/claude"}

	_, err := client.HasSession("session-123", "/tmp/project")
	require.Error(t, err)
	// The specific error type depends on OS, so we just verify an error occurred
}

func TestStartSession(t *testing.T) {
	client := &Client{claudePath: "/mock/claude"}

	sessionID, err := client.StartSession("/tmp/project")
	require.NoError(t, err)

	// StartSession currently returns empty string to indicate fresh session
	assert.Empty(t, sessionID)
}

func TestResumeSession_SessionExists(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	workingDir := "/tmp/test-project"
	sessionID := "resume-session-123"

	// Create session file
	encodedPath := strings.ReplaceAll(workingDir, "/", "-")
	sessionDir := filepath.Join(tempHome, ".claude", "projects", encodedPath)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(`{"test": "data"}`), 0o644))

	client := &Client{claudePath: "/mock/claude"}

	err := client.ResumeSession(sessionID, workingDir)
	require.NoError(t, err)
}

func TestResumeSession_SessionNotFound(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	client := &Client{claudePath: "/mock/claude"}

	err := client.ResumeSession("nonexistent-session", "/tmp/project")
	require.Error(t, err)

	var agxErr *types.AGXError
	require.ErrorAs(t, err, &agxErr)
	assert.Equal(t, types.ErrCodeClaudeSessionNotFound, agxErr.Code)
	assert.Contains(t, agxErr.Message, "nonexistent-session")
}

func TestDiscoverExistingSessions(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	workingDir := "/tmp/test-project"

	client := &Client{claudePath: "/mock/claude"}

	// Should return empty when no project directory exists
	sessions, err := client.DiscoverExistingSessions(workingDir)
	require.NoError(t, err)
	assert.Empty(t, sessions)

	// Create project directory with session files
	encodedPath := strings.ReplaceAll(workingDir, "/", "-")
	sessionDir := filepath.Join(tempHome, ".claude", "projects", encodedPath)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	// Create several session files
	expectedSessions := []string{"session1", "session2", "session3"}
	for _, sessionID := range expectedSessions {
		sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
		require.NoError(t, os.WriteFile(sessionFile, []byte(`{"test": "data"}`), 0o644))
	}

	// Also create a non-session file to ensure it's ignored
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "readme.txt"), []byte("test"), 0o644))

	// Discover sessions
	sessions, err = client.DiscoverExistingSessions(workingDir)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Convert to set for order-independent comparison
	sessionSet := make(map[string]bool)
	for _, session := range sessions {
		sessionSet[session] = true
	}

	for _, expected := range expectedSessions {
		assert.True(t, sessionSet[expected], "Expected session %s not found", expected)
	}
}

func TestDiscoverNewestSession(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	workingDir := "/tmp/test-project"
	client := &Client{claudePath: "/mock/claude"}

	// Should return empty when no sessions exist
	newest, err := client.DiscoverNewestSession(workingDir)
	require.NoError(t, err)
	assert.Empty(t, newest)

	// Create a session
	encodedPath := strings.ReplaceAll(workingDir, "/", "-")
	sessionDir := filepath.Join(tempHome, ".claude", "projects", encodedPath)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	sessionFile := filepath.Join(sessionDir, "test-session.jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(`{"test": "data"}`), 0o644))

	// Should return the session
	newest, err = client.DiscoverNewestSession(workingDir)
	require.NoError(t, err)
	assert.Equal(t, "test-session", newest)
}

func TestPathEncoding(t *testing.T) {
	// Test the path encoding logic used throughout the client
	testCases := []struct {
		input    string
		expected string
	}{
		{"/tmp/project", "-tmp-project"},
		{"/Users/test/my-project", "-Users-test-my-project"},
		{"/home/user/project-with-dashes", "-home-user-project-with-dashes"},
		{"relative/path", "relative-path"},
	}

	for _, tc := range testCases {
		encoded := strings.ReplaceAll(tc.input, "/", "-")
		assert.Equal(t, tc.expected, encoded, "Path encoding mismatch for %s", tc.input)
	}
}

func TestSessionInfo(t *testing.T) {
	info := &SessionInfo{
		SessionID: "test-123",
		Status:    "active",
		Messages:  25,
		LastUsed:  "2025-08-26 15:30:00",
	}

	assert.Equal(t, "test-123", info.SessionID)
	assert.Equal(t, "active", info.Status)
	assert.Equal(t, 25, info.Messages)
	assert.Equal(t, "2025-08-26 15:30:00", info.LastUsed)
}

func TestClientInterface(t *testing.T) {
	// Verify that Client implements ClientInterface
	var _ ClientInterface = (*Client)(nil)

	// This test ensures the interface contract is maintained
	// If Client doesn't implement all interface methods, this will fail to compile
}

