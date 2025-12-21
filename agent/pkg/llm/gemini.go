package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/genai"
)

// LLMOuterResponse defines the top-level structure of LLM responses
type LLMOuterResponse struct {
	Answer      string      `json:"answer"`
	Explanation string      `json:"explanation"`
	Question    string      `json:"question"`
	ToolName    string      `json:"toolName"`
	Arguments   interface{} `json:"arguments"`
	RiskLevel   string      `json:"risk_level"`
}

// GeminiProvider implements the Provider interface for Gemini
type GeminiProvider struct {
	client          *genai.Client
	chat            *genai.Chat
	config          Config
	registeredTools map[string]bool
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string, modelName string) (*GeminiProvider, error) {
	// Default config
	config := Config{
		ModelName:        modelName,
		ResponseMIMEType: "application/json",
		SystemPrompt:     "You are a helpful AI assistant.",
		MaxRetries:       3,
		MaxJSONRetries:   2,
	}

	if modelName == "" {
		config.ModelName = "gemini-3-flash-preview"
	}

	return NewGeminiProviderWithConfig(apiKey, config)
}

// NewGeminiProviderWithConfig creates a new Gemini provider with custom config
func NewGeminiProviderWithConfig(apiKey string, config Config) (*GeminiProvider, error) {
	ctx := context.Background()
	// Initialize the client with the new SDK
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	if config.ModelName == "" {
		config.ModelName = "gemini-3-flash-preview"
	}

	provider := &GeminiProvider{
		client:          client,
		config:          config,
		registeredTools: make(map[string]bool),
	}

	// Populate registered tools map
	for _, toolName := range config.RegisteredTools {
		provider.registeredTools[toolName] = true
	}

	chat, err := provider.startChatSession(ctx)
	if err != nil {
		return nil, err
	}
	provider.chat = chat

	return provider, nil
}

// JSONParseError represents an error when parsing LLM response as JSON
type JSONParseError struct {
	OriginalText string
	Err          error
}

func (e *JSONParseError) Error() string {
	return fmt.Sprintf("failed to parse JSON response: %v", e.Err)
}

// friendlyError converts Gemini API errors to user-friendly messages
func friendlyError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Quota exceeded (429)
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "RESOURCE_EXHAUSTED") || strings.Contains(errStr, "quota") {
		return fmt.Errorf("⏳ API quota exceeded. The free tier has limited requests. Please wait a minute or upgrade your API plan at https://ai.google.dev")
	}

	// Invalid API key (401/403)
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "403") || strings.Contains(errStr, "PERMISSION_DENIED") || strings.Contains(errStr, "UNAUTHENTICATED") {
		return fmt.Errorf("🔑 Invalid or expired API key. Please check your Gemini API key in Settings.")
	}

	// Network/timeout errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") || strings.Contains(errStr, "context canceled") {
		return fmt.Errorf("⏱️ Request timed out. Please try again.")
	}

	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") || strings.Contains(errStr, "dial tcp") {
		return fmt.Errorf("🌐 Network error. Please check your internet connection.")
	}

	// Return original error if no match
	return err
}

// SendMessage sends a message to Gemini
func (p *GeminiProvider) SendMessage(ctx context.Context, message string) (*Response, error) {
	if p.chat == nil {
		return nil, fmt.Errorf("chat session is not initialized")
	}

	var lastErr error
	var currentMessage = message

	// Retry loop for JSON parsing errors
	for attempt := 0; attempt <= p.config.MaxJSONRetries; attempt++ {
		// Send message using the new SDK
		resp, err := p.chat.SendMessage(ctx, genai.Part{Text: currentMessage})
		if err != nil {
			return nil, friendlyError(err)
		}

		parsedResp, err := p.parseResponse(resp)
		if err == nil {
			return parsedResp, nil
		}

		// Check if it's a JSON parse error
		if jsonErr, ok := err.(*JSONParseError); ok {
			lastErr = err
			if attempt < p.config.MaxJSONRetries {
				log.Printf("JSON parse error (attempt %d/%d): %v. Retrying with feedback...", attempt+1, p.config.MaxJSONRetries+1, err)

				// Construct feedback message for the LLM
				currentMessage = fmt.Sprintf("I received an error parsing your last response as JSON. Error: %v\n\nYour previous response was:\n%s\n\nPlease correct the format and respond ONLY with valid JSON matching the schema.", jsonErr.Err, jsonErr.OriginalText)
				continue
			}
		} else {
			// legitimate other error (e.g. empty response)
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
func (p *GeminiProvider) GetHistory() []Message {
	if p.chat == nil {
		return nil
	}
	var history []Message
	// Get curated history (valid turns)
	sdkHistory := p.chat.History(true)

	for _, content := range sdkHistory {
		role := "user"
		if content.Role == "model" {
			role = "assistant"
		}

		var text string
		for _, part := range content.Parts {
			// In new SDK, Part has a Text field directly (if strictly text)
			// or we need to check other fields.
			if part != nil {
				text += part.Text
			}
		}

		history = append(history, Message{
			Role:    role,
			Content: text,
		})
	}
	return history
}

// GetRawHistory returns the raw history for session carry-over
func (p *GeminiProvider) GetRawHistory() any {
	if p.chat == nil {
		return nil
	}
	return p.chat.History(true)
}

// AppendSystemNotice adds a system notice to the chat history
// Note: This modifies the history that will be passed to the next session
// The actual modification happens in the caller (app.go) before creating new provider
func (p *GeminiProvider) AppendSystemNotice(message string) error {
	// This is a no-op for now - the actual history modification
	// happens in app.go before creating the new provider
	// We keep this method to satisfy the interface
	return nil
}

func (p *GeminiProvider) startChatSession(ctx context.Context) (*genai.Chat, error) {
	var chat *genai.Chat
	var err error

	// Automatically append the mandatory JSON format instructions
	fullPrompt := p.config.SystemPrompt + "\n" + JSONFormatInstructions
	// log fullPrompt
	log.Println("Full Prompt:", fullPrompt)

	// Configure generation options
	genConfig := &genai.GenerateContentConfig{
		ResponseMIMEType: p.config.ResponseMIMEType,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: fullPrompt}},
		},
	}

	// Prepare history
	var history []*genai.Content
	if p.config.History != nil {
		if h, ok := p.config.History.([]*genai.Content); ok {
			history = h
		}
	}

	for i := 0; i < p.config.MaxRetries; i++ {
		// Create a new chat session
		// The new SDK uses client.Chats.Create
		chat, err = p.client.Chats.Create(ctx, p.config.ModelName, genConfig, history)
		if err == nil && chat != nil {
			return chat, nil
		}

		if i < p.config.MaxRetries-1 {
			waitTime := time.Duration(1<<uint(i)) * time.Second
			log.Printf("Failed to start chat session, retrying in %v... (attempt %d/%d) Error: %v", waitTime, i+1, p.config.MaxRetries, err)
			time.Sleep(waitTime)
		}
	}
	return nil, fmt.Errorf("failed to start chat session with Gemini after %d attempts: %w", p.config.MaxRetries, err)
}

// Close closes the Gemini client
func (p *GeminiProvider) Close() error {
	// The new client doesn't seem to have a Close method in the examples/API we saw?
	// But api_client usually has one.
	// Looking at example_test.go, client usage doesn't show Close().
	// However, it likely has http connection pools.
	// We can leave it empty or checking if there is a Close method.
	// We'll trust standard Go patterns; if it has it, we call it.
	// If the compiler complains, we'll remove it.
	// I'll assume it doesn't need explicit closing or it's not exposed on the high level client yet?
	// Actually, most Google Cloud clients DO satisfy io.Closer.
	// Let's try to verify via types.go or just comment it out to be safe if compilation fails.
	// The previous code had p.client.Close().
	// Let's assume the new one doesn't for now or use reflection/interface check? No, that's runtime.
	// I'll comment it out with a note to verify.
	// return p.client.Close()
	return nil
}

// parseResponse converts Gemini response to generic Response
func (p *GeminiProvider) parseResponse(resp *genai.GenerateContentResponse) (*Response, error) {
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	var text string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part != nil {
			text += part.Text
		}
	}

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
				RiskLevel:   r.RiskLevel,
			}

			genericResp.ToolCalls = append(genericResp.ToolCalls, toolCall)
		}
	}

	genericResp.Text = strings.TrimSpace(finalAnswer.String())
	return genericResp, nil
}
