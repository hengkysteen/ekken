package adapter

import (
	"context"
	"time"

	"ekken/internal/features/assistant"
	"ekken/internal/features/plugins/passistant"
)

type ProcessProvider struct {
	assistant.BaseProvider
	runner *passistant.ProcessRunner
}

func (p *ProcessProvider) Configure(config map[string]string) {
	p.BaseProvider.Configure(config)
	p.runner.Configure(config)
}

func (p *ProcessProvider) Chat(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	// Map assistant.ChatRequest to passistant.ChatRequest
	pReq := passistant.ChatRequest{
		ConversationID: req.ConversationID,
		Model:          req.Model,
		Agent:          req.Agent,
		Stream:         req.Stream,
		Thinking:       req.Thinking,
	}

	for _, msg := range req.Messages {
		pReq.Messages = append(pReq.Messages, passistant.MessageContent{
			Role:     msg.Role,
			Content:  msg.Content,
			Thinking: msg.Thinking,
			State:    msg.State,
		})
	}

	var onChunk func(content, thinking string)
	if listener != nil {
		onChunk = func(content, thinking string) {
			listener.OnChunk(content, thinking)
		}
	}

	// Execute process runner
	res, err := p.runner.Chat(ctx, pReq, onChunk)
	if err != nil {
		return assistant.MessageContent{}, err
	}

	// Map passistant.MessageContent back to assistant.MessageContent
	return assistant.MessageContent{
		Role:     res.Role,
		Content:  res.Content,
		Thinking: res.Thinking,
		State:    res.State,
	}, nil
}

type assistantRegistryAdapter struct {
}

func (a *assistantRegistryAdapter) Register(providerID string, runner passistant.RunnerSpec, provider passistant.ProviderSpec, sourcePath string, execTimeout time.Duration) error {
	pRunner := passistant.NewProcessRunner(runner, provider, sourcePath, execTimeout)

	pProvider := &ProcessProvider{
		BaseProvider: assistant.BaseProvider{
			ID:           provider.ID,
			Name:         provider.Name,
			Logo:         provider.Icon,
			OfficialURL:  provider.OfficialURL,
			ConfigFields: provider.ConfigFields,
			Config:       make(map[string]string),
		},
		runner: pRunner,
	}

	// Map models
	models := make([]assistant.ModelEntry, len(provider.Models))
	for i, m := range provider.Models {
		models[i] = assistant.ModelEntry{
			Name:          m.Name,
			Origin:        m.Origin,
			ContextWindow: m.ContextWindow,
		}
	}

	// Register with assistant global registry
	assistant.Register(provider.ID, func() assistant.IProvider {
		return pProvider
	}, models)

	// Dynamically register plugin models with ModelManager
	if mm := assistant.GetGlobalModelManager(); mm != nil {
		if err := mm.RegisterPluginModels(provider.ID, models); err != nil {
			return err
		}
	}

	return nil
}

func (a *assistantRegistryAdapter) Unregister(providerID string) error {
	assistant.UnregisterProviderType(providerID)
	assistant.RemoveProvider(providerID)
	return nil
}

func init() {
	passistant.GlobalRegistry = &assistantRegistryAdapter{}
}
