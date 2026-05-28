package module

import (
	"ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/features/assistant"
	"ekken/internal/features/assistant/conversation"
	"ekken/internal/features/credential"
	"ekken/internal/logger"

	_ "ekken/internal/features/assistant/conversation/module"

	"github.com/gin-gonic/gin"
)

type AssistantModule struct {
	config        config.Config
	db            *db.DB
	repo          *assistant.Repository
	conversations conversation.Servicer
	history       *assistant.HistoryManager
	orchestrator  *assistant.Orchestrator
	jobs          *assistant.JobManager
	models        *assistant.ModelManager
	credentials   credential.Servicer
}

func NewModule() *AssistantModule {
	return &AssistantModule{}
}

func init() {
	module.RegisterModule(NewModule())
}

func (m *AssistantModule) Name() string {
	return "assistant"
}

func (m *AssistantModule) Init(database *db.DB, cfg config.Config) error {
	m.db = database
	m.config = cfg

	// Initialize Credential Service (Modular)
	credRepo, err := credential.NewRepository(database)
	if err != nil {
		return err
	}
	m.credentials = credential.New(credRepo)

	// Initialize Conversation Service
	convRepo := conversation.NewRepository(database)
	m.conversations = conversation.NewService(convRepo)
	m.history = assistant.NewHistoryManager(m.conversations)
	m.orchestrator = assistant.NewOrchestrator()
	m.jobs = assistant.NewJobManager()

	m.repo = assistant.NewRepository(database)

	// Initialize providers from DB
	dbProviders, err := m.repo.GetAssistantProviders()
	if err != nil {
		logger.Error("Failed to fetch assistant providers from db", "error", err)
	} else {
		for _, dp := range dbProviders {
			runtimeConfig, err := m.credentials.ResolveConfig(dp.Config)
			if err != nil {
				logger.Error("Failed to resolve credentials for provider", "provider", dp.ProviderID, "error", err)
				runtimeConfig = dp.Config
			}
			assistant.CreateProvider(dp.ProviderID, dp.Config, runtimeConfig)
		}
	}

	modelManager, err := assistant.NewModelManager(cfg.DataDir)
	if err != nil {
		logger.Error("Failed to initialize model manager", "error", err)
		return err
	}
	m.models = modelManager

	return nil
}

func (m *AssistantModule) RegisterRoutes(api *gin.RouterGroup) {
	h := &AssistantHandler{
		Config:        m.config,
		Conversations: m.conversations,
		History:       m.history,
		Orchestrator:  m.orchestrator,
		Jobs:          m.jobs,
		Models:        m.models,
		DB:            m.db,
		Repo:          m.repo,
		Credentials:   m.credentials,
	}

	assistantGroup := api.Group("/assistant")
	{
		// System provider catalogs
		assistantGroup.GET("/catalogs", h.Providers)

		// Agents
		assistantGroup.GET("/agents", h.ListAgents)
		assistantGroup.GET("/jobs", h.RunningJobs)

		// Models
		assistantGroup.POST("/models/sync", h.SyncModels)

		// Conversations
		assistantGroup.POST("/conversations/:id/stop", h.StopChat)
		assistantGroup.GET("/conversations/:id/job", h.ChatJob)
		assistantGroup.GET("/conversations/:id/stream", h.StreamChat)

		// Providers
		assistantGroup.GET("/providers", h.ListProviders)
		assistantGroup.POST("/providers/:id/chat", h.Chat)
		assistantGroup.POST("/providers/setup", h.SetupProvider)
		assistantGroup.DELETE("/providers/:id", h.DeleteProvider)
		assistantGroup.GET("/providers/:id/models", h.ProviderModels)
	}
}
