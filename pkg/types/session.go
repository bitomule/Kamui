// Package types defines the core data structures for AGX session management
package types

import (
	"time"
)

// Session represents a complete AGX session with Claude Code integration
type Session struct {
	Version      string    `json:"version"`
	SessionID    string    `json:"sessionId"`
	Created      time.Time `json:"created"`
	LastAccessed time.Time `json:"lastAccessed"`
	LastModified time.Time `json:"lastModified"`

	Project   ProjectInfo   `json:"project"`
	Claude    ClaudeInfo    `json:"claude"`
	Metadata  SessionMeta   `json:"metadata"`
	Stats     SessionStats  `json:"statistics"`
	Lifecycle LifecycleInfo `json:"lifecycle"`
}

// ProjectInfo contains information about the project this session belongs to
type ProjectInfo struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	WorkingDirectory string `json:"workingDirectory"`
	GitBranch        string `json:"gitBranch"`
	GitCommit        string `json:"gitCommit"`
	GitRemote        string `json:"gitRemote"`
}

// ClaudeInfo contains Claude Code session information
type ClaudeInfo struct {
	SessionID        string      `json:"sessionId"`
	ConversationID   string      `json:"conversationId"`
	ModelUsed        string      `json:"modelUsed"`
	HasActiveContext bool        `json:"hasActiveContext"`
	LastInteraction  time.Time   `json:"lastInteraction"`
	ContextInfo      ContextInfo `json:"contextInfo"`
	ResumeInfo       ResumeInfo  `json:"resumeInfo"`
}

// ContextInfo contains metadata about the Claude conversation state
type ContextInfo struct {
	MessageCount    int      `json:"messageCount"`
	EstimatedTokens int      `json:"estimatedTokens"`
	LastCommand     string   `json:"lastCommand"`
	WorkingFiles    []string `json:"workingFiles"`
}

// ResumeInfo contains information needed for Claude session resumption
type ResumeInfo struct {
	CanResume         bool       `json:"canResume"`
	ResumeCommand     string     `json:"resumeCommand"`
	LastResumeAttempt *time.Time `json:"lastResumeAttempt"`
	ResumeErrors      []string   `json:"resumeErrors"`
}

// SessionMeta contains session metadata and user-defined information
type SessionMeta struct {
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Variant     string                 `json:"variant"`
	IsDefault   bool                   `json:"isDefault"`
	CustomData  map[string]interface{} `json:"customData"`
}

// SessionStats contains usage statistics for the session
type SessionStats struct {
	SessionCount         int    `json:"sessionCount"`
	TotalDuration        string `json:"totalDuration"`
	AverageSessionLength string `json:"averageSessionLength"`
	LastSessionDuration  string `json:"lastSessionDuration"`
	MostActiveDay        string `json:"mostActiveDay"`
	CommandsExecuted     int    `json:"commandsExecuted"`
}

// LifecycleInfo tracks the session lifecycle and state management
type LifecycleInfo struct {
	State        SessionState  `json:"state"`
	StateHistory []StateChange `json:"stateHistory"`
	AutoCleanup  CleanupConfig `json:"autoCleanup"`
}

// SessionState represents the current state of a session
type SessionState string

const (
	SessionStateActive    SessionState = "active"
	SessionStatePaused    SessionState = "paused"
	SessionStateCompleted SessionState = "completed"
	SessionStateArchived  SessionState = "archived"
	SessionStateError     SessionState = "error"
)

// StateChange represents a session state transition
type StateChange struct {
	State     SessionState `json:"state"`
	Timestamp time.Time    `json:"timestamp"`
	Reason    string       `json:"reason"`
}

// CleanupConfig controls automatic session cleanup behavior
type CleanupConfig struct {
	Enabled           bool      `json:"enabled"`
	InactiveThreshold string    `json:"inactiveThreshold"`
	LastCleanupCheck  time.Time `json:"lastCleanupCheck"`
}

// GlobalIndex represents the global session discovery index
type GlobalIndex struct {
	Version       string           `json:"version"`
	LastSync      time.Time        `json:"lastSync"`
	SyncInterval  string           `json:"syncInterval"`
	Sessions      []IndexedSession `json:"sessions"`
	Statistics    IndexStats       `json:"statistics"`
	Configuration IndexConfig      `json:"configuration"`
}

// IndexedSession represents a session entry in the global index
type IndexedSession struct {
	SessionID   string      `json:"sessionId"`
	ProjectName string      `json:"projectName"`
	ProjectPath string      `json:"projectPath"`
	SessionFile string      `json:"sessionFile"`
	Variant     string      `json:"variant"`
	IsDefault   bool        `json:"isDefault"`
	Status      IndexStatus `json:"status"`
	Runtime     RuntimeInfo `json:"runtime"`
	Git         GitInfo     `json:"git"`
	Metadata    IndexMeta   `json:"metadata"`
}

// IndexStatus contains session status information for the index
type IndexStatus struct {
	IsActive     bool         `json:"isActive"`
	LastAccessed time.Time    `json:"lastAccessed"`
	State        SessionState `json:"state"`
}

// RuntimeInfo contains current runtime status
type RuntimeInfo struct {
	ClaudeActive    bool   `json:"claudeActive"`
	ClaudeSessionID string `json:"claudeSessionId"`
}

// GitInfo contains git context information for the index
type GitInfo struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	Dirty  bool   `json:"dirty"`
}

// IndexMeta contains condensed metadata for the index
type IndexMeta struct {
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Created     time.Time `json:"created"`
}

// IndexStats contains statistics about all tracked sessions
type IndexStats struct {
	TotalProjects       int       `json:"totalProjects"`
	TotalSessions       int       `json:"totalSessions"`
	ActiveSessionsCount int       `json:"activeSessionsCount"`
	DiskUsage           string    `json:"diskUsage"`
	LastCleanup         time.Time `json:"lastCleanup"`
}

// IndexConfig contains configuration for index management
type IndexConfig struct {
	AutoIndexing       bool   `json:"autoIndexing"`
	MaxIndexAge        string `json:"maxIndexAge"`
	SyncFailureRetries int    `json:"syncFailureRetries"`
	EnableStatistics   bool   `json:"enableStatistics"`
}

// Config represents the global AGX configuration
type Config struct {
	Version string        `json:"version"`
	Default DefaultConfig `json:"default"`
	Claude  ClaudeConfig  `json:"claude"`
	Session SessionConfig `json:"session"`
	Storage StorageConfig `json:"storage"`
	UI      UIConfig      `json:"ui"`
}

// DefaultConfig contains default behavior settings
type DefaultConfig struct {
	SessionVariant     string `json:"sessionVariant"`
	AutoCreateSessions bool   `json:"autoCreateSessions"`
	ProjectDetection   string `json:"projectDetection"`
}

// ClaudeConfig contains Claude Code integration settings
type ClaudeConfig struct {
	DefaultModel        string   `json:"defaultModel"`
	ResumeTimeout       string   `json:"resumeTimeout"`
	DefaultArgs         []string `json:"defaultArgs"`
	RetryAttempts       int      `json:"retryAttempts"`
	ContextPreservation bool     `json:"contextPreservation"`
}

// SessionConfig contains session management settings
type SessionConfig struct {
	AutoBranchSessions  bool `json:"autoBranchSessions"`
	CleanupInactiveDays int  `json:"cleanupInactiveDays"`
	BackupCount         int  `json:"backupCount"`
	AutoArchive         bool `json:"autoArchive"`
	EnableStatistics    bool `json:"enableStatistics"`
}

// StorageConfig contains storage and indexing settings
type StorageConfig struct {
	IndexSyncInterval string `json:"indexSyncInterval"`
	EnableGlobalIndex bool   `json:"enableGlobalIndex"`
	CompactThreshold  string `json:"compactThreshold"`
	LogRetentionDays  int    `json:"logRetentionDays"`
}

// UIConfig contains user interface settings
type UIConfig struct {
	ColorOutput        bool   `json:"colorOutput"`
	VerboseLogging     bool   `json:"verboseLogging"`
	ConfirmDestructive bool   `json:"confirmDestructive"`
	DefaultEditor      string `json:"defaultEditor"`
}

// ProjectConfig represents project-specific configuration
type ProjectConfig struct {
	Version string               `json:"version"`
	Project ProjectConfigInfo    `json:"project"`
	Claude  ClaudeProjectConfig  `json:"claude"`
	Session SessionProjectConfig `json:"session"`
}

// ProjectConfigInfo contains project identification
type ProjectConfigInfo struct {
	Name                  string `json:"name"`
	DefaultSessionVariant string `json:"defaultSessionVariant"`
	WorkingDirectory      string `json:"workingDirectory"`
}

// ClaudeProjectConfig contains project-specific Claude settings
type ClaudeProjectConfig struct {
	Model        string   `json:"model"`
	ContextFiles []string `json:"contextFiles"`
}

// SessionProjectConfig contains project-specific session settings
type SessionProjectConfig struct {
	Variants       []string `json:"variants"`
	BranchSessions bool     `json:"branchSessions"`
	AutoCleanup    bool     `json:"autoCleanup"`
}
