package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/fileio"
	"github.com/bladeacer/mns/internal/healthcheck"
	"github.com/bladeacer/mns/internal/yaml"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

var repoPathFlag string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new configuration file with default values.",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := fileio.ResolveConfigPath()
		dbPath := fileio.ResolveDbPath()

		dbExists, err := ValidateInitPreconditions(configPath, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		finalRepoPath, err := resolveInitRepoPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nInitialization aborted: %v\n", err)
			os.Exit(1)
		}

		defaultConfig := config.GetMnemoConf()
		defaultConfig.ConfigSchema.IsInit = true
		defaultConfig.ConfigSchema.RepoPath = finalRepoPath
		defaultConfig.ConfigSchema.DbPath = dbPath

		CompleteInitSetup(finalRepoPath, configPath, dbPath, dbExists, defaultConfig)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&repoPathFlag, "repo-path", "r", "", "Specify the path to the target Git repository.")
}

func ValidateInitPreconditions(configPath, dbPath string) (dbExists bool, err error) {
	_, confErr := os.Stat(configPath)
	_, dbErr := os.Stat(dbPath)

	if confErr == nil {
		return false, fmt.Errorf("Cannot run init. Configuration file already exists at %s\nPlease remove the existing configuration file before running 'init'.", configPath)
	} else if !os.IsNotExist(confErr) {
		fmt.Fprintf(os.Stderr, "Error checking for config file at %s: %v\n", configPath, confErr)
	}

	if dbErr != nil && !os.IsNotExist(dbErr) {
		fmt.Fprintf(os.Stderr, "Error checking for database file at %s: %v\n", dbPath, dbErr)
	}

	return dbErr == nil, nil
}

func resolveInitRepoPath() (string, error) {
	if repoPathFlag != "" {
		return ProcessRepoPath(repoPathFlag)
	}
	return getRepoPathInteractive()
}

func CompleteInitSetup(repoPath, configPath, dbPath string, dbExists bool, cfg *config.MnemoConf) {
	exists, _ := healthcheck.GitDirExists(repoPath)

	if exists {
		fmt.Printf("\nRepository path validated: '%s/.git' exists.\n", repoPath)
		yaml.WriteYAML(cfg, configPath)
		if dbExists {
			fmt.Printf("Database file already exists at: '%s'. Leaving as-is.\n", dbPath)
		} else {
			_ = config.GetDataStore().SaveData(dbPath)
			fmt.Printf("\nDatabase created at: '%s'.\n", dbPath)
		}

		if err := EnsureGitignoreInDir(repoPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not ensure .mnemosync is gitignored: %v\n", err)
		}
	} else {
		fmt.Printf("\nDirectory '%s/.git' does not exist.\n", repoPath)
		fmt.Printf("Aborting configuration write.\n")
	}
}

func expandTilde(line string) (string, string, bool) {
	if !strings.HasPrefix(line, "~") {
		return line, "", false
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return line, "", false
	}
	if line == "~" {
		return homeDir + string(os.PathSeparator), homeDir, true
	}
	if strings.HasPrefix(line, "~"+string(os.PathSeparator)) {
		return filepath.Join(homeDir, line[2:]), homeDir, true
	}
	return line, homeDir, true
}

func buildPathSuggestions(entries []os.DirEntry, dir, prefix, homeDir string, homePrefix bool) []string {
	var suggestions []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(prefix, ".") && strings.HasPrefix(name, ".") {
			continue
		}
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		suggestion := filepath.Join(dir, name)
		if homePrefix && strings.HasPrefix(suggestion, homeDir) {
			suggestion = "~" + suggestion[len(homeDir):]
		}
		if entry.IsDir() && suggestion[len(suggestion)-1] != os.PathSeparator {
			suggestion += string(os.PathSeparator)
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions
}

func PathCompleter(line string) []string {
	expanded, homeDir, homePrefix := expandTilde(line)

	dir, prefix := filepath.Split(expanded)
	if dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	return buildPathSuggestions(entries, dir, prefix, homeDir, homePrefix)
}

func ProcessRepoPath(inputPath string) (string, error) {
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
	defer func() { _ = line.Close() }()

	line.SetCompleter(PathCompleter)
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

		finalRepoPath, err := ProcessRepoPath(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Printf("Path accepted: %s\n", finalRepoPath)
		return finalRepoPath, nil
	}
}
