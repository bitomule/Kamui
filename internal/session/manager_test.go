package session

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bitomule/kamui/internal/claude"
	"github.com/bitomule/kamui/internal/storage"
	"github.com/bitomule/kamui/pkg/types"
)

// MockClaudeClient is a mock implementation of claude.ClientInterface
type MockClaudeClient struct {
	mock.Mock
}

func (m *MockClaudeClient) HasSession(sessionID, workingDir string) (bool, error) {
	args := m.Called(sessionID, workingDir)
	return args.Bool(0), args.Error(1)
}

func (m *MockClaudeClient) StartSession(workingDir string) (string, error) {
	args := m.Called(workingDir)
	return args.String(0), args.Error(1)
}

func (m *MockClaudeClient) ResumeSession(sessionID, workingDir string) error {
	args := m.Called(sessionID, workingDir)
	return args.Error(0)
}

func (m *MockClaudeClient) ListSessions() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sessions, ok := args.Get(0).([]string)
	if !ok {
		return nil, args.Error(1)
	}
	return sessions, args.Error(1)
}

func (m *MockClaudeClient) GetSessionInfo(sessionID, workingDir string) (*claude.SessionInfo, error) {
	args := m.Called(sessionID, workingDir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	info, ok := args.Get(0).(*claude.SessionInfo)
	if !ok {
		return nil, args.Error(1)
	}
	return info, args.Error(1)
}

func (m *MockClaudeClient) TerminateSession(sessionID, workingDir string) error {
	args := m.Called(sessionID, workingDir)
	return args.Error(0)
}

func (m *MockClaudeClient) DiscoverExistingSessions(workingDir string) ([]string, error) {
	args := m.Called(workingDir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sessions, ok := args.Get(0).([]string)
	if !ok {
		return nil, args.Error(1)
	}
	return sessions, args.Error(1)
}

func (m *MockClaudeClient) DiscoverNewestSession(workingDir string) (string, error) {
	args := m.Called(workingDir)
	return args.String(0), args.Error(1)
}

func (m *MockClaudeClient) LaunchClaudeInteractively(workingDir string, sessionName string) error {
	args := m.Called(workingDir, sessionName)
	return args.Error(0)
}

func TestNewWithClient(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}

	manager, err := NewWithClient(tempDir, mockClient)
	require.NoError(t, err)

	assert.Equal(t, tempDir, manager.projectPath)
	assert.Equal(t, mockClient, manager.claudeClient)
	assert.NotNil(t, manager.storage)
}

func TestNewWithClientInvalidPath(t *testing.T) {
	invalidPath := "/nonexistent/path"
	mockClient := &MockClaudeClient{}

	_, err := NewWithClient(invalidPath, mockClient)
	require.Error(t, err)

	var agxErr *types.AGXError
	require.ErrorAs(t, err, &agxErr)
	assert.Equal(t, types.ErrCodeProjectNotFound, agxErr.Code)
}

func TestCreateOrResumeSession_NewSession(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	sessionName := "new-session"

	// Mock expectations for new session (no existing sessions)
	// HasSession should return false for the stored session check
	mockClient.On("HasSession", "", tempDir).Return(false, nil).Maybe()
	// LaunchClaudeInteractively should be called to create new session
	mockClient.On("LaunchClaudeInteractively", tempDir, sessionName).Return(nil)

	session, claudeWasExecuted, err := manager.CreateOrResumeSession(sessionName)
	require.NoError(t, err)

	assert.Equal(t, sessionName, session.SessionID)
	assert.True(t, claudeWasExecuted) // New session should execute Claude
	assert.Equal(t, tempDir, session.Project.Path)

	mockClient.AssertExpectations(t)
}

func TestCreateOrResumeSession_ResumeExisting(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	sessionName := "existing-session"
	claudeSessionID := "claude-existing-123"

	// Create existing session first
	session, err := testStorage.CreateSession(sessionName, tempDir)
	require.NoError(t, err)
	session.Claude.SessionID = claudeSessionID
	err = testStorage.SaveSession(session)
	require.NoError(t, err)

	// Mock expectations for resuming existing session
	mockClient.On("HasSession", claudeSessionID, tempDir).Return(true, nil)

	resumedSession, claudeWasExecuted, err := manager.CreateOrResumeSession(sessionName)
	require.NoError(t, err)

	assert.Equal(t, sessionName, resumedSession.SessionID)
	assert.False(t, claudeWasExecuted) // Existing session should not execute Claude again
	assert.Equal(t, claudeSessionID, resumedSession.Claude.SessionID)
	assert.True(t, resumedSession.LastAccessed.After(session.LastAccessed))

	mockClient.AssertExpectations(t)
}

func TestCreateOrResumeSession_StoredSessionMissing(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	sessionName := "session-with-missing-claude"
	claudeSessionID := "claude-missing-123"

	// Create existing session with a Claude session ID
	session, err := testStorage.CreateSession(sessionName, tempDir)
	require.NoError(t, err)
	session.Claude.SessionID = claudeSessionID
	err = testStorage.SaveSession(session)
	require.NoError(t, err)

	// Mock expectations - stored Claude session no longer exists
	mockClient.On("HasSession", claudeSessionID, tempDir).Return(false, nil)
	mockClient.On("LaunchClaudeInteractively", tempDir, sessionName).Return(nil)

	resumedSession, claudeWasExecuted, err := manager.CreateOrResumeSession(sessionName)
	require.NoError(t, err)

	assert.Equal(t, sessionName, resumedSession.SessionID)
	assert.True(t, claudeWasExecuted) // Should execute Claude since stored session was missing

	mockClient.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	sessionName := "test-session"

	// Create and save a session
	originalSession, err := testStorage.CreateSession(sessionName, tempDir)
	require.NoError(t, err)
	err = testStorage.SaveSession(originalSession)
	require.NoError(t, err)

	// Retrieve the session
	retrievedSession, err := manager.GetSession(sessionName)
	require.NoError(t, err)

	assert.Equal(t, sessionName, retrievedSession.SessionID)
	assert.True(t, originalSession.Created.Equal(retrievedSession.Created))
}

func TestGetSessionNotFound(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	_, err = manager.GetSession("nonexistent")
	require.Error(t, err)

	var agxErr *types.AGXError
	require.ErrorAs(t, err, &agxErr)
	assert.Equal(t, types.ErrCodeSessionNotFound, agxErr.Code)
}

func TestListSessions(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	// Should be empty initially
	sessions, err := manager.ListSessions()
	require.NoError(t, err)
	assert.Empty(t, sessions)

	// Create some sessions
	sessionNames := []string{"session1", "session2", "session3"}
	for _, name := range sessionNames {
		session, createErr := testStorage.CreateSession(name, tempDir)
		require.NoError(t, createErr)
		saveErr := testStorage.SaveSession(session)
		require.NoError(t, saveErr)
	}

	// List should return all sessions
	sessions, err = manager.ListSessions()
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

func TestCompleteSession(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	sessionName := "test-session"

	// Create and save a session
	session, err := testStorage.CreateSession(sessionName, tempDir)
	require.NoError(t, err)
	err = testStorage.SaveSession(session)
	require.NoError(t, err)

	// Complete the session
	err = manager.CompleteSession(sessionName)
	require.NoError(t, err)

	// Verify session state changed
	completedSession, err := manager.GetSession(sessionName)
	require.NoError(t, err)

	assert.Equal(t, types.SessionStateCompleted, completedSession.Lifecycle.State)
	assert.Len(t, completedSession.Lifecycle.StateHistory, 2) // Initial + completed
	assert.Equal(t, types.SessionStateCompleted, completedSession.Lifecycle.StateHistory[1].State)
	assert.Equal(t, "manually_completed", completedSession.Lifecycle.StateHistory[1].Reason)
}

func TestDeleteSession(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	sessionName := "test-session"

	// Create and save a session
	session, err := testStorage.CreateSession(sessionName, tempDir)
	require.NoError(t, err)
	err = testStorage.SaveSession(session)
	require.NoError(t, err)

	// Verify it exists
	sessions, err := manager.ListSessions()
	require.NoError(t, err)
	assert.Contains(t, sessions, sessionName)

	// Delete the session
	err = manager.DeleteSession(sessionName)
	require.NoError(t, err)

	// Verify it no longer exists
	sessions, err = manager.ListSessions()
	require.NoError(t, err)
	assert.NotContains(t, sessions, sessionName)
}

func TestGetProjectPath(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	assert.Equal(t, tempDir, manager.GetProjectPath())
}

func TestGetProjectName(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	expectedName := filepath.Base(tempDir)
	assert.Equal(t, expectedName, manager.GetProjectName())
}

func TestGetClaudeCommand(t *testing.T) {
	tempDir := t.TempDir()
	mockClient := &MockClaudeClient{}
	testStorage := storage.NewWithSessionsDir(tempDir, filepath.Join(tempDir, ".claude", "kamui-sessions"))

	manager, err := NewWithDependencies(tempDir, testStorage, mockClient)
	require.NoError(t, err)

	// Test with empty Claude session ID
	session, err := testStorage.CreateSession("test-session", tempDir)
	require.NoError(t, err)

	command := manager.GetClaudeCommand(session)
	assert.Equal(t, "claude", command)

	// Test with Claude session ID
	session.Claude.SessionID = "claude-123456"
	command = manager.GetClaudeCommand(session)
	assert.Equal(t, "claude --resume claude-123456", command)
}
