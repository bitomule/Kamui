package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionStates(t *testing.T) {
	// Test all session states are defined
	states := []SessionState{
		SessionStateActive,
		SessionStatePaused,
		SessionStateCompleted,
		SessionStateArchived,
		SessionStateError,
	}

	expectedValues := []string{
		"active",
		"paused",
		"completed",
		"archived",
		"error",
	}

	for i, state := range states {
		assert.Equal(t, expectedValues[i], string(state))
	}
}

func TestSessionSerialization(t *testing.T) {
	now := time.Now()

	// Create a complete session structure
	session := Session{
		Version:      "1.0.0",
		SessionID:    "test-session-123",
		Created:      now,
		LastAccessed: now,
		LastModified: now,

		Project: ProjectInfo{
			Name:             "test-project",
			Path:             "/tmp/test-project",
			WorkingDirectory: "/tmp/test-project",
			GitBranch:        "main",
			GitCommit:        "abc123",
			GitRemote:        "origin",
		},

		Claude: ClaudeInfo{
			SessionID:        "claude-456789",
			ConversationID:   "conv-123",
			ModelUsed:        "claude-3-sonnet",
			HasActiveContext: true,
			LastInteraction:  now,
			ContextInfo: ContextInfo{
				MessageCount:    10,
				EstimatedTokens: 5000,
				LastCommand:     "test command",
				WorkingFiles:    []string{"file1.go", "file2.go"},
			},
			ResumeInfo: ResumeInfo{
				CanResume:         true,
				ResumeCommand:     "claude --resume claude-456789",
				LastResumeAttempt: &now,
				ResumeErrors:      []string{"error1", "error2"},
			},
		},

		Metadata: SessionMeta{
			Description: "Test session for unit tests",
			Tags:        []string{"test", "development"},
			Variant:     "main",
			IsDefault:   true,
			CustomData: map[string]interface{}{
				"custom_field": "custom_value",
				"number":       42,
			},
		},

		Stats: SessionStats{
			SessionCount:         5,
			TotalDuration:        "2h30m",
			AverageSessionLength: "30m",
			LastSessionDuration:  "45m",
			MostActiveDay:        "2025-08-26",
			CommandsExecuted:     150,
		},

		Lifecycle: LifecycleInfo{
			State: SessionStateActive,
			StateHistory: []StateChange{
				{
					State:     SessionStateActive,
					Timestamp: now,
					Reason:    "session_created",
				},
			},
			AutoCleanup: CleanupConfig{
				Enabled:           true,
				InactiveThreshold: "30d",
				LastCleanupCheck:  now,
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(session)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON unmarshaling
	var unmarshaled Session
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, session.SessionID, unmarshaled.SessionID)
	assert.Equal(t, session.Version, unmarshaled.Version)
	assert.True(t, session.Created.Equal(unmarshaled.Created))
	assert.Equal(t, session.Project.Name, unmarshaled.Project.Name)
	assert.Equal(t, session.Claude.SessionID, unmarshaled.Claude.SessionID)
	assert.Equal(t, session.Claude.HasActiveContext, unmarshaled.Claude.HasActiveContext)
	assert.Equal(t, session.Metadata.Description, unmarshaled.Metadata.Description)
	assert.Equal(t, session.Metadata.Tags, unmarshaled.Metadata.Tags)
	assert.Equal(t, session.Stats.SessionCount, unmarshaled.Stats.SessionCount)
	assert.Equal(t, session.Lifecycle.State, unmarshaled.Lifecycle.State)
	assert.Len(t, unmarshaled.Lifecycle.StateHistory, 1)
}

func TestStateChange(t *testing.T) {
	now := time.Now()

	stateChange := StateChange{
		State:     SessionStateCompleted,
		Timestamp: now,
		Reason:    "manually_completed",
	}

	// Test JSON serialization
	data, err := json.Marshal(stateChange)
	require.NoError(t, err)

	var unmarshaled StateChange
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, stateChange.State, unmarshaled.State)
	assert.True(t, stateChange.Timestamp.Equal(unmarshaled.Timestamp))
	assert.Equal(t, stateChange.Reason, unmarshaled.Reason)
}

func TestContextInfo(t *testing.T) {
	contextInfo := ContextInfo{
		MessageCount:    25,
		EstimatedTokens: 12000,
		LastCommand:     "go test",
		WorkingFiles:    []string{"main.go", "types.go", "session_test.go"},
	}

	// Test JSON serialization
	data, err := json.Marshal(contextInfo)
	require.NoError(t, err)

	var unmarshaled ContextInfo
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, contextInfo.MessageCount, unmarshaled.MessageCount)
	assert.Equal(t, contextInfo.EstimatedTokens, unmarshaled.EstimatedTokens)
	assert.Equal(t, contextInfo.LastCommand, unmarshaled.LastCommand)
	assert.Equal(t, contextInfo.WorkingFiles, unmarshaled.WorkingFiles)
}

func TestResumeInfo(t *testing.T) {
	now := time.Now()

	resumeInfo := ResumeInfo{
		CanResume:         true,
		ResumeCommand:     "claude --resume session-123",
		LastResumeAttempt: &now,
		ResumeErrors:      []string{"timeout", "connection failed"},
	}

	// Test JSON serialization
	data, err := json.Marshal(resumeInfo)
	require.NoError(t, err)

	var unmarshaled ResumeInfo
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, resumeInfo.CanResume, unmarshaled.CanResume)
	assert.Equal(t, resumeInfo.ResumeCommand, unmarshaled.ResumeCommand)
	require.NotNil(t, unmarshaled.LastResumeAttempt)
	assert.True(t, resumeInfo.LastResumeAttempt.Equal(*unmarshaled.LastResumeAttempt))
	assert.Equal(t, resumeInfo.ResumeErrors, unmarshaled.ResumeErrors)
}

func TestResumeInfoNilTimestamp(t *testing.T) {
	resumeInfo := ResumeInfo{
		CanResume:         false,
		ResumeCommand:     "",
		LastResumeAttempt: nil,
		ResumeErrors:      []string{},
	}

	// Test JSON serialization with nil timestamp
	data, err := json.Marshal(resumeInfo)
	require.NoError(t, err)

	var unmarshaled ResumeInfo
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, resumeInfo.CanResume, unmarshaled.CanResume)
	assert.Nil(t, unmarshaled.LastResumeAttempt)
	assert.Equal(t, resumeInfo.ResumeErrors, unmarshaled.ResumeErrors)
}

func TestSessionMetaCustomData(t *testing.T) {
	metadata := SessionMeta{
		Description: "Test metadata",
		Tags:        []string{"tag1", "tag2"},
		Variant:     "feature-branch",
		IsDefault:   false,
		CustomData: map[string]interface{}{
			"string_field": "value",
			"number_field": 123,
			"bool_field":   true,
			"nested": map[string]interface{}{
				"inner": "value",
			},
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(metadata)
	require.NoError(t, err)

	var unmarshaled SessionMeta
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, metadata.Description, unmarshaled.Description)
	assert.Equal(t, metadata.Tags, unmarshaled.Tags)
	assert.Equal(t, metadata.Variant, unmarshaled.Variant)
	assert.Equal(t, metadata.IsDefault, unmarshaled.IsDefault)

	// Verify custom data
	assert.Equal(t, "value", unmarshaled.CustomData["string_field"])
	assert.Equal(t, float64(123), unmarshaled.CustomData["number_field"]) // JSON unmarshals numbers as float64
	assert.Equal(t, true, unmarshaled.CustomData["bool_field"])

	nested, ok := unmarshaled.CustomData["nested"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value", nested["inner"])
}

func TestCleanupConfig(t *testing.T) {
	now := time.Now()

	cleanup := CleanupConfig{
		Enabled:           true,
		InactiveThreshold: "30d",
		LastCleanupCheck:  now,
	}

	// Test JSON serialization
	data, err := json.Marshal(cleanup)
	require.NoError(t, err)

	var unmarshaled CleanupConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, cleanup.Enabled, unmarshaled.Enabled)
	assert.Equal(t, cleanup.InactiveThreshold, unmarshaled.InactiveThreshold)
	assert.True(t, cleanup.LastCleanupCheck.Equal(unmarshaled.LastCleanupCheck))
}

func TestGlobalIndexSerialization(t *testing.T) {
	now := time.Now()

	globalIndex := GlobalIndex{
		Version:      "1.0.0",
		LastSync:     now,
		SyncInterval: "5m",
		Sessions: []IndexedSession{
			{
				SessionID:   "session-1",
				ProjectName: "project-1",
				ProjectPath: "/tmp/project-1",
				SessionFile: "/tmp/project-1/.claude/kamui-sessions/session-1.json",
				Variant:     "main",
				IsDefault:   true,
				Status: IndexStatus{
					IsActive:     true,
					LastAccessed: now,
					State:        SessionStateActive,
				},
				Runtime: RuntimeInfo{
					ClaudeActive:    true,
					ClaudeSessionID: "claude-123",
				},
				Git: GitInfo{
					Branch: "main",
					Commit: "abc123",
					Dirty:  false,
				},
				Metadata: IndexMeta{
					Description: "Test session",
					Tags:        []string{"test"},
					Created:     now,
				},
			},
		},
		Statistics: IndexStats{
			TotalProjects:       1,
			TotalSessions:       1,
			ActiveSessionsCount: 1,
			DiskUsage:           "1.5MB",
			LastCleanup:         now,
		},
		Configuration: IndexConfig{
			AutoIndexing:       true,
			MaxIndexAge:        "24h",
			SyncFailureRetries: 3,
			EnableStatistics:   true,
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(globalIndex)
	require.NoError(t, err)

	var unmarshaled GlobalIndex
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, globalIndex.Version, unmarshaled.Version)
	assert.True(t, globalIndex.LastSync.Equal(unmarshaled.LastSync))
	assert.Equal(t, globalIndex.SyncInterval, unmarshaled.SyncInterval)
	assert.Len(t, unmarshaled.Sessions, 1)

	// Verify indexed session
	session := unmarshaled.Sessions[0]
	assert.Equal(t, "session-1", session.SessionID)
	assert.Equal(t, "project-1", session.ProjectName)
	assert.True(t, session.Status.IsActive)
	assert.Equal(t, SessionStateActive, session.Status.State)
	assert.True(t, session.Runtime.ClaudeActive)
	assert.Equal(t, "claude-123", session.Runtime.ClaudeSessionID)

	// Verify statistics and configuration
	assert.Equal(t, 1, unmarshaled.Statistics.TotalProjects)
	assert.True(t, unmarshaled.Configuration.AutoIndexing)
}
