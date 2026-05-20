package config_test

import (
	"testing"

	"github.com/bladeacer/mns/config"
)

func TestValidateDataStoreSchema_NegativeID(t *testing.T) {
	ds := &config.DataStore{
		CurrentId:   -1,
		TrackedDirs: make(map[string]config.DirData),
	}
	_, err := config.ValidateDataStoreSchema(ds)
	if err == nil {
		t.Error("expected error for negative CurrentId")
	}
}

func TestValidateDataStoreSchema_NilTrackedDirs(t *testing.T) {
	ds := &config.DataStore{
		CurrentId:   0,
		TrackedDirs: nil,
	}
	_, err := config.ValidateDataStoreSchema(ds)
	if err == nil {
		t.Error("expected error for nil TrackedDirs")
	}
}

func TestValidateDataStoreSchema_MissingTargetPath(t *testing.T) {
	ds := &config.DataStore{
		CurrentId: 1,
		TrackedDirs: map[string]config.DirData{
			"1": {TargetPath: "", Alias: "test"},
		},
	}
	repaired, err := config.ValidateDataStoreSchema(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repaired {
		t.Error("expected repair to remove entry with missing target_path")
	}
	if len(ds.TrackedDirs) != 0 {
		t.Errorf("expected 0 entries after removal, got %d", len(ds.TrackedDirs))
	}
}

func TestValidateDataStoreSchema_MissingAlias(t *testing.T) {
	ds := &config.DataStore{
		CurrentId: 1,
		TrackedDirs: map[string]config.DirData{
			"1": {TargetPath: "/tmp/test", Alias: ""},
		},
	}
	repaired, err := config.ValidateDataStoreSchema(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repaired {
		t.Error("expected repair to remove entry with missing alias")
	}
	if len(ds.TrackedDirs) != 0 {
		t.Errorf("expected 0 entries after removal, got %d", len(ds.TrackedDirs))
	}
}

func TestValidateDataStoreSchema_DuplicateTargetPath(t *testing.T) {
	ds := &config.DataStore{
		CurrentId: 2,
		TrackedDirs: map[string]config.DirData{
			"1": {TargetPath: "/tmp/test", Alias: "alias1"},
			"2": {TargetPath: "/tmp/test", Alias: "alias2"},
		},
	}
	repaired, err := config.ValidateDataStoreSchema(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repaired {
		t.Error("expected repair to remove duplicate target_path entry")
	}
	if len(ds.TrackedDirs) != 1 {
		t.Errorf("expected 1 entry after dedup, got %d", len(ds.TrackedDirs))
	}
}

func TestValidateDataStoreSchema_DuplicateAlias(t *testing.T) {
	ds := &config.DataStore{
		CurrentId: 2,
		TrackedDirs: map[string]config.DirData{
			"1": {TargetPath: "/tmp/test1", Alias: "same"},
			"2": {TargetPath: "/tmp/test2", Alias: "same"},
		},
	}
	repaired, err := config.ValidateDataStoreSchema(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repaired {
		t.Error("expected repair to remove duplicate alias entry")
	}
	if len(ds.TrackedDirs) != 1 {
		t.Errorf("expected 1 entry after dedup, got %d", len(ds.TrackedDirs))
	}
}

func TestValidateDataStoreSchema_Valid(t *testing.T) {
	ds := &config.DataStore{
		CurrentId: 2,
		TrackedDirs: map[string]config.DirData{
			"1": {TargetPath: "/tmp/test1", Alias: "alias1"},
			"2": {TargetPath: "/tmp/test2", Alias: "alias2"},
		},
	}
	repaired, err := config.ValidateDataStoreSchema(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repaired {
		t.Error("expected no repair for valid data")
	}
}
