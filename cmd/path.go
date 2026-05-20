package cmd

import (
	"fmt"
	// "github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/fileio"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the mnemosync configuration file",
	Long:  "Provides commands to manage the application's configuration file.\nSet the configuration path either with $MMSYNC_CONF environment variable in your shell configuration file e.g. bashrc or with mns init.\nConfiguration directory should only be under your home directory.\nRestart your shell when changing or clearing the environment variable.",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := fileio.ResolveConfigPath()
		isInit := appConf.ConfigSchema.IsInit

		if !isInit {
			fmt.Printf("\nConfiguration file not found at expected path\n%s\nRun mns init to start.\n", configPath)
		} else {
			fmt.Printf("\nConfiguration file path:\n%s\n", configPath)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Prints the current configuration to stdout",
	Long:  "Prints the content of the mnemosync configuration file to the standard output. If the file doesn't exist, it prints a message.",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := fileio.ResolveConfigPath()
		isInit := appConf.ConfigSchema.IsInit

		if !isInit {
			fmt.Printf("Error: Configuration file not found at expected path:\n%s\nRun mns init to start.\n", configPath)
			os.Exit(1)
		}

		// Read and print the config file contents
		content, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("Error: failed to read config file at %s: %v\n", configPath, err)
			os.Exit(1)
		}

		fmt.Printf("%s", content)
	},
}

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Opens the configuration file with the user's $EDITOR",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := fileio.ResolveConfigPath()
		isInit := appConf.ConfigSchema.IsInit
		editor := os.Getenv("EDITOR")

		if editor == "" {
			fmt.Println("Error: $EDITOR environment variable not set. Please set it to your preferred text editor (e.g., 'vim', 'code').")
			os.Exit(1)
		}

		if !isInit {
			fmt.Printf("\nConfiguration file not found at expected path\n%s\nRun mns init to start.\n", configPath)
			os.Exit(1)
		}

		editorCmd := exec.Command(editor, configPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		err := editorCmd.Run()
		if err != nil {
			fmt.Printf("Error: failed to open config file with %s: %v\n", editor, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(openCmd)
}
