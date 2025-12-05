package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/threefoldtech/grid-agent/agent/pkg/core"
	"github.com/threefoldtech/grid-agent/agent/pkg/llm"
	"github.com/threefoldtech/grid-agent/agent/pkg/tools"
	"github.com/threefoldtech/grid-agent/agent/pkg/tools/builtin"
	"github.com/threefoldtech/grid-agent/agent/pkg/workflow"
	internalConfig "github.com/threefoldtech/grid-agent/grid-agent-gui/internal/config"
	"github.com/threefoldtech/grid-agent/grid-agent-gui/internal/tfcmd"
	"github.com/threefoldtech/grid-agent/grid-cli/cmd"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	agent           *core.Agent
	settings        *Settings
	activeWorkflows map[string]context.CancelFunc
	workflowsMutex  sync.RWMutex
}

// Settings holds user configuration
type Settings struct {
	Mnemonics       string    `json:"mnemonics"`
	Network         string    `json:"network"` // mainnet, testnet, devnet
	GeminiAPIKey    string    `json:"geminiApiKey"`
	Theme           string    `json:"theme"` // light, dark
	IsConfigured    bool      `json:"isConfigured"`
	Profiles        []Profile `json:"profiles"`
	ActiveProfileID string    `json:"activeProfileID"`
}

// Profile represents a user personalization profile
type Profile struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Instructions string `json:"instructions"`
}

// Step represents a single step in the agent's workflow
type Step struct {
	ProgressText string `json:"progressText"` // Visual title with emoji: "⚡ Command Executed"
	ExportPrefix string `json:"exportPrefix"` // Export prefix: "Command:", "URL:"
	CommandID    string `json:"commandID"`    // Unique ID for command steps
	Content      string `json:"content"`      // Command or URL
	Output       string `json:"output"`       // Command output or fetched content
	Error        string `json:"error"`        // Error message if any
}

// Message represents a chat message for the frontend
type Message struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestID"` // Unique identifier for correlation
	Steps     []Step `json:"steps"`     // Workflow steps taken
	// Deprecated fields (kept for backward compatibility)
	IsCommand bool   `json:"isCommand"`
	Output    string `json:"output"`
	Error     string `json:"error"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		settings: &Settings{
			Theme: "dark",
		},
		activeWorkflows: make(map[string]context.CancelFunc),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.loadSettings()
}

// GetSettings returns the current settings
func (a *App) GetSettings() *Settings {
	return a.settings
}

// SaveSettings saves settings and initializes services
func (a *App) SaveSettings(mnemonics, network, apiKey string) error {
	a.settings.Mnemonics = mnemonics
	a.settings.Network = network
	a.settings.GeminiAPIKey = apiKey
	a.settings.IsConfigured = true

	// Save to file
	if err := a.saveSettingsToFile(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Set Gemini API key environment variable
	_ = os.Setenv("GEMINI_API_KEY", apiKey)

	// Run tfcmd login
	if err := a.runTfcmdLogin(); err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	// Initialize agent
	if err := a.initializeAgent(); err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	return nil
}

// Logout clears settings and returns to onboarding
func (a *App) Logout() error {
	// Keep theme and profiles
	currentTheme := a.settings.Theme
	currentProfiles := a.settings.Profiles

	// Clear all settings except theme and profiles
	a.settings = &Settings{
		Theme:        currentTheme,
		Profiles:     currentProfiles,
		IsConfigured: false,
	}

	// Save cleared settings
	if err := a.saveSettingsToFile(); err != nil {
		return fmt.Errorf("failed to clear settings: %w", err)
	}

	// Clear agent
	if a.agent != nil {
		_ = a.agent.Close()
		a.agent = nil
	}

	return nil
}

// GUIMessageCollector collects steps and emits events
type GUIMessageCollector struct {
	ctx         context.Context
	requestID   string
	steps       []Step
	finalAnswer string
}

func (g *GUIMessageCollector) emitEvent(step Step) {
	wailsRuntime.EventsEmit(g.ctx, "agent-progress", map[string]interface{}{
		"requestID": g.requestID,
		"step":      step,
	})
}

func (g *GUIMessageCollector) OnToolExecution(toolCallID, toolName, progressText, exportPrefix, displayArgs string, isStreaming bool) {
	var needsCommandID bool

	// Set needsCommandID based on whether it's streaming
	needsCommandID = isStreaming

	step := Step{
		ProgressText: progressText,
		ExportPrefix: exportPrefix,
		Content:      displayArgs,
		Output:       "",
	}

	if needsCommandID {
		step.CommandID = toolCallID
	}

	g.steps = append(g.steps, step)
	g.emitEvent(step)
}

func (g *GUIMessageCollector) UpdateToolOutput(toolCallID, output string, err error) {
	// Find the tool step by toolCallID and update its output
	for i := len(g.steps) - 1; i >= 0; i-- {
		step := &g.steps[i]
		// Check if this step has a CommandID that matches, or if it's a tool execution step
		if step.CommandID == toolCallID {
			if err != nil {
				step.Output = output
				step.Error = err.Error()
			} else {
				step.Output = output
			}
			g.emitEvent(*step)
			break
		}
	}
}

func (g *GUIMessageCollector) OnAnalyzing() {
	// Optional: emit analysis event
}

func (g *GUIMessageCollector) OnAnswer(answer string) error {
	// If this is the first answer, just set it
	if g.finalAnswer == "" {
		g.finalAnswer = answer
	} else {
		// If the answer is already in the final answer, don't add it again
		if !strings.Contains(g.finalAnswer, answer) {
			// Add a separator and the new answer
			separator := "\n\n" + strings.Repeat("-", 50) + "\n"
			g.finalAnswer += separator + answer
		}
	}

	// Always update the final answer with the latest content
	wailsRuntime.EventsEmit(g.ctx, "agent-answer", g.finalAnswer)

	// Add as a step if it's not already there
	step := Step{
		ProgressText: "💡 Answer",
		ExportPrefix: "",
		Content:      answer,
	}

	// Check if we already have this exact answer in steps
	found := false
	for _, s := range g.steps {
		if s.ProgressText == "💡 Answer" && s.Content == answer {
			found = true
			break
		}
	}

	if !found {
		g.steps = append(g.steps, step)
		g.emitEvent(step)
	}

	return nil
}

func (g *GUIMessageCollector) OnQuestion(question string) error {
	// Append the new question with a newline if there's existing content
	if g.finalAnswer != "" {
		g.finalAnswer += "\n\n" + question
	} else {
		g.finalAnswer = question
	}

	// Emit the updated final answer
	wailsRuntime.EventsEmit(g.ctx, "agent-question", g.finalAnswer)

	// Also emit as a step for consistency
	step := Step{
		ProgressText: "❓ Question",
		ExportPrefix: "",
		Content:      question,
	}
	g.steps = append(g.steps, step)
	g.emitEvent(step)

	return nil
}

func (g *GUIMessageCollector) OnError(message string) {
	step := Step{
		ProgressText: "❌ Error",
		ExportPrefix: "",
		Error:        message,
	}
	if len(g.steps) > 0 {
		g.steps[len(g.steps)-1].Error = message
		// Re-emit the last step with error
		g.emitEvent(g.steps[len(g.steps)-1])
	} else {
		g.steps = append(g.steps, step)
		g.emitEvent(step)
	}
}

func (g *GUIMessageCollector) OnExplanation(text string) {
	step := Step{
		ProgressText: "📊 Analysis",
		ExportPrefix: "",
		Content:      text,
	}
	g.steps = append(g.steps, step)
	g.emitEvent(step)
}

// SendMessage sends a message to the agent and returns a rich message with all steps
func (a *App) SendMessage(message string, requestID string) (*Message, error) {
	if a.agent == nil {
		return nil, fmt.Errorf("agent not initialized")
	}

	// Create cancellable context for this workflow
	ctx, cancel := context.WithCancel(a.ctx)

	// Store cancel function
	a.workflowsMutex.Lock()
	a.activeWorkflows[requestID] = cancel
	a.workflowsMutex.Unlock()

	// Clean up after workflow completes
	defer func() {
		a.workflowsMutex.Lock()
		delete(a.activeWorkflows, requestID)
		a.workflowsMutex.Unlock()
	}()

	collector := &GUIMessageCollector{ctx: a.ctx, requestID: requestID}
	processor := workflow.NewProcessor(a.agent, collector, requestID)

	err := processor.ProcessMessage(ctx, message)
	if err != nil {
		// Check if error is due to cancellation
		if ctx.Err() == context.Canceled {
			return &Message{
				Role:      "agent",
				Content:   "I have interrupted the workflow per your request.",
				Timestamp: time.Now().Format(time.RFC3339),
				RequestID: requestID,
				Steps:     collector.steps,
			}, nil
		}
		return nil, fmt.Errorf("failed to process message: %w", err)
	}

	return &Message{
		Role:      "agent",
		Content:   collector.finalAnswer,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: requestID,
		Steps:     collector.steps,
	}, nil
}

// AbortWorkflow cancels a running workflow by requestID
func (a *App) AbortWorkflow(requestID string) error {
	a.workflowsMutex.Lock()
	defer a.workflowsMutex.Unlock()

	if cancel, exists := a.activeWorkflows[requestID]; exists {
		cancel()
		delete(a.activeWorkflows, requestID)
		log.Printf("Workflow %s aborted by user", requestID)
		return nil
	}
	return fmt.Errorf("workflow %s not found or already completed", requestID)
}

// SetTheme updates the theme
func (a *App) SetTheme(theme string) error {
	a.settings.Theme = theme
	return a.saveSettingsToFile()
}

// Helper functions

func (a *App) loadSettings() {
	settingsPath := a.getSettingsPath()
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return // Settings don't exist yet
	}

	if err := json.Unmarshal(data, a.settings); err != nil {
		log.Printf("Failed to unmarshal settings: %v", err)
	}

	// If configured, initialize services
	if a.settings.IsConfigured {
		_ = os.Setenv("GEMINI_API_KEY", a.settings.GeminiAPIKey)
		if err := a.initializeAgent(); err != nil {
			log.Printf("Failed to initialize agent: %v", err)
		}
	}
}

func (a *App) saveSettingsToFile() error {
	settingsPath := a.getSettingsPath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(a.settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0600)
}

func (a *App) getSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "grid-agent", "settings.json")
}

func (a *App) runTfcmdLogin() error {
	// Use the centralized findTfcmd function from executor
	// executor := tfcmd.NewExecutor()
	// We need to access findTfcmd but it's not exported.
	// For now, let's just assume tfcmd is in path or use a simple check.
	// Or better, expose FindTfcmd in executor package?
	// Let's just use "tfcmd" and rely on PATH for now, or copy the logic.
	// Since I can't easily modify executor.go right now without another tool call,
	// I'll copy the logic here briefly or just use "tfcmd".
	// Actually, I can just use "tfcmd" and let the user ensure it's in PATH.
	// But to be safe, I'll copy the find logic.

	tfcmdPath, err := tfcmd.FindTfcmd()
	if err != nil {
		return fmt.Errorf("tfcmd not found: %w", err)
	}

	cmd := exec.Command(tfcmdPath, "login")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Use a channel to communicate errors from the goroutine back to the main function
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			_ = stdin.Close()
		}()
		// Write mnemonics and network to stdin
		if _, err := io.WriteString(stdin, a.settings.Mnemonics+"\n"); err != nil {
			errChan <- fmt.Errorf("failed to write mnemonics to stdin: %w", err)
			return
		}
		if _, err := io.WriteString(stdin, a.settings.Network+"\n"); err != nil {
			errChan <- fmt.Errorf("failed to write network to stdin: %w", err)
			return
		}
		close(errChan) // Signal that no error occurred
	}()

	output, err := cmd.CombinedOutput()

	// Check for errors from the goroutine
	if writeErr := <-errChan; writeErr != nil {
		return writeErr
	}

	if err != nil {
		return fmt.Errorf("login failed: %s", string(output))
	}

	return nil
}

func (a *App) initializeAgent() error {
	// Create streaming callback for real-time command output
	streamCallback := func(requestID, commandID, line string) {
		// Debug: Log what we're emitting
		log.Printf("[DEBUG] Emitting command-output: requestID=%s, commandID=%s, line=%s", requestID, commandID, line)

		// Emit real-time command output to GUI
		wailsRuntime.EventsEmit(a.ctx, "command-output", map[string]interface{}{
			"requestID": requestID,
			"commandID": commandID,
			"line":      line,
			"type":      "stdout",
		})
	}

	// Create a temporary registry to register tools before creating the agent
	registry := tools.NewRegistry()

	// Register tools
	// 1. Built-in tools with streaming support
	registry.Register(builtin.NewCommandToolWithStreaming(streamCallback))
	registry.Register(builtin.NewURLTool())

	// 2. Tfcmd tool with streaming
	rootCmd := cmd.GetRootCmd()
	registry.Register(tfcmd.NewTool(rootCmd, streamCallback))

	// Generate tool documentation dynamically
	toolDocs := registry.GeneratePromptSection()

	// Get list of registered tool names for dynamic parsing
	var toolNames []string
	for _, tool := range registry.List() {
		toolNames = append(toolNames, tool.Name())
	}

	// Create LLM provider with dynamic tool documentation
	providerConfig := llm.Config{
		ModelName:        "gemini-2.5-flash",
		ResponseMIMEType: "application/json",
		SystemPrompt:     strings.Replace(internalConfig.GetSystemPrompt(a.settings.Network, a.getActiveInstructions()), "{{TOOL_DESCRIPTIONS}}", toolDocs, 1),
		MaxRetries:       3,
		MaxJSONRetries:   2,
		RegisteredTools:  toolNames, // Pass tool names for dynamic parsing
	}

	provider, err := llm.NewGeminiProviderWithConfig(a.settings.GeminiAPIKey, providerConfig)
	if err != nil {
		return err
	}

	// Create agent ONCE with all configuration
	a.agent = core.NewAgent(core.Config{
		LLMProvider: provider,
		Tools:       registry, // Use the registry we already populated
	})

	return nil
}

// getActiveInstructions returns the instructions for the active profile
func (a *App) getActiveInstructions() string {
	if a.settings.ActiveProfileID == "" {
		return ""
	}
	for _, p := range a.settings.Profiles {
		if p.ID == a.settings.ActiveProfileID {
			return p.Instructions
		}
	}
	return ""
}

// AddProfile adds a new profile
func (a *App) AddProfile(name, instructions string) (*Settings, error) {
	// Generate a simple ID
	id := fmt.Sprintf("profile_%d", time.Now().UnixNano())

	profile := Profile{
		ID:           id,
		Name:         name,
		Instructions: instructions,
	}

	a.settings.Profiles = append(a.settings.Profiles, profile)

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	return a.settings, nil
}

// UpdateProfile updates an existing profile
func (a *App) UpdateProfile(id, name, instructions string) (*Settings, error) {
	for i, p := range a.settings.Profiles {
		if p.ID == id {
			a.settings.Profiles[i].Name = name
			a.settings.Profiles[i].Instructions = instructions

			if err := a.saveSettingsToFile(); err != nil {
				return nil, err
			}

			// If this is the active profile, we need to re-initialize the agent to pick up changes
			if a.settings.ActiveProfileID == id {
				// Re-initialize agent in background to avoid blocking UI
				go func() {
					if err := a.initializeAgent(); err != nil {
						log.Printf("Failed to re-initialize agent after profile update: %v", err)
					}
				}()
			}

			return a.settings, nil
		}
	}
	return nil, fmt.Errorf("profile not found")
}

// DeleteProfile deletes a profile
func (a *App) DeleteProfile(id string) (*Settings, error) {
	newProfiles := []Profile{}
	for _, p := range a.settings.Profiles {
		if p.ID != id {
			newProfiles = append(newProfiles, p)
		}
	}

	a.settings.Profiles = newProfiles

	// If active profile was deleted, deactivate it
	if a.settings.ActiveProfileID == id {
		a.settings.ActiveProfileID = ""
		// Re-initialize agent
		go func() {
			if err := a.initializeAgent(); err != nil {
				log.Printf("Failed to re-initialize agent after profile deletion: %v", err)
			}
		}()
	}

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	return a.settings, nil
}

// ActivateProfile sets the active profile
func (a *App) ActivateProfile(id string) (*Settings, error) {
	// Verify profile exists
	found := false
	for _, p := range a.settings.Profiles {
		if p.ID == id {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("profile not found")
	}

	a.settings.ActiveProfileID = id

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	// Re-initialize agent
	if err := a.initializeAgent(); err != nil {
		return nil, fmt.Errorf("failed to re-initialize agent: %w", err)
	}

	return a.settings, nil
}

// DeactivateProfile clears the active profile
func (a *App) DeactivateProfile() (*Settings, error) {
	a.settings.ActiveProfileID = ""

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	// Re-initialize agent
	if err := a.initializeAgent(); err != nil {
		return nil, fmt.Errorf("failed to re-initialize agent: %w", err)
	}

	return a.settings, nil
}

// GenerateSummary uses the agent to summarize the conversation
func (a *App) GenerateSummary(history string) (string, error) {
	if a.agent == nil {
		return "", fmt.Errorf("agent not initialized")
	}

	// Create a temporary context
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	prompt := fmt.Sprintf("Please provide a concise summary (max 2-3 sentences) of the following conversation history. Focus on the main topics and outcomes.\n\n%s", history)

	// Use the agent's LLM directly if possible, or just ask it as a normal message
	// Since we don't have direct access to LLM generate here easily without creating a workflow,
	// let's create a simple workflow or just use the agent's internal LLM if exposed.
	// The agent struct has LLMProvider but it's private in core package?
	// core.Agent has `llmProvider` which is private.
	// But we can use `ProcessMessage` with a special instruction?
	// Or better, just use the agent to "chat" but we want just the summary.

	// Let's use the agent to process this request.
	// We need a collector.
	collector := &GUIMessageCollector{ctx: a.ctx, requestID: "summary-gen"}
	processor := workflow.NewProcessor(a.agent, collector, "summary-gen")

	// We want to avoid tools for this, just pure LLM.
	// But the agent is designed to use tools.
	// However, for a summary request, it should just answer.

	err := processor.ProcessMessage(ctx, prompt)
	if err != nil {
		return "", err
	}

	return collector.finalAnswer, nil
}

// ExportChat saves the chat content to a file
func (a *App) ExportChat(content string, defaultFilename string) error {
	// Open save dialog
	filename, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Title:           "Export Conversation",
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "Markdown Files (*.md)",
				Pattern:     "*.md",
			},
		},
	})

	if err != nil {
		return err
	}

	if filename == "" {
		return nil // User cancelled
	}

	return os.WriteFile(filename, []byte(content), 0644)
}
