// Package mutate generates variants of the harness configuration — the
// skill's rules and the hooks — so the suite can measure which changes raise
// or lower the pass rate and the cost. A mutant is an ordinary variant
// written under eval/variants/mutants/, never an edit of the shipped plugin.
package mutate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Mutant is one proposed change to a plugin.
type Mutant struct {
	// ID names the mutant and its directory: lower-case letters, digits and
	// hyphens.
	ID          string
	Description string
	// Hypothesis says why the change might help.
	Hypothesis string
	// Files maps plugin-relative paths to their new contents; Remove lists
	// paths to delete.
	Files  map[string]string
	Remove []string
}

// Plugin is a materialized plugin directory a mutant is derived from.
type Plugin struct{ Dir string }

// SkillPath and HooksPath are where the mutated texts live in a plugin.
const (
	SkillPath = "skills/prep/SKILL.md"
	HooksPath = "hooks/hooks.json"
)

func (p Plugin) read(rel string) (string, error) {
	b, err := os.ReadFile(filepath.Join(p.Dir, filepath.FromSlash(rel)))
	return string(b), err
}

// Operator generates mutants from a plugin.
type Operator interface {
	Name() string
	Apply(p Plugin) ([]Mutant, error)
}

// Operators returns the built-in operators by name.
func Operators() map[string]Operator {
	ops := []Operator{dropRule{}, firstSentence{}, moveRule{}, replaceWording{}, dropSection{}, dropHook{}}
	m := map[string]Operator{}
	for _, o := range ops {
		m[o.Name()] = o
	}
	return m
}

// OperatorNames lists the built-in operator names, sorted.
func OperatorNames() []string {
	var n []string
	for k := range Operators() {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}

// ---- skill structure

// skill splits a SKILL.md into the text before the numbered rules, the
// rules, and the text after them.
type skill struct {
	head  string
	rules []string // each without its number
	tail  string
}

var ruleLine = regexp.MustCompile(`^(\d+)\. (.*)$`)

func parseSkill(text string) (skill, error) {
	lines := strings.Split(text, "\n")
	first, last := -1, -1
	for i, l := range lines {
		if ruleLine.MatchString(l) {
			if first < 0 {
				first = i
			}
			last = i
		} else if first >= 0 && last == i-1 && strings.TrimSpace(l) != "" && !strings.HasPrefix(l, "#") {
			return skill{}, fmt.Errorf("the rule list is interrupted at line %d", i+1)
		}
	}
	if first < 0 {
		return skill{}, fmt.Errorf("the skill has no numbered rules")
	}
	s := skill{head: strings.Join(lines[:first], "\n"), tail: strings.Join(lines[last+1:], "\n")}
	for _, l := range lines[first : last+1] {
		if m := ruleLine.FindStringSubmatch(l); m != nil {
			s.rules = append(s.rules, m[2])
		}
	}
	return s, nil
}

func (s skill) render() string {
	var b strings.Builder
	b.WriteString(s.head)
	b.WriteString("\n")
	for i, r := range s.rules {
		fmt.Fprintf(&b, "%d. %s\n", i+1, r)
	}
	b.WriteString(strings.TrimLeft(s.tail, "\n"))
	out := b.String()
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	// keep one blank line between the rules and what follows
	return strings.Replace(out, "\n## ", "\n\n## ", 1)
}

func (s skill) clone() skill {
	c := s
	c.rules = append([]string(nil), s.rules...)
	return c
}

func skillMutant(id, desc, hyp, text string) Mutant {
	return Mutant{ID: id, Description: desc, Hypothesis: hyp, Files: map[string]string{SkillPath: text}}
}

// firstWords is the start of a rule, for descriptions.
func firstWords(r string) string {
	r = strings.TrimSpace(r)
	if len(r) > 60 {
		return r[:60] + "…"
	}
	return r
}

// ---- operators

type dropRule struct{}

func (dropRule) Name() string { return "drop-rule" }
func (dropRule) Apply(p Plugin) ([]Mutant, error) {
	text, err := p.read(SkillPath)
	if err != nil {
		return nil, err
	}
	s, err := parseSkill(text)
	if err != nil {
		return nil, err
	}
	var out []Mutant
	for i := range s.rules {
		c := s.clone()
		c.rules = append(c.rules[:i:i], c.rules[i+1:]...)
		out = append(out, skillMutant(fmt.Sprintf("drop-rule-%d", i+1),
			fmt.Sprintf("Skill rule %d removed (%s)", i+1, firstWords(s.rules[i])),
			"If the agents behave the same without the rule, it costs tokens and attention for nothing.", c.render()))
	}
	return out, nil
}

type firstSentence struct{}

func (firstSentence) Name() string { return "first-sentence" }
func (firstSentence) Apply(p Plugin) ([]Mutant, error) {
	text, err := p.read(SkillPath)
	if err != nil {
		return nil, err
	}
	s, err := parseSkill(text)
	if err != nil {
		return nil, err
	}
	var out []Mutant
	for i, r := range s.rules {
		cut := firstSentenceOf(r)
		if len(cut) >= len(r)-20 { // nothing worth cutting
			continue
		}
		c := s.clone()
		c.rules[i] = cut
		out = append(out, skillMutant(fmt.Sprintf("first-sentence-%d", i+1),
			fmt.Sprintf("Skill rule %d cut to its first sentence (%s)", i+1, firstWords(r)),
			"A shorter rule may be read and followed as well, and costs fewer tokens on every session.", c.render()))
	}
	return out, nil
}

// firstSentenceOf returns r up to the end of its first sentence.
func firstSentenceOf(r string) string {
	for i := 0; i+1 < len(r); i++ {
		if r[i] == '.' && r[i+1] == ' ' && i > 20 {
			// not inside backticks
			if strings.Count(r[:i], "`")%2 == 0 {
				return r[:i+1]
			}
		}
	}
	return r
}

type moveRule struct{}

func (moveRule) Name() string { return "move-rule" }
func (moveRule) Apply(p Plugin) ([]Mutant, error) {
	text, err := p.read(SkillPath)
	if err != nil {
		return nil, err
	}
	s, err := parseSkill(text)
	if err != nil {
		return nil, err
	}
	var out []Mutant
	for i, r := range s.rules {
		if i > 0 {
			c := s.clone()
			c.rules = append([]string{r}, append(c.rules[:i:i], c.rules[i+1:]...)...)
			out = append(out, skillMutant(fmt.Sprintf("move-rule-%d-first", i+1),
				fmt.Sprintf("Skill rule %d moved to the top (%s)", i+1, firstWords(r)),
				"Rules near the top are followed more reliably; a rule that is often broken may belong there.", c.render()))
		}
		if i < len(s.rules)-1 {
			c := s.clone()
			c.rules = append(append(c.rules[:i:i], c.rules[i+1:]...), r)
			out = append(out, skillMutant(fmt.Sprintf("move-rule-%d-last", i+1),
				fmt.Sprintf("Skill rule %d moved to the end (%s)", i+1, firstWords(r)),
				"If the rule matters less than its position suggests, moving it down changes nothing.", c.render()))
		}
	}
	return out, nil
}

type replaceWording struct{}

// wording is the replacement table: each entry softens or sharpens a phrase.
var wording = []struct{ id, from, to, hyp string }{
	{"soften-never", "never `cat` entries", "avoid `cat` on entries", "Softer wording may be followed less often."},
	{"soften-follow", "Follow it; it always matches", "Consider it; it matches", "A suggestion may be ignored where a command is followed."},
	{"sharpen-guide", "run `prep guide <id>`.", "ALWAYS run `prep guide <id>` first, before anything else.", "Emphasis may raise how often the guide is read first."},
	{"sharpen-no-files", "Read the knowledge base through prep, not its files", "NEVER read the knowledge base through its files; use prep", "Emphasis may stop direct reads of .prep."},
}

func (replaceWording) Name() string { return "replace-wording" }
func (replaceWording) Apply(p Plugin) ([]Mutant, error) {
	text, err := p.read(SkillPath)
	if err != nil {
		return nil, err
	}
	var out []Mutant
	for _, w := range wording {
		if !strings.Contains(text, w.from) {
			continue
		}
		out = append(out, skillMutant(w.id, fmt.Sprintf("Wording changed: %q becomes %q", w.from, w.to), w.hyp, strings.Replace(text, w.from, w.to, 1)))
	}
	return out, nil
}

type dropSection struct{}

func (dropSection) Name() string { return "drop-section" }
func (dropSection) Apply(p Plugin) ([]Mutant, error) {
	text, err := p.read(SkillPath)
	if err != nil {
		return nil, err
	}
	var out []Mutant
	idx := regexp.MustCompile(`(?m)^## .*$`).FindAllStringIndex(text, -1)
	for i, loc := range idx {
		end := len(text)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		heading := strings.TrimSpace(text[loc[0]:loc[1]])
		id := "drop-section-" + slug(strings.TrimPrefix(heading, "## "))
		out = append(out, skillMutant(id, fmt.Sprintf("Skill section %q removed", heading),
			"A section only needed in rare cases may cost every session its tokens.", strings.TrimRight(text[:loc[0]]+text[end:], "\n")+"\n"))
	}
	return out, nil
}

type dropHook struct{}

func (dropHook) Name() string { return "drop-hook" }
func (dropHook) Apply(p Plugin) ([]Mutant, error) {
	text, err := p.read(HooksPath)
	if err != nil {
		return nil, nil // a variant without hooks has nothing to drop
	}
	var h map[string]any
	if err := json.Unmarshal([]byte(text), &h); err != nil {
		return nil, fmt.Errorf("%s: %w", HooksPath, err)
	}
	var out []Mutant
	if hooks, ok := h["hooks"].(map[string]any); ok {
		var events []string
		for e := range hooks {
			events = append(events, e)
		}
		sort.Strings(events)
		for _, e := range events {
			c := cloneJSON(h)
			delete(c["hooks"].(map[string]any), e)
			data, _ := json.MarshalIndent(c, "", "  ")
			out = append(out, Mutant{ID: "drop-hook-" + slug(e), Description: fmt.Sprintf("The %s hook removed", e),
				Hypothesis: "If agents do as well without the hook's output, the briefing is not what makes them use prep.",
				Files:      map[string]string{HooksPath: string(data) + "\n"}})
		}
	}
	if _, ok := h["modules"]; ok {
		c := cloneJSON(h)
		delete(c, "modules")
		data, _ := json.MarshalIndent(c, "", "  ")
		out = append(out, Mutant{ID: "drop-bridge", Description: "The activity bridge module removed",
			Hypothesis: "The bridge sets the actor and session; without it prep sees agents as anonymous.",
			Files:      map[string]string{HooksPath: string(data) + "\n"}})
	}
	return out, nil
}

func cloneJSON(m map[string]any) map[string]any {
	b, _ := json.Marshal(m)
	var c map[string]any
	_ = json.Unmarshal(b, &c)
	return c
}

func slug(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	dash := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// ---- writing and validating

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// Validate checks that a mutated plugin directory is still a plugin: the
// manifest parses and names it, the hooks file parses, and the skill keeps
// its frontmatter.
func Validate(dir string) error {
	data, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json"))
	if err != nil {
		return fmt.Errorf("plugin manifest: %w", err)
	}
	var man struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &man); err != nil || man.Name == "" {
		return fmt.Errorf("plugin manifest has no name (%v)", err)
	}
	if data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(HooksPath))); err == nil {
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: %w", HooksPath, err)
		}
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "skills"))
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(dir, "skills", e.Name(), "SKILL.md"))
		if err != nil {
			continue
		}
		front := regexp.MustCompile(`(?s)\A---\n(.*?)\n---\n`).FindSubmatch(b)
		if front == nil {
			return fmt.Errorf("skill %s lost its frontmatter", e.Name())
		}
		var meta struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		}
		if err := yaml.Unmarshal(front[1], &meta); err != nil || meta.Name == "" || meta.Description == "" {
			return fmt.Errorf("skill %s frontmatter needs a name and a description (%v)", e.Name(), err)
		}
	}
	return nil
}

// Write stores a mutant as eval/variants/mutants/<id>/ (variant.yaml plus
// the changed files) and returns the variant's name.
func Write(variantsRoot, base string, m Mutant) (string, error) {
	if !idPattern.MatchString(m.ID) {
		return "", fmt.Errorf("mutant id %q must be lower-case letters, digits and hyphens", m.ID)
	}
	name := "mutants/" + m.ID
	dir := filepath.Join(variantsRoot, "mutants", m.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	spec := struct {
		Base        string            `yaml:"base"`
		Description string            `yaml:"description"`
		Replace     map[string]string `yaml:"replace,omitempty"`
		Remove      []string          `yaml:"remove,omitempty"`
	}{Base: base, Description: strings.TrimSpace(m.Description + "\nHypothesis: " + m.Hypothesis), Replace: map[string]string{}, Remove: m.Remove}
	var paths []string
	for p := range m.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for i, p := range paths {
		file := fmt.Sprintf("file%d-%s", i, filepath.Base(p))
		if err := os.WriteFile(filepath.Join(dir, file), []byte(m.Files[p]), 0o644); err != nil {
			return "", err
		}
		spec.Replace[p] = file
	}
	data, err := yaml.Marshal(spec)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "variant.yaml"), data, 0o644); err != nil {
		return "", err
	}
	return name, nil
}
