package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/fluid-movement/prep/internal/domain"
)

func TestKnowledgeRecords(t *testing.T) {
	p, ids := knowledgeSample(t)
	m := withCheck(openModel(t, p, 160, 48), p)
	run(m, "b")
	oldPath := "/decisions/" + ids["oldDecision"] + ".md"
	newPath := "/decisions/" + ids["newDecision"] + ".md"
	paths, _ := m.knowledgePaths()
	if paths[0] != newPath || slices.Contains(paths, oldPath) {
		t.Fatalf("default records: %v", paths)
	}
	golden(t, "knowledge-records", m.View().Content)
	run(m, "u")
	paths, _ = m.knowledgePaths()
	if !slices.Contains(paths, oldPath) {
		t.Fatal("toggle did not show superseded record")
	}
	golden(t, "knowledge-superseded", m.View().Content)
	run(m, "o")
	run(m, "2")
	if m.shownEntry() != oldPath {
		t.Fatalf("supersedes link opened %s", m.shownEntry())
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "Superseded by:") {
		t.Fatalf("missing reverse link: %s", v)
	}
	golden(t, "knowledge-superseded-entry", m.View().Content)
	run(m, "o")
	run(m, "2")
	if m.shownEntry() != newPath {
		t.Fatal("reverse link did not return to newer record")
	}
	run(m, "o")
	run(m, "1")
	if m.screen != screenIssues || m.selected() != ids["doc"] {
		t.Fatal("record issue link failed")
	}
	golden(t, "detail-decisions", m.View().Content)
	run(m, "o")
	for n, k := range linkTargets(m.relations(m.selected())) {
		if m.relations(m.selected())[k].id == oldPath {
			run(m, linkKeys[n:n+1])
			break
		}
	}
	if m.screen != screenKnowledge || m.shownEntry() != oldPath {
		t.Fatal("issue decision link failed")
	}
	run(m, "u")
	if paths, _ := m.knowledgePaths(); slices.Contains(paths, oldPath) {
		t.Fatal("toggle did not hide superseded record")
	}
}

func TestKnowledgeRecordOrder(t *testing.T) {
	m := &Model{tree: &domain.Tree{Knowledge: map[string]*domain.Entry{
		"/a.md":           {Title: "Description"},
		"/decisions/z.md": {ID: "z", Date: "2026-01-01"},
		"/findings/a.md":  {ID: "a", Date: "2026-02-01"},
		"/decisions/b.md": {ID: "b", Date: "2026-02-01"},
	}}}
	got, err := m.knowledgePaths()
	want := []string{"/decisions/b.md", "/findings/a.md", "/decisions/z.md", "/a.md"}
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("paths = %v, %v; want %v", got, err, want)
	}
}
