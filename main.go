package main

import (
	"fmt"
	"os"

	"github.com/bladeacer/mmsync/cmd"
	"github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/confighandler"
)

func main() {
	euid := os.Geteuid()

	// Check if the effective user ID is 0 (which indicates root).
	if euid == 0 {
		fmt.Println("Warning: mnemosync should not be run as root.")
		os.Exit(1)
	}

	appConfig, err := confighandler.LoadConfig()
	dataStore, err2 := config.LoadDataStore()

	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if err2 != nil {
		fmt.Printf("Error loading database: %v\n", err2)
		os.Exit(1)
	}

	cmd.Execute(appConfig, dataStore)
}
