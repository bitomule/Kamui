// Package claude provides integration with Claude Code CLI
package claude

import (
	"bufio"
	"encoding/json"
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

	// Encode the path like Claude does (replace / with -)
	encodedPath := strings.ReplaceAll(workingDir, "/", "-")

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
	// Encode the path like Claude does (replace / with -)
	encodedPath := strings.ReplaceAll(workingDir, "/", "-")

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

// StartFreshAndDiscoverSessionID starts Claude, sends a message to create session, discovers ID, then kills it
// This ensures each AGX session gets a truly independent Claude session
func (c *Client) StartFreshAndDiscoverSessionID(workingDir string) (string, error) {
	fmt.Printf("Kamui: Starting fresh Claude session to ensure independence...\n")

	// Get baseline sessions before starting
	beforeSessions, err := c.DiscoverExistingSessions(workingDir)
	if err != nil {
		return "", err
	}

	fmt.Printf("Kamui: Found %d existing Claude sessions before starting\n", len(beforeSessions))

	// Use Claude with --print to send a dummy message that creates a session
	// This forces Claude to create a new session file
	cmd := exec.Command(c.claudePath, "--print", "KAMUI_INIT_MESSAGE")
	cmd.Dir = workingDir

	fmt.Printf("Kamui: Creating fresh Claude session...\n")
	err = cmd.Run()
	if err != nil {
		return "", types.NewClaudeError(
			types.ErrCodeClaudeStartFailed,
			"failed to start Claude for fresh session creation",
			err,
		)
	}

	// Wait a moment for session file to be written
	time.Sleep(1 * time.Second)

	// Get sessions after the message to find the new one
	afterSessions, err := c.DiscoverExistingSessions(workingDir)
	if err != nil {
		return "", err
	}

	fmt.Printf("Kamui: Found %d Claude sessions after creation\n", len(afterSessions))

	// Find the new session by comparing before and after
	var newSessionID string
	for _, session := range afterSessions {
		found := false
		for _, oldSession := range beforeSessions {
			if session == oldSession {
				found = true
				break
			}
		}
		if !found {
			newSessionID = session
			break
		}
	}

	if newSessionID == "" {
		return "", types.NewClaudeError(
			types.ErrCodeClaudeStartFailed,
			"failed to discover new Claude session ID after creation",
			nil,
		)
	}

	// Clean up the dummy message from the session file
	if err := c.cleanupInitMessage(workingDir, newSessionID); err != nil {
		fmt.Printf("Kamui: Warning - could not clean up init message: %v\n", err)
		// Don't fail the whole operation if cleanup fails
	}

	fmt.Printf("Kamui: Created clean Claude session ID: %s\n", newSessionID)
	return newSessionID, nil
}

// cleanupInitMessage removes the exact dummy KAMUI_INIT_MESSAGE from the Claude session JSONL file
// Only removes user messages that exactly match our init message, not assistant responses
func (c *Client) cleanupInitMessage(workingDir, sessionID string) error {
	sessionFile, err := c.getSessionFilePath(workingDir, sessionID)
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(sessionFile); os.IsNotExist(statErr) {
		return fmt.Errorf("session file not found: %s", sessionFile)
	}

	cleanLines, err := c.filterInitMessages(sessionFile)
	if err != nil {
		return err
	}

	return c.writeCleanedSession(sessionFile, cleanLines)
}

func (c *Client) getSessionFilePath(workingDir, sessionID string) (string, error) {
	encodedPath := strings.ReplaceAll(workingDir, "/", "-")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".claude", "projects", encodedPath, sessionID+".jsonl"), nil
}

func (c *Client) filterInitMessages(sessionFile string) ([]string, error) {
	file, err := os.Open(sessionFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cleanLines []string
	scanner := bufio.NewScanner(file)
	foundInitMessage := false

	for scanner.Scan() {
		line := scanner.Text()
		if c.shouldKeepLine(line, &foundInitMessage) {
			cleanLines = append(cleanLines, line)
		}
	}

	return cleanLines, scanner.Err()
}

func (c *Client) shouldKeepLine(line string, foundInitMessage *bool) bool {
	var message map[string]interface{}
	if err := json.Unmarshal([]byte(line), &message); err != nil {
		return true // Keep unparseable lines
	}

	messageType, hasType := message["type"].(string)
	if !hasType || messageType != "user" {
		return true // Keep non-user messages
	}

	if *foundInitMessage {
		return true // Already found and removed init message
	}

	if c.isInitMessage(message) {
		*foundInitMessage = true
		fmt.Printf("Kamui: Removing exact init message from Claude session\n")
		return false // Skip this line
	}

	return true
}

func (c *Client) isInitMessage(message map[string]interface{}) bool {
	// Check nested message structure
	if messageData, ok := message["message"].(map[string]interface{}); ok {
		if content, ok := messageData["content"].(string); ok && content == "KAMUI_INIT_MESSAGE" {
			return true
		}
	}

	// Check direct content field
	if content, ok := message["content"].(string); ok && content == "KAMUI_INIT_MESSAGE" {
		return true
	}

	return false
}

func (c *Client) writeCleanedSession(sessionFile string, cleanLines []string) error {
	tempFile := sessionFile + ".tmp"
	outFile, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, line := range cleanLines {
		if _, err := outFile.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	_ = outFile.Close()
	if err := os.Rename(tempFile, sessionFile); err != nil {
		_ = os.Remove(tempFile) // cleanup temp file on failure
		return err
	}

	return nil
}
