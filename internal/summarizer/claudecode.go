package summarizer

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner abstracts running an external command with stdin/stdout.
// Tests substitute a fake; production uses execRunner.
type Runner interface {
	Run(ctx context.Context, name string, args []string, stdin string) (string, error)
}

// ClaudeCodeClient calls the local `claude -p` binary, using the user's
// Claude Code login (e.g. Max plan) instead of an API key.
type ClaudeCodeClient struct {
	binary string
	runner Runner
}

// ClaudeCodeOption customizes a ClaudeCodeClient.
type ClaudeCodeOption func(*ClaudeCodeClient)

// WithBinary overrides the claude executable path.
func WithBinary(path string) ClaudeCodeOption {
	return func(c *ClaudeCodeClient) { c.binary = path }
}

// WithRunner injects a custom Runner (mainly for tests).
func WithRunner(r Runner) ClaudeCodeOption {
	return func(c *ClaudeCodeClient) { c.runner = r }
}

// NewClaudeCodeClient builds a Claude Code-backed LLMClient.
func NewClaudeCodeClient(opts ...ClaudeCodeOption) *ClaudeCodeClient {
	c := &ClaudeCodeClient{
		binary: "claude",
		runner: execRunner{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Complete pipes the prompt into `claude -p` and returns the trimmed output.
func (c *ClaudeCodeClient) Complete(ctx context.Context, prompt string) (string, error) {
	out, err := c.runner.Run(ctx, c.binary, []string{"-p"}, prompt)
	if err != nil {
		return "", fmt.Errorf("run claude -p: %w", err)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return "", fmt.Errorf("claude -p returned empty output")
	}
	return trimmed, nil
}

// execRunner runs commands via os/exec.
type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args []string, stdin string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}
