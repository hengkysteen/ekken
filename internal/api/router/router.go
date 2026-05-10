package router

import (
	"ekken/internal/api/handler"
	"ekken/internal/api/module"

	"github.com/gin-gonic/gin"
)

type Router struct {
	h *handler.Handler
}

func New(h *handler.Handler) *Router {
	return &Router{h: h}
}

func (r *Router) Setup(engine *gin.Engine, modules []module.Module) {
	api := engine.Group("/api")
	{
		// Base API check
		api.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"data": "menyala abangku 🔥"})
		})

		// Register modular routes
		for _, m := range modules {
			m.RegisterRoutes(api)
		}

		// System
		api.GET("/system/config", r.h.GetSystemConfig)
		api.POST("/system/restart", r.h.RestartServer)
		api.GET("/system/device", r.h.GetDeviceInfo)
		api.POST("/system/file-picker", r.h.OpenFilePicker)
	}
}
