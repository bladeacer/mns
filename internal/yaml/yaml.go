package yaml

import (
	"fmt"
	"os"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/fileio"
	"gopkg.in/yaml.v3"
)

func WriteYAML(defaultConfig *config.MnemoConf, configPath string) {
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating default config:", err)
		return
	}

	if err := fileio.AtomicWriteFile(configPath, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing config file:", err)
		return
	}

	fmt.Printf("Initialized default configuration file at %s\n", configPath)
}

func SaveConfig(cfg *config.MnemoConf, targetPath string) error {
	jsonData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal MnemoConf to YAML: %w", err)
	}

	if err := fileio.AtomicWriteFile(targetPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write YAML data to file %s: %w", targetPath, err)
	}
	return nil
}

// func UnmarshalWrapper(data File, ) (error) {
// 	if err := yaml.Unmarshal(data, cfg); err != nil {
// 		fmt.Errorf("error unmarshalling YAML data. File may be invalid: %w", err)
// 	}

// 	return nil
// }
