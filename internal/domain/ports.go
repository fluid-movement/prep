package domain

import "fmt"

// IssueStore is the port for transactional issue records. The markdown
// adapter is the first implementation.
type IssueStore interface {
	// Load reads the project and all issues, with format-level diagnostics.
	Load() (Project, []*Issue, []Diagnostic, error)
	// Apply writes a change as a recoverable storage transaction and returns
	// the storage locations it touched. Implementations may expose an
	// interrupted transaction only by returning an error until recovery.
	Apply(*Change) ([]string, error)
}

// KnowledgeStore is the port for the knowledge base, built around retrieval.
// The OKF adapter is the first implementation.
type KnowledgeStore interface {
	// Load reads all entries.
	Load() ([]*Entry, []Diagnostic, error)
	// Render builds the entry a knowledge write would produce, without
	// writing, so the domain can validate it.
	Render(*KnowledgeRecord) (*Entry, error)
	// Apply writes a rendered knowledge write and returns the paths touched.
	Apply(*KnowledgeRecord) ([]string, error)
}

// Load builds a tree from both stores. A project written by a newer schema
// is refused: this binary cannot read it.
func Load(is IssueStore, ks KnowledgeStore) (*Tree, error) {
	p, issues, diags, err := is.Load()
	if err != nil {
		return nil, err
	}
	if p.Schema > SchemaVersion {
		return nil, &Error{Code: ErrSchema, Message: fmt.Sprintf("project uses schema %d; this prep supports %d — update prep", p.Schema, SchemaVersion)}
	}
	entries, kd, err := ks.Load()
	if err != nil {
		return nil, err
	}
	return NewTree(p, issues, entries, append(diags, kd...)), nil
}
