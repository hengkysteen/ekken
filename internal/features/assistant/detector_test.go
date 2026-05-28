package assistant

import (
	"strings"
	"testing"
)

func TestDetector_TextBeforeSkillCall(t *testing.T) {
	o := &Orchestrator{}

	content := "Untuk membuat workflow yang mengambil screenshot...\n\n~ekken skill nodes {} ekken~"

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	t.Logf("Content: %q", content)
	t.Logf("isSkillCall: %v", isSkill)
	t.Logf("Parsed Skill: %q", parsed.Skill)
	t.Logf("Error: %q", retryErr)

	if retryErr != "" {
		t.Errorf("FAIL: Text + skill call menghasilkan error: %s", retryErr)
	} else {
		t.Logf("PASS: Text + skill call tidak error")
	}
}

func TestDetector_PureSkillCall(t *testing.T) {
	o := &Orchestrator{}

	content := "~ekken skill nodes {} ekken~"

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	t.Logf("Content: %q", content)
	t.Logf("isSkillCall: %v", isSkill)
	t.Logf("Parsed Skill: %q", parsed.Skill)
	t.Logf("Error: %q", retryErr)

	if retryErr != "" {
		t.Errorf("FAIL: Pure skill call menghasilkan error: %s", retryErr)
	} else {
		t.Logf("PASS: Pure skill call tidak error")
	}
}

func TestDetector_SplitOpeningSkillMarker(t *testing.T) {
	o := &Orchestrator{}

	content := "~\nekken skill nodes\nekken~"

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	if !isSkill {
		t.Fatal("expected split opening marker to be detected as skill call")
	}
	if parsed.Skill != "nodes" {
		t.Fatalf("expected nodes skill, got %q", parsed.Skill)
	}
	if retryErr != "" {
		t.Fatalf("expected no retry error, got: %s", retryErr)
	}
}

func TestDetector_TildeWithNonEkkenTextIsNotSkillCall(t *testing.T) {
	o := &Orchestrator{}

	content := "~\njembut"

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	if isSkill {
		t.Fatal("expected non-ekken tilde text not to be detected as skill call")
	}
	if retryErr != "" {
		t.Fatalf("expected no retry error for non-ekken tilde text, got: %s", retryErr)
	}
}

func TestDetector_TildeWithNonEkkenTextBeforeSkillCall(t *testing.T) {
	o := &Orchestrator{}

	content := "literal ~\njembut\n\n~\nekken skill nodes\nekken~"

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	if !isSkill {
		t.Fatal("expected later split ekken marker to be detected as skill call")
	}
	if parsed.Skill != "nodes" {
		t.Fatalf("expected nodes skill, got %q", parsed.Skill)
	}
	if retryErr != "" {
		t.Fatalf("expected no retry error, got: %s", retryErr)
	}
}

func TestDetector_SplitClosingSkillMarker(t *testing.T) {
	o := &Orchestrator{}

	content := "~ekken skill nodes\nekken\n~"

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	if !isSkill {
		t.Fatal("expected split closing marker to be detected as skill call")
	}
	if parsed.Skill != "nodes" {
		t.Fatalf("expected nodes skill, got %q", parsed.Skill)
	}
	if retryErr != "" {
		t.Fatalf("expected no retry error, got: %s", retryErr)
	}
}

func TestDetector_TextAfterSkillCall(t *testing.T) {
	o := &Orchestrator{}

	content := "~ekken skill nodes {} ekken~\n\nSelesai."

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	t.Logf("Content: %q", content)
	t.Logf("isSkillCall: %v", isSkill)
	t.Logf("Parsed Skill: %q", parsed.Skill)
	t.Logf("Error: %q", retryErr)

	if retryErr != "" {
		t.Errorf("FAIL: Text after skill call menghasilkan error: %s", retryErr)
	} else {
		t.Logf("PASS: Text after skill call tidak error")
	}
}

func TestDetector_YAMLArguments(t *testing.T) {
	o := &Orchestrator{}

	content := `Saya cek detail action yang dibutuhkan dulu.

~ekken skill nodes_actions
actions:
  - google_chrome.launch
  - chromedp.navigate
ekken~`

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	if !isSkill {
		t.Fatal("expected YAML skill arguments to be detected")
	}
	if retryErr != "" {
		t.Fatalf("expected no retry error, got: %s", retryErr)
	}
	if parsed.Skill != "nodes_actions" {
		t.Fatalf("expected nodes_actions, got: %s", parsed.Skill)
	}

	actions, ok := parsed.Args["actions"].([]interface{})
	if !ok {
		t.Fatalf("expected actions array, got: %#v", parsed.Args["actions"])
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got: %d", len(actions))
	}
}

func TestDetector_InvalidYAMLArgumentsReturnsSpecificError(t *testing.T) {
	o := &Orchestrator{}

	content := `Saya coba lagi.

~ekken skill nodes_actions
actions:
  - google_chrome.launch
  bad: [
ekken~`

	isSkill, parsed, _, _ := o.SkillCallDetector(content)
	retryErr := o.SkillCallFallback(content, isSkill, parsed, nil)

	if isSkill {
		t.Fatal("expected invalid YAML arguments to prevent skill detection")
	}
	if !strings.Contains(retryErr, "Invalid skill arguments for 'nodes_actions'") {
		t.Fatalf("expected specific YAML error, got: %s", retryErr)
	}
	if strings.Contains(retryErr, "Invalid format! You must use") {
		t.Fatalf("expected argument-specific error, got generic format error: %s", retryErr)
	}
}
