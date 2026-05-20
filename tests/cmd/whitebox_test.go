package cmd_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bladeacer/mmsync/cmd"
	"github.com/bladeacer/mmsync/config"
)

func resetGlobals() {
	cmd.SetAppConf(nil)
	cmd.SetDataStore(nil)
}

func setTestGlobals(dir string) {
	resetGlobals()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath:      filepath.Join(dir, "config.yaml"),
			RepoPath:        dir,
			DbPath:          filepath.Join(dir, "state.json"),
			IsInit:          true,
			Archiver:        "tar",
			HistLimitDays:   7,
			HistLimitSizeMb: 1024,
			KeepArchives:    5,
			LfsThresholdMb:  5,
		},
	})
	cmd.SetDataStore(config.GetDataStore())
}

func TestEnsureGitignoreInDir_AddsEntry(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.EnsureGitignoreInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "/.mnemosync/") {
		t.Errorf("expected .gitignore to contain '/.mnemosync/', got: '%s'", string(content))
	}
}

func TestEnsureGitignoreInDir_AlreadyPresent(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("/.mnemosync/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := cmd.EnsureGitignoreInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(gitignorePath)
	count := strings.Count(string(content), "/.mnemosync/")
	if count != 1 {
		t.Errorf("expected exactly 1 entry, got %d", count)
	}
}

func TestEnsureGitignoreInDir_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*.log\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := cmd.EnsureGitignoreInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(gitignorePath)
	if !strings.Contains(string(content), "/.mnemosync/") {
		t.Error("expected .gitignore to contain '/.mnemosync/'")
	}
	if !strings.Contains(string(content), "*.log") {
		t.Error("expected .gitignore to preserve existing content")
	}
}

func TestEnsureGitignoreInDir_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	if err := cmd.EnsureGitignoreInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(gitignorePath)
	if strings.Count(string(content), "/.mnemosync/") != 1 {
		t.Errorf("expected exactly 1 entry in empty file, got content: '%s'", string(content))
	}
}

func TestEnsureGitignoreInDir_ReadError(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("test"), 0000); err != nil {
		t.Fatal(err)
	}
	_ = os.Chmod(gitignorePath, 0000)
	defer func() { _ = os.Chmod(gitignorePath, 0644) }()

	err := cmd.EnsureGitignoreInDir(dir)
	if err == nil {
		t.Error("expected error when .gitignore is not readable")
	}
}

func TestStagingDir(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	expected := filepath.Join(dir, ".mnemosync", "staging")
	if got := cmd.StagingDir(); got != expected {
		t.Errorf("expected '%s', got '%s'", expected, got)
	}
}

func TestRepoPath(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	if got := cmd.RepoPath(); got != dir {
		t.Errorf("expected '%s', got '%s'", dir, got)
	}
}

func TestProcessRepoPath_Absolute(t *testing.T) {
	dir := t.TempDir()
	result, err := cmd.ProcessRepoPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != dir {
		t.Errorf("expected '%s', got '%s'", dir, result)
	}
}

func TestProcessRepoPath_Tilde(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	result, err := cmd.ProcessRepoPath("~")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != homeDir {
		t.Errorf("expected '%s', got '%s'", homeDir, result)
	}
}

func TestProcessRepoPath_TildePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(homeDir, "test-subdir-process")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(subdir) }()

	result, err := cmd.ProcessRepoPath("~/test-subdir-process")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != subdir {
		t.Errorf("expected '%s', got '%s'", subdir, result)
	}
}

func TestProcessRepoPath_NotExist(t *testing.T) {
	_, err := cmd.ProcessRepoPath("/tmp/nonexistent-path-xyz-12345")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestProcessRepoPath_NotDir(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := cmd.ProcessRepoPath(filePath)
	if err == nil {
		t.Error("expected error for file path")
	}
}

func TestProcessRepoPath_TildeHomeError(t *testing.T) {
	orig := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() { _ = os.Setenv("HOME", orig) }()
}

func TestCheckBinary_Found(t *testing.T) {
	result := cmd.CheckBinary("sh", false, false)
	if result != "" {
		t.Errorf("expected empty string for found binary, got: '%s'", result)
	}
}

func TestCheckBinary_NotFoundRequired(t *testing.T) {
	result := cmd.CheckBinary("nonexistent-binary-xyz", false, false)
	if !strings.Contains(result, "[FAIL]") {
		t.Errorf("expected [FAIL] for missing required binary, got: '%s'", result)
	}
}

func TestCheckBinary_NotFoundOptional(t *testing.T) {
	result := cmd.CheckBinary("nonexistent-binary-xyz", true, false)
	if !strings.Contains(result, "[WARNING]") {
		t.Errorf("expected [WARNING] for missing optional binary, got: '%s'", result)
	}
}

func TestCheckBinary_WithVersionError(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "failversion.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho \"error msg\"\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	origPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
	defer func() { _ = os.Setenv("PATH", origPath) }()

	result := cmd.CheckBinary("failversion.sh", false, false)
	if !strings.Contains(result, "[WARNING] Version check failed") {
		t.Errorf("expected version check warning, got: '%s'", result)
	}
}

func TestResolveAndValidatePath_Absolute(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "target")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.ConfigPath = filepath.Join(dir, "config", "config.yaml")
	defer resetGlobals()

	result, err := cmd.ResolveAndValidatePath(subdir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != subdir {
		t.Errorf("expected '%s', got '%s'", subdir, result)
	}
}

func TestResolveAndValidatePath_Tilde(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	result, err := cmd.ResolveAndValidatePath("~/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != homeDir {
		t.Errorf("expected '%s', got '%s'", homeDir, result)
	}
}

func TestResolveAndValidatePath_NotExist(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	_, err := cmd.ResolveAndValidatePath("/tmp/nonexistent-path-xyz-12345")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestResolveAndValidatePath_IsFile(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := cmd.ResolveAndValidatePath(filePath)
	if err == nil {
		t.Error("expected error for file path")
	}
}

func TestResolveAndValidatePath_CircularRepoRef(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	_, err := cmd.ResolveAndValidatePath(dir)
	if err == nil {
		t.Error("expected circular reference error")
	} else if !strings.Contains(err.Error(), "circular reference repo path") {
		t.Errorf("expected circular reference repo path error, got: %v", err)
	}
}

func TestResolveAndValidatePath_CircularConfigRef(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.ConfigPath = filepath.Join(configDir, "config.yaml")
	defer resetGlobals()

	_, err := cmd.ResolveAndValidatePath(configDir)
	if err != nil && strings.Contains(err.Error(), "cannot circular reference config path") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveAndValidatePath_MnemosyncDir(t *testing.T) {
	dir := t.TempDir()
	mnemosyncDir := filepath.Join(dir, "mnemosync")
	if err := os.Mkdir(mnemosyncDir, 0755); err != nil {
		t.Fatal(err)
	}
	setTestGlobals(dir)
	defer resetGlobals()

	_, err := cmd.ResolveAndValidatePath(mnemosyncDir)
	if err == nil {
		t.Error("expected error for 'mnemosync' dir")
	}
}

func TestResolveEntry_ByID(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "testalias"})
	defer resetGlobals()

	_, entry, found := cmd.ResolveEntry("1")
	if !found {
		t.Fatal("expected to find entry by ID")
	}
	if entry.Alias != "testalias" {
		t.Errorf("expected alias 'testalias', got '%s'", entry.Alias)
	}
}

func TestResolveEntry_ByAlias(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "myalias"})
	defer resetGlobals()

	_, entry, found := cmd.ResolveEntry("myalias")
	if !found {
		t.Fatal("expected to find entry by alias")
	}
	if entry.TargetPath != "/tmp/test" {
		t.Errorf("expected path '/tmp/test', got '%s'", entry.TargetPath)
	}
}

func TestResolveEntry_ByPath(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test", Alias: "testalias"})
	defer resetGlobals()

	_, entry, found := cmd.ResolveEntry("/tmp/test")
	if !found {
		t.Fatal("expected to find entry by path")
	}
	if entry.Alias != "testalias" {
		t.Errorf("expected alias 'testalias', got '%s'", entry.Alias)
	}
}

func TestResolveEntry_NotFound(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	_, _, found := cmd.ResolveEntry("nonexistent")
	if found {
		t.Error("expected not to find non-existent entry")
	}
}

func TestAddDirectoryEntry_Success(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	err := cmd.AddDirectoryEntry(dir, "newalias")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cmd.GetDataStore().TrackedDirs) != 1 {
		t.Errorf("expected 1 tracked dir, got %d", len(cmd.GetDataStore().TrackedDirs))
	}
}

func TestAddDirectoryEntry_DuplicatePath(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	_ = cmd.AddDirectoryEntry(dir, "alias1")
	err := cmd.AddDirectoryEntry(dir, "alias2")
	if err == nil {
		t.Error("expected error for duplicate path")
	}
}

func TestAddDirectoryEntry_DuplicateAlias(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	subdir1 := filepath.Join(dir, "dir1")
	subdir2 := filepath.Join(dir, "dir2")
	if err := os.Mkdir(subdir1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(subdir2, 0755); err != nil {
		t.Fatal(err)
	}

	_ = cmd.AddDirectoryEntry(subdir1, "samealias")
	err := cmd.AddDirectoryEntry(subdir2, "samealias")
	if err == nil {
		t.Error("expected error for duplicate alias")
	}
}

func TestEnsureInitialized_NotInitialized(t *testing.T) {
	resetGlobals()
	err := cmd.EnsureInitialized()
	if err == nil {
		t.Error("expected error when not initialized")
	}
}

func TestEnsureInitialized_NoRepoPath(t *testing.T) {
	resetGlobals()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			IsInit:   true,
			RepoPath: "",
		},
	})
	defer resetGlobals()

	err := cmd.EnsureInitialized()
	if err == nil {
		t.Error("expected error when repo path is empty")
	}
}

func TestEnsureInitialized_RepoNotExist(t *testing.T) {
	dir := t.TempDir()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			IsInit:   true,
			RepoPath: filepath.Join(dir, "nonexistent"),
		},
	})
	defer resetGlobals()

	err := cmd.EnsureInitialized()
	if err == nil {
		t.Error("expected error when repo path doesn't exist")
	}
}

func TestEnsureInitialized_Success(t *testing.T) {
	dir := t.TempDir()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			IsInit:   true,
			RepoPath: dir,
		},
	})
	defer resetGlobals()

	err := cmd.EnsureInitialized()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSelectDirs_All(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test1", Alias: "alias1"})
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test2", Alias: "alias2"})
	defer resetGlobals()

	dirs := cmd.SelectDirs([]string{})
	if len(dirs) != 2 {
		t.Errorf("expected 2 dirs, got %d", len(dirs))
	}
}

func TestSelectDirs_Specific(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test1", Alias: "alias1"})
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test2", Alias: "alias2"})
	defer resetGlobals()

	dirs := cmd.SelectDirs([]string{"1"})
	if len(dirs) != 1 {
		t.Errorf("expected 1 dir, got %d", len(dirs))
	}
}

func TestSelectDirs_Unknown(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test1", Alias: "alias1"})
	defer resetGlobals()

	dirs := cmd.SelectDirs([]string{"nonexistent"})
	if len(dirs) != 0 {
		t.Errorf("expected 0 dirs, got %d", len(dirs))
	}
}

func TestSelectDirs_Deduplicate(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/test1", Alias: "alias1"})
	defer resetGlobals()

	dirs := cmd.SelectDirs([]string{"1", "1"})
	if len(dirs) != 1 {
		t.Errorf("expected 1 dir (deduplicated), got %d", len(dirs))
	}
}

func TestPathCompleter_Dir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	suggestions := cmd.PathCompleter(dir + "/sub")
	if len(suggestions) == 0 {
		t.Error("expected suggestions for partial dir path")
	}
}

func TestPathCompleter_NoMatch(t *testing.T) {
	dir := t.TempDir()
	suggestions := cmd.PathCompleter(dir + "/nonexistent")
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(suggestions))
	}
}

func TestPathCompleter_Tilde(t *testing.T) {
	_, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	suggestions := cmd.PathCompleter("~")
	if len(suggestions) == 0 {
		t.Error("expected suggestions for tilde expansion")
		return
	}
	if !strings.HasPrefix(suggestions[0], "~") {
		t.Errorf("expected suggestion to start with '~', got '%s'", suggestions[0])
	}
}

func TestPathCompleter_TildePath(t *testing.T) {
	suggestions := cmd.PathCompleter("~/")
	if len(suggestions) == 0 {
		t.Log("no suggestions for home dir listing (may be empty)")
	}
}

func TestPathCompleter_HiddenFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".hidden"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "visible"), 0755); err != nil {
		t.Fatal(err)
	}

	suggestions := cmd.PathCompleter(dir + "/")
	hasVisible := false
	hasHidden := false
	for _, s := range suggestions {
		if strings.Contains(s, "visible") {
			hasVisible = true
		}
		if strings.Contains(s, ".hidden") {
			hasHidden = true
		}
	}
	if !hasVisible {
		t.Error("expected visible dir in suggestions")
	}
	if hasHidden {
		t.Error("expected hidden dir NOT in suggestions without leading dot")
	}

	suggestions = cmd.PathCompleter(dir + "/.")
	hasHidden = false
	for _, s := range suggestions {
		if strings.Contains(s, ".hidden") {
			hasHidden = true
		}
	}
	if !hasHidden {
		t.Error("expected hidden dir in suggestions when typing leading dot")
	}
}

func TestPathCompleter_Error(t *testing.T) {
	suggestions := cmd.PathCompleter("/nonexistent-path-xyz-12345/")
	if suggestions != nil {
		t.Errorf("expected nil for non-existent directory, got %v", suggestions)
	}
}

func TestPruneStaging_NoStagingDir(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	cmd.PruneStaging()
}

func TestPruneStaging_RemovesOrphans(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	staging := cmd.StagingDir()
	if err := os.MkdirAll(staging, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(staging, "orphan"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(staging, "tracked"), 0755); err != nil {
		t.Fatal(err)
	}

	cmd.GetDataStore().AddDir(config.DirData{TargetPath: "/tmp/tracked", Alias: "tracked"})

	cmd.PruneStaging()

	if _, err := os.Stat(filepath.Join(staging, "orphan")); !os.IsNotExist(err) {
		t.Error("expected orphan dir to be removed")
	}
	if _, err := os.Stat(filepath.Join(staging, "tracked")); os.IsNotExist(err) {
		t.Error("expected tracked dir to remain")
	}
}

func TestPruneOldArchives_RemoveOld(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.KeepArchives = 2
	defer resetGlobals()

	for i := 0; i < 4; i++ {
		path := filepath.Join(dir, "mnemosync-backup-20060102-15040"+string(rune('0'+i))+".tar.gz")
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cmd.PruneOldArchives("tar")

	matches, _ := filepath.Glob(filepath.Join(dir, "mnemosync-backup-*.tar.gz"))
	if len(matches) != 2 {
		t.Errorf("expected 2 archives remaining, got %d", len(matches))
	}
}

func TestPruneOldArchives_KeepAll(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.KeepArchives = 10
	defer resetGlobals()

	for i := 0; i < 3; i++ {
		path := filepath.Join(dir, "mnemosync-backup-20060102-15040"+string(rune('0'+i))+".tar.gz")
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cmd.PruneOldArchives("tar")

	matches, _ := filepath.Glob(filepath.Join(dir, "mnemosync-backup-*.tar.gz"))
	if len(matches) != 3 {
		t.Errorf("expected 3 archives remaining, got %d", len(matches))
	}
}

func TestPruneOldArchives_Zip(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.KeepArchives = 1
	defer resetGlobals()

	for i := 0; i < 3; i++ {
		path := filepath.Join(dir, "mnemosync-backup-20060102-15040"+string(rune('0'+i))+".zip")
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cmd.PruneOldArchives("zip")

	matches, _ := filepath.Glob(filepath.Join(dir, "mnemosync-backup-*.zip"))
	if len(matches) != 1 {
		t.Errorf("expected 1 archive remaining, got %d", len(matches))
	}
}

func TestPruneOldArchives_KeepZero(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.KeepArchives = 0
	defer resetGlobals()

	cmd.PruneOldArchives("tar")
}

func TestCleanupStaging(t *testing.T) {
	dir := t.TempDir()
	staging := filepath.Join(dir, ".mnemosync", "staging")
	if err := os.MkdirAll(staging, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(staging, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	cmd.CleanupStaging(staging)

	entries, _ := os.ReadDir(staging)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after cleanup, got %d", len(entries))
	}
}

func TestCleanupStaging_NonExistent(t *testing.T) {
	cmd.CleanupStaging("/tmp/nonexistent-staging-dir-xyz-12345")
}

func TestCreateTarArchive(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	dstPath := filepath.Join(dir, "archive.tar.gz")

	err := cmd.CreateTarArchive(srcDir, dstPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("expected tar archive to be created")
	}
}

func TestCreateZipArchive(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	dstPath := filepath.Join(dir, "archive.zip")

	err := cmd.CreateZipArchive(srcDir, dstPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("expected zip archive to be created")
	}
}

func TestEnsureGitignore(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	err := cmd.EnsureGitignore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error("expected .gitignore to be created")
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()

	prevMMSync := os.Getenv("MMSYNC_CONF")
	_ = os.Setenv("MMSYNC_CONF", dir)

	homeDir, _ := os.UserHomeDir()
	backupDir := filepath.Join(dir, ".backup-mmsync")
	oldConfigDir := filepath.Join(homeDir, ".config/mmsync")
	hadOldConfig := false
	if _, err := os.Stat(oldConfigDir); err == nil {
		hadOldConfig = true
		_ = os.Rename(oldConfigDir, backupDir)
	}

	setTestGlobals(dir)
	defer func() {
		resetGlobals()
		_ = os.Setenv("MMSYNC_CONF", prevMMSync)
		if hadOldConfig {
			_ = os.RemoveAll(oldConfigDir)
			_ = os.Rename(backupDir, oldConfigDir)
		}
	}()

	cmd.GetAppConf().ConfigSchema.ConfigPath = filepath.Join(dir, "config.yaml")

	cmd.SaveConfig()

	configPath := filepath.Join(dir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to be saved")
	}
}

func TestRunGit(t *testing.T) {
	dir := t.TempDir()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RepoPath: dir,
			IsInit:   true,
		},
	})
	defer resetGlobals()

	err := cmd.RunGit("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureLfsTracking_SmallFile(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.LfsThresholdMb = 100
	defer resetGlobals()

	archivePath := filepath.Join(dir, "archive.tar.gz")
	if err := os.WriteFile(archivePath, []byte("small"), 0644); err != nil {
		t.Fatal(err)
	}

	err := cmd.EnsureLfsTracking(archivePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureLfsTracking_NoLFS(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.LfsThresholdMb = 0
	defer resetGlobals()

	archivePath := filepath.Join(dir, "archive.tar.gz")
	if err := os.WriteFile(archivePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	err := cmd.EnsureLfsTracking(archivePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureLfsTracking_NoGitLFSBinary(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.LfsThresholdMb = 1
	defer resetGlobals()

	archivePath := filepath.Join(dir, "archive.tar.gz")
	largeData := make([]byte, 2*1024*1024)
	if err := os.WriteFile(archivePath, largeData, 0644); err != nil {
		t.Fatal(err)
	}

	err := cmd.EnsureLfsTracking(archivePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureLfsTracking_AlreadyTracked(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.LfsThresholdMb = 1
	defer resetGlobals()

	archivePath := filepath.Join(dir, "mnemosync-backup-20060102-150405.tar.gz")
	largeData := make([]byte, 2*1024*1024)
	if err := os.WriteFile(archivePath, largeData, 0644); err != nil {
		t.Fatal(err)
	}

	attrPath := filepath.Join(dir, ".gitattributes")
	if err := os.WriteFile(attrPath, []byte("mnemosync-backup-*.tar.gz filter=lfs diff=lfs merge=lfs -text\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := cmd.EnsureLfsTracking(archivePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureLfsTracking_Zip(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.LfsThresholdMb = 1
	defer resetGlobals()

	archivePath := filepath.Join(dir, "mnemosync-backup-20060102-150405.zip")
	largeData := make([]byte, 2*1024*1024)
	if err := os.WriteFile(archivePath, largeData, 0644); err != nil {
		t.Fatal(err)
	}

	attrPath := filepath.Join(dir, ".gitattributes")
	if err := os.WriteFile(attrPath, []byte("mnemosync-backup-*.zip filter=lfs diff=lfs merge=lfs -text\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := cmd.EnsureLfsTracking(archivePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecute(t *testing.T) {
	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			AppVersion: "0.1.0",
		},
	}
	ds := config.GetDataStore()

	stdoutBak := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	origArgs := os.Args
	os.Args = []string{"mns", "version"}

	done := make(chan bool)
	go func() {
		defer func() { _ = recover() }()
		cmd.Execute(cfg, ds)
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	_ = w.Close()
	os.Stdout = stdoutBak
	os.Args = origArgs

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "mnemosync") {
		t.Errorf("expected version output, got: '%s'", output)
	}
}

func TestHealthCmd(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("config: test"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			ConfigPath: configPath,
			RepoPath:   dir,
			DbPath:     filepath.Join(dir, "state.json"),
			IsInit:     true,
		},
	})
	defer resetGlobals()

	stdoutBak := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	origArgs := os.Args
	os.Args = []string{"mns", "health"}

	done := make(chan bool)
	go func() {
		defer func() { _ = recover() }()
		_ = cmd.RootCmd.Execute()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	_ = w.Close()
	os.Stdout = stdoutBak
	os.Args = origArgs

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Health Check") {
		t.Errorf("expected health output, got: '%s'", output)
	}
}

func TestPruneStaging_NoStagingDirExists(t *testing.T) {
	dir := t.TempDir()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RepoPath: dir,
		},
	})
	defer resetGlobals()

	cmd.PruneStaging()
}

func TestPruneOldArchives_GlobError(t *testing.T) {
	dir := t.TempDir()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RepoPath:     dir,
			KeepArchives: 3,
		},
	})
	defer resetGlobals()

	cmd.PruneOldArchives("tar")
}

func TestEnsureGitignoreInDir_SymlinkDir(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("existing content\n"), 0000); err != nil {
		t.Fatal(err)
	}
	_ = os.Chmod(gitignorePath, 0000)
	defer func() { _ = os.Chmod(gitignorePath, 0644) }()

	err := cmd.EnsureGitignoreInDir(dir)
	if err == nil {
		t.Error("expected error when .gitignore is not writable")
	}
}

func TestEnsureLfsTracking_NonExistentArchive(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.LfsThresholdMb = 1
	defer resetGlobals()

	archivePath := filepath.Join(dir, "nonexistent.tar.gz")

	err := cmd.EnsureLfsTracking(archivePath)
	if err == nil {
		t.Error("expected error for non-existent archive")
	}
}

func TestDisplayManPage_PagerFallback(t *testing.T) {
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			AppVersion: "0.1.0",
		},
	})
	defer resetGlobals()

	origPager := os.Getenv("PAGER")
	_ = os.Setenv("PAGER", "cat")
	defer func() { _ = os.Setenv("PAGER", origPager) }()

	stdoutBak := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.DisplayManPage()

	_ = w.Close()
	os.Stdout = stdoutBak

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "mns") {
		t.Errorf("expected man page output to contain 'mns', got: '%s'", output)
	}
}

func TestSaveConfig_WithHOME(t *testing.T) {
	dir := t.TempDir()
	realHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", realHome) }()

	prevMMSync := os.Getenv("MMSYNC_CONF")
	_ = os.Unsetenv("MMSYNC_CONF")
	defer func() { _ = os.Setenv("MMSYNC_CONF", prevMMSync) }()

	setTestGlobals(dir)
	cmd.GetAppConf().ConfigSchema.ConfigPath = filepath.Join(dir, ".config/mmsync/config.yaml")

	_ = os.MkdirAll(filepath.Join(dir, ".config/mmsync"), 0755)
	if err := os.WriteFile(cmd.GetAppConf().ConfigSchema.ConfigPath, []byte("old: data"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd.SaveConfig()

	data, _ := os.ReadFile(cmd.GetAppConf().ConfigSchema.ConfigPath)
	if !strings.Contains(string(data), "config_schema") {
		t.Error("expected saved config to contain config_schema")
	}

	resetGlobals()
}

func TestProcessRepoPath_TildeHomeDirErr(t *testing.T) {
	realHome := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() { _ = os.Setenv("HOME", realHome) }()

	_, err := cmd.ProcessRepoPath("~/test")
	if err == nil {
		t.Error("expected error when HOME is not set")
	}
}

func TestResolveAndValidatePath_TildeHomeErr(t *testing.T) {
	dir := t.TempDir()
	setTestGlobals(dir)
	defer resetGlobals()

	realHome := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() { _ = os.Setenv("HOME", realHome) }()

	_, err := cmd.ResolveAndValidatePath("~/test")
	if err == nil {
		t.Error("expected error when HOME is not set")
	}
}

func TestPathCompleter_TildeHomeErr(t *testing.T) {
	realHome := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() { _ = os.Setenv("HOME", realHome) }()

	suggestions := cmd.PathCompleter("~")
	if suggestions != nil {
		t.Error("expected nil when HOME is not set")
	}
}

func TestPathCompleter_NonExistentDir(t *testing.T) {
	suggestions := cmd.PathCompleter("/nonexistent-path-xyz/")
	if suggestions != nil {
		t.Errorf("expected nil, got %v", suggestions)
	}
}

func TestPersistManPage_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Unsetenv("HOME") }()

	err := cmd.PersistManPage(".TH MNS 1 \"test\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manPath := filepath.Join(dir, ".local", "share", "man", "man1", "mns.1")
	if _, err := os.Stat(manPath); os.IsNotExist(err) {
		t.Error("expected man page file to be created")
	}
}

func TestPersistManPage_SkipRewrite(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Unsetenv("HOME") }()

	manPath := filepath.Join(dir, ".local", "share", "man", "man1", "mns.1")
	_ = os.MkdirAll(filepath.Dir(manPath), 0755)
	_ = os.WriteFile(manPath, []byte(".TH MNS 1 \"test\""), 0644)
	origModTime, _ := os.Stat(manPath)

	err := cmd.PersistManPage(".TH MNS 1 \"test\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	newModTime, _ := os.Stat(manPath)
	if newModTime.ModTime() != origModTime.ModTime() {
		t.Error("expected file not to be rewritten when content is the same")
	}
}

func TestPersistManPage_NoHome(t *testing.T) {
	realHome := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() { _ = os.Setenv("HOME", realHome) }()

	err := cmd.PersistManPage("test")
	if err == nil {
		t.Error("expected error when HOME is not set")
	}
}

func TestDisplayManPage_NoPanic(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Unsetenv("HOME") }()

	origPager := os.Getenv("PAGER")
	_ = os.Setenv("PAGER", "cat")
	defer func() { _ = os.Setenv("PAGER", origPager) }()

	stdoutBak := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan bool, 1)
	go func() {
		defer func() { _ = recover(); done <- true }()
		cmd.DisplayManPage()
		done <- true
	}()

	<-done
	_ = w.Close()
	os.Stdout = stdoutBak
	_ = r.Close()
}
