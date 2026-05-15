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

func TestGeminiClient_Complete_Success(t *testing.T) {
	var gotReq geminiRequest
	var gotPath string
	var gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotReq); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		_, _ = w.Write([]byte(`{
			"candidates": [
				{ "content": { "parts": [ { "text": "summary result" } ] } }
			]
		}`))
	}))
	defer srv.Close()

	c := NewGeminiClient("test-key",
		WithGeminiBaseURL(srv.URL),
		WithGeminiModel("gemini-2.5-flash"),
	)

	got, err := c.Complete(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if got != "summary result" {
		t.Errorf("Complete() = %q, want %q", got, "summary result")
	}
	if !strings.Contains(gotPath, "gemini-2.5-flash") {
		t.Errorf("path should include model name, got %q", gotPath)
	}
	if !strings.Contains(gotPath, "generateContent") {
		t.Errorf("path should call generateContent, got %q", gotPath)
	}
	if !strings.Contains(gotQuery, "key=test-key") {
		t.Errorf("query should include api key, got %q", gotQuery)
	}
	if len(gotReq.Contents) != 1 || len(gotReq.Contents[0].Parts) != 1 || gotReq.Contents[0].Parts[0].Text != "hello" {
		t.Errorf("request body = %+v, want single content with prompt", gotReq)
	}
}

func TestGeminiClient_DefaultModel(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"x"}]}}]}`))
	}))
	defer srv.Close()

	c := NewGeminiClient("k", WithGeminiBaseURL(srv.URL))
	if _, err := c.Complete(context.Background(), "x"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if !strings.Contains(gotPath, "gemini-2.5-flash") {
		t.Errorf("default model path should contain gemini-2.5-flash, got %q", gotPath)
	}
}

func TestGeminiClient_Complete_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"code":429,"message":"quota"}}`))
	}))
	defer srv.Close()

	c := NewGeminiClient("k", WithGeminiBaseURL(srv.URL))

	_, err := c.Complete(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error on 429, got nil")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error should mention status, got: %v", err)
	}
}

func TestGeminiClient_Complete_EmptyCandidates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"candidates":[]}`))
	}))
	defer srv.Close()

	c := NewGeminiClient("k", WithGeminiBaseURL(srv.URL))

	_, err := c.Complete(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error on empty candidates, got nil")
	}
}
