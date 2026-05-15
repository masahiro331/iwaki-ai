package summarizer

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeRunner struct {
	gotName  string
	gotArgs  []string
	gotStdin string
	reply    string
	err      error
}

func (f *fakeRunner) Run(_ context.Context, name string, args []string, stdin string) (string, error) {
	f.gotName = name
	f.gotArgs = args
	f.gotStdin = stdin
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

func TestClaudeCodeClient_Complete_Success(t *testing.T) {
	r := &fakeRunner{reply: "  要約結果  \n"}
	c := NewClaudeCodeClient(WithRunner(r))

	got, err := c.Complete(context.Background(), "Hello world")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if got != "要約結果" {
		t.Errorf("Complete() = %q, want trimmed reply", got)
	}
	if r.gotName != "claude" {
		t.Errorf("default binary = %q, want claude", r.gotName)
	}
	if !containsArg(r.gotArgs, "-p") {
		t.Errorf("args = %v, must include -p", r.gotArgs)
	}
	if r.gotStdin != "Hello world" {
		t.Errorf("stdin = %q, want prompt verbatim", r.gotStdin)
	}
}

func TestClaudeCodeClient_Complete_CustomBinary(t *testing.T) {
	r := &fakeRunner{reply: "ok"}
	c := NewClaudeCodeClient(WithBinary("/usr/local/bin/claude"), WithRunner(r))

	if _, err := c.Complete(context.Background(), "x"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if r.gotName != "/usr/local/bin/claude" {
		t.Errorf("binary = %q, want override", r.gotName)
	}
}

func TestClaudeCodeClient_Complete_RunnerError(t *testing.T) {
	runErr := errors.New("exec failed")
	r := &fakeRunner{err: runErr}
	c := NewClaudeCodeClient(WithRunner(r))

	_, err := c.Complete(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, runErr) {
		t.Errorf("error should wrap runner error, got %v", err)
	}
}

func TestClaudeCodeClient_Complete_EmptyOutput(t *testing.T) {
	r := &fakeRunner{reply: "   \n\t"}
	c := NewClaudeCodeClient(WithRunner(r))

	_, err := c.Complete(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error on empty output, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention empty output, got %v", err)
	}
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
