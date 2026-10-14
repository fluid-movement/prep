package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestScopedEvidenceBlocksMissingFailedAndChangedReports(t *testing.T) {
	h := newHarness(t)
	id := h.newIssue("--title", "Scoped work", "--kind", "manual", "--body", "Inspect work.")
	h.ok("define", id)
	h.ok("criterion", id, "--add", "Work observed")
	h.ok("ready", id)
	h.ok("claim", id)
	h.ok("criterion", id, "--check", "1")
	work := filepath.Join(h.dir, "work.txt")
	config := filepath.Join(h.dir, "checks.txt")
	for _, p := range []string{work, config} {
		if err := os.WriteFile(p, []byte("original"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var policy struct {
		Bound *domain.CompletionContract `json:"bound"`
	}
	h.jsonOf(&policy, "policy", "show", id)
	requirement := policy.Bound.Requirements[0].ID
	record := func(result string) {
		h.ok("evidence", "record", id, "--requirement", requirement, "--result", result,
			"--method", "automated", "--command", "test -f work.txt", "--observation", "Recorded run result.",
			"--scope", "work.txt", "--config-scope", "checks.txt")
	}
	h.fails("G_EVIDENCE", "complete", id, "--no-impact", "No impact", "--json")
	h.fails("command", "evidence", "record", id, "--requirement", requirement,
		"--result", "pass", "--method", "automated", "--observation", "Missing command", "--scope", "work.txt")
	h.fails("unknown requirement", "evidence", "record", id, "--requirement", "unknown",
		"--result", "pass", "--method", "manual", "--observation", "Wrong check", "--scope", "work.txt")
	// A command is recorded, never executed by prep.
	marker := filepath.Join(h.dir, "should-not-exist")
	h.ok("evidence", "record", id, "--requirement", requirement, "--result", "pass",
		"--method", "automated", "--command", "touch "+marker, "--observation", "Agent report.", "--scope", "work.txt")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("prep executed the reported command")
	}
	for _, r := range policy.Bound.Requirements[1:] {
		h.ok("evidence", "record", id, "--requirement", r.ID, "--result", "pass", "--method", "manual",
			"--observation", "Observed requirement.", "--scope", "work.txt", "--config-scope", "checks.txt")
	}
	record("fail")
	h.fails("G_EVIDENCE", "complete", id, "--no-impact", "No impact", "--json")
	record("pass")
	if got := h.ok("guide", id); !strings.Contains(got, "prep evidence record") {
		t.Fatal(got)
	}
	if err := os.WriteFile(config, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.fails("G_EVIDENCE", "complete", id, "--no-impact", "No impact", "--json")
	if got := h.ok("evidence", "show", id, "--json"); !strings.Contains(got, `"applicable": false`) {
		t.Fatal(got)
	}
	if err := os.WriteFile(config, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(work, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.fails("G_EVIDENCE", "complete", id, "--no-impact", "No impact", "--json")
	if err := os.WriteFile(work, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.ok("fmt", "--check")
	before := h.read(id, "verification.md")
	h.ok("evidence", "show", id)
	if before != h.read(id, "verification.md") {
		t.Fatal("read rewrote reports")
	}
	h.ok("complete", id, "--no-impact", "No impact", "--json")
	var resolved struct {
		Issue struct {
			Resolution *domain.Resolution `json:"resolution"`
		} `json:"issue"`
	}
	h.jsonOf(&resolved, "show", id)
	if resolved.Issue.Resolution == nil || resolved.Issue.Resolution.Contract == nil || len(resolved.Issue.Resolution.Verification) != len(policy.Bound.Requirements) {
		t.Fatal("resolution omitted completion evidence")
	}
	h.fails("unresolved issue", "evidence", "record", id, "--requirement", requirement,
		"--result", "pass", "--method", "manual", "--observation", "Late report", "--scope", "work.txt")
}
