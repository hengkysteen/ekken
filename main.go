package main

import (
	"fmt"
	"os"

	"ekken/banner"
	"ekken/cmd"
	"ekken/embed"
	"ekken/internal/api"
	"ekken/internal/db"
	"ekken/internal/logger"

	_ "ekken/internal/features"
	_ "ekken/nodes"
)

func main() {
	// Initialize CLI and Load Config
	cfg := cmd.Execute()

	// Initialize logging
	logger.NewCleanLogger(cfg.Verbose)

	// Initialize database
	database, err := db.Open(cfg.DataDir)
	if err != nil {
		logger.Error("Failed to open database", "error", err, "dir", cfg.DataDir)
		os.Exit(1)
	}
	defer database.Close()

	// Create server (Modules are auto-registered via blank imports and initialized in NewServer)
	server := api.NewServer(cfg, database)
	engine := server.Engine()

	// Serve embedded UI
	embed.ServeEmbedded(engine, cfg.Mode)

	// Print mode-specific info and start server
	address := fmt.Sprintf(":%d", cfg.Port)

	if cfg.Mode == "production" {
		banner.PrintProd(cfg.AppVersion, cfg.Host, cfg.Port)
	} else {
		banner.PrintDev(cfg.AppVersion, cfg.Host, cfg.Port)
	}

	logger.Info("Server is running", "address", address)
	if err := engine.Run(address); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
