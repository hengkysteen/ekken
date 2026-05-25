package webhook

import (
	apimodule "ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"

	"github.com/gin-gonic/gin"
)

type Module struct{}

func init() {
	apimodule.RegisterModule(&Module{})
}

func (m *Module) Name() string {
	return "webhook"
}

func (m *Module) Init(database *db.DB, cfg config.Config) error {
	return nil
}

func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	h := &Handler{}
	api.Any("/webhook/:id", h.Handle)
}
