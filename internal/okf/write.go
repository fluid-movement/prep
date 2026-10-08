package okf

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/domain"
)

// rendered remembers what Render produced and the file it read, so Apply
// writes exactly what was validated and refuses a file changed meanwhile.
type rendered struct {
	content string
	before  [32]byte
	existed bool
}

var (
	renderMu    sync.Mutex
	renderCache = map[*domain.KnowledgeEdit]rendered{}
)

func (s *Store) file(p string) string {
	return filepath.Join(s.Root, filepath.FromSlash(Dir), filepath.FromSlash(strings.TrimPrefix(p, "/")))
}

// Render builds the entry file for an edit and parses it back.
func (s *Store) Render(e *domain.KnowledgeEdit) (*domain.Entry, error) {
	raw, err := os.ReadFile(s.file(e.Path))
	existed := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if e.New && existed {
		// prep knowledge new only plans entries the tree does not have, so
		// this file failed to load; writing over it would lose it.
		return nil, &domain.Error{Code: domain.ErrConflict, Message: fmt.Sprintf("%s exists but is not a readable entry (see prep check); fix or remove it first", Dir+e.Path)}
	}
	var content string
	if e.New {
		content, err = renderNew(e)
	} else {
		content, err = renderUpdate(string(raw), e)
	}
	if err != nil {
		return nil, err
	}
	entry, err := parseEntry(e.Path, content)
	if err != nil {
		return nil, err
	}
	renderMu.Lock()
	renderCache[e] = rendered{content: content, before: sha256.Sum256(raw), existed: existed}
	renderMu.Unlock()
	return entry, nil
}

// Apply writes what Render produced, atomically.
func (s *Store) Apply(e *domain.KnowledgeEdit) ([]string, error) {
	renderMu.Lock()
	r, ok := renderCache[e]
	delete(renderCache, e)
	renderMu.Unlock()
	if !ok {
		if _, err := s.Render(e); err != nil {
			return nil, err
		}
		return s.Apply(e)
	}
	p := s.file(e.Path)
	cur, err := os.ReadFile(p)
	if (err == nil) != r.existed || (err == nil && sha256.Sum256(cur) != r.before) {
		return nil, &domain.Error{Code: domain.ErrConflict, Message: fmt.Sprintf("%s changed since it was read; re-run the command", Dir+e.Path)}
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), ".prep-tmp-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(r.content); err != nil {
		tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp.Name(), p); err != nil {
		return nil, err
	}
	touched := []string{Dir + e.Path}
	idx, err := s.WriteIndexes(false)
	return append(touched, idx...), err
}

func renderNew(e *domain.KnowledgeEdit) (string, error) {
	m := &yaml.Node{Kind: yaml.MappingNode}
	set(m, "type", *e.Type)
	set(m, "title", *e.Title)
	set(m, "description", *e.Description)
	if e.Status != nil && *e.Status != "" {
		set(m, "status", *e.Status)
	}
	gen := &yaml.Node{Kind: yaml.MappingNode}
	set(gen, "by", e.Actor)
	gen.Content = append(gen.Content, scalar("at"), &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!timestamp", Value: e.At.UTC().Format(time.RFC3339)})
	m.Content = append(m.Content, scalar("generated"), gen)
	if e.Scope != nil && len(*e.Scope) > 0 {
		setSeq(m, "scope", *e.Scope)
	}
	if e.ConfirmedCommit != nil && *e.ConfirmedCommit != "" {
		set(m, "confirmed_commit", *e.ConfirmedCommit)
	}
	return compose(m, *e.Body)
}

func renderUpdate(raw string, e *domain.KnowledgeEdit) (string, error) {
	fm, body, err := splitEntry(raw)
	if err != nil {
		return "", fmt.Errorf("%s: %v", Dir+e.Path, err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(fm), &doc); err != nil {
		return "", fmt.Errorf("%s: frontmatter: %v", Dir+e.Path, err)
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return "", fmt.Errorf("%s: frontmatter is not a mapping", Dir+e.Path)
	}
	m := doc.Content[0]
	for _, f := range []struct {
		key string
		v   *string
	}{{"type", e.Type}, {"title", e.Title}, {"description", e.Description}, {"status", e.Status}, {"confirmed_commit", e.ConfirmedCommit}} {
		switch {
		case f.v == nil:
		case *f.v == "":
			remove(m, f.key)
		default:
			set(m, f.key, *f.v)
		}
	}
	if e.Scope != nil {
		if len(*e.Scope) == 0 {
			remove(m, "scope")
			remove(m, "confirmed_commit")
		} else {
			setSeq(m, "scope", *e.Scope)
		}
	}
	if e.Body != nil {
		body = *e.Body
	}
	return compose(m, body)
}

func compose(m *yaml.Node, body string) (string, error) {
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(m); err != nil {
		return "", err
	}
	enc.Close()
	return "---\n" + b.String() + "---\n\n" + strings.TrimSpace(body) + "\n", nil
}

func scalar(v string) *yaml.Node { return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v} }

// set replaces a key's value in place or appends the key.
func set(m *yaml.Node, key, val string) {
	for k := 0; k+1 < len(m.Content); k += 2 {
		if m.Content[k].Value == key {
			m.Content[k+1] = scalar(val)
			return
		}
	}
	m.Content = append(m.Content, scalar(key), scalar(val))
}

func setSeq(m *yaml.Node, key string, vals []string) {
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, v := range vals {
		seq.Content = append(seq.Content, scalar(v))
	}
	for k := 0; k+1 < len(m.Content); k += 2 {
		if m.Content[k].Value == key {
			m.Content[k+1] = seq
			return
		}
	}
	m.Content = append(m.Content, scalar(key), seq)
}

func remove(m *yaml.Node, key string) {
	for k := 0; k+1 < len(m.Content); k += 2 {
		if m.Content[k].Value == key {
			m.Content = append(m.Content[:k], m.Content[k+2:]...)
			return
		}
	}
}
