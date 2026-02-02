#!/bin/bash
# Claude Sneakpeek installation
# https://github.com/mikekelly/claude-sneakpeek

set -e

echo "Installing Claude Sneakpeek..."

# Install using npx quick installer with custom name
npx @realmikekelly/claude-sneakpeek quick --name claudesp

# Verify installation
if command -v claudesp &> /dev/null; then
    echo "Claude Sneakpeek installed successfully"
else
    # Check if it's in ~/.local/bin
    if [ -f "$HOME/.local/bin/claudesp" ]; then
        echo "Claude Sneakpeek installed to ~/.local/bin/claudesp"
    else
        echo "Warning: claudesp command not found after installation"
    fi
fi
