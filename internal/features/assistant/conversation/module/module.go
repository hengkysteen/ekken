package module

import (
	"ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/features/assistant/conversation"

	"github.com/gin-gonic/gin"
)

type ConversationModule struct {
	service conversation.Servicer
}

func NewModule() *ConversationModule {
	return &ConversationModule{}
}

func init() {
	module.RegisterModule(NewModule())
}

func (m *ConversationModule) Name() string {
	return "conversation"
}

func (m *ConversationModule) Init(database *db.DB, cfg config.Config) error {
	repo := conversation.NewRepository(database)
	m.service = conversation.NewService(repo)
	return nil
}

func (m *ConversationModule) RegisterRoutes(api *gin.RouterGroup) {
	h := &ConversationHandler{service: m.service}

	convs := api.Group("/conversations")
	{
		convs.GET("", h.ListConversations)
		convs.POST("", h.CreateConversation)
		convs.DELETE("", h.DeleteAllConversations)
		convs.GET("/:id", h.GetConversationById)
		convs.PUT("/:id/rename", h.RenameConversationById)
		convs.DELETE("/:id", h.DeleteConversationById)
		convs.POST("/:id/messages", h.AddMessageToConversation)
	}
}
