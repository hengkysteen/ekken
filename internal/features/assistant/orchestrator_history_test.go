package assistant

import (
	"context"
	"testing"

	"ekken/internal/features/assistant/conversation"
	"ekken/internal/features/assistant/skills"
)

type recordedMessage struct {
	role     string
	content  string
	thinking string
	isSystem bool
}

type fakeConversationService struct {
	messages []recordedMessage
}

func (s *fakeConversationService) Create(title string) (conversation.Conversation, error) {
	return conversation.Conversation{ID: "conv_test", Title: title}, nil
}

func (s *fakeConversationService) List() ([]conversation.Conversation, error) {
	return nil, nil
}

func (s *fakeConversationService) Get(id string) (conversation.Conversation, []conversation.Message, error) {
	return conversation.Conversation{ID: id}, nil, nil
}

func (s *fakeConversationService) Rename(id, title string) error {
	return nil
}

func (s *fakeConversationService) Delete(id string) error {
	return nil
}

func (s *fakeConversationService) DeleteAll() error {
	return nil
}

func (s *fakeConversationService) AddMessage(conversationID, role, content, thinking, provider, model, agent string, isSystem bool) error {
	s.messages = append(s.messages, recordedMessage{role: role, content: content, thinking: thinking, isSystem: isSystem})
	return nil
}

type fakeProvider struct{}

func (fakeProvider) Info() ProviderType {
	return ProviderType{ID: "fake", Name: "Fake"}
}

func (fakeProvider) Configure(config map[string]string) {}

func (fakeProvider) Chat(ctx context.Context, req ChatRequest, listener StreamListener) (MessageContent, error) {
	return MessageContent{}, nil
}

type discardStreamSink struct{}

func (discardStreamSink) Prepare(convID, model, provider string) error { return nil }
func (discardStreamSink) Send(data ChatResponse) error                 { return nil }
func (discardStreamSink) Done(convID, model string) error              { return nil }

func TestCommitToHistoryStoresOnlyVisibleStreamContent(t *testing.T) {
	service := &fakeConversationService{}
	session := &loopSession{
		orchestrator: NewOrchestrator(),
		sink:         discardStreamSink{},
		request: ChatRequest{
			ConversationID: "conv_test",
			Model:          "test-model",
			Stream:         true,
		},
		provider: fakeProvider{},
		history:  NewHistoryManager(service),
	}

	session.OnChunk("~ekken skill skill_workflow_nodes {} ekken~", "")
	if leftover := session.filter.Flush(); leftover != "" {
		session.visibleContent.WriteString(leftover)
	}
	session.filter = contentFilter{}

	session.OnChunk("The workflow has been created as a temporary draft.", "")
	session.commitToHistory()

	if len(service.messages) != 1 {
		t.Fatalf("expected 1 committed message, got %d", len(service.messages))
	}

	got := service.messages[0].content
	want := "The workflow has been created as a temporary draft."
	if got != want {
		t.Fatalf("expected visible content %q, got %q", want, got)
	}
	if service.messages[0].role != "assistant" {
		t.Fatalf("expected assistant role, got %q", service.messages[0].role)
	}
}

func TestCommitToHistoryStoresNonStreamFinalMessage(t *testing.T) {
	service := &fakeConversationService{}
	session := &loopSession{
		request: ChatRequest{
			ConversationID: "conv_test",
			Model:          "test-model",
		},
		provider: fakeProvider{},
		history:  NewHistoryManager(service),
		lastMsg:  MessageContent{Role: "assistant", Content: "Final response"},
	}

	session.commitToHistory()

	if len(service.messages) != 1 {
		t.Fatalf("expected 1 committed message, got %d", len(service.messages))
	}
	if service.messages[0].content != "Final response" {
		t.Fatalf("expected non-stream final response, got %q", service.messages[0].content)
	}
}

type noOpSkill struct{}

func (noOpSkill) GetID() string          { return "test_noop" }
func (noOpSkill) GetName() string        { return "Test Noop" }
func (noOpSkill) GetDescription() string { return "No-op skill for tests." }
func (noOpSkill) Execute(args map[string]interface{}) (string, error) {
	return "ok", nil
}

func TestExecuteSkillIfDetectedDoesNotStoreIntermediateThinking(t *testing.T) {
	skills.Register(noOpSkill{})

	service := &fakeConversationService{}
	session := &loopSession{
		orchestrator: NewOrchestrator(),
		sink:         discardStreamSink{},
		request: ChatRequest{
			ConversationID: "conv_test",
			Model:          "test-model",
		},
		provider: fakeProvider{},
		history:  NewHistoryManager(service),
	}

	msg := MessageContent{
		Role:     "assistant",
		Content:  "~ekken skill test_noop {} ekken~",
		Thinking: "intermediate thinking",
	}

	shouldContinue, _, err := session.executeSkillIfDetected(msg, nil, ChatRequest{}, 0, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !shouldContinue {
		t.Fatal("expected skill execution to continue loop")
	}
	if len(service.messages) != 2 {
		t.Fatalf("expected 2 stored messages, got %d", len(service.messages))
	}
	if service.messages[0].thinking != "" {
		t.Fatalf("expected intermediate assistant thinking not stored, got %q", service.messages[0].thinking)
	}
}
