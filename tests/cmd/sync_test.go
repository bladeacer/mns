package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladeacer/mns/cmd"
	"github.com/bladeacer/mns/config"
)

func TestBuildRsyncArgs_DefaultFlags(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RespectGitignore: true,
		},
	}
	cmd.SetAppConf(cfg)

	entry := config.DirData{
		TargetPath: "/some/path",
		Alias:      "myalias",
	}

	args := cmd.BuildRsyncArgs(entry, dir, "-a")
	expected := []string{"-a", "--delete", "/some/path/", filepath.Join(dir, "myalias")}

	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
	for i := range expected {
		if args[i] != expected[i] {
			t.Errorf("arg[%d]: expected '%s', got '%s'", i, expected[i], args[i])
		}
	}
}

func TestBuildRsyncArgs_RespectGitignoreDisabled(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RespectGitignore: false,
		},
	}
	cmd.SetAppConf(cfg)

	entry := config.DirData{
		TargetPath: "/some/path",
		Alias:      "myalias",
	}

	args := cmd.BuildRsyncArgs(entry, dir, "-av")
	hasExclude := false
	for _, a := range args {
		if a == "--exclude=.gitignore" {
			hasExclude = true
			break
		}
	}
	if !hasExclude {
		t.Errorf("expected --exclude=.gitignore in args when RespectGitignore=false, got: %v", args)
	}
}

func TestBuildRsyncArgs_CustomFlags(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RespectGitignore: true,
		},
	}
	cmd.SetAppConf(cfg)

	entry := config.DirData{
		TargetPath: "/other/path",
		Alias:      "otheralias",
	}

	args := cmd.BuildRsyncArgs(entry, dir, "-a", "--verbose")
	containsA := false
	containsVerbose := false
	for _, a := range args {
		if a == "-a" {
			containsA = true
		}
		if a == "--verbose" {
			containsVerbose = true
		}
	}
	if !containsA {
		t.Error("expected -a flag in args")
	}
	if !containsVerbose {
		t.Error("expected --verbose flag in args")
	}
}

func TestSelectDirs_AllWhenNoArgs(t *testing.T) {
	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/a", Alias: "alias-a"})
	ds.AddDir(config.DirData{TargetPath: "/b", Alias: "alias-b"})
	cmd.SetDataStore(ds)

	dirs := cmd.SelectDirs(nil)
	if len(dirs) != 2 {
		t.Fatalf("expected 2 dirs, got %d", len(dirs))
	}
}

func TestSelectDirs_ByAlias(t *testing.T) {
	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/a", Alias: "alias-a"})
	ds.AddDir(config.DirData{TargetPath: "/b", Alias: "alias-b"})
	cmd.SetDataStore(ds)

	dirs := cmd.SelectDirs([]string{"alias-a"})
	if len(dirs) != 1 {
		t.Fatalf("expected 1 dir, got %d", len(dirs))
	}
	if dirs[0].Alias != "alias-a" {
		t.Errorf("expected alias-a, got %s", dirs[0].Alias)
	}
}

func TestSelectDirs_SkipsUnknown(t *testing.T) {
	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/a", Alias: "alias-a"})
	cmd.SetDataStore(ds)

	dirs := cmd.SelectDirs([]string{"nonexistent"})
	if len(dirs) != 0 {
		t.Fatalf("expected 0 dirs for unknown alias, got %d", len(dirs))
	}
}

func TestEnsureInitialized_NilConfig(t *testing.T) {
	cmd.SetAppConf(nil)
	err := cmd.EnsureInitialized()
	if err == nil {
		t.Fatal("expected error when AppConf is nil")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("expected 'not initialized' error, got: %v", err)
	}
}

func TestEnsureInitialized_NotInit(t *testing.T) {
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			IsInit:   false,
			RepoPath: "/some/path",
		},
	})
	err := cmd.EnsureInitialized()
	if err == nil {
		t.Fatal("expected error when IsInit is false")
	}
}

func TestEnsureInitialized_EmptyRepoPath(t *testing.T) {
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			IsInit:   true,
			RepoPath: "",
		},
	})
	err := cmd.EnsureInitialized()
	if err == nil {
		t.Fatal("expected error when RepoPath is empty")
	}
}

func TestEnsureInitialized_RepoPathNotExist(t *testing.T) {
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			IsInit:   true,
			RepoPath: "/nonexistent-path-xyz-98765",
		},
	})
	err := cmd.EnsureInitialized()
	if err == nil {
		t.Fatal("expected error when RepoPath does not exist")
	}
}

func TestEnsureInitialized_Valid(t *testing.T) {
	dir := t.TempDir()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			IsInit:   true,
			RepoPath: dir,
		},
	})
	err := cmd.EnsureInitialized()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPruneStagingSync_RemovesOrphans(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "repo")
	stagingDir := filepath.Join(repoPath, ".mnemosync", "staging")

	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(stagingDir, "tracked-dir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(stagingDir, "orphan-dir"), 0755); err != nil {
		t.Fatal(err)
	}

	ds := config.GetDataStore()
	ds.AddDir(config.DirData{TargetPath: "/a", Alias: "tracked-dir"})
	cmd.SetDataStore(ds)

	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RepoPath: repoPath,
		},
	})

	cmd.PruneStaging()

	if _, err := os.Stat(filepath.Join(stagingDir, "tracked-dir")); os.IsNotExist(err) {
		t.Error("expected tracked-dir to remain after prune")
	}
	if _, err := os.Stat(filepath.Join(stagingDir, "orphan-dir")); !os.IsNotExist(err) {
		t.Error("expected orphan-dir to be removed after prune")
	}
}

func TestPruneStagingSync_NonExistentStaging(t *testing.T) {
	dir := t.TempDir()
	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RepoPath: filepath.Join(dir, "nonexistent"),
		},
	})

	cmd.PruneStaging()
}

func TestCleanupStagingSync(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd.CleanupStaging(dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty staging dir after cleanup, got %d entries", len(entries))
	}
}

func TestRecordPushHistory(t *testing.T) {
	dir := t.TempDir()

	savedHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	_ = os.Setenv("MMSYNC_CONF", dir)
	defer func() {
		_ = os.Setenv("HOME", savedHome)
		_ = os.Unsetenv("MMSYNC_CONF")
	}()

	ds := config.GetDataStore()
	cmd.SetDataStore(ds)

	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{},
	})

	staging := filepath.Join(dir, "staging")
	if err := os.MkdirAll(staging, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(staging, "dir-a"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(staging)
	if err != nil {
		t.Fatal(err)
	}

	cmd.RecordPushHistory(entries, "archive.tar.gz", 1024)

	if len(ds.StagingHistory) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(ds.StagingHistory))
	}
	if ds.StagingHistory[0].Archive != "archive.tar.gz" {
		t.Errorf("expected archive.tar.gz, got %s", ds.StagingHistory[0].Archive)
	}
	if ds.StagingHistory[0].SizeBytes != 1024 {
		t.Errorf("expected 1024 bytes, got %d", ds.StagingHistory[0].SizeBytes)
	}
	if len(ds.StagingHistory[0].Dirs) != 1 || ds.StagingHistory[0].Dirs[0] != "dir-a" {
		t.Errorf("expected [dir-a] dirs, got %v", ds.StagingHistory[0].Dirs)
	}

	expectedDbPath := filepath.Join(dir, "mmsync-state.json")
	if _, err := os.Stat(expectedDbPath); os.IsNotExist(err) {
		t.Errorf("expected db file to be created at %s", expectedDbPath)
	}
}

func TestPruneOldArchives_KeepsRecent(t *testing.T) {
	dir := t.TempDir()

	cmd.SetAppConf(&config.MnemoConf{
		ConfigSchema: config.ConfigSchema{
			RepoPath:     dir,
			KeepArchives: 2,
		},
	})

	for i := 0; i < 3; i++ {
		fname := filepath.Join(dir, "mnemosync-backup-20060102-15040"+string(rune('0'+i))+".tar.gz")
		if err := os.WriteFile(fname, []byte("archive"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cmd.PruneOldArchives("tar")

	matches, _ := filepath.Glob(filepath.Join(dir, "mnemosync-backup-*.tar.gz"))
	if len(matches) != 2 {
		t.Errorf("expected 2 archives remaining, got %d: %v", len(matches), matches)
	}
}
