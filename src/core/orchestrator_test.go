package core

import (
	"strings"
	"testing"

	"github.com/jedi4ever/addt/provider"
)

// mockProvider implements provider.Provider for testing
type mockProvider struct{}

func (m *mockProvider) Initialize(cfg *provider.Config) error              { return nil }
func (m *mockProvider) Run(spec *provider.RunSpec) error                   { return nil }
func (m *mockProvider) Shell(spec *provider.RunSpec) error                 { return nil }
func (m *mockProvider) Cleanup() error                                     { return nil }
func (m *mockProvider) Exists(name string) bool                            { return false }
func (m *mockProvider) IsRunning(name string) bool                         { return false }
func (m *mockProvider) Start(name string) error                            { return nil }
func (m *mockProvider) Stop(name string) error                             { return nil }
func (m *mockProvider) Remove(name string) error                           { return nil }
func (m *mockProvider) List() ([]provider.Environment, error)              { return nil, nil }
func (m *mockProvider) GeneratePersistentName() string                     { return "test-persistent" }
func (m *mockProvider) GenerateEphemeralName() string                      { return "test-ephemeral" }
func (m *mockProvider) GetStatus(cfg *provider.Config, name string) string { return "test" }
func (m *mockProvider) GetName() string                                    { return "mock" }
func (m *mockProvider) GetExtensionEnvVars(imageName string) []string      { return nil }
func (m *mockProvider) DetermineImageName() string                         { return "test-image" }
func (m *mockProvider) BuildIfNeeded(rebuild bool, rebuildBase bool) error { return nil }

func TestBuildPorts_Empty(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{},
		PortRangeStart: 30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	ports := orch.buildPorts()

	if len(ports) != 0 {
		t.Errorf("Expected 0 ports, got %d", len(ports))
	}
}

func TestBuildPorts_Single(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{"3000"},
		PortRangeStart: 30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	ports := orch.buildPorts()

	if len(ports) != 1 {
		t.Fatalf("Expected 1 port, got %d", len(ports))
	}

	if ports[0].Container != 3000 {
		t.Errorf("Container port = %d, want 3000", ports[0].Container)
	}

	if ports[0].Host < 30000 {
		t.Errorf("Host port = %d, want >= 30000", ports[0].Host)
	}
}

func TestBuildPorts_Multiple(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{"3000", "8080", "5432"},
		PortRangeStart: 30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	ports := orch.buildPorts()

	if len(ports) != 3 {
		t.Fatalf("Expected 3 ports, got %d", len(ports))
	}

	expectedContainerPorts := []int{3000, 8080, 5432}
	for i, expectedPort := range expectedContainerPorts {
		if ports[i].Container != expectedPort {
			t.Errorf("Port %d: Container = %d, want %d", i, ports[i].Container, expectedPort)
		}
	}

	// Host ports should be unique and >= 30000
	usedPorts := make(map[int]bool)
	for i, port := range ports {
		if port.Host < 30000 {
			t.Errorf("Port %d: Host = %d, want >= 30000", i, port.Host)
		}
		if usedPorts[port.Host] {
			t.Errorf("Port %d: Host port %d is duplicated", i, port.Host)
		}
		usedPorts[port.Host] = true
	}
}

func TestBuildPorts_WhitespaceHandling(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{" 3000 ", "8080", " 5432"},
		PortRangeStart: 30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	ports := orch.buildPorts()

	if len(ports) != 3 {
		t.Fatalf("Expected 3 ports, got %d", len(ports))
	}

	// Should trim whitespace
	if ports[0].Container != 3000 {
		t.Errorf("Port 0: Container = %d, want 3000 (whitespace not trimmed)", ports[0].Container)
	}
}

func TestBuildEnvironment_PortMap_Empty(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{},
		PortRangeStart: 30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	env := orch.buildEnvironment()

	if _, ok := env["ADDT_PORT_MAP"]; ok {
		t.Error("ADDT_PORT_MAP should not be set when no ports configured")
	}
}

func TestBuildEnvironment_PortMap_Single(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{"3000"},
		PortRangeStart: 30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	env := orch.buildEnvironment()

	portMap, ok := env["ADDT_PORT_MAP"]
	if !ok {
		t.Fatal("ADDT_PORT_MAP not set")
	}

	// Should be in format "containerPort:hostPort"
	if !strings.HasPrefix(portMap, "3000:") {
		t.Errorf("ADDT_PORT_MAP = %q, want prefix '3000:'", portMap)
	}
}

func TestBuildEnvironment_PortMap_Multiple(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{"3000", "8080"},
		PortRangeStart: 30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	env := orch.buildEnvironment()

	portMap, ok := env["ADDT_PORT_MAP"]
	if !ok {
		t.Fatal("ADDT_PORT_MAP not set")
	}

	// Should be comma-separated
	parts := strings.Split(portMap, ",")
	if len(parts) != 2 {
		t.Errorf("Expected 2 port mappings, got %d: %q", len(parts), portMap)
	}

	// First mapping should start with "3000:"
	if !strings.HasPrefix(parts[0], "3000:") {
		t.Errorf("First mapping = %q, want prefix '3000:'", parts[0])
	}

	// Second mapping should start with "8080:"
	if !strings.HasPrefix(parts[1], "8080:") {
		t.Errorf("Second mapping = %q, want prefix '8080:'", parts[1])
	}
}

func TestBuildEnvironment_PortMap_Format(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{"3000", "8080", "5432"},
		PortRangeStart: 30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	env := orch.buildEnvironment()

	portMap := env["ADDT_PORT_MAP"]

	// Verify format: "containerPort:hostPort,containerPort:hostPort,..."
	parts := strings.Split(portMap, ",")
	for i, part := range parts {
		mapping := strings.Split(part, ":")
		if len(mapping) != 2 {
			t.Errorf("Mapping %d = %q, expected format 'container:host'", i, part)
		}
	}

	t.Logf("ADDT_PORT_MAP = %q", portMap)
}

func TestBuildEnvironment_Firewall(t *testing.T) {
	cfg := &provider.Config{
		FirewallEnabled: true,
		FirewallMode:    "allowlist",
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	env := orch.buildEnvironment()

	if env["ADDT_FIREWALL_ENABLED"] != "true" {
		t.Errorf("ADDT_FIREWALL_ENABLED = %q, want 'true'", env["ADDT_FIREWALL_ENABLED"])
	}

	if env["ADDT_FIREWALL_MODE"] != "allowlist" {
		t.Errorf("ADDT_FIREWALL_MODE = %q, want 'allowlist'", env["ADDT_FIREWALL_MODE"])
	}
}

func TestBuildEnvironment_Command(t *testing.T) {
	cfg := &provider.Config{
		Command: "codex",
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	env := orch.buildEnvironment()

	if env["ADDT_COMMAND"] != "codex" {
		t.Errorf("ADDT_COMMAND = %q, want 'codex'", env["ADDT_COMMAND"])
	}
}

func TestBuildEnvironment_TerminalVars(t *testing.T) {
	cfg := &provider.Config{}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	env := orch.buildEnvironment()

	// COLUMNS and LINES should always be set
	if env["COLUMNS"] == "" {
		t.Error("COLUMNS not set")
	}

	if env["LINES"] == "" {
		t.Error("LINES not set")
	}
}

func TestBuildVolumes_AutomountEnabled(t *testing.T) {
	cfg := &provider.Config{
		WorkdirAutomount: true,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	volumes := orch.buildVolumes("/home/user/project")

	if len(volumes) != 1 {
		t.Fatalf("Expected 1 volume, got %d", len(volumes))
	}

	if volumes[0].Source != "/home/user/project" {
		t.Errorf("Volume source = %q, want '/home/user/project'", volumes[0].Source)
	}

	if volumes[0].Target != "/workspace" {
		t.Errorf("Volume target = %q, want '/workspace'", volumes[0].Target)
	}

	if volumes[0].ReadOnly {
		t.Error("Volume should not be read-only")
	}
}

func TestBuildVolumes_AutomountDisabled(t *testing.T) {
	cfg := &provider.Config{
		WorkdirAutomount: false,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	volumes := orch.buildVolumes("/home/user/project")

	if len(volumes) != 0 {
		t.Errorf("Expected 0 volumes when automount disabled, got %d", len(volumes))
	}
}

func TestBuildRunSpec_Basic(t *testing.T) {
	cfg := &provider.Config{
		ImageName:        "test-image",
		WorkdirAutomount: true,
		PortRangeStart:   30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	spec := orch.buildRunSpec("test-container", []string{"--help"}, false)

	if spec.Name != "test-container" {
		t.Errorf("Name = %q, want 'test-container'", spec.Name)
	}

	if spec.ImageName != "test-image" {
		t.Errorf("ImageName = %q, want 'test-image'", spec.ImageName)
	}

	if len(spec.Args) != 1 || spec.Args[0] != "--help" {
		t.Errorf("Args = %v, want ['--help']", spec.Args)
	}
}

func TestBuildRunSpec_ShellMode(t *testing.T) {
	cfg := &provider.Config{
		ImageName:        "test-image",
		WorkdirAutomount: true,
		PortRangeStart:   30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	// Shell mode with no args
	spec := orch.buildRunSpec("test-container", []string{}, true)

	if len(spec.Args) != 0 {
		t.Errorf("Args = %v, want empty for shell mode", spec.Args)
	}

	// Shell mode with args
	spec = orch.buildRunSpec("test-container", []string{"-c", "ls"}, true)

	if len(spec.Args) != 2 {
		t.Errorf("Args = %v, want ['-c', 'ls'] for shell mode with args", spec.Args)
	}
}

func TestBuildRunSpec_Persistent(t *testing.T) {
	cfg := &provider.Config{
		ImageName:        "test-image",
		Persistent:       true,
		WorkdirAutomount: true,
		PortRangeStart:   30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	spec := orch.buildRunSpec("test-container", []string{}, false)

	if !spec.Persistent {
		t.Error("Persistent should be true")
	}
}

func TestBuildRunSpec_SSHAndGPG(t *testing.T) {
	cfg := &provider.Config{
		ImageName:        "test-image",
		SSHForward:       "keys",
		GPGForward:       true,
		WorkdirAutomount: true,
		PortRangeStart:   30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	spec := orch.buildRunSpec("test-container", []string{}, false)

	if spec.SSHForward != "keys" {
		t.Errorf("SSHForward = %q, want 'keys'", spec.SSHForward)
	}

	if !spec.GPGForward {
		t.Error("GPGForward should be true")
	}
}

func TestBuildRunSpec_DindMode(t *testing.T) {
	cfg := &provider.Config{
		ImageName:        "test-image",
		DindMode:         "isolated",
		WorkdirAutomount: true,
		PortRangeStart:   30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	spec := orch.buildRunSpec("test-container", []string{}, false)

	if spec.DindMode != "isolated" {
		t.Errorf("DindMode = %q, want 'isolated'", spec.DindMode)
	}
}

func TestBuildRunSpec_Resources(t *testing.T) {
	cfg := &provider.Config{
		ImageName:        "test-image",
		CPUs:             "2",
		Memory:           "4g",
		WorkdirAutomount: true,
		PortRangeStart:   30000,
	}
	orch := NewOrchestrator(&mockProvider{}, cfg)

	spec := orch.buildRunSpec("test-container", []string{}, false)

	if spec.CPUs != "2" {
		t.Errorf("CPUs = %q, want '2'", spec.CPUs)
	}

	if spec.Memory != "4g" {
		t.Errorf("Memory = %q, want '4g'", spec.Memory)
	}
}
