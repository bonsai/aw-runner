package repos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

type Repo struct {
	Name          string   `json:"name"`
	FullName      string   `json:"full_name"`
	Description   string   `json:"description"`
	HTMLURL       string   `json:"html_url"`
	Language      string   `json:"language"`
	DefaultBranch string   `json:"default_branch"`
	Topics        []string `json:"topics"`
	Stars         int      `json:"stargazers_count"`
	Forks         int      `json:"forks_count"`
	UpdatedAt     string   `json:"updated_at"`
	Archived      bool     `json:"archived"`
	Fork          bool     `json:"fork"`
	Owner         string   `json:"-"`
}

// githubRepo is the API shape we decode. GitHub's owner field is an object;
// Repo.Owner stays a simple string because the graph does not need the object.
type githubRepo struct {
	Repo
	Owner githubOwner `json:"owner"`
}

type githubOwner struct {
	Login string `json:"login"`
}

// FetchRecent returns at most limit repositories ordered by GitHub's updated_at.
// It intentionally keeps the observation window small: the graph is a live view,
// not a historical warehouse.
func FetchRecent(owner string, limit int) ([]Repo, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 30 {
		limit = 30
	}

	endpoint := "https://api.github.com/users/" + url.PathEscape(owner) + "/repos?sort=updated&direction=desc&per_page=30&type=owner"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API: %s", resp.Status)
	}

	var raw []githubRepo
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	rs := make([]Repo, 0, len(raw))
	for _, item := range raw {
		r := item.Repo
		r.Owner = item.Owner.Login
		rs = append(rs, r)
	}
	if len(rs) > limit {
		rs = rs[:limit]
	}

	// Deterministic order makes graph snapshots diff-friendly.
	sort.Slice(rs, func(i, j int) bool { return rs[i].FullName < rs[j].FullName })
	return rs, nil
}

func Tokens(r Repo) []string {
	text := strings.ToLower(strings.Join([]string{r.Name, r.Description, r.Language, strings.Join(r.Topics, " ")}, " "))
	fields := strings.FieldsFunc(text, func(ch rune) bool {
		return !(ch >= 'a' && ch <= 'z') && !(ch >= '0' && ch <= '9')
	})
	seen := map[string]bool{}
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if len(f) < 2 || seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}
