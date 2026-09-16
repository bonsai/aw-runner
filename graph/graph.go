package graph

import (
	"github.com/bonsai/aw-runner/repos"
	"math"
	"sort"
)

type Edge struct {
	From, To string
	Weight   float64
}
type Node struct {
	ID, Label string
	Group     int
}
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Build creates a self-organizing similarity graph. Similarity combines lexical
// TF-style features from name/description/language/topics with a topic bonus.
// Communities are then discovered by label propagation, so no department names
// are hard-coded.
func Build(rs []repos.Repo) Graph {
	vecs := make([]map[string]float64, len(rs))
	for i, r := range rs {
		vecs[i] = vector(rs, r)
	}
	var edges []Edge
	for i := range rs {
		for j := i + 1; j < len(rs); j++ {
			w := cosine(vecs[i], vecs[j])
			if sharedTopics(rs[i], rs[j]) > 0 {
				w += 0.15
			}
			if w >= 0.18 {
				edges = append(edges, Edge{rs[i].FullName, rs[j].FullName, math.Round(w*1000) / 1000})
			}
		}
	}
	groups := propagate(rs, edges)
	nodes := make([]Node, len(rs))
	for i, r := range rs {
		nodes[i] = Node{r.FullName, r.Name, groups[r.FullName]}
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Group == nodes[j].Group {
			return nodes[i].ID < nodes[j].ID
		}
		return nodes[i].Group < nodes[j].Group
	})
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Weight == edges[j].Weight {
			return edges[i].From < edges[j].From
		}
		return edges[i].Weight > edges[j].Weight
	})
	return Graph{nodes, edges}
}

func vector(all []repos.Repo, r repos.Repo) map[string]float64 {
	terms := repos.Tokens(r)
	v := map[string]float64{}
	for _, t := range terms {
		v[t]++
	}
	for t := range v {
		df := 0
		for _, x := range all {
			for _, xt := range repos.Tokens(x) {
				if xt == t {
					df++
					break
				}
			}
		}
		v[t] *= math.Log(1 + float64(len(all))/float64(1+df))
	}
	if r.Language != "" {
		v["lang:"+r.Language] += 2
	}
	return v
}
func cosine(a, b map[string]float64) float64 {
	var dot, aa, bb float64
	for k, x := range a {
		dot += x * b[k]
		aa += x * x
	}
	for _, x := range b {
		bb += x * x
	}
	if aa == 0 || bb == 0 {
		return 0
	}
	return dot / math.Sqrt(aa*bb)
}
func sharedTopics(a, b repos.Repo) int {
	n := 0
	for _, x := range a.Topics {
		for _, y := range b.Topics {
			if x == y {
				n++
			}
		}
	}
	return n
}

func propagate(rs []repos.Repo, edges []Edge) map[string]int {
	adj := map[string][]string{}
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
		adj[e.To] = append(adj[e.To], e.From)
	}
	label := map[string]string{}
	for _, r := range rs {
		label[r.FullName] = r.FullName
	}
	for iter := 0; iter < 20; iter++ {
		changed := false
		for _, r := range rs {
			counts := map[string]int{}
			for _, n := range adj[r.FullName] {
				counts[label[n]]++
			}
			best := label[r.FullName]
			max := 0
			for k, v := range counts {
				if v > max || (v == max && k < best) {
					best, max = k, v
				}
			}
			if best != label[r.FullName] {
				label[r.FullName] = best
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	ids := map[string]int{}
	next := 0
	out := map[string]int{}
	for _, r := range rs {
		k := label[r.FullName]
		if _, ok := ids[k]; !ok {
			ids[k] = next
			next++
		}
		out[r.FullName] = ids[k]
	}
	return out
}
