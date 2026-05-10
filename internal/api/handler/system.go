package handler

import (
	"ekken/internal/logger"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetSystemConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"app_name":    h.Config.AppName,
			"data_dir":    h.Config.DataDir,
			"plugin_dir":  h.Config.PluginDir,
			"port":        h.Config.Port,
			"app_version": h.Config.AppVersion,
			"mode":        h.Config.Mode,
			"repo_url":    h.Config.RepoURL,
			"author":      h.Config.Author,
		},
	})
}

func (h *Handler) RestartServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Server is restarting...",
	})

	go func() {
		logger.Info("Restarting server in 1 second...")
		time.Sleep(1 * time.Second)

		// Close database connection gracefully
		if h.DB != nil {
			logger.Info("Closing database connection...")
			h.DB.Close()
		}

		err := h.restartApp()
		if err != nil {
			logger.Error("Failed to restart server", "error", err)
			os.Exit(1)
		}
	}()
}
