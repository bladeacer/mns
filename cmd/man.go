package cmd

import (
	"bytes"
	"fmt"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

// manCmd represents the man command
var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generates the manual page for mnemosync",
	Long: `Generates and displays manual page for mnemosync

Does not persist it to a file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use a better name to avoid conflict with the standard man command
		DisplayManPage()
	},
}

// TODO: Persist the generated man page to local man-db once it is called,
// add a flag to force persisting to local man-db

func DisplayManPage() {
	manPage, err := mcobra.NewManPage(1, RootCmd)
	if err != nil {
		panic(err)
	}

	manPage = manPage.WithSection("Copyright", "(C) 2025 bladeacer.\n"+
		"Released under GPLv3 license.")

	// Get the generated man page content.
	manContent := manPage.Build(roff.NewDocument())

	// 1. Create a buffer to hold the man page content.
	var buf bytes.Buffer
	buf.WriteString(manContent)

	// 2. Get the user's preferred pager (like `less` or `more`).
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less" // Default to `less`
	}

	// 3. Set up the man page viewer command.
	// We use `man` as the viewer to get proper formatting, with `less` as the pager.
	// `man -l` command formats and displays a local man page file.
	// We pipe the content to `man`'s standard input.
	manCmd := exec.Command("man", "-l", "-")

	// Set the command's standard input to our buffer.
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
}
