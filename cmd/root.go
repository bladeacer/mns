package cmd

import (
	"fmt"
	"github.com/bladeacer/mmsync/config"
	"github.com/spf13/cobra"
	"os"
)

var dataStore *config.DataStore
var appConf *config.MnemoConf
var versionFlag bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mns",
	Short: "A CLI tool that lets you add folders to backup manually to a target Git repository.",

	Long: `mns - short for mnemosync.

mnemosync is a CLI tool that lets you add folders to backup manually to a target Git repository.
The name is inspired by the Greek Goddess of memory Mnemosyne.

This application assumes that you know how to create and set up a Git repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			schema_ver := appConf.ConfigSchema.AppVersion
			fmt.Printf("mnemosync %s\n", schema_ver)
			return
		}
		// Your original root command logic goes here
		// e.g., print help or run default action
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(cfg *config.MnemoConf, data *config.DataStore) {
	appConf = cfg
	dataStore = data
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mmsync.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Gets the version of mnemosync running")
}
