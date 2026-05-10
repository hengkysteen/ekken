package cmd

import (
	"ekken/internal/config"
	"flag"
	"fmt"
	"os"
)

// Execute parses CLI flags and returns the application configuration.
// It also handles immediate commands like --version.
func Execute() config.Config {
	showVersion := flag.Bool("version", false, "Print version")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	disableGinLog := flag.Bool("disable-gin-log", false, "Disable Gin's default logging")
	flag.Parse()

	cfg := config.LoadConfig()
	cfg.Verbose = cfg.Mode == "development" || *verbose
	cfg.DisableGinLog = *disableGinLog

	if *showVersion {
		fmt.Println(cfg.AppVersion)
		os.Exit(0)
	}

	return cfg
}
