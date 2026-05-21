package passistant

type RunnerSpec struct {
	Type    string `json:"type,omitempty"`
	Command string `json:"command"`
}

type ModelSpecEntry struct {
	Name          string `json:"name"`
	Origin        string `json:"origin"`
	ContextWindow int    `json:"context_window"`
}

type ProviderSpec struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Icon         string           `json:"icon"`
	OfficialURL  string           `json:"official_url"`
	ConfigFields []string         `json:"config_fields"`
	Models       []ModelSpecEntry `json:"models"`
}

type PluginSpec struct {
	Runner   RunnerSpec   `json:"runner"`
	Provider ProviderSpec `json:"provider"`
}

type MessageContent struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
	State    string `json:"state,omitempty"`
}

type ChatRequest struct {
	ConversationID string           `json:"conversation_id"`
	Model          string           `json:"model"`
	Messages       []MessageContent `json:"messages"`
	Agent          string           `json:"agent,omitempty"`
	Stream         bool             `json:"stream"`
	Thinking       string           `json:"thinking"`
}

// StdinRequest is the request payload written to the subprocess's stdin.
type StdinRequest struct {
	Kind       string         `json:"kind"`
	ProviderID string         `json:"provider_id"`
	Request    ChatRequest    `json:"request"`
	Config     map[string]any `json:"config"`
}

// StreamResponseChunk is the JSON line structure read from the subprocess's stdout.
type StreamResponseChunk struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
	Error    string `json:"error,omitempty"`
	Message  *struct {
		Role     string `json:"role"`
		Content  string `json:"content"`
		Thinking string `json:"thinking,omitempty"`
	} `json:"message,omitempty"`
}
