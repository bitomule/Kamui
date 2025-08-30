// Package session provides the core session management functionality
package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bitomule/kamui/internal/claude"
	"github.com/bitomule/kamui/internal/storage"
	"github.com/bitomule/kamui/pkg/types"
)

// Manager handles session lifecycle and coordination
type Manager struct {
	storage      storage.Interface
	claudeClient claude.ClientInterface
	projectPath  string
}

// New creates a new session manager for the current working directory
func New() (*Manager, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, types.NewStorageError(
			types.ErrCodeProjectNotFound,
			"failed to get current working directory",
			err,
		)
	}

	return NewForPath(cwd)
}

// NewForPath creates a new session manager for a specific project path
func NewForPath(projectPath string) (*Manager, error) {
	claudeClient, err := claude.New()
	if err != nil {
		return nil, err
	}

	return NewWithClient(projectPath, claudeClient)
}

func NewWithClient(projectPath string, claudeClient claude.ClientInterface) (*Manager, error) {
	storage := storage.New(projectPath)
	return NewWithDependencies(projectPath, storage, claudeClient)
}

func NewWithDependencies(projectPath string, storageImpl storage.Interface, claudeClient claude.ClientInterface) (*Manager, error) {
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, types.NewStorageError(
			types.ErrCodeProjectNotFound,
			fmt.Sprintf("project path does not exist: %s", projectPath),
			err,
		)
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, types.NewStorageError(
			types.ErrCodeProjectInvalid,
			"failed to resolve absolute path",
			err,
		)
	}

	return &Manager{
		storage:      storageImpl,
		claudeClient: claudeClient,
		projectPath:  absPath,
	}, nil
}

// CreateOrResumeSession creates a new session or resumes an existing one
// Returns session data and whether Claude was already executed (for new sessions)
func (m *Manager) CreateOrResumeSession(sessionName string) (*types.Session, bool, error) {
	var session *types.Session
	var err error

	// Check if session already exists in storage
	if m.storage.SessionExists(sessionName) {
		// Load existing session data
		session, err = m.storage.LoadSession(sessionName)
		if err != nil {
			return nil, false, err
		}
	} else {
		// Create new session
		session, err = m.storage.CreateSession(sessionName, m.projectPath)
		if err != nil {
			return nil, false, err
		}
	}

	// Check if this session has a stored Claude session to restore
	var shouldStartFreshClaude bool
	if session.Claude.SessionID != "" {
		// Check if the stored Claude session still exists
		exists, err := m.claudeClient.HasSession(session.Claude.SessionID, session.Project.WorkingDirectory)
		if err != nil || !exists {
			shouldStartFreshClaude = true
		} else {
			shouldStartFreshClaude = false
		}
	} else {
		shouldStartFreshClaude = true
	}

	// Set up Claude session
	if shouldStartFreshClaude {
		if err := m.setupClaudeSession(session, true); err != nil {
			return nil, false, fmt.Errorf("failed to setup Claude session: %w", err)
		}
	}

	// Update access time and save
	session.LastAccessed = time.Now()
	session.LastModified = time.Now()
	if err := m.storage.SaveSession(session); err != nil {
		return nil, false, err
	}

	// Return whether Claude was already executed (true for new sessions, false for resume)
	return session, shouldStartFreshClaude, nil
}

// GetSession retrieves an existing session
func (m *Manager) GetSession(sessionName string) (*types.Session, error) {
	return m.storage.LoadSession(sessionName)
}

// ListSessions returns all sessions for the current project
func (m *Manager) ListSessions() ([]string, error) {
	return m.storage.ListSessions()
}

// CompleteSession marks a session as completed
func (m *Manager) CompleteSession(sessionName string) error {
	session, err := m.storage.LoadSession(sessionName)
	if err != nil {
		return err
	}

	// Update session state
	session.Lifecycle.State = types.SessionStateCompleted
	session.Lifecycle.StateHistory = append(session.Lifecycle.StateHistory, types.StateChange{
		State:     types.SessionStateCompleted,
		Timestamp: session.LastModified,
		Reason:    "manually_completed",
	})

	// Save updated session
	return m.storage.SaveSession(session)
}

// DeleteSession removes a session
func (m *Manager) DeleteSession(sessionName string) error {
	return m.storage.DeleteSession(sessionName)
}

// GetProjectPath returns the current project path
func (m *Manager) GetProjectPath() string {
	return m.projectPath
}

// GetProjectName returns the current project name
func (m *Manager) GetProjectName() string {
	return filepath.Base(m.projectPath)
}

// setupClaudeSession configures the Claude session using subprocess monitoring
func (m *Manager) setupClaudeSession(session *types.Session, startFresh bool) error {
	if startFresh {
		// Launch Claude with monitor subprocess - this blocks until Claude exits
		if err := m.claudeClient.LaunchClaudeInteractively(session.Project.WorkingDirectory, session.SessionID); err != nil {
			return err
		}

		// After Claude exits, the monitor subprocess should have saved the mapping
		// Try to reload the session to get the updated Claude session ID
		if updatedSession, err := m.storage.LoadSession(session.SessionID); err == nil {
			session.Claude = updatedSession.Claude
		}
	}

	return nil
}

// GetClaudeCommand returns the command to resume the Claude session
func (m *Manager) GetClaudeCommand(session *types.Session) string {
	if session.Claude.SessionID == "" {
		return "claude"
	}
	return fmt.Sprintf("claude --resume %s", session.Claude.SessionID)
}
