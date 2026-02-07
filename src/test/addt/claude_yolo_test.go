//go:build addt

package addt

import (
	"strings"
	"testing"
)

// extractYoloResult extracts the value after "YOLO_RESULT:" from output.
func extractYoloResult(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "YOLO_RESULT:") {
			return strings.TrimPrefix(line, "YOLO_RESULT:")
		}
	}
	return ""
}

// extractArgsResult extracts the value after "ARGS_RESULT:" from output.
func extractArgsResult(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ARGS_RESULT:") {
			return strings.TrimPrefix(line, "ARGS_RESULT:")
		}
	}
	return ""
}

// Scenario: A user enables yolo mode via project config so that
// claude receives --dangerously-skip-permissions. The env var
// ADDT_EXTENSION_CLAUDE_YOLO should be set inside the container.
func TestClaudeYolo_Addt_ConfigSetsEnvVar(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, `
extensions:
  claude:
    flags:
      yolo: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Check the env var inside the container
			output, err := runShellCommand(t, dir,
				"claude", "-c", "echo YOLO_RESULT:${ADDT_EXTENSION_CLAUDE_YOLO:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractYoloResult(output)
			if result != "true" {
				t.Errorf("Expected YOLO_RESULT:true, got YOLO_RESULT:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: A user does NOT enable yolo mode. The env var
// ADDT_EXTENSION_CLAUDE_YOLO should not be set inside the container.
func TestClaudeYolo_Addt_NotSetByDefault(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Check the env var is absent inside the container
			output, err := runShellCommand(t, dir,
				"claude", "-c", "echo YOLO_RESULT:${ADDT_EXTENSION_CLAUDE_YOLO:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractYoloResult(output)
			if result != "UNSET" {
				t.Errorf("Expected YOLO_RESULT:UNSET when yolo not configured, got YOLO_RESULT:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: Inside the container, the claude extension's args.sh script
// transforms --yolo into --dangerously-skip-permissions so that claude
// receives the correct flag. Verify the transformation works.
func TestClaudeYolo_Addt_ArgsTransformation(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Run args.sh directly inside the container with --yolo flag
			// and capture the null-delimited output as newline-separated
			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo ARGS_RESULT:$(bash /usr/local/share/addt/extensions/claude/args.sh --yolo 2>/dev/null | tr '\\0' ' ')")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractArgsResult(output)
			if !strings.Contains(result, "--dangerously-skip-permissions") {
				t.Errorf("Expected args.sh to transform --yolo to --dangerously-skip-permissions, got ARGS_RESULT:%s\nFull output:\n%s",
					result, output)
			}
			if strings.Contains(result, "--yolo") {
				t.Errorf("Expected --yolo to be removed after transformation, got ARGS_RESULT:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: When yolo is enabled via config (env var), args.sh should
// inject --dangerously-skip-permissions even without --yolo on the command
// line. This is the config-driven path.
func TestClaudeYolo_Addt_ArgsTransformationViaEnv(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, `
extensions:
  claude:
    flags:
      yolo: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Run args.sh with NO --yolo flag but with env var set
			// args.sh reads ADDT_EXTENSION_CLAUDE_YOLO and injects the flag
			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo ARGS_RESULT:$(bash /usr/local/share/addt/extensions/claude/args.sh 2>/dev/null | tr '\\0' ' ')")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractArgsResult(output)
			if !strings.Contains(result, "--dangerously-skip-permissions") {
				t.Errorf("Expected args.sh to inject --dangerously-skip-permissions from env var, got ARGS_RESULT:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}
