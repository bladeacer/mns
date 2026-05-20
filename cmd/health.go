package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/healthcheck"
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
	var statusBuilder strings.Builder

	configPath := cfg.ConfigSchema.ConfigPath
	repoPath := cfg.ConfigSchema.RepoPath
	dbPath := cfg.ConfigSchema.DbPath

	if shouldPrintOutput {
		header := "-- Health Check --"
		fmt.Printf("\n%s\n\n", header)
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

	if shouldPrintOutput {
		fmt.Println("Binaries:")
	}
	for _, dep := range deps {
		msg := CheckBinary(dep.name, dep.isOptional, shouldPrintOutput)
		if msg != "" {
			statusBuilder.WriteString(msg)
			statusBuilder.WriteString("\n")
		}
	}

	if shouldPrintOutput {
		fmt.Println()
		fmt.Println("Configuration:")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("  \u2717 %s (not found - run 'mns init')", configPath)
		statusBuilder.WriteString(msg)
		statusBuilder.WriteString("\n")
		if shouldPrintOutput {
			fmt.Println(msg)
		}
	} else if shouldPrintOutput {
		fmt.Printf("  \u2713 %s\n", configPath)
	}

	if shouldPrintOutput {
		fmt.Println()
		fmt.Println("Repository:")
	}
	if repoPath == "" {
		msg := "  \u2717 repo path not set (run 'mns init')"
		statusBuilder.WriteString(msg)
		statusBuilder.WriteString("\n")
		if shouldPrintOutput {
			fmt.Println(msg)
		}
	} else {
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			msg := fmt.Sprintf("  \u2717 %s (path not found)", repoPath)
			statusBuilder.WriteString(msg)
			statusBuilder.WriteString("\n")
			if shouldPrintOutput {
				fmt.Println(msg)
			}
		} else {
			gitDirExists, err := healthcheck.GitDirExists(repoPath)
			if err != nil {
				msg := fmt.Sprintf("  \u26a0 %s (git check failed: %v)", repoPath, err)
				statusBuilder.WriteString(msg)
				statusBuilder.WriteString("\n")
				if shouldPrintOutput {
					fmt.Println(msg)
				}
			} else if !gitDirExists {
				msg := fmt.Sprintf("  \u26a0 %s (not a git repository)", repoPath)
				statusBuilder.WriteString(msg)
				statusBuilder.WriteString("\n")
				if shouldPrintOutput {
					fmt.Println(msg)
				}
			} else if shouldPrintOutput {
				fmt.Printf("  \u2713 %s\n", repoPath)
			}
		}
	}

	if shouldPrintOutput {
		fmt.Println()
		fmt.Println("Database:")
	}
	if dbPath == "" {
		msg := "  \u2717 db path not set (run 'mns init')"
		statusBuilder.WriteString(msg)
		statusBuilder.WriteString("\n")
		if shouldPrintOutput {
			fmt.Println(msg)
		}
	} else if shouldPrintOutput {
		fmt.Printf("  \u2713 %s\n", dbPath)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("  \u26a0 %s (not found on disk)", dbPath)
		statusBuilder.WriteString(msg)
		statusBuilder.WriteString("\n")
		if shouldPrintOutput {
			fmt.Println(msg)
		}
	} else if shouldPrintOutput {
		fmt.Printf("  \u2713 %s\n", dbPath)
	}

	if shouldPrintOutput {
		fmt.Println()
		fmt.Println("-- Health Check Complete --")
	}

	return statusBuilder.String()
}

func CheckBinary(binaryName string, isOptional bool, shouldPrintOutput bool) string {
	result := healthcheck.CheckBinary(binaryName)

	if !result.Found {
		if !isOptional {
			msg := fmt.Sprintf("  \u2717 %s (required - not found in PATH)", binaryName)
			if shouldPrintOutput {
				fmt.Println(msg)
			}
			return msg
		}
		msg := fmt.Sprintf("  \u26a0 %s (optional - not found in PATH)", binaryName)
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg
	}

	msg := fmt.Sprintf("  \u2713 %s at %s", binaryName, result.Path)
	if shouldPrintOutput {
		fmt.Println(msg)
	}

	if result.Error != nil {
		var warnMsg string
		if result.ExitCode != 0 {
			warnMsg = fmt.Sprintf("      \u26a0 version check: exit %d (%s)", result.ExitCode, result.Version)
		} else {
			warnMsg = fmt.Sprintf("      \u26a0 version check: %v", result.Error)
		}
		if shouldPrintOutput {
			fmt.Println(warnMsg)
		}
		return msg + "\n" + warnMsg
	}

	if result.Version != "" {
		verMsg := fmt.Sprintf("      %s", result.Version)
		if shouldPrintOutput {
			fmt.Println(verMsg)
		}
	}

	return ""
}

func init() {
	RootCmd.AddCommand(healthCmd)
}
