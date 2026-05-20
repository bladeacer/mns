package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mmsync/cmd"
	"github.com/bladeacer/mmsync/config"
)

func TestRunHealthCheck_AllMissing(t *testing.T) {
	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: filepath.Join(os.TempDir(), "mns-test-config.yaml"),
			RepoPath:   "",
			DbPath:     "",
		},
	}

	result := cmd.RunHealthCheck(cfg, false)

	if !strings.Contains(result, "[NOT FOUND] Configuration file") {
		t.Error("expected config file not found warning")
	}
	if !strings.Contains(result, "[NOT SET] Repository Path") {
		t.Error("expected repo path not set warning")
	}
	if !strings.Contains(result, "[NOT SET] Database Path") {
		t.Error("expected db path not set warning")
	}
}

func TestRunHealthCheck_AllValid(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")
	gitDir := filepath.Join(dir, ".git")

	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   dir,
			DbPath:     dbPath,
			IsInit:     true,
		},
	}

	result := cmd.RunHealthCheck(cfg, false)

	if strings.Contains(result, "[NOT FOUND]") {
		t.Errorf("unexpected 'not found' error: %s", result)
	}
	if strings.Contains(result, "[NOT SET]") {
		t.Errorf("unexpected 'not set' error: %s", result)
	}
	if strings.Contains(result, "[WARNING]") {
		t.Errorf("unexpected warning: %s", result)
	}
}

func TestRunHealthCheck_RepoPathNotExist(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   filepath.Join(dir, "nonexistent-repo"),
			DbPath:     dbPath,
			IsInit:     true,
		},
	}

	result := cmd.RunHealthCheck(cfg, false)

	if !strings.Contains(result, "[WARNING] Repository directory does not exist") {
		t.Errorf("expected repo path not exist warning, got: %s", result)
	}
}

func TestRunHealthCheck_RepoPathNotGitDir(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   dir,
			DbPath:     dbPath,
			IsInit:     true,
		},
	}

	result := cmd.RunHealthCheck(cfg, false)

	if !strings.Contains(result, "[WARNING] Repository path exists but is not a git repository") {
		t.Errorf("expected not a git repo warning, got: %s", result)
	}
}

func TestRunHealthCheck_RepoIsGitDir(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")
	gitDir := filepath.Join(dir, ".git")

	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   dir,
			DbPath:     dbPath,
			IsInit:     true,
		},
	}

	result := cmd.RunHealthCheck(cfg, false)

	if strings.Contains(result, "[WARNING]") {
		t.Errorf("unexpected warning for valid git repo: %s", result)
	}
}

func TestRunHealthCheck_DbPathNotExist(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	gitDir := filepath.Join(dir, ".git")

	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   dir,
			DbPath:     filepath.Join(dir, "nonexistent-db.json"),
			IsInit:     true,
		},
	}

	result := cmd.RunHealthCheck(cfg, false)

	if !strings.Contains(result, "[WARNING] Database file not found") {
		t.Errorf("expected db not found warning, got: %s", result)
	}
}

func TestRunHealthCheck_WithOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	dbPath := filepath.Join(dir, "state.json")

	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   dir,
			DbPath:     dbPath,
			IsInit:     true,
		},
	}

	result := cmd.RunHealthCheck(cfg, true)

	if strings.Contains(result, "[FAIL]") {
		if !strings.Contains(result, "zip") {
			t.Errorf("unexpected fail in output: %s", result)
		}
	}
}

func TestRunHealthCheck_ConfigFileExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   dir,
			DbPath:     filepath.Join(dir, "nonexistent-db.json"),
			IsInit:     true,
		},
	}

	result := cmd.RunHealthCheck(cfg, false)

	if strings.Contains(result, "[NOT FOUND] Configuration file") {
		t.Error("expected config file to be found")
	}
}

func TestCheckBinary_OptionalNotFound(t *testing.T) {
	result := cmd.CheckBinary("nonexistent-binary-xyz", true, false)
	if !strings.Contains(result, "[WARNING]") {
		t.Errorf("expected [WARNING] for missing optional binary, got: '%s'", result)
	}
}

func TestCheckBinary_WithOutput(t *testing.T) {
	result := cmd.CheckBinary("sh", false, true)
	if result != "" {
		t.Errorf("expected empty result for found binary, got: '%s'", result)
	}
}

func TestCheckBinary_NotFoundWithOutput(t *testing.T) {
	result := cmd.CheckBinary("nonexistent-binary-xyz", false, true)
	if !strings.Contains(result, "[FAIL]") {
		t.Errorf("expected [FAIL] for missing required binary, got: '%s'", result)
	}
}

func TestRunHealthCheck_DbPathSetButNotExist(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	gitDir := filepath.Join(dir, ".git")

	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   dir,
			DbPath:     filepath.Join(dir, "nonexistent-db.json"),
			IsInit:     true,
		},
	}

	result := cmd.RunHealthCheck(cfg, true)

	if !strings.Contains(result, "[WARNING] Database file not found") {
		t.Errorf("expected db not found warning when db path is set but file doesn't exist")
	}
}
