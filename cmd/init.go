package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/fileio"
	"github.com/bladeacer/mmsync/internal/healthcheck"
	"github.com/bladeacer/mmsync/internal/yaml"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

// Global variable to hold the path passed via flag
var repoPathFlag string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new configuration file with default values.",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := fileio.ResolveConfigPath()
		dbPath := fileio.ResolveDbPath()
		_, confErr := os.Stat(configPath)
		_, dbErr := os.Stat(dbPath)

		if confErr == nil || dbErr == nil {
			fmt.Fprintf(os.Stderr, "Error: Cannot run init. The following files already exist:\n")
			if confErr == nil {
				fmt.Fprintf(os.Stderr, "- Configuration file at %s\n", configPath)
			}
			if dbErr == nil {
				fmt.Fprintf(os.Stderr, "- Database file at %s\n", dbPath)
			}
			fmt.Fprintf(os.Stderr, "Please remove the existing files before running 'init'.\n")
			os.Exit(1)
		} else {
			if !os.IsNotExist(confErr) {
				fmt.Fprintf(os.Stderr, "Error checking for config file at %s: %v\n", configPath, confErr)
			}
			if !os.IsNotExist(dbErr) {
				fmt.Fprintf(os.Stderr, "Error checking for database file at %s: %v\n", dbPath, dbErr)
			}
		}

		var finalRepoPath string
		var err error

		if repoPathFlag != "" {
			finalRepoPath, err = processRepoPath(repoPathFlag)
		} else {
			finalRepoPath, err = getRepoPathInteractive()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "\nInitialization aborted: %v\n", err)
			os.Exit(1)
		}

		defaultConfig := config.GetMnemoConf()
		defaultConfig.ConfigSchema.IsInit = true
		defaultConfig.ConfigSchema.RepoPath = finalRepoPath

		exists, _ := healthcheck.GitDirExists(finalRepoPath)

		if exists {
			fmt.Printf("\nRepository path validated: '%s/.git' exists.\n", finalRepoPath)
			yaml.WriteYAML(defaultConfig, configPath)
			config.GetDataStore().SaveData(dbPath)
			fmt.Printf("\nDatabase created at: '%s'.\n", dbPath)

			repoPath := finalRepoPath
			if err := ensureGitignoreInDir(repoPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not ensure .mnemosync is gitignored: %v\n", err)
			}
		} else {
			fmt.Printf("\nDirectory '%s/.git' does not exist.\n", finalRepoPath)
			fmt.Printf("Aborting configuration write.\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&repoPathFlag, "repo-path", "r", "", "Specify the path to the target Git repository.")
}

func pathCompleter(line string) []string {
	var homeDir string
	var err error
	homePrefix := strings.HasPrefix(line, "~")

	if homePrefix {
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil
		}

		if line == "~" {
			line = homeDir + string(os.PathSeparator)
		} else if strings.HasPrefix(line, "~"+string(os.PathSeparator)) {
			line = filepath.Join(homeDir, line[2:])
		}
	}

	dir, prefix := filepath.Split(line)

	targetDir := dir
	if targetDir == "" {
		targetDir = "."
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil
	}

	var suggestions []string
	for _, entry := range entries {
		name := entry.Name()

		if !strings.HasPrefix(prefix, ".") && strings.HasPrefix(name, ".") {
			continue
		}

		if strings.HasPrefix(name, prefix) {

			suggestion := filepath.Join(dir, name)

			if homePrefix && strings.HasPrefix(suggestion, homeDir) {
				suggestion = "~" + suggestion[len(homeDir):]
			}

			if entry.IsDir() {
				if suggestion[len(suggestion)-1] != os.PathSeparator {
					suggestion += string(os.PathSeparator)
				}
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

func processRepoPath(inputPath string) (string, error) {
	if strings.HasPrefix(inputPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}

		if inputPath == "~" {
			inputPath = homeDir
		} else if strings.HasPrefix(inputPath, "~"+string(os.PathSeparator)) {
			relativePath := inputPath[2:]
			inputPath = filepath.Join(homeDir, relativePath)
		}
	}

	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for '%s': %w", inputPath, err)
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("directory '%s' does not exist", absPath)
	} else if err != nil {
		return "", fmt.Errorf("error checking path '%s': %w", absPath, err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("'%s' is not a directory", absPath)
	}

	return absPath, nil
}

func getRepoPathInteractive() (string, error) {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCompleter(pathCompleter)
	line.SetTabCompletionStyle(liner.TabPrints)
	line.SetCtrlCAborts(true)

	fmt.Println("Ensure that the target repository path is correct and does not contain other important files.\nDatabase for storing directories and their aliases would use the same parent directory.")
	for {
		prompt := "Enter a valid path to the target repository to archive files to (e.g., /path/to/repo or ~/myrepo): "

		fmt.Printf("\n\n")
		inputPath, err := line.Prompt(prompt)

		if err != nil {
			if err == liner.ErrPromptAborted {
				fmt.Fprintln(os.Stderr, "\nInput cancelled by user (Ctrl+C).")
				return "", fmt.Errorf("user cancelled input (Ctrl+C)")
			}
			if err == io.EOF {
				fmt.Fprintln(os.Stderr, "\nInput cancelled by user (Ctrl+D).")
				return "", fmt.Errorf("user cancelled input (Ctrl+D)")
			}
		}

		inputPath = strings.TrimSpace(inputPath)

		if inputPath == "" {
			continue
		}

		line.AppendHistory(inputPath)

		finalRepoPath, err := processRepoPath(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Printf("Path accepted: %s\n", finalRepoPath)
		return finalRepoPath, nil
	}
}
