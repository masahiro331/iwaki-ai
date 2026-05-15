package discord

import "strings"

// DiscordMessageLimit is Discord's per-message character cap.
const DiscordMessageLimit = 2000

// SplitForDiscord breaks `text` into chunks no longer than `limit` characters.
// It prefers line boundaries; lines longer than `limit` are hard-cut to avoid
// silently dropping content. Returns nil for empty input.
func SplitForDiscord(text string, limit int) []string {
	if text == "" {
		return nil
	}
	if limit <= 0 {
		return []string{text}
	}

	var chunks []string
	var current strings.Builder

	flush := func() {
		if current.Len() > 0 {
			chunks = append(chunks, current.String())
			current.Reset()
		}
	}

	for _, line := range strings.Split(text, "\n") {
		// Hard-cut lines that exceed the limit on their own.
		for len(line) > limit {
			flush()
			chunks = append(chunks, line[:limit])
			line = line[limit:]
		}

		// Would adding this line exceed the limit? Need to account for the
		// joining newline when the buffer is non-empty.
		extra := len(line)
		if current.Len() > 0 {
			extra += 1 // newline separator
		}
		if current.Len()+extra > limit {
			flush()
		}

		if current.Len() > 0 {
			current.WriteByte('\n')
		}
		current.WriteString(line)
	}
	flush()
	return chunks
}
