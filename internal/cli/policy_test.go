package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/mdstore"
)

func TestProjectPolicyCommand(t *testing.T) {
	h := newHarness(t)
	// Use a fresh project, not the helper's explicitly configured project.
	h.dir = t.TempDir()
	h.ok("init", "--no-bootstrap")
	if got := h.ok("policy", "show", "--json"); !strings.Contains(got, `"configured": false`) {
		t.Fatal(got)
	}
	id := h.newIssue("--title", "Small task", "--kind", "manual", "--body", "Verify the behavior.")
	h.ok("define", id)
	h.ok("criterion", id, "--add", "Behavior verified")
	h.fails("completion policy", "ready", id)
	h.fails("documentation", "policy", "set", "--verification", "Manual observation")
	h.ok("policy", "set", "--verification", "Manual observation", "--documentation", "Explain impact", "--by", "agent/test")
	h.ok("ready", id)
	h.ok("fmt", "--check")
	p, err := mdstore.ReadProject(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Policy == nil || p.Policy.By != "agent/test" || p.Policy.Verification != "Manual observation" {
		t.Fatalf("policy %+v", p.Policy)
	}
	if len(p.DoD) != 1 || p.DoD[0] != "prep check passes" {
		t.Fatalf("lost inherited DoD: %v", p.DoD)
	}
	// Identity rewrites and canonicalization must retain the policy.
	h.ok("init", "--new-id")
	p2, err := mdstore.ReadProject(h.dir)
	if err != nil || p2.Policy == nil || *p.Policy != *p2.Policy {
		t.Fatalf("identity rewrite: %+v %v", p2, err)
	}
	before, err := os.ReadFile(filepath.Join(h.dir, ".prep", "project.md"))
	if err != nil {
		t.Fatal(err)
	}
	h.ok("policy", "show")
	after, _ := os.ReadFile(filepath.Join(h.dir, ".prep", "project.md"))
	if string(before) != string(after) {
		t.Fatal("read mutated records")
	}
}
