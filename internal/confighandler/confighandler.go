package confighandler

import (
	"errors"
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

// type healRule struct {
// 	check func(loaded config.ConfigSchema) (string, bool)
// 	heal  func(loaded *config.ConfigSchema, defaultVal config.ConfigSchema)
// }

func checkEmptyField(val string, name, def string) (string, bool) {
	if val != "" {
		return "", false
	}
	return fmt.Sprintf("invalid or empty field '%s': Cannot be empty when initialized. Overridden with default: '%s'", name, def), true
}

func checkNonNegative(val int, name string, def int) (string, bool) {
	if val >= 0 {
		return "", false
	}
	return fmt.Sprintf("invalid %s: %d. Reset to default: %d", name, val, def), true
}

func checkNonNegative64(val int64, name string, def int64) (string, bool) {
	if val >= 0 {
		return "", false
	}
	return fmt.Sprintf("invalid %s: %d. Reset to default: %d", name, val, def), true
}

func healStringField(field *string, def string) {
	*field = def
}

func healIntField(field *int, def int) {
	*field = def
}

func healInt64Field(field *int64, def int64) {
	*field = def
}

func repoPathRule(l config.ConfigSchema, d config.ConfigSchema) (string, bool) {
	return checkEmptyField(l.RepoPath, "RepoPath", d.RepoPath)
}

func healRepoPath(l *config.ConfigSchema, d config.ConfigSchema) {
	healStringField(&l.RepoPath, d.RepoPath)
}

func dbPathCheck(l config.ConfigSchema, d config.ConfigSchema) (string, bool) {
	if l.DbPath == "" {
		return fmt.Sprintf("invalid or empty field 'DbPath': Cannot be empty when initialized. Overridden with default: '%s'", d.DbPath), true
	}
	if l.DbPath != d.DbPath {
		return fmt.Sprintf("invalid or empty field 'DbPath': Migrating to expected database path. Overridden with default: '%s'", d.DbPath), true
	}
	return "", false
}

func healDbPath(l *config.ConfigSchema, d config.ConfigSchema) {
	healStringField(&l.DbPath, d.DbPath)
}

func archiverRule(l config.ConfigSchema, d config.ConfigSchema) (string, bool) {
	return checkEmptyField(l.Archiver, "Archiver", d.Archiver)
}

func healArchiver(l *config.ConfigSchema, d config.ConfigSchema) {
	healStringField(&l.Archiver, d.Archiver)
}

func commitFmtRule(l config.ConfigSchema, d config.ConfigSchema) (string, bool) {
	return checkEmptyField(l.CommitFmt, "CommitFmt", d.CommitFmt)
}

func healCommitFmt(l *config.ConfigSchema, d config.ConfigSchema) {
	healStringField(&l.CommitFmt, d.CommitFmt)
}

func histLimitDaysRule(l config.ConfigSchema, d config.ConfigSchema) (string, bool) {
	return checkNonNegative(l.HistLimitDays, "HistLimitDays", d.HistLimitDays)
}

func healHistLimitDays(l *config.ConfigSchema, d config.ConfigSchema) {
	healIntField(&l.HistLimitDays, d.HistLimitDays)
}

func histLimitSizeMbRule(l config.ConfigSchema, d config.ConfigSchema) (string, bool) {
	return checkNonNegative64(l.HistLimitSizeMb, "HistLimitSizeMb", d.HistLimitSizeMb)
}

func healHistLimitSizeMb(l *config.ConfigSchema, d config.ConfigSchema) {
	healInt64Field(&l.HistLimitSizeMb, d.HistLimitSizeMb)
}

func keepArchivesRule(l config.ConfigSchema, d config.ConfigSchema) (string, bool) {
	return checkNonNegative(l.KeepArchives, "KeepArchives", d.KeepArchives)
}

func healKeepArchives(l *config.ConfigSchema, d config.ConfigSchema) {
	healIntField(&l.KeepArchives, d.KeepArchives)
}

func lfsThresholdRule(l config.ConfigSchema, d config.ConfigSchema) (string, bool) {
	return checkNonNegative64(l.LfsThresholdMb, "LfsThresholdMb", d.LfsThresholdMb)
}

func healLfsThreshold(l *config.ConfigSchema, d config.ConfigSchema) {
	healInt64Field(&l.LfsThresholdMb, d.LfsThresholdMb)
}

func applyHealRules(loaded *config.ConfigSchema, default_ config.ConfigSchema) (bool, []error) {
	warnings := make([]error, 0)
	healed := false

	type healEntry struct {
		check func(l config.ConfigSchema, d config.ConfigSchema) (string, bool)
		heal  func(l *config.ConfigSchema, d config.ConfigSchema)
	}

	rules := []healEntry{
		{check: repoPathRule, heal: healRepoPath},
		{check: dbPathCheck, heal: healDbPath},
		{check: archiverRule, heal: healArchiver},
		{check: commitFmtRule, heal: healCommitFmt},
		{check: histLimitDaysRule, heal: healHistLimitDays},
		{check: histLimitSizeMbRule, heal: healHistLimitSizeMb},
		{check: keepArchivesRule, heal: healKeepArchives},
		{check: lfsThresholdRule, heal: healLfsThreshold},
	}

	for _, rule := range rules {
		if msg, needsHeal := rule.check(*loaded, default_); needsHeal {
			rule.heal(loaded, default_)
			healed = true
			warnings = append(warnings, errors.New(msg))
		}
	}

	return healed, warnings
}

func checkRepoPathExists(loaded config.ConfigSchema) []error {
	if loaded.RepoPath == "" {
		return nil
	}
	if _, err := os.Stat(loaded.RepoPath); os.IsNotExist(err) {
		return []error{fmt.Errorf("RepoPath path does not exist on disk: %s", loaded.RepoPath)}
	}
	return nil
}

func healConfigSchema(loadedCfg *config.MnemoConf, defaultCfg *config.MnemoConf) (bool, []error) {
	loaded := &loadedCfg.ConfigSchema
	default_ := defaultCfg.ConfigSchema

	if !loaded.IsInit {
		return false, []error{fmt.Errorf("configuration is not initialized; run 'mns init' first")}
	}

	healed, warnings := applyHealRules(loaded, default_)
	warnings = append(warnings, checkRepoPathExists(*loaded)...)
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
