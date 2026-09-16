package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bonsai/aw-runner/ghaw"
	"github.com/spf13/cobra"
)

var repo string

func main() {
	root := &cobra.Command{
		Use:   "aw",
		Short: "gh aw control surface — dispatch / run / status / logs",
		Long:  "aw is the aw-runner boundary over the `gh aw` CLI.\nIt does not think or solve; it dispatches and executes.",
	}
	root.PersistentFlags().StringVarP(&repo, "repo", "r", "", "target repository ([HOST/]owner/repo). Default: current")

	listCmd := &cobra.Command{
		Use:   "list [pattern]",
		Short: "List agentic workflows",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := ""
			if len(args) == 1 {
				pattern = args[0]
			}
			wf, err := ghaw.List(repo, pattern)
			if err != nil {
				return err
			}
			for _, w := range wf {
				fmt.Printf("%-32s engine=%-12s compiled=%s\n", w.Workflow, emptyDash(w.EngineID), emptyDash(w.Compiled))
			}
			return nil
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status [pattern]",
		Short: "Show status of agentic workflows",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := ""
			if len(args) == 1 {
				pattern = args[0]
			}
			st, err := ghaw.Status(repo, pattern)
			if err != nil {
				return err
			}
			for _, s := range st {
				t := s.TimeRemaining
				if t != "" {
					t = " " + t
				}
				fmt.Printf("%-32s engine=%-12s %s%s\n", s.Workflow, emptyDash(s.EngineID), s.Status, t)
			}
			return nil
		},
	}

	runCmd := &cobra.Command{
		Use:   "run <workflow>",
		Short: "Dispatch an agentic workflow (workflow_dispatch)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			fields, _ := cmd.Flags().GetStringArray("raw-field")
			res, err := ghaw.Run(repo, args[0], dryRun, fields)
			if err != nil {
				return err
			}
			for _, r := range res {
				state := r.Status
				if dryRun {
					state = "preview"
				}
				fmt.Printf("%s → %s (%s)\n", r.Workflow, state, emptyDash(r.LockFile))
			}
			return nil
		},
	}
	runCmd.Flags().Bool("dry-run", false, "preview the run without triggering GitHub Actions")
	runCmd.Flags().StringArray("raw-field", nil, "workflow input as name=value (repeatable)")

	logsCmd := &cobra.Command{
		Use:   "logs [workflow]",
		Short: "Download and analyze workflow logs / artifacts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflow := ""
			if len(args) == 1 {
				workflow = args[0]
			}
			dir, _ := cmd.Flags().GetString("output")
			out, err := ghaw.Logs(repo, workflow, dir)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
	logsCmd.Flags().StringP("output", "o", "", "output directory for extracted artifacts")

	root.AddCommand(listCmd, statusCmd, runCmd, logsCmd)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func emptyDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
