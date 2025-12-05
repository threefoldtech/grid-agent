package builtin

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/threefoldtech/grid-agent/agent/pkg/tools"
)

type contextKey string

const (
	// RequestIDKey is the context key for the request ID
	RequestIDKey contextKey = "requestID"
	// CommandIDKey is the context key for the command ID
	CommandIDKey contextKey = "commandID"
)

// CommandTool executes shell commands with optional real-time streaming
type CommandTool struct {
	streamCallback StreamCallback
	executor       *CommandExecutor
	isStreaming    bool
}

func NewCommandTool() *CommandTool {
	return &CommandTool{
		isStreaming: false,
	}
}

func NewCommandToolWithStreaming(callback StreamCallback) *CommandTool {
	return &CommandTool{
		streamCallback: callback,
		executor:       NewCommandExecutor(callback),
		isStreaming:    true,
	}
}

func (t *CommandTool) HasStreamingCallback() bool {
	return t.isStreaming
}

func (t *CommandTool) SkipUpdateToolOutput() bool {
	return false
}

func (t *CommandTool) Name() string {
	return "command"
}

func (t *CommandTool) Description() tools.ToolDescriptor {
	return tools.ToolDescriptor{
		Name:        "command",
		Description: "Execute a shell command (read-only and safe commands recommended)",
		CallFormat: `{
  "toolName": "command",
  "arguments": ["ls", "-la", "/tmp"],
  "explanation": "explain why you need to use this command"
}`,
		Instructions: `Use this tool to execute system commands. The arguments should be a list of strings representing the command and its arguments. You can execute ANY system command including file operations, SSH, kubectl, and other CLI tools. Always use appropriate commands for the operating system.`,
		Examples: []string{
			`{
  "toolName": "command",
  "arguments": ["cat", "~/.ssh/id_rsa.pub"],
}`,
			`{
  "toolName": "command",
  "arguments": ["ls", "-la", "/tmp"]
}`,
		},
		ProgressText: "⚡ Command Executed",
		ExportPrefix: "Command:",
	}
}

func (t *CommandTool) FormatDisplayArgs(args any) string {
	// Only support array format
	if argsArray, ok := args.([]interface{}); ok {
		var parts []string
		for _, v := range argsArray {
			if s, ok := v.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " ")
	}
	return "invalid arguments format"
}

func (t *CommandTool) Execute(ctx context.Context, args any) (map[string]any, error) {
	var parts []string

	// Only support array of strings
	if argsArray, ok := args.([]interface{}); ok {
		parts = make([]string, len(argsArray))
		for i, v := range argsArray {
			if s, ok := v.(string); ok {
				parts[i] = s
			} else {
				return nil, fmt.Errorf("command argument at index %d must be a string", i)
			}
		}
	} else {
		return nil, fmt.Errorf("arguments must be a list of strings")
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	var output string
	var err error

	// Use shared executor if streaming, otherwise fallback to direct execution
	if t.isStreaming {
		requestID, _ := ctx.Value(RequestIDKey).(string)
		commandID, _ := ctx.Value(CommandIDKey).(string)
		output, err = t.executor.ExecuteCommand(ctx, parts, requestID, commandID)
	} else {
		// Fallback for non-streaming case - need to expand args here too
		parts = ExpandArguments(parts)
		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		if cmd.SysProcAttr == nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		}
		var outputBytes []byte
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)
	}

	result := map[string]any{
		"output": output,
	}

	if err != nil {
		// Check if command failed to start (not found, permission denied, etc.)
		if _, ok := err.(*exec.ExitError); !ok {
			// Command didn't even start - this is a critical error
			result["error"] = err.Error()
			return result, fmt.Errorf("command execution failed: %w", err)
		}
		// Command ran but exited with non-zero code
		// Include exit code in error message for better LLM understanding
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["error"] = fmt.Sprintf("Command failed with exit code %d: %s",
				exitErr.ExitCode(), output)
		} else {
			result["error"] = err.Error()
		}
	}

	return result, nil
}
