// Package storage handles session file operations and metadata management
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bitomule/kamui/pkg/types"
)

// Storage manages session file operations
type Storage struct {
	projectPath string
	sessionsDir string // Global sessions directory in ~/.claude/kamui-sessions/
}

// New creates a new Storage instance for the given project path
func New(projectPath string) *Storage {
	// Use global sessions directory in user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current working directory if home directory isn't accessible
		homeDir = "."
	}
	sessionsDir := filepath.Join(homeDir, ".claude", "kamui-sessions")

	return &Storage{
		projectPath: projectPath,
		sessionsDir: sessionsDir,
	}
}

// Initialize creates the necessary directories for session storage
func (s *Storage) Initialize() error {
	// Create .claude/kamui-sessions directory structure
	if err := os.MkdirAll(s.sessionsDir, 0o700); err != nil {
		return types.NewStorageError(
			types.ErrCodeStoragePermission,
			"failed to create sessions directory",
			err,
		)
	}

	return nil
}

// SaveSession saves a session to disk using friendly name as filename
func (s *Storage) SaveSession(session *types.Session) error {
	if err := s.Initialize(); err != nil {
		return err
	}

	// Use SessionID (which contains friendly name like "Undolly") as filename
	sessionFile := filepath.Join(s.sessionsDir, session.SessionID+".json")

	// Create temporary file for atomic write
	tempFile := sessionFile + ".tmp"

	// Marshal session to JSON
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return types.NewStorageError(
			types.ErrCodeStorageCorrupted,
			"failed to marshal session data",
			err,
		)
	}

	// Write to temporary file
	if err := os.WriteFile(tempFile, data, 0o600); err != nil {
		return types.NewStorageError(
			types.ErrCodeStoragePermission,
			"failed to write session file",
			err,
		)
	}

	// Atomic move to final location
	if err := os.Rename(tempFile, sessionFile); err != nil {
		os.Remove(tempFile) // cleanup temp file
		return types.NewStorageError(
			types.ErrCodeStoragePermission,
			"failed to save session file",
			err,
		)
	}

	return nil
}

// LoadSession loads a session from disk
func (s *Storage) LoadSession(sessionID string) (*types.Session, error) {
	sessionFile := filepath.Join(s.sessionsDir, sessionID+".json")

	// Check if file exists
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return nil, types.NewStorageError(
			types.ErrCodeSessionNotFound,
			fmt.Sprintf("session '%s' not found", sessionID),
			err,
		)
	}

	// Read file
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, types.NewStorageError(
			types.ErrCodeStoragePermission,
			"failed to read session file",
			err,
		)
	}

	// Unmarshal JSON
	var session types.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, types.NewStorageError(
			types.ErrCodeStorageCorrupted,
			"failed to parse session data",
			err,
		)
	}

	return &session, nil
}

// SessionExists checks if a session file exists
func (s *Storage) SessionExists(sessionID string) bool {
	sessionFile := filepath.Join(s.sessionsDir, sessionID+".json")
	_, err := os.Stat(sessionFile)
	return err == nil
}

// ListSessions returns a list of all session IDs in the project
func (s *Storage) ListSessions() ([]string, error) {
	if _, err := os.Stat(s.sessionsDir); os.IsNotExist(err) {
		return []string{}, nil // no sessions yet
	}

	entries, err := os.ReadDir(s.sessionsDir)
	if err != nil {
		return nil, types.NewStorageError(
			types.ErrCodeStoragePermission,
			"failed to read sessions directory",
			err,
		)
	}

	var sessionIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Remove .json extension to get session ID
			sessionID := entry.Name()[:len(entry.Name())-5]
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

// DeleteSession removes a session file
func (s *Storage) DeleteSession(sessionID string) error {
	sessionFile := filepath.Join(s.sessionsDir, sessionID+".json")

	if err := os.Remove(sessionFile); err != nil {
		if os.IsNotExist(err) {
			return types.NewStorageError(
				types.ErrCodeSessionNotFound,
				fmt.Sprintf("session '%s' not found", sessionID),
				err,
			)
		}
		return types.NewStorageError(
			types.ErrCodeStoragePermission,
			"failed to delete session file",
			err,
		)
	}

	return nil
}

// CreateSession creates a new session with minimal required data
func (s *Storage) CreateSession(sessionID, projectPath string) (*types.Session, error) {
	now := time.Now()

	session := &types.Session{
		Version:      "1.0.0",
		SessionID:    sessionID,
		Created:      now,
		LastAccessed: now,
		LastModified: now,

		Project: types.ProjectInfo{
			Path:             projectPath,
			WorkingDirectory: projectPath,
		},

		Claude: types.ClaudeInfo{
			SessionID: "", // Will be set when Claude session is created
		},
	}

	return session, nil
}

// UpdateSessionAccess updates the last accessed time for a session
func (s *Storage) UpdateSessionAccess(sessionID string) error {
	session, err := s.LoadSession(sessionID)
	if err != nil {
		return err
	}

	session.LastAccessed = time.Now()
	return s.SaveSession(session)
}

// GetProjectPath returns the project path for this storage instance
func (s *Storage) GetProjectPath() string {
	return s.projectPath
}

// GetSessionsPath returns the sessions directory path
func (s *Storage) GetSessionsPath() string {
	return s.sessionsDir
}
