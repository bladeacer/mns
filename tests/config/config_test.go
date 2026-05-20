package config_test

import (
	"testing"

	"github.com/bladeacer/mmsync/config"
)

func TestGetMnemoConf_ReturnsValidConfig(t *testing.T) {
	cfg := config.GetMnemoConf()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.ConfigSchema.Archiver != "tar" {
		t.Errorf("expected default archiver 'tar', got '%s'", cfg.ConfigSchema.Archiver)
	}
	if cfg.ConfigSchema.RespectGitignore != true {
		t.Error("expected RespectGitignore to default to true")
	}
	if cfg.ConfigSchema.HistLimitDays != 7 {
		t.Errorf("expected HistLimitDays 7, got %d", cfg.ConfigSchema.HistLimitDays)
	}
	if cfg.ConfigSchema.HistLimitSizeMb != 1024 {
		t.Errorf("expected HistLimitSizeMb 1024, got %d", cfg.ConfigSchema.HistLimitSizeMb)
	}
	if cfg.ConfigSchema.KeepArchives != 5 {
		t.Errorf("expected KeepArchives 5, got %d", cfg.ConfigSchema.KeepArchives)
	}
	if cfg.ConfigSchema.LfsThresholdMb != 5 {
		t.Errorf("expected LfsThresholdMb 5, got %d", cfg.ConfigSchema.LfsThresholdMb)
	}
	if cfg.ConfigSchema.AppVersion != "0.1.0" {
		t.Errorf("expected AppVersion '0.1.0', got '%s'", cfg.ConfigSchema.AppVersion)
	}
	if cfg.ConfigSchema.IsInit != false {
		t.Error("expected IsInit to default to false")
	}
	if cfg.ConfigSchema.ConfigPath == "" {
		t.Error("expected non-empty ConfigPath")
	}
	if cfg.ConfigSchema.DbPath == "" {
		t.Error("expected non-empty DbPath")
	}
}
