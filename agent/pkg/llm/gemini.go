package llm

import (
	"context"
	"fmt"
	"log"
	"time"

	"encoding/json"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface for Gemini
type GeminiProvider struct {
	client          *genai.Client
	model           *genai.GenerativeModel
	cs              *genai.ChatSession
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
		config.ModelName = "gemini-2.5-flash"
	}

	return NewGeminiProviderWithConfig(apiKey, config)
}

// NewGeminiProviderWithConfig creates a new Gemini provider with custom config
func NewGeminiProviderWithConfig(apiKey string, config Config) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	if config.ModelName == "" {
		config.ModelName = "gemini-2.5-flash"
	}

	model := client.GenerativeModel(config.ModelName)
	model.ResponseMIMEType = config.ResponseMIMEType
	// Automatically append the mandatory JSON format instructions
	fullPrompt := config.SystemPrompt + "\n" + JSONFormatInstructions
	model.SystemInstruction = genai.NewUserContent(genai.Text(fullPrompt))

	provider := &GeminiProvider{
		client:          client,
		model:           model,
		config:          config,
		registeredTools: make(map[string]bool),
	}

	// Populate registered tools map
	for _, toolName := range config.RegisteredTools {
		provider.registeredTools[toolName] = true
	}

	cs, err := provider.startChatSession()
	if err != nil {
		return nil, err
	}
	provider.cs = cs

	return provider, nil
}

func (p *GeminiProvider) startChatSession() (*genai.ChatSession, error) {
	var session *genai.ChatSession
	for i := 0; i < p.config.MaxRetries; i++ {
		session = p.model.StartChat()
		if session != nil {
			return session, nil
		}
		if i < p.config.MaxRetries-1 {
			waitTime := time.Duration(1<<uint(i)) * time.Second
			log.Printf("Failed to start chat session, retrying in %v... (attempt %d/%d)", waitTime, i+1, p.config.MaxRetries)
			time.Sleep(waitTime)
		}
	}
	return nil, fmt.Errorf("failed to start chat session with Gemini after %d attempts", p.config.MaxRetries)
}

// SendMessage sends a message to Gemini
func (p *GeminiProvider) SendMessage(ctx context.Context, message string) (*Response, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in SendMessage: %v. Restarting session...", r)
			var oldHistory []*genai.Content
			if p.cs != nil {
				oldHistory = p.cs.History
			}

			newCs, err := p.startChatSession()
			if err == nil {
				if len(oldHistory) > 0 {
					newCs.History = oldHistory
				}
				p.cs = newCs
				log.Printf("Session restarted successfully and history restored")
			} else {
				log.Printf("Failed to restart session after panic: %v", err)
			}
		}
	}()

	resp, err := p.cs.SendMessage(ctx, genai.Text(message))
	if err != nil {
		return nil, err
	}

	return p.parseResponse(resp)
}

// GetHistory returns the conversation history
func (p *GeminiProvider) GetHistory() []Message {
	var history []Message
	for _, content := range p.cs.History {
		role := "user"
		if content.Role == "model" {
			role = "assistant"
		}

		var text string
		for _, part := range content.Parts {
			if t, ok := part.(genai.Text); ok {
				text += string(t)
			}
		}

		history = append(history, Message{
			Role:    role,
			Content: text,
		})
	}
	return history
}

// Close closes the Gemini client
func (p *GeminiProvider) Close() error {
	return p.client.Close()
}

// parseResponse converts Gemini response to generic Response
func (p *GeminiProvider) parseResponse(resp *genai.GenerateContentResponse) (*Response, error) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}
	var text string
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}

	// Sanitize JSON string: escape control characters that might be unescaped
	text = sanitizeJSON(text)

	// Try to parse JSON
	var responses []map[string]interface{}
	if err := json.Unmarshal([]byte(text), &responses); err != nil {
		// Try single object
		var single map[string]interface{}
		if err2 := json.Unmarshal([]byte(text), &single); err2 != nil {
			// Not JSON, return raw text
			log.Printf("[DEBUG] Failed to parse JSON response. Error: %v. Raw text (first 200 chars): %s", err2, text[:min(200, len(text))])
			return &Response{Text: text}, nil
		}
		responses = []map[string]interface{}{single}
	}

	// For now, we only handle the first response in the list for the generic interface
	genericResp := &Response{}
	var finalAnswer strings.Builder

	for _, r := range responses {
		if answer, ok := r["answer"].(string); ok && answer != "" {
			finalAnswer.WriteString(answer + "\n")
		}
		if question, ok := r["question"].(string); ok && question != "" {
			genericResp.Question = question
		}
		// Look for unified tool call format: toolName and arguments
		if toolNameValue, exists := r["toolName"]; exists && toolNameValue != nil {
			toolName, ok := toolNameValue.(string)
			if !ok {
				continue // Invalid toolName type
			}

			// Check if this tool is registered
			if !p.registeredTools[toolName] {
				continue // Tool not registered, skip
			}

			// Get arguments
			var arguments any
			if argsValue, exists := r["arguments"]; exists && argsValue != nil {
				arguments = argsValue
			} else {
				// Default empty arguments based on tool type
				if toolName == "tfcmd" {
					arguments = []interface{}{}
				} else {
					arguments = ""
				}
			}

			// Create tool call
			toolCall := ToolCall{
				ToolName:  toolName,
				Arguments: arguments,
			}

			genericResp.ToolCalls = append(genericResp.ToolCalls, toolCall)
		}

		// Handle explanation for Answer/Question/Tool
		// We always add explanation to text so it appears as an analysis step
		if explanation, ok := r["explanation"].(string); ok && explanation != "" {
			finalAnswer.WriteString("\n" + explanation + "\n")
		}
	}

	genericResp.Text = strings.TrimSpace(finalAnswer.String())
	return genericResp, nil
}

// sanitizeJSON attempts to fix common JSON formatting issues from LLMs
func sanitizeJSON(s string) string {
	// If the string contains unescaped newlines within quotes, we need to escape them
	// This is a naive implementation but handles the most common case
	var result strings.Builder
	inString := false
	escaped := false

	for _, r := range s {
		if escaped {
			result.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			result.WriteRune(r)
			continue
		}

		if r == '"' {
			inString = !inString
		}

		if inString {
			if r == '\n' {
				result.WriteString("\\n")
				continue
			}
			if r == '\t' {
				result.WriteString("\\t")
				continue
			}
			if r == '\r' {
				continue // skip carriage returns
			}
		}

		result.WriteRune(r)
	}

	return result.String()
}
