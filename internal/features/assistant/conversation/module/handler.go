package module

import (
	"ekken/internal/api"
	"ekken/internal/features/assistant/conversation"

	"net/http"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	service conversation.Servicer
}

func (h *ConversationHandler) ListConversations(c *gin.Context) {
	items, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: items})
}

func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Title is optional
	}

	item, err := h.service.Create(req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, api.Response{OK: true, Data: item})
}

func (h *ConversationHandler) GetConversationById(c *gin.Context) {
	id := c.Param("id")
	conv, msgs, err := h.service.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, api.Response{OK: false, Error: err.Error()})
		return
	}

	// Filter system messages for user display
	userMessages := conversation.FilterMessagesForDisplay(msgs)

	c.JSON(http.StatusOK, api.Response{OK: true, Data: gin.H{
		"conversation": conv,
		"messages":     userMessages,
	}})
}

func (h *ConversationHandler) RenameConversationById(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	if err := h.service.Rename(id, req.Title); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *ConversationHandler) DeleteConversationById(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *ConversationHandler) DeleteAllConversations(c *gin.Context) {
	if err := h.service.DeleteAll(); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *ConversationHandler) AddMessageToConversation(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Role     string `json:"role" binding:"required"`
		Content  string `json:"content"`
		Thinking string `json:"thinking"`
		Provider string `json:"provider"`
		Model    string `json:"model"`
		Agent    string `json:"agent"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	if err := h.service.AddMessage(id, req.Role, req.Content, req.Thinking, req.Provider, req.Model, req.Agent, false); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}
