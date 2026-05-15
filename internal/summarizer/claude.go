package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultClaudeEndpoint = "https://api.anthropic.com/v1/messages"
	defaultClaudeModel    = "claude-sonnet-4-6"
	defaultMaxTokens      = 2048
	anthropicVersion      = "2023-06-01"
)

// ClaudeClient calls the Anthropic Messages API and satisfies LLMClient.
type ClaudeClient struct {
	apiKey     string
	model      string
	endpoint   string
	maxTokens  int
	httpClient *http.Client
}

// ClaudeOption customizes the ClaudeClient.
type ClaudeOption func(*ClaudeClient)

// WithModel overrides the model name.
func WithModel(model string) ClaudeOption {
	return func(c *ClaudeClient) { c.model = model }
}

// WithMaxTokens overrides the response token limit.
func WithMaxTokens(n int) ClaudeOption {
	return func(c *ClaudeClient) { c.maxTokens = n }
}

// WithHTTPClient injects a custom http.Client (e.g. for tests).
func WithHTTPClient(h *http.Client) ClaudeOption {
	return func(c *ClaudeClient) { c.httpClient = h }
}

// WithEndpoint overrides the API endpoint (e.g. for tests).
func WithEndpoint(url string) ClaudeOption {
	return func(c *ClaudeClient) { c.endpoint = url }
}

// NewClaudeClient builds a Claude-backed LLMClient.
func NewClaudeClient(apiKey string, opts ...ClaudeOption) *ClaudeClient {
	c := &ClaudeClient{
		apiKey:     apiKey,
		model:      defaultClaudeModel,
		endpoint:   defaultClaudeEndpoint,
		maxTokens:  defaultMaxTokens,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Complete sends the prompt as a user message and returns the assistant text.
func (c *ClaudeClient) Complete(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(claudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages:  []claudeMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call claude api: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		err := fmt.Errorf("claude api status %d: %s", resp.StatusCode, string(raw))
		if isRetryableStatus(resp.StatusCode) {
			return "", &RetryableError{Cause: err}
		}
		return "", err
	}

	var parsed claudeResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("claude api error: %s", parsed.Error.Message)
	}
	for _, blk := range parsed.Content {
		if blk.Type == "text" {
			return blk.Text, nil
		}
	}
	return "", fmt.Errorf("claude api: empty text content")
}
