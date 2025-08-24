// Package types defines error types for AGX
package types

import (
	"fmt"
)

// AGXError represents a base error type for AGX operations
type AGXError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"cause,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

func (e *AGXError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AGXError) Unwrap() error {
	return e.Cause
}

// ErrorCode represents different types of errors in AGX
type ErrorCode string

const (
	// Dependency errors
	ErrCodeDependencyMissing   ErrorCode = "DEPENDENCY_MISSING"
	ErrCodeDependencyVersion   ErrorCode = "DEPENDENCY_VERSION"
	ErrCodeDependencyFailed    ErrorCode = "DEPENDENCY_FAILED"
	
	// Session errors  
	ErrCodeSessionNotFound     ErrorCode = "SESSION_NOT_FOUND"
	ErrCodeSessionExists       ErrorCode = "SESSION_EXISTS"
	ErrCodeSessionCorrupted    ErrorCode = "SESSION_CORRUPTED"
	ErrCodeSessionLocked       ErrorCode = "SESSION_LOCKED"
	ErrCodeSessionInvalid      ErrorCode = "SESSION_INVALID"
	
	// Storage errors
	ErrCodeStoragePermission   ErrorCode = "STORAGE_PERMISSION"
	ErrCodeStorageNotFound     ErrorCode = "STORAGE_NOT_FOUND"
	ErrCodeStorageCorrupted    ErrorCode = "STORAGE_CORRUPTED"
	ErrCodeStorageFull         ErrorCode = "STORAGE_FULL"
	ErrCodeStorageLocked       ErrorCode = "STORAGE_LOCKED"
	
	// Claude integration errors
	ErrCodeClaudeNotFound        ErrorCode = "CLAUDE_NOT_FOUND"
	ErrCodeClaudeSessionInvalid  ErrorCode = "CLAUDE_SESSION_INVALID"
	ErrCodeClaudeSessionNotFound ErrorCode = "CLAUDE_SESSION_NOT_FOUND"
	ErrCodeClaudeResumeFailed    ErrorCode = "CLAUDE_RESUME_FAILED"
	ErrCodeClaudeStartFailed     ErrorCode = "CLAUDE_START_FAILED"
	ErrCodeClaudeCommandFailed   ErrorCode = "CLAUDE_COMMAND_FAILED"
	ErrCodeClaudeTimeout         ErrorCode = "CLAUDE_TIMEOUT"
	
	// Configuration errors
	ErrCodeConfigInvalid       ErrorCode = "CONFIG_INVALID"
	ErrCodeConfigNotFound      ErrorCode = "CONFIG_NOT_FOUND"
	ErrCodeConfigPermission    ErrorCode = "CONFIG_PERMISSION"
	
	// Project errors
	ErrCodeProjectNotFound     ErrorCode = "PROJECT_NOT_FOUND"
	ErrCodeProjectInvalid      ErrorCode = "PROJECT_INVALID"
	ErrCodeProjectPermission   ErrorCode = "PROJECT_PERMISSION"
	
	// General errors
	ErrCodeInvalidInput        ErrorCode = "INVALID_INPUT"
	ErrCodeTimeout             ErrorCode = "TIMEOUT"
	ErrCodeInterrupted         ErrorCode = "INTERRUPTED"
	ErrCodeUnknown             ErrorCode = "UNKNOWN"
)

// Error constructor functions

// NewDependencyError creates a new dependency-related error
func NewDependencyError(message string, cause error) *AGXError {
	return &AGXError{
		Code:    ErrCodeDependencyMissing,
		Message: message,
		Cause:   cause,
	}
}

// NewSessionError creates a new session-related error  
func NewSessionError(code ErrorCode, message string, cause error) *AGXError {
	return &AGXError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewStorageError creates a new storage-related error
func NewStorageError(code ErrorCode, message string, cause error) *AGXError {
	return &AGXError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewClaudeError creates a new Claude-related error
func NewClaudeError(code ErrorCode, message string, cause error) *AGXError {
	return &AGXError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WithContext adds context information to an error
func (e *AGXError) WithContext(key string, value interface{}) *AGXError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// IsRecoverable returns true if the error represents a recoverable condition
func (e *AGXError) IsRecoverable() bool {
	switch e.Code {
	case ErrCodeSessionLocked, ErrCodeStorageLocked:
		return true // Can retry after lock is released
	case ErrCodeClaudeResumeFailed:  
		return true // Can attempt alternative approaches
	case ErrCodeTimeout:
		return true // Can retry operation
	default:
		return false
	}
}

// IsUserError returns true if the error is due to user input/configuration
func (e *AGXError) IsUserError() bool {
	switch e.Code {
	case ErrCodeInvalidInput, ErrCodeConfigInvalid, ErrCodeProjectNotFound:
		return true
	default:
		return false
	}
}

// GetRecoveryHint returns a hint for how to recover from the error
func (e *AGXError) GetRecoveryHint() string {
	switch e.Code {
	case ErrCodeDependencyMissing:
		return "Install required dependencies (claude)"
	case ErrCodeSessionLocked:
		return "Wait for lock to be released or remove stale lock file"
	case ErrCodeStoragePermission:
		return "Check file permissions for AGX directories"
	case ErrCodeClaudeNotFound:
		return "Install Claude Code CLI"
	case ErrCodeSessionCorrupted:
		return "Session data may be corrupted, consider creating a new session"
	case ErrCodeConfigInvalid:
		return "Check configuration file syntax and values"
	default:
		return "Check the error message for specific details"
	}
}