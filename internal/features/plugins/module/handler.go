package module

import (
	"net/http"

	"ekken/internal/api"
	"ekken/internal/features/plugins"

	"github.com/gin-gonic/gin"
)

type PluginHandler struct {
	service plugins.PluginServicer
}

func (h *PluginHandler) ListPlugins(c *gin.Context) {
	c.JSON(http.StatusOK, api.Response{OK: true, Data: h.service.List()})
}

func (h *PluginHandler) ReloadPlugins(c *gin.Context) {
	if err := h.service.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: gin.H{"reloaded": true}})
}

func (h *PluginHandler) PluginRegistry(c *gin.Context) {
	registry, err := h.service.Registry(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: registry})
}

func (h *PluginHandler) InstallPlugin(c *gin.Context) {
	task, err := h.service.Install(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, api.Response{OK: true, Data: task})
}

func (h *PluginHandler) PluginInstallStatus(c *gin.Context) {
	task, ok := h.service.InstallStatus(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, api.Response{OK: false, Error: "install task not found"})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: task})
}

func (h *PluginHandler) StopPluginInstall(c *gin.Context) {
	task, err := h.service.StopInstall(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: task})
}

func (h *PluginHandler) UninstallPlugin(c *gin.Context) {
	if err := h.service.Uninstall(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: gin.H{"uninstalled": true}})
}

func (h *PluginHandler) HandlePluginAction(c *gin.Context) {
	id := c.Param("id")
	action := c.Param("action")

	if err := h.service.Manage(id, action); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: gin.H{"action": action, "status": "executed"}})
}
