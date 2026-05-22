package assistant

import (
	"regexp"
	"strings"
)

// contentFilter prevents skill call blocks from leaking to the user during streaming.
// Normal content flows through immediately. Once '~' is seen, buffering begins.
// If a complete ~ekken...ekken~ block arrives, SkillBlock is set and nothing is flushed.
// If the stream ends with buffered content that never formed a skill call, Flush() returns it.
type contentFilter struct {
	buf        strings.Builder
	triggered  bool
	SkillBlock string
}

func (f *contentFilter) Write(chunk string) string {
	if f.SkillBlock != "" {
		return ""
	}
	f.buf.WriteString(chunk)
	current := f.buf.String()

	idx := strings.Index(current, "~")
	if idx == -1 {
		// no ~ at all — safe to flush everything
		f.buf.Reset()
		return current
	}

	// flush everything before ~
	safe := current[:idx]
	held := current[idx:]
	f.buf.Reset()
	f.buf.WriteString(held)
	f.triggered = true

	// check if closing tag already present in held
	if regexp.MustCompile(`(?is)ekken\s*~`).MatchString(held) {
		f.SkillBlock = held
		f.buf.Reset()
		return safe
	}

	return safe
}

func (f *contentFilter) Flush() string {
	if f.SkillBlock != "" {
		return ""
	}
	s := f.buf.String()
	f.buf.Reset()
	f.triggered = false
	return s
}
