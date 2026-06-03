package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mns/cmd"
)

func TestResolveAndValidatePath_Nonexistent(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	_, err := cmd.ResolveAndValidatePath(filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestResolveAndValidatePath_FilePath(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := cmd.ResolveAndValidatePath(filePath)
	if err == nil {
		t.Fatal("expected error when path is a file")
	}
}

func TestAddDirectoryEntry_EmptyAlias(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	subDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := cmd.AddDirectoryEntry(subDir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
