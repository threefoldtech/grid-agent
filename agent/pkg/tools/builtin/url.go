package builtin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/threefoldtech/grid-agent/agent/pkg/tools"
)

// URLTool fetches content from a URL
type URLTool struct{}

func NewURLTool() *URLTool {
	return &URLTool{}
}

func (t *URLTool) Name() string {
	return "fetch_url"
}

func (t *URLTool) SkipUpdateToolOutput() bool {
	return true // URL content is displayed differently, not as tool output
}

func (t *URLTool) Description() tools.ToolDescriptor {
	return tools.ToolDescriptor{
		Name:        "fetch_url",
		Description: "Fetch content from a URL (HTTP GET request)",
		CallFormat: `{
  "toolName": "fetch_url",
  "arguments": "https://example.com/api/endpoint",
  "explanation": "explain why you need to fetch this URL",
  "risk_level": "low|medium|high"
}`,
		Instructions: `Use this tool to fetch content from URLs. The arguments should be a string containing the URL to fetch. 

RISK ASSESSMENT:
- "low": Fetching public documentation, README files, or non-sensitive data.
- "medium": Fetching internal APIs or potentially sensitive configuration files.
- "high": Fetching executable content (e.g., install scripts) or communicating with critical control planes.
Default to "low" for most documentation reading tasks.`,
		Examples: []string{
			`{
  "toolName": "fetch_url",
  "arguments": "https://api.github.com/repos/threefoldtech/tf-images/contents/tfgrid3"
}`,
			`{
  "toolName": "fetch_url",
  "arguments": "https://raw.githubusercontent.com/threefoldtech/tf-images/development/tfgrid3/wordpress/README.md"
}`,
		},
		ProgressText: "🌐 URL Fetched",
		ExportPrefix: "URL:",
	}
}

func (t *URLTool) FormatDisplayArgs(args any) string {
	// Only support direct string
	if url, ok := args.(string); ok {
		return url
	}
	return "invalid URL format"
}

func (t *URLTool) Execute(ctx context.Context, args any) (map[string]any, error) {
	var url string
	var ok bool

	// Only support direct string
	if url, ok = args.(string); !ok {
		return nil, fmt.Errorf("URL argument must be a string")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Use a custom client with parameters to avoid hanging
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"content": string(body),
		"status":  resp.Status,
	}, nil
}
