package message

import (
	"strings"
	"testing"
	"time"
)

func TestMessage_Format(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	tests := []struct {
		name string
		msg  Message
		want string
	}{
		{
			name: "basic message",
			msg: Message{
				Author:    "alice",
				Content:   "hello world",
				Timestamp: time.Date(2026, 5, 15, 10, 30, 0, 0, jst),
			},
			want: "[2026-05-15 10:30] alice: hello world",
		},
		{
			name: "multi-line content is preserved",
			msg: Message{
				Author:    "bob",
				Content:   "line1\nline2",
				Timestamp: time.Date(2026, 5, 15, 10, 30, 0, 0, jst),
			},
			want: "[2026-05-15 10:30] bob: line1\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.Format()
			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatAll(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	msgs := []Message{
		{Author: "alice", Content: "hi", Timestamp: time.Date(2026, 5, 15, 10, 30, 0, 0, jst)},
		{Author: "bob", Content: "hello", Timestamp: time.Date(2026, 5, 15, 10, 31, 0, 0, jst)},
	}
	got := FormatAll(msgs)
	if !strings.Contains(got, "alice: hi") || !strings.Contains(got, "bob: hello") {
		t.Errorf("FormatAll() missing entries: %q", got)
	}
	if !strings.Contains(got, "\n") {
		t.Errorf("FormatAll() should join with newlines: %q", got)
	}
}

func TestFormatAll_Empty(t *testing.T) {
	got := FormatAll(nil)
	if got != "" {
		t.Errorf("FormatAll(nil) = %q, want empty string", got)
	}
}
