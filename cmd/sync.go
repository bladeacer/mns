package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/fileio"
	"github.com/spf13/cobra"
)

func StagingDir() string {
	return filepath.Join(AppConf.ConfigSchema.RepoPath, ".mnemosync", "staging")
}

func RepoPath() string {
	return AppConf.ConfigSchema.RepoPath
}

func RunGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = RepoPath()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var stageCmd = &cobra.Command{
	Use:   "stage [id-or-alias...]",
	Short: "Rsync tracked directories to the repo staging area",
	Long: `Rsync all (or specified) tracked directories to the target repository's staging area.
Files are mirrored under <repo>/.mnemosync/staging/ which is gitignored.
They will be archived and committed on 'mns push'.

Examples:
  mns stage
  mns stage 1 myalias`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		runStage(args, "-a")
	},
}

var unstageCmd = &cobra.Command{
	Use:   "unstage [id-or-alias...]",
	Short: "Remove tracked directories from the staging area",
	Long: `Remove all (or specified) tracked directories from the staging area.
Since staging files are gitignored, this simply deletes the mirrored files.

Examples:
  mns unstage
  mns unstage 1 myalias`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := EnsureGitignore(); err != nil {
			fmt.Fprintf(os.Stderr, "Error ensuring .gitignore: %v\n", err)
			os.Exit(1)
		}

		PruneStaging()
		runStage(args, "-av")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show git status of the target repository",
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := RunGit("status"); err != nil {
			os.Exit(1)
		}
	},
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show git log of the target repository",
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		logLimit, _ := cmd.Flags().GetInt("limit")
		gitArgs := []string{"log", "--oneline"}
		if logLimit > 0 {
			gitArgs = append(gitArgs, fmt.Sprintf("-%d", logLimit))
		}
		gitArgs = append(gitArgs, args...)

		if err := RunGit(gitArgs...); err != nil {
			os.Exit(1)
		}
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Archive staged files, commit, and push to remote",
	Long: `Archives the staged files in the repository using the configured archiver (tar or zip),
commits with the configured message format, and pushes to the remote.

Examples:
  mns push
  mns push --no-push`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsureInitialized(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		noPush, _ := cmd.Flags().GetBool("no-push")
		runPush(noPush)
	},
}

func runStage(args []string, rsyncFlags ...string) {
	dirs := SelectDirs(args)

	if len(dirs) == 0 {
		fmt.Println("No directories to stage.")
		return
	}

	staging := StagingDir()
	if err := os.MkdirAll(staging, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating staging directory %s: %v\n", staging, err)
		os.Exit(1)
	}

	for _, entry := range dirs {
		rsyncEntry(entry, staging, rsyncFlags...)
	}

	fmt.Println("Staging complete.")
}

func rsyncEntry(entry config.DirData, staging string, rsyncFlags ...string) {
	dest := filepath.Join(staging, entry.Alias)
	fmt.Printf("Staging '%s' (%s) -> %s\n", entry.Alias, entry.TargetPath, dest)

	args := BuildRsyncArgs(entry, staging, rsyncFlags...)
	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error rsyncing '%s': %v\n", entry.Alias, err)
		os.Exit(1)
	}
}

func BuildRsyncArgs(entry config.DirData, staging string, rsyncFlags ...string) []string {
	dest := filepath.Join(staging, entry.Alias)
	args := append([]string{}, rsyncFlags...)
	args = append(args, "--delete")
	if !AppConf.ConfigSchema.RespectGitignore {
		args = append(args, "--exclude=.gitignore")
	}
	args = append(args, entry.TargetPath+"/", dest)
	return args
}

func runPush(noPush bool) {
	if err := EnsureGitignore(); err != nil {
		fmt.Fprintf(os.Stderr, "Error ensuring .gitignore: %v\n", err)
		os.Exit(1)
	}

	PruneStaging()

	staging := StagingDir()
	stagingInfo, err := os.Stat(staging)
	if err != nil || !stagingInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: staging directory not found at %s.\nRun 'mns stage' first.\n", staging)
		os.Exit(1)
	}

	entries, err := os.ReadDir(staging)
	if err != nil || len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "Error: staging directory is empty.\nRun 'mns stage' first.\n")
		os.Exit(1)
	}

	archivePath, archiveName, archiveSize := CreateArchive(staging, AppConf.ConfigSchema.Archiver)
	fmt.Printf("Created archive: %s (%d bytes)\n", archiveName, archiveSize)

	PruneOldArchives(AppConf.ConfigSchema.Archiver)

	if err := EnsureLfsTracking(archivePath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not configure Git LFS: %v\n", err)
	}

	fmt.Println("Running git add -A...")
	if err := RunGit("add", "-A"); err != nil {
		os.Exit(1)
	}

	commitMsg := time.Now().Format(AppConf.ConfigSchema.CommitFmt)
	fmt.Printf("Committing: %s\n", commitMsg)
	if err := RunGit("commit", "-m", commitMsg); err != nil {
		os.Exit(1)
	}

	if !noPush {
		fmt.Println("Pushing to remote...")
		if err := RunGit("push"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: git push failed\n")
		} else {
			fmt.Println("Push complete.")
		}
	} else {
		fmt.Println("Skipping push (--no-push flag set).")
	}

	RecordPushHistory(entries, archiveName, archiveSize)
	CleanupStaging(staging)
	fmt.Println("Push complete. Staging directory cleaned.")
}

func CreateArchive(staging, archiver string) (archivePath, archiveName string, size int64) {
	timestamp := time.Now().Format("20060102-150405")

	switch archiver {
	case "zip":
		archiveName = fmt.Sprintf("mnemosync-backup-%s.zip", timestamp)
		archivePath = filepath.Join(RepoPath(), archiveName)
		if err := CreateZipArchive(staging, archivePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating zip archive: %v\n", err)
			os.Exit(1)
		}
	default:
		archiveName = fmt.Sprintf("mnemosync-backup-%s.tar.gz", timestamp)
		archivePath = filepath.Join(RepoPath(), archiveName)
		if err := CreateTarArchive(staging, archivePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tar archive: %v\n", err)
			os.Exit(1)
		}
	}

	info, _ := os.Stat(archivePath)
	return archivePath, archiveName, info.Size()
}

func RecordPushHistory(entries []os.DirEntry, archiveName string, archiveSize int64) {
	aliases := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			aliases = append(aliases, e.Name())
		}
	}

	dbPath := fileio.ResolveDbPath()
	DataStore.AddHistory(config.StagingHistoryEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Archive:   archiveName,
		SizeBytes: archiveSize,
		Dirs:      aliases,
	})
	if err := DataStore.SaveData(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save staging history: %v\n", err)
	}
}

func CreateTarArchive(srcDir, dstPath string) error {
	parent := filepath.Dir(srcDir)
	base := filepath.Base(srcDir)
	args := []string{"-czf", dstPath, "-C", parent, base}
	cmd := exec.Command("tar", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CreateZipArchive(srcDir, dstPath string) error {
	cmd := exec.Command("zip", "-r", dstPath, ".")
	cmd.Dir = srcDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func EnsureGitignore() error {
	return EnsureGitignoreInDir(RepoPath())
}

func EnsureGitignoreInDir(dir string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading .gitignore: %w", err)
		}
		content = []byte{}
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "/.mnemosync/" {
			return nil
		}
	}

	hasContent := len(content) > 0 && (len(lines) != 1 || lines[0] != "")
	entry := "/.mnemosync/\n"
	if hasContent {
		entry = "\n" + entry
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening .gitignore: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("writing to .gitignore: %w", err)
	}

	fmt.Println("Added '/.mnemosync/' to repo .gitignore")
	return nil
}

func PruneStaging() {
	staging := StagingDir()
	entries, err := os.ReadDir(staging)
	if err != nil {
		return
	}

	aliases := make(map[string]bool)
	for _, entry := range DataStore.TrackedDirs {
		aliases[entry.Alias] = true
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if !aliases[e.Name()] {
			path := filepath.Join(staging, e.Name())
			fmt.Printf("Removing orphan staging dir '%s' (no longer tracked)\n", e.Name())
			if err := os.RemoveAll(path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not remove orphan '%s': %v\n", e.Name(), err)
			}
		}
	}
}

func PruneOldArchives(archiver string) {
	keep := AppConf.ConfigSchema.KeepArchives
	if keep <= 0 {
		return
	}

	var pattern string
	switch archiver {
	case "zip":
		pattern = "mnemosync-backup-*.zip"
	default:
		pattern = "mnemosync-backup-*.tar.gz"
	}

	matches, err := filepath.Glob(filepath.Join(RepoPath(), pattern))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not list archives: %v\n", err)
		return
	}

	sort.Strings(matches)

	if len(matches) <= keep {
		return
	}

	toRemove := matches[:len(matches)-keep]
	for _, path := range toRemove {
		fmt.Printf("Removing old archive: %s\n", filepath.Base(path))
		if err := os.Remove(path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove old archive '%s': %v\n", path, err)
		}
	}
}

func lfsArchivePattern(ext string) string {
	if ext == ".zip" {
		return "mnemosync-backup-*.zip"
	}
	return "mnemosync-backup-*.tar.gz"
}

func gitAttributesHasPattern(repoPath, pattern string) (bool, error) {
	attrPath := filepath.Join(repoPath, ".gitattributes")
	content, err := os.ReadFile(attrPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("reading .gitattributes: %w", err)
		}
		return false, nil
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), pattern) {
			return true, nil
		}
	}
	return false, nil
}

func EnsureLfsTracking(archivePath string) error {
	threshold := AppConf.ConfigSchema.LfsThresholdMb
	if threshold <= 0 {
		return nil
	}

	info, err := os.Stat(archivePath)
	if err != nil {
		return err
	}

	if info.Size() < threshold*1024*1024 {
		return nil
	}

	lfsPath, err := exec.LookPath("git-lfs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: archive is %d bytes but git-lfs is not installed\n", info.Size())
		return nil
	}

	pattern := lfsArchivePattern(filepath.Ext(archivePath))
	fmt.Printf("Archive exceeds %d MB threshold, configuring Git LFS for '%s'...\n", threshold, pattern)

	exists, err := gitAttributesHasPattern(RepoPath(), pattern)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	cmd := exec.Command(lfsPath, "track", pattern)
	cmd.Dir = RepoPath()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git lfs track failed: %w", err)
	}

	fmt.Printf("Git LFS configured for '%s'\n", pattern)
	return nil
}

func CleanupStaging(StagingDir string) {
	entries, err := os.ReadDir(StagingDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		path := filepath.Join(StagingDir, e.Name())
		_ = os.RemoveAll(path)
	}
}

func EnsureInitialized() error {
	if AppConf == nil || !AppConf.ConfigSchema.IsInit {
		return fmt.Errorf("configuration not initialized. Run 'mns init' first")
	}
	if AppConf.ConfigSchema.RepoPath == "" {
		return fmt.Errorf("repository path not set in configuration")
	}
	if _, err := os.Stat(AppConf.ConfigSchema.RepoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository path '%s' does not exist", AppConf.ConfigSchema.RepoPath)
	}
	return nil
}

func SelectDirs(args []string) []config.DirData {
	if len(args) == 0 {
		result := make([]config.DirData, 0, len(DataStore.TrackedDirs))
		for _, entry := range DataStore.TrackedDirs {
			result = append(result, entry)
		}
		return result
	}

	seen := make(map[string]bool)
	var result []config.DirData
	for _, arg := range args {
		if seen[arg] {
			continue
		}
		seen[arg] = true

		_, entry, found := ResolveEntry(arg)
		if !found {
			fmt.Fprintf(os.Stderr, "Warning: no tracked directory matches '%s', skipping.\n", arg)
			continue
		}
		result = append(result, entry)
	}
	return result
}

func init() {
	RootCmd.AddCommand(stageCmd)
	RootCmd.AddCommand(unstageCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(logCmd)
	RootCmd.AddCommand(pushCmd)

	logCmd.Flags().IntP("limit", "n", 0, "Limit the number of log entries")
	pushCmd.Flags().Bool("no-push", false, "Create archive and commit without pushing")
}
