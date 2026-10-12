package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/matrix"
	"github.com/fluid-movement/prep/internal/eval/results"
	"github.com/fluid-movement/prep/internal/eval/run"
)

func evalCell(c, v string, k, f int) *matrix.CellResult {
	cr := &matrix.CellResult{Cell: matrix.Cell{Case: c, Variant: v, Model: "haiku"}}
	for i := 0; i < k+f; i++ {
		cr.Runs = append(cr.Runs, matrix.RunResult{Index: i, Status: run.Completed, Cost: matrix.Cost{USD: 0.01, Calls: 4},
			Assertions: []assert.Result{{Kind: "ran", Desc: "ran guide", Pass: i < k}}})
	}
	matrix.Reduce(cr)
	return cr
}

func TestEvalReportCompareAndPrune(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()
	st := results.Store{Root: filepath.Join(dir, "results")}
	for id, cells := range map[string][]*matrix.CellResult{
		"20261012-100000-aaaa": {evalCell("a", "base", 9, 1), evalCell("a", "mutant", 1, 9)},
		"20261013-100000-bbbb": {evalCell("a", "base", 2, 8)},
	} {
		if err := st.Save(&results.Run{ID: id, Created: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Definition: results.Definition{N: 10}, Cells: cells}); err != nil {
			t.Fatal(err)
		}
	}
	out := h.ok("eval", "report", "--dir", dir, "20261012")
	for _, want := range []string{"run 20261012-100000-aaaa", "base / haiku", "9/10 90%", "mutant / haiku"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if out := h.ok("eval", "report", "--dir", dir, "--failures"); !strings.Contains(out, "run 20261013-100000-bbbb") || !strings.Contains(out, "failed 8x: ran guide") {
		t.Errorf("latest with failures:\n%s", out)
	}
	var j struct {
		ID    string `json:"id"`
		Cells []any  `json:"cells"`
	}
	h.jsonOf(&j, "eval", "report", "--dir", dir, "20261012")
	if j.ID != "20261012-100000-aaaa" || len(j.Cells) != 2 {
		t.Errorf("json: %+v", j)
	}
	cmp := h.ok("eval", "compare", "--dir", dir, "20261012", "20261013")
	if !strings.Contains(cmp, "REGRESSED") || !strings.Contains(cmp, "a / base / haiku") || !strings.Contains(cmp, "ONLY IN A") {
		t.Errorf("compare:\n%s", cmp)
	}
	vc := h.ok("eval", "compare", "--dir", dir, "20261012", "--variants", "base,mutant")
	if !strings.Contains(vc, "REGRESSED (1)") {
		t.Errorf("variant compare:\n%s", vc)
	}
	h.fails("usage", "eval", "compare", "--dir", dir, "20261012")
	h.fails("no stored run", "eval", "report", "--dir", dir, "zzz")
	h.fails("unknown eval command", "eval", "bogus")
	if out := h.ok("eval", "prune", "--dir", dir, "--keep", "1"); !strings.Contains(out, "removed 1") {
		t.Errorf("prune:\n%s", out)
	}
	h.fails("no stored run", "eval", "report", "--dir", dir, "20261012")
}

// copyEvalDir copies the repository's eval directory (cases, fixtures,
// variants, judge samples) so a test can write results next to it.
func copyEvalDir(t *testing.T) string {
	t.Helper()
	dst := t.TempDir()
	src := "../../eval"
	err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if strings.HasPrefix(rel, "results") {
			return nil
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

const fakeStream = `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"prep next"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"01X Payment integration"}]}}
{"type":"result","subtype":"success","result":"Work on Payment integration.","num_turns":2,"total_cost_usd":0.02}`

func fakeClaudeFor(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo 9.9.9; exit 0; fi\ncat >/dev/null\n" +
		"printf '{\"type\":\"system\",\"subtype\":\"init\",\"cwd\":\"%s\",\"session_id\":\"s\",\"model\":\"claude-haiku-5-5\",\"skills\":[\"prep\"],\"plugins\":[{\"name\":\"prep\"}]}\\n' \"$PWD\"\n" +
		"cat <<'S'\n" + fakeStream + "\nS\n"
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEvalListAndDryRun(t *testing.T) {
	h := newHarness(t)
	dir := copyEvalDir(t)
	out := h.ok("eval", "list", "--dir", dir)
	for _, want := range []string{"session-what-next", "links-cross-dependency", "define-from-scratch", "#smoke"} {
		if !strings.Contains(out, want) {
			t.Errorf("list lacks %q:\n%s", want, out)
		}
	}
	if out := h.ok("eval", "list", "--dir", dir, "--smoke"); strings.Contains(out, "enrich-tax-rules") || !strings.Contains(out, "session-what-next") {
		t.Errorf("smoke list:\n%s", out)
	}
	if out := h.ok("eval", "list", "--dir", dir, "--area", "links"); strings.Count(out, "\n") != 3 {
		t.Errorf("area list:\n%s", out)
	}
	h.fails("no case matches", "eval", "list", "--dir", dir, "--case", "nonsense")
	dry := h.ok("eval", "run", "--dir", dir, "--smoke", "--models", "haiku,opus", "--n", "2", "--dry-run")
	if !strings.Contains(dry, "cell(s) x 2 run(s)") || !strings.Contains(dry, "session-what-next / current / opus") || !strings.Contains(dry, "~$0.900/run") {
		t.Errorf("dry run:\n%s", dry)
	}
	h.fails("exceeds --max-usd", "eval", "run", "--dir", dir, "--smoke", "--models", "opus", "--n", "5", "--max-usd", "1", "--dry-run=false")
}

func TestEvalRunDoctorAndJudgeCheck(t *testing.T) {
	h := newHarness(t)
	dir := copyEvalDir(t)
	claude := fakeClaudeFor(t)
	h.fails("claude setup-token", "eval", "run", "--dir", dir, "--case", "session-what-next", "--models", "haiku", "--n", "1", "--claude", claude)
	t.Setenv("ANTHROPIC_API_KEY", "k")
	plugin := "../../plugins/claude-code"
	out := h.ok("eval", "run", "--dir", dir, "--case", "session-what-next", "--models", "haiku", "--n", "2", "--no-judge", "--claude", claude, "--plugin", plugin, "--binary", prepBinary(t))
	for _, want := range []string{"session-what-next", "current / haiku", "stored as "} {
		if !strings.Contains(out, want) {
			t.Errorf("run output lacks %q:\n%s", want, out)
		}
	}
	rep := h.ok("eval", "report", "--dir", dir)
	if !strings.Contains(rep, "session-what-next") {
		t.Errorf("stored run not reported:\n%s", rep)
	}
	var j struct {
		Cells []struct {
			Runs []struct {
				Status string `json:"status"`
			} `json:"runs"`
			ModelResolved []string `json:"resolved"`
		} `json:"cells"`
	}
	h.jsonOf(&j, "eval", "report", "--dir", dir)
	if len(j.Cells) != 1 || len(j.Cells[0].Runs) != 2 || j.Cells[0].Runs[0].Status != "completed" || j.Cells[0].ModelResolved[0] != "claude-haiku-5-5" {
		t.Errorf("stored cells: %+v", j)
	}
	// doctor: the fake harness reports the variant's plugin and skill, nothing else.
	if out := h.ok("eval", "doctor", "--dir", dir, "--claude", claude, "--plugin", plugin, "--binary", prepBinary(t)); !strings.Contains(out, "ok") {
		t.Errorf("doctor:\n%s", out)
	}
	// judge-check with a fake judge that always says pass: untrusted.
	judgeClaude := filepath.Join(t.TempDir(), "claude")
	os.WriteFile(judgeClaude, []byte("#!/bin/sh\ncat >/dev/null\necho '{\"type\":\"result\",\"is_error\":false,\"result\":\"{\\\"verdict\\\":\\\"pass\\\",\\\"reason\\\":\\\"fine\\\"}\"}'\n"), 0o755)
	h.fails("untrusted", "eval", "judge-check", "--dir", dir, "--claude", judgeClaude, "--repeats", "1")
}

// prepBinary builds the prep binary once for tests that run it as the agent's.
func prepBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "prep")
	if out, err := exec.Command("go", "build", "-o", bin, "../../cmd/prep").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func TestEvalModels(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()
	st := results.Store{Root: filepath.Join(dir, "results")}
	mk := func(c, model string, k, f int, usd float64) *matrix.CellResult {
		cr := &matrix.CellResult{Cell: matrix.Cell{Case: c, Variant: "current", Model: model}}
		for i := 0; i < k+f; i++ {
			cr.Runs = append(cr.Runs, matrix.RunResult{Index: i, Status: run.Completed, ModelResolved: "id-" + model, Cost: matrix.Cost{USD: usd},
				Assertions: []assert.Result{{Kind: "ran", Desc: "ran guide", Pass: i < k}}})
		}
		matrix.Reduce(cr)
		return cr
	}
	err := st.Save(&results.Run{ID: "20261012-100000-aaaa", Created: time.Now(),
		Definition: results.Definition{N: 10, Areas: map[string]string{"c1": "queries"}},
		Cells:      []*matrix.CellResult{mk("c1", "haiku", 20, 0, 0.01), mk("c1", "sonnet", 20, 0, 0.05)}})
	if err != nil {
		t.Fatal(err)
	}
	out := h.ok("eval", "models", "--dir", dir)
	if !strings.Contains(out, "queries") || !strings.Contains(out, "haiku") || !strings.Contains(out, "next up: sonnet") {
		t.Errorf("models:\n%s", out)
	}
	if md := h.ok("eval", "models", "--dir", dir, "--md"); !strings.Contains(md, "| queries | haiku |") {
		t.Errorf("markdown:\n%s", md)
	}
	var j struct {
		Areas []struct {
			Area, Cheapest string
		} `json:"areas"`
	}
	h.jsonOf(&j, "eval", "models", "--dir", dir)
	if len(j.Areas) != 1 || j.Areas[0].Cheapest != "haiku" {
		t.Errorf("json: %+v", j)
	}
	h.fails("no cells for variant", "eval", "models", "--dir", dir, "--variant", "nope")
}
