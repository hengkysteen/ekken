package agents

import _ "embed"

//go:embed prompts/chat.md
var chatPrompt string

func init() {
	Register("chat", func() Agent {
		return Agent{
			Name:         "Standard Assistant",
			Description:  "Standard conversational AI for general tasks.",
			SystemPrompt: chatPrompt,
		}
	})
}
