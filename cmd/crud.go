package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bladeacer/mmsync/config"
	"github.com/bladeacer/mmsync/internal/fileio"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked directories",
	Run: func(cmd *cobra.Command, args []string) {
		if len(DataStore.TrackedDirs) == 0 {
			fmt.Println("No tracked directories. Use 'mns add <path>' to add one.")
			return
		}

		ids := make([]string, 0, len(DataStore.TrackedDirs))
		for id := range DataStore.TrackedDirs {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			ai, bi := ids[i], ids[j]
			return len(ai) < len(bi) || (len(ai) == len(bi) && ai < bi)
		})

		fmt.Println("Tracked directories:")
		for _, id := range ids {
			entry := DataStore.TrackedDirs[id]
			fmt.Printf("  %s | %-20s | %s\n", id, entry.Alias, entry.TargetPath)
		}
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm <id-or-alias>",
	Short: "Remove a tracked directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		id, entry, found := ResolveEntry(query)
		if !found {
			fmt.Fprintf(os.Stderr, "Error: no tracked directory matches '%s'\n", query)
			os.Exit(1)
		}

		fmt.Printf("Removing:\n  ID: %s | %s | %s\n", id, entry.Alias, entry.TargetPath)

		if !DataStore.DeleteDir(id) {
			fmt.Fprintf(os.Stderr, "Error: failed to remove entry '%s'\n", query)
			os.Exit(1)
		}

		dbPath := fileio.ResolveDbPath()
		if err := DataStore.SaveData(dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Removed successfully.")
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all tracked directories (with confirmation)",
	Run: func(cmd *cobra.Command, args []string) {
		if len(DataStore.TrackedDirs) == 0 {
			fmt.Println("No tracked directories to clear.")
			return
		}

		fmt.Printf("Warning: This will remove all %d tracked directories.\n", len(DataStore.TrackedDirs))
		fmt.Print("Are you sure? [y/N]: ")

		var response string
		_, err := fmt.Scanln(&response)
		if err != nil || (strings.ToLower(response) != "y" && strings.ToLower(response) != "yes") {
			fmt.Println("Aborted.")
			return
		}

		DataStore.TrackedDirs = make(map[string]config.DirData)
		dbPath := fileio.ResolveDbPath()
		if err := DataStore.SaveData(dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All tracked directories removed.")
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tracked directories by path or alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.ToLower(args[0])
		found := false

		ids := make([]string, 0, len(DataStore.TrackedDirs))
		for id := range DataStore.TrackedDirs {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			ai, bi := ids[i], ids[j]
			return len(ai) < len(bi) || (len(ai) == len(bi) && ai < bi)
		})

		for _, id := range ids {
			entry := DataStore.TrackedDirs[id]
			if strings.Contains(strings.ToLower(entry.Alias), query) ||
				strings.Contains(strings.ToLower(entry.TargetPath), query) ||
				strings.Contains(id, query) {
				if !found {
					fmt.Println("Matching entries:")
					found = true
				}
				fmt.Printf("  %s | %-20s | %s\n", id, entry.Alias, entry.TargetPath)
			}
		}

		if !found {
			fmt.Printf("No entries matching '%s'.\n", query)
		}
	},
}

var changeCmd = &cobra.Command{
	Use:   "change <id-or-alias>",
	Short: "Change the path or alias of a tracked directory",
	Long: `Change the path or alias of a tracked directory.
At least one of --path or --alias must be provided.

Examples:
  mns change 1 --path /new/path
  mns change myalias --alias newalias
  mns change 1 --path /new/path --alias newalias`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		newPath, _ := cmd.Flags().GetString("path")
		newAlias, _ := cmd.Flags().GetString("alias")

		if newPath == "" && newAlias == "" {
			fmt.Fprintf(os.Stderr, "Error: at least one of --path or --alias must be provided.\n")
			os.Exit(1)
		}

		id, entry, found := ResolveEntry(query)
		if !found {
			fmt.Fprintf(os.Stderr, "Error: no tracked directory matches '%s'\n", query)
			os.Exit(1)
		}

		if newPath != "" {
			resolved, err := ResolveAndValidatePath(newPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			entry.TargetPath = resolved
		}

		if newAlias != "" {
			if _, _, exists := DataStore.FindDirByAlias(newAlias); exists {
				fmt.Fprintf(os.Stderr, "Error: alias '%s' is already in use.\n", newAlias)
				os.Exit(1)
			}
			entry.Alias = newAlias
		}

		DataStore.TrackedDirs[id] = entry
		dbPath := fileio.ResolveDbPath()
		if err := DataStore.SaveData(dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Updated entry:")
		fmt.Printf("  ID: %s | %s | %s\n", id, entry.Alias, entry.TargetPath)
	},
}

func ResolveEntry(query string) (string, config.DirData, bool) {
	if entry, ok := DataStore.TrackedDirs[query]; ok {
		return query, entry, true
	}

	if id, entry, ok := DataStore.FindDirByAlias(query); ok {
		return id, entry, true
	}

	if id, entry, ok := DataStore.FindDirByPath(query); ok {
		return id, entry, true
	}

	return "", config.DirData{}, false
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
