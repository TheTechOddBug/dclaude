#!/bin/bash

# dclaude.sh - Wrapper script to run Claude Code in Docker container
# Usage: ./dclaude.sh [claude-options] [prompt]
# Special commands:
#   ./dclaude.sh shell  - Open bash shell in container
# Examples:
#   ./dclaude.sh --help
#   ./dclaude.sh --version
#   ./dclaude.sh "Fix the bug in app.js"
#   ./dclaude.sh --model opus "Explain this codebase"


set -e

# Default to latest Claude Code version, or use specified version
DCLAUDE_CLAUDE_VERSION="${DCLAUDE_CLAUDE_VERSION:-latest}"
# Default to Node 20, or use specified version (can be "20", "lts", "current", etc.)
DCLAUDE_NODE_VERSION="${DCLAUDE_NODE_VERSION:-20}"
IMAGE_NAME="dclaude:latest"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="$SCRIPT_DIR/dclaude.log"

# Function to log commands
log_command() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $*" >> "$LOG_FILE"
}

# Check for special "shell" command
OPEN_SHELL=false
if [ "$1" = "shell" ]; then
    OPEN_SHELL=true
    shift  # Remove "shell" from arguments
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    echo "Please install Docker from: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo "Error: Docker daemon is not running"
    echo "Please start Docker and try again"
    exit 1
fi

# Check if we need a specific Claude Code version or latest from npm
if [ "$DCLAUDE_CLAUDE_VERSION" = "latest" ]; then
    # Check npm for the stable version (not pre-release)
    echo "Checking npm for stable Claude Code version..."
    NPM_LATEST=$(npm info @anthropic-ai/claude-code dist-tags.stable 2>/dev/null)

    if [ -n "$NPM_LATEST" ]; then
        echo "Latest stable version on npm: $NPM_LATEST"

        # Check if we already have an image with this version
        EXISTING_IMAGE=$(docker images --filter "label=tools.claude.version=$NPM_LATEST" --format "{{.Repository}}:{{.Tag}}" | head -1)

        if [ -n "$EXISTING_IMAGE" ]; then
            echo "✓ Already have Claude Code $NPM_LATEST: $EXISTING_IMAGE"
            IMAGE_NAME="$EXISTING_IMAGE"
        else
            echo "Building image with Claude Code $NPM_LATEST..."
            DCLAUDE_CLAUDE_VERSION="$NPM_LATEST"
            IMAGE_NAME="dclaude:claude-$NPM_LATEST"
        fi
    else
        echo "Warning: Could not check npm, using existing dclaude:latest if available"
    fi
else
    # Specific version requested
    # Check if an image with this Claude version already exists
    EXISTING_IMAGE=$(docker images --filter "label=tools.claude.version=$DCLAUDE_CLAUDE_VERSION" --format "{{.Repository}}:{{.Tag}}" | head -1)

    if [ -n "$EXISTING_IMAGE" ]; then
        echo "Found existing image with Claude Code $DCLAUDE_CLAUDE_VERSION: $EXISTING_IMAGE"
        IMAGE_NAME="$EXISTING_IMAGE"
    else
        echo "No image found with Claude Code $DCLAUDE_CLAUDE_VERSION. Building now..."
        IMAGE_NAME="dclaude:claude-$DCLAUDE_CLAUDE_VERSION"
    fi
fi

# Check if Docker image exists, build if needed
if ! docker image inspect "$IMAGE_NAME" &> /dev/null; then
    echo "Docker image '$IMAGE_NAME' not found. Building it now..."
    if [ "$DCLAUDE_CLAUDE_VERSION" != "latest" ]; then
        echo "Installing Claude Code version: $DCLAUDE_CLAUDE_VERSION"
    fi
    echo "This may take a few minutes on first run..."
    echo ""

    # Build the image with user's UID/GID, Node version, and Claude version
    if docker build \
        --build-arg NODE_VERSION="$DCLAUDE_NODE_VERSION" \
        --build-arg USER_ID=$(id -u) \
        --build-arg GROUP_ID=$(id -g) \
        --build-arg USERNAME=$(whoami) \
        --build-arg CLAUDE_VERSION="$DCLAUDE_CLAUDE_VERSION" \
        -t "$IMAGE_NAME" "$SCRIPT_DIR"; then
        echo ""
        echo "✓ Image built successfully!"
        echo ""
        echo "Detecting tool versions..."

        # Get versions from the built image
        CLAUDE_VERSION=$(docker run --rm --entrypoint claude "$IMAGE_NAME" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        GH_VERSION=$(docker run --rm --entrypoint gh "$IMAGE_NAME" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        RG_VERSION=$(docker run --rm --entrypoint rg "$IMAGE_NAME" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        GIT_VERSION=$(docker run --rm --entrypoint git "$IMAGE_NAME" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        NODE_VERSION_ACTUAL=$(docker run --rm --entrypoint node "$IMAGE_NAME" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)

        # Create a temporary Dockerfile to add version labels
        TEMP_DOCKERFILE=$(mktemp)
        cat > "$TEMP_DOCKERFILE" << EOF
FROM $IMAGE_NAME
LABEL tools.claude.version="$CLAUDE_VERSION"
LABEL tools.gh.version="$GH_VERSION"
LABEL tools.ripgrep.version="$RG_VERSION"
LABEL tools.git.version="$GIT_VERSION"
LABEL tools.node.version="$NODE_VERSION_ACTUAL"
EOF

        # Rebuild with version labels (fast - just adds metadata layer)
        echo "Adding version labels..."
        if docker build -f "$TEMP_DOCKERFILE" -t "$IMAGE_NAME" "$SCRIPT_DIR" &> /dev/null; then
            # Also tag as dclaude:latest if we built the latest version
            if [ "$DCLAUDE_CLAUDE_VERSION" = "latest" ]; then
                docker tag "$IMAGE_NAME" "dclaude:latest" &> /dev/null
            fi
            # Tag with claude version for easy reference
            if [ -n "$CLAUDE_VERSION" ]; then
                docker tag "$IMAGE_NAME" "dclaude:claude-$CLAUDE_VERSION" &> /dev/null
            fi

            rm "$TEMP_DOCKERFILE"
            echo ""
            echo "Installed versions:"
            [ -n "$NODE_VERSION_ACTUAL" ] && echo "  • Node.js:     $NODE_VERSION_ACTUAL"
            [ -n "$CLAUDE_VERSION" ] && echo "  • Claude Code: $CLAUDE_VERSION"
            [ -n "$GH_VERSION" ] && echo "  • GitHub CLI:  $GH_VERSION"
            [ -n "$RG_VERSION" ] && echo "  • Ripgrep:     $RG_VERSION"
            [ -n "$GIT_VERSION" ] && echo "  • Git:         $GIT_VERSION"
            echo ""
            echo "Image tagged as: $IMAGE_NAME"
            [ "$DCLAUDE_CLAUDE_VERSION" = "latest" ] && echo "Also tagged as: dclaude:latest"
            [ -n "$CLAUDE_VERSION" ] && echo "Also tagged as: dclaude:claude-$CLAUDE_VERSION"
            echo ""
            echo "View labels: docker inspect $IMAGE_NAME --format '{{json .Config.Labels}}' | jq"
            echo ""
        else
            rm "$TEMP_DOCKERFILE"
            echo "Warning: Could not add version labels, but image is functional"
            echo ""
        fi
    else
        echo ""
        echo "Error: Failed to build Docker image"
        echo "Please check the Dockerfile and try again"
        exit 1
    fi
fi

# Load .env file if it exists
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# Check if ANTHROPIC_API_KEY is set (not required for shell mode)
if [ "$OPEN_SHELL" = false ] && [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "Error: ANTHROPIC_API_KEY environment variable is not set"
    echo "Please set it with: export ANTHROPIC_API_KEY='your-key'"
    echo "Or add it to your .env file"
    exit 1
fi

# Generate unique container name
CONTAINER_NAME="dclaude-$(date +%Y%m%d-%H%M%S)-$$"

# Build docker run command
# Detect if we're running in an interactive terminal
if [ -t 0 ] && [ -t 1 ]; then
    # Interactive mode - use -it flags
    DOCKER_CMD="docker run -it --rm --name $CONTAINER_NAME"
else
    # Non-interactive mode (piped, scripted, etc) - don't use TTY
    DOCKER_CMD="docker run -i --rm --name $CONTAINER_NAME"
fi

# Mount current directory
DOCKER_CMD="$DOCKER_CMD -v $(pwd):/workspace"

# Mount .gitconfig for git identity
if [ -f "$HOME/.gitconfig" ]; then
    DOCKER_CMD="$DOCKER_CMD -v $HOME/.gitconfig:/home/$(whoami)/.gitconfig:ro"
fi

# Mount .claude directory for session persistence
if [ -d "$HOME/.claude" ]; then
    DOCKER_CMD="$DOCKER_CMD -v $HOME/.claude:/home/$(whoami)/.claude"
fi

# Mount .gnupg directory for GPG commit signing support (opt-in)
if [ "$DCLAUDE_GPG_FORWARD" = "true" ] && [ -d "$HOME/.gnupg" ]; then
    DOCKER_CMD="$DOCKER_CMD -v $HOME/.gnupg:/home/$(whoami)/.gnupg"
    # Set GPG_TTY for proper TTY handling
    DOCKER_CMD="$DOCKER_CMD -e GPG_TTY=/dev/console"
fi

# Pass environment variables (if set)
if [ -n "$ANTHROPIC_API_KEY" ]; then
    DOCKER_CMD="$DOCKER_CMD -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY"
fi

# Add GH_TOKEN if it's set
if [ -n "$GH_TOKEN" ]; then
    DOCKER_CMD="$DOCKER_CMD -e GH_TOKEN=$GH_TOKEN"
fi

# Handle shell mode or normal mode
if [ "$OPEN_SHELL" = true ]; then
    # Override entrypoint for shell mode
    echo "Opening bash shell in container..."
    DOCKER_CMD="$DOCKER_CMD --entrypoint /bin/bash $IMAGE_NAME"
    # Add any remaining arguments
    if [ $# -gt 0 ]; then
        DOCKER_CMD="$DOCKER_CMD $@"
    fi
else
    # Normal mode - run claude command
    DOCKER_CMD="$DOCKER_CMD $IMAGE_NAME"
    # Add all arguments passed to this script
    if [ $# -gt 0 ]; then
        DOCKER_CMD="$DOCKER_CMD $@"
    fi
fi

# Log the command
log_command "PWD: $(pwd) | Container: $CONTAINER_NAME | Command: $@"

# Execute the command
eval $DOCKER_CMD
