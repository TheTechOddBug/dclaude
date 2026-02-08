#!/bin/bash
echo "Setup [codex]: Initializing OpenAI Codex environment"

# login_method: env = API key, native = interactive login, auto = try env first
if [ "$ADDT_EXT_AUTO_LOGIN" = "true" ]; then
    method="${ADDT_EXT_LOGIN_METHOD:-auto}"

    if [ "$method" = "env" ] || [ "$method" = "auto" ]; then
        if [ -n "$OPENAI_API_KEY" ]; then
            echo "Setup [codex]: Auto-logging in with API key"
            printenv OPENAI_API_KEY | codex login --with-api-key
        fi
    fi
fi
