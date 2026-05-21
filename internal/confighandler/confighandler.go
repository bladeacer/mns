package confighandler

import (
	"fmt"
	"os"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/fileio"
	yamlwrapper "github.com/bladeacer/mns/internal/yaml"
	"gopkg.in/yaml.v3"
)

var knownSchemaFields = map[string]bool{
	"config_path":        true,
	"app_version":        true,
	"is_init":            true,
	"repo_path":          true,
	"db_path":            true,
	"archiver":           true,
	"commit_fmt":         true,
	"respect_gitignore":  true,
	"hist_limit_days":    true,
	"hist_limit_size_mb": true,
	"keep_archives":      true,
	"lfs_threshold_mb":   true,
}

func LoadConfig() (*config.MnemoConf, error) {
	configPath := fileio.ResolveConfigPath()
	return LoadConfigWithPath(configPath)
}

func LoadConfigWithPath(configPath string) (*config.MnemoConf, error) {
	if err := fileio.MigrateConfigData(configPath); err != nil {
		return nil, fmt.Errorf("configuration migration failed: %w", err)
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

	versionUpdated := false
	if tempCfg.ConfigSchema.AppVersion != defaultCfg.ConfigSchema.AppVersion {
		old := tempCfg.ConfigSchema.AppVersion
		tempCfg.ConfigSchema.AppVersion = defaultCfg.ConfigSchema.AppVersion
		fmt.Fprintf(os.Stderr, "Config Warning: AppVersion updated from '%s' to '%s'\n", old, defaultCfg.ConfigSchema.AppVersion)
		versionUpdated = true
	}

	oldDbPath := tempCfg.ConfigSchema.DbPath
	healed, warnings := healConfigSchema(tempCfg, defaultCfg)

	if healed && len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Configuration Healing Performed \n")
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "Config Warning: %v\n", w)
		}

		if err := migrateDbFile(oldDbPath, tempCfg.ConfigSchema.DbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Config Warning: %v\n", err)
		}

		fmt.Fprintf(os.Stderr, "Saving Repaired Configuration \n\n")

		if saveErr := yamlwrapper.MergeAndSaveConfig(tempCfg, configPath, data); saveErr != nil {
			return nil, fmt.Errorf("critical error: failed to save repaired configuration: %w", saveErr)
		}
		return tempCfg, nil
	}

	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Config Warning: %v\n", w)
	}

	if needsSchemaUpdate(data) {
		fmt.Fprintf(os.Stderr, "Updating configuration file with new schema fields\n")
		if saveErr := yamlwrapper.MergeAndSaveConfig(tempCfg, configPath, data); saveErr != nil {
			return nil, fmt.Errorf("critical error: failed to save updated configuration: %w", saveErr)
		}
		return tempCfg, nil
	}

	if versionUpdated {
		fmt.Fprintf(os.Stderr, "Updating configuration file with current version\n\n")
		if saveErr := yamlwrapper.MergeAndSaveConfig(tempCfg, configPath, data); saveErr != nil {
			return nil, fmt.Errorf("critical error: failed to save updated configuration: %w", saveErr)
		}
	}

	return tempCfg, nil
}

func HealAndSaveConfig(configPath string) (*config.MnemoConf, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := fileio.MigrateConfigData(configPath); err != nil {
		return nil, fmt.Errorf("configuration migration failed: %w", err)
	}
	defaultCfg := config.GetMnemoConf()

	tempCfg := config.GetMnemoConf()
	if err := yaml.Unmarshal(data, tempCfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML data: %w", err)
	}

	// Force AppVersion update
	if tempCfg.ConfigSchema.AppVersion != defaultCfg.ConfigSchema.AppVersion {
		old := tempCfg.ConfigSchema.AppVersion
		tempCfg.ConfigSchema.AppVersion = defaultCfg.ConfigSchema.AppVersion
		fmt.Fprintf(os.Stderr, "Config Warning: AppVersion updated from '%s' to '%s'\n", old, defaultCfg.ConfigSchema.AppVersion)
	}

	oldDbPath := tempCfg.ConfigSchema.DbPath
	healed, warnings := healConfigSchema(tempCfg, defaultCfg)

	if healed && len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Configuration Healing Performed\n")
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "Config Warning: %v\n", w)
		}
	} else {
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "Config Warning: %v\n", w)
		}
	}

	if err := migrateDbFile(oldDbPath, tempCfg.ConfigSchema.DbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Config Warning: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "Saving configuration (--heal forced save)\n\n")
	if err := yamlwrapper.MergeAndSaveConfig(tempCfg, configPath, data); err != nil {
		return nil, fmt.Errorf("failed to save healed configuration: %w", err)
	}

	return tempCfg, nil
}

func needsSchemaUpdate(data []byte) bool {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return false
	}
	schema, ok := raw["config_schema"].(map[string]interface{})
	if !ok {
		return false
	}
	for field := range knownSchemaFields {
		if _, exists := schema[field]; !exists {
			return true
		}
	}
	return false
}

func healConfigSchema(loadedCfg *config.MnemoConf, defaultCfg *config.MnemoConf) (bool, []error) {
	warnings := make([]error, 0)
	healed := false

	loadedSchema := &loadedCfg.ConfigSchema
	defaultSchema := defaultCfg.ConfigSchema

	replaceField := func(field *string, defaultVal string, fieldName string, reason string) {
		*field = defaultVal
		healed = true
		warnings = append(warnings, fmt.Errorf("invalid or empty field '%s': %s Overridden with default: '%s'", fieldName, reason, defaultVal))
	}

	if !loadedSchema.IsInit {
		warnings = append(warnings, fmt.Errorf("configuration is not initialized; run 'mns init' first"))
		return healed, warnings
	}

	if loadedSchema.RepoPath == "" {
		replaceField(&loadedSchema.RepoPath, defaultSchema.RepoPath, "RepoPath", "Cannot be empty when initialized")
	} else if _, err := os.Stat(loadedSchema.RepoPath); os.IsNotExist(err) {
		warnings = append(warnings, fmt.Errorf("RepoPath path does not exist on disk: %s", loadedSchema.RepoPath))
	}

	if loadedSchema.DbPath == "" {
		replaceField(&loadedSchema.DbPath, defaultSchema.DbPath, "DbPath", "Cannot be empty when initialized")
	} else if loadedSchema.DbPath != defaultSchema.DbPath {
		replaceField(&loadedSchema.DbPath, defaultSchema.DbPath, "DbPath", "Migrating to expected database path")
	}

	if loadedSchema.Archiver == "" {
		replaceField(&loadedSchema.Archiver, defaultSchema.Archiver, "Archiver", "Empty or invalid")
	}
	if loadedSchema.CommitFmt == "" {
		replaceField(&loadedSchema.CommitFmt, defaultSchema.CommitFmt, "CommitFmt", "Empty or invalid")
	}
	if loadedSchema.HistLimitDays < 0 {
		loadedSchema.HistLimitDays = defaultSchema.HistLimitDays
		healed = true
		warnings = append(warnings, fmt.Errorf("invalid HistLimitDays: %d. Reset to default: %d", loadedSchema.HistLimitDays, defaultSchema.HistLimitDays))
	}
	if loadedSchema.HistLimitSizeMb < 0 {
		loadedSchema.HistLimitSizeMb = defaultSchema.HistLimitSizeMb
		healed = true
		warnings = append(warnings, fmt.Errorf("invalid HistLimitSizeMb: %d. Reset to default: %d", loadedSchema.HistLimitSizeMb, defaultSchema.HistLimitSizeMb))
	}

	if loadedSchema.KeepArchives < 0 {
		loadedSchema.KeepArchives = defaultSchema.KeepArchives
		healed = true
		warnings = append(warnings, fmt.Errorf("invalid KeepArchives: %d. Reset to default: %d", loadedSchema.KeepArchives, defaultSchema.KeepArchives))
	}

	if loadedSchema.LfsThresholdMb < 0 {
		loadedSchema.LfsThresholdMb = defaultSchema.LfsThresholdMb
		healed = true
		warnings = append(warnings, fmt.Errorf("invalid LfsThresholdMb: %d. Reset to default: %d", loadedSchema.LfsThresholdMb, defaultSchema.LfsThresholdMb))
	}

	return healed, warnings
}

func migrateDbFile(oldPath, newPath string) error {
	if oldPath == "" || oldPath == newPath {
		return nil
	}
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil
	}
	if _, err := os.Stat(newPath); err == nil {
		if err := os.Remove(oldPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove stale database file '%s': %v\n", oldPath, err)
		}
		return nil
	}
	if err := fileio.CopyFile(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to migrate database file from '%s' to '%s': %w", oldPath, newPath, err)
	}
	if err := os.Remove(oldPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove old database file '%s': %v\n", oldPath, err)
	}
	fmt.Fprintf(os.Stderr, "Database file migrated from '%s' to '%s'\n", oldPath, newPath)
	return nil
}
