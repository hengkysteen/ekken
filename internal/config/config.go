package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	AppName    string
	Host       string
	Port       int
	DataDir    string
	AppVersion string
	PluginDir  string
	RepoURL    string
	Author     string
	Mode       string

	// cmd flag
	Verbose       bool
	DisableGinLog bool
}

// Do not edit this manually. Use Makefile for production builds.
var mode = "development"

// Source of Truth for Ekken versioning (Format: vYYYY.m.DD).
// After updating this version, use 'make tag' to create a new git tag.
var buildVersion = "v2026.5.10-alpha"

func LoadConfig() Config {
	dataDir := getEnv("EKKENDATA_DIR", defaultDataDir())
	return Config{
		AppName:    "Ekken",
		Host:       getEnv("EKKENAPI_HOST", "localhost"),
		Port:       getEnvInt("EKKENAPI_PORT", 11245),
		DataDir:    dataDir,
		AppVersion: buildVersion,
		Mode:       mode,
		RepoURL:    "https://github.com/hengkysteen/ekken",
		Author:     "hengkysteen",
	}
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".ekken")
	}
	return filepath.Join(home, ".ekken")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
