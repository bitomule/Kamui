# AGX Storage Format Specification

## Overview

AGX uses a JSON-based storage system with two layers:
1. **Local Session Storage**: Project-specific session metadata stored in `.agx/sessions/`
2. **Global Session Index**: Cross-project discovery index stored in `~/.agx/index.json`

## Directory Structure

```
project-root/
  .agx/
    sessions/
      main.json                    # Default session metadata
      feature-auth.json            # Branch-specific session
      testing.json                 # Named variant session
      .current                     # Current active session name
      .lock                        # File lock for atomic operations
    config.json                    # Project-specific configuration

~/.agx/
  config.json                      # Global configuration
  index.json                       # Global session discovery index
  logs/                           # Operation logs
    agx-2025-01-24.log
  temp/                          # Temporary files for atomic operations
  cache/                         # Performance optimization cache
```

## Session Metadata Format

### Local Session File (`project/.agx/sessions/{session-id}.json`)

```json
{
  "version": "1.0.0",
  "sessionId": "main",
  "created": "2025-01-24T10:30:00Z",
  "lastAccessed": "2025-01-24T14:45:00Z",
  "lastModified": "2025-01-24T14:40:00Z",
  
  "project": {
    "name": "myproject",
    "path": "/Users/user/projects/myproject",
    "workingDirectory": "/Users/user/projects/myproject/src",
    "gitBranch": "main",
    "gitCommit": "abc123def456",
    "gitRemote": "origin/main"
  },
  
  "tmux": {
    "sessionName": "agx-myproject-main",
    "windowCount": 2,
    "isActive": true,
    "lastAttached": "2025-01-24T14:45:00Z",
    "config": {
      "defaultWindow": {
        "name": "main",
        "path": "/Users/user/projects/myproject",
        "command": null
      },
      "additionalWindows": [
        {
          "name": "test",
          "path": "/Users/user/projects/myproject/test",
          "command": "npm test"
        }
      ]
    }
  },
  
  "claude": {
    "sessionId": "abc123-def456-ghi789",
    "conversationId": "conv_xyz789",
    "modelUsed": "claude-3-sonnet",
    "hasActiveContext": true,
    "lastInteraction": "2025-01-24T14:40:00Z",
    "contextInfo": {
      "messageCount": 45,
      "estimatedTokens": 15000,
      "lastCommand": "/help tmux",
      "workingFiles": [
        "src/main.go",
        "internal/session/manager.go"
      ]
    },
    "resumeInfo": {
      "canResume": true,
      "resumeCommand": "claude --resume abc123-def456-ghi789",
      "lastResumeAttempt": null,
      "resumeErrors": []
    }
  },
  
  "metadata": {
    "description": "Main development session for user authentication",
    "tags": ["development", "auth", "backend"],
    "variant": "main",
    "isDefault": true,
    "customData": {
      "environment": "development",
      "debugMode": false,
      "notes": "Working on JWT implementation"
    }
  },
  
  "statistics": {
    "sessionCount": 12,
    "totalDuration": "4h 23m",
    "averageSessionLength": "22m",
    "lastSessionDuration": "45m",
    "mostActiveDay": "2025-01-20",
    "commandsExecuted": 156
  },
  
  "lifecycle": {
    "state": "active",
    "stateHistory": [
      {
        "state": "created",
        "timestamp": "2025-01-24T10:30:00Z",
        "reason": "initial_creation"
      },
      {
        "state": "active",
        "timestamp": "2025-01-24T10:35:00Z",
        "reason": "first_attachment"
      }
    ],
    "autoCleanup": {
      "enabled": true,
      "inactiveThreshold": "30d",
      "lastCleanupCheck": "2025-01-24T00:00:00Z"
    }
  }
}
```

### Field Definitions

#### Core Fields
- `version`: Storage format version for migration support
- `sessionId`: Unique identifier within project (e.g., "main", "feature-auth")  
- `created`: ISO timestamp of session creation
- `lastAccessed`: ISO timestamp of last access (read or write)
- `lastModified`: ISO timestamp of last modification

#### Project Information
- `project.name`: Human-readable project name
- `project.path`: Absolute path to project root
- `project.workingDirectory`: Current working directory for new sessions
- `project.git*`: Git context information for branch awareness

#### Tmux Integration
- `tmux.sessionName`: Actual tmux session name (includes AGX prefix)
- `tmux.isActive`: Whether tmux session is currently running
- `tmux.config`: Window configuration for session creation

#### Claude Code Integration  
- `claude.sessionId`: Claude Code session identifier for `--resume`
- `claude.hasActiveContext`: Whether Claude session has conversation history
- `claude.contextInfo`: Metadata about conversation state
- `claude.resumeInfo`: Information needed for session resumption

#### Session Management
- `metadata.variant`: Session variant (branch name, custom name, or "main")
- `metadata.isDefault`: Whether this is the default session for the project
- `lifecycle.state`: Current session state (active, paused, completed, archived)

## Global Index Format

### Global Index File (`~/.agx/index.json`)

```json
{
  "version": "1.0.0",
  "lastSync": "2025-01-24T14:45:00Z",
  "syncInterval": "5m",
  
  "sessions": [
    {
      "sessionId": "main",
      "projectName": "myproject",
      "projectPath": "/Users/user/projects/myproject",
      "sessionFile": "/Users/user/projects/myproject/.agx/sessions/main.json",
      "variant": "main",
      "isDefault": true,
      
      "status": {
        "isActive": true,
        "lastAccessed": "2025-01-24T14:45:00Z",
        "state": "active"
      },
      
      "runtime": {
        "tmuxActive": true,
        "claudeActive": true,
        "claudeSessionId": "abc123-def456-ghi789"
      },
      
      "git": {
        "branch": "main",
        "commit": "abc123def456",
        "dirty": false
      },
      
      "metadata": {
        "description": "Main development session",
        "tags": ["development", "auth"],
        "created": "2025-01-24T10:30:00Z"
      }
    }
  ],
  
  "statistics": {
    "totalProjects": 5,
    "totalSessions": 8,
    "activeSessionsCount": 2,
    "diskUsage": "2.4MB",
    "lastCleanup": "2025-01-23T02:00:00Z"
  },
  
  "configuration": {
    "autoIndexing": true,
    "maxIndexAge": "7d",
    "syncFailureRetries": 3,
    "enableStatistics": true
  }
}
```

## Configuration Formats

### Global Configuration (`~/.agx/config.json`)

```json
{
  "version": "1.0.0",
  "
ult": {
    "sessionVariant": "main",
    "autoCreateSessions": true,
    "projectDetection": "auto"
  },
  
  "tmux": {
    "defaultWindowCount": 2,
    "windowNames": ["main", "test"],
    "startupCommand": null,
    "sessionPrefix": "agx",
    "attachTimeout": "10s"
  },
  
  "claude": {
    "defaultModel": "claude-3-sonnet",
    "resumeTimeout": "30s",
    "defaultArgs": [],
    "retryAttempts": 3,
    "contextPreservation": true
  },
  
  "session": {
    "autoBranchSessions": true,
    "cleanupInactiveDays": 30,
    "backupCount": 5,
    "autoArchive": true,
    "enableStatistics": true
  },
  
  "storage": {
    "indexSyncInterval": "5m",
    "enableGlobalIndex": true,
    "compactThreshold": "100MB",
    "logRetentionDays": 7
  },
  
  "ui": {
    "colorOutput": true,
    "verboseLogging": false,
    "confirmDestructive": true,
    "defaultEditor": "nano"
  }
}
```

### Project Configuration (`project/.agx/config.json`)

```json
{
  "version": "1.0.0",
  "project": {
    "name": "myproject",
    "defaultSessionVariant": "main",
    "workingDirectory": "src/"
  },
  
  "tmux": {
    "windowCount": 3,
    "windows": [
      {
        "name": "main",
        "path": ".",
        "command": null
      },
      {
        "name": "test",
        "path": "test/",
        "command": "npm run test:watch"
      },
      {
        "name": "server",
        "path": ".",
        "command": "npm run dev"
      }
    ]
  },
  
  "claude": {
    "model": "claude-3-sonnet",
    "contextFiles": [
      "README.md",
      "package.json",
      "src/main.go"
    ]
  },
  
  "session": {
    "variants": ["main", "testing", "debug"],
    "branchSessions": true,
    "autoCleanup": false
  }
}
```

## File Operations

### Atomic Operations
All session file modifications use atomic operations to prevent corruption:

1. **Write**: Create temporary file → Write data → Move to final location
2. **Update**: Read current → Modify → Write to temporary → Move to final
3. **Delete**: Move to trash location → Confirm operation → Remove permanently

### File Locking
Concurrent access prevention using lock files:

```
project/.agx/sessions/.lock
~/.agx/.index.lock
```

Lock files contain process ID and timestamp for stale lock detection.

### Backup Strategy
- **Automatic backups** before destructive operations
- **Configurable retention** (default: 5 backups)
- **Timestamp-based naming** for backup files
- **Automatic cleanup** of old backups

## Data Migration

### Version Compatibility
Storage format includes version field for migration support:

- **Forward Compatibility**: Newer versions ignore unknown fields
- **Backward Compatibility**: Migration scripts for format upgrades
- **Breaking Changes**: Clear error messages with upgrade instructions

### Migration Process
1. **Detection**: Compare file version with current format version
2. **Backup**: Create backup of original data
3. **Transform**: Apply migration transformations
4. **Validate**: Verify migrated data integrity
5. **Cleanup**: Remove backup after successful migration

## Performance Considerations

### Indexing Strategy
- **Lazy Loading**: Load session details only when needed
- **Cached Metadata**: Keep frequently accessed data in memory
- **Efficient Queries**: Index by common search criteria
- **Batch Operations**: Group multiple file operations

### Storage Optimization
- **Compact Format**: Remove redundant data and whitespace
- **Compression**: Optional compression for large session files
- **Cleanup**: Regular removal of orphaned and expired data
- **Defragmentation**: Periodic reorganization of storage files

## Security Considerations

### File Permissions
- Session files: `600` (user read/write only)
- Directories: `700` (user access only)
- Lock files: `600` (user read/write only)

### Data Privacy
- **No sensitive data**: Avoid storing passwords, tokens, or secrets
- **Local only**: All data stored locally, no cloud synchronization
- **User control**: User can inspect and modify all stored data

### Validation
- **Schema validation**: Ensure data conforms to expected format
- **Sanitization**: Clean user input to prevent injection attacks  
- **Error handling**: Graceful handling of corrupted data

This storage format provides a robust, scalable foundation for AGX session management while maintaining simplicity and reliability.