package cmd

import "github.com/bladeacer/mmsync/config"

var (
	ProcessRepoPath        = processRepoPath
	CheckBinary            = checkBinary
	ResolveAndValidatePath = resolveAndValidatePath
	ResolveEntry           = resolveEntry
	AddDirectoryEntry      = addDirectoryEntry
	EnsureInitialized      = ensureInitialized
	SelectDirs             = selectDirs
	PathCompleter          = pathCompleter
	PruneStaging           = pruneStaging
	PruneOldArchives       = pruneOldArchives
	CleanupStaging         = cleanupStaging
	CreateTarArchive       = createTarArchive
	CreateZipArchive       = createZipArchive
	EnsureGitignore        = ensureGitignore
	EnsureGitignoreInDir   = ensureGitignoreInDir
	SaveConfig             = saveConfig
	RunGit                 = runGit
	EnsureLfsTracking      = ensureLfsTracking
	DisplayManPage         = displayManPage
	StagingDir             = stagingDir
	RepoPathFn             = repoPath
	RootCmd                = rootCmd
)

func SetAppConf(cfg *config.MnemoConf)  { appConf = cfg }
func GetAppConf() *config.MnemoConf     { return appConf }
func SetDataStore(ds *config.DataStore) { dataStore = ds }
func GetDataStore() *config.DataStore   { return dataStore }
