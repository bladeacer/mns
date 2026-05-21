package yaml_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/fileio"
	"github.com/bladeacer/mns/internal/yaml"
)

func TestMergeAndSaveConfig_PreservesExtraFields(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	originalContent := `# user comment
extra_top_key: preserve_me
config_schema:
  config_path: "` + configPath + `"
  app_version: "` + config.AppVersion + `"
  is_init: true
  repo_path: "` + dir + `"
  db_path: "` + filepath.Join(dir, "state.json") + `"
  archiver: tar
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
  keep_archives: 5
  lfs_threshold_mb: 5
  extra_schema_field: should_survive
`
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath:       configPath,
			AppVersion:       config.AppVersion,
			IsInit:           true,
			RepoPath:         dir,
			DbPath:           filepath.Join(dir, "state.json"),
			Archiver:         "zip",
			CommitFmt:        "mnemosync archive 2006-01-02",
			RespectGitignore: true,
			HistLimitDays:    7,
			HistLimitSizeMb:  1024,
			KeepArchives:     5,
			LfsThresholdMb:   5,
		},
	}

	data, _ := os.ReadFile(configPath)
	if err := yaml.MergeAndSaveConfig(cfg, configPath, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	saved, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(saved)

	if !strings.Contains(content, "extra_top_key: preserve_me") {
		t.Error("MergeAndSaveConfig dropped top-level extra field")
	}
	if !strings.Contains(content, "extra_schema_field: should_survive") {
		t.Error("MergeAndSaveConfig dropped schema-level extra field")
	}
	if !strings.Contains(content, "archiver: zip") {
		t.Error("MergeAndSaveConfig did not update field to new value")
	}
}

func TestSaveConfig_DropsExtraFields(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	originalContent := `extra_top_key: will_be_lost
config_schema:
  config_path: "` + configPath + `"
  app_version: "` + config.AppVersion + `"
  is_init: true
  repo_path: "` + dir + `"
  db_path: "` + filepath.Join(dir, "state.json") + `"
  archiver: tar
  commit_fmt: "mnemosync archive 2006-01-02"
  respect_gitignore: true
  hist_limit_days: 7
  hist_limit_size_mb: 1024
  keep_archives: 5
  lfs_threshold_mb: 5
`
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.GetMnemoConf()
	cfg.ConfigSchema.IsInit = true
	cfg.ConfigSchema.RepoPath = dir
	if err := yaml.SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	saved, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(saved)

	if strings.Contains(content, "extra_top_key") {
		t.Error("SaveConfig should drop extra fields, but they appeared")
	}
}

func TestAtomicWriteFile_WithDirSync(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "test.yaml")
	data := []byte("key: value\n")

	if err := fileio.AtomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(readBack) != string(data) {
		t.Errorf("content mismatch: got '%s', want '%s'", string(readBack), string(data))
	}
}

func TestAtomicWriteFile_NoTempFilesLeftBehind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clean.yaml")
	data := []byte("test: data\n")

	if err := fileio.AtomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestAtomicWriteFile_MultipleWritesSamePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.yaml")

	for i := 0; i < 10; i++ {
		data := []byte("iteration: " + string(rune('0'+i)) + "\n")
		if err := fileio.AtomicWriteFile(path, data, 0644); err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
	}

	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readBack), "iteration: 9") {
		t.Errorf("expected final iteration content, got: %s", string(readBack))
	}
}
