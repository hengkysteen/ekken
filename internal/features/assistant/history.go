package assistant

import (
	"ekken/internal/features/assistant/conversation"
	"regexp"
	"strings"
)

func NewHistoryManager(svc conversation.Servicer) *HistoryManager {
	return &HistoryManager{
		convService: svc,
	}
}

// GetContext retrieves the full conversation history and formats it for the AI model.
func (m *HistoryManager) GetContext(convID string) ([]MessageContent, error) {
	if convID == "" {
		return []MessageContent{}, nil
	}

	_, messages, err := m.convService.Get(convID)
	if err != nil {
		return nil, err
	}

	// Regex to remove reasoning blocks like <think> or <thinking>
	// Clean up the assistant's internal thoughts to keep the model focused on the actual conversation.
	re := regexp.MustCompile(`(?s)<think>.*?</think>|<thinking>.*?</thinking>`)

	var context []MessageContent
	for _, msg := range messages {

		cleanContent := msg.Content
		if msg.Role == "assistant" {
			cleanContent = re.ReplaceAllString(cleanContent, "")
			cleanContent = strings.TrimSpace(cleanContent)
		}

		if cleanContent != "" {
			context = append(context, MessageContent{
				Role:    msg.Role,
				Content: cleanContent,
			})
		}
	}

	return context, nil
}

// AddMessage saves a new message to the history via the conversation service.
func (m *HistoryManager) AddMessage(convID, role, content, thinking, provider, model, agent string, isSystem bool) error {
	return m.convService.AddMessage(convID, role, content, thinking, provider, model, agent, isSystem)
}
