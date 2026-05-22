package conversation

import (
	"strings"
	"testing"
)

func TestMessage_IsSystemFlag(t *testing.T) {
	msg := Message{
		IsSystem: true,
	}

	if !msg.IsSystem {
		t.Error("Expected IsSystem to be true")
	}
}

func TestService_FilterSystemMessages(t *testing.T) {
	// Simulate messages from DB
	allMessages := []Message{
		{ID: "msg_1", Role: "user", Content: "Hello", IsSystem: false},
		{ID: "msg_2", Role: "assistant", Content: "Hi", IsSystem: false},
		{ID: "msg_3", Role: "user", Content: "[SYSTEM][SKILL_RESULT]: data", IsSystem: true},
		{ID: "msg_4", Role: "assistant", Content: "Response", IsSystem: false},
	}

	// Filter logic (same as handler)
	userMessages := make([]Message, 0)
	for _, msg := range allMessages {
		if !msg.IsSystem {
			userMessages = append(userMessages, msg)
		}
	}

	// Verify
	if len(userMessages) != 3 {
		t.Errorf("Expected 3 user messages, got %d", len(userMessages))
	}

	for _, msg := range userMessages {
		if msg.IsSystem {
			t.Errorf("System message leaked: %s", msg.ID)
		}
	}

	// Verify system message was filtered
	hasSystemMsg := false
	for _, msg := range userMessages {
		if msg.ID == "msg_3" {
			hasSystemMsg = true
		}
	}
	if hasSystemMsg {
		t.Error("System message (msg_3) should be filtered out")
	}
}

func TestService_ModelStillGetsSystemMessages(t *testing.T) {
	// Simulate GetContext behavior - returns ALL messages
	allMessages := []Message{
		{ID: "msg_1", Role: "user", Content: "Hello", IsSystem: false},
		{ID: "msg_2", Role: "assistant", Content: "Hi", IsSystem: false},
		{ID: "msg_3", Role: "user", Content: "[SYSTEM][SKILL_RESULT]: data", IsSystem: true},
		{ID: "msg_4", Role: "assistant", Content: "Response", IsSystem: false},
	}

	// Model gets all messages (no filter)
	modelMessages := allMessages

	// Verify model gets system messages
	if len(modelMessages) != 4 {
		t.Errorf("Expected model to get 4 messages, got %d", len(modelMessages))
	}

	hasSystemMsg := false
	for _, msg := range modelMessages {
		if msg.IsSystem {
			hasSystemMsg = true
			break
		}
	}
	if !hasSystemMsg {
		t.Error("Model should receive system messages")
	}
}

func TestFilterMessagesForDisplay_StripSkillCall(t *testing.T) {
	messages := []Message{
		{ID: "msg_1", Role: "user", Content: "Create workflow", IsSystem: false},
		{ID: "msg_2", Role: "assistant", Content: "Baik, saya akan buat workflow.\n\n~ekken skill skill_workflow_nodes {} ekken~", IsSystem: false},
		{ID: "msg_3", Role: "user", Content: "[SYSTEM][SKILL_RESULT]: nodes data", IsSystem: true},
		{ID: "msg_4", Role: "assistant", Content: "Workflow created", IsSystem: false},
	}

	filtered := FilterMessagesForDisplay(messages)

	// Verify system message filtered
	if len(filtered) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(filtered))
	}

	// Verify skill call stripped from assistant message
	for _, msg := range filtered {
		if msg.ID == "msg_2" {
			if strings.Contains(msg.Content, "~ekken") {
				t.Error("Skill call should be stripped from assistant message")
			}
			want := "Baik, saya akan buat workflow.\n\nWorkflow created"
			if msg.Content != want {
				t.Errorf("Expected %q, got %q", want, msg.Content)
			}
		}
	}
}

func TestFilterMessagesForDisplay_MergeAssistantToolLoopSession(t *testing.T) {
	messages := []Message{
		{ID: "msg_1", Role: "user", Content: "Create workflow", IsSystem: false},
		{ID: "msg_2", Role: "assistant", Content: "I'll check nodes.\n\n~ekken skill nodes ekken~", IsSystem: false},
		{ID: "msg_3", Role: "user", Content: "[SYSTEM][SKILL_RESULT]: nodes data", IsSystem: true},
		{ID: "msg_4", Role: "assistant", Content: "Now I'll create it.\n\n~ekken skill create_workflow\nname: test\nekken~", IsSystem: false},
		{ID: "msg_5", Role: "user", Content: "[SYSTEM][SKILL_RESULT]: created", IsSystem: true},
		{ID: "msg_6", Role: "assistant", Content: "I'll check nodes.\n\nNow I'll create it.\n\nWorkflow created successfully.", Thinking: "reasoning 1\n\nreasoning 2", IsSystem: false},
		{ID: "msg_7", Role: "user", Content: "simpan", IsSystem: false},
		{ID: "msg_8", Role: "assistant", Content: "Saving now.\n\n~ekken skill save_workflow\nid: tmp_1\nekken~", IsSystem: false},
		{ID: "msg_9", Role: "user", Content: "[SYSTEM][SKILL_RESULT]: saved", IsSystem: true},
		{ID: "msg_10", Role: "assistant", Content: "Saving now.\n\nWorkflow saved.", IsSystem: false},
	}

	filtered := FilterMessagesForDisplay(messages)

	if len(filtered) != 4 {
		t.Fatalf("expected 4 display messages, got %d", len(filtered))
	}

	if filtered[1].Role != "assistant" {
		t.Fatalf("expected second display message to be assistant, got %q", filtered[1].Role)
	}

	wantFirstAssistant := "I'll check nodes.\n\nNow I'll create it.\n\nWorkflow created successfully."
	if filtered[1].Content != wantFirstAssistant {
		t.Fatalf("expected first assistant session %q, got %q", wantFirstAssistant, filtered[1].Content)
	}
	if filtered[1].Thinking != "reasoning 1\n\nreasoning 2" {
		t.Fatalf("expected thinking to be preserved, got %q", filtered[1].Thinking)
	}

	if filtered[3].Role != "assistant" {
		t.Fatalf("expected fourth display message to be assistant, got %q", filtered[3].Role)
	}
	wantSecondAssistant := "Saving now.\n\nWorkflow saved."
	if filtered[3].Content != wantSecondAssistant {
		t.Fatalf("expected second assistant session %q, got %q", wantSecondAssistant, filtered[3].Content)
	}
}

func TestFilterMessagesForDisplay_PreserveNormalMessages(t *testing.T) {
	messages := []Message{
		{ID: "msg_1", Role: "user", Content: "Hello", IsSystem: false},
		{ID: "msg_2", Role: "assistant", Content: "Hi there!", IsSystem: false},
	}

	filtered := FilterMessagesForDisplay(messages)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(filtered))
	}

	if filtered[0].Content != "Hello" || filtered[1].Content != "Hi there!" {
		t.Error("Normal messages should be preserved unchanged")
	}
}

func TestFilterMessagesForDisplay_DoesNotMutateOriginal(t *testing.T) {
	originalContent := "Baik, saya akan buat workflow.\n\n~ekken skill skill_workflow_nodes {} ekken~"
	messages := []Message{
		{ID: "msg_1", Role: "assistant", Content: originalContent, IsSystem: false},
	}

	// Filter messages
	filtered := FilterMessagesForDisplay(messages)

	// Verify filtered message has stripped content
	if strings.Contains(filtered[0].Content, "~ekken") {
		t.Error("Filtered message should have skill call stripped")
	}

	// Verify original message is unchanged
	if messages[0].Content != originalContent {
		t.Errorf("Original message was mutated! Expected '%s', got '%s'", originalContent, messages[0].Content)
	}
}

func TestFilterMessagesForDisplay_StripSplitSkillMarkers(t *testing.T) {
	messages := []Message{
		{ID: "msg_1", Role: "assistant", Content: "Checking nodes.\n\n~\nekken skill nodes\nekken\n~", IsSystem: false},
	}

	filtered := FilterMessagesForDisplay(messages)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered message, got %d", len(filtered))
	}
	if filtered[0].Content != "Checking nodes." {
		t.Fatalf("expected split skill call to be stripped, got %q", filtered[0].Content)
	}
}
