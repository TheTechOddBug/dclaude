# Dockerized Claude Code

Run Claude Code in Docker containers with automatic building and easy execution.

## Overview

This project provides a complete Docker-based setup for running Claude Code in an isolated container environment. It includes:

- **Dockerfile**: Container definition with Claude Code, Git, GitHub CLI, and Ripgrep pre-installed
- **dclaude.sh**: Smart wrapper script that auto-builds, auto-loads .env, and handles everything
- **Volume mounting**: Access to your local files from within the container
- **Session persistence**: Your conversation history persists across runs
- **Environment-based auth**: Secure API key handling via .env file
- **GitHub CLI**: Full gh CLI support for GitHub operations
- **Non-root execution**: Runs as your local user with correct file permissions
- **Git Identity**: Automatic git configuration from your local machine

## Prerequisites

- Docker installed and running
- ANTHROPIC_API_KEY environment variable
- GH_TOKEN environment variable (optional, for GitHub CLI authentication)

## Quick Start

1. **Set your API keys in .env file:**
   ```bash
   # Create or edit .env file
   echo "ANTHROPIC_API_KEY=your-anthropic-api-key" >> .env
   echo "GH_TOKEN=your-github-token" >> .env  # Optional, for GitHub operations
   ```

   Or export them in your shell:
   ```bash
   export ANTHROPIC_API_KEY='your-anthropic-api-key'
   export GH_TOKEN='your-github-token'  # Optional
   ```

2. **Run Claude Code:**
   ```bash
   # Just run it - automatically builds if needed!
   ./dclaude.sh
   ```

   **Note:** The script automatically builds the Docker image on first run. No manual build step needed!

## Usage

### Using dclaude.sh (Recommended)

The `dclaude.sh` script is the easiest way to run Claude Code. It automatically loads your `.env` file and passes all arguments to Claude:

```bash
# Interactive mode (default)
./dclaude.sh

# Display help
./dclaude.sh --help

# Check version
./dclaude.sh --version

# Run with a specific prompt
./dclaude.sh "Fix the bug in app.js"

# Use different model
./dclaude.sh --model opus "Explain this codebase"

# Continue previous conversation
./dclaude.sh --continue

# Non-interactive mode (for scripts/automation)
./dclaude.sh --print "List all files"

# Non-interactive with file write permissions
./dclaude.sh --print --permission-mode acceptEdits "Create a config.json file"

# Open a bash shell in the container
./dclaude.sh shell
```

**Special Commands:**
- `./dclaude.sh shell` - Opens a bash shell in the container for debugging and manual operations

**Permission Modes for Non-Interactive Use:**
When using `--print` mode for scripting/automation, Claude Code can't ask for permissions interactively. Use these flags:
- `--permission-mode acceptEdits` - Automatically accept file edits (recommended)
- `--permission-mode dontAsk` - Don't ask for permissions
- `--dangerously-skip-permissions` - Skip all permission checks (works with non-root user)
- For interactive use, permissions are prompted normally

**Benefits:**
- Automatically loads `.env` file (no need to export variables)
- Mounts current directory, `.gitconfig`, and `.claude` directories
- Session persistence - `--continue` works to resume conversations
- Optional GPG commit signing support (opt-in with `DCLAUDE_GPG_FORWARD=true`)
- Passes through all Claude Code arguments
- Built-in shell access for debugging
- Auto-detects interactive vs non-interactive mode (proper TTY handling)
- Works seamlessly in pipes, scripts, and automation
- Containers are named with `dclaude-` prefix for easy identification
- Automatic command logging to `dclaude.log` for audit trail
- Validates Docker installation and image availability before running
- Simple and convenient

**Automatic Checks & Build:**
The script performs these checks before running:
1. ✅ Docker is installed
2. ✅ Docker daemon is running
3. ✅ Image `dclaude:latest` exists (automatically builds if missing)
4. ✅ API key is configured (except for shell mode)

If the Docker image doesn't exist, the script will automatically build it for you. This means you can run `./dclaude.sh` immediately after cloning the repo!

**Container Naming:**
Each container gets a unique name in the format `dclaude-YYYYMMDD-HHMMSS-PID` (e.g., `dclaude-20260131-122028-23245`). This makes it easy to identify running Claude Code sessions:

```bash
# View running dclaude containers
docker ps --filter "name=dclaude"

# View all dclaude containers (including stopped)
docker ps -a --filter "name=dclaude"
```

**Command Logging:**
All commands are automatically logged to `dclaude.log` in the project directory. Each log entry includes:
- Timestamp
- Working directory where the command was executed
- Container name
- Command arguments

```bash
# View recent commands
tail -20 dclaude.log

# Monitor commands in real-time
tail -f dclaude.log

# Example log entry:
# [2026-01-31 12:37:57] PWD: /tmp | Container: dclaude-20260131-123757-35140 | Command: --version
```

### Direct Docker Commands

You can also run Docker directly:

```bash
# Run interactively
docker run -it --rm \
  -v $(pwd):/workspace \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  -e GH_TOKEN=$GH_TOKEN \
  dclaude:latest

# Run with specific command
docker run -it --rm \
  -v $(pwd):/workspace \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  -e GH_TOKEN=$GH_TOKEN \
  dclaude:latest "List all files in the project"
```

## Architecture

### Base Image
- Uses `node:20-slim` for lightweight and reliable Node.js environment
- Debian-based for easy package installation
- Automatically tagged with Claude Code version (e.g., `dclaude:2.1.27` and `dclaude:latest`)

### Non-Root User Setup
- Container runs as your local user (not root) for security
- Build-time arguments automatically set UID/GID to match your local user
- Files created in container have correct ownership on host
- Enables use of `--dangerously-skip-permissions` flag if needed
- Git config and file permissions work seamlessly

### Installed Tools
- **Claude Code**: Global npm installation of `@anthropic-ai/claude-code`
- **Git**: Version control system for repository operations
- **GitHub CLI (gh)**: Official GitHub CLI for PR management, issues, and more
- **Ripgrep (rg)**: Fast search tool for code exploration and file searching

### Volume Mounting
- Current directory mounted to `/workspace` in container
- Local `~/.gitconfig` mounted to `/root/.gitconfig` (read-only)
  - Preserves your git identity (name and email) in commits
  - All git aliases and configurations available in container
- Claude Code can read and write files in your project
- Changes persist to your local filesystem

### Authentication & Identity
- **ANTHROPIC_API_KEY**: Required for Claude Code API access
- **GH_TOKEN**: Optional, enables GitHub CLI authentication for private repos and API operations
- **Git Identity**: Automatically uses your local git configuration
  - Commits made in the container will use your name and email
  - No need to configure git identity separately
- Keys and config passed securely (not stored in the image)

### Image Metadata & Labels

Images include metadata labels for tracking installed tools:

```bash
# View all labels
docker inspect dclaude:latest --format '{{json .Config.Labels}}' | jq

# Example output:
{
  "org.opencontainers.image.title": "dclaude",
  "org.opencontainers.image.description": "Claude Code with Git, GitHub CLI, and Ripgrep",
  "tools.claude": "installed",
  "tools.gh": "installed",
  "tools.git": "installed",
  "tools.ripgrep": "installed",
  "tools.node": "20"
}

# Filter images by label
docker images --filter "label=tools.claude=installed"

# Get specific label value
docker inspect dclaude:latest --format '{{index .Config.Labels "tools.claude"}}'
```

**Installed Tool Versions:**

After building, the script displays versions of all tools:
```
Installed versions:
  • Claude Code: 2.1.27
  • GitHub CLI:  2.86.0
  • Ripgrep:     13.0.0
  • Git:         2.39.5
```

Benefits:
- 🏷️ **Metadata tracking** - Labels show what's installed in each image
- 🔍 **Queryable** - Filter and search images by tools
- 📊 **Documentation** - Self-documenting images
- 🎯 **Standards compliant** - Uses OCI image spec labels

## File Structure

```
/Users/patrickdebois/dev/dclaude/
├── Dockerfile         # Container definition with all tools pre-installed
├── dclaude.sh         # Smart wrapper script (auto-builds, auto-loads .env)
├── dclaude.log        # Command log (auto-generated)
├── .dockerignore      # Exclude unnecessary files from build context
├── .env               # Environment variables (ANTHROPIC_API_KEY, GH_TOKEN)
└── README.md          # This file
```

## Configuration

### Environment Variables

- **ANTHROPIC_API_KEY** (required): Your Anthropic API key for authentication
- **GH_TOKEN** (optional): GitHub personal access token for gh CLI authentication
  - Required for private repository access
  - Required for creating PRs, issues, and other write operations
  - Get yours at: https://github.com/settings/tokens
- **DCLAUDE_CLAUDE_VERSION** (optional, default: `latest`): Pin to a specific Claude Code version
  - Set to `latest` to always use the newest version
  - Set to a specific version like `2.1.27` to pin to that version
  - Automatically reuses existing images with matching version labels
- **DCLAUDE_GPG_FORWARD** (optional, default: `false`): Enable GPG commit signing
  - Set to `true` to mount `~/.gnupg` for commit signing

### GitHub CLI Integration

The container includes the official GitHub CLI (`gh`) for seamless GitHub operations. Claude Code can use it to:

- Create and manage pull requests
- View and create issues
- Check CI/CD status
- Clone and manage repositories
- And more

To enable GitHub CLI authentication, set the `GH_TOKEN` environment variable:

```bash
export GH_TOKEN='ghp_your_github_personal_access_token'
```

Inside the container, you can use gh commands:

```bash
# Example: Check gh CLI status
docker run --rm -e GH_TOKEN=$GH_TOKEN --entrypoint gh dclaude:latest auth status

# Example: List PRs in current repo
docker run --rm -v $(pwd):/workspace -e GH_TOKEN=$GH_TOKEN --entrypoint gh dclaude:latest pr list
```

### Volume Mounts

The following directories are automatically mounted by `dclaude.sh`:

1. **Current directory** → `/workspace` - Your project files
2. **`~/.gitconfig`** → `/home/<user>/.gitconfig` (read-only) - Git identity
3. **`~/.claude`** → `/home/<user>/.claude` - Session persistence and history
4. **`~/.gnupg`** → `/home/<user>/.gnupg` - GPG keys (opt-in, see GPG section below)

This means:
- ✅ `--continue` flag works to resume previous sessions
- ✅ Conversation history persists across container runs
- ✅ Settings and preferences are maintained
- ✅ Project-specific session data is preserved

```bash
# Example: Sessions now persist automatically
./dclaude.sh "Write a hello world script"
# Later, continue the conversation:
./dclaude.sh --continue
```

### Version Pinning

Control which Claude Code version to use with `DCLAUDE_CLAUDE_VERSION`:

**Use latest version (default):**
```bash
./dclaude.sh  # Automatically uses latest
```

**Pin to specific version:**
```bash
# One-time use
DCLAUDE_CLAUDE_VERSION=2.1.27 ./dclaude.sh

# Or add to .env file
echo "DCLAUDE_CLAUDE_VERSION=2.1.27" >> .env
./dclaude.sh
```

**How it works:**
1. Script checks if an image with the requested Claude version exists (via labels)
2. If found, uses the existing image (fast!)
3. If not found, builds a new image with that version
4. Images are tagged with version for easy reuse

**Benefits:**
- 📌 **Pin to stable versions** for production
- 🔄 **Easy rollback** if new version has issues
- 🏷️ **Automatic image reuse** - no rebuilding if version exists
- 🚀 **Test new versions** without losing old ones

**Example - Managing multiple versions:**
```bash
# Build version 2.1.27
DCLAUDE_CLAUDE_VERSION=2.1.27 ./dclaude.sh --version
# Output: 2.1.27 (Claude Code)

# Build version 2.1.26 (if available)
DCLAUDE_CLAUDE_VERSION=2.1.26 ./dclaude.sh --version
# Builds new image with 2.1.26

# Switch back to 2.1.27
DCLAUDE_CLAUDE_VERSION=2.1.27 ./dclaude.sh --version
# Output: Found existing image with Claude Code 2.1.27
# (No rebuild needed!)

# List all versions you have
docker images dclaude --format "table {{.Tag}}\t{{.CreatedAt}}"
```

### GPG Commit Signing (Opt-In)

GPG commit signing is **disabled by default** for security. To enable it, set the `DCLAUDE_GPG_FORWARD` environment variable:

**Enable GPG signing:**
```bash
# One-time use
export DCLAUDE_GPG_FORWARD=true
./dclaude.sh

# Or add to your .env file
echo "DCLAUDE_GPG_FORWARD=true" >> .env
./dclaude.sh
```

**Prerequisites:**
1. GPG keys set up on your host system
2. Git configured to use GPG signing:
   ```bash
   git config --global user.signingkey YOUR_KEY_ID
   git config --global commit.gpgsign true  # Optional: always sign
   ```

**How it works:**
- When `DCLAUDE_GPG_FORWARD=true`, your `~/.gnupg` directory is mounted into the container
- GPG keys and agent socket are accessible
- Commits can be signed just like on your host

**Test GPG signing:**
```bash
# Enable GPG forwarding
export DCLAUDE_GPG_FORWARD=true

# Verify GPG access in container
./dclaude.sh shell -c "gpg --list-secret-keys"

# Create a signed commit
./dclaude.sh "Make a change and commit it with git commit -S"
```

**Security Note:**
- GPG keys are sensitive - only enable forwarding when needed
- You may see a GPG ownership warning, which is harmless

### Image Customization

Edit the `Dockerfile` to customize the image:

```dockerfile
# Add additional tools (git and gh are already included)
RUN apt-get update && apt-get install -y vim

# Install additional npm packages
RUN npm install -g typescript
```

Then rebuild by removing the image and running dclaude.sh:
```bash
docker rmi dclaude:latest
./dclaude.sh  # Will auto-rebuild
```

## Examples

### Example 1: Quick Start with dclaude.sh
```bash
# Just run it - auto-builds and loads .env
./dclaude.sh
```

### Example 2: One-off Command
```bash
# Analyze code
./dclaude.sh "Analyze the Dockerfile and suggest improvements"

# Check for bugs
./dclaude.sh --print "Review app.js for potential bugs"
```

### Example 3: Creating Files (Non-Interactive)
```bash
# Create a file in non-interactive mode
./dclaude.sh --print --permission-mode acceptEdits "Create a hello.json file with a greeting message"

# Generate configuration files
./dclaude.sh --print --permission-mode acceptEdits "Create a package.json for a Node.js project"
```

### Example 4: Using Different Models
```bash
# Use Opus for complex tasks
./dclaude.sh --model opus "Design a new authentication system"

# Use Haiku for quick tasks
./dclaude.sh --model haiku "Fix the typo in README.md"
```

### Example 4: Shell Access for Debugging
```bash
# Open a shell using dclaude.sh
./dclaude.sh shell

# Or use make
./dclaude.sh shell

# Now you're in a bash shell inside the container
# You can explore the environment, test commands, etc.
git config --global user.name  # Check git identity
gh --version                    # Check GitHub CLI
claude --version                # Check Claude Code
ls -la /workspace              # View mounted files
```

### Example 4: Using GitHub CLI
```bash
# Set both API keys
export ANTHROPIC_API_KEY='your-anthropic-key'
export GH_TOKEN='your-github-token'

# Run Claude and ask it to create a PR
./dclaude.sh "Create a pull request for the current branch"

# Or use gh CLI directly
docker run --rm -v $(pwd):/workspace -e GH_TOKEN=$GH_TOKEN --entrypoint gh dclaude:latest pr list
```

### Example 5: Verify Git Identity
```bash
# Check that your git identity is correctly configured in the container
docker run --rm \
  -v $HOME/.gitconfig:/root/.gitconfig:ro \
  --entrypoint git dclaude:latest config --global user.name

docker run --rm \
  -v $HOME/.gitconfig:/root/.gitconfig:ro \
  --entrypoint git dclaude:latest config --global user.email
```

### Example 6: Custom Volume Mounts
```bash
docker run -it --rm \
  -v $(pwd):/workspace \
  -v ~/my-project:/external \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  -e GH_TOKEN=$GH_TOKEN \
  dclaude:latest
```

## Troubleshooting

### API Key Not Set
```
Error: ANTHROPIC_API_KEY environment variable is not set
```
**Solution**: Export your API key:
```bash
export ANTHROPIC_API_KEY='your-key'
```

### Image Not Found
```
Error: Docker image 'dclaude:latest' not found
```
**Solution**: The image will be built automatically when you run:
```bash
./dclaude.sh
```

### Permission Issues
If you encounter permission errors with files:
```bash
# Check file ownership
ls -la

# If needed, run with user mapping
docker run -it --rm \
  -v $(pwd):/workspace \
  -u $(id -u):$(id -g) \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  dclaude:latest
```

### Git Identity Not Set
If commits in the container don't have your identity:
```bash
# Check if .gitconfig is mounted correctly
docker run --rm \
  -v $HOME/.gitconfig:/root/.gitconfig:ro \
  --entrypoint ls dclaude:latest -la /root/.gitconfig

# Verify git configuration
./dclaude.sh shell
git config --global user.name
git config --global user.email
```

The `dclaude.sh` script automatically mounts your `~/.gitconfig`, so commits will use your local git identity.

### Debugging Container Issues
```bash
# Open a shell to inspect the container
./dclaude.sh shell

# Check if Claude Code is installed
which claude
claude --version

# Test Claude Code manually
claude --help
```

## Advanced Usage

### Building with Different Node Version
Edit `Dockerfile` and change the base image:
```dockerfile
FROM node:18-slim  # or node:22-slim
```

### Using Alpine for Smaller Image
```dockerfile
FROM node:20-alpine
```

Note: Alpine may have compatibility issues with some native modules.

### Persisting Claude Sessions
Mount the Claude config directory:
```bash
docker run -it --rm \
  -v $(pwd):/workspace \
  -v ~/.claude:/root/.claude \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  dclaude:latest
```

## Contributing

Feel free to submit issues or pull requests to improve this setup.

## License

This project is provided as-is for use with Claude Code.
