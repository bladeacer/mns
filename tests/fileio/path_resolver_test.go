package fileio_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mns/internal/fileio"
)

func TestResolveConfigPath_Fallback(t *testing.T) {
	realHome := os.Getenv("HOME")
	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Unsetenv("HOME")
	_ = os.Unsetenv("MMSYNC_CONF")
	defer func() {
		_ = os.Setenv("HOME", realHome)
		_ = os.Setenv("MMSYNC_CONF", prevConf)
	}()

	path := fileio.ResolveConfigPath()
	if path == "" {
		t.Error("expected non-empty fallback path")
	}
}

func TestResolveConfigPath_XdgOverride(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	prevXdg := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Unsetenv("MMSYNC_CONF")
	_ = os.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	defer func() {
		_ = os.Setenv("MMSYNC_CONF", prevConf)
		_ = os.Setenv("XDG_CONFIG_HOME", prevXdg)
	}()

	path := fileio.ResolveConfigPath()
	if !strings.HasPrefix(path, "/custom/xdg") {
		t.Errorf("expected path to start with '/custom/xdg', got '%s'", path)
	}
}

func TestResolveDbPath_XdgOverride(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	prevXdgData := os.Getenv("XDG_DATA_HOME")
	_ = os.Unsetenv("MMSYNC_CONF")
	_ = os.Setenv("XDG_DATA_HOME", "/custom/xdg-data")
	defer func() {
		_ = os.Setenv("MMSYNC_CONF", prevConf)
		_ = os.Setenv("XDG_DATA_HOME", prevXdgData)
	}()

	path := fileio.ResolveDbPath()
	if !strings.HasPrefix(path, "/custom/xdg-data") {
		t.Errorf("expected path to start with '/custom/xdg-data', got '%s'", path)
	}
}

func TestConfigDir_FromMMSYNCConf(t *testing.T) {
	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", filepath.Join("some", "path"))
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	dir := fileio.ConfigDir()
	if dir == "" {
		t.Error("expected non-empty config dir")
	}
	_ = dir
}

var _ = fileio.ConfigDir
var _ = fileio.DbDir
