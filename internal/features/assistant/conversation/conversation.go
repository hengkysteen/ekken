package conversation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Servicer interface {
	Create(title string) (Conversation, error)
	List() ([]Conversation, error)
	Get(id string) (Conversation, []Message, error)
	Rename(id, title string) error
	Delete(id string) error
	DeleteAll() error
	AddMessage(conversationID, role, content, thinking, provider, model, agent string, isSystem bool) error
}

// Database is the modular storage interface for conversations.
type Database interface {
	CreateConversation(id, title string) error
	ListConversations() ([]ConversationItem, error)
	GetConversation(id string) (ConversationItem, error)
	UpdateConversationTitle(id, title string) error
	DeleteConversation(id string) error
	DeleteAllConversations() error
	AddMessage(id, conversationID, role, content, thinking, provider, model, agent string, isSystem bool) error
	GetMessages(conversationID string) ([]MessageItem, error)
}

type Service struct {
	db Database
}

func NewService(db Database) *Service {
	return &Service{db: db}
}

func (s *Service) Create(title string) (Conversation, error) {
	id := fmt.Sprintf("conv_%d", time.Now().UnixNano())
	if title == "" {
		title = "New Chat"
	}
	if err := s.db.CreateConversation(id, title); err != nil {
		return Conversation{}, err
	}
	return Conversation{
		ID:        id,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *Service) List() ([]Conversation, error) {
	items, err := s.db.ListConversations()
	if err != nil {
		return nil, err
	}
	res := make([]Conversation, 0, len(items))
	for _, item := range items {
		res = append(res, Conversation{
			ID:        item.ID,
			Title:     item.Title,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return res, nil
}

func (s *Service) Get(id string) (Conversation, []Message, error) {
	convItem, err := s.db.GetConversation(id)
	if err != nil {
		return Conversation{}, nil, err
	}

	msgItems, err := s.db.GetMessages(id)
	if err != nil {
		return Conversation{}, nil, err
	}

	conv := Conversation{
		ID:        convItem.ID,
		Title:     convItem.Title,
		CreatedAt: convItem.CreatedAt,
		UpdatedAt: convItem.UpdatedAt,
	}

	msgs := make([]Message, 0, len(msgItems))
	for _, m := range msgItems {
		msgs = append(msgs, Message{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Role:           m.Role,
			Content:        m.Content,
			Thinking:       m.Thinking,
			Provider:       m.Provider,
			Model:          m.Model,
			Agent:          m.Agent,
			IsSystem:       m.IsSystem,
			CreatedAt:      m.CreatedAt,
		})
	}

	return conv, msgs, nil
}

func (s *Service) Rename(id, title string) error {
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	return s.db.UpdateConversationTitle(id, title)
}

func (s *Service) Delete(id string) error {
	return s.db.DeleteConversation(id)
}

func (s *Service) DeleteAll() error {
	return s.db.DeleteAllConversations()
}

func (s *Service) AddMessage(conversationID, role, content, thinking, provider, model, agent string, isSystem bool) error {
	id := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	return s.db.AddMessage(id, conversationID, role, content, thinking, provider, model, agent, isSystem)
}

// FilterMessagesForDisplay filters out system messages for user display.
func FilterMessagesForDisplay(messages []Message) []Message {
	filtered := make([]Message, 0)

	// Regex untuk mencari dan menghapus blok skill call
	re := regexp.MustCompile(`(?is)~\s*ekken\s+skill\s+([a-zA-Z0-9_]+)\s*(.*?)\s*ekken\s*~`)
	for _, msg := range messages {
		// Abaikan pesan sistem
		if msg.IsSystem {
			continue
		}

		// Hapus tag skill call jika pesan dari assistant
		if msg.Role == "assistant" {
			msg.Content = strings.TrimSpace(re.ReplaceAllString(msg.Content, ""))
		}

		// Jangan masukkan jika setelah difilter pesannya kosong
		if msg.Content != "" {
			if msg.Role == "assistant" && len(filtered) > 0 && filtered[len(filtered)-1].Role == "assistant" {
				last := &filtered[len(filtered)-1]
				last.Content = joinDisplayText(last.Content, msg.Content)
				last.Thinking = joinDisplayText(last.Thinking, msg.Thinking)
				if msg.Provider != "" {
					last.Provider = msg.Provider
				}
				if msg.Model != "" {
					last.Model = msg.Model
				}
				if msg.Agent != "" {
					last.Agent = msg.Agent
				}
				continue
			}
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

func joinDisplayText(left, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	if strings.HasPrefix(right, left) {
		return right
	}
	if strings.HasPrefix(left, right) {
		return left
	}
	return left + "\n\n" + right
}
