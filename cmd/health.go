package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checks the health of mnemosync",
	Long: `Checks the health of mnemosync
Checks if the required system binaries are installed

Also checks if the mnemosync configuration files have been created.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunHealthCheck(true)
	},
}

func RunHealthCheck(shouldPrintOutput bool) string {
	var errStrBuilder strings.Builder
	separator := "_"
	repeatedSeparator := strings.Repeat(separator, 72)

	configPath := appConf.ConfigSchema.ConfigPath
	repoPath := appConf.ConfigSchema.RepoPath
	dbPath := appConf.ConfigSchema.DbPath

	fmt.Println("\n\tRunning Health Check")

	if err := checkBinWrapper("git", false); err != "" {
		errStrBuilder.WriteString(err)
	}
	if err := checkBinWrapper("rsync", false); err != "" {
		errStrBuilder.WriteString(err)
	}
	if err := checkBinWrapper("tar", false); err != "" {
		errStrBuilder.WriteString(err)
	}
	if err := checkBinWrapper("zip", true); err != "" {
		errStrBuilder.WriteString(err)
	}

	fmt.Printf("\t%s\n\n", repeatedSeparator)

	fmt.Println("\tConfiguration File:")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("\t\t[NOT FOUND] Configuration file not found at:\n\t\t%s\n\t\tRun 'mns init' to start.\n", configPath)
		errStrBuilder.WriteString(msg)
		fmt.Print(msg)
	} else {
		fmt.Printf("\t\t[FOUND] at %s\n", configPath)
	}
	fmt.Printf("\t%s\n", repeatedSeparator)

	fmt.Println("\tRepository Path:")
	if repoPath == "" {
		msg := "\t\t[NOT SET] Repository Path is not defined.\n\t\tRun 'mns init' to set.\n"
		errStrBuilder.WriteString(msg)
		fmt.Print(msg)
	} else {
		fmt.Printf("\t\t[SET] %s\n", repoPath)

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			msg := fmt.Sprintf("\t\t[WARNING] Repository directory does not exist on disk: %s\n", repoPath)
			errStrBuilder.WriteString(msg)
			fmt.Print(msg)
		}
	}
	fmt.Printf("\t%s\n", repeatedSeparator)

	fmt.Println("\tDatabase Path:")
	if dbPath == "" {
		msg := "\t\t[NOT SET] Database Path is not defined.\n\t\tRun 'mns init' to start.\n"
		errStrBuilder.WriteString(msg)
		fmt.Print(msg)
	} else {
		fmt.Printf("\t\t[SET] %s\n", dbPath)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("\t\t[WARNING] Database file not found on disk: %s\n", dbPath)
		errStrBuilder.WriteString(msg)
		fmt.Print(msg)
	}

	fmt.Printf("\t%s\n", repeatedSeparator)
	fmt.Println("\n\tHealth Check Complete")

	return errStrBuilder.String()
}

func checkBinWrapper(binaryName string, isOptional bool) string {
	var msgBuilder strings.Builder

	path, err := exec.LookPath(binaryName)

	if err != nil {
		if !isOptional {
			return fmt.Sprintf("[FAIL] Required Binary '%s' not found in PATH.", binaryName)
		}
		return fmt.Sprintf("[WARNING] Optional Binary '%s' not found in PATH.", binaryName)
	}

	msgBuilder.WriteString(fmt.Sprintf("[PASS] Binary '%s' found at: %s\n", binaryName, path))

	cmd := exec.Command(binaryName, "--version")
	output, versionErr := cmd.CombinedOutput()

	if versionErr != nil {
		msgBuilder.WriteString(fmt.Sprintf("\t[WARNING] Version check failed for '%s'. ", binaryName))

		if exitError, ok := versionErr.(*exec.ExitError); ok {
			msgBuilder.WriteString(fmt.Sprintf("Exit Code %d. Output:\n\t\t%s\n", exitError.ExitCode(), strings.TrimSpace(string(output))))
		} else {
			msgBuilder.WriteString(fmt.Sprintf("Failed to execute: %v\n", versionErr))
		}

		return msgBuilder.String()
	}

	versionLine := strings.SplitN(string(output), "\n", 2)[0]
	msgBuilder.WriteString(fmt.Sprintf("\tVersion: %s\n", strings.TrimSpace(versionLine)))

	return msgBuilder.String()
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

// Here you will define your flags and configuration settings.

// Cobra supports Persistent Flags which will work for this command
// and all subcommands, e.g.:
// healthCmd.PersistentFlags().String("foo", "", "A help for foo")

// Cobra supports local flags which will only run when this command
// is called directly, e.g.:
// healthCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
