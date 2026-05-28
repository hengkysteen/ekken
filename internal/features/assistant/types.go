package assistant

import (
	"ekken/internal/features/assistant/conversation"
	"strings"
	"sync"
)

const (
	TagSkillResult = "[SYSTEM][SKILL_RESULT]: "
	TagSystemError = "[SYSTEM][ERROR]: "
)

type ChatRequest struct {
	ConversationID string           `json:"conversation_id"`
	Model          string           `json:"model" binding:"required"`
	Messages       []MessageContent `json:"messages"`
	Agent          string           `json:"agent,omitempty"`
	Stream         bool             `json:"stream"`
	Thinking       string           `json:"thinking"`
}

type MessageContent struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
	State    string `json:"state,omitempty"`
}

type ChatResponse struct {
	ConversationID string         `json:"conversation_id,omitempty"`
	Model          string         `json:"model"`
	ProviderName   string         `json:"provider"`
	Message        MessageContent `json:"message"`
	Done           bool           `json:"done"`
}

type ProviderType struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Logo         string   `json:"logo"`
	BaseURL      string   `json:"base_url"`
	OfficialURL  string   `json:"official_url"`
	ConfigFields []string `json:"config_fields"`
}

type Provider struct {
	ID          string            `json:"id"`
	ProviderID  string            `json:"provider_id"`
	Name        string            `json:"name"`
	Logo        string            `json:"logo"`
	BaseURL     string            `json:"base_url"`
	OfficialURL string            `json:"official_url"`
	Config      map[string]string `json:"config"`
}

type ModelInfo struct {
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	Name          string `json:"name"`
	ContextWindow int    `json:"context_window"`
	Type          string `json:"type"`
}

type ModelEntry struct {
	Name          string `json:"name"`
	Origin        string `json:"origin"`
	ContextWindow int    `json:"context_window"`
}

type ProviderModels struct {
	Provider string       `json:"provider"`
	Models   []ModelEntry `json:"models"`
}

type ModelConfig struct {
	Date   string           `json:"date"`
	System []ProviderModels `json:"system"`
	User   []ProviderModels `json:"user"`
}

type ModelManager struct {
	mu       sync.RWMutex
	filePath string
	config   ModelConfig
}

type Orchestrator struct {
}

type loopSession struct {
	orchestrator   *Orchestrator
	sink           StreamSink
	request        ChatRequest
	provider       IProvider
	history        *HistoryManager
	lastMsg        MessageContent
	visibleContent strings.Builder
	totalThinking  strings.Builder
	filter         contentFilter
}

type HistoryManager struct {
	convService conversation.Servicer
}
