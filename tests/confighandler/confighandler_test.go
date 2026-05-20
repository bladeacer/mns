package confighandler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mmsync/internal/confighandler"
)

func withFakeHome(t *testing.T, fn func(homeDir string)) {
	t.Helper()
	fakeHome := t.TempDir()

	// Save real HOME and set a fake one so ResolveConfigPath
	// points into the temp dir with no pre-existing config.
	realHome := os.Getenv("HOME")
	os.Setenv("HOME", fakeHome)
	prevMMSync := os.Getenv("MMSYNC_CONF")
	os.Unsetenv("MMSYNC_CONF")

	defer func() {
		os.Setenv("HOME", realHome)
		os.Setenv("MMSYNC_CONF", prevMMSync)
	}()

	fn(fakeHome)
}

func writeConfig(t *testing.T, dir string, content string) {
	t.Helper()
	configPath := filepath.Join(dir, ".config/mmsync/config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		cfg, err := confighandler.LoadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.ConfigSchema.IsInit {
			t.Error("expected IsInit to be false for default config")
		}
	})
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.1.0"
  is_init: true
  repo_path: "` + homeDir + `"
  db_path: "` + filepath.Join(homeDir, ".config/mmsync/mmsync-state.json") + `"
  archiver: tar
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
  keep_archives: 5
  lfs_threshold_mb: 5
`
		writeConfig(t, homeDir, yamlContent)

		cfg, err := confighandler.LoadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.ConfigSchema.IsInit {
			t.Error("expected IsInit to be true")
		}
	})
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		writeConfig(t, homeDir, "{invalid yaml: [}")

		_, err := confighandler.LoadConfig()
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})
}

func TestLoadConfig_HealsOnVersionMismatch(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "9.9.9"
  is_init: true
  repo_path: "` + homeDir + `"
  db_path: "` + filepath.Join(homeDir, ".config/mmsync/mmsync-state.json") + `"
  archiver: tar
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
  keep_archives: 5
  lfs_threshold_mb: 5
`
		writeConfig(t, homeDir, yamlContent)

		cfg, err := confighandler.LoadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ConfigSchema.AppVersion != "0.1.0" {
			t.Errorf("expected AppVersion healed to '0.1.0', got '%s'", cfg.ConfigSchema.AppVersion)
		}
	})
}

func TestLoadConfig_HealsOnNotInit(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.1.0"
  is_init: false
  repo_path: "` + homeDir + `"
  db_path: "` + filepath.Join(homeDir, ".config/mmsync/mmsync-state.json") + `"
  archiver: tar
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
  keep_archives: 5
  lfs_threshold_mb: 5
`
		writeConfig(t, homeDir, yamlContent)

		cfg, err := confighandler.LoadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ConfigSchema.IsInit {
			t.Error("expected IsInit to remain false when not init")
		}
		if cfg.ConfigSchema.RepoPath != "" {
			t.Error("expected RepoPath to be cleared when not init")
		}
	})
}

func TestLoadConfig_HealsMissingFields(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.1.0"
  is_init: true
  repo_path: "` + homeDir + `"
  db_path: ""
  archiver: ""
  commit_fmt: ""
  respect_gitignore: true
  hist_limit_days: -1
  hist_limit_size_mb: -1
  keep_archives: 0
  lfs_threshold_mb: 0
`
		writeConfig(t, homeDir, yamlContent)

		cfg, err := confighandler.LoadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ConfigSchema.Archiver != "tar" {
			t.Errorf("expected Archiver healed to 'tar', got '%s'", cfg.ConfigSchema.Archiver)
		}
		if cfg.ConfigSchema.CommitFmt != "mnemosync archive 2006-01-02" {
			t.Errorf("expected CommitFmt healed to default, got '%s'", cfg.ConfigSchema.CommitFmt)
		}
		if cfg.ConfigSchema.HistLimitDays != 7 {
			t.Errorf("expected HistLimitDays healed to 7, got %d", cfg.ConfigSchema.HistLimitDays)
		}
		if cfg.ConfigSchema.HistLimitSizeMb != 1024 {
			t.Errorf("expected HistLimitSizeMb healed to 1024, got %d", cfg.ConfigSchema.HistLimitSizeMb)
		}
		if cfg.ConfigSchema.KeepArchives != 5 {
			t.Errorf("expected KeepArchives healed to 5, got %d", cfg.ConfigSchema.KeepArchives)
		}
		if cfg.ConfigSchema.LfsThresholdMb != 5 {
			t.Errorf("expected LfsThresholdMb healed to 5, got %d", cfg.ConfigSchema.LfsThresholdMb)
		}
	})
}

func TestLoadConfig_UpdatesSchema(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.1.0"
  is_init: true
  repo_path: "` + homeDir + `"
  db_path: "` + filepath.Join(homeDir, ".config/mmsync/mmsync-state.json") + `"
  archiver: tar
  commit_fmt: "mnemosync archive 2006-01-02"
  hist_limit_days: 7
  hist_limit_size_mb: 1024
`
		writeConfig(t, homeDir, yamlContent)

		cfg, err := confighandler.LoadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ConfigSchema.KeepArchives != 5 {
			t.Errorf("expected KeepArchives default 5, got %d", cfg.ConfigSchema.KeepArchives)
		}
		if cfg.ConfigSchema.LfsThresholdMb != 5 {
			t.Errorf("expected LfsThresholdMb default 5, got %d", cfg.ConfigSchema.LfsThresholdMb)
		}
	})
}

func TestLoadConfig_ReadError(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(configPath, []byte("valid: yaml"), 0000); err != nil {
			t.Fatal(err)
		}
		os.Chmod(configPath, 0000)
		defer os.Chmod(configPath, 0644)

		_, err := confighandler.LoadConfig()
		if err == nil {
			t.Error("expected error when config file is not readable")
		}
	})
}
