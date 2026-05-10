package module

import (
	"ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/features/workflow"

	"github.com/gin-gonic/gin"
)

type WorkflowModule struct {
	workflows workflow.WorkflowServicer
	runtime   workflow.RuntimeServicer
	sse       workflow.SSEServicer
}

func NewModule() *WorkflowModule {
	return &WorkflowModule{}
}

func init() {
	module.RegisterModule(NewModule())
}

func (m *WorkflowModule) Name() string {
	return "workflow"
}

func (m *WorkflowModule) Init(database *db.DB, cfg config.Config) error {
	database.AddColumnIfNotExists("workflows", "created_by", "TEXT")
	repo := workflow.NewRepository(database)
	sse := workflow.NewWorkflowEventStream()
	store := workflow.NewWorkflowStore(repo, cfg.DataDir)
	workflows := workflow.NewWorkflowService(store)
	runtime := workflow.NewRuntimeService(workflows, repo, sse, cfg.DataDir)

	m.workflows = workflows
	m.runtime = runtime
	m.sse = sse

	return nil
}

func (m *WorkflowModule) RegisterRoutes(api *gin.RouterGroup) {
	h := &WorkflowHandler{
		workflows: m.workflows,
		runtime:   m.runtime,
		sse:       m.sse,
	}

	// Workflows - collection
	api.GET("/workflows", h.ListWorkflows)
	api.GET("/workflows/status", h.WorkflowsStatus)
	api.POST("/workflows", h.CreateWorkflow)
	api.POST("/workflows/validate", h.ValidateWorkflow)
	api.DELETE("/workflows", h.DeleteAllWorkflows)
	api.POST("/workflows/run", h.RunWorkflowPayload)
	api.GET("/nodes/catalog", h.NodeCatalog)

	// Workflows - single resource
	wf := api.Group("/workflows/:id")
	{
		wf.GET("", h.GetWorkflow)
		wf.PUT("", h.UpdateWorkflow)
		wf.DELETE("", h.DeleteWorkflow)
		wf.POST("/run", h.RunWorkflow)
		wf.POST("/stop", h.StopWorkflow)
		wf.GET("/status", h.WorkflowStatus)
		wf.GET("/logs", h.WorkflowLogs)
		wf.DELETE("/logs", h.DeleteWorkflowLogs)
		wf.GET("/events", h.SSEStream)
	}
}
