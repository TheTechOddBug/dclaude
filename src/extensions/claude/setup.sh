#!/bin/bash
echo "Setup [claude]: Initializing Claude Code environment"

CLAUDE_DIR="$HOME/.claude"
CLAUDE_JSON="$HOME/.claude.json"
SETTINGS_JSON="$CLAUDE_DIR/settings.json"
API_HELPER="$CLAUDE_DIR/api-key-helper.sh"

# If ANTHROPIC_API_KEY is set, configure Claude Code to use it
if [ -n "$ANTHROPIC_API_KEY" ]; then
    # Create .claude directory
    mkdir -p "$CLAUDE_DIR"

    # Create API key helper script
    if [ ! -f "$API_HELPER" ]; then
        echo "Setup [claude]: Creating API key helper script"
        cat > "$API_HELPER" << 'HELPER'
#!/bin/bash
echo "$ANTHROPIC_API_KEY"
HELPER
        chmod +x "$API_HELPER"
    fi

    # Create settings.json with apiKeyHelper
    if [ ! -f "$SETTINGS_JSON" ]; then
        echo "Setup [claude]: Creating $SETTINGS_JSON with apiKeyHelper"
        cat > "$SETTINGS_JSON" << SETTINGS
{
  "apiKeyHelper": "$API_HELPER"
}
SETTINGS
    elif ! grep -q '"apiKeyHelper"' "$SETTINGS_JSON" 2>/dev/null; then
        echo "Setup [claude]: Adding apiKeyHelper to $SETTINGS_JSON"
        if command -v jq &> /dev/null; then
            tmp=$(mktemp)
            jq --arg helper "$API_HELPER" '. + {"apiKeyHelper": $helper}' "$SETTINGS_JSON" > "$tmp" && mv "$tmp" "$SETTINGS_JSON"
        fi
    fi

    # Create/update .claude.json to skip onboarding
    if [ ! -f "$CLAUDE_JSON" ]; then
        echo "Setup [claude]: Creating $CLAUDE_JSON (skipping onboarding)"
        cat > "$CLAUDE_JSON" << 'EOF'
{
  "hasCompletedOnboarding": true,
  "hasTrustDialogAccepted": true
}
EOF
    elif ! grep -q '"hasCompletedOnboarding"' "$CLAUDE_JSON" 2>/dev/null; then
        echo "Setup [claude]: Adding hasCompletedOnboarding to $CLAUDE_JSON"
        if command -v jq &> /dev/null; then
            tmp=$(mktemp)
            jq '. + {"hasCompletedOnboarding": true, "hasTrustDialogAccepted": true}' "$CLAUDE_JSON" > "$tmp" && mv "$tmp" "$CLAUDE_JSON"
        else
            sed -i 's/}$/,"hasCompletedOnboarding":true,"hasTrustDialogAccepted":true}/' "$CLAUDE_JSON"
        fi
    fi

    echo "Setup [claude]: Configured for API key authentication"
fi
