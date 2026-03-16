package fileio

import (
	"os"
	"path/filepath"
)

const (
	DefaultConfigDir  = ".config/mmsync"
	DefaultConfigFile = "config.yaml"
	DefaultDbFile     = "mmsync-state.json"
)

func ResolveConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(DefaultConfigDir, DefaultConfigFile)
	}

	if envPath := os.Getenv("MMSYNC_CONF"); envPath != "" {
		resolvedPath := envPath

		if !filepath.IsAbs(envPath) {
			resolvedPath = filepath.Join(homeDir, envPath)
		}

		if filepath.Base(resolvedPath) != DefaultConfigFile {
			resolvedPath = filepath.Join(resolvedPath, DefaultConfigFile)
		}

		return resolvedPath
	}

	return filepath.Join(homeDir, DefaultConfigDir, DefaultConfigFile)
}

func ResolveDbPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(DefaultConfigDir, DefaultDbFile)
	}

	if envPath := os.Getenv("MMSYNC_CONF"); envPath != "" {
		configPath := ResolveConfigPath()

		configDir := filepath.Dir(configPath)
		return filepath.Join(configDir, DefaultDbFile)
	}

	return filepath.Join(homeDir, DefaultConfigDir, DefaultDbFile)
}
