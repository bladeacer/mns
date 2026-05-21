package confighandler_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/confighandler"
	"github.com/bladeacer/mns/internal/fileio"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	realStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = realStderr
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestLoadConfig_NotInitDoesNotSaveToDisk(t *testing.T) {
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

		statBefore, err := os.Stat(configPath)
		if err != nil {
			t.Fatal(err)
		}

		out := captureStderr(t, func() {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}
		})

		statAfter, err := os.Stat(configPath)
		if err != nil {
			t.Fatal(err)
		}

		if statBefore.ModTime() != statAfter.ModTime() {
			t.Error("config file was modified on disk when IsInit=false and nothing was healed")
		}

		if strings.Contains(out, "Saving Repaired Configuration") {
			t.Error("unexpected save triggered for IsInit=false with valid config")
		}

		if !strings.Contains(out, "configuration is not initialized") {
			t.Error("expected warning about not initialized")
		}
	})
}

func TestLoadConfig_NotInitPreservesValuesOnDisk(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		originalContent := `config_schema:
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
		writeConfig(t, homeDir, originalContent)

		_, err := confighandler.LoadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		contentAfter, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(contentAfter) != originalContent {
			t.Errorf("file content changed after loading IsInit=false config\n--- original\n+++ after\n-%s\n+%s", originalContent, string(contentAfter))
		}
	})
}

func TestLoadConfig_ValidInitDoesNotSaveToDisk(t *testing.T) {
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

		statBefore, err := os.Stat(configPath)
		if err != nil {
			t.Fatal(err)
		}

		out := captureStderr(t, func() {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}
		})

		statAfter, err := os.Stat(configPath)
		if err != nil {
			t.Fatal(err)
		}

		if statBefore.ModTime() != statAfter.ModTime() {
			t.Error("config file was modified on disk for valid IsInit=true config")
		}

		if strings.Contains(out, "Saving Repaired Configuration") {
			t.Error("unexpected save triggered for valid IsInit=true config")
		}
	})
}

func TestLoadConfig_MultipleLoadsDontChangeFile(t *testing.T) {
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

		originalData, _ := os.ReadFile(configPath)

		for i := 0; i < 10; i++ {
			_, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("iteration %d: unexpected error: %v", i, err)
			}
		}

		dataAfter, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(dataAfter) != string(originalData) {
			t.Error("file content changed after 10 successive load calls")
		}
	})
}

func TestLoadConfig_HealedFieldsPersistToDisk(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.1.0"
  is_init: true
  repo_path: "` + homeDir + `"
  db_path: ""
  archiver: ""
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
  keep_archives: 5
  lfs_threshold_mb: 5
`
		writeConfig(t, homeDir, yamlContent)

		out := captureStderr(t, func() {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}
		})

		if !strings.Contains(out, "Saving Repaired Configuration") {
			t.Error("expected save for healed config")
		}

		dataAfter, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(string(dataAfter), "db_path: \"\"") {
			t.Error("DbPath was not healed on disk - still empty")
		}
		if strings.Contains(string(dataAfter), "archiver: \"\"") {
			t.Error("Archiver was not healed on disk - still empty")
		}
	})
}

func TestLoadConfig_IsInitFalseDoesNotRepairRepoPathOnDisk(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		customRepoPath := filepath.Join(homeDir, "my-repo")
		_ = os.MkdirAll(customRepoPath, 0755)

		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.1.0"
  is_init: false
  repo_path: "` + customRepoPath + `"
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
		if cfg.ConfigSchema.RepoPath != customRepoPath {
			t.Errorf("expected RepoPath preserved as '%s', got '%s'", customRepoPath, cfg.ConfigSchema.RepoPath)
		}
	})
}

func TestLoadConfig_MissingKeysGetAddedBySchemaUpdate(t *testing.T) {
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
`
		writeConfig(t, homeDir, yamlContent)

		out := captureStderr(t, func() {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}
		})

		if !strings.Contains(out, "Updating configuration file with new schema fields") {
			t.Error("expected schema update message for config with missing fields")
		}

		cfg, err := confighandler.LoadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.ConfigSchema.KeepArchives != 5 {
			t.Errorf("expected KeepArchives default 5, got %d", cfg.ConfigSchema.KeepArchives)
		}
		if cfg.ConfigSchema.LfsThresholdMb != 5 {
			t.Errorf("expected LfsThresholdMb default 5, got %d", cfg.ConfigSchema.LfsThresholdMb)
		}
	})
}

func TestHealAndSaveConfig_AlwaysSaves(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")

		_ = os.MkdirAll(filepath.Dir(configPath), 0755)
		initialContent := `config_schema:
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
		if err := os.WriteFile(configPath, []byte(initialContent), 0644); err != nil {
			t.Fatal(err)
		}

		statBefore, _ := os.Stat(configPath)

		cfg, err := confighandler.HealAndSaveConfig(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}

		statAfter, _ := os.Stat(configPath)
		if statBefore.ModTime() == statAfter.ModTime() {
			t.Error("expected HealAndSaveConfig to modify file on disk")
		}
	})
}

func TestAtomicWriteFile_PersistsData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	data := []byte("key: value\n")

	if err := fileio.AtomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(readBack) != string(data) {
		t.Errorf("read back data differs: got '%s', want '%s'", string(readBack), string(data))
	}
}

func TestAtomicWriteFile_OverwritesExistingContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	if err := os.WriteFile(path, []byte("old: data"), 0644); err != nil {
		t.Fatal(err)
	}

	newData := []byte("new: content\n")
	if err := fileio.AtomicWriteFile(path, newData, 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(readBack) != string(newData) {
		t.Errorf("expected old content overwritten, got '%s'", string(readBack))
	}
}

func TestAtomicWriteFile_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "config.yaml")
	data := []byte("test: data\n")

	if err := fileio.AtomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to be created in subdirectory")
	}
}

func TestAtomicWriteFile_ValidPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	data := []byte("test: data\n")

	if err := fileio.AtomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("expected permissions 0644, got %o", info.Mode().Perm())
	}
}

func TestGetMnemoConf_DefaultsDoNotMutate(t *testing.T) {
	cfg1 := config.GetMnemoConf()
	cfg2 := config.GetMnemoConf()

	cfg1.ConfigSchema.IsInit = true
	cfg1.ConfigSchema.RepoPath = "/some/path"

	if cfg2.ConfigSchema.IsInit {
		t.Error("GetMnemoConf returned mutated config - IsInit should be false")
	}
	if cfg2.ConfigSchema.RepoPath != "" {
		t.Error("GetMnemoConf returned mutated config - RepoPath should be empty")
	}
}

func TestLoadConfigWithPath_InvalidYAMLReturnsError(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		_ = os.MkdirAll(filepath.Dir(configPath), 0755)
		_ = os.WriteFile(configPath, []byte("{invalid: [broken}"), 0644)

		_, err := confighandler.LoadConfig()
		if err == nil {
			t.Fatal("expected error for invalid YAML")
		}
	})
}

func TestLoadConfigWithPath_NoFileReturnsDefaults(t *testing.T) {
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

func TestLoadConfig_VersionUpdatedPersistsToDisk(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		// Old version 0.4.0 in config, binary has 0.1.0
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.4.0"
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

		out := captureStderr(t, func() {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.ConfigSchema.AppVersion != "0.1.0" {
				t.Errorf("expected in-memory version '0.1.0', got '%s'", cfg.ConfigSchema.AppVersion)
			}
		})

		if !strings.Contains(out, "AppVersion updated from '0.4.0' to '0.1.0'") {
			t.Error("expected version update message")
		}
		if !strings.Contains(out, "Updating configuration file with current version") {
			t.Error("expected file update message for version change")
		}

		dataAfter, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(dataAfter), "app_version: 0.1.0") {
			t.Errorf("version not persisted to disk. File content:\n%s", string(dataAfter))
		}
	})
}

func TestLoadConfig_VersionUpdatedOnlyOnceAfterPersist(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.4.0"
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

		// First load updates and persists version
		out1 := captureStderr(t, func() {
			_, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatal(err)
			}
		})
		if !strings.Contains(out1, "Updating configuration file with current version") {
			t.Error("first load should update version on disk")
		}

		// Second load should see matching version and NOT trigger another save
		out2 := captureStderr(t, func() {
			_, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatal(err)
			}
		})
		if strings.Contains(out2, "Updating configuration file with current version") {
			t.Error("second load should NOT update version - it was already persisted")
		}
		if strings.Contains(out2, "AppVersion updated") {
			t.Error("second load should NOT show version mismatch warning")
		}
	})
}

func TestLoadConfig_VersionUpdatedEvenWhenNotInit(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.4.0"
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

		out := captureStderr(t, func() {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.ConfigSchema.AppVersion != "0.1.0" {
				t.Errorf("expected in-memory version '0.1.0', got '%s'", cfg.ConfigSchema.AppVersion)
			}
		})

		if !strings.Contains(out, "AppVersion updated from '0.4.0' to '0.1.0'") {
			t.Error("expected version update message")
		}
		if !strings.Contains(out, "Updating configuration file with current version") {
			t.Error("expected file update for version change even when IsInit=false")
		}

		dataAfter, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(dataAfter), "app_version: \"0.4.0\"") {
			t.Error("version was NOT updated on disk for IsInit=false config")
		}
	})
}

func TestLoadConfig_HealingAndVersionUpdateOnlyOneSave(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		// Config with old version AND empty fields that need healing
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.4.0"
  is_init: true
  repo_path: "` + homeDir + `"
  db_path: ""
  archiver: ""
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
  keep_archives: 5
  lfs_threshold_mb: 5
`
		writeConfig(t, homeDir, yamlContent)

		out := captureStderr(t, func() {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.ConfigSchema.AppVersion != "0.1.0" {
				t.Errorf("expected version '0.1.0', got '%s'", cfg.ConfigSchema.AppVersion)
			}
		})

		if !strings.Contains(out, "Saving Repaired Configuration") {
			t.Error("expected healing save")
		}
		// Should NOT also have a version update save - healing save already covers it
		if strings.Contains(out, "Updating configuration file with current version") {
			t.Error("should not have separate version update save when healing already saved")
		}
	})
}

func TestLoadConfig_SchemaUpdateAndVersionUpdateOnlyOneSave(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := filepath.Join(homeDir, ".config/mmsync/config.yaml")
		// Config with old version AND missing schema fields
		yamlContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.4.0"
  is_init: true
  repo_path: "` + homeDir + `"
  db_path: "` + filepath.Join(homeDir, ".config/mmsync/mmsync-state.json") + `"
  archiver: tar
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
`
		writeConfig(t, homeDir, yamlContent)

		out := captureStderr(t, func() {
			cfg, err := confighandler.LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.ConfigSchema.AppVersion != "0.1.0" {
				t.Errorf("expected version '0.1.0', got '%s'", cfg.ConfigSchema.AppVersion)
			}
		})

		if !strings.Contains(out, "Updating configuration file with new schema fields") {
			t.Error("expected schema update save")
		}
		if strings.Contains(out, "Updating configuration file with current version") {
			t.Error("should not have separate version update save when schema update already saved")
		}
	})
}
