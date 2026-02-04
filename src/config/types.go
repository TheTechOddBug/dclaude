package config

// ExtensionSettings holds per-extension configuration settings
type ExtensionSettings struct {
	Version         string   `yaml:"version,omitempty"`
	Automount       *bool    `yaml:"automount,omitempty"`
	FirewallAllowed []string `yaml:"firewall_allowed,omitempty"`
	FirewallDenied  []string `yaml:"firewall_denied,omitempty"`
}

// SecuritySettings holds container security configuration
type SecuritySettings struct {
	PidsLimit         *int     `yaml:"pids_limit,omitempty"`          // Max number of processes (default: 200)
	UlimitNofile      string   `yaml:"ulimit_nofile,omitempty"`       // File descriptor limit "soft:hard" (default: "4096:8192")
	UlimitNproc       string   `yaml:"ulimit_nproc,omitempty"`        // Process limit "soft:hard" (default: "256:512")
	NoNewPrivileges   *bool    `yaml:"no_new_privileges,omitempty"`   // Prevent privilege escalation (default: true)
	CapDrop           []string `yaml:"cap_drop,omitempty"`            // Capabilities to drop (default: [ALL])
	CapAdd            []string `yaml:"cap_add,omitempty"`             // Capabilities to add back (default: [CHOWN, SETUID, SETGID])
	ReadOnlyRootfs    *bool    `yaml:"read_only_rootfs,omitempty"`    // Read-only root filesystem (default: false)
	SeccompProfile    string   `yaml:"seccomp_profile,omitempty"`     // Seccomp profile: "default", "unconfined", or path
}

// GlobalConfig represents the persistent configuration stored in ~/.addt/config.yaml
type GlobalConfig struct {
	Dind             *bool    `yaml:"dind,omitempty"`
	DindMode         string   `yaml:"dind_mode,omitempty"`
	DockerCPUs       string   `yaml:"docker_cpus,omitempty"`
	DockerMemory     string   `yaml:"docker_memory,omitempty"`
	Firewall         *bool    `yaml:"firewall,omitempty"`
	FirewallMode     string   `yaml:"firewall_mode,omitempty"`
	FirewallAllowed  []string `yaml:"firewall_allowed,omitempty"`
	FirewallDenied   []string `yaml:"firewall_denied,omitempty"`
	GitHubDetect     *bool    `yaml:"github_detect,omitempty"`
	GoVersion        string   `yaml:"go_version,omitempty"`
	GPGForward       *bool    `yaml:"gpg_forward,omitempty"`
	Log              *bool    `yaml:"log,omitempty"`
	LogFile          string   `yaml:"log_file,omitempty"`
	NodeVersion      string   `yaml:"node_version,omitempty"`
	Persistent       *bool    `yaml:"persistent,omitempty"`
	PortRangeStart   *int     `yaml:"port_range_start,omitempty"`
	SSHForward       string   `yaml:"ssh_forward,omitempty"`
	UvVersion        string   `yaml:"uv_version,omitempty"`
	Workdir          string   `yaml:"workdir,omitempty"`
	WorkdirAutomount *bool    `yaml:"workdir_automount,omitempty"`

	// Per-extension configuration
	Extensions map[string]*ExtensionSettings `yaml:"extensions,omitempty"`

	// Security configuration
	Security *SecuritySettings `yaml:"security,omitempty"`
}

// Config holds all configuration options
type Config struct {
	AddtVersion              string
	NodeVersion              string
	GoVersion                string
	UvVersion                string
	EnvVars                  []string
	GitHubDetect             bool
	Ports                    []string
	PortRangeStart           int
	SSHForward               string
	GPGForward               bool
	DindMode                 string
	EnvFile                  string
	LogEnabled               bool
	LogFile                  string
	ImageName                string
	Persistent               bool              // Enable persistent container mode
	WorkdirAutomount         bool              // Auto-mount working directory
	Workdir                  string            // Override working directory (default: current directory)
	FirewallEnabled          bool              // Enable network firewall
	FirewallMode             string            // Firewall mode: strict, permissive, off
	GlobalFirewallAllowed    []string          // Global allowed domains
	GlobalFirewallDenied     []string          // Global denied domains
	ProjectFirewallAllowed   []string          // Project allowed domains
	ProjectFirewallDenied    []string          // Project denied domains
	ExtensionFirewallAllowed []string          // Extension allowed domains
	ExtensionFirewallDenied  []string          // Extension denied domains
	Mode                     string            // container or shell
	Provider                 string            // Provider type: docker or daytona
	Extensions               string            // Comma-separated list of extensions to install (e.g., "claude,codex")
	Command                  string            // Command to run instead of claude (e.g., "gt" for gastown)
	ExtensionVersions        map[string]string // Per-extension versions (e.g., {"claude": "1.0.5", "codex": "latest"})
	ExtensionAutomount       map[string]bool   // Per-extension automount control (e.g., {"claude": true, "codex": false})
	CPUs                     string            // CPU limit (e.g., "2", "0.5", "1.5")
	Memory                   string            // Memory limit (e.g., "512m", "2g", "4gb")

	// Security settings
	Security SecurityConfig
}

// SecurityConfig holds runtime security configuration with defaults applied
type SecurityConfig struct {
	PidsLimit        int      // Max number of processes (default: 200)
	UlimitNofile     string   // File descriptor limit "soft:hard" (default: "4096:8192")
	UlimitNproc      string   // Process limit "soft:hard" (default: "256:512")
	NoNewPrivileges  bool     // Prevent privilege escalation (default: true)
	CapDrop          []string // Capabilities to drop (default: [ALL])
	CapAdd           []string // Capabilities to add back (default: [CHOWN, SETUID, SETGID])
	ReadOnlyRootfs   bool     // Read-only root filesystem (default: false)
	SeccompProfile   string   // Seccomp profile (default: "")
}
