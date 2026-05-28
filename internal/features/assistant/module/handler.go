package module

import (
	"context"
	"fmt"
	"net/http"

	"ekken/internal/api"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/features/assistant"
	"ekken/internal/features/assistant/agents"
	"ekken/internal/features/assistant/conversation"
	"ekken/internal/features/credential"

	"github.com/gin-gonic/gin"
)

type AssistantHandler struct {
	Config        config.Config
	Conversations conversation.Servicer
	History       *assistant.HistoryManager
	Orchestrator  *assistant.Orchestrator
	Jobs          *assistant.JobManager
	Models        *assistant.ModelManager
	Repo          *assistant.Repository
	DB            *db.DB
	Credentials   credential.Servicer
}

func (h *AssistantHandler) Chat(c *gin.Context) {
	providerID := c.Param("id")
	if providerID == "" {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: "provider id is required"})
		return
	}

	var req assistant.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	provider, err := assistant.GetProvider(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, api.Response{OK: false, Error: err.Error()})
		return
	}

	convID := req.ConversationID
	if convID == "" {
		title := "New Chat"
		if len(req.Messages) > 0 {
			title = req.Messages[len(req.Messages)-1].Content
			if len(title) > 30 {
				title = title[:30] + "..."
			}
		}
		conv, _ := h.Conversations.Create(title)
		convID = conv.ID
	}
	req.ConversationID = convID

	if req.Stream && h.Jobs != nil {
		if snapshot, ok := h.Jobs.Snapshot(convID); ok && snapshot.Running {
			c.JSON(http.StatusConflict, api.Response{OK: false, Error: "assistant chat is already running"})
			return
		}
	}

	historyMessages, err := h.History.GetContext(convID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: "failed to load history: " + err.Error()})
		return
	}

	if len(req.Messages) > 0 {
		userMsg := req.Messages[len(req.Messages)-1]
		if userMsg.Role == "user" {
			h.History.AddMessage(convID, "user", userMsg.Content, "", providerID, req.Model, req.Agent, false)
			historyMessages = append(historyMessages, userMsg)
		}
	}

	req.Messages = historyMessages

	if req.Stream {
		if h.Jobs == nil {
			c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: "assistant job manager is not initialized"})
			return
		}

		job, err := h.Jobs.Start(convID, func(ctx context.Context, sink assistant.StreamSink) error {
			_, err := h.Orchestrator.Execute(ctx, sink, req, provider, h.History)
			return err
		})
		if err != nil {
			status := http.StatusInternalServerError
			if err == assistant.ErrJobAlreadyRunning {
				status = http.StatusConflict
			}
			c.JSON(status, api.Response{OK: false, Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, api.Response{OK: true, Data: job.Snapshot()})
		return
	}

	lastAssistantMsg, err := h.Orchestrator.Execute(c.Request.Context(), nil, req, provider, h.History)
	if err != nil {
		h.handleChatError(c, req, lastAssistantMsg, err)
		return
	}

	if !req.Stream {
		c.JSON(http.StatusOK, api.Response{OK: true, Data: lastAssistantMsg})
	}
}

func (h *AssistantHandler) handleChatError(c *gin.Context, req assistant.ChatRequest, assistantMsg assistant.MessageContent, err error) {
	if !req.Stream && assistantMsg.Content == "" {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
	}
}

func (h *AssistantHandler) Providers(c *gin.Context) {
	catalog := assistant.ListProviderTypes()
	c.JSON(http.StatusOK, api.Response{OK: true, Data: catalog})
}

type providerWithModels struct {
	assistant.Provider
	Models []assistant.ModelInfo `json:"models"`
}

func (h *AssistantHandler) ListProviders(c *gin.Context) {
	providers := assistant.ListProviders()
	resData := make([]providerWithModels, len(providers))
	for i, p := range providers {
		var models []assistant.ModelInfo
		if h.Models != nil {
			models = h.Models.GetModels(p.ID)
		}
		resData[i] = providerWithModels{
			Provider: p,
			Models:   models,
		}
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: resData})
}

func (h *AssistantHandler) ProviderModels(c *gin.Context) {
	providerID := c.Param("id")
	if h.Models == nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: "Model manager not initialized"})
		return
	}
	uiModels := h.Models.GetModels(providerID)
	c.JSON(http.StatusOK, api.Response{OK: true, Data: uiModels})
}

func (h *AssistantHandler) ListAgents(c *gin.Context) {
	modes := agents.ListAgents()
	c.JSON(http.StatusOK, api.Response{OK: true, Data: modes})
}

func (h *AssistantHandler) SetupProvider(c *gin.Context) {
	var p assistant.Provider
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	if err := h.Repo.SaveAssistantProvider(p.ProviderID, p.Config); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}

	runtimeConfig, err := h.Credentials.ResolveConfig(p.Config)
	if err != nil {
		fmt.Printf("Warning: Failed to resolve credentials during setup: %v\n", err)
		runtimeConfig = p.Config
	}

	if err := assistant.CreateProvider(p.ProviderID, p.Config, runtimeConfig); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *AssistantHandler) DeleteProvider(c *gin.Context) {
	id := c.Param("id")
	assistant.RemoveProvider(id)
	if err := h.Repo.DeleteAssistantProvider(id); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

// UNIMPLEMENTED UNTIL WE HAVE A USE CASE
func (h *AssistantHandler) SyncModels(c *gin.Context) {
	if h.Models == nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: "Model manager not initialized"})
		return
	}
	force := c.Query("force") == "true"
	if err := h.Models.SyncWithEmbeddedDefaults(force); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *AssistantHandler) ChatJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: "conversation id is required"})
		return
	}
	if h.Jobs == nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: "assistant job manager is not initialized"})
		return
	}
	snapshot, ok := h.Jobs.Snapshot(id)
	if !ok {
		c.JSON(http.StatusOK, api.Response{OK: true, Data: assistant.JobSnapshot{
			ConversationID: id,
			Status:         assistant.JobStatusDone,
			Running:        false,
		}})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: snapshot})
}

func (h *AssistantHandler) RunningJobs(c *gin.Context) {
	if h.Jobs == nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: "assistant job manager is not initialized"})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: h.Jobs.Running()})
}

func (h *AssistantHandler) StreamChat(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: "conversation id is required"})
		return
	}
	if h.Jobs == nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: "assistant job manager is not initialized"})
		return
	}
	job, ok := h.Jobs.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, api.Response{OK: false, Error: "assistant job not found"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	ch, unsubscribe := job.Subscribe()
	defer unsubscribe()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			if _, err := c.Writer.Write(event); err != nil {
				return
			}
			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			}
		case <-c.Request.Context().Done():
			return
		}
	}
}

func (h *AssistantHandler) StopChat(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: "conversation id is required"})
		return
	}
	if h.Jobs != nil {
		if err := h.Jobs.Stop(id); err != nil {
			c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, api.Response{OK: true})
		return
	}
	c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: "assistant job manager is not initialized"})
}
