package config

import (
	"github.com/bladeacer/mmsync/internal/fileio"
)

type ConfigSchema struct {
	ConfigPath string `yaml:"config_path"`
	AppVersion string `yaml:"app_version"`
	IsInit     bool   `yaml:"is_init"`
	RepoPath   string `yaml:"repo_path"`
	DbPath     string `yaml:"db_path"`
}

type MnemoConf struct {
	ConfigSchema ConfigSchema `yaml:"config_schema"`
}


func GetMnemoConf() *MnemoConf {
	return &MnemoConf{
		ConfigSchema{
			ConfigPath: fileio.ResolveConfigPath(),
			AppVersion: "Version 0.0.1",
			IsInit:     false,
			RepoPath:   "",
			DbPath:     fileio.ResolveDbPath(),
		},
	}
}
