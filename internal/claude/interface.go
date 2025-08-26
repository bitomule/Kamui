// Package claude provides integration with Claude Code CLI
package claude

// ClientInterface defines the methods required for Claude Code integration
// This interface allows for easy mocking in unit tests
type ClientInterface interface {
	// HasSession checks if a Claude session exists by ID for the given working directory
	HasSession(sessionID, workingDir string) (bool, error)

	// StartSession creates a fresh Claude session
	StartSession(workingDir string) (string, error)

	// ResumeSession resumes an existing Claude session
	ResumeSession(sessionID, workingDir string) error

	// ListSessions returns a list of all Claude sessions
	ListSessions() ([]string, error)

	// GetSessionInfo returns information about a Claude session
	GetSessionInfo(sessionID, workingDir string) (*SessionInfo, error)

	// TerminateSession terminates a Claude session
	TerminateSession(sessionID, workingDir string) error

	// DiscoverExistingSessions finds existing Claude sessions for the current directory
	DiscoverExistingSessions(workingDir string) ([]string, error)

	// DiscoverNewestSession finds the newest Claude session (most recently created)
	DiscoverNewestSession(workingDir string) (string, error)

	// StartFreshAndDiscoverSessionID starts Claude, creates a session, and returns the ID
	StartFreshAndDiscoverSessionID(workingDir string) (string, error)
}

// Verify that Client implements ClientInterface at compile time
var _ ClientInterface = (*Client)(nil)
