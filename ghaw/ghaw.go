package ghaw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Workflow is a single `gh aw list --json` entry (simplified table).
type Workflow struct {
	Workflow string `json:"workflow"`
	EngineID string `json:"engine_id"`
	Compiled string `json:"compiled"`
}

// WorkflowStatus is a single `gh aw status --json` entry.
type WorkflowStatus struct {
	Workflow      string `json:"workflow"`
	EngineID      string `json:"engine_id"`
	Compiled      string `json:"compiled"`
	Status        string `json:"status"`
	TimeRemaining string `json:"time_remaining"`
}

// RunResult is a single `gh aw run --json` entry.
type RunResult struct {
	Workflow string `json:"workflow"`
	LockFile string `json:"lock_file"`
	Status   string `json:"status"`
}

// Exec runs `gh aw <args...>` and returns stdout. stderr is merged into errors
// so the runner surfaces gh's diagnostic stream instead of swallowing it.
func Exec(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", append([]string{"aw"}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return stdout.Bytes(), fmt.Errorf("gh aw %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.Bytes(), nil
}

// List returns the agentic workflows known to gh aw. An empty repo uses the
// current repository (same default as `gh aw list`).
func List(repo, pattern string) ([]Workflow, error) {
	args := []string{"list", "--json"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if pattern != "" {
		args = append(args, pattern)
	}
	out, err := Exec(args...)
	if err != nil {
		return nil, err
	}
	var wf []Workflow
	if err := json.Unmarshal(out, &wf); err != nil {
		return nil, fmt.Errorf("parse gh aw list: %w", err)
	}
	return wf, nil
}

// Status returns the status of the agentic workflows in the repository.
func Status(repo, pattern string) ([]WorkflowStatus, error) {
	args := []string{"status", "--json"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if pattern != "" {
		args = append(args, pattern)
	}
	out, err := Exec(args...)
	if err != nil {
		return nil, err
	}
	var st []WorkflowStatus
	if err := json.Unmarshal(out, &st); err != nil {
		return nil, fmt.Errorf("parse gh aw status: %w", err)
	}
	return st, nil
}

// Run dispatches one agentic workflow through `gh aw run` (workflow_dispatch).
// dryRun previews the trigger without executing it on GitHub Actions.
func Run(repo, workflow string, dryRun bool, fields []string) ([]RunResult, error) {
	args := []string{"run", workflow, "--json"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if dryRun {
		args = append(args, "--dry-run")
	}
	for _, f := range fields {
		args = append(args, "--raw-field", f)
	}
	out, err := Exec(args...)
	if err != nil {
		return nil, err
	}
	var res []RunResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, fmt.Errorf("parse gh aw run: %w", err)
	}
	return res, nil
}

// Logs downloads and analyzes workflow run logs/artifacts via `gh aw logs`.
// The overview report is returned; artifacts are extracted into <dir>.
func Logs(repo, workflow, dir string) (string, error) {
	args := []string{"logs"}
	if workflow != "" {
		args = append(args, workflow)
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if dir != "" {
		args = append(args, "-o", dir)
	}
	out, err := Exec(args...)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
