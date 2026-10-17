package okf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fluid-movement/prep/internal/domain"
	"gopkg.in/yaml.v3"
)

var renderMu sync.Mutex
var renderCache = map[*domain.KnowledgeRecord]string{}

func (s *Store) file(p string) string {
	return filepath.Join(s.Root, filepath.FromSlash(Dir), filepath.FromSlash(strings.TrimPrefix(p, "/")))
}

// Render validates a new immutable record without changing the bundle.
func (s *Store) Render(r *domain.KnowledgeRecord) (*domain.Entry, error) {
	expected := "/" + r.Type + "s/" + r.ID + ".md"
	if !domain.ValidID(r.ID) || (r.Type != "decision" && r.Type != "finding") || r.Path != expected {
		return nil, &domain.Error{Code: domain.ErrUsage, Message: "invalid record path or type"}
	}
	if strings.TrimSpace(r.Title) == "" || strings.TrimSpace(r.Description) == "" || strings.ContainsAny(r.Description, "\r\n") {
		return nil, &domain.Error{Code: domain.ErrUsage, Message: "record needs a title and one-line description"}
	}
	if _, err := os.Lstat(s.file(r.Path)); err == nil {
		return nil, &domain.Error{Code: domain.ErrConflict, Message: "record already exists: " + r.ID}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	m := &yaml.Node{Kind: yaml.MappingNode}
	set(m, "type", r.Type)
	set(m, "title", r.Title)
	set(m, "description", r.Description)
	set(m, "date", r.At.UTC().Format("2006-01-02"))
	if r.Issue != "" {
		set(m, "issue", r.Issue)
	}
	if len(r.Scope) > 0 {
		setSeq(m, "scope", r.Scope)
	}
	if len(r.Supersedes) > 0 {
		setSeq(m, "supersedes", r.Supersedes)
	}
	if r.Outcome {
		m.Content = append(m.Content, scalar("outcome"), &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"})
	}
	content, err := compose(m, r.Body)
	if err != nil {
		return nil, err
	}
	e, err := parseEntry(r.Path, content)
	if err != nil {
		return nil, err
	}
	renderMu.Lock()
	renderCache[r] = content
	renderMu.Unlock()
	return e, nil
}

// Apply atomically installs only a new file. Link refuses even a concurrent
// creation at the destination, so no existing record can be overwritten.
func (s *Store) Apply(r *domain.KnowledgeRecord) ([]string, error) {
	renderMu.Lock()
	content, ok := renderCache[r]
	delete(renderCache, r)
	renderMu.Unlock()
	if !ok {
		if _, err := s.Render(r); err != nil {
			return nil, err
		}
		return s.Apply(r)
	}
	p := s.file(r.Path)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), ".prep-tmp-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.WriteString(content); err != nil {
		tmp.Close()
		return nil, err
	}
	if err = tmp.Chmod(0644); err != nil {
		tmp.Close()
		return nil, err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return nil, err
	}
	if err = tmp.Close(); err != nil {
		return nil, err
	}
	if err = os.Link(tmp.Name(), p); err != nil {
		if os.IsExist(err) {
			return nil, &domain.Error{Code: domain.ErrConflict, Message: fmt.Sprintf("record already exists: %s", r.ID)}
		}
		return nil, err
	}
	idx, err := s.WriteIndexes(false)
	return append([]string{Dir + r.Path}, idx...), err
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
func scalar(v string) *yaml.Node        { return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v} }
func set(m *yaml.Node, key, val string) { m.Content = append(m.Content, scalar(key), scalar(val)) }
func setSeq(m *yaml.Node, key string, vals []string) {
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, v := range vals {
		seq.Content = append(seq.Content, scalar(v))
	}
	m.Content = append(m.Content, scalar(key), seq)
}
