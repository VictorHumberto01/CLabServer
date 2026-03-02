package security

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type SecurityLevel int

const (
	SecurityMaximum SecurityLevel = iota
)

type SecurityConfig struct {
	Level            SecurityLevel
	MaxExecutionTime time.Duration
	MaxMemoryMB      int
	MaxCPUPercent    int
	AllowNetwork     bool
	AllowFileWrite   bool
	TempDirOnly      bool
	WorkspaceDir     string
	WorkspaceRO      bool
	UseContainer     bool
	ContainerImage   string
	MaxProcesses     int
	MaxFileSizeMB    int
}

type SecurityManager struct {
	config       SecurityConfig
	capabilities map[string]bool
}

var DefaultManager *SecurityManager

func init() {
	DefaultManager = NewSecurityManager()
}

func NewSecurityManager() *SecurityManager {
	sm := &SecurityManager{
		config: SecurityConfig{
			Level:            SecurityMaximum,
			MaxExecutionTime: 30 * time.Second,
			MaxMemoryMB:      128,
			MaxCPUPercent:    50,
			AllowNetwork:     false,
			AllowFileWrite:   false,
			TempDirOnly:      true,
			WorkspaceDir:     "",
			UseContainer:     true,
			ContainerImage:   "gcc:latest",
			MaxProcesses:     64,
			MaxFileSizeMB:    50,
		},
		capabilities: make(map[string]bool),
	}

	sm.detectCapabilities()
	sm.adjustConfigBasedOnCapabilities()
	go sm.startOrphanCleanupRoutine()
	return sm
}

// How the new security engine works:
// This sandbox engine uses a "Docker-in-Docker" (DinD) isolation architecture via
// the mounted host socket (/var/run/docker.sock) rather than inside-container limits.
//
//  1. Compile Phase: A completely new, isolated Docker container is started. The workspace
//     directory (containing the C code) is copied INTO this container via `docker cp`.
//     The compiler (`gcc`) output goes to this directory. The container is then destroyed,
//     and the compiled binary is preserved on the host.
//
//  2. Execution Phase: ANOTHER fresh container is started. The compiled binary is copied in.
//     Crucially, execution runs with strict limits:
//     - --user=65534:65534 (Nobody, preventing reads/writes to /etc, /bin, /root, etc.)
//     - --network=none (Total network isolation)
//     - --cap-drop=ALL & --security-opt=no-new-privileges (Blocks privilege escalation)
//     - --pids-limit & --memory (Defeats fork bombs and RAM exhaustion)
//     - Workspace permissions are locked to read/execute only (chmod 555).
//
// This guarantees that even if a malicious program executes destructive system calls, it
// only impacts a customized, unprivileged throwaway instance that is instantly destroyed.
// Note that this new implementation is slower than the older caused by container overhead.

// TODO: To reduce container overhead, I could implement a container pool.
// However, this would require a more complex cleanup mechanism to ensure that
// containers are properly cleaned up after use.

func (sm *SecurityManager) detectCapabilities() {
	sm.capabilities["docker"] = isCommandAvailable("docker")
	sm.capabilities["podman"] = isCommandAvailable("podman")
	sm.capabilities["firejail"] = isCommandAvailable("firejail")
	sm.capabilities["bubblewrap"] = isCommandAvailable("bwrap")
	sm.capabilities["systemd-run"] = isCommandAvailable("systemd-run")

	sm.capabilities["in_container"] = sm.isRunningInContainer()

	sm.capabilities["can_create_namespace"] = sm.canCreateNamespace()
}

func (sm *SecurityManager) startOrphanCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		if !sm.capabilities["docker"] && !sm.capabilities["podman"] {
			continue
		}

		runtime := "docker"
		if !sm.capabilities["docker"] && sm.capabilities["podman"] {
			runtime = "podman"
		}

		cmd := exec.Command(runtime, "ps", "-q", "-f", "name=clab-sandbox-")
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		containerIDs := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, id := range containerIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}

			inspectCmd := exec.Command(runtime, "inspect", "-f", "{{.State.StartedAt}}", id)
			inspectOut, err := inspectCmd.Output()
			if err != nil {
				continue
			}

			startedAtStr := strings.TrimSpace(string(inspectOut))
			startedAt, err := time.Parse(time.RFC3339Nano, startedAtStr)
			if err == nil {
				if time.Since(startedAt) > 5*time.Minute {
					exec.Command(runtime, "rm", "-f", id).Run()
					log.Printf("SecurityManager Cleanup: Removed orphaned container %s", id)
				}
			}
		}
	}
}

func (sm *SecurityManager) adjustConfigBasedOnCapabilities() {
	if !sm.capabilities["docker"] && !sm.capabilities["podman"] {
		if sm.capabilities["in_container"] {
			panic("CRITICAL SECURITY ERROR: Running inside a container but Docker socket is not available. Mount /var/run/docker.sock to enable sandboxing.")
		}
		panic("CRITICAL SECURITY ERROR: No suitable container runtime (Docker or Podman) found. The server cannot start securely.")
	}

	sm.config.Level = SecurityMaximum
	sm.config.UseContainer = true
}

func (sm *SecurityManager) CreateSecureCommand(ctx context.Context, executable string, args ...string) (*exec.Cmd, func(), error) {
	if sm.config.Level != SecurityMaximum {
		return nil, nil, fmt.Errorf("unknown or unsupported security level")
	}
	return sm.createContainerCommand(ctx, executable, args...)
}

func (sm *SecurityManager) createContainerCommand(ctx context.Context, executable string, args ...string) (*exec.Cmd, func(), error) {
	if !sm.capabilities["docker"] && !sm.capabilities["podman"] {
		return nil, nil, fmt.Errorf("container runtime not available")
	}

	runtime := "docker"
	if !sm.capabilities["docker"] && sm.capabilities["podman"] {
		runtime = "podman"
	}

	containerName := fmt.Sprintf("clab-sandbox-%d", time.Now().UnixNano())

	createArgs := []string{
		"run", "-d", "--name", containerName,
		"--network=none",
		fmt.Sprintf("--memory=%dm", sm.config.MaxMemoryMB),
		fmt.Sprintf("--cpus=%.2f", float64(sm.config.MaxCPUPercent)/100),
		"--security-opt=no-new-privileges",
		"--cap-drop=ALL",
		fmt.Sprintf("--pids-limit=%d", sm.config.MaxProcesses),
		sm.config.ContainerImage,
		"sleep", "86400", // Sleep for a day (will be killed by cleanup)
	}

	createCmd := exec.CommandContext(ctx, runtime, createArgs...)
	if out, err := createCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("failed to start sandbox container: %w\n%s", err, string(out))
	}

	if sm.config.WorkspaceDir != "" {
		exec.CommandContext(ctx, runtime, "exec", "-u", "root", containerName, "mkdir", "-p", sm.config.WorkspaceDir).Run()

		cpCmd := exec.CommandContext(ctx, runtime, "cp",
			sm.config.WorkspaceDir+"/.", containerName+":"+sm.config.WorkspaceDir)
		if out, err := cpCmd.CombinedOutput(); err != nil {
			exec.Command(runtime, "rm", "-f", containerName).Run()
			return nil, nil, fmt.Errorf("failed to copy workspace: %w\n%s", err, string(out))
		}

		// Determine permissions based on WorkspaceRO
		// 777 = read/write/execute (needed for compilation)
		// 555 = read/execute only (needed for running, prevents self-modification or new files)
		perms := "777"
		if sm.config.WorkspaceRO {
			perms = "555"
		}

		chmodCmd := exec.CommandContext(ctx, runtime, "exec", "-u", "root", containerName, "chmod", "-R", perms, sm.config.WorkspaceDir)
		if out, err := chmodCmd.CombinedOutput(); err != nil {
			exec.Command(runtime, "rm", "-f", containerName).Run()
			return nil, nil, fmt.Errorf("failed to set workspace permissions: %w\n%s", err, string(out))
		}
	}

	// Step 3: Return `docker exec -i -u 65534` to run the actual command as nobody
	execArgs := []string{"exec", "-i", "-u", "65534:65534"}
	if sm.config.WorkspaceDir != "" {
		// Set working directory to workspace
		execArgs = append(execArgs, "-w", sm.config.WorkspaceDir)
	}
	execArgs = append(execArgs, containerName, executable)
	execArgs = append(execArgs, args...)

	cmd := exec.CommandContext(ctx, runtime, execArgs...)

	cmd.Cancel = func() error {
		cleanupCmd := exec.Command(runtime, "rm", "-f", containerName)
		return cleanupCmd.Run()
	}

	cleanup := func() {
		if sm.config.WorkspaceDir != "" && !sm.config.WorkspaceRO {
			// Copy the workspace back to host to preserve compiled binaries
			exec.Command(runtime, "cp", "-a", containerName+":"+sm.config.WorkspaceDir+"/.", sm.config.WorkspaceDir).Run()
		}
		exec.Command(runtime, "rm", "-f", containerName).Run()
	}

	return cmd, cleanup, nil
}

func (sm *SecurityManager) SetWorkspaceDir(dir string, readOnly bool) {
	sm.config.WorkspaceDir = dir
	sm.config.WorkspaceRO = readOnly
}

func (sm *SecurityManager) ValidateExecutable(path string) error {
	if sm.config.TempDirOnly {
		tmpDir := os.TempDir()
		if !filepath.HasPrefix(path, tmpDir) {
			return fmt.Errorf("executable must be in temp directory")
		}
	}

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
