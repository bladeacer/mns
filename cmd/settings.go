package cmd

import (
	"fmt"
	"os"

	"github.com/bladeacer/mns/internal/fileio"
	yamlwrapper "github.com/bladeacer/mns/internal/yaml"
	"github.com/spf13/cobra"
)

var getArchiverCmd = &cobra.Command{
	Use:   "get-archiver",
	Short: "Show the current archiver (tar or zip)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(AppConf.ConfigSchema.Archiver)
	},
}

var setArchiverCmd = &cobra.Command{
	Use:   "set-archiver <tar|zip>",
	Short: "Set the archiver to use (tar or zip)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		val := args[0]
		if val != "tar" && val != "zip" {
			fmt.Fprintf(os.Stderr, "Error: archiver must be 'tar' or 'zip', got '%s'\n", val)
			os.Exit(1)
		}
		AppConf.ConfigSchema.Archiver = val
		SaveConfig()
		fmt.Printf("Archiver set to '%s'.\n", val)
	},
}

var getCommitFmtCmd = &cobra.Command{
	Use:   "get-commit-fmt",
	Short: "Show the current commit message format",
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(AppConf.ConfigSchema.CommitFmt)
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
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		AppConf.ConfigSchema.CommitFmt = args[0]
		SaveConfig()
		fmt.Printf("Commit format set to '%s'.\n", args[0])
	},
}

var getIgnoreCmd = &cobra.Command{
	Use:   "get-ignore",
	Short: "Show whether .gitignore is respected during staging",
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if AppConf.ConfigSchema.RespectGitignore {
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
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		switch args[0] {
		case "0":
			AppConf.ConfigSchema.RespectGitignore = false
		case "1":
			AppConf.ConfigSchema.RespectGitignore = true
		default:
			fmt.Fprintf(os.Stderr, "Error: value must be '0' or '1', got '%s'\n", args[0])
			os.Exit(1)
		}
		SaveConfig()
		fmt.Printf("Respect .gitignore set to %s.\n", args[0])
	},
}

var getHistLimitCmd = &cobra.Command{
	Use:   "get-hist-limit",
	Short: "Show the history retention limits",
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("History retention:\n")
		fmt.Printf("  Days: %d\n", AppConf.ConfigSchema.HistLimitDays)
		fmt.Printf("  Max size: %d MB\n", AppConf.ConfigSchema.HistLimitSizeMb)
	},
}

var setHistLimitCmd = &cobra.Command{
	Use:   "set-hist-limit",
	Short: "Set the history retention limits",
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
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
			AppConf.ConfigSchema.HistLimitDays = days
		}
		if size > 0 {
			AppConf.ConfigSchema.HistLimitSizeMb = size
		}
		SaveConfig()
		fmt.Println("History limits updated.")
	},
}

var clearHistCmd = &cobra.Command{
	Use:   "clear-hist",
	Short: "Clear staging history and remove staged files",
	Long: `Removes all files from the staging area and clears the recorded history.
Requires confirmation.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
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

		staging := StagingDir()
		if err := os.RemoveAll(staging); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove staging directory: %v\n", err)
		}

		DataStore.ClearHistory()
		dbPath := fileio.ResolveDbPath()
		if err := DataStore.SaveData(dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Staging area and history cleared.")
	},
}

func SaveConfig() {
	configPath := fileio.ResolveConfigPath()
	if err := yamlwrapper.SaveConfig(AppConf, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(getArchiverCmd)
	RootCmd.AddCommand(setArchiverCmd)
	RootCmd.AddCommand(getCommitFmtCmd)
	RootCmd.AddCommand(setCommitFmtCmd)
	RootCmd.AddCommand(getIgnoreCmd)
	RootCmd.AddCommand(setIgnoreCmd)
	RootCmd.AddCommand(getHistLimitCmd)
	RootCmd.AddCommand(setHistLimitCmd)
	RootCmd.AddCommand(clearHistCmd)

	setHistLimitCmd.Flags().IntP("days", "d", 0, "Number of days to retain history")
	setHistLimitCmd.Flags().Int64P("size", "s", 0, "Maximum size in MB for history")
}
