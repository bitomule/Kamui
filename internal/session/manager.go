// Package session provides the core session management functionality
package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/davidcollado/kamui/internal/claude"
	"github.com/davidcollado/kamui/internal/storage"
	"github.com/davidcollado/kamui/pkg/types"
)

// Manager handles session lifecycle and coordination
type Manager struct {
	storage     *storage.Storage
	claudeClient *claude.Client
	projectPath string
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
	// Verify project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, types.NewStorageError(
			types.ErrCodeProjectNotFound,
			fmt.Sprintf("project path does not exist: %s", projectPath),
			err,
		)
	}
	
	// Get absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, types.NewStorageError(
			types.ErrCodeProjectInvalid,
			"failed to resolve absolute path",
			err,
		)
	}
	
	// Initialize clients
	claudeClient, err := claude.New()
	if err != nil {
		return nil, err
	}
	
	storage := storage.New(absPath)
	
	return &Manager{
		storage:      storage,
		claudeClient: claudeClient,
		projectPath:  absPath,
	}, nil
}

// CreateOrResumeSession creates a new session or resumes an existing one
func (m *Manager) CreateOrResumeSession(sessionName string) (*types.Session, error) {
	var session *types.Session
	var err error
	
	// Check if session already exists in storage
	if m.storage.SessionExists(sessionName) {
		// Load existing session data
		session, err = m.storage.LoadSession(sessionName)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Kamui: Resuming session '%s'\n", sessionName)
	} else {
		// Create new session
		session, err = m.storage.CreateSession(sessionName, m.projectPath)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Kamui: Created new session '%s'\n", sessionName)
	}
	
	// Check if this AGX session has a stored Claude session to restore
	var shouldStartFreshClaude bool
	if session.Claude.SessionID != "" {
		// Check if the stored Claude session still exists
		exists, err := m.claudeClient.HasSession(session.Claude.SessionID, session.Project.WorkingDirectory)
		if err != nil {
			fmt.Printf("Kamui: Error checking stored Claude session: %v\n", err)
			shouldStartFreshClaude = true
		} else if !exists {
			fmt.Printf("Kamui: Stored Claude session '%s' no longer exists, starting fresh\n", session.Claude.SessionID)
			shouldStartFreshClaude = true
		} else {
			fmt.Printf("Kamui: Claude session '%s' is ready\n", session.Claude.SessionID)
			shouldStartFreshClaude = false
		}
	} else {
		fmt.Printf("Kamui: No stored Claude session for '%s', starting fresh\n", session.SessionID)
		shouldStartFreshClaude = true
	}
	
	// Set up Claude session
	if err := m.setupClaudeSession(session, shouldStartFreshClaude); err != nil {
		return nil, fmt.Errorf("failed to setup Claude session: %w", err)
	}
	
	// Update access time and save
	session.LastAccessed = time.Now()
	session.LastModified = time.Now()
	if err := m.storage.SaveSession(session); err != nil {
		return nil, err
	}
	
	return session, nil
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

// setupClaudeSession configures the Claude session
func (m *Manager) setupClaudeSession(session *types.Session, startFresh bool) error {
	if startFresh {
		// Create a fresh Claude session
		sessionID, err := m.claudeClient.StartFreshAndDiscoverSessionID(session.Project.WorkingDirectory)
		if err != nil {
			return err
		}
		
		// Store the discovered session ID
		session.Claude.SessionID = sessionID
		session.Claude.HasActiveContext = true
		session.Claude.LastInteraction = time.Now()
		session.LastModified = time.Now()
		
		fmt.Printf("Kamui: Created fresh Claude session: %s\n", sessionID)
	} else {
		// Existing Claude session is already ready
		fmt.Printf("Kamui: Using existing Claude session: %s\n", session.Claude.SessionID)
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