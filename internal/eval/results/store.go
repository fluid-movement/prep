// Package results stores matrix runs on disk and reports and compares them.
package results

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/eval/matrix"
)

// VariantRef names a variant and the content hash that ran.
type VariantRef struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

// Definition is what a matrix run was asked to do.
type Definition struct {
	Cases []string `json:"cases"`
	// Areas maps each case to its area, so reports stand alone.
	Areas      map[string]string `json:"areas,omitempty"`
	Variants   []VariantRef      `json:"variants"`
	Models     []string          `json:"models"`
	N          int               `json:"n"`
	Seed       int64             `json:"seed"`
	Parallel   int               `json:"parallel"`
	Sequential bool              `json:"sequential,omitempty"`
	MaxUSD     float64           `json:"max_usd,omitempty"`
}

// Versions are what ran it.
type Versions struct {
	Prep    string `json:"prep"`
	Harness string `json:"harness,omitempty"`
	// Git is the repository revision the suite was run at, when known.
	Git string `json:"git,omitempty"`
}

// Run is a stored matrix run.
type Run struct {
	ID         string     `json:"id"`
	Created    time.Time  `json:"created"`
	Note       string     `json:"note,omitempty"`
	Definition Definition `json:"definition"`
	Versions   Versions   `json:"versions"`

	Cells     []*matrix.CellResult `json:"cells"`
	StoppedBy string               `json:"stopped_by,omitempty"`
	SpentUSD  float64              `json:"spent_usd"`
}

// Store keeps runs under Root/<run-id>/.
type Store struct{ Root string }

// NewID makes a run ID from the time and a few random characters.
func NewID(now time.Time) string {
	b := make([]byte, 2)
	rand.Read(b)
	return now.UTC().Format("20060102-150405") + "-" + hex.EncodeToString(b)
}

// Dir returns the directory of a run; per-run files go under its runs/.
func (s Store) Dir(id string) string { return filepath.Join(s.Root, id) }

// Save writes matrix.json (what was asked and what ran it) and cells.json
// (the results).
func (s Store) Save(r *Run) error {
	if r.ID == "" {
		return fmt.Errorf("results: run has no ID")
	}
	dir := s.Dir(r.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	head := struct {
		ID         string     `json:"id"`
		Created    time.Time  `json:"created"`
		Note       string     `json:"note,omitempty"`
		Definition Definition `json:"definition"`
		Versions   Versions   `json:"versions"`
	}{r.ID, r.Created, r.Note, r.Definition, r.Versions}
	if err := writeJSON(filepath.Join(dir, "matrix.json"), head); err != nil {
		return err
	}
	return writeJSON(filepath.Join(dir, "cells.json"), struct {
		Cells     []*matrix.CellResult `json:"cells"`
		StoppedBy string               `json:"stopped_by,omitempty"`
		SpentUSD  float64              `json:"spent_usd"`
	}{r.Cells, r.StoppedBy, r.SpentUSD})
}

func writeJSON(path string, v any) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".tmp-")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// Load reads a stored run. The ID may be abbreviated to a unique prefix, or
// "latest" for the newest.
func (s Store) Load(id string) (*Run, error) {
	full, err := s.resolve(id)
	if err != nil {
		return nil, err
	}
	var r Run
	if err := readJSON(filepath.Join(s.Dir(full), "matrix.json"), &r); err != nil {
		return nil, err
	}
	var c struct {
		Cells     []*matrix.CellResult `json:"cells"`
		StoppedBy string               `json:"stopped_by"`
		SpentUSD  float64              `json:"spent_usd"`
	}
	if err := readJSON(filepath.Join(s.Dir(full), "cells.json"), &c); err != nil {
		return nil, err
	}
	r.Cells, r.StoppedBy, r.SpentUSD = c.Cells, c.StoppedBy, c.SpentUSD
	for _, cr := range r.Cells {
		matrix.Reduce(cr)
	}
	return &r, nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (s Store) resolve(id string) (string, error) {
	ids, err := s.IDs()
	if err != nil {
		return "", err
	}
	if id == "latest" {
		if len(ids) == 0 {
			return "", fmt.Errorf("no stored runs in %s", s.Root)
		}
		return ids[0], nil
	}
	var match []string
	for _, x := range ids {
		if x == id {
			return x, nil
		}
		if len(id) > 0 && len(x) >= len(id) && x[:len(id)] == id {
			match = append(match, x)
		}
	}
	switch len(match) {
	case 0:
		return "", fmt.Errorf("no stored run %q in %s", id, s.Root)
	case 1:
		return match[0], nil
	}
	return "", fmt.Errorf("%q matches %d runs: %v", id, len(match), match)
}

// IDs lists the stored run IDs, newest first.
func (s Store) IDs() ([]string, error) {
	entries, err := os.ReadDir(s.Root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(s.Root, e.Name(), "matrix.json")); err == nil {
				ids = append(ids, e.Name())
			}
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	return ids, nil
}

// Prune removes all but the newest keep runs, including their kept environments,
// and returns the removed IDs.
func (s Store) Prune(keep int) ([]string, error) {
	ids, err := s.IDs()
	if err != nil {
		return nil, err
	}
	if keep < 0 {
		keep = 0
	}
	if len(ids) <= keep {
		return nil, nil
	}
	var removed []string
	for _, id := range ids[keep:] {
		r, err := s.Load(id)
		if err != nil {
			return removed, err
		}
		for _, cell := range r.Cells {
			for _, run := range cell.Runs {
				if err := removeWorkspace(run.Dir); err != nil {
					return removed, err
				}
			}
		}
		if err := os.RemoveAll(s.Dir(id)); err != nil {
			return removed, err
		}
		removed = append(removed, id)
	}
	return removed, nil
}

// Merge combines runs into one: the runs of a cell in several stored runs are
// pooled (so a model compared in a later run, or a cell topped up with more
// runs, reads as one result). The ID joins the sources.
func Merge(runs ...*Run) *Run {
	if len(runs) == 1 {
		return runs[0]
	}
	out := &Run{Definition: Definition{Areas: map[string]string{}}}
	cells := map[matrix.Cell]*matrix.CellResult{}
	var order []matrix.Cell
	seenCase, seenModel := map[string]bool{}, map[string]bool{}
	seenVar := map[string]bool{}
	var ids []string
	for _, r := range runs {
		ids = append(ids, r.ID)
		if out.Created.IsZero() || r.Created.Before(out.Created) {
			out.Created = r.Created
		}
		if out.Versions.Prep == "" {
			out.Versions = r.Versions
		}
		out.SpentUSD += r.SpentUSD
		if r.Definition.N > out.Definition.N {
			out.Definition.N = r.Definition.N
		}
		for k, v := range r.Definition.Areas {
			out.Definition.Areas[k] = v
		}
		for _, c := range r.Definition.Cases {
			if !seenCase[c] {
				seenCase[c] = true
				out.Definition.Cases = append(out.Definition.Cases, c)
			}
		}
		for _, m := range r.Definition.Models {
			if !seenModel[m] {
				seenModel[m] = true
				out.Definition.Models = append(out.Definition.Models, m)
			}
		}
		for _, v := range r.Definition.Variants {
			if !seenVar[v.Name] {
				seenVar[v.Name] = true
				out.Definition.Variants = append(out.Definition.Variants, v)
			}
		}
		for _, c := range r.Cells {
			into := cells[c.Cell]
			if into == nil {
				into = &matrix.CellResult{Cell: c.Cell}
				cells[c.Cell] = into
				order = append(order, c.Cell)
			}
			for _, run := range c.Runs {
				run.Index = len(into.Runs)
				into.Runs = append(into.Runs, run)
			}
			into.Planned += c.Planned
		}
	}
	for _, k := range order {
		matrix.Reduce(cells[k])
		out.Cells = append(out.Cells, cells[k])
	}
	out.ID = strings.Join(ids, "+")
	return out
}

// LoadAll loads several runs by ID and merges them.
func (s Store) LoadAll(ids ...string) (*Run, error) {
	var runs []*Run
	for _, id := range ids {
		r, err := s.Load(id)
		if err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	if len(runs) == 0 {
		return nil, fmt.Errorf("no run given")
	}
	return Merge(runs...), nil
}
