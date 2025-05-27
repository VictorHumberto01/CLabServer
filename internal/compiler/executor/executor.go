package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type ExecutionResult struct {
	Output   string
	Error    string
	ExitCode int
	Duration time.Duration
}

type Executor struct {
	compilerPath   string
	maxMemoryUsage int64
	timeout        time.Duration
}

func NewExecutor(compilerPath string, maxMemoryUsage int64, timeout time.Duration) *Executor {
	return &Executor{
		compilerPath:   compilerPath,
		maxMemoryUsage: maxMemoryUsage,
		timeout:        timeout,
	}
}

func (e *Executor) CompileAndRun(code string, input string) (*ExecutionResult, error) {
	// Create temporary directory for compilation
	tempDir, err := os.MkdirTemp("", "clab-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write source code to file
	sourcePath := filepath.Join(tempDir, "main.c")
	if err := os.WriteFile(sourcePath, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	// Compile the code
	executablePath := filepath.Join(tempDir, "program")
	compileCmd := exec.Command(e.compilerPath, "-o", executablePath, sourcePath)
	var compileOutput bytes.Buffer
	compileCmd.Stderr = &compileOutput

	if err := compileCmd.Run(); err != nil {
		return &ExecutionResult{
			Error:    compileOutput.String(),
			ExitCode: compileCmd.ProcessState.ExitCode(),
		}, nil
	}

	// Run the program with input
	runCmd := exec.Command(executablePath)
	runCmd.Dir = tempDir

	// Set up pipes for input/output
	var stdout, stderr bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr
	runCmd.Stdin = bytes.NewReader([]byte(input))

	// Set resource limits
	runCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Set memory limit using ulimit
	if err := syscall.Setrlimit(syscall.RLIMIT_AS, &syscall.Rlimit{
		Cur: uint64(e.maxMemoryUsage),
		Max: uint64(e.maxMemoryUsage),
	}); err != nil {
		return nil, fmt.Errorf("failed to set memory limit: %w", err)
	}

	// Start the process
	startTime := time.Now()
	if err := runCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start program: %w", err)
	}

	// Create a channel to receive the result
	done := make(chan error, 1)
	go func() {
		done <- runCmd.Wait()
	}()

	// Wait for completion or timeout
	var result ExecutionResult
	select {
	case err := <-done:
		result.Duration = time.Since(startTime)
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitErr.ExitCode()
			}
			result.Error = stderr.String()
		} else {
			result.Output = stdout.String()
			result.ExitCode = 0
		}
	case <-time.After(e.timeout):
		// Kill the process group
		syscall.Kill(-runCmd.Process.Pid, syscall.SIGKILL)
		result.Error = "Execution timed out"
		result.ExitCode = -1
	}

	return &result, nil
}
