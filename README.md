# 🎯 Kamui - Advanced Session Manager

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey.svg)]()

**Kamui is an advanced session manager for Claude Code** with automatic status line integration and project-local session isolation.

## Features

- **Project-local sessions** - Each project gets its own Kamui sessions
- **Claude Code integration** - Automatic status line showing current session
- **Session isolation** - Independent Claude conversations per Kamui session
- **Interactive picker** - Browse and select sessions with rich metadata
- **Zero configuration** - Automatic setup on first use
- **Clean terminal title** - Shows `Claude - SessionName` 

## 📦 Installation

### Package Managers (Recommended)

**macOS/Linux - Homebrew**
```bash
brew install bitomule/tap/kamui
```

**Go Developers**
```bash
go install github.com/bitomule/kamui/cmd/kam@latest
```

### Manual Installation

**Option 1: Install Script (Recommended)**
```bash
git clone https://github.com/bitomule/kamui.git
cd kamui
./install.sh
```

**Option 2: Build from Source**
```bash
git clone https://github.com/bitomule/kamui.git
cd kamui
go build -o kam cmd/kam/main.go
sudo mv kam /usr/local/bin/kam
```

### Requirements

Before installing, make sure you have:
- **Claude Code CLI** - [Download from claude.ai/code](https://claude.ai/code)
- **Node.js** (for status line features)
- **Go 1.19+** (only if building from source)

### Verification

Verify your installation:
```bash
kam --version
kam --help
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

## Architecture

Kamui uses a clean, modular architecture:

- **CLI Layer** (`cmd/kam`): User interface and command handling
- **Session Management** (`internal/session`): Core business logic
- **Storage Layer** (`internal/storage`): Atomic file operations  
- **Claude Integration** (`internal/claude`): Claude Code CLI wrapper
- **Types** (`pkg/types`): Shared data structures and errors

## Troubleshooting

**Claude Code not found**
```bash
# Make sure Claude Code CLI is installed
claude --version

# Install if missing
# macOS: Download from claude.ai/code
# Or use package managers like Homebrew when available
```

**Status line not appearing**
```bash
# Manually configure Claude Code integration
kam setup

# Check if Node.js is installed (required for status line)
node --version
```

**Session not resuming correctly**
```bash
# List existing sessions to verify they exist
kam

# Check session files exist
ls ~/.claude/projects/*/
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.