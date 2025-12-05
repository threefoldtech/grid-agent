package builtin

import (
	"context"
	"fmt"
	"os/exec"
)

// CommandExecutor provides shared command execution with streaming support
type CommandExecutor struct {
	streamCallback StreamCallback
	expandArgs     bool // whether to expand arguments (tilde/globs)
}

// NewCommandExecutor creates a new command executor with streaming support
func NewCommandExecutor(callback StreamCallback) *CommandExecutor {
	return &CommandExecutor{
		streamCallback: callback,
		expandArgs:     true, // expand by default
	}
}

// NewCommandExecutorWithOptions creates a command executor with custom options
func NewCommandExecutorWithOptions(callback StreamCallback, expandArgs bool) *CommandExecutor {
	return &CommandExecutor{
		streamCallback: callback,
		expandArgs:     expandArgs,
	}
}

// ExecuteCommand executes a command with streaming support
// args[0] should be the command, args[1:] should be arguments
// Arguments are automatically expanded (tilde and globs) unless disabled
func (e *CommandExecutor) ExecuteCommand(ctx context.Context, args []string, requestID, commandID string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("no command provided")
	}

	// Expand arguments (tilde and globs) before execution if enabled
	if e.expandArgs {
		args = ExpandArguments(args)
	}

	// Create command with context for proper cancellation support
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	// Execute with streaming if callback is available
	if e.streamCallback != nil {
		return ExecuteWithStreaming(cmd, requestID, commandID, e.streamCallback)
	}

	// Fallback to non-streaming execution
	output, err := cmd.CombinedOutput()
	return string(output), err
}
