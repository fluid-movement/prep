// Package projects keeps the machine-local registry of prep projects: every
// project root a prep command ran in, with the time it was last used. It is
// state, not a choice, so it lives in $XDG_STATE_HOME/prep/projects.json
// (default ~/.local/state/prep/projects.json), not in the user configuration.
// Clients such as the TUI list it to switch between
// projects, and the CLI resolves --project and cross-project links through it.
//
// A project's identity is the permanent ID in its project.md, shared by every
// clone; its name is a handle and its remote only says where to clone it from.
package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
	"github.com/fluid-movement/prep/internal/mdstore"
)

// Project is one recorded project root. ID, Name and Remote are read from
// the checkout whenever the registry is listed; the file keeps the last
// values seen.
type Project struct {
	Root string `json:"root"`
	// ID is the permanent project ID from project.md; empty for a project
	// from before IDs (prep fix writes one).
	ID string `json:"id,omitempty"`
	// Name is project.md's name, else DefaultName.
	Name string `json:"name"`
	// Remote is origin normalized to host/owner/repo; empty without one.
	Remote   string    `json:"remote,omitempty"`
	LastUsed time.Time `json:"last_used"`
	// Collision marks a name that projects with different IDs share; it is
	// derived when listing, never stored.
	Collision bool `json:"collision,omitempty"`
}

// key identifies the project a root belongs to: clones share their ID.
func (p Project) key() string {
	if p.ID != "" {
		return p.ID
	}
	return "root:" + p.Root
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
	p := Describe(root)
	p.LastUsed = at
	return save(append(ps, p))
}

// Add registers the project at dir, which prep may not have run in yet, as
// used at at. It changes nothing in the project.
func Add(dir string, at time.Time) (Project, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return Project{}, err
	}
	if _, err := os.Stat(filepath.Join(root, mdstore.Dir, "project.md")); err != nil {
		return Project{}, &domain.Error{Code: domain.ErrNoProject, Message: fmt.Sprintf("%s has no %s/project.md; run prep init there", root, mdstore.Dir)}
	}
	ps, err := load()
	if err != nil {
		return Project{}, err
	}
	p := Describe(root)
	p.LastUsed = at.UTC().Truncate(time.Second)
	kept := ps[:0]
	for _, q := range ps {
		if q.Root != root {
			kept = append(kept, q)
		}
	}
	return p, save(append(kept, p))
}

// Remove forgets the projects query names: a root path, or a name or ID
// (Resolve's rules, but every clone of the project). It changes nothing in
// the projects and returns the roots it forgot.
func Remove(query string) ([]string, error) {
	ps, err := List()
	if err != nil {
		return nil, err
	}
	gone := map[string]bool{}
	if abs, err := filepath.Abs(query); err == nil {
		for _, p := range ps {
			if p.Root == abs {
				gone[p.Root] = true
			}
		}
	}
	if len(gone) == 0 {
		p, err := resolveIn(ps, query)
		if err != nil {
			return nil, err
		}
		for _, q := range ps {
			if q.key() == p.key() {
				gone[q.Root] = true
			}
		}
	}
	all, err := load()
	if err != nil {
		return nil, err
	}
	var kept []Project
	var roots []string
	for _, p := range all {
		if gone[p.Root] {
			roots = append(roots, p.Root)
			continue
		}
		kept = append(kept, p)
	}
	return roots, save(kept)
}

// List returns the recorded projects that still exist, newest first, with
// their identity read fresh from each checkout. It rewrites the registry
// when it dropped missing ones or an identity changed.
func List() ([]Project, error) {
	ps, err := load()
	if err != nil {
		return nil, err
	}
	changed := false
	var kept []Project
	for _, p := range ps {
		fi, err := os.Stat(filepath.Join(p.Root, mdstore.Dir))
		if errors.Is(err, fs.ErrNotExist) || (err == nil && !fi.IsDir()) {
			changed = true
			continue
		}
		if err != nil {
			kept = append(kept, p)
			continue
		}
		d := Describe(p.Root)
		if d.ID != p.ID || d.Name != p.Name || d.Remote != p.Remote {
			p.ID, p.Name, p.Remote = d.ID, d.Name, d.Remote
			changed = true
		}
		kept = append(kept, p)
	}
	if changed {
		if err := save(kept); err != nil {
			return nil, err
		}
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].LastUsed.After(kept[j].LastUsed) })
	keys := map[string]map[string]bool{}
	for _, p := range kept {
		if keys[p.Name] == nil {
			keys[p.Name] = map[string]bool{}
		}
		keys[p.Name][p.key()] = true
	}
	for k := range kept {
		kept[k].Collision = len(keys[kept[k].Name]) > 1
	}
	if kept == nil {
		kept = []Project{}
	}
	return kept, nil
}

// Describe reads the identity of the project at root: ID and name from its
// project.md, remote from git. It never fails; what it cannot read stays
// empty, and the name falls back to DefaultName.
func Describe(root string) Project {
	p := Project{Root: root, Remote: NormalizeRemote(gitx.Remote(root))}
	if pr, err := mdstore.ReadProject(root); err == nil {
		p.ID, p.Name = pr.ID, pr.Name
	}
	if p.Name == "" {
		p.Name = DefaultName(root)
	}
	return p
}

// Resolve finds the registered project query names: an exact name, else an
// ID or a unique suffix of one, else a remote, else a path to a checkout.
// Clones of one project resolve to the most recently used. A name that
// several projects share, or a suffix of several IDs, is ambiguous.
func Resolve(query string) (Project, error) {
	ps, err := List()
	if err != nil {
		return Project{}, err
	}
	return resolveIn(ps, query)
}

// ByID returns the most recently used checkout of the project with the
// given ID.
func ByID(id string) (Project, bool, error) {
	ps, err := List()
	if err != nil {
		return Project{}, false, err
	}
	for _, p := range ps {
		if p.ID == id {
			return p, true, nil
		}
	}
	return Project{}, false, nil
}

// resolveIn applies Resolve's rules to ps, newest first.
func resolveIn(ps []Project, query string) (Project, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return Project{}, &domain.Error{Code: domain.ErrUsage, Message: "no project given; prep projects lists them"}
	}
	// pick returns the newest of the matches, or an ambiguity error when
	// they belong to several projects.
	pick := func(what string, match func(Project) bool) (Project, bool, error) {
		var hits []Project
		seen := map[string]bool{}
		for _, p := range ps {
			if match(p) && !seen[p.key()] {
				seen[p.key()] = true
				hits = append(hits, p)
			}
		}
		switch len(hits) {
		case 0:
			return Project{}, false, nil
		case 1:
			return hits[0], true, nil
		}
		var c []string
		for _, h := range hits {
			c = append(c, fmt.Sprintf("%s (%s)", h.Root, describeID(h)))
		}
		return Project{}, false, &domain.Error{Code: domain.ErrAmbiguous,
			Message: fmt.Sprintf("%s %q matches several projects: %s; give an ID or a path", what, q, strings.Join(c, ", "))}
	}
	if p, ok, err := pick("name", func(p Project) bool { return p.Name == q }); ok || err != nil {
		return p, err
	}
	if up := strings.ToUpper(q); len(up) >= 4 && len(up) <= domain.ULIDLen {
		if p, ok, err := pick("ID", func(p Project) bool { return p.ID != "" && strings.HasSuffix(p.ID, up) }); ok || err != nil {
			return p, err
		}
	}
	if r := NormalizeRemote(q); r != "" {
		if p, ok, err := pick("remote", func(p Project) bool { return p.Remote == r }); ok || err != nil {
			return p, err
		}
	}
	if abs, err := filepath.Abs(q); err == nil {
		if fi, err := os.Stat(filepath.Join(abs, mdstore.Dir)); err == nil && fi.IsDir() {
			for _, p := range ps {
				if p.Root == abs {
					return p, nil
				}
			}
			return Describe(abs), nil
		}
	}
	var names []string
	seen := map[string]bool{}
	for _, p := range ps {
		if !seen[p.Name] {
			seen[p.Name] = true
			names = append(names, p.Name)
		}
	}
	known := "none are registered; prep projects add <dir> registers one"
	if len(names) > 0 {
		known = "registered: " + strings.Join(names, ", ")
	}
	return Project{}, &domain.Error{Code: domain.ErrNotFound, Message: fmt.Sprintf("no project %q (%s)", q, known)}
}

func describeID(p Project) string {
	if p.ID == "" {
		return "no id"
	}
	return "id …" + p.ID[len(p.ID)-6:]
}

// NormalizeRemote turns a git remote URL into host/owner/repo, so the SSH,
// scp-like and HTTPS forms of one repository are equal: no scheme, user,
// port or .git suffix, the host in lower case. A local path stays a cleaned
// path. An empty URL gives "".
func NormalizeRemote(url string) string {
	u := strings.TrimSpace(url)
	if u == "" {
		return ""
	}
	u = strings.TrimSuffix(strings.TrimSuffix(u, "/"), ".git")
	host, path := "", ""
	switch {
	case strings.Contains(u, "://"):
		rest := u[strings.Index(u, "://")+3:]
		if strings.HasPrefix(u, "file://") {
			return filepath.Clean(rest)
		}
		host, path, _ = strings.Cut(rest, "/")
		if at := strings.LastIndex(host, "@"); at >= 0 {
			host = host[at+1:]
		}
		if c := strings.Index(host, ":"); c >= 0 {
			host = host[:c]
		}
	case strings.Contains(u, ":") && !strings.HasPrefix(u, "/") && !strings.HasPrefix(u, "."):
		// scp-like: [user@]host:path
		host, path, _ = strings.Cut(u, ":")
		if at := strings.LastIndex(host, "@"); at >= 0 {
			host = host[at+1:]
		}
	default:
		return filepath.Clean(u)
	}
	path = strings.Trim(path, "/")
	if host == "" {
		return path
	}
	return strings.ToLower(host) + "/" + path
}

// DefaultName is the name of a project that declares none: the last path
// segment of its origin remote, else its directory name, lowered and with
// characters a name cannot hold replaced by '-'.
func DefaultName(root string) string {
	base := filepath.Base(root)
	if r := NormalizeRemote(gitx.Remote(root)); r != "" {
		base = r[strings.LastIndex(r, "/")+1:]
	}
	return SanitizeName(base)
}

// SanitizeName turns s into a valid project name.
func SanitizeName(s string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(s) {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-', c == '_', c == '.':
			b.WriteRune(c)
		default:
			b.WriteByte('-')
		}
	}
	n := strings.TrimLeft(b.String(), "-_.")
	if len(n) > 64 {
		n = n[:64]
	}
	if n == "" {
		return "project"
	}
	return n
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
	kept := f.Projects[:0]
	for _, project := range f.Projects {
		if _, err := os.Stat(project.Root); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		kept = append(kept, project)
	}
	if len(kept) != len(f.Projects) {
		if err := save(kept); err != nil {
			return nil, err
		}
	}
	return kept, nil
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
	for k := range ps {
		ps[k].Collision = false
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
