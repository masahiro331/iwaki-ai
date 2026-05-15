// Package message defines the common Message type used across the application.
package message

import (
	"fmt"
	"strings"
	"time"
)

// Message represents a single chat message normalized from any source.
type Message struct {
	Author    string
	Content   string
	Timestamp time.Time
}

// Format returns a human-readable single-entry representation
// suitable for feeding into an LLM prompt.
func (m Message) Format() string {
	return fmt.Sprintf("[%s] %s: %s",
		m.Timestamp.Format("2006-01-02 15:04"),
		m.Author,
		m.Content,
	)
}

// FormatAll joins formatted messages with newlines.
func FormatAll(msgs []Message) string {
	if len(msgs) == 0 {
		return ""
	}
	lines := make([]string, 0, len(msgs))
	for _, m := range msgs {
		lines = append(lines, m.Format())
	}
	return strings.Join(lines, "\n")
}
