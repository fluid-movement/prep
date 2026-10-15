package cli

import (
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestArtifactReferenceAddShowFmtEditRemove(t *testing.T) {
	h := newHarness(t)
	id := h.newIssue("--title", "External item", "--kind", "manual", "--body", "Review the upstream item.")
	if raw := h.read(id, "issue.md"); strings.Contains(raw, "artifacts:") {
		t.Fatalf("legacy issue without references gained an empty artifacts field:\n%s", raw)
	}
	initialState, initialStale := h.state(id)
	const source = "tracker:team/project"
	const location = "https://tracker.example/items/42"
	const title = "Upstream decision"
	h.ok("artifact", "add", id, "--source", source, "--location", location, "--title", title)

	var shown struct {
		Issue struct {
			Title     string                     `json:"title"`
			Artifacts []domain.ArtifactReference `json:"artifacts"`
		} `json:"issue"`
		State string `json:"state"`
		Stale bool   `json:"stale"`
	}
	h.jsonOf(&shown, "show", id)
	want := domain.ArtifactReference{Source: source, Location: location, Title: title}
	if len(shown.Issue.Artifacts) != 1 || shown.Issue.Artifacts[0] != want {
		t.Fatalf("JSON references = %#v, want %#v", shown.Issue.Artifacts, want)
	}
	if shown.State != initialState || shown.Stale != initialStale {
		t.Fatalf("add changed lifecycle state: %s stale=%v; want %s stale=%v", shown.State, shown.Stale, initialState, initialStale)
	}
	if out := h.ok("show", id); !strings.Contains(out, "artifact (reference only): "+source+" — "+location+" ("+title+")") {
		t.Fatalf("text show omitted reference: %s", out)
	}

	h.fails("already exists", "artifact", "add", id, "--source", source, "--location", location)
	h.fails("control characters", "artifact", "add", id, "--source", "tracker\nteam", "--location", "/items/43")
	h.fails("control characters", "artifact", "add", id, "--source", source, "--location", location, "--title", "bad\ntitle")
	if out := h.ok("fmt"); strings.Contains(out, "invalid artifact") {
		t.Fatalf("fmt rejected a valid reference: %s", out)
	}

	h.ok("edit", id, "--title", "Edited title")
	var edited struct {
		Issue struct {
			Title     string                     `json:"title"`
			Artifacts []domain.ArtifactReference `json:"artifacts"`
		} `json:"issue"`
		State string `json:"state"`
		Stale bool   `json:"stale"`
	}
	h.jsonOf(&edited, "show", id)
	if edited.Issue.Title != "Edited title" || len(edited.Issue.Artifacts) != 1 || edited.Issue.Artifacts[0] != want {
		t.Fatalf("ordinary edit did not preserve artifact metadata: %+v", edited)
	}
	if edited.State != initialState || edited.Stale != initialStale {
		t.Fatalf("metadata/edit changed lifecycle state: %s stale=%v; want %s stale=%v", edited.State, edited.Stale, initialState, initialStale)
	}

	h.ok("artifact", "remove", id, "--source", source, "--location", location)
	var removed struct {
		Issue struct {
			Artifacts []domain.ArtifactReference `json:"artifacts"`
		} `json:"issue"`
		State string `json:"state"`
		Stale bool   `json:"stale"`
	}
	h.jsonOf(&removed, "show", id)
	if len(removed.Issue.Artifacts) != 0 {
		t.Fatalf("remove left references: %#v", removed.Issue.Artifacts)
	}
	if removed.State != initialState || removed.Stale != initialStale {
		t.Fatalf("remove changed lifecycle state: %s stale=%v; want %s stale=%v", removed.State, removed.Stale, initialState, initialStale)
	}
	if out := h.ok("show", id); strings.Contains(out, "artifact (reference only):") {
		t.Fatalf("text show retained removed reference: %s", out)
	}
	h.fails("reference not found", "artifact", "remove", id, "--source", source, "--location", location)
}
