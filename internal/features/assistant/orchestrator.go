package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"ekken/internal/features/assistant/agents"
	"ekken/internal/features/assistant/skills"
	"ekken/internal/logger"
)

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{}
}

// OnChunk implements StreamListener.
func (s *loopSession) OnChunk(content, reasoning string) {
	if reasoning != "" {
		s.totalThinking.WriteString(reasoning)
	}

	if s.request.Stream && s.sink != nil {
		if reasoning != "" {
			_ = s.sink.Send(ChatResponse{
				ConversationID: s.request.ConversationID,
				Model:          s.request.Model,
				ProviderName:   s.provider.Info().Name,
				Message:        MessageContent{Role: "assistant", Thinking: reasoning},
			})
		}
		if content != "" {
			wasTriggered := s.filter.triggered
			if safe := s.filter.Write(content); safe != "" {
				s.visibleContent.WriteString(safe)
				_ = s.sink.Send(ChatResponse{
					ConversationID: s.request.ConversationID,
					Model:          s.request.Model,
					ProviderName:   s.provider.Info().Name,
					Message:        MessageContent{Role: "assistant", Content: safe},
				})
			}
			if !wasTriggered && s.filter.triggered {
				s.orchestrator.SendState(s.sink, s.request.ConversationID, s.request.Model, s.provider.Info().Name, "writing")
			}
		}
	}
}

func (o *Orchestrator) Execute(ctx context.Context, sink StreamSink, req ChatRequest, provider IProvider, hist *HistoryManager) (MessageContent, error) {
	session := &loopSession{
		orchestrator: o,
		sink:         sink,
		request:      req,
		provider:     provider,
		history:      hist,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	activeAgent := session.prepareStream()

	maxLoops := 5
	internalReq := session.request

	for i := range maxLoops {
		// Log history stack for debugging
		logger.DevPrintf("\n📜 [HISTORY STACK] Loop %d | Sending to Model: %s", i+1, internalReq.Model)
		logger.DevPrintf("--------------------------------------------------")
		for idx, m := range internalReq.Messages {
			content := m.Content
			if len(content) > 100 {
				content = content[:97] + "..."
			}
			content = strings.ReplaceAll(content, "\n", " ")
			logger.DevPrintf("[%d] %-10s | %s", idx, strings.ToUpper(m.Role), content)
		}
		logger.DevPrintf("--------------------------------------------------")

		logger.DevPrintf("\n🔄 [LOOP %d/%d] Requesting response from: %s", i+1, maxLoops, internalReq.Model)

		// Temporary debug logging
		o.logHistoryDebug(internalReq)

		assistantMsg, err := provider.Chat(ctx, internalReq, session)
		session.lastMsg = assistantMsg

		if leftover := session.filter.Flush(); leftover != "" && internalReq.Stream && session.sink != nil {
			session.visibleContent.WriteString(leftover)
			_ = session.sink.Send(ChatResponse{
				ConversationID: internalReq.ConversationID,
				Model:          internalReq.Model,
				ProviderName:   provider.Info().Name,
				Message:        MessageContent{Role: "assistant", Content: leftover},
			})
		}
		session.filter = contentFilter{}

		if assistantMsg.Thinking != "" {
			session.totalThinking.WriteString("\n\n")
		}

		if err != nil {
			logger.DevPrintf("❌ [LOOP %d] Model Error: %v", i+1, err)
			errMsg := fmt.Sprintf("\n\n%s%v", TagSystemError, err)
			state := "Error"
			if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
				errMsg = " [Stopped]"
				state = "stopped"
			}
			session.visibleContent.WriteString(errMsg)

			if internalReq.Stream && session.sink != nil {
				_ = session.sink.Send(ChatResponse{
					ConversationID: internalReq.ConversationID,
					Model:          internalReq.Model,
					ProviderName:   provider.Info().Name,
					Message: MessageContent{
						Role:    "assistant",
						Content: errMsg,
					},
					Done: true,
				})
				o.SendState(session.sink, internalReq.ConversationID, internalReq.Model, provider.Info().Name, state)
			}

			session.lastMsg.Content += errMsg
			session.commitToHistory()
			return assistantMsg, err
		}

		logger.DevPrintf("\n--- RAW OUTPUT ---\n%s\n----------------------------", assistantMsg.Content)

		shouldContinue, nextReq, err := session.executeSkillIfDetected(assistantMsg, activeAgent, internalReq, i, maxLoops)
		if err != nil {
			return assistantMsg, err
		}
		if !shouldContinue {
			break
		}
		internalReq = nextReq
	}

	session.commitToHistory()
	logger.DevPrintf("\n--- SESSION END ---")

	return session.lastMsg, nil
}

func (o *Orchestrator) logHistoryDebug(req ChatRequest) {
	path := "/Users/steen/.ekken/history-debug.txt"
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	jb, _ := json.MarshalIndent(req, "", "  ")
	f.WriteString("\n\n##################################################\n")
	f.WriteString(fmt.Sprintf("🚀 PAYLOAD TO MODEL AT: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	f.WriteString("##################################################\n\n")
	f.Write(jb)
	f.WriteString("\n\n")
}

func (s *loopSession) prepareStream() *agents.Agent {
	if s.request.Stream && s.sink != nil {
		_ = s.sink.Prepare(s.request.ConversationID, s.request.Model, s.provider.Info().Name)
		s.orchestrator.SendState(s.sink, s.request.ConversationID, s.request.Model, s.provider.Info().Name, "Loading")
	}

	var activeAgent *agents.Agent
	if s.request.Agent != "" {
		agent, err := agents.GetAgent(s.request.Agent)
		if err == nil {
			activeAgent = &agent
			systemMsg := agent.BuildSystemPrompt()
			logger.DevPrintf("\n🚀 Agent: %s | System Prompt:\n%s", s.request.Agent, systemMsg)

			cleanMessages := make([]MessageContent, 0)
			for _, m := range s.request.Messages {
				if m.Role != "system" {
					cleanMessages = append(cleanMessages, m)
				}
			}
			s.request.Messages = append([]MessageContent{{Role: "system", Content: systemMsg}}, cleanMessages...)
		}
	}
	return activeAgent
}

func (s *loopSession) executeSkillIfDetected(msg MessageContent, agent *agents.Agent, currentReq ChatRequest, loopIdx, maxLoops int) (bool, ChatRequest, error) {
	isSkillCall, parsedCall, _, _ := s.orchestrator.SkillCallDetector(msg.Content)
	retryErr := s.orchestrator.SkillCallFallback(msg.Content, isSkillCall, parsedCall, agent)

	if retryErr != "" {
		retryMsg := fmt.Sprintf("%s\nFormat Error! Fix the format. (Turn %d/%d)", retryErr, loopIdx+1, maxLoops)
		logger.DevPrintf("⚠️ Format Error. Retrying...")

		if s.request.Stream {
			s.orchestrator.SendState(s.sink, s.request.ConversationID, s.request.Model, s.provider.Info().Name, "fixing_format")
		}

		s.history.AddMessage(s.request.ConversationID, "assistant", msg.Content, "", s.provider.Info().ID, s.request.Model, s.request.Agent, false)
		s.history.AddMessage(s.request.ConversationID, "user", retryMsg, "", s.provider.Info().ID, s.request.Model, s.request.Agent, true)

		currentReq.Messages = append(currentReq.Messages, msg)
		currentReq.Messages = append(currentReq.Messages, MessageContent{Role: "user", Content: retryMsg})
		return true, currentReq, nil
	}

	if isSkillCall {
		logger.DevPrintf("🎯 Skill Call Detected: %s", parsedCall.Skill)
		if skill, ok := skills.Registry[parsedCall.Skill]; ok {
			if s.request.Stream {
				stateName := parsedCall.Skill
				s.orchestrator.SendState(s.sink, s.request.ConversationID, s.request.Model, s.provider.Info().Name, stateName)
			}

			logger.DevPrintf("🛠️  Executing: %s", parsedCall.Skill)
			res, err := skill.Execute(parsedCall.Args)
			if err != nil {
				res = "Error: " + err.Error()
			}

			skillResultContent := TagSkillResult + res
			logger.DevPrintf("✅ Skill Result: %s", skillResultContent)

			s.history.AddMessage(s.request.ConversationID, "assistant", msg.Content, "", s.provider.Info().ID, s.request.Model, s.request.Agent, false)
			s.history.AddMessage(s.request.ConversationID, "user", skillResultContent, "", s.provider.Info().ID, s.request.Model, s.request.Agent, true)

			currentReq.Messages = append(currentReq.Messages, msg)
			currentReq.Messages = append(currentReq.Messages, MessageContent{Role: "user", Content: skillResultContent})

			return true, currentReq, nil
		}
	}

	return false, currentReq, nil
}

func (s *loopSession) commitToHistory() {
	content := s.visibleContent.String()
	if content == "" && !s.request.Stream {
		content = s.lastMsg.Content
	}
	thinking := s.totalThinking.String()

	if content != "" || thinking != "" {
		s.history.AddMessage(s.request.ConversationID, "assistant", content, thinking, s.provider.Info().ID, s.request.Model, s.request.Agent, false)
	}

	if s.request.Stream && s.sink != nil {
		_ = s.sink.Done(s.request.ConversationID, s.request.Model)
	}
}
