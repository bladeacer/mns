package healthcheck_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mmsync/internal/healthcheck"
)

func TestGitDirExists_ValidGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	exists, err := healthcheck.GitDirExists(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true for a directory with .git subdirectory")
	}
}

func TestGitDirExists_NoGitDir(t *testing.T) {
	dir := t.TempDir()

	exists, err := healthcheck.GitDirExists(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for a directory without .git")
	}
}

func TestGitDirExists_GitFile(t *testing.T) {
	dir := t.TempDir()
	gitFile := filepath.Join(dir, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: ../.git"), 0644); err != nil {
		t.Fatal(err)
	}

	exists, err := healthcheck.GitDirExists(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false when .git is a file, not a directory")
	}
}

func TestGitDirExists_NonExistentPath(t *testing.T) {
	exists, err := healthcheck.GitDirExists("/tmp/nonexistent-path-42abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for non-existent path")
	}
}

func TestCheckBinary_Found(t *testing.T) {
	result := healthcheck.CheckBinary("sh")
	if !result.Found {
		t.Fatal("expected 'sh' to be found in PATH")
	}
	if result.Path == "" {
		t.Error("expected non-empty path")
	}
	if result.Version == "" {
		t.Error("expected non-empty version string")
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestCheckBinary_NotFound(t *testing.T) {
	result := healthcheck.CheckBinary("this-binary-does-not-exist-12345")
	if result.Found {
		t.Error("expected Found to be false for non-existent binary")
	}
	if result.Path != "" {
		t.Error("expected empty path for non-existent binary")
	}
	if result.Error == nil {
		t.Error("expected non-nil error for non-existent binary")
	}
}

func TestCheckBinary_NonZeroExit(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "failing.sh")
	content := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
	defer os.Setenv("PATH", origPath)

	result := healthcheck.CheckBinary("failing.sh")
	if !result.Found {
		t.Fatal("expected the script to be found")
	}
	if result.Error == nil {
		t.Error("expected error for binary that exits with non-zero")
	}
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
}

func TestCheckBinary_VersionOutput(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "versioner.sh")
	content := "#!/bin/sh\necho \"myapp version 1.2.3\"\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
	defer os.Setenv("PATH", origPath)

	result := healthcheck.CheckBinary("versioner.sh")
	if !result.Found {
		t.Fatal("expected the script to be found")
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if !strings.Contains(result.Version, "1.2.3") {
		t.Errorf("expected version to contain '1.2.3', got: %s", result.Version)
	}
}

func TestCheckBinary_MultiLineVersion(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "multiline.sh")
	content := "#!/bin/sh\necho \"first line version\"\necho \"second line\"\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
	defer os.Setenv("PATH", origPath)

	result := healthcheck.CheckBinary("multiline.sh")
	if !result.Found {
		t.Fatal("expected the script to be found")
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Version != "first line version" {
		t.Errorf("expected 'first line version', got: '%s'", result.Version)
	}
}
