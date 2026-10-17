package eval

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/fluid-movement/prep/internal/eval/assert"
	_ "github.com/fluid-movement/prep/internal/eval/assert/outcome"
	_ "github.com/fluid-movement/prep/internal/eval/assert/tools"
	"github.com/fluid-movement/prep/internal/eval/evalcase"
	"github.com/fluid-movement/prep/internal/eval/fixture"
	"github.com/fluid-movement/prep/internal/eval/transcript"
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
		if testing.Short() && c.Fixture == "large" {
			continue // sixty issues take a minute to build
		}
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
	for _, a := range []string{"session", "define", "enrich", "implement", "issues", "knowledge-read", "knowledge-write", "links", "lifecycle-other", "batch", "real", "natural", "adversarial", "scale"} {
		if areas[a] == 0 {
			t.Errorf("no cases in area %s", a)
		}
	}
	if smoke < 4 || smoke > 8 {
		t.Errorf("%d smoke cases", smoke)
	}
}

func TestRespectDecisionTools(t *testing.T) {
	cases, err := evalcase.Load("../../eval")
	if err != nil {
		t.Fatal(err)
	}
	var c *evalcase.Compiled
	for _, candidate := range cases {
		if candidate.Name == "knowledge-read-respect-decision" {
			c, err = candidate.Compile(func(s string) (string, error) { return s, nil })
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if c == nil {
		t.Fatal("missing respect-decision case")
	}
	decision := transcript.Step{Tool: "Bash", Shell: "prep decide --title Cache --description Reason --supersedes 01OLD"}
	decision.Commands = transcript.ParseShell(decision.Shell)
	for _, change := range []transcript.Step{
		{Tool: "Edit", Path: "/work/src/cart.go"},
		{Tool: "Write", Path: "/work/src/cart.go"},
		{Tool: "Bash", Shell: "cat > src/cart.go <<'EOF'\npackage shop\nEOF"},
		{Tool: "Bash", Shell: "python -c 'open(\"src/cart.go\", \"w\").write(code)'"},
	} {
		for _, tc := range []struct {
			name  string
			steps []transcript.Step
			want  bool
		}{
			{"no supersede", []transcript.Step{change}, false},
			{"supersede first", []transcript.Step{decision, change}, true},
			{"supersede too late", []transcript.Step{change, decision}, false},
		} {
			t.Run(change.Tool+"/"+tc.name, func(t *testing.T) {
				tr := &transcript.Transcript{Steps: append([]transcript.Step(nil), tc.steps...)}
				for i := range tr.Steps {
					tr.Steps[i].Index = i
				}
				passed := true
				for _, a := range c.Tools {
					if result := a.Check(&assert.Input{Transcript: tr}); !result.Pass {
						passed = false
					}
				}
				if passed != tc.want {
					t.Errorf("tools pass = %v, want %v", passed, tc.want)
				}
			})
		}
	}
	tr := &transcript.Transcript{Steps: []transcript.Step{{Tool: "Read", Path: "/work/src/cart.go"}, {Index: 1, Tool: "Bash", Shell: "cat src/cart.go"}}}
	for _, a := range c.Tools {
		if result := a.Check(&assert.Input{Transcript: tr}); !result.Pass {
			t.Errorf("unchanged code failed: %+v", result)
		}
	}
}
