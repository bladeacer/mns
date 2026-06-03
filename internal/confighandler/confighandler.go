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

	versionUpdated := updateAppVersion(tempCfg, defaultCfg)
	oldDbPath := tempCfg.ConfigSchema.DbPath
	healed, warnings := healConfigSchema(tempCfg, defaultCfg)

	if healed && len(warnings) > 0 {
		printWarnings("Configuration Healing Performed", warnings)
		if err := migrateDbFile(oldDbPath, tempCfg.ConfigSchema.DbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Config Warning: %v\n", err)
		}
		if err := saveConfig(tempCfg, configPath, data, "Saving Repaired Configuration"); err != nil {
			return nil, fmt.Errorf("critical error: failed to save repaired configuration: %w", err)
		}
		return tempCfg, nil
	}

	printWarnings("", warnings)

	if needsSchemaUpdate(data) {
		if err := saveConfig(tempCfg, configPath, data, "Updating configuration file with new schema fields"); err != nil {
			return nil, fmt.Errorf("critical error: failed to save updated configuration: %w", err)
		}
		return tempCfg, nil
	}

	if versionUpdated {
		if err := saveConfig(tempCfg, configPath, data, "Updating configuration file with current version"); err != nil {
			return nil, fmt.Errorf("critical error: failed to save updated configuration: %w", err)
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

	updateAppVersion(tempCfg, defaultCfg)

	oldDbPath := tempCfg.ConfigSchema.DbPath
	healed, warnings := healConfigSchema(tempCfg, defaultCfg)

	if healed && len(warnings) > 0 {
		printWarnings("Configuration Healing Performed", warnings)
	} else {
		printWarnings("", warnings)
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

func updateAppVersion(loaded, defaultCfg *config.MnemoConf) bool {
	if loaded.ConfigSchema.AppVersion == defaultCfg.ConfigSchema.AppVersion {
		return false
	}
	old := loaded.ConfigSchema.AppVersion
	loaded.ConfigSchema.AppVersion = defaultCfg.ConfigSchema.AppVersion
	fmt.Fprintf(os.Stderr, "Config Warning: AppVersion updated from '%s' to '%s'\n", old, defaultCfg.ConfigSchema.AppVersion)
	return true
}

func printWarnings(header string, warnings []error) {
	if header != "" {
		fmt.Fprintf(os.Stderr, "%s \n", header)
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Config Warning: %v\n", w)
	}
}

func saveConfig(cfg *config.MnemoConf, configPath string, data []byte, message string) error {
	fmt.Fprintf(os.Stderr, "%s\n\n", message)
	return yamlwrapper.MergeAndSaveConfig(cfg, configPath, data)
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

type healRule struct {
	check func(loaded config.ConfigSchema) (string, bool)
	heal  func(loaded *config.ConfigSchema, defaultVal config.ConfigSchema)
}

func healConfigSchema(loadedCfg *config.MnemoConf, defaultCfg *config.MnemoConf) (bool, []error) {
	warnings := make([]error, 0)
	healed := false

	loaded := &loadedCfg.ConfigSchema
	default_ := defaultCfg.ConfigSchema

	if !loaded.IsInit {
		warnings = append(warnings, fmt.Errorf("configuration is not initialized; run 'mns init' first"))
		return healed, warnings
	}

	rules := []healRule{
		{
			check: func(l config.ConfigSchema) (string, bool) {
				if l.RepoPath == "" {
					return fmt.Sprintf("invalid or empty field 'RepoPath': Cannot be empty when initialized. Overridden with default: '%s'", default_.RepoPath), true
				}
				return "", false
			},
			heal: func(l *config.ConfigSchema, d config.ConfigSchema) { l.RepoPath = d.RepoPath },
		},
		{
			check: func(l config.ConfigSchema) (string, bool) {
				if l.DbPath == "" {
					return fmt.Sprintf("invalid or empty field 'DbPath': Cannot be empty when initialized. Overridden with default: '%s'", default_.DbPath), true
				}
				if l.DbPath != default_.DbPath {
					return fmt.Sprintf("invalid or empty field 'DbPath': Migrating to expected database path. Overridden with default: '%s'", default_.DbPath), true
				}
				return "", false
			},
			heal: func(l *config.ConfigSchema, d config.ConfigSchema) { l.DbPath = d.DbPath },
		},
		{
			check: func(l config.ConfigSchema) (string, bool) {
				if l.Archiver == "" {
					return fmt.Sprintf("invalid or empty field 'Archiver': Empty or invalid. Overridden with default: '%s'", default_.Archiver), true
				}
				return "", false
			},
			heal: func(l *config.ConfigSchema, d config.ConfigSchema) { l.Archiver = d.Archiver },
		},
		{
			check: func(l config.ConfigSchema) (string, bool) {
				if l.CommitFmt == "" {
					return fmt.Sprintf("invalid or empty field 'CommitFmt': Empty or invalid. Overridden with default: '%s'", default_.CommitFmt), true
				}
				return "", false
			},
			heal: func(l *config.ConfigSchema, d config.ConfigSchema) { l.CommitFmt = d.CommitFmt },
		},
		{
			check: func(l config.ConfigSchema) (string, bool) {
				if l.HistLimitDays < 0 {
					return fmt.Sprintf("invalid HistLimitDays: %d. Reset to default: %d", l.HistLimitDays, default_.HistLimitDays), true
				}
				return "", false
			},
			heal: func(l *config.ConfigSchema, d config.ConfigSchema) { l.HistLimitDays = d.HistLimitDays },
		},
		{
			check: func(l config.ConfigSchema) (string, bool) {
				if l.HistLimitSizeMb < 0 {
					return fmt.Sprintf("invalid HistLimitSizeMb: %d. Reset to default: %d", l.HistLimitSizeMb, default_.HistLimitSizeMb), true
				}
				return "", false
			},
			heal: func(l *config.ConfigSchema, d config.ConfigSchema) { l.HistLimitSizeMb = d.HistLimitSizeMb },
		},
		{
			check: func(l config.ConfigSchema) (string, bool) {
				if l.KeepArchives < 0 {
					return fmt.Sprintf("invalid KeepArchives: %d. Reset to default: %d", l.KeepArchives, default_.KeepArchives), true
				}
				return "", false
			},
			heal: func(l *config.ConfigSchema, d config.ConfigSchema) { l.KeepArchives = d.KeepArchives },
		},
		{
			check: func(l config.ConfigSchema) (string, bool) {
				if l.LfsThresholdMb < 0 {
					return fmt.Sprintf("invalid LfsThresholdMb: %d. Reset to default: %d", l.LfsThresholdMb, default_.LfsThresholdMb), true
				}
				return "", false
			},
			heal: func(l *config.ConfigSchema, d config.ConfigSchema) { l.LfsThresholdMb = d.LfsThresholdMb },
		},
	}

	for _, rule := range rules {
		if msg, needsHeal := rule.check(*loaded); needsHeal {
			rule.heal(loaded, default_)
			healed = true
			warnings = append(warnings, fmt.Errorf(msg))
		}
	}

	if loaded.RepoPath != "" {
		if _, err := os.Stat(loaded.RepoPath); os.IsNotExist(err) {
			warnings = append(warnings, fmt.Errorf("RepoPath path does not exist on disk: %s", loaded.RepoPath))
		}
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
		return nil
	}
	if err := fileio.CopyFile(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to migrate database file from '%s' to '%s': %w", oldPath, newPath, err)
	}
	fmt.Fprintf(os.Stderr, "Database file migrated from '%s' to '%s'\n", oldPath, newPath)
	return nil
}
