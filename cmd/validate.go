package cmd

import (
	"fmt"
	"os"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/confighandler"
	"github.com/bladeacer/mns/internal/fileio"
	"github.com/spf13/cobra"
)

var validateConfigPath string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and database files",
	Long: `Loads the configuration file (and optionally database), applies
any schema healing, and reports whether the files are valid.

If --config is specified, the given file is loaded instead of the
default config path. This is useful for testing that schema healing
does not corrupt an existing configuration.

Exits with code 0 if both config and database are valid,
1 if warnings were emitted, or 2 if errors occurred.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := validateConfigPath
		exitCode := ValidateConfigAndDataStore(path)
		os.Exit(exitCode)
	},
}

func ValidateConfigAndDataStore(configPath string) int {
	exitCode := 0

	if err := validateConfig(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		exitCode = 2
	}

	if err := validateDataStore(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		exitCode = 2
	}

	if exitCode == 0 {
		fmt.Println("Validation passed.")
	}

	return exitCode
}

func validateConfig(configPath string) error {
	if configPath == "" {
		configPath = fileio.ResolveConfigPath()
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	cfg, err := confighandler.LoadConfigWithPath(configPath)
	if err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("  AppVersion: %s\n", cfg.ConfigSchema.AppVersion)
	fmt.Printf("  IsInit: %t\n", cfg.ConfigSchema.IsInit)
	if cfg.ConfigSchema.IsInit {
		fmt.Printf("  RepoPath: %s\n", cfg.ConfigSchema.RepoPath)
		fmt.Printf("  DbPath: %s\n", cfg.ConfigSchema.DbPath)
	}
	fmt.Printf("  Archiver: %s\n", cfg.ConfigSchema.Archiver)
	fmt.Printf("  CommitFmt: %s\n", cfg.ConfigSchema.CommitFmt)
	fmt.Printf("  KeepArchives: %d\n", cfg.ConfigSchema.KeepArchives)
	fmt.Printf("  LfsThresholdMb: %d\n", cfg.ConfigSchema.LfsThresholdMb)

	return nil
}

func validateDataStore() error {
	dbPath := fileio.ResolveDbPath()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	}

	ds, err := config.LoadDataStore()
	if err != nil {
		return fmt.Errorf("database validation failed: %w", err)
	}

	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("  CurrentId: %d\n", ds.CurrentId)
	fmt.Printf("  TrackedDirs: %d\n", len(ds.TrackedDirs))
	fmt.Printf("  StagingHistory: %d entries\n", len(ds.StagingHistory))

	return nil
}

func init() {
	RootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&validateConfigPath, "config", "c", "", "Path to a specific configuration file to validate")
}
