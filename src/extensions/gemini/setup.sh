#!/bin/bash
echo "Setup [gemini]: Initializing Gemini CLI environment"

# Pre-configure auth type so gemini-cli skips the interactive first-run wizard.
# login_method: env = API key, native = Google OAuth, auto = try env first
if [ "$ADDT_EXT_AUTO_LOGIN" = "true" ]; then
    method="${ADDT_EXT_LOGIN_METHOD:-auto}"

    if [ "$method" = "env" ] || [ "$method" = "auto" ]; then
        if [ -n "$GEMINI_API_KEY" ]; then
            echo "Setup [gemini]: Auto-configuring API key authentication"
            mkdir -p "$HOME/.gemini"
            echo '{"security":{"auth":{"selectedType":"gemini-api-key"}}}' > "$HOME/.gemini/settings.json"
        fi
    fi
fi
