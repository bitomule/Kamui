#!/bin/bash
# Kamui Installation Script

set -e

echo "🎯 Installing Kamui - Advanced Session Manager"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is required but not installed"
    echo "   Please install Go from: https://golang.org/downloads/"
    exit 1
fi

# Check if claude is installed
if ! command -v claude &> /dev/null; then
    echo "❌ Claude Code is required but not installed"
    echo "   Please install Claude Code from: https://claude.ai/code"
    exit 1
fi

echo "✅ Prerequisites check passed"
echo ""

# Build Kamui
echo "🔨 Building Kamui..."
go build -o kam cmd/kam/main.go

# Make executable
chmod +x kam

# Optionally install to system PATH
read -p "Install Kamui to /usr/local/bin? (y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    sudo cp kam /usr/local/bin/kam
    echo "✅ Kamui installed to /usr/local/bin/kam"
    echo "   You can now run 'kam' from anywhere"
else
    echo "✅ Kamui built as './kam'"
    echo "   Run with: ./kam <session-name>"
fi

echo ""
echo "🎉 Kamui installation complete!"
echo ""
echo "Quick start:"
echo "  kam setup           # Configure Claude Code integration (optional)"
echo "  kam MyProject       # Create/resume a session"
echo "  kam                 # Interactive session picker"
echo ""
echo "The Claude Code status line will be configured automatically"
echo "on first use to show your Kamui session info."