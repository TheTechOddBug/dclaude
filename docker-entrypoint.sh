#!/bin/bash
set -e

# Start Docker daemon if in DinD mode
if [ "$DCLAUDE_DIND" = "true" ]; then
    echo "Starting Docker daemon in isolated mode..."

    # Start dockerd in the background
    sudo dockerd --host=unix:///var/run/docker.sock >/tmp/docker.log 2>&1 &

    # Wait for Docker daemon to be ready
    echo "Waiting for Docker daemon..."
    for i in {1..30}; do
        if [ -S /var/run/docker.sock ]; then
            # Socket exists, fix permissions
            sudo chmod 666 /var/run/docker.sock
            if docker info >/dev/null 2>&1; then
                echo "✓ Docker daemon ready (isolated environment)"
                break
            fi
        fi
        if [ $i -eq 30 ]; then
            echo "Error: Docker daemon failed to start"
            cat /tmp/docker.log 2>/dev/null || echo "No log file available"
            exit 1
        fi
        sleep 1
    done
fi

# Execute claude with all arguments
exec claude "$@"
