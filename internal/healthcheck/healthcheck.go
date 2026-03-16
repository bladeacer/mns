package healthcheck

import (
	"os"
	"path/filepath"
)

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
