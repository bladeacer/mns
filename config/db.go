package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bladeacer/mmsync/internal/fileio"
)

type DirData struct {
	TargetPath string `json:"target_path"`
	Alias      string `json:"alias"`
}

type StagingHistoryEntry struct {
	Timestamp string   `json:"timestamp"`
	Archive   string   `json:"archive"`
	SizeBytes int64    `json:"size_bytes"`
	Dirs      []string `json:"dirs"`
}

type DataStore struct {
	CurrentId      int64                 `json:"current_id"`
	TrackedDirs    map[string]DirData    `json:"tracked_dirs"`
	StagingHistory []StagingHistoryEntry `json:"staging_history"`
}

func GetDataStore() *DataStore {
	return &DataStore{
		CurrentId:      0,
		TrackedDirs:    make(map[string]DirData),
		StagingHistory: make([]StagingHistoryEntry, 0),
	}
}
func LoadDataStore() (*DataStore, error) {
	dbPath := fileio.ResolveDbPath()

	defaultDS := GetDataStore()

	data, err := os.ReadFile(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultDS, nil
		}
		return nil, fmt.Errorf("error reading database file %s: %w", dbPath, err)
	}

	tempDS := GetDataStore()

	if err := json.Unmarshal(data, tempDS); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON data from %s. File may be corrupt: %w", dbPath, err)
	}

	if err := ValidateDataStoreSchema(tempDS); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Database at %s failed schema validation: %v. Overwriting with default data.\n", dbPath, err)

		tempDS = defaultDS

		if saveErr := tempDS.SaveData(dbPath); saveErr != nil {
			return nil, fmt.Errorf("critical error: failed to repair and save default data store: %w", saveErr)
		}

		return tempDS, nil
	}

	return tempDS, nil
}

func (ds *DataStore) AddDir(data DirData) string {
	ds.CurrentId += 1

	newIDStr := strconv.FormatInt(ds.CurrentId, 10)
	ds.TrackedDirs[newIDStr] = data
	return newIDStr
}
func (ds *DataStore) DeleteDir(id string) bool {
	if _, exists := ds.TrackedDirs[id]; exists {
		delete(ds.TrackedDirs, id)
		return true
	}
	return false
}

func (ds *DataStore) FindDirByAlias(alias string) (string, DirData, bool) {
	for id, entry := range ds.TrackedDirs {
		if entry.Alias == alias {
			return id, entry, true
		}
	}
	return "", DirData{}, false
}

func (ds *DataStore) FindDirByPath(path string) (string, DirData, bool) {
	for id, entry := range ds.TrackedDirs {
		if entry.TargetPath == path {
			return id, entry, true
		}
	}
	return "", DirData{}, false
}

func (ds *DataStore) AddHistory(entry StagingHistoryEntry) {
	ds.StagingHistory = append(ds.StagingHistory, entry)
}

func (ds *DataStore) ClearHistory() {
	ds.StagingHistory = make([]StagingHistoryEntry, 0)
}

func (ds *DataStore) SaveData(targetPath string) error {
	jsonData, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal DataStore to JSON: %w", err)
	}

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory structure for %s: %w", targetPath, err)
	}
	if err := os.WriteFile(targetPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON data to file %s: %w", targetPath, err)
	}
	return nil
}

func ValidateDataStoreSchema(ds *DataStore) error {
	if ds.CurrentId < 0 {
		return fmt.Errorf("current_id cannot be negative; found: %d", ds.CurrentId)
	}
	if ds.TrackedDirs == nil {
		return fmt.Errorf("required field 'tracked_dirs' is missing from the database schema")
	}

	seenTargetPaths := make(map[string]struct{})
	seenAliases := make(map[string]struct{})

	for id, data := range ds.TrackedDirs {
		if data.TargetPath == "" {
			return fmt.Errorf("entry with ID '%s' is missing a required target_path", id)
		}
		if data.Alias == "" {
			return fmt.Errorf("entry with ID '%s' is missing a required alias", id)
		}

		if _, exists := seenTargetPaths[data.TargetPath]; exists {
			return fmt.Errorf("duplicate target_path found: '%s'", data.TargetPath)
		}
		seenTargetPaths[data.TargetPath] = struct{}{}

		if _, exists := seenAliases[data.Alias]; exists {
			return fmt.Errorf("duplicate alias found: '%s'", data.Alias)
		}
		seenAliases[data.Alias] = struct{}{}
	}

	return nil
}
