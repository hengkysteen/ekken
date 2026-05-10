package api

import (
	"ekken/internal/api/handler"
	"ekken/internal/api/module"
	"ekken/internal/api/router"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router  *router.Router
	cfg     config.Config
	db      *db.DB
	modules []module.Module
}

func NewServer(cfg config.Config, database *db.DB) *Server {
	h := handler.New(cfg, database)

	s := &Server{
		router:  router.New(h),
		cfg:     cfg,
		db:      database,
		modules: make([]module.Module, 0),
	}

	// Auto-initialize modules from registry
	for _, m := range module.ModuleRegistry {
		if err := m.Init(database, cfg); err != nil {
			logger.Error("Failed to initialize module", "module", m.Name(), "error", err)
			continue
		}
		s.modules = append(s.modules, m)
	}

	return s
}

func (s *Server) Engine() *gin.Engine {
	if s.cfg.Verbose && !s.cfg.DisableGinLog {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.New()
	if s.cfg.Verbose && !s.cfg.DisableGinLog {
		e.Use(gin.Logger())
	}
	e.Use(gin.Recovery())
	s.router.Setup(e, s.modules)
	return e
}
