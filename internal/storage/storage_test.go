package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitomule/kamui/pkg/types"
)

func TestNew(t *testing.T) {
	projectPath := "/tmp/test-project"
	storage := New(projectPath)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	expectedSessionsDir := filepath.Join(homeDir, ".claude", "kamui-sessions")

	assert.Equal(t, projectPath, storage.projectPath)
	assert.Equal(t, expectedSessionsDir, storage.sessionsDir)
}

func TestInitialize(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	err := storage.Initialize()
	require.NoError(t, err)

	info, err := os.Stat(sessionsDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}

func TestSessionExists(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	// Session should not exist initially
	exists := storage.SessionExists("test-session")
	assert.False(t, exists)

	// Create session directory and file
	require.NoError(t, os.MkdirAll(sessionsDir, 0o700))

	sessionFile := filepath.Join(sessionsDir, "test-session.json")
	require.NoError(t, os.WriteFile(sessionFile, []byte("{}"), 0o600))

	// Session should exist now
	exists = storage.SessionExists("test-session")
	assert.True(t, exists)
}

func TestCreateSession(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	sessionID := "test-session"
	projectPath := tempDir

	session, err := storage.CreateSession(sessionID, projectPath)
	require.NoError(t, err)

	// Verify session properties (simplified structure)
	assert.Equal(t, sessionID, session.SessionID)
	assert.Equal(t, "1.0.0", session.Version)
	assert.Equal(t, projectPath, session.Project.Path)
	assert.Equal(t, projectPath, session.Project.WorkingDirectory)

	// Verify timestamps are recent
	now := time.Now()
	assert.True(t, session.Created.After(now.Add(-time.Minute)))
	assert.True(t, session.LastAccessed.After(now.Add(-time.Minute)))
	assert.True(t, session.LastModified.After(now.Add(-time.Minute)))

	// Verify Claude session ID is initially empty
	assert.Equal(t, "", session.Claude.SessionID)
}

func TestSaveAndLoadSession(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	// Create a test session
	originalSession, err := storage.CreateSession("test-session", tempDir)
	require.NoError(t, err)

	// Modify some fields to test serialization
	originalSession.Claude.SessionID = "claude-session-123"
	originalSession.Claude.HasActiveContext = true
	originalSession.Metadata.Description = "Test session description"
	originalSession.Metadata.Tags = []string{"test", "example"}

	// Save the session
	err = storage.SaveSession(originalSession)
	require.NoError(t, err)

	// Load the session back
	loadedSession, err := storage.LoadSession("test-session")
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, originalSession.SessionID, loadedSession.SessionID)
	assert.Equal(t, originalSession.Version, loadedSession.Version)
	assert.Equal(t, originalSession.Claude.SessionID, loadedSession.Claude.SessionID)
	assert.Equal(t, originalSession.Claude.HasActiveContext, loadedSession.Claude.HasActiveContext)
	assert.Equal(t, originalSession.Metadata.Description, loadedSession.Metadata.Description)
	assert.Equal(t, originalSession.Metadata.Tags, loadedSession.Metadata.Tags)
	assert.Equal(t, originalSession.Project.Name, loadedSession.Project.Name)
	assert.Equal(t, originalSession.Lifecycle.State, loadedSession.Lifecycle.State)
}

func TestSaveSessionAtomic(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	session, err := storage.CreateSession("test-session", tempDir)
	require.NoError(t, err)

	// Save the session
	err = storage.SaveSession(session)
	require.NoError(t, err)

	// Verify the temp file was cleaned up
	entries, err := os.ReadDir(sessionsDir)
	require.NoError(t, err)

	// Should only have the session file, no temp files
	assert.Len(t, entries, 1)
	assert.Equal(t, "test-session.json", entries[0].Name())
}

func TestLoadSessionNotFound(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	// Try to load non-existent session
	_, err := storage.LoadSession("non-existent")
	require.Error(t, err)

	// Should be a storage error with correct code
	var agxErr *types.AGXError
	require.ErrorAs(t, err, &agxErr)
	assert.Equal(t, types.ErrCodeSessionNotFound, agxErr.Code)
}

func TestListSessions(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	// Initially should return empty slice
	sessions, err := storage.ListSessions()
	require.NoError(t, err)
	assert.Empty(t, sessions)

	// Create some sessions
	sessionNames := []string{"session1", "session2", "session3"}
	for _, name := range sessionNames {
		session, createErr := storage.CreateSession(name, tempDir)
		require.NoError(t, createErr)
		saveErr := storage.SaveSession(session)
		require.NoError(t, saveErr)
	}

	// List sessions should return all created sessions
	sessions, err = storage.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Convert to set for order-independent comparison
	sessionSet := make(map[string]bool)
	for _, session := range sessions {
		sessionSet[session] = true
	}

	for _, expectedName := range sessionNames {
		assert.True(t, sessionSet[expectedName], "Expected session %s not found", expectedName)
	}
}

func TestDeleteSession(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	// Create and save a session
	session, err := storage.CreateSession("test-session", tempDir)
	require.NoError(t, err)
	err = storage.SaveSession(session)
	require.NoError(t, err)

	// Verify it exists
	exists := storage.SessionExists("test-session")
	assert.True(t, exists)

	// Delete the session
	err = storage.DeleteSession("test-session")
	require.NoError(t, err)

	// Verify it no longer exists
	exists = storage.SessionExists("test-session")
	assert.False(t, exists)
}

func TestDeleteSessionNotFound(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	// Try to delete non-existent session
	err := storage.DeleteSession("non-existent")
	require.Error(t, err)

	// Should be a storage error with correct code
	var agxErr *types.AGXError
	require.ErrorAs(t, err, &agxErr)
	assert.Equal(t, types.ErrCodeSessionNotFound, agxErr.Code)
}

func TestUpdateSessionAccess(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	// Create and save a session
	session, err := storage.CreateSession("test-session", tempDir)
	require.NoError(t, err)
	originalAccessTime := session.LastAccessed
	err = storage.SaveSession(session)
	require.NoError(t, err)

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update session access
	err = storage.UpdateSessionAccess("test-session")
	require.NoError(t, err)

	// Reload and verify access time was updated
	updatedSession, err := storage.LoadSession("test-session")
	require.NoError(t, err)

	assert.True(t, updatedSession.LastAccessed.After(originalAccessTime))
}

func TestGetProjectPath(t *testing.T) {
	projectPath := "/tmp/test-project"
	storage := New(projectPath)

	assert.Equal(t, projectPath, storage.GetProjectPath())
}

func TestGetSessionsPath(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".claude", "kamui-sessions")
	storage := NewWithSessionsDir(tempDir, sessionsDir)

	assert.Equal(t, sessionsDir, storage.GetSessionsPath())
}
