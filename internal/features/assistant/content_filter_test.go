package assistant

import (
	"testing"
)

func TestContentFilter_NormalContent(t *testing.T) {
	f := contentFilter{}
	out := f.Write("hello world")
	if out != "hello world" {
		t.Fatalf("expected 'hello world', got %q", out)
	}
	if f.SkillBlock != "" {
		t.Fatal("expected no skill block")
	}
}

func TestContentFilter_SkillCallBlocked(t *testing.T) {
	f := contentFilter{}
	chunks := []string{"~ekken skill ", "skill_workflow_nodes ", "{} ekken~"}
	for _, c := range chunks {
		f.Write(c)
	}
	if f.SkillBlock == "" {
		t.Fatal("expected skill block to be captured")
	}
}

func TestContentFilter_SkillCallNotLeaked(t *testing.T) {
	f := contentFilter{}
	var sent string
	for _, c := range []string{"~ekken skill skill_foo {} ekken~"} {
		sent += f.Write(c)
	}
	sent += f.Flush()
	if sent != "" {
		t.Fatalf("skill call leaked to user: %q", sent)
	}
}

func TestContentFilter_SplitSkillMarkersNotLeaked(t *testing.T) {
	f := contentFilter{}
	var sent string
	for _, c := range []string{"~\nekken skill skill_foo {}", " ekken\n~"} {
		sent += f.Write(c)
	}
	sent += f.Flush()
	if sent != "" {
		t.Fatalf("split skill call leaked to user: %q", sent)
	}
}

func TestContentFilter_TextBeforeSkillCall(t *testing.T) {
	f := contentFilter{}
	out := f.Write("sure thing ")
	out += f.Write("~ekken skill skill_foo {} ekken~")
	out += f.Flush()
	if out != "sure thing " {
		t.Fatalf("expected 'sure thing ', got %q", out)
	}
	if f.SkillBlock == "" {
		t.Fatal("expected skill block captured")
	}
}

func TestContentFilter_FlushLeftover(t *testing.T) {
	f := contentFilter{}
	out := f.Write("hello ~world")
	out += f.Flush()
	if out != "hello ~world" {
		t.Fatalf("expected 'hello ~world', got %q", out)
	}
}

func TestContentFilter_NonSkillTextAfterTildeFlushes(t *testing.T) {
	f := contentFilter{}
	out := f.Write("hello ")
	out += f.Write("~\njembut")
	out += f.Flush()
	if out != "hello ~\njembut" {
		t.Fatalf("expected non-skill tilde text to flush, got %q", out)
	}
}
