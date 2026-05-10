package module

import (
	"ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/features/mynode"

	"github.com/gin-gonic/gin"
)

type MyNodeModule struct {
	service mynode.MyNodesServicer
}

func NewModule() *MyNodeModule {
	return &MyNodeModule{}
}

func init() {
	module.RegisterModule(NewModule())
}

func (m *MyNodeModule) Name() string {
	return "mynode"
}

func (m *MyNodeModule) Init(database *db.DB, cfg config.Config) error {
	repo := mynode.NewMyNodesRepository(database)
	m.service = mynode.NewMyNodesService(repo)
	return nil
}

func (m *MyNodeModule) RegisterRoutes(api *gin.RouterGroup) {
	h := &MyNodeHandler{service: m.service}

	nodes := api.Group("/mynodes")
	{
		nodes.GET("", h.ListMyNodes)
		nodes.POST("", h.SaveMyNodes)
		nodes.PUT("/:id", h.UpdateMyNodes)
		nodes.DELETE("/:id", h.DeleteMyNodes)
	}
}
