package fileio

import (
	"os"
	"path/filepath"
)

func AtomicWriteFile(targetPath string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(targetPath))
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Track if we need to clean up the temp file on failure
	var keep bool
	defer func() {
		if !keep {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		return err
	}

	keep = true // Successfully renamed, cancel the deferred cleanup

	if f, err := os.Open(dir); err == nil {
		_ = f.Sync()
		_ = f.Close()
	}

	return nil
}
