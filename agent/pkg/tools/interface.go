package tools

import (
	"context"
	"fmt"
	"strings"
)

// ToolDescriptor describes a tool's capabilities and usage
type ToolDescriptor struct {
	Name         string   // Tool name (e.g., "command", "fetch_url", "tfcmd")
	Description  string   // Brief description of what the tool does
	CallFormat   string   // JSON format example for calling this tool
	Instructions string   // Detailed instructions on when and how to use the tool
	Examples     []string // Example JSON calls
	ProgressText string   // UI progress text with emoji: "⚡ Command Executed", "🌐 URL Fetched"
	ExportPrefix string   // Export prefix: "Command:", "URL:" (for markdown export)
	DisplayArgs  bool     // Whether to display arguments in the UI
}

// Tool defines the interface for agent tools
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns structured metadata about the tool
	Description() ToolDescriptor

	// FormatDisplayArgs formats the tool arguments for UI display
	// Returns the formatted string to show in the UI for this tool call
	FormatDisplayArgs(args any) string

	// Execute runs the tool with given arguments
	Execute(ctx context.Context, args any) (map[string]any, error)

	// SkipUpdateToolOutput returns true if this tool should not call UpdateToolOutput
	// Streaming tools typically skip this since output is handled progressively
	SkipUpdateToolOutput() bool
}

// Registry manages available tools
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GeneratePromptSection generates formatted documentation for all registered tools
func (r *Registry) GeneratePromptSection() string {
	if len(r.tools) == 0 {
		return "No tools available."
	}

	var builder strings.Builder
	builder.WriteString("Available tools:\n\n")

	for _, tool := range r.tools {
		desc := tool.Description()

		builder.WriteString(fmt.Sprintf("### %s\n", desc.Name))
		builder.WriteString(fmt.Sprintf("%s\n\n", desc.Description))

		if desc.CallFormat != "" {
			builder.WriteString("**Call Format:**\n```json\n")
			builder.WriteString(desc.CallFormat)
			builder.WriteString("\n```\n\n")
		}

		if desc.Instructions != "" {
			builder.WriteString("**Instructions:**\n")
			builder.WriteString(desc.Instructions)
			builder.WriteString("\n\n")
		}

		if len(desc.Examples) > 0 {
			builder.WriteString("**Examples:**\n")
			for _, example := range desc.Examples {
				builder.WriteString("```json\n")
				builder.WriteString(example)
				builder.WriteString("\n```\n")
			}
			builder.WriteString("\n")
		}

		builder.WriteString("---\n\n")
	}

	return builder.String()
}
