package fileio_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mns/internal/fileio"
)

func TestResolveConfigPath_HomeDirErr(t *testing.T) {
	realHome := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() { _ = os.Setenv("HOME", realHome) }()

	path := fileio.ResolveConfigPath()
	if path == "" {
		t.Error("expected fallback path when HOME is unset")
	}
}

func TestResolveConfigPath_Default(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Unsetenv("MMSYNC_CONF")
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	path := fileio.ResolveConfigPath()
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if !strings.HasSuffix(path, "config.yaml") {
		t.Errorf("expected path to end with 'config.yaml', got '%s'", path)
	}
	if !strings.HasSuffix(path, "mmsync/config.yaml") {
		t.Errorf("expected path to end with 'mmsync/config.yaml', got '%s'", path)
	}
}

func TestResolveConfigPath_WithMMSYNCConf(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", "/custom/path")
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	path := fileio.ResolveConfigPath()
	if !strings.HasPrefix(path, "/custom/path") {
		t.Errorf("expected path to start with '/custom/path', got '%s'", path)
	}
	if !strings.HasSuffix(path, "config.yaml") {
		t.Errorf("expected path to end with 'config.yaml', got '%s'", path)
	}
}

func TestResolveConfigPath_WithRelativeMMSYNCConf(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", "relative/path")
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	path := fileio.ResolveConfigPath()
	expected := filepath.Join(homeDir, "relative/path", "config.yaml")
	if path != expected {
		t.Errorf("expected '%s', got '%s'", expected, path)
	}
}

func TestResolveConfigPath_WithMMSYNCConfFile(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", "/custom/path/config.yaml")
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	path := fileio.ResolveConfigPath()
	if path != "/custom/path/config.yaml" {
		t.Errorf("expected '/custom/path/config.yaml', got '%s'", path)
	}
}

func TestResolveDbPath_HomeDirErr(t *testing.T) {
	realHome := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() { _ = os.Setenv("HOME", realHome) }()

	path := fileio.ResolveDbPath()
	if path == "" {
		t.Error("expected fallback path when HOME is unset")
	}
}

func TestResolveDbPath_Default(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Unsetenv("MMSYNC_CONF")
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	path := fileio.ResolveDbPath()
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if !strings.HasSuffix(path, "mmsync-state.json") {
		t.Errorf("expected path to end with 'mmsync-state.json', got '%s'", path)
	}
}

func TestResolveDbPath_WithMMSYNCConf(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", "/custom/path")
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	path := fileio.ResolveDbPath()
	expected := "/custom/path/mmsync-state.json"
	if path != expected {
		t.Errorf("expected '%s', got '%s'", expected, path)
	}
}

func TestMigrateConfigData_NoMigration(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Unsetenv("MMSYNC_CONF")
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	err := fileio.MigrateConfigData(fileio.ResolveConfigPath())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrateConfigData_WithMMSYNCConfNoOldFile(t *testing.T) {
	dir := t.TempDir()
	newConfigPath := filepath.Join(dir, "config.yaml")

	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", dir)
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	err := fileio.MigrateConfigData(newConfigPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrateConfigData_MigrationSuccess(t *testing.T) {
	dir := t.TempDir()
	realHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", realHome) }()

	oldConfigDir := filepath.Join(dir, ".config/mmsync")
	_ = os.MkdirAll(oldConfigDir, 0755)
	oldConfigFile := filepath.Join(oldConfigDir, "config.yaml")
	oldDbFile := filepath.Join(oldConfigDir, "mmsync-state.json")

	_ = os.WriteFile(oldConfigFile, []byte("old: config"), 0644)
	_ = os.WriteFile(oldDbFile, []byte("{}"), 0644)

	newDir := t.TempDir()
	newConfigPath := filepath.Join(newDir, "config.yaml")

	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", newDir)
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	err := fileio.MigrateConfigData(newConfigPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(newConfigPath); os.IsNotExist(err) {
		t.Error("expected config file to be migrated")
	}
	newDbFile := filepath.Join(newDir, "mmsync-state.json")
	if _, err := os.Stat(newDbFile); os.IsNotExist(err) {
		t.Error("expected db file to be migrated")
	}
}

func TestMigrateConfigData_TargetExists(t *testing.T) {
	dir := t.TempDir()
	realHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", realHome) }()

	oldConfigDir := filepath.Join(dir, ".config/mmsync")
	_ = os.MkdirAll(oldConfigDir, 0755)
	oldConfigFile := filepath.Join(oldConfigDir, "config.yaml")
	_ = os.WriteFile(oldConfigFile, []byte("old: config"), 0644)

	newDir := t.TempDir()
	newConfigPath := filepath.Join(newDir, "config.yaml")
	_ = os.WriteFile(newConfigPath, []byte("new: config"), 0644)

	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", newDir)
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	err := fileio.MigrateConfigData(newConfigPath)
	if err == nil {
		t.Error("expected error when target already exists")
	}
}

func TestMigrateConfigData_CopyConfigError(t *testing.T) {
	dir := t.TempDir()
	realHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", realHome) }()

	oldConfigDir := filepath.Join(dir, ".config/mmsync")
	_ = os.MkdirAll(oldConfigDir, 0755)
	oldConfigFile := filepath.Join(oldConfigDir, "config.yaml")
	_ = os.WriteFile(oldConfigFile, []byte("old: config"), 0644)

	newDir := t.TempDir()
	newConfigPath := filepath.Join(newDir, "nested", "config.yaml")

	if err := os.MkdirAll(filepath.Dir(newConfigPath), 0555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(filepath.Dir(newConfigPath), 0755) }()

	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", newDir)
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	err := fileio.MigrateConfigData(newConfigPath)
	if err == nil {
		t.Error("expected error when dest dir is not writable")
	}
}

func TestMigrateConfigData_OldDirSameAsNew(t *testing.T) {
	dir := t.TempDir()
	realHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", realHome) }()

	oldConfigDir := filepath.Join(dir, ".config/mmsync")
	_ = os.MkdirAll(oldConfigDir, 0755)

	newConfigPath := filepath.Join(oldConfigDir, "config.yaml")

	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", oldConfigDir)
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	err := fileio.MigrateConfigData(newConfigPath)
	if err != nil {
		t.Fatalf("expected no error when old=new dirs, got: %v", err)
	}
}
