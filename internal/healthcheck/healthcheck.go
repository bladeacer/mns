package healthcheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BinaryResult struct {
	Name     string
	Found    bool
	Path     string
	Version  string
	ExitCode int
	Error    error
}

func GitDirExists(path string) (bool, error) {
	info, err := os.Stat(filepath.Join(path, ".git"))
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CheckBinary(binaryName string) BinaryResult {
	result := BinaryResult{Name: binaryName}

	path, err := exec.LookPath(binaryName)
	if err != nil {
		result.Error = err
		return result
	}

	result.Found = true
	result.Path = path

	cmd := exec.Command(binaryName, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = err
		result.Version = strings.TrimSpace(string(output))
		return result
	}

	firstLine := strings.SplitN(string(output), "\n", 2)[0]
	result.Version = strings.TrimSpace(firstLine)
	return result
}
