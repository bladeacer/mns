package fileio

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName           = "mmsync"
	LegacyConfigDir   = ".config/mmsync"
	DefaultConfigFile = "config.yaml"
	DefaultDbFile     = "mmsync-state.json"
)

func xdgConfigHome() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}

	userDir, err := os.UserConfigDir()
	if err == nil {
		return userDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".config"
	}
	return filepath.Join(home, ".config")
}

func xdgDataHome() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg
	}

	if runtime.GOOS != "linux" {
		// macOS: ~/Library/Application Support
		// Windows: %AppData%
		return xdgConfigHome()
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".local", "share")
	}
	return filepath.Join(home, ".local", "share")
}

func ResolveConfigPath() string {
	if envPath := os.Getenv("MMSYNC_CONF"); envPath != "" {
		resolvedPath := envPath

		if !filepath.IsAbs(envPath) {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return filepath.Join(AppName, DefaultConfigFile)
			}
			resolvedPath = filepath.Join(homeDir, envPath)
		}

		if filepath.Base(resolvedPath) != DefaultConfigFile {
			resolvedPath = filepath.Join(resolvedPath, DefaultConfigFile)
		}

		return resolvedPath
	}

	return filepath.Join(xdgConfigHome(), AppName, DefaultConfigFile)
}

func ResolveDbPath() string {
	if envPath := os.Getenv("MMSYNC_CONF"); envPath != "" {
		configPath := ResolveConfigPath()
		configDir := filepath.Dir(configPath)
		return filepath.Join(configDir, DefaultDbFile)
	}

	return filepath.Join(xdgDataHome(), AppName, DefaultDbFile)
}

func LegacyConfigDirPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(LegacyConfigDir)
	}
	return filepath.Join(home, LegacyConfigDir)
}

func LegacyDbPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(LegacyConfigDir, DefaultDbFile)
	}
	return filepath.Join(home, LegacyConfigDir, DefaultDbFile)
}

func ConfigDir() string {
	return filepath.Dir(ResolveConfigPath())
}

func DbDir() string {
	return filepath.Dir(ResolveDbPath())
}
