// Package projects keeps the machine-local registry of prep projects: every
// project root a prep command ran in, with the time it was last used. It is
// state, not a choice, so it lives in $XDG_STATE_HOME/prep/projects.json
// (default ~/.local/state/prep/projects.json), not in the user configuration.
// Clients such as the desktop app list it to switch between projects.
package projects

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Project is one recorded project root.
type Project struct {
	Root     string    `json:"root"`
	Name     string    `json:"name"`
	LastUsed time.Time `json:"last_used"`
}

// file is the registry on disk.
type file struct {
	Projects []Project `json:"projects"`
}

// fresh is how long a touch stays current: touching a root again within it
// writes nothing, so a burst of commands costs one write.
const fresh = time.Minute

// Path returns the registry location.
func Path() (string, error) {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, "prep", "projects.json"), nil
}

// Touch records root as used at at.
func Touch(root string, at time.Time) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	ps, err := load()
	if err != nil {
		return err
	}
	at = at.UTC().Truncate(time.Second)
	for k, p := range ps {
		if p.Root == root {
			if at.Sub(p.LastUsed) < fresh {
				return nil
			}
			ps[k].LastUsed = at
			return save(ps)
		}
	}
	return save(append(ps, Project{Root: root, Name: filepath.Base(root), LastUsed: at}))
}

// List returns the recorded projects that still exist, newest first. It
// rewrites the registry when it dropped missing ones.
func List() ([]Project, error) {
	ps, err := load()
	if err != nil {
		return nil, err
	}
	var kept []Project
	for _, p := range ps {
		if fi, err := os.Stat(filepath.Join(p.Root, ".prep")); err == nil && fi.IsDir() {
			kept = append(kept, p)
		}
	}
	if len(kept) != len(ps) {
		if err := save(kept); err != nil {
			return nil, err
		}
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].LastUsed.After(kept[j].LastUsed) })
	if kept == nil {
		kept = []Project{}
	}
	return kept, nil
}

func load() ([]Project, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var f file
	if err := json.Unmarshal(b, &f); err != nil {
		// A damaged registry is rebuilt from the next touches.
		return nil, nil
	}
	return f.Projects, nil
}

// save writes the registry atomically, creating its directory.
func save(ps []Project) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	if ps == nil {
		ps = []Project{}
	}
	b, err := json.MarshalIndent(file{Projects: ps}, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), ".projects-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), p)
}
