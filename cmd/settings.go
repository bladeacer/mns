package cmd

import (
	"fmt"
	"os"

	"github.com/bladeacer/mmsync/internal/fileio"
	yamlwrapper "github.com/bladeacer/mmsync/internal/yaml"
	"github.com/spf13/cobra"
)

var getArchiverCmd = &cobra.Command{
	Use:   "get-archiver",
	Short: "Show the current archiver (tar or zip)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(appConf.ConfigSchema.Archiver)
	},
}

var setArchiverCmd = &cobra.Command{
	Use:   "set-archiver <tar|zip>",
	Short: "Set the archiver to use (tar or zip)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		val := args[0]
		if val != "tar" && val != "zip" {
			fmt.Fprintf(os.Stderr, "Error: archiver must be 'tar' or 'zip', got '%s'\n", val)
			os.Exit(1)
		}
		appConf.ConfigSchema.Archiver = val
		saveConfig()
		fmt.Printf("Archiver set to '%s'.\n", val)
	},
}

var getCommitFmtCmd = &cobra.Command{
	Use:   "get-commit-fmt",
	Short: "Show the current commit message format",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(appConf.ConfigSchema.CommitFmt)
	},
}

var setCommitFmtCmd = &cobra.Command{
	Use:   "set-commit-fmt <format>",
	Short: "Set the commit message format (Go time layout)",
	Long: `Set the commit message format using Go's time layout reference.
Reference time: Mon Jan 2 15:04:05 MST 2006

Examples:
  mns set-commit-fmt "mnemosync archive 2006-01-02"
  mns set-commit-fmt "backup 2006-01-02 150405"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		appConf.ConfigSchema.CommitFmt = args[0]
		saveConfig()
		fmt.Printf("Commit format set to '%s'.\n", args[0])
	},
}

var getIgnoreCmd = &cobra.Command{
	Use:   "get-ignore",
	Short: "Show whether .gitignore is respected during staging",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if appConf.ConfigSchema.RespectGitignore {
			fmt.Println("1")
		} else {
			fmt.Println("0")
		}
	},
}

var setIgnoreCmd = &cobra.Command{
	Use:   "set-ignore <0|1>",
	Short: "Set whether to respect .gitignore during staging (1=yes, 0=no)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		switch args[0] {
		case "0":
			appConf.ConfigSchema.RespectGitignore = false
		case "1":
			appConf.ConfigSchema.RespectGitignore = true
		default:
			fmt.Fprintf(os.Stderr, "Error: value must be '0' or '1', got '%s'\n", args[0])
			os.Exit(1)
		}
		saveConfig()
		fmt.Printf("Respect .gitignore set to %s.\n", args[0])
	},
}

var getHistLimitCmd = &cobra.Command{
	Use:   "get-hist-limit",
	Short: "Show the history retention limits",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("History retention:\n")
		fmt.Printf("  Days: %d\n", appConf.ConfigSchema.HistLimitDays)
		fmt.Printf("  Max size: %d MB\n", appConf.ConfigSchema.HistLimitSizeMb)
	},
}

var setHistLimitCmd = &cobra.Command{
	Use:   "set-hist-limit",
	Short: "Set the history retention limits",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		days, _ := cmd.Flags().GetInt("days")
		size, _ := cmd.Flags().GetInt64("size")

		if days == 0 && size == 0 {
			fmt.Fprintf(os.Stderr, "Error: at least one of --days or --size must be provided.\n")
			os.Exit(1)
		}
		if days > 0 {
			appConf.ConfigSchema.HistLimitDays = days
		}
		if size > 0 {
			appConf.ConfigSchema.HistLimitSizeMb = size
		}
		saveConfig()
		fmt.Println("History limits updated.")
	},
}

var clearHistCmd = &cobra.Command{
	Use:   "clear-hist",
	Short: "Clear staging history and remove staged files",
	Long: `Removes all files from the staging area and clears the recorded history.
Requires confirmation.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print("Warning: This will delete all staged files and clear history. Are you sure? [y/N]: ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil || (response != "y" && response != "Y" && response != "yes" && response != "YES") {
			fmt.Println("Aborted.")
			return
		}

		staging := stagingDir()
		if err := os.RemoveAll(staging); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove staging directory: %v\n", err)
		}

		dataStore.ClearHistory()
		dbPath := fileio.ResolveDbPath()
		if err := dataStore.SaveData(dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Staging area and history cleared.")
	},
}

func saveConfig() {
	configPath := fileio.ResolveConfigPath()
	if err := yamlwrapper.SaveConfig(appConf, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(getArchiverCmd)
	rootCmd.AddCommand(setArchiverCmd)
	rootCmd.AddCommand(getCommitFmtCmd)
	rootCmd.AddCommand(setCommitFmtCmd)
	rootCmd.AddCommand(getIgnoreCmd)
	rootCmd.AddCommand(setIgnoreCmd)
	rootCmd.AddCommand(getHistLimitCmd)
	rootCmd.AddCommand(setHistLimitCmd)
	rootCmd.AddCommand(clearHistCmd)

	setHistLimitCmd.Flags().IntP("days", "d", 0, "Number of days to retain history")
	setHistLimitCmd.Flags().Int64P("size", "s", 0, "Maximum size in MB for history")
}
