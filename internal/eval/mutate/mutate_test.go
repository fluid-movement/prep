package mutate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/matrix"
	"github.com/fluid-movement/prep/internal/eval/run"
	"github.com/fluid-movement/prep/internal/eval/variant"
)

func shipped(t *testing.T) Plugin {
	t.Helper()
	return Plugin{Dir: "../../../plugins/claude-code"}
}

func TestSkillRoundTrip(t *testing.T) {
	p := shipped(t)
	text, err := p.read(SkillPath)
	if err != nil {
		t.Fatal(err)
	}
	s, err := parseSkill(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.rules) < 10 {
		t.Fatalf("only %d rules parsed", len(s.rules))
	}
	if got := s.render(); got != text {
		i := 0
		for i < len(got) && i < len(text) && got[i] == text[i] {
			i++
		}
		t.Fatalf("round trip differs at byte %d: %q vs %q", i, got[max(0, i-30):min(len(got), i+30)], text[max(0, i-30):min(len(text), i+30)])
	}
}

func TestOperatorsOnShippedPlugin(t *testing.T) {
	p := shipped(t)
	text, _ := p.read(SkillPath)
	s, _ := parseSkill(text)
	seen := map[string]bool{}
	for name, op := range Operators() {
		ms, err := op.Apply(p)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(ms) == 0 {
			t.Errorf("%s produced no mutants", name)
		}
		for _, m := range ms {
			if seen[m.ID] {
				t.Errorf("duplicate mutant id %s", m.ID)
			}
			seen[m.ID] = true
			if !idPattern.MatchString(m.ID) || m.Description == "" || m.Hypothesis == "" {
				t.Errorf("%s: incomplete mutant %+v", name, m)
			}
			for path, content := range m.Files {
				if path == SkillPath && content == text {
					t.Errorf("%s: mutant %s does not change the skill", name, m.ID)
				}
			}
		}
	}
	for _, id := range []string{"drop-rule-1", "drop-rule-" + itoa(len(s.rules)), "move-rule-3-first", "move-rule-1-last", "drop-hook-sessionstart", "drop-bridge", "drop-section-without-the-binary"} {
		if !seen[id] {
			t.Errorf("mutant %s not generated (have %d)", id, len(seen))
		}
	}
	// drop-rule removes exactly one rule and renumbers
	ms, _ := dropRule{}.Apply(p)
	d := ms[3].Files[SkillPath]
	ds, err := parseSkill(d)
	if err != nil || len(ds.rules) != len(s.rules)-1 || !strings.HasPrefix(ds.rules[3], s.rules[4][:20]) {
		t.Errorf("drop-rule-4: %v %d rules", err, len(ds.rules))
	}
	if !strings.Contains(d, "\n4. "+s.rules[4][:20]) {
		t.Error("rules not renumbered")
	}
}

func itoa(n int) string {
	return strings.TrimSpace(strings.Repeat(" ", 0) + string(rune('0'+n/10)) + string(rune('0'+n%10)))
}

func TestWriteMaterializeAndValidate(t *testing.T) {
	root := t.TempDir()
	res := variant.Resolver{Root: filepath.Join(root, "variants"), PluginDir: "../../../plugins/claude-code"}
	ms, err := dropRule{}.Apply(shipped(t))
	if err != nil {
		t.Fatal(err)
	}
	hooks, _ := dropHook{}.Apply(shipped(t))
	hashes := map[string]bool{}
	for _, m := range []Mutant{ms[0], ms[1], hooks[0]} {
		name, err := Write(res.Root, "current", m)
		if err != nil {
			t.Fatal(err)
		}
		v, err := res.Materialize(name, filepath.Join(t.TempDir(), "p"))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := Validate(v.PluginDir); err != nil {
			t.Errorf("%s invalid: %v", name, err)
		}
		hashes[v.Hash] = true
	}
	if len(hashes) != 3 {
		t.Errorf("mutants share hashes: %v", hashes)
	}
	if _, err := Write(res.Root, "current", Mutant{ID: "Bad ID"}); err == nil {
		t.Error("bad id accepted")
	}
	// the shipped plugin is untouched
	if err := Validate("../../../plugins/claude-code"); err != nil {
		t.Errorf("shipped plugin: %v", err)
	}
}

func TestValidateRejectsBrokenPlugins(t *testing.T) {
	mk := func(skill, hooks, manifest string) string {
		d := t.TempDir()
		os.MkdirAll(filepath.Join(d, ".claude-plugin"), 0o755)
		os.MkdirAll(filepath.Join(d, "skills", "prep"), 0o755)
		os.MkdirAll(filepath.Join(d, "hooks"), 0o755)
		os.WriteFile(filepath.Join(d, ".claude-plugin", "plugin.json"), []byte(manifest), 0o644)
		os.WriteFile(filepath.Join(d, "skills", "prep", "SKILL.md"), []byte(skill), 0o644)
		os.WriteFile(filepath.Join(d, "hooks", "hooks.json"), []byte(hooks), 0o644)
		return d
	}
	good := "---\nname: prep\ndescription: d\n---\n\nbody\n"
	if err := Validate(mk(good, "{}", `{"name":"prep"}`)); err != nil {
		t.Errorf("good plugin: %v", err)
	}
	for name, d := range map[string]string{
		"no frontmatter": mk("body", "{}", `{"name":"prep"}`),
		"no name":        mk("---\ndescription: d\n---\n", "{}", `{"name":"prep"}`),
		"bad hooks":      mk(good, "{", `{"name":"prep"}`),
		"bad manifest":   mk(good, "{}", `{}`),
	} {
		if Validate(d) == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestPropose(t *testing.T) {
	p := shipped(t)
	text, _ := p.read(SkillPath)
	short := strings.SplitN(text, "\n1. ", 2)[0] + "\n1. Run prep prime, then prep guide.\n"
	reply := `Sure! {"mutants":[
	  {"id":"Tiny Skill","hypothesis":"shorter is followed better","description":"cut to two rules","skill_md":` + jsonString(short) + `},
	  {"id":"tiny skill","hypothesis":"dup","description":"dup","skill_md":` + jsonString(short+"x") + `},
	  {"id":"same","hypothesis":"h","description":"d","skill_md":` + jsonString(text) + `},
	  {"id":"no-front","hypothesis":"h","description":"d","skill_md":"just text"},
	  {"id":"huge","hypothesis":"h","description":"d","skill_md":` + jsonString(text+text+text) + `}
	]}`
	var prompt string
	ms, err := Propose(context.Background(), func(ctx context.Context, pr string) (string, error) { prompt = pr; return reply, nil }, p, "ran guide failed 5x", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 1 || ms[0].ID != "ai-tiny-skill" || ms[0].Hypothesis == "" || !strings.Contains(ms[0].Description, "Agent-proposed") {
		t.Errorf("%+v", ms)
	}
	if !strings.Contains(prompt, "ran guide failed 5x") || !strings.Contains(prompt, "3 DIFFERENT") || !strings.Contains(prompt, "name: prep") {
		t.Error("prompt lacks the failures, the count or the skill")
	}
	for _, bad := range []string{"no json", `{"mutants": []}`, `{"mutants": [{"id":"x"}]}`} {
		if _, err := Propose(context.Background(), func(ctx context.Context, pr string) (string, error) { return bad, nil }, p, "", 2); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func jsonString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func cellOf(c, v, m string, k, f int, usd float64) *matrix.CellResult {
	cr := &matrix.CellResult{Cell: matrix.Cell{Case: c, Variant: v, Model: m}}
	for i := 0; i < k+f; i++ {
		cr.Runs = append(cr.Runs, matrix.RunResult{Index: i, Status: run.Completed, Cost: matrix.Cost{USD: usd},
			Assertions: []assert.Result{{Kind: "ran", Desc: "ran guide", Pass: i < k}}})
	}
	matrix.Reduce(cr)
	return cr
}

func TestRank(t *testing.T) {
	areas := map[string]string{"a1": "define", "a2": "define", "b1": "queries"}
	cells := []*matrix.CellResult{
		cellOf("a1", "base", "m", 15, 5, 0.10), cellOf("a2", "base", "m", 15, 5, 0.10), cellOf("b1", "base", "m", 20, 0, 0.10),
		// better: define up, queries same, cheaper
		cellOf("a1", "mutants/better", "m", 20, 0, 0.08), cellOf("a2", "mutants/better", "m", 20, 0, 0.08), cellOf("b1", "mutants/better", "m", 20, 0, 0.08),
		// worse: define collapses
		cellOf("a1", "mutants/worse", "m", 2, 18, 0.10), cellOf("a2", "mutants/worse", "m", 2, 18, 0.10), cellOf("b1", "mutants/worse", "m", 20, 0, 0.10),
		// noise: one point different
		cellOf("a1", "mutants/noise", "m", 16, 4, 0.10), cellOf("a2", "mutants/noise", "m", 15, 5, 0.10), cellOf("b1", "mutants/noise", "m", 20, 0, 0.10),
	}
	rows := Rank(cells, "base", []string{"mutants/noise", "mutants/worse", "mutants/better"}, areas, map[string]string{"mutants/better": "the better one"})
	if len(rows) != 3 || rows[0].Variant != "mutants/better" || rows[2].Variant != "mutants/worse" {
		t.Fatalf("order: %+v", rows)
	}
	b := rows[0]
	if !b.Significant || b.Delta <= 0 || b.CostPct > -19 || b.CostPct < -21 || b.ByArea["define"] <= 0 || b.ByArea["queries"] != 0 || b.Cells != 3 {
		t.Errorf("better: %+v", b)
	}
	if w := rows[2]; !w.Significant || w.Delta >= 0 || w.Regressed != 2 {
		t.Errorf("worse: %+v", w)
	}
	if n := rows[1]; n.Significant {
		t.Errorf("noise counted as significant: %+v", n)
	}
	out := Format("base", rows[0].BaseRate, rows)
	for _, want := range []string{"BETTER beyond the noise", "WORSE beyond the noise", "within the noise", "the better one", "by area (pts): define"} {
		if !strings.Contains(out, want) {
			t.Errorf("format lacks %q:\n%s", want, out)
		}
	}
}
