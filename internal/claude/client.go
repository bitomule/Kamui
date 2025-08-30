// Package claude provides integration with Claude Code CLI
package claude

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitomule/kamui/pkg/types"
)

// Client manages Claude Code operations
type Client struct {
	claudePath string
}

// New creates a new Claude client
func New() (*Client, error) {
	// Find claude executable
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return nil, types.NewClaudeError(
			types.ErrCodeClaudeNotFound,
			"claude not found in PATH",
			err,
		)
	}

	return &Client{
		claudePath: claudePath,
	}, nil
}

// HasSession checks if a Claude session exists by ID for the given working directory
func (c *Client) HasSession(sessionID, workingDir string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}

	// Use the same path resolution as other methods to handle symlinks
	canonicalPath, err := filepath.EvalSymlinks(workingDir)
	if err != nil {
		// If we can't resolve symlinks, use the original path
		canonicalPath = workingDir
	}

	// Encode the path like Claude does (replace / with -)
	encodedPath := strings.ReplaceAll(canonicalPath, "/", "-")

	// Check if session file exists in ~/.claude/projects/[encoded-path]/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	sessionFile := filepath.Join(homeDir, ".claude", "projects", encodedPath, sessionID+".jsonl")
	_, err = os.Stat(sessionFile)

	return err == nil, nil
}

// StartSession creates a fresh Claude session
func (c *Client) StartSession(_ string) (string, error) {
	// For AGX, we want each session to have its own Claude session
	// Don't reuse existing Claude sessions - let each AGX session be independent
	fmt.Printf("Kamui: Will start fresh Claude session\n")

	// Return empty string to indicate no existing session to reuse
	return "", nil
}

// ResumeSession resumes an existing Claude session
func (c *Client) ResumeSession(sessionID, workingDir string) error {
	// Check if session exists
	exists, err := c.HasSession(sessionID, workingDir)
	if err != nil {
		return err
	}

	if !exists {
		return types.NewClaudeError(
			types.ErrCodeClaudeSessionNotFound,
			fmt.Sprintf("Claude session '%s' not found", sessionID),
			nil,
		)
	}

	// Provide the exact command to resume the session
	fmt.Printf("Kamui: Resume Claude session with: claude --resume %s\n", sessionID)

	return nil
}

// ListSessions returns a list of all Claude sessions
func (c *Client) ListSessions() ([]string, error) {
	cmd := exec.Command(c.claudePath, "sessions", "list")
	output, err := cmd.Output()
	if err != nil {
		// If no sessions exist, claude may return exit code 1
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return []string{}, nil // no sessions
			}
		}
		return nil, types.NewClaudeError(
			types.ErrCodeClaudeCommandFailed,
			"failed to list Claude sessions",
			err,
		)
	}

	// Parse output
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	sessions := strings.Split(outputStr, "\n")
	return sessions, nil
}

// GetSessionInfo returns information about a Claude session
func (c *Client) GetSessionInfo(sessionID, workingDir string) (*SessionInfo, error) {
	// Check if session exists
	exists, err := c.HasSession(sessionID, workingDir)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, types.NewClaudeError(
			types.ErrCodeClaudeSessionNotFound,
			fmt.Sprintf("Claude session '%s' not found", sessionID),
			nil,
		)
	}

	// Get session information (just verify it exists)
	cmd := exec.Command(c.claudePath, "sessions", "info", sessionID)
	_, err = cmd.Output()
	if err != nil {
		return nil, types.NewClaudeError(
			types.ErrCodeClaudeCommandFailed,
			fmt.Sprintf("failed to get Claude session info for '%s'", sessionID),
			err,
		)
	}

	// Return basic session info - full parsing not needed for current use cases
	info := &SessionInfo{
		SessionID: sessionID,
		Status:    "active", // Default status
		Messages:  0,        // Basic default
		LastUsed:  "",       // Basic default
	}

	return info, nil
}

// TerminateSession terminates a Claude session
func (c *Client) TerminateSession(sessionID, workingDir string) error {
	// Check if session exists
	exists, err := c.HasSession(sessionID, workingDir)
	if err != nil {
		return err
	}

	if !exists {
		return types.NewClaudeError(
			types.ErrCodeClaudeSessionNotFound,
			fmt.Sprintf("Claude session '%s' not found", sessionID),
			nil,
		)
	}

	// Terminate session
	cmd := exec.Command(c.claudePath, "sessions", "terminate", sessionID)
	if err := cmd.Run(); err != nil {
		return types.NewClaudeError(
			types.ErrCodeClaudeCommandFailed,
			fmt.Sprintf("failed to terminate Claude session '%s'", sessionID),
			err,
		)
	}

	return nil
}

// SessionInfo contains information about a Claude session
type SessionInfo struct {
	SessionID string
	Status    string
	Messages  int
	LastUsed  string
}

// Message represents a message in a Claude session JSONL file
type Message struct {
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	GitBranch string `json:"gitBranch"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
}

// DiscoverExistingSessions finds existing Claude sessions for the current directory
func (c *Client) DiscoverExistingSessions(workingDir string) ([]string, error) {
	// Resolve canonical path to handle symlinks like /tmp -> /private/tmp
	canonicalPath, err := filepath.EvalSymlinks(workingDir)
	if err != nil {
		// If we can't resolve symlinks, use the original path
		canonicalPath = workingDir
	}

	// Encode the path like Claude does (replace / with -)
	encodedPath := strings.ReplaceAll(canonicalPath, "/", "-")

	// Check if project directory exists in ~/.claude/projects/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectDir := filepath.Join(homeDir, ".claude", "projects", encodedPath)
	if _, statErr := os.Stat(projectDir); os.IsNotExist(statErr) {
		return []string{}, nil // No sessions for this project
	}

	// Read all .jsonl files in the project directory
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}

	var sessionIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".jsonl" {
			// Remove .jsonl extension to get session ID
			sessionID := entry.Name()[:len(entry.Name())-6]
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

// DiscoverNewestSession finds the newest Claude session (most recently created)
func (c *Client) DiscoverNewestSession(workingDir string) (string, error) {
	sessions, err := c.DiscoverExistingSessions(workingDir)
	if err != nil {
		return "", err
	}

	if len(sessions) == 0 {
		return "", nil
	}

	// For now, just return the first one found
	// In a more sophisticated implementation, we'd parse timestamps to find newest
	return sessions[0], nil
}

// LaunchClaudeInteractively spawns a monitor subprocess and runs Claude in main process
func (c *Client) LaunchClaudeInteractively(workingDir string, sessionName string) error {
	// Spawn monitor subprocess first
	monitorCmd, err := c.spawnMonitorProcess(sessionName, workingDir)
	if err != nil {
		return fmt.Errorf("failed to spawn monitor process: %w", err)
	}
	
	// Set up cleanup timer for monitor process (1 minute timeout)
	go func() {
		time.Sleep(1 * time.Minute)
		if monitorCmd.Process != nil {
			monitorCmd.Process.Kill()
		}
	}()
	
	// Run Claude in main process (blocking with full terminal access)
	cmd := exec.Command(c.claudePath)
	cmd.Dir = workingDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout  
	cmd.Stderr = os.Stderr
	
	// Set up Claude environment for hooks
	env := os.Environ()
	env = append(env, fmt.Sprintf("KAMUI_SESSION_ID=%s", sessionName))
	env = append(env, "KAMUI_ACTIVE=1")
	env = append(env, fmt.Sprintf("KAMUI_PROJECT_NAME=%s", filepath.Base(workingDir)))
	cmd.Env = env
	
	// This blocks until Claude exits - main process handles user interaction
	if err := cmd.Run(); err != nil {
		return types.NewClaudeError(
			types.ErrCodeClaudeStartFailed,
			"Claude session ended with error",
			err,
		)
	}
	
	return nil
}

// spawnMonitorProcess starts the monitor subprocess
func (c *Client) spawnMonitorProcess(sessionName, workingDir string) (*exec.Cmd, error) {
	// Get path to current executable
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	
	// Spawn monitor subprocess with no stdio (truly background)
	cmd := exec.Command(executable, "monitor", sessionName, workingDir)
	cmd.Dir = workingDir
	// Don't attach stdin/stdout/stderr - runs in background
	
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	
	return cmd, nil
}


// monitorForSession monitors filesystem for new Claude sessions
func (c *Client) monitorForSession(workingDir string, beforeSessions []string, timeout time.Duration) (string, error) {
	start := time.Now()
	
	for time.Since(start) < timeout {
		// Check for new sessions
		afterSessions, err := c.DiscoverExistingSessions(workingDir)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue // Keep trying
		}
		
		// Find any new session
		for _, sessionID := range afterSessions {
			found := false
			for _, oldSession := range beforeSessions {
				if sessionID == oldSession {
					found = true
					break
				}
			}
			if !found {
				// Found new session
				return sessionID, nil
			}
		}
		
		// Wait before checking again
		time.Sleep(1 * time.Second)
	}
	
	// Timeout reached
	return "", types.NewClaudeError(
		types.ErrCodeClaudeStartFailed,
		"timeout monitoring for Claude session creation",
		nil,
	)
}


// getSessionFilePath returns the path to a Claude session file
func (c *Client) getSessionFilePath(workingDir, sessionID string) (string, error) {
	// Resolve canonical path to handle symlinks like /tmp -> /private/tmp
	canonicalPath, err := filepath.EvalSymlinks(workingDir)
	if err != nil {
		// If we can't resolve symlinks, use the original path
		canonicalPath = workingDir
	}

	encodedPath := strings.ReplaceAll(canonicalPath, "/", "-")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".claude", "projects", encodedPath, sessionID+".jsonl"), nil
}
