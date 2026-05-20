package confighandler

import (
	"fmt"
	"os"

	"github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/fileio"
	yamlwrapper "github.com/bladeacer/mmsync/internal/yaml"
	"gopkg.in/yaml.v3"
)


func LoadConfig() (*config.MnemoConf, error) {
	configPath := fileio.ResolveConfigPath()

	if err := fileio.MigrateConfigData(configPath); err != nil {
		return nil, fmt.Errorf("Configuration migration failed: %w", err)
	}
	defaultCfg := config.GetMnemoConf()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultCfg, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	tempCfg := config.GetMnemoConf()

	if err := yaml.Unmarshal(data, tempCfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML data. File may be invalid: %w", err)
	}

	warnings := healConfigSchema(tempCfg, defaultCfg)

	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Configuration Healing Performed \n")
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "Config Warning: %v\n", w)
		}
		fmt.Fprintf(os.Stderr, "Saving Repaired Configuration \n\n")

		if saveErr := yamlwrapper.SaveConfig(tempCfg, configPath); saveErr != nil {
			return nil, fmt.Errorf("critical error: failed to save repaired configuration: %w", saveErr)
		}
	}

	return tempCfg, nil
}


func healConfigSchema(loadedCfg *config.MnemoConf, defaultCfg *config.MnemoConf) []error {
	warnings := make([]error, 0)

	loadedSchema := &loadedCfg.ConfigSchema
	defaultSchema := defaultCfg.ConfigSchema

	replaceField := func(field *string, defaultVal string, fieldName string, reason string) {
		*field = defaultVal
		warnings = append(warnings, fmt.Errorf("invalid or empty field '%s': %s Overridden with default: '%s'", fieldName, reason, defaultVal))
	}

	if loadedSchema.AppVersion != defaultSchema.AppVersion {
		replaceField(&loadedSchema.AppVersion, defaultSchema.AppVersion, "AppVersion", "")
	}

	if !loadedSchema.IsInit {
		warnings = append(warnings, fmt.Errorf("found configuration file marked IsInit=false. Resetting RepoPath/DbPath."))

		loadedSchema.RepoPath = defaultSchema.RepoPath
		loadedSchema.DbPath = defaultSchema.DbPath
		return warnings
	}

	if loadedSchema.RepoPath == "" {
		replaceField(&loadedSchema.RepoPath, defaultSchema.RepoPath, "RepoPath", "Cannot be empty when initialized")
	} else if _, err := os.Stat(loadedSchema.RepoPath); os.IsNotExist(err) {
		replaceField(&loadedSchema.RepoPath, defaultSchema.RepoPath, "RepoPath", fmt.Sprintf("Path does not exist on disk: %s", loadedSchema.RepoPath))
	}

	if loadedSchema.DbPath == "" {
		replaceField(&loadedSchema.DbPath, defaultSchema.DbPath, "DbPath", "Cannot be empty when initialized")
	}

	if loadedSchema.ConfigPath == "" {
		replaceField(&loadedSchema.ConfigPath, defaultSchema.ConfigPath, "ConfigPath", fmt.Sprintf("File path mismatch: %s", loadedSchema.ConfigPath))
	}

	if loadedSchema.Archiver == "" {
		replaceField(&loadedSchema.Archiver, defaultSchema.Archiver, "Archiver", "Empty or invalid")
	}
	if loadedSchema.CommitFmt == "" {
		replaceField(&loadedSchema.CommitFmt, defaultSchema.CommitFmt, "CommitFmt", "Empty or invalid")
	}
	if loadedSchema.HistLimitDays <= 0 {
		loadedSchema.HistLimitDays = defaultSchema.HistLimitDays
		warnings = append(warnings, fmt.Errorf("invalid HistLimitDays: %d. Reset to default: %d", loadedSchema.HistLimitDays, defaultSchema.HistLimitDays))
	}
	if loadedSchema.HistLimitSizeMb <= 0 {
		loadedSchema.HistLimitSizeMb = defaultSchema.HistLimitSizeMb
		warnings = append(warnings, fmt.Errorf("invalid HistLimitSizeMb: %d. Reset to default: %d", loadedSchema.HistLimitSizeMb, defaultSchema.HistLimitSizeMb))
	}

	return warnings
}
