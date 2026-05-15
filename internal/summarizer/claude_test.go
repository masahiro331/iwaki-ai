package summarizer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClaudeClient_Complete_Success(t *testing.T) {
	var gotReq claudeRequest
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotReq); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"summary result"}]}`))
	}))
	defer srv.Close()

	c := NewClaudeClient("test-key",
		WithEndpoint(srv.URL),
		WithModel("test-model"),
		WithMaxTokens(123),
	)

	got, err := c.Complete(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if got != "summary result" {
		t.Errorf("Complete() = %q, want %q", got, "summary result")
	}

	if gotHeaders.Get("x-api-key") != "test-key" {
		t.Errorf("missing api key header: %v", gotHeaders.Get("x-api-key"))
	}
	if gotHeaders.Get("anthropic-version") == "" {
		t.Errorf("missing anthropic-version header")
	}
	if gotReq.Model != "test-model" {
		t.Errorf("model = %q, want test-model", gotReq.Model)
	}
	if gotReq.MaxTokens != 123 {
		t.Errorf("max_tokens = %d, want 123", gotReq.MaxTokens)
	}
	if len(gotReq.Messages) != 1 || gotReq.Messages[0].Content != "hello" {
		t.Errorf("messages = %+v, want single user 'hello'", gotReq.Messages)
	}
}

func TestClaudeClient_Complete_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"type":"server_error","message":"boom"}}`))
	}))
	defer srv.Close()

	c := NewClaudeClient("k", WithEndpoint(srv.URL))

	_, err := c.Complete(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error on 500 response, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention status, got: %v", err)
	}
}

func TestClaudeClient_Complete_EmptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"content":[]}`))
	}))
	defer srv.Close()

	c := NewClaudeClient("k", WithEndpoint(srv.URL))

	_, err := c.Complete(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error on empty content, got nil")
	}
}
