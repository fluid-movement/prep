package domain

import (
	"strings"
	"testing"
)

func TestSections(t *testing.T) {
	body := "Intro line.\n\n## Setup\n\nInstall it.\n\n### Flags\n\n- `--x`\n\n```sh\n# not a heading\n```\n\n## Setup\n\nAgain.\n\n## Use it\n\nRun it."
	got := Sections(body)
	var anchors []string
	for _, s := range got {
		anchors = append(anchors, s.Anchor)
	}
	if strings.Join(anchors, ",") != ",setup,flags,setup-1,use-it" {
		t.Fatalf("anchors = %v", anchors)
	}
	if got[0].Text != "Intro line." {
		t.Errorf("intro = %q", got[0].Text)
	}
	if !strings.Contains(got[1].Text, "### Flags") || !strings.Contains(got[1].Text, "# not a heading") || strings.Contains(got[1].Text, "Again.") {
		t.Errorf("setup takes its subsections up to the next ## heading: %q", got[1].Text)
	}
	if strings.Contains(got[1].Own, "Flags") {
		t.Errorf("own text stops at the next heading: %q", got[1].Own)
	}
	if Slug("Write commands: `new` & edit!") != "write-commands-new-edit" {
		t.Errorf("slug = %q", Slug("Write commands: `new` & edit!"))
	}
}

func TestBlocks(t *testing.T) {
	text := "# Title\n\nApplies when\nwrapped.\n\n- one\n  - nested\n- two\n1. three\n\n```\na\n\nb\n```"
	got := Blocks(text)
	want := []string{"Applies when\nwrapped.", "- one\n  - nested", "- two", "1. three", "```\na\n\nb\n```"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("blocks = %q", got)
	}
}

func TestFindKnowledge(t *testing.T) {
	entries := []*Entry{
		{Path: "/components/tui.md", Title: "TUI", Description: "Screens and keys.", Scope: []string{"internal/tui"},
			Body: "# TUI\n\n- **Keys**: letter sequences select actions.\n- **Wheel**: scrolls the list view, never the selection.\n- Mouse clicks select rows."},
		{Path: "/components/cli.md", Title: "CLI", Description: "Commands.",
			Body: "# CLI\n\n- Writes stage their files with git.\n\n## Wheels\n\nNothing about scrolling here."},
	}
	tree := NewTree(Project{}, nil, entries, nil)

	hits := tree.FindKnowledge("wheel scrolls", 0)
	if len(hits) == 0 || !strings.HasPrefix(hits[0].Text, "- **Wheel**") {
		t.Fatalf("best hit = %+v", hits)
	}
	if hits[0].Matched != 2 || hits[0].Path != "/components/tui.md" {
		t.Errorf("hit = %+v", hits[0])
	}
	// A typo still finds the block; a heading match counts for the blocks under it.
	if hits := tree.FindKnowledge("selecton", 1); len(hits) != 1 || !strings.Contains(hits[0].Text, "selection") {
		t.Errorf("typo: %+v", hits)
	}
	if hits := tree.FindKnowledge("wheels scrolling", 1); len(hits) != 1 || hits[0].Anchor != "wheels" || hits[0].Matched != 2 {
		t.Errorf("heading: %+v", hits)
	}
	// Description matches only rank: a block must match a term itself.
	if hits := tree.FindKnowledge("commands git", 0); len(hits) != 1 || !strings.Contains(hits[0].Text, "git") {
		t.Errorf("description: %+v", hits)
	}
	if hits := tree.FindKnowledge("zzzz", 0); len(hits) != 0 {
		t.Errorf("no match: %+v", hits)
	}
	if !oneEdit("selection", "selecton") || !oneEdit("wheel", "whele") || oneEdit("wheel", "wheel") || oneEdit("abc", "xyz") {
		t.Error("oneEdit")
	}
}

func TestSectionsAndBlocksTrackFenceType(t *testing.T) {
	body := "# Top\n\n````md\n```go\n## not a heading\n````\n\n## Real\n\ntext"
	secs := Sections(body)
	if len(secs) != 2 || secs[1].Heading != "Real" {
		t.Fatalf("sections = %+v", secs)
	}
	bs := Blocks("```\n~~~\n\n## inside\n```\n\nafter")
	if len(bs) != 2 || bs[1] != "after" {
		t.Fatalf("blocks = %q", bs)
	}
}
