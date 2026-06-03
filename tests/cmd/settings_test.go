package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mns/cmd"
)

func TestSaveConfig_WritesFile(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	configPath := filepath.Join(dir, "config.yaml")
	_ = configPath

	cmd.SaveConfig()
	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}

func TestSaveConfig_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("config_schema:\n  app_version: \"0.0.0\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd.SaveConfig()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty config after save")
	}
}
