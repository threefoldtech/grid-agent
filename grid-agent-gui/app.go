package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	"google.golang.org/genai"
)

// Version is the current application version (set via ldflags during build)
var Version = "dev"

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
	Mnemonics           string    `json:"mnemonics"`
	Network             string    `json:"network"` // mainnet, testnet, devnet
	GeminiAPIKey        string    `json:"geminiApiKey"`
	Model               string    `json:"model"`
	Theme               string    `json:"theme"` // light, dark
	IsConfigured        bool      `json:"isConfigured"`
	Profiles            []Profile `json:"profiles"`
	ActiveProfileID     string    `json:"activeProfileID"`
	EnableExportSummary bool      `json:"enableExportSummary"` // Generate AI summary on export (uses tokens)
}

// Profile represents a user personalization profile
type Profile struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Instructions string `json:"instructions"`
}

// Step types
const (
	StepTypeTool     = "tool"
	StepTypeAnalysis = "analysis"
	StepTypeQuestion = "question"
	StepTypeAnswer   = "answer"
	StepTypeError    = "error"
)

// Step represents a single step in the agent's workflow
type Step struct {
	Type         string `json:"type"`         // Step type: "tool", "analysis", "question", "answer", "error"
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

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return Version
}

// UpdateInfo contains version update information
type UpdateInfo struct {
	UpdateAvailable bool   `json:"updateAvailable"`
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	ReleaseURL      string `json:"releaseURL"`
}

// CheckForUpdates checks GitHub for a newer version
func (a *App) CheckForUpdates() UpdateInfo {
	result := UpdateInfo{
		CurrentVersion: Version,
	}

	// Create HTTP client with redirect policy to capture final URL
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
		Timeout: 5 * time.Second,
	}

	// Check the latest release URL (it redirects to the actual version)
	resp, err := client.Get("https://github.com/threefoldtech/grid-agent/releases/latest")
	if err != nil {
		log.Printf("Failed to check for updates: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Get the redirect location which contains the version
	location := resp.Header.Get("Location")
	if location == "" {
		return result
	}

	// Extract version from URL like: https://github.com/threefoldtech/grid-agent/releases/tag/v0.2.1
	parts := strings.Split(location, "/tag/")
	if len(parts) != 2 {
		return result
	}

	latestVersion := parts[1]
	result.LatestVersion = latestVersion
	result.ReleaseURL = location

	// Compare versions (simple string comparison, assumes semantic versioning)
	if latestVersion != Version && latestVersion > Version {
		result.UpdateAvailable = true
	}

	return result
}

// VersionMismatchInfo contains version comparison information
type VersionMismatchInfo struct {
	HasMismatch  bool   `json:"hasMismatch"`
	AppVersion   string `json:"appVersion"`
	TfcmdVersion string `json:"tfcmdVersion"`
	ReleaseURL   string `json:"releaseURL"`
}

// CheckTfcmdVersion checks if the installed tfcmd version matches the app version
func (a *App) CheckTfcmdVersion() VersionMismatchInfo {
	result := VersionMismatchInfo{
		AppVersion: Version,
		ReleaseURL: "https://github.com/threefoldtech/grid-agent/releases/tag/" + Version,
	}

	// Find tfcmd using the same function used for command execution
	tfcmdPath, err := tfcmd.FindTfcmd()
	if err != nil {
		log.Printf("Failed to find tfcmd: %v", err)
		// If tfcmd is not found, we can't check version - treat as mismatch
		result.HasMismatch = true
		result.TfcmdVersion = "not found"
		return result
	}

	// Run tfcmd version to get the version string
	cmd := exec.Command(tfcmdPath, "version")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get tfcmd version: %v", err)
		result.HasMismatch = true
		result.TfcmdVersion = "unknown"
		return result
	}

	// Parse the version from the first line of output
	// tfcmd version outputs: "v0.3.0\n<commit-hash>"
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		result.HasMismatch = true
		result.TfcmdVersion = "unknown"
		return result
	}

	tfcmdVersion := strings.TrimSpace(lines[0])
	result.TfcmdVersion = tfcmdVersion

	// Compare versions - they should match exactly
	if tfcmdVersion != Version {
		result.HasMismatch = true
	}

	return result
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
	if err := a.initializeAgent(AgentInitOptions{}); err != nil {
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
		Type:         StepTypeTool,
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
		Type:         StepTypeAnswer,
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
		Type:         StepTypeQuestion,
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
		Type:         StepTypeError,
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
		Type:         StepTypeAnalysis,
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
		return nil, fmt.Errorf("Sorry, something went wrong while processing your message.\n %w", err)
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
		if err := a.initializeAgent(AgentInitOptions{}); err != nil {
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

// AgentInitOptions holds options for agent initialization
type AgentInitOptions struct {
	ContextChangeMsg string // Message to add to history as system notice
	ClearHistory     bool   // Whether to clear history instead of carrying it over
}

func (a *App) initializeAgent(opts AgentInitOptions) error {
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

	// Capture history from existing agent if available (for session carry-over)
	var history any
	if !opts.ClearHistory && a.agent != nil {
		if provider := a.agent.GetProvider(); provider != nil {
			// Get raw history using interface method
			if rawHistory := provider.GetRawHistory(); rawHistory != nil {
				// Make a copy to avoid modifying previous session
				if h, ok := rawHistory.([]*genai.Content); ok {
					historySlice := make([]*genai.Content, len(h))
					copy(historySlice, h)

					// If we have a context change message, append it to history
					if opts.ContextChangeMsg != "" {
						// Append User message with the notice
						userContent := &genai.Content{
							Role: "user",
							Parts: []*genai.Part{
								{Text: fmt.Sprintf("\n\n[SYSTEM NOTICE: %s]\n\n", opts.ContextChangeMsg)},
							},
						}
						// Append Model acknowledgement to keep the turn structure valid (User -> Model)
						modelContent := &genai.Content{
							Role: "model",
							Parts: []*genai.Part{
								{Text: "[System state update acknowledged.]"},
							},
						}
						historySlice = append(historySlice, userContent, modelContent)
						log.Printf("Carrying over chat history with context marker: %s", opts.ContextChangeMsg)
					} else {
						log.Printf("Carrying over chat history without context marker")
					}

					history = historySlice
				}
			}
		}
	} else if opts.ClearHistory {
		log.Printf("Clearing chat history as requested")
	}

	// Determine model
	modelName := "gemini-3-flash-preview"
	if a.settings.Model != "" {
		modelName = a.settings.Model
	}

	// Create LLM provider with dynamic tool documentation
	providerConfig := llm.Config{
		ModelName:        modelName,
		ResponseMIMEType: "application/json",
		SystemPrompt:     strings.Replace(internalConfig.GetSystemPrompt(a.settings.Network, a.getActiveInstructions()), "{{TOOL_DESCRIPTIONS}}", toolDocs, 1),
		MaxRetries:       3,
		MaxJSONRetries:   2,
		RegisteredTools:  toolNames, // Pass tool names for dynamic parsing
		History:          history,   // Pass previous history
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

// getProfileByID returns a pointer to the profile and its index, or nil and -1 if not found
func (a *App) getProfileByID(id string) (*Profile, int) {
	for i := range a.settings.Profiles {
		if a.settings.Profiles[i].ID == id {
			return &a.settings.Profiles[i], i
		}
	}
	return nil, -1
}

// getActiveInstructions returns the instructions for the active profile
func (a *App) getActiveInstructions() string {
	if a.settings.ActiveProfileID == "" {
		return ""
	}
	if p, _ := a.getProfileByID(a.settings.ActiveProfileID); p != nil {
		return p.Instructions
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
	_, idx := a.getProfileByID(id)
	if idx == -1 {
		return nil, fmt.Errorf("profile not found")
	}

	a.settings.Profiles[idx].Name = name
	a.settings.Profiles[idx].Instructions = instructions

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	// If this is the active profile, we need to re-initialize the agent to pick up changes
	if a.settings.ActiveProfileID == id {
		// Re-initialize agent in background to avoid blocking UI
		go func() {
			if err := a.initializeAgent(AgentInitOptions{
				ContextChangeMsg: "Context Update: The user has modified the profile instructions. You must adhere to the new system prompt and ignore previous persona instructions if they conflict.",
				ClearHistory:     false,
			}); err != nil {
				log.Printf("Failed to re-initialize agent after profile update: %v", err)
			}
		}()
	}

	return a.settings, nil
}

// DeleteProfile deletes a profile
func (a *App) DeleteProfile(id string) (*Settings, error) {
	_, idx := a.getProfileByID(id)
	if idx != -1 {
		// Remove the profile at idx
		a.settings.Profiles = append(a.settings.Profiles[:idx], a.settings.Profiles[idx+1:]...)
	}

	// If active profile was deleted, deactivate it
	if a.settings.ActiveProfileID == id {
		a.settings.ActiveProfileID = ""
		// Re-initialize agent
		go func() {
			if err := a.initializeAgent(AgentInitOptions{
				ContextChangeMsg: "Profile deleted. Reverting to default instructions.",
				ClearHistory:     false,
			}); err != nil {
				log.Printf("Failed to re-initialize agent after profile deletion: %v", err)
			}
		}()
	}

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	return a.settings, nil
}

// UpdateAdvancedSettings updates the API key, model, and export options
func (a *App) UpdateAdvancedSettings(apiKey, model string, enableExportSummary bool) (*Settings, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	a.settings.GeminiAPIKey = apiKey
	a.settings.Model = model
	a.settings.EnableExportSummary = enableExportSummary

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	// Re-initialize agent to apply new settings
	if err := a.initializeAgent(AgentInitOptions{}); err != nil {
		log.Printf("Failed to re-initialize agent after settings update: %v", err)
		// We return the settings anyway as they are saved
	}

	return a.settings, nil
}

// UpdateGridSettings updates grid configuration (mnemonics and network)
// This clears chat history since changing network/mnemonics means different twin/contracts
func (a *App) UpdateGridSettings(mnemonics string, network string) (*Settings, error) {
	a.settings.Mnemonics = mnemonics
	a.settings.Network = network

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	// Run tfcmd login with new credentials
	if err := a.runTfcmdLogin(); err != nil {
		return nil, fmt.Errorf("failed to login with new grid credentials: %w", err)
	}

	// Re-initialize agent with CLEARED history (new network = new context)
	if err := a.initializeAgent(AgentInitOptions{
		ContextChangeMsg: fmt.Sprintf("Network changed to %s. Chat history cleared for new grid context.", network),
		ClearHistory:     true, // Clear history for grid config changes
	}); err != nil {
		log.Printf("Failed to re-initialize agent after grid settings update: %v", err)
		// We return the settings anyway as they are saved
	}

	return a.settings, nil
}

// ActivateProfile sets the active profile
func (a *App) ActivateProfile(id string) (*Settings, error) {
	// Verify profile exists
	if _, idx := a.getProfileByID(id); idx == -1 {
		return nil, fmt.Errorf("profile not found")
	}

	a.settings.ActiveProfileID = id

	if err := a.saveSettingsToFile(); err != nil {
		return nil, err
	}

	// Re-initialize agent
	if err := a.initializeAgent(AgentInitOptions{
		ContextChangeMsg: "New profile activated with updated instructions.",
		ClearHistory:     false,
	}); err != nil {
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
	if err := a.initializeAgent(AgentInitOptions{
		ContextChangeMsg: "Profile deactivated. Reverting to default instructions.",
		ClearHistory:     false,
	}); err != nil {
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

	// Check for active persona
	activePersona := ""
	if a.settings.ActiveProfileID != "" {
		if p, _ := a.getProfileByID(a.settings.ActiveProfileID); p != nil {
			activePersona = p.Name
		}
	}

	// Prepend metadata
	metadata := fmt.Sprintf("# Chat Export\n\n**App Version:** %s\n**Active Persona:** %s\n\n---\n\n", Version, activePersona)
	fullContent := metadata + content

	return os.WriteFile(filename, []byte(fullContent), 0644)
}
