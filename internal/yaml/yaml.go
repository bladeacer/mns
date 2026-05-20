package yaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bladeacer/mns/config"
	"gopkg.in/yaml.v3"
)

// Optimise with goroutines in the background?

func WriteYAML(defaultConfig *config.MnemoConf, configPath string) {
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating default config:", err)
		return
	}

	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintln(os.Stderr, "Error creating config directory:", err)
			return
		}
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
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

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory structure for %s: %w", targetPath, err)
	}
	if err := os.WriteFile(targetPath, jsonData, 0644); err != nil {
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
