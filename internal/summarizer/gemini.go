package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	defaultGeminiModel   = "gemini-1.5-flash"
)

// GeminiClient calls Google AI Studio's generateContent API and satisfies LLMClient.
type GeminiClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// GeminiOption customizes the GeminiClient.
type GeminiOption func(*GeminiClient)

// WithGeminiModel overrides the model name (e.g. "gemini-1.5-flash-8b").
func WithGeminiModel(model string) GeminiOption {
	return func(c *GeminiClient) { c.model = model }
}

// WithGeminiBaseURL overrides the API base URL (mainly for tests).
func WithGeminiBaseURL(u string) GeminiOption {
	return func(c *GeminiClient) { c.baseURL = u }
}

// WithGeminiHTTPClient injects a custom http.Client.
func WithGeminiHTTPClient(h *http.Client) GeminiOption {
	return func(c *GeminiClient) { c.httpClient = h }
}

// NewGeminiClient builds a Gemini-backed LLMClient.
func NewGeminiClient(apiKey string, opts ...GeminiOption) *GeminiClient {
	c := &GeminiClient{
		apiKey:     apiKey,
		model:      defaultGeminiModel,
		baseURL:    defaultGeminiBaseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Complete sends the prompt as a single user content and returns the first candidate text.
func (c *GeminiClient) Complete(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		c.baseURL, c.model, url.QueryEscape(c.apiKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call gemini api: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("gemini api status %d: %s", resp.StatusCode, string(raw))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("gemini api error: %s", parsed.Error.Message)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini api: empty candidates")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
