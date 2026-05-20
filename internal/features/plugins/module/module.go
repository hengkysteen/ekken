package module

import (
	"ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/features/plugins"
	"ekken/internal/features/plugins/kind"
	"ekken/internal/features/plugins/passistant"
	"ekken/internal/features/plugins/pnode"

	"github.com/gin-gonic/gin"
)

type PluginModule struct {
	service plugins.PluginServicer
}

func NewModule() *PluginModule {
	return &PluginModule{}
}

func init() {
	// Register plugin kinds
	kind.Register("node", func(config kind.Config) kind.Kind {
		return pnode.NewKind(config.ExecTimeout)
	})
	kind.Register("assistant", func(config kind.Config) kind.Kind {
		return passistant.NewKind(config.ExecTimeout)
	})

	module.RegisterModule(NewModule())
}

func (m *PluginModule) Name() string {
	return "plugins"
}

func (m *PluginModule) Init(database *db.DB, cfg config.Config) error {
	service, err := plugins.NewPluginService(cfg.AppVersion, cfg.PluginDir)
	if err != nil {
		return err
	}
	m.service = service
	return nil
}

func (m *PluginModule) RegisterRoutes(api *gin.RouterGroup) {
	h := &PluginHandler{service: m.service}

	p := api.Group("/plugins")
	{
		p.GET("", h.ListPlugins)
		p.POST("/reload", h.ReloadPlugins)
		p.GET("/registry", h.PluginRegistry)
		p.POST("/registry/:id/install", h.InstallPlugin)
		p.GET("/registry/:id/install", h.PluginInstallStatus)
		p.DELETE("/registry/:id/install", h.StopPluginInstall)
		p.DELETE("/:id", h.UninstallPlugin)
		p.POST("/:id/:action", h.HandlePluginAction)
	}
}
