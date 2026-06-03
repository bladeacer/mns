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
	var status strings.Builder

	printOutput(shouldPrintOutput, "", "-- Health Check --", "")

	status.WriteString(checkDeps(shouldPrintOutput))
	status.WriteString(checkConfigPath(cfg.ConfigSchema.ConfigPath, shouldPrintOutput))
	status.WriteString(checkRepoPath(cfg.ConfigSchema.RepoPath, shouldPrintOutput))
	status.WriteString(checkDbPath(cfg.ConfigSchema.DbPath, shouldPrintOutput))

	printOutput(shouldPrintOutput, "", "-- Health Check Complete --")

	return status.String()
}

func printOutput(shouldPrintOutput bool, lines ...string) {
	if !shouldPrintOutput {
		return
	}
	for _, l := range lines {
		fmt.Println(l)
	}
}

var deps = []struct {
	name       string
	isOptional bool
}{
	{"git", false},
	{"rsync", false},
	{"tar", false},
	{"zip", true},
	{"git-lfs", true},
}

func checkDeps(shouldPrintOutput bool) string {
	var status strings.Builder

	printOutput(shouldPrintOutput, "Binaries:")

	for _, dep := range deps {
		msg := CheckBinary(dep.name, dep.isOptional, shouldPrintOutput)
		if msg != "" {
			status.WriteString(msg)
			status.WriteString("\n")
		}
	}

	return status.String()
}

func checkConfigPath(configPath string, shouldPrintOutput bool) string {
	printOutput(shouldPrintOutput, "", "Configuration:")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("  [!!] %s (not found - run 'mns init')", configPath)
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg + "\n"
	}

	if shouldPrintOutput {
		fmt.Printf("  [OK] %s\n", configPath)
	}
	return ""
}

func checkRepoPath(repoPath string, shouldPrintOutput bool) string {
	printOutput(shouldPrintOutput, "", "Repository:")

	if repoPath == "" {
		msg := "  [!!] repo path not set (run 'mns init')"
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg + "\n"
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("  [!!] %s (path not found)", repoPath)
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg + "\n"
	}

	gitDirExists, err := healthcheck.GitDirExists(repoPath)
	if err != nil {
		msg := fmt.Sprintf("  [..] %s (git check failed: %v)", repoPath, err)
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg + "\n"
	}

	if !gitDirExists {
		msg := fmt.Sprintf("  [..] %s (not a git repository)", repoPath)
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg + "\n"
	}

	if shouldPrintOutput {
		fmt.Printf("  [OK] %s\n", repoPath)
	}
	return ""
}

func checkDbPath(dbPath string, shouldPrintOutput bool) string {
	printOutput(shouldPrintOutput, "", "Database:")

	if dbPath == "" {
		msg := "  [!!] db path not set (run 'mns init')"
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg + "\n"
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("  [..] %s (not found on disk — first backup will create it)", dbPath)
		if shouldPrintOutput {
			fmt.Println(msg)
		}
		return msg + "\n"
	}

	if shouldPrintOutput {
		fmt.Printf("  [OK] %s\n", dbPath)
	}
	return ""
}

func printIf(should bool, msg string) {
	if should {
		fmt.Println(msg)
	}
}

func CheckBinary(binaryName string, isOptional bool, shouldPrintOutput bool) string {
	result := healthcheck.CheckBinary(binaryName)

	if !result.Found {
		if !isOptional {
			msg := fmt.Sprintf("  [!!] %s (required - not found in PATH)", binaryName)
			printIf(shouldPrintOutput, msg)
			return msg
		}
		msg := fmt.Sprintf("  [..] %s (optional - not found in PATH)", binaryName)
		printIf(shouldPrintOutput, msg)
		return msg
	}

	msg := fmt.Sprintf("  [OK] %s at %s", binaryName, result.Path)
	printIf(shouldPrintOutput, msg)

	if result.Error != nil {
		var warnMsg string
		if result.ExitCode != 0 {
			warnMsg = fmt.Sprintf("      [..] version check: exit %d (%s)", result.ExitCode, result.Version)
		} else {
			warnMsg = fmt.Sprintf("      [..] version check: %v", result.Error)
		}
		printIf(shouldPrintOutput, warnMsg)
		return msg + "\n" + warnMsg
	}

	if result.Version != "" {
		printIf(shouldPrintOutput, fmt.Sprintf("      %s", result.Version))
	}

	return ""
}

func init() {
	RootCmd.AddCommand(healthCmd)
}
