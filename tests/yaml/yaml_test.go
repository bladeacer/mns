package yaml_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/yaml"
)

func TestWriteYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			AppVersion: "0.1.0",
			IsInit:     true,
			RepoPath:   dir,
			DbPath:     filepath.Join(dir, "state.json"),
			Archiver:   "tar",
		},
	}

	yaml.WriteYAML(cfg, configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected config file to be created")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "is_init: true") {
		t.Error("expected config to contain 'is_init: true'")
	}
}

func TestWriteYAML_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "subdir", "config.yaml")

	cfg := config.GetMnemoConf()
	yaml.WriteYAML(cfg, configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to be created in subdirectory")
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			IsInit:     true,
			RepoPath:   dir,
		},
	}

	err := yaml.SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected config file to be saved")
	}

	content, _ := os.ReadFile(configPath)
	if !strings.Contains(string(content), "is_init: true") {
		t.Error("expected saved config to contain 'is_init: true'")
	}
}

func TestSaveConfig_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "deep", "nested", "config.yaml")

	cfg := config.GetMnemoConf()
	err := yaml.SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveConfig_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	os.WriteFile(configPath, []byte("old: data"), 0644)

	cfg := config.GetMnemoConf()
	cfg.ConfigSchema.IsInit = true
	err := yaml.SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(configPath)
	if strings.Contains(string(content), "old: data") {
		t.Error("expected old content to be overwritten")
	}
}
