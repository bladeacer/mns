package config

import (
	"testing"
)

func TestValidateDataStoreSchema_NegativeID(t *testing.T) {
	ds := &DataStore{
		CurrentId:   -1,
		TrackedDirs: make(map[string]DirData),
	}
	err := validateDataStoreSchema(ds)
	if err == nil {
		t.Error("expected error for negative CurrentId")
	}
}

func TestValidateDataStoreSchema_NilTrackedDirs(t *testing.T) {
	ds := &DataStore{
		CurrentId:   0,
		TrackedDirs: nil,
	}
	err := validateDataStoreSchema(ds)
	if err == nil {
		t.Error("expected error for nil TrackedDirs")
	}
}

func TestValidateDataStoreSchema_MissingTargetPath(t *testing.T) {
	ds := &DataStore{
		CurrentId: 1,
		TrackedDirs: map[string]DirData{
			"1": {TargetPath: "", Alias: "test"},
		},
	}
	err := validateDataStoreSchema(ds)
	if err == nil {
		t.Error("expected error for missing target_path")
	}
}

func TestValidateDataStoreSchema_MissingAlias(t *testing.T) {
	ds := &DataStore{
		CurrentId: 1,
		TrackedDirs: map[string]DirData{
			"1": {TargetPath: "/tmp/test", Alias: ""},
		},
	}
	err := validateDataStoreSchema(ds)
	if err == nil {
		t.Error("expected error for missing alias")
	}
}

func TestValidateDataStoreSchema_DuplicateTargetPath(t *testing.T) {
	ds := &DataStore{
		CurrentId: 2,
		TrackedDirs: map[string]DirData{
			"1": {TargetPath: "/tmp/test", Alias: "alias1"},
			"2": {TargetPath: "/tmp/test", Alias: "alias2"},
		},
	}
	err := validateDataStoreSchema(ds)
	if err == nil {
		t.Error("expected error for duplicate target_path")
	}
}

func TestValidateDataStoreSchema_DuplicateAlias(t *testing.T) {
	ds := &DataStore{
		CurrentId: 2,
		TrackedDirs: map[string]DirData{
			"1": {TargetPath: "/tmp/test1", Alias: "same"},
			"2": {TargetPath: "/tmp/test2", Alias: "same"},
		},
	}
	err := validateDataStoreSchema(ds)
	if err == nil {
		t.Error("expected error for duplicate alias")
	}
}

func TestValidateDataStoreSchema_Valid(t *testing.T) {
	ds := &DataStore{
		CurrentId: 2,
		TrackedDirs: map[string]DirData{
			"1": {TargetPath: "/tmp/test1", Alias: "alias1"},
			"2": {TargetPath: "/tmp/test2", Alias: "alias2"},
		},
	}
	err := validateDataStoreSchema(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
