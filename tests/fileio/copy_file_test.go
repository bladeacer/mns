package fileio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mns/internal/fileio"
)

func TestCopyFile_EmptySource(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := fileio.CopyFile(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(data))
	}
}

func TestCopyFile_BinaryContent(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")
	content := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := fileio.CopyFile(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(content) {
		t.Errorf("expected %d bytes, got %d", len(content), len(data))
	}
}

func TestMigrateConfigData_StatOldConfigError(t *testing.T) {
	dir := t.TempDir()
	realHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", realHome) }()

	oldConfigDir := filepath.Join(dir, ".config/mmsync")
	_ = os.MkdirAll(oldConfigDir, 0755)

	newDir := t.TempDir()
	newConfigPath := filepath.Join(newDir, "config.yaml")

	prevConf := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", newDir)
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevConf) }()

	err := fileio.MigrateConfigData(newConfigPath)
	if err != nil {
		t.Fatalf("expected no error when old config not found, got: %v", err)
	}
}
