package repo

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bonsai/aw-runner/repos"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// selectCmd presents the list of known repositories via `peco` fuzzy finder
// and returns the chosen repo identifier (full name). It assumes `peco`
// is installed and available in $PATH.
var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactively pick a repository using peco",
	RunE:  runSelect,
}

func init() {
	// allow owner override
	selectCmd.Flags().StringP("owner", "o", "", "GitHub owner / organization (default from GITHUB_OWNER env)")
	viper.BindPFlag("repo.owner", selectCmd.Flags().Lookup("owner"))
	rootCmd.AddCommand(selectCmd)
}

func runSelect(cmd *cobra.Command, _ []string) error {
	// Load repos from the ontology (inventory DB)
	repos, err := repos.FetchRecent(viper.GetString("repo.owner"), 30)
	if err != nil {
		return fmt.Errorf("failed to load repos: %w", err)
	}
	if len(repos) == 0 {
		return fmt.Errorf("no repositories found in inventory")
	}
	// Build slice of display strings "<full_name> - <description>"
	var lines []string
	for _, r := range repos {
		line := fmt.Sprintf("%s - %s", r.FullName, r.Description)
		lines = append(lines, line)
	}
	// Run peco and feed the list
	sel, err := runPeco(lines)
	if err != nil {
		return err
	}
	// Extract repo full name (before first space)
	chosen := strings.SplitN(sel, " ", 2)[0]
	fmt.Printf("Selected repository: %s\n", chosen)
	// Store selection in Viper for later sub‑commands (e.g. stream, health)
	viper.Set("repo.selected", chosen)
	return nil
}

// runPeco executes the peco binary, passing the provided lines via stdin
// and returns the line the user selected.
func runPeco(lines []string) (string, error) {
	cmd := exec.Command("peco")
	// Use a pipe for stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin pipe for peco: %w", err)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start peco: %w", err)
	}
	// Write lines
	w := bufio.NewWriter(stdin)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	w.Flush()
	stdin.Close()
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("peco exited with error: %w", err)
	}
	sel := strings.TrimSpace(out.String())
	return sel, nil
}
