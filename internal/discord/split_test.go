package discord

import (
	"strings"
	"testing"
)

func TestSplitForDiscord_ShortText(t *testing.T) {
	got := SplitForDiscord("hello", 2000)
	if len(got) != 1 || got[0] != "hello" {
		t.Errorf("short text should be a single chunk, got %v", got)
	}
}

func TestSplitForDiscord_Empty(t *testing.T) {
	got := SplitForDiscord("", 2000)
	if len(got) != 0 {
		t.Errorf("empty input should produce no chunks, got %v", got)
	}
}

func TestSplitForDiscord_SplitsOnLineBoundaries(t *testing.T) {
	// 4 lines of 50 chars + newline = 51 chars each => 204 chars total
	// With limit=100 we expect ~2 chunks split on line boundaries.
	line := strings.Repeat("a", 50)
	in := strings.Join([]string{line, line, line, line}, "\n")

	got := SplitForDiscord(in, 100)

	if len(got) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(got))
	}
	for i, chunk := range got {
		if len(chunk) > 100 {
			t.Errorf("chunk %d exceeds limit: %d chars", i, len(chunk))
		}
	}
	// Joining the chunks (with newlines as separators) should reconstruct the input.
	if strings.Join(got, "\n") != in {
		t.Errorf("chunks should reconstruct input when joined with newlines:\ngot=%q\nwant=%q", strings.Join(got, "\n"), in)
	}
}

func TestSplitForDiscord_LongLineGetsHardCut(t *testing.T) {
	// One line of 150 chars with limit=100 => must hard-cut, not silently lose content.
	in := strings.Repeat("x", 150)

	got := SplitForDiscord(in, 100)
	if len(got) < 2 {
		t.Fatalf("expected hard-cut into multiple chunks, got %d", len(got))
	}
	total := 0
	for _, c := range got {
		if len(c) > 100 {
			t.Errorf("chunk exceeds limit: %d", len(c))
		}
		total += len(c)
	}
	if total != 150 {
		t.Errorf("total length lost: got %d, want 150", total)
	}
}
