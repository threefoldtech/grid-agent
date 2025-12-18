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
		Instructions: fmt.Sprintf(`Execute ThreeFold Grid CLI commands using the provided schema.

TFCMD SCHEMA:
%s

VALIDATION CHECKLIST:
- You MUST use alphanumeric only for deployment and project names
- You MUST provide all required flags (check "required": true in schema)
- You MUST provide required positional args (check "args" in schema)  
- You MUST satisfy flag groups (see FLAG GROUPS below)
- You MUST respect mutually exclusive flags (see EXCLUSIONS below)
- You MUST use file path for --ssh flag, not key content

CRITICAL FLAG RULES:

VALIDATION RULES:
- Check that ALL required flags are provided (look for "required": true in the schema)
- Check that ALL required positional arguments are provided (look for "args" in the schema)
- Evaluate the 'Mutually Exclusive Flags' and 'Flag Groups' sections for the target subcommand.
- If any required flags or arguments are missing, ask the user first
- resources names must contain only letters and numbers
- Always remember that deploy kubernetes|vm --ssh flag takes the file path not the file content

FLAG GROUPS (must be set together):
- deploy vm: --flist + --entrypoint (both required if either is used)

EXCLUSIONS (only ONE can be set):
- deploy vm: --node OR --farm (never both)
- deploy kubernetes: --master-node OR --master-farm
- deploy kubernetes: --workers-nodes OR --workers-farm  
- deploy gateway name: --node OR --farm
- deploy gateway fqdn: --node OR --farm
- deploy zdb: --node OR --farm

BOOLEAN FLAGS:
- Enable: --flag or --flag=true
- Disable: --flag=false (equals sign REQUIRED)
- WRONG: --mycelium false
- RIGHT: --mycelium=false

SSH & ENVIRONMENT:
- --ssh flag: file path (e.g., ~/.ssh/id_rsa.pub)
- ENV vars like SSH_KEY: actual key content, not file path


PRO TIPS:
- Find twin ID from "get contracts" output - don't ask user

DOMAIN PLANNING PATTERN (breaks circular dependency):
1. List available gateways and pick node:
   tfcmd list gateways name --farm 1

2. Plan your full domain (gateway domain + your subdomain):
   If gateway shows: "gent01.dev.grid.tf"
   Your planned domain: "myapp.gent01.dev.grid.tf"

3. Deploy VM with planned domain in env vars:
   tfcmd deploy vm --name api --project-name myapp --env DOMAIN=myapp.gent01.dev.grid.tf

4. Get VM IP after deployment:
   tfcmd get vm api --project-name myapp
   Note the private IP: e.g., 10.20.2.2

5. Deploy gateway to planned node with backend IP:
   tfcmd deploy gateway name --name myapp --node <gateway-node-id> --backends http://10.20.2.2:8080 --project-name myapp

This allows configuring app with domain before gateway exists!

`, schemaJSON),
		Examples: []string{
			`{
  "toolName": "tfcmd",
  "arguments": ["tfcmd", "list", "gateways", "name"]
}`,
			`{
  "toolName": "tfcmd", 
  "arguments": ["tfcmd", "deploy", "vm", "--name", "myvm", "--ssh", "~/.ssh/id_rsa.pub", "--cpu", "2", "--memory", "4"]
}`,
			`{
  "toolName": "tfcmd",
  "arguments": ["tfcmd", "deploy", "gateway", "name", "--name", "myapp", "--node", "11", "--backends", "http://10.20.2.2:8080", "--network", "mynetwork", "--project-name", "myapp"]
}`,
		},
		ProgressText: "Command Executed",
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
