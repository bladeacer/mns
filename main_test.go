package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T, dir string) string {
	t.Helper()
	binaryPath := filepath.Join(dir, "mns-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\nOutput: %s", err, string(out))
	}
	return binaryPath
}

func TestMain_RootRejected(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("cannot test root rejection when running as root")
	}
}

func TestMain_VersionFlag(t *testing.T) {
	dir := t.TempDir()
	binaryPath := buildBinary(t, dir)

	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "mnemosync") {
		t.Errorf("expected version output to contain 'mnemosync', got: '%s'", string(output))
	}
}

func TestMain_HealthCmd(t *testing.T) {
	dir := t.TempDir()
	binaryPath := buildBinary(t, dir)

	cmd := exec.Command(binaryPath, "health")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "Health Check") {
		t.Errorf("expected health output, got: '%s'", string(output))
	}
}

func TestMain_VersionCmd(t *testing.T) {
	dir := t.TempDir()
	binaryPath := buildBinary(t, dir)

	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "mnemosync") {
		t.Errorf("expected version output to contain 'mnemosync', got: '%s'", string(output))
	}
}

func TestMain_Help(t *testing.T) {
	dir := t.TempDir()
	binaryPath := buildBinary(t, dir)

	cmd := exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "mns") {
		t.Errorf("expected help output to contain 'mns', got: '%s'", string(output))
	}
}
