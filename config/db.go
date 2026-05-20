package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bladeacer/mns/internal/fileio"
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

	repaired, err := ValidateDataStoreSchema(tempDS)
	if err != nil {
		return nil, fmt.Errorf("database at %s failed schema validation and cannot be repaired: %w", dbPath, err)
	}

	if repaired {
		if saveErr := tempDS.SaveData(dbPath); saveErr != nil {
			return nil, fmt.Errorf("failed to persist repaired database: %w", saveErr)
		}
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

	tmpPath := targetPath + ".tmp"
	if err := os.WriteFile(tmpPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON data to temp file %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("failed to rename temp file %s to %s: %w", tmpPath, targetPath, err)
	}
	return nil
}

func ValidateDataStoreSchema(ds *DataStore) (repaired bool, err error) {
	if ds.CurrentId < 0 {
		return false, fmt.Errorf("current_id cannot be negative; found: %d", ds.CurrentId)
	}
	if ds.TrackedDirs == nil {
		return false, fmt.Errorf("required field 'tracked_dirs' is missing from the database schema")
	}

	removed := 0
	seenTargetPaths := make(map[string]struct{})
	seenAliases := make(map[string]struct{})

	for id, data := range ds.TrackedDirs {
		if data.TargetPath == "" {
			fmt.Fprintf(os.Stderr, "Warning: removing entry '%s' with missing target_path\n", id)
			delete(ds.TrackedDirs, id)
			removed++
			continue
		}
		if data.Alias == "" {
			fmt.Fprintf(os.Stderr, "Warning: removing entry '%s' with missing alias\n", id)
			delete(ds.TrackedDirs, id)
			removed++
			continue
		}

		if _, exists := seenTargetPaths[data.TargetPath]; exists {
			fmt.Fprintf(os.Stderr, "Warning: removing duplicate target_path entry '%s': '%s'\n", id, data.TargetPath)
			delete(ds.TrackedDirs, id)
			removed++
			continue
		}
		seenTargetPaths[data.TargetPath] = struct{}{}

		if _, exists := seenAliases[data.Alias]; exists {
			fmt.Fprintf(os.Stderr, "Warning: removing duplicate alias entry '%s': '%s'\n", id, data.Alias)
			delete(ds.TrackedDirs, id)
			removed++
			continue
		}
		seenAliases[data.Alias] = struct{}{}
	}

	if removed > 0 {
		fmt.Fprintf(os.Stderr, "Warning: removed %d corrupt/duplicate entries from database\n", removed)
		return true, nil
	}

	return false, nil
}
