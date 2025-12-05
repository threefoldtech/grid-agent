package tfcmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/threefoldtech/grid-agent/agent/pkg/tools/builtin"
)

// Executor handles command execution
type Executor struct {
	commandExecutor *builtin.CommandExecutor
	isStreaming     bool
}

// NewExecutor creates a new non-streaming executor
func NewExecutor() *Executor {
	return &Executor{
		isStreaming: false,
	}
}

// NewExecutorWithStreaming creates a new streaming executor with shared command executor
func NewExecutorWithStreaming(commandExecutor *builtin.CommandExecutor) *Executor {
	return &Executor{
		commandExecutor: commandExecutor,
		isStreaming:     true,
	}
}

// Execute executes a command and returns its output
func (e *Executor) Execute(ctx context.Context, command []string, requestID, commandID string) (string, error) {
	// Strictly expect "tfcmd" as the first argument
	if len(command) == 0 || command[0] != "tfcmd" {
		return "", fmt.Errorf("invalid command format: expected 'tfcmd' as first argument")
	}

	// Find tfcmd executable
	tfcmdPath, err := FindTfcmd()
	if err != nil {
		return "", err
	}

	// Replace "tfcmd" with actual path
	args := make([]string, len(command))
	copy(args, command)
	args[0] = tfcmdPath

	// Execute based on streaming capability
	if e.isStreaming {
		// Use shared executor for streaming (handles expansion automatically)
		return e.commandExecutor.ExecuteCommand(ctx, args, requestID, commandID)
	}

	// Non-streaming execution - need to expand args here
	args = builtin.ExpandArguments(args)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// FindTfcmd finds the tfcmd executable (exported for use by app.go)
func FindTfcmd() (string, error) {
	// Determine executable name based on OS
	exeName := "tfcmd"
	if runtime.GOOS == "windows" {
		exeName = "tfcmd.exe"
	}

	// Get user home directory (cross-platform)
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = "" // Fallback, skip home-based paths
	}

	possiblePaths := []string{
		exeName, // In PATH
	}

	// Add standard installation paths (cross-platform)
	if userHome != "" {
		possiblePaths = append(possiblePaths,
			filepath.Join(userHome, ".local", "bin", exeName), // Linux user install
			filepath.Join(userHome, "go", "bin", exeName),     // Go bin directory
		)
	}

	possiblePaths = append(possiblePaths,
		filepath.Join("/usr", "local", "bin", exeName), // macOS/Linux system install
	)

	// Add GOPATH bin if set
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		possiblePaths = append(possiblePaths, filepath.Join(gopath, "bin", exeName))
	}

	// Add development path and current directory
	possiblePaths = append(possiblePaths,
		filepath.Join("grid-cli", "build", "bin", exeName),
		filepath.Join(".", exeName),
	)

	// Try each path
	for _, path := range possiblePaths {
		if p, err := exec.LookPath(path); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("%s not found in PATH or common locations", exeName)
}
