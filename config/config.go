package config

import (
	"github.com/bladeacer/mmsync/internal/fileio"
)

var AppVersion = "0.1.0"

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
	KeepArchives     int    `yaml:"keep_archives"`
	LfsThresholdMb   int64  `yaml:"lfs_threshold_mb"`
}

type MnemoConf struct {
	ConfigSchema ConfigSchema `yaml:"config_schema"`
}


func GetMnemoConf() *MnemoConf {
	return &MnemoConf{
		ConfigSchema{
			ConfigPath:       fileio.ResolveConfigPath(),
			AppVersion:       AppVersion,
			IsInit:           false,
			RepoPath:         "",
			DbPath:           fileio.ResolveDbPath(),
			Archiver:         "tar",
			CommitFmt:        "mnemosync archive 2006-01-02",
			RespectGitignore: true,
			HistLimitDays:    7,
			HistLimitSizeMb:  1024,
			KeepArchives:     5,
			LfsThresholdMb:   5,
		},
	}
}
