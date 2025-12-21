package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	baseURL       string
	config        Config
	registeredTools map[string]bool
	messages      []OllamaMessage
}

// OllamaMessage represents a message in Ollama format
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Message OllamaMessage `json:"message"`
	Done    bool          `json:"done"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string, modelName string) (*OllamaProvider, error) {
	config := Config{
		Provider:         "ollama",
		ModelName:        modelName,
		ResponseMIMEType: "application/json",
		SystemPrompt:     getOllamaSystemPrompt(),
		MaxRetries:       3,
		MaxJSONRetries:   2,
	}

	if modelName == "" {
		config.ModelName = "llama3.1:8b"
	}

	return NewOllamaProviderWithConfig(baseURL, config)
}

// getOllamaSystemPrompt returns an optimized system prompt for Ollama models
func getOllamaSystemPrompt() string {
	return `You are a helpful AI assistant with access to various tools and functions.

CRITICAL: You MUST respond with valid JSON only. No markdown, no explanations, no extra text.

When you need to use a tool, respond with:
{"toolName": "tool_name", "arguments": {...}, "explanation": "why"}

For answers, respond with:
{"answer": "your response here", "explanation": "context"}

For questions, respond with:
{"question": "what you need to know"}

Keep responses simple and valid JSON.`
}

// NewOllamaProviderWithConfig creates a new Ollama provider with custom config
func NewOllamaProviderWithConfig(baseURL string, llmConfig Config) (*OllamaProvider, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	if llmConfig.ModelName == "" {
		llmConfig.ModelName = "llama3.1:8b"
	}

	provider := &OllamaProvider{
		baseURL:          baseURL,
		config:           llmConfig,
		registeredTools:  make(map[string]bool),
		messages:         []OllamaMessage{},
	}

	// Populate registered tools map
	for _, toolName := range llmConfig.RegisteredTools {
		provider.registeredTools[toolName] = true
	}

	// Initialize with system message
	provider.messages = append(provider.messages, OllamaMessage{
		Role:    "system",
		Content: llmConfig.SystemPrompt,
	})

	return provider, nil
}

// SendMessage sends a message to Ollama
func (p *OllamaProvider) SendMessage(ctx context.Context, message string) (*Response, error) {
	// Add user message
	p.messages = append(p.messages, OllamaMessage{
		Role:    "user",
		Content: message,
	})

	var lastErr error
	var currentMessages []OllamaMessage

	// Retry loop for JSON parsing errors
	for attempt := 0; attempt <= p.config.MaxJSONRetries; attempt++ {
		currentMessages = make([]OllamaMessage, len(p.messages))
		copy(currentMessages, p.messages)

		// Create Ollama request
		req := OllamaRequest{
			Model:    p.config.ModelName,
			Messages: currentMessages,
			Stream:   false,
		}

		resp, err := p.makeRequest(ctx, req)
		if err != nil {
			return nil, p.friendlyError(err)
		}

		content := resp.Message.Content
		parsedResp, err := p.parseResponse(content)
		if err == nil {
			// Add assistant message to history
			p.messages = append(p.messages, resp.Message)
			return parsedResp, nil
		}

		// Check if it's a JSON parse error
		if jsonErr, ok := err.(*JSONParseError); ok {
			lastErr = err
			if attempt < p.config.MaxJSONRetries {
				log.Printf("JSON parse error (attempt %d/%d): %v. Retrying with feedback...", attempt+1, p.config.MaxJSONRetries+1, err)

				// Construct feedback message for the LLM
				feedbackMsg := fmt.Sprintf("I received an error parsing your last response as JSON. Error: %v\n\nYour previous response was:\n%s\n\nPlease correct the format and respond ONLY with valid JSON matching the schema.", jsonErr.Err, jsonErr.OriginalText)
				currentMessages = append(currentMessages, OllamaMessage{
					Role:    "user",
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

// makeRequest sends a request to Ollama API
func (p *OllamaProvider) makeRequest(ctx context.Context, req OllamaRequest) (*OllamaResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ollamaResp, nil
}

// GetHistory returns the conversation history
func (p *OllamaProvider) GetHistory() []Message {
	var history []Message
	for _, msg := range p.messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "assistant"
		} else if msg.Role == "system" {
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
func (p *OllamaProvider) GetRawHistory() any {
	return p.messages
}

// AppendSystemNotice adds a system notice to the history
func (p *OllamaProvider) AppendSystemNotice(message string) error {
	p.messages = append(p.messages, OllamaMessage{
		Role:    "system",
		Content: message,
	})
	return nil
}

// Close closes the provider and releases resources
func (p *OllamaProvider) Close() error {
	// Ollama doesn't require explicit closing
	return nil
}

// parseResponse converts Ollama response to generic Response with enhanced flexibility
func (p *OllamaProvider) parseResponse(text string) (*Response, error) {
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
func (p *OllamaProvider) tryParseStrategies(text string) (*Response, error) {
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
func (p *OllamaProvider) parseStandard(text string) (*Response, error) {
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
func (p *OllamaProvider) parseFromMarkdown(text string) (*Response, error) {
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
func (p *OllamaProvider) parseFlexible(text string) (*Response, error) {
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
func (p *OllamaProvider) parseExtractedJSON(text string) (*Response, error) {
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
func (p *OllamaProvider) processLLMResponses(outerResponses []LLMOuterResponse) *Response {
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

// friendlyError converts Ollama API errors to user-friendly messages
func (p *OllamaProvider) friendlyError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for common error patterns
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "dial tcp") {
		return fmt.Errorf("🚫 Cannot connect to Ollama. Please ensure Ollama is running with 'ollama serve'")
	}

	if strings.Contains(errStr, "model not found") || strings.Contains(errStr, "model not available") {
		return fmt.Errorf("📦 Model '%s' not found. Please run 'ollama pull %s' to download it", p.config.ModelName, p.config.ModelName)
	}

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") || strings.Contains(errStr, "context canceled") {
		return fmt.Errorf("⏱️ Request timed out. Ollama may be busy or the model may be too large for your system")
	}

	// Return original error if no match
	return err
}
