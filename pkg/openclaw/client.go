package openclaw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Config holds OpenClaw gateway configuration
type Config struct {
	Port  int    `json:"port"`
	Token string `json:"-"` // From auth.token
}

// Client is an OpenClaw gateway HTTP client
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseChunk represents a streamed response chunk
type ResponseChunk struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Done    bool   `json:"done,omitempty"`
	Error   string `json:"error,omitempty"`
}

// LoadConfig loads OpenClaw config from ~/.openclaw/openclaw.json
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".openclaw", "openclaw.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read openclaw config: %w", err)
	}

	var raw struct {
		Gateway struct {
			Port int `json:"port"`
			Auth struct {
				Token string `json:"token"`
			} `json:"auth"`
		} `json:"gateway"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("could not parse openclaw config: %w", err)
	}

	return &Config{
		Port:  raw.Gateway.Port,
		Token: raw.Gateway.Auth.Token,
	}, nil
}

// NewClient creates a new OpenClaw client
func NewClient(config *Config) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://localhost:%d", config.Port),
		token:   config.Token,
		http:    &http.Client{},
	}
}

// NewClientFromConfig loads config and creates a client
func NewClientFromConfig() (*Client, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return NewClient(config), nil
}

// SendMessage sends a message to OpenClaw
// Note: Full bidirectional chat requires OpenResponses API or WebSocket
// For now, we show status and guide user to full TUI
func (c *Client) SendMessage(message string, projectContext string, onChunk func(chunk string)) error {
	// Get current session status
	result, err := c.InvokeTool("session_status", map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("gateway error: %w", err)
	}

	// Extract status text
	if details, ok := result["details"].(map[string]interface{}); ok {
		if statusText, ok := details["statusText"].(string); ok {
			// Parse out key info
			lines := strings.Split(statusText, "\n")
			for _, line := range lines {
				if strings.Contains(line, "Model:") || strings.Contains(line, "Context:") {
					onChunk(line + " | Press 'c' for full chat")
					return nil
				}
			}
		}
	}
	
	// Fallback
	onChunk("🦞 Connected to OpenClaw. Press 'c' to launch full TUI in project context.")
	return nil
}

// SendMessageSync sends a message and returns the full response
func (c *Client) SendMessageSync(message string, projectContext string) (string, error) {
	var response strings.Builder
	err := c.SendMessage(message, projectContext, func(chunk string) {
		response.WriteString(chunk)
	})
	if err != nil {
		return "", err
	}
	return response.String(), nil
}

// InvokeTool invokes a tool via the /tools/invoke endpoint
func (c *Client) InvokeTool(tool string, args map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"tool": tool,
		"args": args,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/tools/invoke", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Ok     bool                   `json:"ok"`
		Result map[string]interface{} `json:"result,omitempty"`
		Error  struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Ok {
		return nil, fmt.Errorf("%s: %s", result.Error.Type, result.Error.Message)
	}

	return result.Result, nil
}

// Ping checks if the gateway is reachable
func (c *Client) Ping() error {
	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("gateway unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gateway returned status %d", resp.StatusCode)
	}

	return nil
}
