#!/bin/bash
echo "Setup [gemini]: Initializing Gemini CLI environment"

# Pre-configure auth type so gemini-cli skips the interactive first-run wizard.
# When GEMINI_API_KEY is set, this tells the CLI to use it directly.
if [ -n "$GEMINI_API_KEY" ]; then
    mkdir -p "$HOME/.gemini"
    echo '{"security":{"auth":{"selectedType":"gemini-api-key"}}}' > "$HOME/.gemini/settings.json"
fi
