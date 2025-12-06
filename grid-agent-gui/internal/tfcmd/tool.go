package tfcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/grid-agent/agent/pkg/tools"
	"github.com/threefoldtech/grid-agent/agent/pkg/tools/builtin"
)

// Tool implements the agent.Tool interface for tfcmd
type Tool struct {
	schema   *CommandSchema
	executor *Executor
}

// NewTool creates a new tfcmd tool
func NewTool(rootCmd *cobra.Command, callback builtin.StreamCallback) *Tool {
	var executor *Executor
	if callback != nil {
		// Create streaming executor with shared command executor
		commandExecutor := builtin.NewCommandExecutor(callback)
		executor = NewExecutorWithStreaming(commandExecutor)
	} else {
		// Create non-streaming executor
		executor = NewExecutor()
	}

	return &Tool{
		schema:   GenerateSchema(rootCmd),
		executor: executor,
	}
}

const ToolName = "tfcmd"

func (t *Tool) Name() string {
	return ToolName
}

func (t *Tool) HasStreamingCallback() bool {
	return t.executor.isStreaming
}

func (t *Tool) SkipUpdateToolOutput() bool {
	return false
}

func (t *Tool) Description() tools.ToolDescriptor {
	// Generate schema JSON for instructions
	schemaJSON, _ := json.MarshalIndent(t.schema, "", "  ")

	return tools.ToolDescriptor{
		Name:        ToolName,
		Description: "Execute ThreeFold Grid CLI commands for managing VMs, Kubernetes, gateways, and contracts",
		CallFormat: `{
  "toolName": "tfcmd",
  "arguments": ["tfcmd", "deploy", "vm", "--name", "myvm"],
  "explanation": "explain why you need to execute this command"
}`,
		Instructions: fmt.Sprintf(`Use this tool to interact with the ThreeFold Grid.
- Deploy and manage VMs, Kubernetes clusters, gateways, and ZDBs
- Query and cancel contracts
- All tfcmd commands are available

VALIDATION RULES:
- Check that ALL required flags are provided (look for "required": true in the schema)
- Check that ALL required positional arguments are provided (look for "args" in the schema)
- Evaluate the 'Mutually Exclusive Flags' and 'Flag Groups' sections for the target subcommand.
- If any required flags or arguments are missing, ask the user first
- resources names must contain only letters and numbers
- Always remember that deploy kubernetes|vm --ssh flag takes the file path not the file content

FLAG GROUPS (must be set together):
- deploy vm: if --flist is provided, --entrypoint MUST also be provided

MUTUALLY EXCLUSIVE FLAGS (only ONE can be set):
- deploy vm: --node OR --farm (not both)
- deploy kubernetes: --master-node OR --master-farm (not both)
- deploy kubernetes: --workers-nodes OR --workers-farm (not both)
- deploy gateway name: --node OR --farm (not both)
- deploy zdb: --node OR --farm (not both)

BOOLEAN FLAG SYNTAX:
- To enable: --flag or --flag=true
- To disable: --flag=false (MUST use = sign)
- WRONG: --mycelium false
- CORRECT: --mycelium=false

SSH KEYS IN ENV VARS:
- Some flists require SSH key as ENV VAR (e.g., SSH_KEY, pub_key)
- For ENV VAR: pass the actual key CONTENT, not the file path

CANCEL/DELETE:
- By name: tfcmd cancel <deployment-name> (preferred for single deployments)
- By contract: tfcmd cancel contracts <contract-id>
- Cancel all: tfcmd cancel contracts -a (REQUIRES explicit user confirmation!)

OTHER RELEVANT INFORMATION:
- Don't Ask user for his twin ID, instead look it up as it it usually can be seen in the output of tfcmd, such "get" or "contracts", etc.
- Always use --disable-sentry with any deploy commands unless user specified otherwise

%s`, schemaJSON),
		Examples: []string{
			`{
  "toolName": "tfcmd",
  "arguments": ["tfcmd", "list"]
}`,
			`{
  "toolName": "tfcmd",
  "arguments": ["tfcmd", "deploy", "vm", "--name", "myvm", "--ssh", "~/.ssh/id_rsa.pub"]
}`,
		},
		ProgressText: "⚡ Command Executed",
		ExportPrefix: "Command:",
	}
}

func (t *Tool) FormatDisplayArgs(args any) string {
	// Only support array format
	if cmdList, ok := args.([]interface{}); ok {
		var parts []string
		for _, arg := range cmdList {
			parts = append(parts, fmt.Sprint(arg))
		}
		return strings.Join(parts, " ")
	}
	return "invalid tfcmd arguments"
}

func (t *Tool) Execute(ctx context.Context, args any) (map[string]any, error) {
	// Parse args to build command
	// Expecting args as list of strings
	var cmdArgs []string

	// Only support array format
	if cmdList, ok := args.([]interface{}); ok {
		for _, arg := range cmdList {
			cmdArgs = append(cmdArgs, fmt.Sprint(arg))
		}
	} else {
		return nil, fmt.Errorf("arguments must be a list of strings")
	}

	if len(cmdArgs) == 0 {
		return nil, fmt.Errorf("missing or invalid arguments")
	}

	// Extract requestID and commandID from context
	requestID, _ := ctx.Value(builtin.RequestIDKey).(string)
	commandID, _ := ctx.Value(builtin.CommandIDKey).(string)

	output, err := t.executor.Execute(ctx, cmdArgs, requestID, commandID)

	result := map[string]any{
		"output": output,
	}

	if err != nil {
		// Check if command failed to start (not found, permission denied, etc.)
		result["error"] = err.Error()
		// Return error for critical failures (similar to command tool)
		// This ensures the LLM knows the command failed
	}

	return result, err
}

// Ensure Tool implements tools.Tool
var _ tools.Tool = (*Tool)(nil)
