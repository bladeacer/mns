package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of mnemosync",
	Run: func(cmd *cobra.Command, args []string) {
		schemaVer := AppConf.ConfigSchema.AppVersion
		fmt.Printf("mnemosync %s\n", schemaVer)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
