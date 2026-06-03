package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mns/cmd"
)

func TestResolveAndValidatePath_Root(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	result, err := cmd.ResolveAndValidatePath("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "/" {
		t.Errorf("expected '/', got '%s'", result)
	}
}

func TestAddDirectoryEntry_Symlinked(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	targetDir := filepath.Join(dir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "mylink")
	if err := os.Symlink(targetDir, linkDir); err != nil {
		t.Fatal(err)
	}

	err := cmd.AddDirectoryEntry(linkDir, "linked")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
