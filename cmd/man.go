package cmd

import (
	"bytes"
	"fmt"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

var manForce bool

// manCmd represents the man command
var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generates the manual page for mnemosync",
	Long: `Generates and displays manual page for mnemosync,
persisting it to ~/.local/share/man/man1/mns.1 for system man(1) access.

Use --force to overwrite the local man page even if unchanged.`,
	Run: func(cmd *cobra.Command, args []string) {
		DisplayManPage()
	},
}

func generateManPage() (string, error) {
	manPage, err := mcobra.NewManPage(1, RootCmd)
	if err != nil {
		return "", err
	}

	manPage = manPage.WithSection("Copyright", "(C) 2025 bladeacer.\n"+
		"Released under GPLv3 license.")

	return manPage.Build(roff.NewDocument()), nil
}

func PersistManPage(content string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	manDir := filepath.Join(homeDir, ".local", "share", "man", "man1")
	manPath := filepath.Join(manDir, "mns.1")

	if err := os.MkdirAll(manDir, 0755); err != nil {
		return fmt.Errorf("cannot create man directory: %w", err)
	}

	if !manForce {
		existing, err := os.ReadFile(manPath)
		if err == nil && string(existing) == content {
			return nil
		}
	}

	if err := os.WriteFile(manPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write man page: %w", err)
	}

	_ = exec.Command("mandb", "-q").Run()

	return nil
}

func DisplayManPage() {
	manContent, err := generateManPage()
	if err != nil {
		panic(err)
	}

	if err := PersistManPage(manContent); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not persist man page to man-db: %v\n", err)
	}

	var buf bytes.Buffer
	buf.WriteString(manContent)

	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	manCmd := exec.Command("man", "-l", "-")
	manCmd.Stdin = &buf
	manCmd.Stdout = os.Stdout
	manCmd.Stderr = os.Stderr

	if err := manCmd.Run(); err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "Error running 'man' command, falling back to 'nroff'.\n")

	nroffCmd := exec.Command("nroff", "-man")
	nroffCmd.Stdin = &buf
	nroffCmd.Stdout = os.Stdout
	nroffCmd.Stderr = os.Stderr

	if err := nroffCmd.Run(); err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "Error running 'nroff', falling back to pager (%s).\n", pager)

	pagerCmd := exec.Command(pager)
	pagerCmd.Stdin = &buf
	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr

	if err := pagerCmd.Run(); err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "Error running pager (%s), displaying raw content.\n", pager)
	fmt.Println(manContent)
}

func init() {
	RootCmd.AddCommand(manCmd)
	manCmd.Flags().BoolVarP(&manForce, "force", "f", false, "Force overwrite of existing man page")
}
