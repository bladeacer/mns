package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/mns/config"
)

func writeConfigDir(t *testing.T) (dir string, cleanup func()) {
	t.Helper()
	dir = t.TempDir()

	prevMMSync := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", dir)

	backupDir := filepath.Join(dir, ".backup-mmsync")
	homeDir, _ := os.UserHomeDir()
	oldConfigDir := filepath.Join(homeDir, ".config/mmsync")
	if _, err := os.Stat(oldConfigDir); err == nil {
		_ = os.Rename(oldConfigDir, backupDir)
	}

	cleanup = func() {
		_ = os.Setenv("MMSYNC_CONF", prevMMSync)
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(oldConfigDir)
			_ = os.Rename(backupDir, oldConfigDir)
		}
	}
	return
}

func TestGetDataStore_ReturnsEmptyStore(t *testing.T) {
	ds := config.GetDataStore()
	if ds == nil {
		t.Fatal("expected non-nil data store")
	}
	if ds.CurrentId != 0 {
		t.Errorf("expected CurrentId 0, got %d", ds.CurrentId)
	}
	if len(ds.TrackedDirs) != 0 {
		t.Errorf("expected empty TrackedDirs, got %d entries", len(ds.TrackedDirs))
	}
	if len(ds.StagingHistory) != 0 {
		t.Errorf("expected empty StagingHistory, got %d entries", len(ds.StagingHistory))
	}
}

func TestDataStore_AddDir(t *testing.T) {
	ds := config.GetDataStore()
	id := ds.AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "test"})
	if id != "1" {
		t.Errorf("expected id '1', got '%s'", id)
	}
	if ds.CurrentId != 1 {
		t.Errorf("expected CurrentId 1, got %d", ds.CurrentId)
	}
	if len(ds.TrackedDirs) != 1 {
		t.Errorf("expected 1 tracked dir, got %d", len(ds.TrackedDirs))
	}

	id2 := ds.AddDir(config.DirData{TargetPath: "/tmp/test2", Alias: "test2"})
	if id2 != "2" {
		t.Errorf("expected id '2', got '%s'", id2)
	}
}

func TestDataStore_DeleteDir(t *testing.T) {
	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "test"})

	if !ds.DeleteDir("1") {
		t.Error("expected DeleteDir to return true")
	}
	if len(ds.TrackedDirs) != 0 {
		t.Errorf("expected 0 tracked dirs after delete, got %d", len(ds.TrackedDirs))
	}

	if ds.DeleteDir("nonexistent") {
		t.Error("expected DeleteDir to return false for non-existent id")
	}
}

func TestDataStore_FindDirByAlias(t *testing.T) {
	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "myalias"})

	id, entry, found := ds.FindDirByAlias("myalias")
	if !found {
		t.Fatal("expected to find dir by alias")
	}
	if id != "1" {
		t.Errorf("expected id '1', got '%s'", id)
	}
	if entry.TargetPath != "/tmp/test" {
		t.Errorf("expected TargetPath '/tmp/test', got '%s'", entry.TargetPath)
	}

	_, _, found = ds.FindDirByAlias("nonexistent")
	if found {
		t.Error("expected not to find non-existent alias")
	}
}

func TestDataStore_FindDirByPath(t *testing.T) {
	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "myalias"})

	id, entry, found := ds.FindDirByPath("/tmp/test")
	if !found {
		t.Fatal("expected to find dir by path")
	}
	if id != "1" {
		t.Errorf("expected id '1', got '%s'", id)
	}
	if entry.Alias != "myalias" {
		t.Errorf("expected Alias 'myalias', got '%s'", entry.Alias)
	}

	_, _, found = ds.FindDirByPath("/nonexistent")
	if found {
		t.Error("expected not to find non-existent path")
	}
}

func TestDataStore_AddHistory(t *testing.T) {
	ds := config.GetDataStore()
	entry := config.StagingHistoryEntry{
		Timestamp: "2024-01-01T00:00:00Z",
		Archive:   "backup.tar.gz",
		SizeBytes: 1024,
		Dirs:      []string{"dir1"},
	}
	ds.AddHistory(entry)

	if len(ds.StagingHistory) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(ds.StagingHistory))
	}
	if ds.StagingHistory[0].Archive != "backup.tar.gz" {
		t.Errorf("expected archive 'backup.tar.gz', got '%s'", ds.StagingHistory[0].Archive)
	}
}

func TestDataStore_ClearHistory(t *testing.T) {
	ds := config.GetDataStore()
	ds.AddHistory(config.StagingHistoryEntry{Timestamp: "t1", Archive: "a1"})
	ds.ClearHistory()

	if len(ds.StagingHistory) != 0 {
		t.Errorf("expected 0 history entries after clear, got %d", len(ds.StagingHistory))
	}
}

func TestDataStore_SaveData(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "state.json")

	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "test"})

	if err := ds.SaveData(dbPath); err != nil {
		t.Fatalf("unexpected error saving data: %v", err)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected db file to be created")
	}
}

func TestDataStore_SaveData_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "subdir", "state.json")

	ds := config.GetDataStore()
	if err := ds.SaveData(dbPath); err != nil {
		t.Fatalf("unexpected error saving data to nested path: %v", err)
	}
}

func TestDataStore_SaveData_WriteError(t *testing.T) {
	dir := t.TempDir()

	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "test"})

	err := ds.SaveData(dir)
	if err == nil {
		t.Error("expected error when target path is a directory")
	}
}

func TestDataStore_SaveData_MkdirAllError(t *testing.T) {
	dir := t.TempDir()

	blockPath := filepath.Join(dir, "block")
	if err := os.WriteFile(blockPath, []byte("block"), 0644); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(blockPath, "subdir", "state.json")

	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "test"})

	err := ds.SaveData(dbPath)
	if err == nil {
		t.Error("expected error when mkdir is blocked by a file")
	}
}

func TestLoadDataStore_NoFile(t *testing.T) {
	ds, err := config.LoadDataStore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Fatal("expected non-nil data store")
	}
}

func TestLoadDataStore_ValidFile(t *testing.T) {
	dir, cleanup := writeConfigDir(t)
	defer cleanup()

	dbPath := filepath.Join(dir, "mmsync-state.json")
	content := `{"current_id": 3, "tracked_dirs": {"1": {"target_path": "/tmp/test", "alias": "test"}}, "staging_history": []}`
	if err := os.WriteFile(dbPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ds, err := config.LoadDataStore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds.CurrentId != 3 {
		t.Errorf("expected CurrentId 3, got %d", ds.CurrentId)
	}
	if len(ds.TrackedDirs) != 1 {
		t.Errorf("expected 1 tracked dir, got %d", len(ds.TrackedDirs))
	}
}

func TestLoadDataStore_InvalidJSON(t *testing.T) {
	dir, cleanup := writeConfigDir(t)
	defer cleanup()

	dbPath := filepath.Join(dir, "mmsync-state.json")
	if err := os.WriteFile(dbPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.LoadDataStore()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadDataStore_CorruptData_Repairs(t *testing.T) {
	dir, cleanup := writeConfigDir(t)
	defer cleanup()

	dbPath := filepath.Join(dir, "mmsync-state.json")
	corruptContent := `{"current_id": -1, "tracked_dirs": null, "staging_history": []}`
	if err := os.WriteFile(dbPath, []byte(corruptContent), 0644); err != nil {
		t.Fatal(err)
	}

	ds, err := config.LoadDataStore()
	if err != nil {
		t.Fatalf("expected repair to succeed, got: %v", err)
	}
	if ds.CurrentId != 0 {
		t.Errorf("expected CurrentId reset to 0 after repair, got %d", ds.CurrentId)
	}
	if len(ds.TrackedDirs) != 0 {
		t.Errorf("expected TrackedDirs to be empty after repair, got %d", len(ds.TrackedDirs))
	}

	repairedContent, _ := os.ReadFile(dbPath)
	if len(repairedContent) == 0 {
		t.Error("expected repaired file to contain data")
	}
}
