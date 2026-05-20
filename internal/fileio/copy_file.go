package fileio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Optimise this operation to just move?

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer func() { _ = sourceFile.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory for %s: %w", dst, err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy content from %s to %s: %w", src, dst, err)
	}

	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// Copies configuration and database files when new MMSYNC_CONF set
func MigrateConfigData(newConfigPath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	oldConfigDir := filepath.Join(homeDir, DefaultConfigDir)
	oldConfigFile := filepath.Join(oldConfigDir, DefaultConfigFile)
	oldDbFile := filepath.Join(oldConfigDir, DefaultDbFile)

	newConfigDir := filepath.Dir(newConfigPath)
	newDbFile := filepath.Join(newConfigDir, DefaultDbFile)

	if os.Getenv("MMSYNC_CONF") != "" && oldConfigDir != newConfigDir {

		if _, err := os.Stat(oldConfigFile); os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return fmt.Errorf("error checking old configuration file %s: %w", oldConfigFile, err)
		}

		if _, err := os.Stat(newConfigPath); err == nil {
			return fmt.Errorf("cannot migrate: new configuration file already exists at %s", newConfigPath)
		}

		fmt.Fprintf(os.Stderr, "Migrating configuration files from %s to %s...\n", oldConfigDir, newConfigDir)

		if err := copyFile(oldConfigFile, newConfigPath); err != nil {
			return fmt.Errorf("failed to copy configuration file: %w", err)
		}

		if _, err := os.Stat(oldDbFile); err == nil {
			if err := copyFile(oldDbFile, newDbFile); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to copy database file: %v\n", err)
			}
		}

		if err := os.RemoveAll(oldConfigDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to clean up old configuration directory %s: %v\n", oldConfigDir, err)
		}

		fmt.Fprintf(os.Stderr, "Configuration migration complete.\n")
	}

	return nil
}
