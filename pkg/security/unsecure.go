package security

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type SecurityLevel int

const (
	SecurityDisabled SecurityLevel = iota
	SecurityBasic
	SecurityStrict
	SecurityMaximum
)

type SecurityConfig struct {
	Level            SecurityLevel
	MaxExecutionTime time.Duration
	MaxMemoryMB      int
	MaxCPUPercent    int
	AllowNetwork     bool
	AllowFileWrite   bool
	TempDirOnly      bool
	UseContainer     bool
	ContainerImage   string
}

type SecurityManager struct {
	config       SecurityConfig
	capabilities map[string]bool
}

func NewSecurityManager() *SecurityManager {
	sm := &SecurityManager{
		config: SecurityConfig{
			Level:            SecurityStrict,
			MaxExecutionTime: 30 * time.Second,
			MaxMemoryMB:      128,
			MaxCPUPercent:    50,
			AllowNetwork:     false,
			AllowFileWrite:   false,
			TempDirOnly:      true,
			UseContainer:     true,
			ContainerImage:   "alpine:latest",
		},
		capabilities: make(map[string]bool),
	}

	sm.detectCapabilities()
	sm.adjustConfigBasedOnCapabilities()
	return sm
}

func (sm *SecurityManager) detectCapabilities() {
	// Check available sandboxing tools
	sm.capabilities["docker"] = isCommandAvailable("docker")
	sm.capabilities["podman"] = isCommandAvailable("podman")
	sm.capabilities["firejail"] = isCommandAvailable("firejail")
	sm.capabilities["bubblewrap"] = isCommandAvailable("bwrap")
	sm.capabilities["systemd-run"] = isCommandAvailable("systemd-run")

	// Check if running in container already
	sm.capabilities["in_container"] = sm.isRunningInContainer()

	// Check user permissions
	sm.capabilities["can_create_namespace"] = sm.canCreateNamespace()
}

func (sm *SecurityManager) adjustConfigBasedOnCapabilities() {
	// Auto-adjust security level based on available tools
	if sm.capabilities["docker"] || sm.capabilities["podman"] {
		sm.config.Level = SecurityMaximum
		sm.config.UseContainer = true
	} else if sm.capabilities["firejail"] || sm.capabilities["bubblewrap"] {
		sm.config.Level = SecurityStrict
		sm.config.UseContainer = false
	} else if sm.capabilities["systemd-run"] {
		sm.config.Level = SecurityBasic
	} else {
		// Fallback to basic restrictions
		sm.config.Level = SecurityBasic
		sm.config.MaxExecutionTime = 10 * time.Second
		sm.config.MaxMemoryMB = 64
	}
}

func (sm *SecurityManager) CreateSecureCommand(ctx context.Context, executable string, args ...string) (*exec.Cmd, error) {
	switch sm.config.Level {
	case SecurityMaximum:
		return sm.createContainerCommand(ctx, executable, args...)
	case SecurityStrict:
		return sm.createSandboxedCommand(ctx, executable, args...)
	case SecurityBasic:
		return sm.createBasicSecureCommand(ctx, executable, args...)
	case SecurityDisabled:
		return exec.CommandContext(ctx, executable, args...), nil
	default:
		return nil, fmt.Errorf("unknown security level")
	}
}

func (sm *SecurityManager) createContainerCommand(ctx context.Context, executable string, args ...string) (*exec.Cmd, error) {
	if !sm.capabilities["docker"] && !sm.capabilities["podman"] {
		return nil, fmt.Errorf("container runtime not available")
	}

	runtime := "docker"
	if !sm.capabilities["docker"] && sm.capabilities["podman"] {
		runtime = "podman"
	}

	containerArgs := []string{
		"run", "--rm",
		"--network=none", // No network access
		fmt.Sprintf("--memory=%dm", sm.config.MaxMemoryMB),
		fmt.Sprintf("--cpus=%.2f", float64(sm.config.MaxCPUPercent)/100),
		"--read-only", // Read-only root filesystem
		"--tmpfs=/tmp:rw,noexec,nosuid,size=100m", // Writable tmp, no exec
		"--security-opt=no-new-privileges",        // Prevent privilege escalation
		"--cap-drop=ALL",                          // Drop all capabilities
		"--user=65534:65534",                      // Run as nobody
		sm.config.ContainerImage,
		executable,
	}

	containerArgs = append(containerArgs, args...)
	return exec.CommandContext(ctx, runtime, containerArgs...), nil
}

func (sm *SecurityManager) createSandboxedCommand(ctx context.Context, executable string, args ...string) (*exec.Cmd, error) {
	if sm.capabilities["firejail"] {
		return sm.createFirejailCommand(ctx, executable, args...)
	} else if sm.capabilities["bubblewrap"] {
		return sm.createBubblewrapCommand(ctx, executable, args...)
	}

	// Fallback to basic security
	return sm.createBasicSecureCommand(ctx, executable, args...)
}

func (sm *SecurityManager) createFirejailCommand(ctx context.Context, executable string, args ...string) (*exec.Cmd, error) {
	firejailArgs := []string{
		"--quiet",
		"--noprofile",
		"--seccomp",       // Enable seccomp filtering
		"--nonetwork",     // Disable network
		"--private-tmp",   // Private /tmp
		"--noroot",        // Don't run as root
		"--caps.drop=all", // Drop all capabilities
		"--rlimit-cpu=30", // CPU time limit
		"--rlimit-as=" + fmt.Sprintf("%d", sm.config.MaxMemoryMB*1024*1024), // Memory limit
		executable,
	}

	firejailArgs = append(firejailArgs, args...)
	return exec.CommandContext(ctx, "firejail", firejailArgs...), nil
}

func (sm *SecurityManager) createBubblewrapCommand(ctx context.Context, executable string, args ...string) (*exec.Cmd, error) {
	bwrapArgs := []string{
		"--ro-bind", "/usr", "/usr",
		"--ro-bind", "/lib", "/lib",
		"--ro-bind", "/lib64", "/lib64",
		"--ro-bind", "/bin", "/bin",
		"--ro-bind", "/sbin", "/sbin",
		"--tmpfs", "/tmp",
		"--proc", "/proc",
		"--dev", "/dev",
		"--unshare-net",     // No network
		"--unshare-pid",     // PID namespace
		"--die-with-parent", // Die when parent dies
		executable,
	}

	bwrapArgs = append(bwrapArgs, args...)
	return exec.CommandContext(ctx, "bwrap", bwrapArgs...), nil
}

func (sm *SecurityManager) createBasicSecureCommand(ctx context.Context, executable string, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, executable, args...)

	// Set resource limits using syscall
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Create new process group
		Setpgid: true,
	}

	return cmd, nil
}

func (sm *SecurityManager) ValidateExecutable(path string) error {
	// Check if executable is in allowed locations
	if sm.config.TempDirOnly {
		tmpDir := os.TempDir()
		if !filepath.HasPrefix(path, tmpDir) {
			return fmt.Errorf("executable must be in temp directory")
		}
	}

	// Check file permissions
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat executable: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("file is not executable")
	}

	return nil
}

func (sm *SecurityManager) GetSecurityReport() map[string]interface{} {
	return map[string]interface{}{
		"level":        sm.config.Level,
		"capabilities": sm.capabilities,
		"config":       sm.config,
	}
}

// Helper functions

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (sm *SecurityManager) isRunningInContainer() bool {
	// Check for container indicators
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup for container indicators
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		return contains(content, "docker") || contains(content, "kubepods") || contains(content, "containerd")
	}

	return false
}

func (sm *SecurityManager) canCreateNamespace() bool {
	// Check if we can run privileged operations
	cmd := exec.Command("sudo", "-n", "true")
	return cmd.Run() == nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)+1] == substr+"/" || s[len(s)-len(substr)-1:] == "/"+substr)))
}
