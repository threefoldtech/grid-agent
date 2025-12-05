package workflow

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/threefoldtech/grid-agent/agent/pkg/core"
	"github.com/threefoldtech/grid-agent/agent/pkg/llm"
	"github.com/threefoldtech/grid-agent/agent/pkg/tools/builtin"
)

var commandCounter uint64

// ResponseHandler defines the interface for handling streaming responses
type ResponseHandler interface {
	OnToolExecution(toolCallID, toolName, progressText, exportPrefix, displayArgs string, isStreaming bool)
	OnAnalyzing()
	OnAnswer(answer string) error
	OnQuestion(question string) error
	OnExplanation(text string)
	OnError(message string)
	UpdateToolOutput(toolCallID, output string, err error)
}

// Processor handles the response processing loop
type Processor struct {
	agent     *core.Agent
	handler   ResponseHandler
	requestID string
}

// NewProcessor creates a new response processor
func NewProcessor(agent *core.Agent, handler ResponseHandler, requestID string) *Processor {
	return &Processor{
		agent:     agent,
		handler:   handler,
		requestID: requestID,
	}
}

// ProcessMessage sends a message and processes the response stream
func (p *Processor) ProcessMessage(ctx context.Context, message string) error {
	resp, err := p.agent.SendMessage(ctx, message)
	if err != nil {
		return err
	}

	return p.processResponseLoop(ctx, resp)
}

func (p *Processor) processResponseLoop(ctx context.Context, resp *llm.Response) error {
	// Loop limit to prevent infinite loops
	const maxSteps = 20 // Increased from 10 to allow for more complex workflows
	steps := 0

	for {
		if steps >= maxSteps {
			return fmt.Errorf("maximum number of steps (%d) reached. This might indicate a loop in the agent's responses", maxSteps)
		}
		steps++

		// 1. Handle Question
		if resp.Question != "" {
			if err := p.handler.OnQuestion(resp.Question); err != nil {
				return err
			}
			// If we have a question, we pause for user input.
			// Even if there are tools, we usually want to ask first.
			// But if the model asks AND calls tools, we might want to run tools?
			// For safety, let's assume Question means "Stop and Ask".
			return nil
		}

		// 2. Handle Text/Answer/Explanation
		if resp.Text != "" {
			if len(resp.ToolCalls) > 0 {
				// Intermediate explanation - treat as analysis step
				p.handler.OnExplanation(resp.Text)
			} else {
				// Final answer - accumulate
				if err := p.handler.OnAnswer(resp.Text); err != nil {
					return err
				}
			}
		}

		// Handle Tool Calls
		if len(resp.ToolCalls) > 0 {
			for _, toolCall := range resp.ToolCalls {
				var toolCallID string
				var isStreaming bool

				// Execute tool
				tool, ok := p.agent.GetTool(toolCall.ToolName)
				if !ok {
					// Tool not found - report error
					feedback := fmt.Sprintf("Error: Tool '%s' not found.", toolCall.ToolName)
					var sendErr error
					resp, sendErr = p.agent.SendMessage(ctx, feedback)
					if sendErr != nil {
						return sendErr
					}
					continue
				}

				// Generate toolCallID for tracking all tool executions
				toolCallID = fmt.Sprintf("tool_%d", atomic.AddUint64(&commandCounter, 1))

				// Get tool metadata from the tool's descriptor
				toolDesc := tool.Description()

				// Check if tool supports streaming (dynamic detection)
				if streamingTool, hasMethod := tool.(interface{ HasStreamingCallback() bool }); hasMethod {
					isStreaming = streamingTool.HasStreamingCallback()
				}

				// Format display arguments using the tool's own method
				displayArgs := tool.FormatDisplayArgs(toolCall.Arguments)

				// Notify UI about tool execution start using unified handler
				p.handler.OnToolExecution(toolCallID, toolCall.ToolName, toolDesc.ProgressText, toolDesc.ExportPrefix, displayArgs, isStreaming)

				// Set up context with IDs
				ctxWithID := context.WithValue(ctx, builtin.RequestIDKey, p.requestID)
				if isStreaming {
					ctxWithID = context.WithValue(ctxWithID, builtin.CommandIDKey, toolCallID)
				}
				output, err := tool.Execute(ctxWithID, toolCall.Arguments)

				// Update GUI with actual tool output (unless tool opts out)
				if !tool.SkipUpdateToolOutput() {
					var outputStr string
					if out, ok := output["output"]; ok {
						outputStr = fmt.Sprintf("%v", out)
					}
					p.handler.UpdateToolOutput(toolCallID, outputStr, err)
				}

				// Format feedback for LLM
				var feedback string
				if err != nil {
					feedback = fmt.Sprintf("Tool '%s' failed: %v", toolCall.ToolName, err)
					p.handler.OnError(feedback)
				} else {
					// Assume output has "output" or "content" key
					if out, ok := output["output"]; ok {
						feedback = fmt.Sprintf("Tool output:\n%v", out)
					} else if content, ok := output["content"]; ok {
						feedback = fmt.Sprintf("Tool output:\n%v", content)
					} else {
						feedback = fmt.Sprintf("Tool executed successfully. Result: %v", output)
					}

					if isStreaming {
						p.handler.OnAnalyzing()
					}
				}

				// Send feedback to LLM
				resp, err = p.agent.SendMessage(ctx, feedback)
				if err != nil {
					return err
				}
			}
			// Continue loop with new response
			continue
		}

		// If no tools (and text was already handled), we are done
		return nil
	}
}
