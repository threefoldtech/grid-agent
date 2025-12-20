package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// OpenrouterProvider implements the Provider interface for Openrouter
type OpenrouterProvider struct {
	client          *openai.Client
	config          Config
	registeredTools map[string]bool
	messages        []openai.ChatCompletionMessage
}

// NewOpenrouterProvider creates a new Openrouter provider
func NewOpenrouterProvider(apiKey string, modelName string) (*OpenrouterProvider, error) {
	config := Config{
		Provider:         "openrouter",
		ModelName:        modelName,
		ResponseMIMEType: "application/json",
		SystemPrompt:     "You are a helpful AI assistant.",
		MaxRetries:       3,
		MaxJSONRetries:   2,
	}

	if modelName == "" {
		config.ModelName = "anthropic/claude-3.5-sonnet"
	}

	return NewOpenrouterProviderWithConfig(apiKey, config)
}

// NewOpenrouterProviderWithConfig creates a new Openrouter provider with custom config
func NewOpenrouterProviderWithConfig(apiKey string, llmConfig Config) (*OpenrouterProvider, error) {
	if llmConfig.ModelName == "" {
		llmConfig.ModelName = "anthropic/claude-3.5-sonnet"
	}

	openaiConfig := openai.DefaultConfig(apiKey)
	openaiConfig.BaseURL = "https://openrouter.ai/api/v1"
	client := openai.NewClientWithConfig(openaiConfig)

	provider := &OpenrouterProvider{
		client:          client,
		config:          llmConfig,
		registeredTools: make(map[string]bool),
		messages:        []openai.ChatCompletionMessage{},
	}

	// Populate registered tools map
	for _, toolName := range llmConfig.RegisteredTools {
		provider.registeredTools[toolName] = true
	}

	// Initialize with system message
	provider.messages = append(provider.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: llmConfig.SystemPrompt,
	})

	return provider, nil
}

// SendMessage sends a message to Openrouter
func (p *OpenrouterProvider) SendMessage(ctx context.Context, message string) (*Response, error) {
	// Add user message
	p.messages = append(p.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	var lastErr error
	var currentMessages []openai.ChatCompletionMessage

	// Retry loop for JSON parsing errors
	for attempt := 0; attempt <= p.config.MaxJSONRetries; attempt++ {
		currentMessages = make([]openai.ChatCompletionMessage, len(p.messages))
		copy(currentMessages, p.messages)

		// Create chat completion
		req := openai.ChatCompletionRequest{
			Model:    p.config.ModelName,
			Messages: currentMessages,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		}

		resp, err := p.client.CreateChatCompletion(ctx, req)
		if err != nil {
			return nil, p.friendlyError(err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("empty response from Openrouter")
		}

		content := resp.Choices[0].Message.Content
		parsedResp, err := p.parseResponse(content)
		if err == nil {
			// Add assistant message to history
			p.messages = append(p.messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: content,
			})
			return parsedResp, nil
		}

		// Check if it's a JSON parse error
		if jsonErr, ok := err.(*JSONParseError); ok {
			lastErr = err
			if attempt < p.config.MaxJSONRetries {
				log.Printf("JSON parse error (attempt %d/%d): %v. Retrying with feedback...", attempt+1, p.config.MaxJSONRetries+1, err)

				// Construct feedback message for the LLM
				feedbackMsg := fmt.Sprintf("I received an error parsing your last response as JSON. Error: %v\n\nYour previous response was:\n%s\n\nPlease correct the format and respond ONLY with valid JSON matching the schema.", jsonErr.Err, jsonErr.OriginalText)
				currentMessages = append(currentMessages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: feedbackMsg,
				})
				continue
			}
		} else {
			// legitimate other error
			return nil, err
		}
	}

	// If we exhausted retries, fallback to returning the text from the last error if available
	if jsonErr, ok := lastErr.(*JSONParseError); ok {
		log.Printf("Exhausted JSON retries. Falling back to raw text.")
		return &Response{Text: jsonErr.OriginalText}, nil
	}

	return nil, lastErr
}

// GetHistory returns the conversation history
func (p *OpenrouterProvider) GetHistory() []Message {
	var history []Message
	for _, msg := range p.messages {
		role := "user"
		if msg.Role == openai.ChatMessageRoleAssistant {
			role = "assistant"
		} else if msg.Role == openai.ChatMessageRoleSystem {
			role = "system"
		}

		history = append(history, Message{
			Role:    role,
			Content: msg.Content,
		})
	}
	return history
}

// GetRawHistory returns the raw history for session carry-over
func (p *OpenrouterProvider) GetRawHistory() any {
	return p.messages
}

// AppendSystemNotice adds a system notice to the history
func (p *OpenrouterProvider) AppendSystemNotice(message string) error {
	// For Openrouter, we can add a system message
	p.messages = append(p.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	})
	return nil
}

// Close closes the provider and releases resources
func (p *OpenrouterProvider) Close() error {
	// OpenAI client doesn't require explicit closing
	return nil
}

// parseResponse converts Openrouter response to generic Response
func (p *OpenrouterProvider) parseResponse(text string) (*Response, error) {
	// Try to parse JSON using the defined struct
	var outerResponses []LLMOuterResponse
	if err := json.Unmarshal([]byte(text), &outerResponses); err != nil {
		// Try single object
		var single LLMOuterResponse
		if err2 := json.Unmarshal([]byte(text), &single); err2 != nil {
			// Return special error to trigger retry loop in SendMessage
			return nil, &JSONParseError{
				OriginalText: text,
				Err:          err2,
			}
		}
		outerResponses = []LLMOuterResponse{single}
	}

	// Handle all responses in the list for multiple tool calls
	genericResp := &Response{}
	var finalAnswer strings.Builder

	for _, r := range outerResponses {
		if r.Answer != "" {
			finalAnswer.WriteString(r.Answer + "\n")
		}
		if r.Question != "" {
			genericResp.Question = r.Question
		}
		// Look for unified tool call format: toolName and arguments
		if r.ToolName != "" {
			// Check if this tool is registered
			if !p.registeredTools[r.ToolName] {
				continue // Tool not registered, skip
			}

			// Get arguments (LLM should always provide them)
			arguments := r.Arguments

			// Create tool call with its explanation
			toolCall := ToolCall{
				ToolName:    r.ToolName,
				Arguments:   arguments,
				Explanation: r.Explanation,
			}

			genericResp.ToolCalls = append(genericResp.ToolCalls, toolCall)
		}
	}

	genericResp.Text = strings.TrimSpace(finalAnswer.String())
	return genericResp, nil
}

// friendlyError converts Openrouter API errors to user-friendly messages
func (p *OpenrouterProvider) friendlyError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for common error patterns
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "invalid_api_key") {
		return fmt.Errorf("🔑 Invalid or expired API key. Please check your Openrouter API key in Settings.")
	}

	// Quota/rate limit errors
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "quota") {
		return fmt.Errorf("⏳ API quota exceeded. Please check your Openrouter account limits.")
	}

	// Network/timeout errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") || strings.Contains(errStr, "context canceled") {
		return fmt.Errorf("⏱️ Request timed out. Please try again.")
	}

	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") {
		return fmt.Errorf("🌐 Network error. Please check your internet connection.")
	}

	// Return original error if no match
	return err
}
