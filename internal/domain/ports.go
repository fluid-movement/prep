package domain

// IssueStore is the port for transactional issue records. The markdown
// adapter is the first implementation.
type IssueStore interface {
	// Load reads the project and all issues, with format-level diagnostics.
	Load() (Project, []*Issue, []Diagnostic, error)
	// Apply writes the records of a change atomically per file and returns
	// the storage locations it touched.
	Apply(*Change) ([]string, error)
}

// KnowledgeStore is the port for the knowledge base, built around retrieval.
// The OKF adapter is the first implementation.
type KnowledgeStore interface {
	// Load reads all entries; with drift it also computes code drift.
	Load(drift bool) ([]*Entry, []Diagnostic, error)
}

// Load builds a tree from both stores.
func Load(is IssueStore, ks KnowledgeStore, drift bool) (*Tree, error) {
	p, issues, diags, err := is.Load()
	if err != nil {
		return nil, err
	}
	entries, kd, err := ks.Load(drift)
	if err != nil {
		return nil, err
	}
	return NewTree(p, issues, entries, append(diags, kd...)), nil
}
