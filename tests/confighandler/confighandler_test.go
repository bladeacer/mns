package confighandler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/confighandler"
)

func withFakeHome(t *testing.T, fn func(homeDir string)) {
	t.Helper()
	fakeHome := t.TempDir()

	// Save real HOME and set a fake one so ResolveConfigPath
	// points into the temp dir with no pre-existing config.
	realHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", fakeHome)
	prevMMSync := os.Getenv("MMSYNC_CONF")
	_ = os.Unsetenv("MMSYNC_CONF")

	defer func() {
		_ = os.Setenv("HOME", realHome)
		_ = os.Setenv("MMSYNC_CONF", prevMMSync)
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

func TestLoadConfig_VersionMismatchUpdatesToBinaryVersion(t *testing.T) {
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
		// AppVersion should be updated to the current binary version in memory
		if cfg.ConfigSchema.AppVersion != "0.1.0" {
			t.Errorf("expected AppVersion updated to '0.1.0', got '%s'", cfg.ConfigSchema.AppVersion)
		}

		// AppVersion should also be persisted to disk
		dataAfter, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(dataAfter), "app_version: \"0.1.0\"") &&
			!strings.Contains(string(dataAfter), "app_version: 0.1.0") {
			t.Errorf("expected file on disk to contain updated app_version '0.1.0', got:\n%s", string(dataAfter))
		}
	})
}

func TestLoadConfig_NotInitPreservesPaths(t *testing.T) {
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
		// RepoPath should NOT be cleared — healing no longer resets paths on IsInit=false
		if cfg.ConfigSchema.RepoPath != homeDir {
			t.Errorf("expected RepoPath to be preserved, got '%s'", cfg.ConfigSchema.RepoPath)
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
		// 0 is a valid value for KeepArchives/LfsThresholdMb (means "disabled")
		if cfg.ConfigSchema.KeepArchives != 0 {
			t.Errorf("expected KeepArchives to remain 0, got %d", cfg.ConfigSchema.KeepArchives)
		}
		if cfg.ConfigSchema.LfsThresholdMb != 0 {
			t.Errorf("expected LfsThresholdMb to remain 0, got %d", cfg.ConfigSchema.LfsThresholdMb)
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

func TestLoadConfig_RepoPathNotExist(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.1.0"
  is_init: true
  repo_path: "` + filepath.Join(homeDir, "nonexistent-repo") + `"
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
		if cfg == nil {
			t.Fatal("expected non-nil config")
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
		_ = os.Chmod(configPath, 0000)
		defer func() { _ = os.Chmod(configPath, 0644) }()

		_, err := confighandler.LoadConfig()
		if err == nil {
			t.Error("expected error when config file is not readable")
		}
	})
}

func TestLoadConfig_HealingDoesNotBreakValidConfig(t *testing.T) {
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

		// Load config multiple times — healing should not corrupt the file
		var prevCfg *config.MnemoConf
		for i := 0; i < 5; i++ {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("iteration %d: unexpected error: %v", i, err)
			}
			if prevCfg != nil {
				if cfg.ConfigSchema.AppVersion != prevCfg.ConfigSchema.AppVersion {
					t.Errorf("iteration %d: AppVersion changed from '%s' to '%s'", i, prevCfg.ConfigSchema.AppVersion, cfg.ConfigSchema.AppVersion)
				}
				if cfg.ConfigSchema.IsInit != prevCfg.ConfigSchema.IsInit {
					t.Errorf("iteration %d: IsInit changed from %t to %t", i, prevCfg.ConfigSchema.IsInit, cfg.ConfigSchema.IsInit)
				}
				if cfg.ConfigSchema.RepoPath != prevCfg.ConfigSchema.RepoPath {
					t.Errorf("iteration %d: RepoPath changed from '%s' to '%s'", i, prevCfg.ConfigSchema.RepoPath, cfg.ConfigSchema.RepoPath)
				}
				if cfg.ConfigSchema.DbPath != prevCfg.ConfigSchema.DbPath {
					t.Errorf("iteration %d: DbPath changed from '%s' to '%s'", i, prevCfg.ConfigSchema.DbPath, cfg.ConfigSchema.DbPath)
				}
				if cfg.ConfigSchema.KeepArchives != prevCfg.ConfigSchema.KeepArchives {
					t.Errorf("iteration %d: KeepArchives changed from %d to %d", i, prevCfg.ConfigSchema.KeepArchives, cfg.ConfigSchema.KeepArchives)
				}
				if cfg.ConfigSchema.LfsThresholdMb != prevCfg.ConfigSchema.LfsThresholdMb {
					t.Errorf("iteration %d: LfsThresholdMb changed from %d to %d", i, prevCfg.ConfigSchema.LfsThresholdMb, cfg.ConfigSchema.LfsThresholdMb)
				}
			}
			prevCfg = cfg
		}
	})
}
