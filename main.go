package main

import (
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
	cfg := cmd.Execute()

	logger.NewCleanLogger(cfg.Verbose)

	database, err := db.Open(cfg.DataDir)
	if err != nil {
		logger.Error("Failed to open database", "error", err, "dir", cfg.DataDir)
		os.Exit(1)
	}
	defer database.Close()

	server := api.NewServer(cfg, database)
	engine := server.Engine()

	embed.ServeEmbedded(engine, cfg.Mode)

	if cfg.Mode == "production" {
		banner.PrintProd(cfg.AppVersion, cfg.Address)
	} else {
		banner.PrintDev(cfg.AppVersion, cfg.Address)
	}

	logger.Info("Server is running", "address", cfg.Address)
	if err := engine.Run(cfg.Address); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
