package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"ekken/internal/features/assistant"
)

// OpenAI-style streaming and non-streaming response structures
type streamResponse struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content,omitempty"`
			Reasoning        string `json:"reasoning,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type nonStreamResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// thinkExtractor is a stateful parser that extracts <think>/<thinking> blocks
// from streaming content chunks that may be split across multiple calls.
type thinkExtractor struct {
	buf       strings.Builder
	inThink   bool
	openTags  []string
	closeTags []string
}

func newThinkExtractor() *thinkExtractor {
	return &thinkExtractor{
		openTags:  []string{"<think>", "<thinking>"},
		closeTags: []string{"</think>", "</thinking>"},
	}
}

// Process takes a raw content chunk and returns (cleanContent, reasoning).
func (e *thinkExtractor) Process(chunk string) (string, string) {
	e.buf.WriteString(chunk)
	raw := e.buf.String()
	e.buf.Reset()

	var content, reasoning strings.Builder

	for len(raw) > 0 {
		if e.inThink {
			// Look for closing tag
			found := false
			for _, ct := range e.closeTags {
				if idx := strings.Index(raw, ct); idx != -1 {
					reasoning.WriteString(raw[:idx])
					raw = raw[idx+len(ct):]
					e.inThink = false
					found = true
					break
				}
			}
			if !found {
				// Might be partial closing tag at end — buffer it
				for _, ct := range e.closeTags {
					for i := len(ct) - 1; i > 0; i-- {
						if strings.HasSuffix(raw, ct[:i]) {
							reasoning.WriteString(raw[:len(raw)-i])
							e.buf.WriteString(raw[len(raw)-i:])
							raw = ""
							break
						}
					}
					if raw == "" {
						break
					}
				}
				if raw != "" {
					reasoning.WriteString(raw)
					raw = ""
				}
			}
		} else {
			// Look for opening tag
			found := false
			for _, ot := range e.openTags {
				if idx := strings.Index(raw, ot); idx != -1 {
					content.WriteString(raw[:idx])
					raw = raw[idx+len(ot):]
					e.inThink = true
					found = true
					break
				}
			}
			if !found {
				// Might be partial opening tag at end — buffer it
				for _, ot := range e.openTags {
					for i := len(ot) - 1; i > 0; i-- {
						if strings.HasSuffix(raw, ot[:i]) {
							content.WriteString(raw[:len(raw)-i])
							e.buf.WriteString(raw[len(raw)-i:])
							raw = ""
							break
						}
					}
					if raw == "" {
						break
					}
				}
				if raw != "" {
					content.WriteString(raw)
					raw = ""
				}
			}
		}
	}

	return content.String(), reasoning.String()
}

// HandleOpenAIStream is the standard OpenAI-compatible stream handler.
func CommonChatStream(ctx context.Context, body io.Reader, listener assistant.StreamListener) (assistant.MessageContent, error) {
	extractor := newThinkExtractor()
	return assistant.StreamHandler(ctx, body, func(line string) (*assistant.StreamChunk, error) {
		lineStr := strings.TrimSpace(line)
		if lineStr == "" || strings.HasPrefix(lineStr, ":") {
			return nil, nil
		}
		if lineStr == "data: [DONE]" {
			return &assistant.StreamChunk{Done: true}, nil
		}
		if !strings.HasPrefix(lineStr, "data:") {
			if strings.HasPrefix(lineStr, "{") {
				var errObj map[string]any
				if json.Unmarshal([]byte(lineStr), &errObj) == nil {
					if errData, ok := errObj["error"]; ok {
						return nil, fmt.Errorf("provider error: %v", errData)
					}
				}
			}
			return nil, nil
		}

		data := strings.TrimPrefix(lineStr, "data:")
		var chunk streamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return nil, fmt.Errorf("stream parse error: %w, raw data: %s", err, data)
		}
		if len(chunk.Choices) == 0 {
			return nil, nil
		}

		delta := chunk.Choices[0].Delta

		// Priority: explicit reasoning field first, then extract from content
		reasoning := delta.ReasoningContent
		if reasoning == "" {
			reasoning = delta.Reasoning
		}

		cleanContent, extracted := extractor.Process(delta.Content)
		if extracted != "" {
			reasoning += extracted
		}

		done := chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != ""

		return &assistant.StreamChunk{
			Content:   cleanContent,
			Reasoning: reasoning,
			Done:      done,
		}, nil
	}, listener)
}

// HandleOpenAINonStream is the standard OpenAI-compatible non-stream handler.
func CommonChatNonStream(body io.Reader) (assistant.MessageContent, error) {
	rawBody, err := io.ReadAll(body)
	if err != nil {
		return assistant.MessageContent{}, err
	}
	var fullResp nonStreamResponse
	if err := json.Unmarshal(rawBody, &fullResp); err != nil {
		return assistant.MessageContent{}, err
	}

	content := ""
	if len(fullResp.Choices) > 0 {
		content = fullResp.Choices[0].Message.Content
	}

	return assistant.MessageContent{Role: "assistant", Content: content}, nil
}

// MapMessagesWithRoles allows custom role mapping for non-standard providers (e.g., Google's "model" role).
func MapMessagesWithRoles(req assistant.ChatRequest, roleMap map[string]string) []map[string]any {
	var cleanMessages []map[string]any
	for _, m := range req.Messages {
		role := m.Role

		// Apply custom mapping if provided
		if mappedRole, ok := roleMap[role]; ok {
			role = mappedRole
		}

		msg := map[string]any{"role": role, "content": m.Content}
		cleanMessages = append(cleanMessages, msg)
	}
	return cleanMessages
}

// MapMessages sanitizes the message content for AI providers, removing internal fields.
func MapMessages(messages []assistant.MessageContent) []map[string]any {
	var cleanMessages []map[string]any
	for _, m := range messages {
		msg := map[string]any{"role": m.Role, "content": m.Content}
		cleanMessages = append(cleanMessages, msg)
	}
	return cleanMessages
}

// NewOpenAIPayload creates a standardized payload map for OpenAI-compatible providers.
func NewOpenAIPayload(req assistant.ChatRequest) map[string]any {

	payload := map[string]any{
		"messages": MapMessages(req.Messages),
		"model":    req.Model,
		"stream":   req.Stream,
	}

	// Fallback to "low" if thinking is empty or "none" for strict providers like NVIDIA
	// if req.Thinking == "" || req.Thinking == "none" {
	// 	payload["reasoning_effort"] = "low"
	// } else {
	// 	payload["reasoning_effort"] = req.Thinking
	// }

	// Logging final payload for debugging
	logPayloadDebug(payload)

	return payload
}

func logPayloadDebug(payload any) {
	path := "/Users/steen/.ekken/history-debug.txt"
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	jb, _ := json.MarshalIndent(payload, "", "  ")
	f.WriteString("\n\n##################################################\n")
	f.WriteString(fmt.Sprintf("🚀 REAL PAYLOAD TO PROVIDER AT: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	f.WriteString("##################################################\n\n")
	f.Write(jb)
	f.WriteString("\n\n")
}
