package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/healthcheck"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checks the health of mnemosync",
	Long: `Checks the health of mnemosync
Checks if the required system binaries are installed

Also checks if the mnemosync configuration files have been created.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunHealthCheck(AppConf, true)
	},
}

func RunHealthCheck(cfg *config.MnemoConf, shouldPrintOutput bool) string {
	var errStrBuilder strings.Builder
	separator := "_"
	repeatedSeparator := strings.Repeat(separator, 72)

	configPath := cfg.ConfigSchema.ConfigPath
	repoPath := cfg.ConfigSchema.RepoPath
	dbPath := cfg.ConfigSchema.DbPath

	if shouldPrintOutput {
		fmt.Println("\n\tRunning Health Check")
	}

	deps := []struct {
		name       string
		isOptional bool
	}{
		{"git", false},
		{"rsync", false},
		{"tar", false},
		{"zip", true},
	}

	for _, dep := range deps {
		if err := CheckBinary(dep.name, dep.isOptional, shouldPrintOutput); err != "" {
			errStrBuilder.WriteString(err)
		}
	}

	if shouldPrintOutput {
		fmt.Printf("\t%s\n\n", repeatedSeparator)
	}

	if shouldPrintOutput {
		fmt.Println("\tConfiguration File:")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("\t\t[NOT FOUND] Configuration file not found at:\n\t\t%s\n\t\tRun 'mns init' to start.\n", configPath)
		errStrBuilder.WriteString(msg)
		if shouldPrintOutput {
			fmt.Print(msg)
		}
	} else if shouldPrintOutput {
		fmt.Printf("\t\t[FOUND] at %s\n", configPath)
	}
	if shouldPrintOutput {
		fmt.Printf("\t%s\n", repeatedSeparator)
	}

	if shouldPrintOutput {
		fmt.Println("\tRepository Path:")
	}
	if repoPath == "" {
		msg := "\t\t[NOT SET] Repository Path is not defined.\n\t\tRun 'mns init' to set.\n"
		errStrBuilder.WriteString(msg)
		if shouldPrintOutput {
			fmt.Print(msg)
		}
	} else {
		if shouldPrintOutput {
			fmt.Printf("\t\t[SET] %s\n", repoPath)
		}

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			msg := fmt.Sprintf("\t\t[WARNING] Repository directory does not exist on disk: %s\n", repoPath)
			errStrBuilder.WriteString(msg)
			if shouldPrintOutput {
				fmt.Print(msg)
			}
		} else {
			gitDirExists, err := healthcheck.GitDirExists(repoPath)
			if err != nil {
				msg := fmt.Sprintf("\t\t[WARNING] Could not check repository git status: %v\n", err)
				errStrBuilder.WriteString(msg)
				if shouldPrintOutput {
					fmt.Print(msg)
				}
			} else if !gitDirExists {
				msg := fmt.Sprintf("\t\t[WARNING] Repository path exists but is not a git repository: %s\n", repoPath)
				errStrBuilder.WriteString(msg)
				if shouldPrintOutput {
					fmt.Print(msg)
				}
			} else if shouldPrintOutput {
				fmt.Printf("\t\t[PASS] Valid git repository\n")
			}
		}
	}
	if shouldPrintOutput {
		fmt.Printf("\t%s\n", repeatedSeparator)
	}

	if shouldPrintOutput {
		fmt.Println("\tDatabase Path:")
	}
	if dbPath == "" {
		msg := "\t\t[NOT SET] Database Path is not defined.\n\t\tRun 'mns init' to start.\n"
		errStrBuilder.WriteString(msg)
		if shouldPrintOutput {
			fmt.Print(msg)
		}
	} else if shouldPrintOutput {
		fmt.Printf("\t\t[SET] %s\n", dbPath)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("\t\t[WARNING] Database file not found on disk: %s\n", dbPath)
		errStrBuilder.WriteString(msg)
		if shouldPrintOutput {
			fmt.Print(msg)
		}
	} else if shouldPrintOutput {
		fmt.Printf("\t\t[FOUND] at %s\n", dbPath)
	}

	if shouldPrintOutput {
		fmt.Printf("\t%s\n", repeatedSeparator)
		fmt.Println("\n\tHealth Check Complete")
	}

	return errStrBuilder.String()
}

func CheckBinary(binaryName string, isOptional bool, shouldPrintOutput bool) string {
	result := healthcheck.CheckBinary(binaryName)

	if !result.Found {
		if !isOptional {
			msg := fmt.Sprintf("[FAIL] Required Binary '%s' not found in PATH.", binaryName)
			if shouldPrintOutput {
				fmt.Println(msg)
			}
			return msg
		}
		msg := fmt.Sprintf("[WARNING] Optional Binary '%s' not found in PATH.", binaryName)
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg
	}

	if shouldPrintOutput {
		fmt.Printf("[PASS] Binary '%s' found at: %s\n", binaryName, result.Path)
	}

	if result.Error != nil {
		var msg string
		if result.ExitCode != 0 {
			msg = fmt.Sprintf("\t[WARNING] Version check failed for '%s'. Exit Code %d. Output:\n\t\t%s\n",
				binaryName, result.ExitCode, result.Version)
		} else {
			msg = fmt.Sprintf("\t[WARNING] Version check failed for '%s': %v\n", binaryName, result.Error)
		}
		if shouldPrintOutput {
			fmt.Print(msg)
		}
		return msg
	}

	if shouldPrintOutput {
		fmt.Printf("\tVersion: %s\n", result.Version)
	}

	return ""
}

func init() {
	RootCmd.AddCommand(healthCmd)
}
