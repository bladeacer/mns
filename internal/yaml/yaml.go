package yaml

import (
	"fmt"
	"os"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/fileio"
	"gopkg.in/yaml.v3"
)

func deepMergeMaps(original, updated map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(original))
	for k, v := range original {
		result[k] = v
	}
	for k, v := range updated {
		if origVal, ok := original[k]; ok {
			if origMap, okOrig := origVal.(map[string]interface{}); okOrig {
				if updMap, okUpd := v.(map[string]interface{}); okUpd {
					result[k] = deepMergeMaps(origMap, updMap)
					continue
				}
			}
		}
		result[k] = v
	}
	return result
}

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

func MergeAndSaveConfig(cfg *config.MnemoConf, targetPath string, originalData []byte) error {
	newData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var original map[string]interface{}
	if err := yaml.Unmarshal(originalData, &original); err != nil {
		return SaveConfig(cfg, targetPath)
	}

	var updated map[string]interface{}
	if err := yaml.Unmarshal(newData, &updated); err != nil {
		return SaveConfig(cfg, targetPath)
	}

	merged := deepMergeMaps(original, updated)

	mergedData, err := yaml.Marshal(merged)
	if err != nil {
		return fmt.Errorf("failed to marshal merged config: %w", err)
	}

	return fileio.AtomicWriteFile(targetPath, mergedData, 0644)
}
