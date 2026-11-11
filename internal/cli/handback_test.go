package cli

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

func handbackIssue(t *testing.T, kind string) (*harness, string) {
	t.Helper()
	h := newHarness(t)
	id := h.newIssue("--title", "Handback task", "--kind", kind, "--body", "Do the task.")
	h.ok("define", id)
	h.enrich(id)
	h.ok("ready", id)
	return h, id
}

func writeHandback(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertReportMoved(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("report not consumed: %v", err)
	}
	ext := filepath.Ext(path)
	if _, err := os.Stat(strings.TrimSuffix(path, ext) + ".done" + ext); err != nil {
		t.Fatal(err)
	}
}

func TestHandbackImplementer(t *testing.T) {
	h, id := handbackIssue(t, "code")
	path := mdstore.Open(h.dir).HandoffPath(id, mdstore.HandoffReport)
	writeHandback(t, path, "## Summary\nBuilt the exporter.\n## Decisions\n### Stream\nScope: internal/**, cmd/**\nMemory stays flat. Buffers would grow too large.\n## Findings\n### Library\nQuoting works. Tested embedded newlines.\n## Problems\nSlow benchmark remains.")
	out := h.ok("handback", id)
	for _, want := range []string{"decision /decisions/", "finding /findings/", "log " + id, "claim " + id} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q: %s", want, out)
		}
	}
	h.expectState(id, "in-progress", false)
	if got := h.read(id, "acceptance.md"); strings.Contains(got, "[x]") {
		t.Fatalf("implementer checked criteria: %s", got)
	}
	for _, want := range []string{"Implemented, ready for check: Built the exporter.", "Problems: Slow benchmark remains."} {
		if !strings.Contains(h.read(id, "history.md"), want) {
			t.Fatalf("missing log %q", want)
		}
	}
	var listed struct {
		Entries []domain.Entry
	}
	h.jsonOf(&listed, "knowledge", "list")
	var records []domain.Entry
	for _, e := range listed.Entries {
		if e.Issue == id {
			records = append(records, e)
		}
	}
	if len(records) != 2 {
		t.Fatalf("linked records: %+v", records)
	}
	for _, e := range records {
		if e.Type == "decision" && (e.Title != "Stream" || e.Description != "Memory stays flat." || !reflect.DeepEqual(e.Scope, []string{"internal/**", "cmd/**"})) {
			t.Fatalf("decision: %+v", e)
		}
		if e.Type == "decision" && !strings.Contains(h.ok("knowledge", "show", e.Path), "Buffers would grow too large.") {
			t.Fatal("decision body missing")
		}
	}
	assertReportMoved(t, path)
	h.fails("--report", "handback", id)
	h.ok("check")
	h.ok("fmt", "--check")
}

func TestHandbackCheckerPass(t *testing.T) {
	h, id := handbackIssue(t, "code")
	path := filepath.Join(t.TempDir(), "external.md")
	writeHandback(t, path, "# Report\n## sUMMARY\nChecked locally.\n## CRITERIA\n\n1) **MET**: go test passed\n## Extra\nIgnored.\n## Problems\nNone.")
	var result writeResult
	h.jsonOf(&result, "handback", id, "--check", "--report", path, "--commit", "abc1234")
	h.expectState(id, "done", false)
	if !strings.Contains(h.read(id, "acceptance.md"), "[x]") || !strings.Contains(h.read(id, "resolution.md"), "abc1234") {
		t.Fatal("missing checked criterion or commit evidence")
	}
	if len(result.Files) == 0 || !strings.HasSuffix(result.Files[len(result.Files)-1], "resolution.md") {
		t.Fatalf("resolution not written last: %v", result.Files)
	}
	assertReportMoved(t, path)
	h.ok("check")
	h.ok("fmt", "--check")
}

func TestHandbackCheckerFail(t *testing.T) {
	for _, tc := range []struct{ name, verdict, problems, want string }{
		{"unmet", "not met — missing header", "", "criterion 1 not met: missing header"},
		{"problem", "met — test passed", "benchmark failed", "benchmark failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, id := handbackIssue(t, "code")
			h.ok("criterion", id, "--check", "1")
			path := mdstore.Open(h.dir).HandoffPath(id, mdstore.HandoffCheckReport)
			writeHandback(t, path, "## Summary\nChecked.\n## Criteria\n1. "+tc.verdict+"\n## Problems\n"+tc.problems)
			h.ok("handback", id, "--check")
			h.expectState(id, "in-progress", false)
			history := h.read(id, "history.md")
			if !strings.Contains(history, "Check failed: "+tc.want) {
				t.Fatalf("failure not logged: %s", history)
			}
			if strings.Contains(h.read(id, "acceptance.md"), "[x]") != (tc.name == "problem") {
				t.Fatal("wrong criterion status")
			}
			assertReportMoved(t, path)
			h.ok("check")
		})
	}
}

func TestHandbackResearch(t *testing.T) {
	for _, tc := range []struct{ name, verdict, problems, state, failure string }{
		{"complete", "met — reproduced quoting test", "", "done", ""},
		{"unmet", "not met — newline handling unverified", "", "in-progress", "criterion 1 not met: newline handling unverified"},
		{"problem", "met — reproduced quoting test", "Source unavailable", "in-progress", "Source unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, id := handbackIssue(t, "research")
			brief := h.ok("brief", id)
			if !strings.Contains(brief, "## Criteria\n") || !strings.Contains(brief, "except ## Summary and ## Criteria") {
				t.Fatalf("research brief lacks required Criteria: %s", brief)
			}
			path := mdstore.Open(h.dir).HandoffPath(id, mdstore.HandoffReport)
			writeHandback(t, path, "## Summary\nInvestigated quoting.\n## Criteria\n1. "+tc.verdict+"\n## Findings\n### Result\nThe library handles quotes. Tested newlines too.\n## Problems\n"+tc.problems)
			var result writeResult
			h.jsonOf(&result, "handback", id)
			h.expectState(id, tc.state, false)
			if strings.Contains(h.read(id, "acceptance.md"), "[x]") != (tc.name != "unmet") {
				t.Fatal("wrong research criterion status")
			}
			if tc.failure != "" && !strings.Contains(h.read(id, "history.md"), "Check failed: "+tc.failure) {
				t.Fatal("research failure not logged")
			}
			var listed struct{ Entries []domain.Entry }
			h.jsonOf(&listed, "knowledge", "list")
			found := false
			for _, e := range listed.Entries {
				if e.Issue == id && e.Type == "finding" && e.Title == "Result" {
					found = true
				}
			}
			if !found {
				t.Fatal("research finding not written")
			}
			if tc.state == "done" && (len(result.Files) == 0 || !strings.HasSuffix(result.Files[len(result.Files)-1], "resolution.md")) {
				t.Fatalf("resolution not last: %v", result.Files)
			}
			assertReportMoved(t, path)
			h.ok("check")
			h.ok("fmt", "--check")
		})
	}
}

// Snapshot only durable records: local reports and lock files are not records.
func handbackRecords(t *testing.T, h *harness) map[string]string {
	t.Helper()
	files := map[string]string{}
	for _, dir := range []string{"issues", "knowledge"} {
		root := filepath.Join(h.dir, ".prep", dir)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			b, err := os.ReadFile(path)
			files[path] = string(b)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

func TestHandbackRefusalWritesNothing(t *testing.T) {
	for _, tc := range []struct {
		name, text, want string
		check            bool
	}{
		{"unknown criterion", "## Summary\nChecked.\n## Criteria\n2. met", "no criterion 2", true},
		{"missing verdict", "## Summary\nChecked.\n## Criteria", "missing verdict", true},
		{"missing summary", "## Problems\nStopped.", "missing Summary", false},
		{"no reason", "## Summary\nBuilt.\n## Findings\n### Valid\nUseful finding.\n## Decisions\n### Invalid", "has no reason", false},
		{"implementer verdict", "## Summary\nBuilt.\n## Criteria\n1. met", "implementer report cannot", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, id := handbackIssue(t, "code")
			path := filepath.Join(t.TempDir(), "report.md")
			writeHandback(t, path, tc.text)
			before := handbackRecords(t, h)
			args := []string{"handback", id, "--report", path}
			if tc.check {
				args = append(args, "--check")
			}
			h.fails(tc.want, args...)
			if !reflect.DeepEqual(before, handbackRecords(t, h)) {
				t.Fatal("refused report changed records")
			}
			if _, err := os.Stat(path); err != nil {
				t.Fatal("refused report consumed", err)
			}
		})
	}
}

func TestHandbackCompletionGateWritesNothing(t *testing.T) {
	h, id := handbackIssue(t, "research")
	path := filepath.Join(t.TempDir(), "check.md")
	writeHandback(t, path, "## Summary\nChecked.\n## Criteria\n1. met — test")
	before := handbackRecords(t, h)
	h.fails("G_FINDINGS", "handback", id, "--check", "--report", path, "--json")
	if !reflect.DeepEqual(before, handbackRecords(t, h)) {
		t.Fatal("failed completion gate wrote partial records")
	}
}

func TestHandbackStdinAndMissingDefault(t *testing.T) {
	h, id := handbackIssue(t, "code")
	path := mdstore.Open(h.dir).HandoffPath(id, mdstore.HandoffReport)
	code, out := h.run("handback", id)
	if code == 0 || !strings.Contains(out, path) || !strings.Contains(out, "--report") {
		t.Fatalf("missing default error: %s", out)
	}
	var stdout, stderr bytes.Buffer
	code = Main([]string{"--root", h.dir, "handback", id, "--report", "-"}, strings.NewReader("## Summary\nBuilt from sandbox."), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("stdin: %s %s", &stdout, &stderr)
	}
	h.expectState(id, "in-progress", false)
}
