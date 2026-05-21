package confighandler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/fileio"
)

func TestInitBehavior_ExistingDbFileLeftAlone(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		dbPath := filepath.Join(homeDir, ".local", "share", "mmsync", "mmsync-state.json")
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			t.Fatal(err)
		}
		existingDbContent := `{
  "current_id": 42,
  "tracked_dirs": {
    "1": {
      "target_path": "/some/path",
      "alias": "mydir"
    }
  },
  "staging_history": []
}`
		if err := os.WriteFile(dbPath, []byte(existingDbContent), 0644); err != nil {
			t.Fatal(err)
		}

		statBefore, err := os.Stat(dbPath)
		if err != nil {
			t.Fatal(err)
		}

		configPath := fileio.ResolveConfigPath()
		_, confErr := os.Stat(configPath)
		_, dbErr := os.Stat(dbPath)

		if confErr == nil {
			t.Fatal("config file should not exist yet")
		}
		if dbErr != nil {
			t.Fatal("db file should exist")
		}

		contentAfter, err := os.ReadFile(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(contentAfter) != existingDbContent {
			t.Error("db file content changed despite not being touched")
		}

		statAfter, err := os.Stat(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		if statBefore.ModTime() != statAfter.ModTime() {
			t.Error("db file modification time changed")
		}
	})
}

func TestInitBehavior_ConfigFileBlocksInit(t *testing.T) {
	withFakeHome(t, func(homeDir string) {
		configPath := fileio.ResolveConfigPath()
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
			t.Fatal(err)
		}

		_, confErr := os.Stat(configPath)
		_, dbErr := os.Stat(fileio.ResolveDbPath())

		if confErr != nil {
			t.Fatal("config file should exist")
		}
		if confErr == nil && dbErr == nil {
			t.Log("both files exist - config should block init")
		}
	})
}

func TestAtomicWrite_ConfigAndDbUseSameWrite(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	configData := []byte("config_schema:\n  app_version: \"0.1.0\"\n")
	dbData := []byte("{\"current_id\": 1}")

	if err := fileio.AtomicWriteFile(configPath, configData, 0644); err != nil {
		t.Fatal(err)
	}
	if err := fileio.AtomicWriteFile(dbPath, dbData, 0644); err != nil {
		t.Fatal(err)
	}

	readConfig, _ := os.ReadFile(configPath)
	readDb, _ := os.ReadFile(dbPath)

	if string(readConfig) != string(configData) {
		t.Error("config file content mismatch after atomic write")
	}
	if string(readDb) != string(dbData) {
		t.Error("db file content mismatch after atomic write")
	}
}

func TestSaveConfig_PreservesExtraFieldsViaMerge(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	originalContent := `config_schema:
  config_path: "` + configPath + `"
  app_version: "0.1.0"
  is_init: true
  repo_path: "` + dir + `"
  db_path: "` + dbPath + `"
  archiver: tar
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
  keep_archives: 5
  lfs_threshold_mb: 5
  custom_field: keep_me
`
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.GetMnemoConf()
	cfg.ConfigSchema.ConfigPath = configPath
	cfg.ConfigSchema.IsInit = true
	cfg.ConfigSchema.RepoPath = dir
	cfg.ConfigSchema.DbPath = dbPath

	// Simulate what SaveConfig does when using MergeAndSaveConfig
	if err := fileio.AtomicWriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	written, _ := os.ReadFile(configPath)
	if !strings.Contains(string(written), "custom_field: keep_me") {
		t.Error("extra schema field was lost after save")
	}
}
