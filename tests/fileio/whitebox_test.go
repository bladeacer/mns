package fileio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mns/internal/fileio"
)

func TestCopyFile_SrcNotExist(t *testing.T) {
	dir := t.TempDir()
	err := fileio.CopyFile(filepath.Join(dir, "nonexistent"), filepath.Join(dir, "dest"))
	if err == nil {
		t.Error("expected error when source does not exist")
	}
}

func TestCopyFile_DstDirNotWritable(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subdir, 0555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(subdir, 0755) }()

	err := fileio.CopyFile(srcPath, filepath.Join(subdir, "dest.txt"))
	if err == nil {
		t.Error("expected error when dest dir is not writable")
	}
}

func TestCopyFile_MkdirAllError(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	blockPath := filepath.Join(dir, "block")
	if err := os.WriteFile(blockPath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	dstPath := filepath.Join(blockPath, "subdir", "dest.txt")

	err := fileio.CopyFile(srcPath, dstPath)
	if err == nil {
		t.Error("expected error when a file blocks directory creation")
	}
}
