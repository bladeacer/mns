package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mns/cmd"
)

func TestValidateConfigAndDataStore_NoConfig(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	exitCode := cmd.ValidateConfigAndDataStore("")
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestValidateConfigAndDataStore_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	cmd.SaveConfig()
	exitCode := cmd.ValidateConfigAndDataStore(filepath.Join(dir, "config.yaml"))
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestValidateConfigAndDataStore_InvalidYaml(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: [bad"), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := cmd.ValidateConfigAndDataStore(configPath)
	if exitCode != 2 {
		t.Errorf("expected exit code 2, got %d", exitCode)
	}
}
