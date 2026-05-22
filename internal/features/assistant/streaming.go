package assistant

import (
	"bufio"
	"context"
	"io"
	"strings"
)

// SendState sends a transient state update via SSE (e.g., "Thinking...", "searching").
func (o *Orchestrator) SendState(sink StreamSink, convID, model, provider, state string) {
	if sink == nil {
		return
	}
	_ = sink.Send(ChatResponse{
		ConversationID: convID,
		Model:          model,
		ProviderName:   provider,
		Message: MessageContent{
			Role:  "assistant",
			State: state,
		},
	})
}

// StreamChunk represents a normalized chunk of data from any provider.
type StreamChunk struct {
	Content   string
	Reasoning string
	Done      bool
}

// ChunkMapper is a function strategy to parse a raw line from a provider's stream.
type ChunkMapper func(line string) (*StreamChunk, error)

// StreamListener is an interface for observing streaming chunks.
type StreamListener interface {
	OnChunk(content, reasoning string)
}

type StreamSink interface {
	Prepare(convID, model, provider string) error
	Send(data ChatResponse) error
	Done(convID, model string) error
}

// StreamHandler is the unified engine for handling streaming responses from any AI provider.
func StreamHandler(ctx context.Context, body io.Reader, mapper ChunkMapper, listener StreamListener) (MessageContent, error) {
	scanner := bufio.NewScanner(body)
	var fullContent, fullThinking strings.Builder

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return MessageContent{Role: "assistant", Content: fullContent.String(), Thinking: fullThinking.String()}, ctx.Err()
		default:
		}

		line := scanner.Text()
		chunk, err := mapper(line)
		if err != nil {
			return MessageContent{Role: "assistant", Content: fullContent.String(), Thinking: fullThinking.String()}, err
		}
		if chunk == nil {
			continue
		}

		if chunk.Reasoning != "" {
			fullThinking.WriteString(chunk.Reasoning)
			if listener != nil {
				listener.OnChunk("", chunk.Reasoning)
			}
		}

		if chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
			if listener != nil {
				listener.OnChunk(chunk.Content, "")
			}
		}

		if chunk.Done {
			break
		}
	}

	return MessageContent{Role: "assistant", Content: fullContent.String(), Thinking: fullThinking.String()}, nil
}
