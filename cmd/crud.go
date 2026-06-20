package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bladeacer/mns/config"
	"github.com/bladeacer/mns/internal/fileio"
	"github.com/spf13/cobra"
)

// ListedEntry represents a standardized data structure for viewing or searching data
type ListedEntry struct {
	ID         string
	Alias      string
	TargetPath string
}

// GetSortedTrackedDirectories returns an ordered slice of tracked directories.
// It sorts primarily by length of ID, then alphanumerically.
func GetSortedTrackedDirectories() []ListedEntry {
	store := GetDataStore()
	ids := make([]string, 0, len(store.TrackedDirs))
	for id := range store.TrackedDirs {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		ai, bi := ids[i], ids[j]
		return len(ai) < len(bi) || (len(ai) == len(bi) && ai < bi)
	})

	entries := make([]ListedEntry, 0, len(ids))
	for _, id := range ids {
		entry := store.TrackedDirs[id]
		entries = append(entries, ListedEntry{
			ID:         id,
			Alias:      entry.Alias,
			TargetPath: entry.TargetPath,
		})
	}
	return entries
}

// RemoveTrackedDirectory deletes an entry from the datastore by its ID, path, or alias.
func RemoveTrackedDirectory(query string) error {
	id, _, found := ResolveEntry(query)
	if !found {
		return fmt.Errorf("no tracked directory matches '%s'", query)
	}

	store := GetDataStore()
	if !store.DeleteDir(id) {
		return fmt.Errorf("failed to remove entry '%s'", query)
	}

	dbPath := fileio.ResolveDbPath()
	if err := store.SaveData(dbPath); err != nil {
		return fmt.Errorf("error saving database: %w", err)
	}
	return nil
}

// ClearAllTrackedDirectories completely flushes the store and commits the empty database.
func ClearAllTrackedDirectories() error {
	store := GetDataStore()
	store.TrackedDirs = make(map[string]config.DirData)

	dbPath := fileio.ResolveDbPath()
	if err := store.SaveData(dbPath); err != nil {
		return fmt.Errorf("error saving database: %w", err)
	}
	return nil
}

// SearchTrackedDirectories evaluates a case-insensitive query match on ID, Alias, or TargetPath.
func SearchTrackedDirectories(query string) []ListedEntry {
	normalizedQuery := strings.ToLower(query)
	allEntries := GetSortedTrackedDirectories()

	var matches []ListedEntry
	for _, entry := range allEntries {
		if strings.Contains(strings.ToLower(entry.Alias), normalizedQuery) ||
			strings.Contains(strings.ToLower(entry.TargetPath), normalizedQuery) ||
			strings.Contains(strings.ToLower(entry.ID), normalizedQuery) {
			matches = append(matches, entry)
		}
	}
	return matches
}

// ChangeTrackedDirectory updates the path or alias of a tracked directory.
func ChangeTrackedDirectory(query, newPath, newAlias string) error {
	if newPath == "" && newAlias == "" {
		return fmt.Errorf("at least one of --path or --alias must be provided")
	}

	id, entry, found := ResolveEntry(query)
	if !found {
		return fmt.Errorf("no tracked directory matches '%s'", query)
	}

	store := GetDataStore()

	if newPath != "" {
		resolved, err := ResolveAndValidatePath(newPath)
		if err != nil {
			return err
		}
		entry.TargetPath = resolved
	}

	if newAlias != "" {
		if _, _, exists := store.FindDirByAlias(newAlias); exists {
			return fmt.Errorf("alias '%s' is already in use", newAlias)
		}
		entry.Alias = newAlias
	}

	store.TrackedDirs[id] = entry
	dbPath := fileio.ResolveDbPath()
	if err := store.SaveData(dbPath); err != nil {
		return fmt.Errorf("error saving database: %w", err)
	}
	return nil
}

// ResolveEntry attempts to locate a directory via its explicit map ID, assigned alias, or target path.
func ResolveEntry(query string) (string, config.DirData, bool) {
	store := GetDataStore()
	if entry, ok := store.TrackedDirs[query]; ok {
		return query, entry, true
	}

	if id, entry, ok := store.FindDirByAlias(query); ok {
		return id, entry, true
	}

	if id, entry, ok := store.FindDirByPath(query); ok {
		return id, entry, true
	}

	return "", config.DirData{}, false
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked directories",
	Run: func(cmd *cobra.Command, args []string) {
		entries := GetSortedTrackedDirectories()
		if len(entries) == 0 {
			cmd.Println("No tracked directories. Use 'mns add <path>' to add one.")
			return
		}

		cmd.Println("Tracked directories:")
		for _, entry := range entries {
			cmd.Printf("  %s | %-20s | %s\n", entry.ID, entry.Alias, entry.TargetPath)
		}
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm <id-or-alias>",
	Short: "Remove a tracked directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := RemoveTrackedDirectory(args[0]); err != nil {
			return err
		}
		cmd.Println("Removed successfully.")
		return nil
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all tracked directories (with confirmation)",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := GetDataStore()
		if len(store.TrackedDirs) == 0 {
			cmd.Println("No tracked directories to clear.")
			return nil
		}

		cmd.Printf("Warning: This will remove all %d tracked directories.\n", len(store.TrackedDirs))
		cmd.Print("Are you sure? [y/N]: ")

		var response string
		_, err := fmt.Fscanln(cmd.InOrStdin(), &response)
		if err != nil || (strings.ToLower(response) != "y" && strings.ToLower(response) != "yes") {
			cmd.Println("Aborted.")
			return nil
		}

		if err := ClearAllTrackedDirectories(); err != nil {
			return err
		}
		cmd.Println("All tracked directories removed.")
		return nil
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tracked directories by path or alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		matches := SearchTrackedDirectories(args[0])
		if len(matches) == 0 {
			cmd.Printf("No entries matching '%s'.\n", args[0])
			return
		}

		cmd.Println("Matching entries:")
		for _, entry := range matches {
			cmd.Printf("  %s | %-20s | %s\n", entry.ID, entry.Alias, entry.TargetPath)
		}
	},
}

var changeCmd = &cobra.Command{
	Use:   "change <id-or-alias>",
	Short: "Change the path or alias of a tracked directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		newPath, _ := cmd.Flags().GetString("path")
		newAlias, _ := cmd.Flags().GetString("alias")

		if err := ChangeTrackedDirectory(args[0], newPath, newAlias); err != nil {
			return err
		}
		cmd.Println("Updated entry successfully.")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(rmCmd)
	RootCmd.AddCommand(clearCmd)
	RootCmd.AddCommand(searchCmd)
	RootCmd.AddCommand(changeCmd)

	changeCmd.Flags().String("path", "", "New path for the tracked directory")
	changeCmd.Flags().String("alias", "", "New alias for the tracked directory")
}
