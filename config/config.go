package config

import (
	"github.com/bladeacer/mmsync/internal/fileio"
)

type ConfigSchema struct {
	ConfigPath       string `yaml:"config_path"`
	AppVersion       string `yaml:"app_version"`
	IsInit           bool   `yaml:"is_init"`
	RepoPath         string `yaml:"repo_path"`
	DbPath           string `yaml:"db_path"`
	Archiver         string `yaml:"archiver"`
	CommitFmt        string `yaml:"commit_fmt"`
	RespectGitignore bool   `yaml:"respect_gitignore"`
	HistLimitDays    int    `yaml:"hist_limit_days"`
	HistLimitSizeMb  int64  `yaml:"hist_limit_size_mb"`
}

type MnemoConf struct {
	ConfigSchema ConfigSchema `yaml:"config_schema"`
}


func GetMnemoConf() *MnemoConf {
	return &MnemoConf{
		ConfigSchema{
			ConfigPath:       fileio.ResolveConfigPath(),
			AppVersion:       "Version 0.0.1",
			IsInit:           false,
			RepoPath:         "",
			DbPath:           fileio.ResolveDbPath(),
			Archiver:         "tar",
			CommitFmt:        "mnemosync archive 2006-01-02",
			RespectGitignore: true,
			HistLimitDays:    7,
			HistLimitSizeMb:  1024,
		},
	}
}
