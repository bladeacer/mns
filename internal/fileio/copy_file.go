package fileio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Optimise this operation to just move?

func CopyFile(src, dst string) error {
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
	if os.Getenv("MMSYNC_CONF") == "" {
		return nil
	}

	oldConfigDir := LegacyConfigDirPath()
	newConfigDir := filepath.Dir(newConfigPath)

	if oldConfigDir == newConfigDir {
		return nil
	}

	oldConfigFile := filepath.Join(oldConfigDir, DefaultConfigFile)

	if _, err := os.Stat(oldConfigFile); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking old configuration file %s: %w", oldConfigFile, err)
	}

	if _, err := os.Stat(newConfigPath); err == nil {
		return fmt.Errorf("cannot migrate: new configuration file already exists at %s", newConfigPath)
	}

	fmt.Fprintf(os.Stderr, "Migrating configuration files from %s to %s...\n", oldConfigDir, newConfigDir)

	if err := CopyFile(oldConfigFile, newConfigPath); err != nil {
		return fmt.Errorf("failed to copy configuration file: %w", err)
	}

	migrateDbFile(filepath.Join(oldConfigDir, DefaultDbFile), filepath.Join(newConfigDir, DefaultDbFile))

	fmt.Fprintf(os.Stderr, "Configuration migration complete.\n")
	return nil
}

func migrateDbFile(oldDbFile, newDbFile string) {
	if _, err := os.Stat(oldDbFile); os.IsNotExist(err) {
		return
	}
	if err := CopyFile(oldDbFile, newDbFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to copy database file: %v\n", err)
	}
}
