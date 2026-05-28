package conversation

import (
	"os"
	"testing"

	"ekken/internal/db"
)

func setupTestDB(t *testing.T) (*Repository, func()) {
	t.Helper()

	// Create temp DB
	tmpDir := t.TempDir()

	database, err := db.Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}

	repo := NewRepository(database)

	cleanup := func() {
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return repo, cleanup
}

func TestIntegration_FilterMessagesWithRealDB(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create conversation
	convID := "test_conv_1"
	err := repo.CreateConversation(convID, "Test Conversation")
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Add messages
	messages := []struct {
		id       string
		role     string
		content  string
		isSystem bool
	}{
		{"msg_1", "user", "Create workflow", false},
		{"msg_2", "assistant", "Baik, saya akan buat workflow.\n\n~ekken skill skill_workflow_nodes {} ekken~", false},
		{"msg_3", "user", "[SYSTEM][SKILL_RESULT]: nodes data", true},
		{"msg_4", "assistant", "Workflow created successfully", false},
	}

	for _, msg := range messages {
		err := repo.AddMessage(msg.id, convID, msg.role, msg.content, "", "", "", "", msg.isSystem)
		if err != nil {
			t.Fatalf("Failed to add message %s: %v", msg.id, err)
		}
	}

	// Get messages from DB
	dbMessages, err := repo.GetMessages(convID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	// Verify all messages in DB
	if len(dbMessages) != 4 {
		t.Errorf("Expected 4 messages in DB, got %d", len(dbMessages))
	}

	// Convert to domain messages
	domainMessages := make([]Message, len(dbMessages))
	for i, m := range dbMessages {
		domainMessages[i] = Message{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Role:           m.Role,
			Content:        m.Content,
			IsSystem:       m.IsSystem,
		}
	}

	// Filter for display
	filtered := FilterMessagesForDisplay(domainMessages)

	// Verify filtered results
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered messages, got %d", len(filtered))
	}

	// Verify system message filtered out
	for _, msg := range filtered {
		if msg.IsSystem {
			t.Errorf("System message should be filtered: %s", msg.ID)
		}
		if msg.ID == "msg_3" {
			t.Error("System message msg_3 should not be in filtered results")
		}
	}

	// Verify skill call stripped from msg_2
	for _, msg := range filtered {
		if msg.ID == "msg_2" {
			if msg.Content != "Baik, saya akan buat workflow.\n\nWorkflow created successfully" {
				t.Errorf("Expected skill call stripped, got: %s", msg.Content)
			}
		}
	}

	// Verify original DB data unchanged
	dbMessagesAfter, _ := repo.GetMessages(convID)
	for _, m := range dbMessagesAfter {
		if m.ID == "msg_2" {
			if m.Content != "Baik, saya akan buat workflow.\n\n~ekken skill skill_workflow_nodes {} ekken~" {
				t.Error("Original DB content should be unchanged")
			}
		}
	}
}

func TestIntegration_IsSystemColumnExists(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create conversation and add message with is_system flag
	convID := "test_conv_2"
	repo.CreateConversation(convID, "Test")

	err := repo.AddMessage("msg_1", convID, "user", "[SYSTEM][SKILL_RESULT]: test", "", "", "", "", true)
	if err != nil {
		t.Fatalf("Failed to add message with is_system flag: %v", err)
	}

	// Retrieve and verify
	messages, err := repo.GetMessages(convID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if !messages[0].IsSystem {
		t.Error("Expected IsSystem to be true")
	}
}
