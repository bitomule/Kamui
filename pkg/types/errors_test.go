package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAGXError_Error(t *testing.T) {
	// Test without cause
	err := &AGXError{
		Code:    ErrCodeSessionNotFound,
		Message: "session does not exist",
	}

	expected := "[SESSION_NOT_FOUND] session does not exist"
	assert.Equal(t, expected, err.Error())

	// Test with cause
	cause := errors.New("file not found")
	errWithCause := &AGXError{
		Code:    ErrCodeStoragePermission,
		Message: "failed to read session",
		Cause:   cause,
	}

	expected = "[STORAGE_PERMISSION] failed to read session: file not found"
	assert.Equal(t, expected, errWithCause.Error())
}

func TestAGXError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &AGXError{
		Code:    ErrCodeUnknown,
		Message: "wrapper error",
		Cause:   cause,
	}

	assert.Equal(t, cause, err.Unwrap())

	// Test without cause
	errNoCause := &AGXError{
		Code:    ErrCodeUnknown,
		Message: "error without cause",
	}

	assert.Nil(t, errNoCause.Unwrap())
}

func TestNewDependencyError(t *testing.T) {
	cause := errors.New("binary not found")
	err := NewDependencyError("claude CLI not installed", cause)

	assert.Equal(t, ErrCodeDependencyMissing, err.Code)
	assert.Equal(t, "claude CLI not installed", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestNewSessionError(t *testing.T) {
	cause := errors.New("file corrupted")
	err := NewSessionError(ErrCodeSessionCorrupted, "session data invalid", cause)

	assert.Equal(t, ErrCodeSessionCorrupted, err.Code)
	assert.Equal(t, "session data invalid", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestNewStorageError(t *testing.T) {
	cause := errors.New("permission denied")
	err := NewStorageError(ErrCodeStoragePermission, "cannot write to directory", cause)

	assert.Equal(t, ErrCodeStoragePermission, err.Code)
	assert.Equal(t, "cannot write to directory", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestNewClaudeError(t *testing.T) {
	cause := errors.New("connection failed")
	err := NewClaudeError(ErrCodeClaudeTimeout, "Claude CLI timed out", cause)

	assert.Equal(t, ErrCodeClaudeTimeout, err.Code)
	assert.Equal(t, "Claude CLI timed out", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestAGXError_WithContext(t *testing.T) {
	err := &AGXError{
		Code:    ErrCodeSessionNotFound,
		Message: "session missing",
	}

	updatedErr := err.WithContext("sessionID", "test-session-123")

	// Should return the same error instance
	assert.Equal(t, err, updatedErr)

	// Should have added context
	assert.NotNil(t, err.Context)
	assert.Equal(t, "test-session-123", err.Context["sessionID"])

	// Add another context value
	err = err.WithContext("projectPath", "/tmp/test")
	assert.Equal(t, "/tmp/test", err.Context["projectPath"])
	assert.Equal(t, "test-session-123", err.Context["sessionID"])
}

func TestAGXError_IsRecoverable(t *testing.T) {
	recoverableCodes := []ErrorCode{
		ErrCodeSessionLocked,
		ErrCodeStorageLocked,
		ErrCodeClaudeResumeFailed,
		ErrCodeTimeout,
	}

	for _, code := range recoverableCodes {
		err := &AGXError{Code: code, Message: "test"}
		assert.True(t, err.IsRecoverable(), "Expected %s to be recoverable", code)
	}

	nonRecoverableCodes := []ErrorCode{
		ErrCodeSessionNotFound,
		ErrCodeDependencyMissing,
		ErrCodeStorageCorrupted,
		ErrCodeInvalidInput,
	}

	for _, code := range nonRecoverableCodes {
		err := &AGXError{Code: code, Message: "test"}
		assert.False(t, err.IsRecoverable(), "Expected %s to not be recoverable", code)
	}
}

func TestAGXError_IsUserError(t *testing.T) {
	userErrorCodes := []ErrorCode{
		ErrCodeInvalidInput,
		ErrCodeConfigInvalid,
		ErrCodeProjectNotFound,
	}

	for _, code := range userErrorCodes {
		err := &AGXError{Code: code, Message: "test"}
		assert.True(t, err.IsUserError(), "Expected %s to be a user error", code)
	}

	systemErrorCodes := []ErrorCode{
		ErrCodeStoragePermission,
		ErrCodeClaudeNotFound,
		ErrCodeTimeout,
		ErrCodeUnknown,
	}

	for _, code := range systemErrorCodes {
		err := &AGXError{Code: code, Message: "test"}
		assert.False(t, err.IsUserError(), "Expected %s to not be a user error", code)
	}
}

func TestAGXError_GetRecoveryHint(t *testing.T) {
	testCases := []struct {
		code         ErrorCode
		expectedHint string
	}{
		{ErrCodeDependencyMissing, "Install required dependencies (claude)"},
		{ErrCodeSessionLocked, "Wait for lock to be released or remove stale lock file"},
		{ErrCodeStoragePermission, "Check file permissions for AGX directories"},
		{ErrCodeClaudeNotFound, "Install Claude Code CLI"},
		{ErrCodeSessionCorrupted, "Session data may be corrupted, consider creating a new session"},
		{ErrCodeConfigInvalid, "Check configuration file syntax and values"},
		{ErrCodeUnknown, "Check the error message for specific details"},
	}

	for _, tc := range testCases {
		err := &AGXError{Code: tc.code, Message: "test"}
		assert.Equal(t, tc.expectedHint, err.GetRecoveryHint(), "Recovery hint mismatch for %s", tc.code)
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined as expected
	allCodes := []ErrorCode{
		// Dependency errors
		ErrCodeDependencyMissing,
		ErrCodeDependencyVersion,
		ErrCodeDependencyFailed,

		// Session errors
		ErrCodeSessionNotFound,
		ErrCodeSessionExists,
		ErrCodeSessionCorrupted,
		ErrCodeSessionLocked,
		ErrCodeSessionInvalid,

		// Storage errors
		ErrCodeStoragePermission,
		ErrCodeStorageNotFound,
		ErrCodeStorageCorrupted,
		ErrCodeStorageFull,
		ErrCodeStorageLocked,

		// Claude integration errors
		ErrCodeClaudeNotFound,
		ErrCodeClaudeSessionInvalid,
		ErrCodeClaudeSessionNotFound,
		ErrCodeClaudeResumeFailed,
		ErrCodeClaudeStartFailed,
		ErrCodeClaudeCommandFailed,
		ErrCodeClaudeTimeout,

		// Configuration errors
		ErrCodeConfigInvalid,
		ErrCodeConfigNotFound,
		ErrCodeConfigPermission,

		// Project errors
		ErrCodeProjectNotFound,
		ErrCodeProjectInvalid,
		ErrCodeProjectPermission,

		// General errors
		ErrCodeInvalidInput,
		ErrCodeTimeout,
		ErrCodeInterrupted,
		ErrCodeUnknown,
	}

	// Ensure all codes have non-empty string values
	for _, code := range allCodes {
		assert.NotEmpty(t, string(code), "Error code should not be empty")
	}

	// Test specific values for key error codes
	assert.Equal(t, "SESSION_NOT_FOUND", string(ErrCodeSessionNotFound))
	assert.Equal(t, "CLAUDE_NOT_FOUND", string(ErrCodeClaudeNotFound))
	assert.Equal(t, "STORAGE_PERMISSION", string(ErrCodeStoragePermission))
}
