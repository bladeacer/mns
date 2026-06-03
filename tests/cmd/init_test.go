package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mns/cmd"
	"github.com/bladeacer/mns/config"
)

func TestValidateInitPreconditions_ConfigExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	if err := os.WriteFile(configPath, []byte("key: value"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := cmd.ValidateInitPreconditions(configPath, dbPath)
	if err == nil {
		t.Fatal("expected error when config file already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected error about config existing, got: %v", err)
	}
}

func TestValidateInitPreconditions_ConfigNotExist_DbExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	if err := os.WriteFile(dbPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	dbExists, err := cmd.ValidateInitPreconditions(configPath, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dbExists {
		t.Error("expected dbExists=true when db file exists")
	}
}

func TestValidateInitPreconditions_NeitherExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	dbExists, err := cmd.ValidateInitPreconditions(configPath, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dbExists {
		t.Error("expected dbExists=false when db file does not exist")
	}
}

func TestCompleteInitSetup_WithGitDir(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")
	gitDir := filepath.Join(dir, ".git")

	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath:      configPath,
			RepoPath:        dir,
			DbPath:          dbPath,
			IsInit:          true,
			Archiver:        "tar",
			HistLimitDays:   7,
			HistLimitSizeMb: 1024,
			KeepArchives:    5,
			LfsThresholdMb:  5,
		},
	}

	cmd.CompleteInitSetup(dir, configPath, dbPath, false, cfg)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected database file to be created")
	}
}

func TestCompleteInitSetup_WithoutGitDir(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath:      configPath,
			RepoPath:        dir,
			DbPath:          dbPath,
			IsInit:          true,
			Archiver:        "tar",
			HistLimitDays:   7,
			HistLimitSizeMb: 1024,
			KeepArchives:    5,
			LfsThresholdMb:  5,
		},
	}

	cmd.CompleteInitSetup(dir, configPath, dbPath, false, cfg)

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("expected no config file when .git missing")
	}
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Error("expected no db file when .git missing")
	}
}

func TestCompleteInitSetup_DbAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")
	gitDir := filepath.Join(dir, ".git")

	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath, []byte(`{"current_id": 0}`), 0644); err != nil {
		t.Fatal(err)
	}

	origModTime, err := getModTime(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath:      configPath,
			RepoPath:        dir,
			DbPath:          dbPath,
			IsInit:          true,
			Archiver:        "tar",
			HistLimitDays:   7,
			HistLimitSizeMb: 1024,
			KeepArchives:    5,
			LfsThresholdMb:  5,
		},
	}

	cmd.CompleteInitSetup(dir, configPath, dbPath, true, cfg)

	newModTime, err := getModTime(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if newModTime != origModTime {
		t.Error("expected existing db file to not be modified")
	}
}

func TestCompleteInitSetup_GitignoreCreated(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")
	gitDir := filepath.Join(dir, ".git")

	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath:      configPath,
			RepoPath:        dir,
			DbPath:          dbPath,
			IsInit:          true,
			Archiver:        "tar",
			HistLimitDays:   7,
			HistLimitSizeMb: 1024,
			KeepArchives:    5,
			LfsThresholdMb:  5,
		},
	}

	cmd.CompleteInitSetup(dir, configPath, dbPath, false, cfg)

	gitignoreContent, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gitignoreContent), "/.mnemosync/") {
		t.Errorf("expected .gitignore to contain '/.mnemosync/', got: '%s'", string(gitignoreContent))
	}
}

func getModTime(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.ModTime().UnixNano(), nil
}
