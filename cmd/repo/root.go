package repo

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "repo",
	Short: "Repository tooling for aw-runner",
}

func Execute() error {
	return rootCmd.Execute()
}
