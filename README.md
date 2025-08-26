# Kamui - Advanced Session Manager

🎯 **Kamui is an advanced session manager for Claude Code** with automatic status line integration and project-local session isolation.

## Features

- **Project-local sessions** - Each project gets its own Kamui sessions
- **Claude Code integration** - Automatic status line showing current session
- **Session isolation** - Independent Claude conversations per Kamui session
- **Interactive picker** - Browse and select sessions with rich metadata
- **Zero configuration** - Automatic setup on first use
- **Clean terminal title** - Shows `Claude - SessionName` 

## Installation

```bash
# Clone the repository
git clone https://github.com/davidcollado/kamui.git
cd kamui

# Run the installation script
./install.sh
```

Or install manually:
```bash
go build -o kam cmd/kam/main.go
sudo cp kam /usr/local/bin/kam
```

## Quick Start

```bash
# Create or resume a session
kam MyProject

# Interactive session picker
kam

# Manual Claude Code setup (optional)
kam setup
```

## How It Works

### Session Management
- Sessions are stored in `.claude/kamui-sessions/` in each project
- Each Kamui session maps to an independent Claude Code conversation
- Sessions persist across runs and show rich metadata

### Claude Code Integration
- **Automatic setup** on first use
- **Status line** shows `🎯 SessionName • ProjectName`
- **Terminal title** shows `Claude - SessionName`
- Uses Claude Code's built-in `statusLine` feature

### Session Isolation
Kamui ensures each session name gets its own Claude conversation:
- `kam Tasks` in ProjectA → Independent Claude session
- `kam Tasks` in ProjectB → Different Claude session  
- `kam Development` in ProjectA → Another independent session

## Commands

- `kam <session-name>` - Create or resume a session
- `kam` - Interactive session picker
- `kam setup` - Configure Claude Code integration
- `kam list` - List all sessions
- `kam info <session>` - Show session details
- `kam complete <session>` - Mark session as completed

## Requirements

- **Go** 1.19+ for building
- **Claude Code** CLI installed and configured
- **Node.js** for status line script

## Architecture

Kamui uses a clean, modular architecture:

- **CLI Layer** (`cmd/kam`): User interface and command handling
- **Session Management** (`internal/session`): Core business logic
- **Storage Layer** (`internal/storage`): Atomic file operations  
- **Claude Integration** (`internal/claude`): Claude Code CLI wrapper
- **Types** (`pkg/types`): Shared data structures and errors