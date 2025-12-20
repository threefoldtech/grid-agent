package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
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
		SystemPrompt:     getOpenrouterSystemPrompt(),
		MaxRetries:       3,
		MaxJSONRetries:   2,
	}

	if modelName == "" {
		config.ModelName = "anthropic/claude-3.5-sonnet"
	}

	return NewOpenrouterProviderWithConfig(apiKey, config)
}

// getOpenrouterSystemPrompt returns an optimized system prompt for OpenRouter models
func getOpenrouterSystemPrompt() string {
	return `You are a helpful AI assistant with access to various tools and functions.

CRITICAL: You MUST respond with valid JSON only. No markdown, no explanations, no extra text.

When you need to use a tool, respond with:
{"toolName": "tool_name", "arguments": {...}, "explanation": "why"}

For answers, respond with:
{"answer": "your response here", "explanation": "context"}

For questions, respond with:
{"question": "what you need to know"}

Multiple responses can be in an array, but keep it simple and valid JSON.`
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

// parseResponse converts Openrouter response to generic Response with enhanced flexibility
func (p *OpenrouterProvider) parseResponse(text string) (*Response, error) {
	// Clean the text first
	text = strings.TrimSpace(text)

	// Try multiple parsing strategies for better robustness
	genericResp, err := p.tryParseStrategies(text)
	if err != nil {
		return nil, err
	}

	return genericResp, nil
}

// tryParseStrategies attempts multiple JSON parsing approaches
func (p *OpenrouterProvider) tryParseStrategies(text string) (*Response, error) {
	// Strategy 1: Standard array/object parsing (original approach)
	if resp, err := p.parseStandard(text); err == nil {
		return resp, nil
	}

	// Strategy 2: Extract JSON from markdown code blocks
	if resp, err := p.parseFromMarkdown(text); err == nil {
		return resp, nil
	}

	// Strategy 3: Flexible key-based parsing (look for tool/action patterns)
	if resp, err := p.parseFlexible(text); err == nil {
		return resp, nil
	}

	// Strategy 4: Try to extract any valid JSON object/array from the text
	if resp, err := p.parseExtractedJSON(text); err == nil {
		return resp, nil
	}

	// All strategies failed
	return nil, &JSONParseError{
		OriginalText: text,
		Err: fmt.Errorf("all parsing strategies failed"),
	}
}

// parseStandard - original parsing approach
func (p *OpenrouterProvider) parseStandard(text string) (*Response, error) {
	var outerResponses []LLMOuterResponse
	if err := json.Unmarshal([]byte(text), &outerResponses); err != nil {
		// Try single object
		var single LLMOuterResponse
		if err2 := json.Unmarshal([]byte(text), &single); err2 != nil {
			return nil, err2
		}
		outerResponses = []LLMOuterResponse{single}
	}

	return p.processLLMResponses(outerResponses), nil
}

// parseFromMarkdown - extract JSON from markdown code blocks
func (p *OpenrouterProvider) parseFromMarkdown(text string) (*Response, error) {
	// Look for JSON in markdown code blocks
	jsonRegex := regexp.MustCompile("```(?:json)?\\s*(\\{[\\s\\S]*?\\}|\\[[\\s\\S]*?\\])\\s*```")
	matches := jsonRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return p.parseStandard(matches[1])
	}

	// Also try without language specifier
	jsonRegex2 := regexp.MustCompile("```\\s*(\\{[\\s\\S]*?\\}|\\[[\\s\\S]*?\\])\\s*```")
	matches2 := jsonRegex2.FindStringSubmatch(text)
	if len(matches2) > 1 {
		return p.parseStandard(matches2[1])
	}

	return nil, fmt.Errorf("no JSON found in markdown")
}

// parseFlexible - look for tool/action patterns in various formats
func (p *OpenrouterProvider) parseFlexible(text string) (*Response, error) {
	resp := &Response{}

	// Look for tool call patterns
	toolPatterns := []string{
		`"toolName"\s*:\s*"([^"]+)"`,
		`"tool"\s*:\s*"([^"]+)"`,
		`"action"\s*:\s*"([^"]+)"`,
		`"function"\s*:\s*"([^"]+)"`,
	}

	for _, pattern := range toolPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 && p.registeredTools[matches[1]] {
			// Found a valid tool call
			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ToolName:    matches[1],
				Arguments:   make(map[string]interface{}),
				Explanation: "Tool call detected",
			})
			break
		}
	}

	// Look for answer patterns
	answerPatterns := []string{
		`"answer"\s*:\s*"([^"]*(?:\\.[^"]*)*)"`,
		`"response"\s*:\s*"([^"]*(?:\\.[^"]*)*)"`,
		`"result"\s*:\s*"([^"]*(?:\\.[^"]*)*)"`,
	}

	for _, pattern := range answerPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			resp.Text = matches[1]
			break
		}
	}

	// Look for question patterns
	questionPatterns := []string{
		`"question"\s*:\s*"([^"]*(?:\\.[^"]*)*)"`,
	}

	for _, pattern := range questionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			resp.Question = matches[1]
			break
		}
	}

	// If we found any structured content, return it
	if len(resp.ToolCalls) > 0 || resp.Text != "" || resp.Question != "" {
		return resp, nil
	}

	return nil, fmt.Errorf("no structured content found")
}

// parseExtractedJSON - try to find and parse any valid JSON in the text
func (p *OpenrouterProvider) parseExtractedJSON(text string) (*Response, error) {
	// Try to find JSON objects or arrays in the text
	jsonPatterns := []string{
		`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`, // Simple objects (may not handle nested)
		`\[[\s\S]*?\]`,                     // Arrays
	}

	for _, pattern := range jsonPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		for _, match := range matches {
			if resp, err := p.parseStandard(match); err == nil {
				return resp, nil
			}
		}
	}

	return nil, fmt.Errorf("no valid JSON found")
}

// processLLMResponses - common processing logic for LLMOuterResponse arrays
func (p *OpenrouterProvider) processLLMResponses(outerResponses []LLMOuterResponse) *Response {
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
	return genericResp
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
