package cmd

import (
	"fmt"
	"github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/fileio"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

// TODO: This command helps add directory paths to be staged before performing backup. Have CRUD in this.
// Somehow rsync directories to the target directory and then tar archive all of them when push is called

var aliases []string
var addCmd = &cobra.Command{
	Use:   "add [path_1] [path_2]...",
	Short: "Add one or more target paths to be tracked for backup",
	Long: `Add one or more target paths to be tracked for backup.
If provided, the number of aliases must match the number of paths.

Examples:

mmsync add ./
mmsync add ./ --alias="test"
mmsync add ./ ~/test_dir --alias="test","test_dir_w_alias"

Adds the current directory recursively to be staged.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := fileio.ResolveConfigPath()
		isInit := appConf.ConfigSchema.IsInit

		if !isInit {
			fmt.Printf("\nConfiguration file not found at expected path\n%s\nRun mns init to start.\n", configPath)
		} else {
			addWrapper(args)
			fmt.Println("\nFinished adding entries.")
		}

	},
}

func addWrapper(args []string) {
	if len(aliases) > 0 && len(aliases) != len(args) {
		fmt.Fprintf(os.Stderr, "Error: Number of paths (%d) must match number of aliases (%d).\n", len(args), len(aliases))
		os.Exit(1)
	}

	for i, argPath := range args {
		resolvedPath, err := resolveAndValidatePath(argPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing path '%s': %v\n", argPath, err)
			os.Exit(1)
		}

		var alias string
		if len(aliases) > i {
			alias = aliases[i]
		} else {
			alias = filepath.Base(resolvedPath)
		}
		err = addDirectoryEntry(resolvedPath, alias)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal Error adding path '%s': %v\n", argPath, err)
			os.Exit(1)
		}
	}
}

func resolveAndValidatePath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory for tilde expansion: %w", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	targetPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path provided: %w", err)
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path '%s' does not exist", targetPath)
		}
		return "", fmt.Errorf("error checking path '%s': %w", targetPath, err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("path '%s' is a file, only directories can be added", targetPath)
	}
	if targetPath == appConf.ConfigSchema.RepoPath {
		return "", fmt.Errorf("cannot circular reference repo path: '%s'", targetPath)
	}
	if filepath.Dir(targetPath) == filepath.Dir(appConf.ConfigSchema.ConfigPath) {
		return "", fmt.Errorf("cannot circular reference config path: '%s'", targetPath)
	}
	if filepath.Base(targetPath) == "mnemosync" {
		return "", fmt.Errorf("do not use the dev repo: '%s'", targetPath)
	}

	return targetPath, nil
}

func addDirectoryEntry(targetPath string, alias string) error {
	for newID, entry := range dataStore.TrackedDirs {
		if entry.TargetPath == targetPath {
			return fmt.Errorf("path '%s' is already being tracked (ID: %s, Alias: %s)",
				targetPath, newID, entry.Alias)
		}

		if entry.Alias == alias {
			return fmt.Errorf("alias '%s' is already in use by path '%s' (ID: %s)",
				alias, entry.TargetPath, newID)
		}
	}

	dbPath := fileio.ResolveDbPath()

	newEntry := config.DirData{
		TargetPath: targetPath,
		Alias:      alias,
	}

	newID := dataStore.AddDir(newEntry)

	if err := dataStore.SaveData(dbPath); err != nil {
		return fmt.Errorf("failed to save data store after adding entry: %w", err)
	}

	fmt.Printf("Successfully added directory:\n")
	fmt.Printf("\tID: %s\n", newID)
	fmt.Printf("\tPath: %s\n", targetPath)
	fmt.Printf("\tAlias: %s\n", alias)

	return nil
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringSliceVarP(&aliases, "alias", "a", []string{}, "Comma-separated list of aliases for the corresponding paths.")
}
