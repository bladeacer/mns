package yaml_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/yaml"
)

func TestWriteYAML_MkdirAllError(t *testing.T) {
	dir := t.TempDir()
	// Create a file where the directory should be, causing MkdirAll to fail
	blockPath := filepath.Join(dir, "block")
	if err := os.WriteFile(blockPath, []byte("not-a-dir"), 0644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(blockPath, "config.yaml")

	cfg := config.GetMnemoConf()
	yaml.WriteYAML(cfg, configPath)
}

func TestWriteYAML_WriteFileError(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(configDir, 0555); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.yaml")

	cfg := config.GetMnemoConf()
	yaml.WriteYAML(cfg, configPath)
}

func TestSaveConfig_MkdirAllError(t *testing.T) {
	dir := t.TempDir()
	blockPath := filepath.Join(dir, "block")
	if err := os.WriteFile(blockPath, []byte("not-a-dir"), 0644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(blockPath, "config.yaml")

	cfg := config.GetMnemoConf()
	err := yaml.SaveConfig(cfg, configPath)
	if err == nil {
		t.Error("expected error when MkdirAll fails")
	}
}

func TestSaveConfig_WriteFileError(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(configDir, 0555); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.yaml")

	cfg := config.GetMnemoConf()
	err := yaml.SaveConfig(cfg, configPath)
	if err == nil {
		t.Error("expected error when WriteFile fails")
	}
}
