package eval

import (
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/fluid-movement/prep/internal/eval/assert/outcome"
	_ "github.com/fluid-movement/prep/internal/eval/assert/tools"
	"github.com/fluid-movement/prep/internal/eval/evalcase"
	"github.com/fluid-movement/prep/internal/eval/fixture"
)

// TestShippedCases loads every case under eval/ (which validates them) and
// builds each distinct fixture once, checking every case's aliases expand
// against the built fixture and its assertions decode.
func TestShippedCases(t *testing.T) {
	cases, err := evalcase.Load("../../eval")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) < 20 {
		t.Fatalf("only %d cases", len(cases))
	}
	bin := filepath.Join(t.TempDir(), "prep")
	if out, err := exec.Command("go", "build", "-o", bin, "../../cmd/prep").CombinedOutput(); err != nil {
		t.Fatalf("build prep: %v\n%s", err, out)
	}
	built := map[string]*fixture.Workspace{}
	areas := map[string]int{}
	smoke := 0
	for _, c := range cases {
		areas[c.Area]++
		for _, tag := range c.Tags {
			if tag == "smoke" {
				smoke++
			}
		}
		key := c.Fixture + "|" + c.Workspace
		w, ok := built[key]
		if !ok {
			w, err = fixture.BuildWorkspace(c.Resolved(), t.TempDir(), fixture.Options{Binary: bin})
			if err != nil {
				t.Fatalf("%s: build fixture %q: %v", c.Name, key, err)
			}
			built[key] = w
		}
		if _, err := c.Compile(w.PrimaryProject().Expand); err != nil {
			t.Errorf("%s: %v", c.Name, err)
		}
	}
	for _, a := range []string{"session", "define", "enrich", "implement", "issues", "knowledge-read", "knowledge-write", "links", "queries", "lifecycle-other"} {
		if areas[a] == 0 {
			t.Errorf("no cases in area %s", a)
		}
	}
	if smoke < 4 || smoke > 8 {
		t.Errorf("%d smoke cases", smoke)
	}
}
