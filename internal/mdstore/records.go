package mdstore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/okf"
)

// Frontmatter shapes. Field order is the canonical key order.

type issueFM struct {
	Title     string                     `yaml:"title"`
	Kind      string                     `yaml:"kind"`
	Parent    string                     `yaml:"parent,omitempty"`
	DependsOn []string                   `yaml:"depends_on,omitempty"`
	Tags      []string                   `yaml:"tags,omitempty"`
	Priority  string                     `yaml:"priority,omitempty"`
	Artifacts []domain.ArtifactReference `yaml:"artifacts,omitempty"`
}

type baselineFM struct {
	By   string    `yaml:"by"`
	At   time.Time `yaml:"at"`
	Kind string    `yaml:"kind"`
	Ack  bool      `yaml:"ack,omitempty"`
}

type readyFM struct {
	By       string    `yaml:"by"`
	At       time.Time `yaml:"at"`
	Baseline string    `yaml:"baseline"`
}

type claimFM struct {
	By string    `yaml:"by"`
	At time.Time `yaml:"at"`
}

type docFM struct {
	Entries  []string `yaml:"entries,omitempty"`
	NoImpact string   `yaml:"no_impact,omitempty"`
}

type resolutionFM struct {
	Outcome       string    `yaml:"outcome"`
	By            string    `yaml:"by"`
	At            time.Time `yaml:"at"`
	Reason        string    `yaml:"reason,omitempty"`
	Evidence      string    `yaml:"evidence,omitempty"`
	Documentation *docFM    `yaml:"documentation,omitempty"`
}

type projectFM struct {
	Schema int               `yaml:"schema"`
	ID     string            `yaml:"id,omitempty"`
	Name   string            `yaml:"name,omitempty"`
	Links  map[string]linkFM `yaml:"links,omitempty"`
}

// linkFM is one link under its local name; yaml writes the names sorted.
type linkFM struct {
	ID     string `yaml:"id"`
	Remote string `yaml:"remote,omitempty"`
}

// project returns the frontmatter's facts.
func (f projectFM) project() domain.Project {
	p := domain.Project{Schema: f.Schema, ID: f.ID, Name: f.Name}
	for name, l := range f.Links {
		p.Links = append(p.Links, domain.ProjectLink{Name: name, ID: l.ID, Remote: l.Remote})
	}
	sort.Slice(p.Links, func(a, b int) bool { return p.Links[a].Name < p.Links[b].Name })
	return p
}

func projectFMOf(p domain.Project) projectFM {
	f := projectFM{Schema: p.Schema, ID: p.ID, Name: p.Name}
	if len(p.Links) > 0 {
		f.Links = map[string]linkFM{}
		for _, l := range p.Links {
			f.Links[l.Name] = linkFM{ID: l.ID, Remote: l.Remote}
		}
	}
	return f
}

func fmtTime(t time.Time) time.Time { return t.UTC().Truncate(time.Second) }

func parseTime(t time.Time) (time.Time, error) {
	if t.IsZero() {
		return t, fmt.Errorf("at is missing")
	}
	return t.UTC(), nil
}

// --- issue.md ---

var openQuestionsRe = regexp.MustCompile(`(?i)^##\s+open questions\s*$`)
var headingRe = regexp.MustCompile(`^#{1,2}\s`)

// splitOpenQuestions separates the Open questions section from the prose.
// The section runs to the next level one or two heading; headings inside
// fenced code blocks count for neither.
func splitOpenQuestions(body string) (prose, questions string) {
	lines := strings.Split(body, "\n")
	start, end := -1, len(lines)
	var fence domain.Fence
	for k, l := range lines {
		if in, _ := fence.Line(l); in {
			continue
		}
		if start < 0 && openQuestionsRe.MatchString(l) {
			start = k
		} else if start >= 0 && headingRe.MatchString(l) {
			end = k
			break
		}
	}
	if start < 0 {
		return body, ""
	}
	questions = strings.TrimSpace(strings.Join(lines[start+1:end], "\n"))
	prose = strings.TrimSpace(strings.Join(lines[:start], "\n") + "\n" + strings.Join(lines[end:], "\n"))
	return prose, questions
}

// hasOpenQuestions reports whether text has an Open questions heading
// outside fenced code blocks.
func hasOpenQuestions(text string) bool {
	var fence domain.Fence
	for l := range strings.SplitSeq(text, "\n") {
		if in, _ := fence.Line(l); !in && openQuestionsRe.MatchString(l) {
			return true
		}
	}
	return false
}

func parseIssue(raw string, i *domain.Issue) error {
	fm, body, ok, err := splitFrontmatter(raw)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("issue.md must start with frontmatter (---)")
	}
	var f issueFM
	if err := decodeStrict(fm, &f); err != nil {
		return err
	}
	i.Title = strings.TrimSpace(f.Title)
	i.Kind = domain.Kind(f.Kind)
	i.Parent = f.Parent
	i.DependsOn = f.DependsOn
	i.Tags = f.Tags
	i.Priority = domain.Priority(f.Priority)
	i.Artifacts = f.Artifacts
	i.Body = normalize(body)
	i.Prose, i.OpenQuestions = splitOpenQuestions(i.Body)
	return nil
}

func renderIssue(i *domain.Issue) string {
	// Medium is the default and never written; an invalid level is kept for prep check to report.
	prio := string(i.Priority)
	if i.Priority.Effective() == domain.PriorityMedium {
		prio = ""
	}
	return withFrontmatter(encodeYAML(issueFM{Title: i.Title, Kind: string(i.Kind), Parent: i.Parent, DependsOn: i.DependsOn, Tags: i.Tags, Priority: prio, Artifacts: i.Artifacts}), i.Body)
}

// requirementBody is the body of issue.md for a requirement given as text.
// It keeps the text's own Open questions section and adds an empty one only
// when the text has none.
func requirementBody(text string) string {
	text = normalize(text)
	if hasOpenQuestions(text) {
		return text
	}
	if text == "" {
		return "## Open questions"
	}
	return text + "\n\n## Open questions"
}

// --- acceptance.md ---

var dodHeadingRe = regexp.MustCompile(`(?i)^##\s+definition of done\s*$`)
var h2Re = regexp.MustCompile(`^##\s`)
var bulletRe = regexp.MustCompile(`^\s*[-*+]\s+(?:\[[ xX]\]\s+)?(.*)$`)

// parseAcceptance reads checkable criteria.
func parseAcceptance(raw string, i *domain.Issue) {
	for _, l := range strings.Split(normalize(raw), "\n") {
		if m := checkboxRe.FindStringSubmatch(l); m != nil {
			i.Criteria = append(i.Criteria, domain.Criterion{Text: strings.TrimSpace(m[3]), Checked: m[2] != " "})
		}
	}
}

// applyAcceptance changes criteria while preserving surrounding prose.
func applyAcceptance(raw string, ops []domain.AcceptanceOp) string {
	set := map[int]string{}
	removed := map[int]bool{}
	var adds []string
	for _, o := range ops {
		switch o.Op {
		case "check":
			set[o.Index] = "x"
		case "uncheck":
			set[o.Index] = " "
		case "remove":
			removed[o.Index] = true
		case "add":
			adds = append(adds, "- [ ] "+strings.TrimSpace(o.Text))
		}
	}
	var out []string
	n, addAt := 0, -1
	for _, l := range strings.Split(normalize(raw), "\n") {
		if m := checkboxRe.FindStringSubmatch(l); m != nil {
			n++
			if removed[n] {
				addAt = len(out)
				continue
			}
			if mark, ok := set[n]; ok {
				l = m[1] + "- [" + mark + "] " + m[3]
			}
			out = append(out, l)
			addAt = len(out)
		} else {
			out = append(out, l)
		}
	}
	if len(adds) > 0 {
		if addAt < 0 {
			addAt = len(out)
		}
		out = insertLines(out, addAt, adds)
	}
	return normalize(strings.Join(out, "\n"))
}

// insertLines inserts ins at position at, joining the list that ends there:
// blank lines before the position are skipped, a blank line separates the
// inserted lines from preceding non-list text and from what follows.
func insertLines(lines []string, at int, ins []string) []string {
	for at > 0 && strings.TrimSpace(lines[at-1]) == "" {
		at--
	}
	out := append([]string(nil), lines[:at]...)
	if at > 0 && !bulletRe.MatchString(lines[at-1]) {
		out = append(out, "")
	}
	out = append(out, ins...)
	if at < len(lines) {
		out = append(out, "")
	}
	return append(out, lines[at:]...)
}

// renderDecision renders one decisions.md entry.
func renderDecision(d *domain.Decision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s: %s\ndate: %s\n", d.ID, d.Title, d.Date)
	if d.Supersedes != "" {
		fmt.Fprintf(&b, "supersedes: %s\n", d.Supersedes)
	}
	if d.Outcome {
		b.WriteString("outcome: true\n")
	}
	if d.Body != "" {
		b.WriteString("\n" + d.Body + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// --- context.md ---

var bundleLinkRe = regexp.MustCompile(`\]\((?:\.prep/knowledge)?(/[^)\s#]+\.md)(?:#[^)]*)?\)`)
var codeSpanRe = regexp.MustCompile("`([^`\\s]+)`")
var pathLikeRe = regexp.MustCompile(`^[A-Za-z0-9_.][A-Za-z0-9_./*-]*$`)

func parseContext(raw string, i *domain.Issue) {
	i.Context = normalize(raw)
	seen := map[string]bool{}
	for _, m := range bundleLinkRe.FindAllStringSubmatch(okf.StripCode(raw), -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			i.ContextLinks = append(i.ContextLinks, m[1])
		}
	}
	for _, m := range codeSpanRe.FindAllStringSubmatch(raw, -1) {
		p := m[1]
		if !pathLikeRe.MatchString(p) || seen[p] {
			continue
		}
		if strings.Contains(p, "/") || regexp.MustCompile(`\.[A-Za-z0-9]{1,8}$`).MatchString(p) {
			seen[p] = true
			i.ContextPaths = append(i.ContextPaths, p)
		}
	}
}

// --- decisions.md ---

var decisionHeadRe = regexp.MustCompile(`^##\s+([A-Za-z0-9][A-Za-z0-9_-]*):\s*(.+)$`)
var decisionKVRe = regexp.MustCompile(`^(date|supersedes|outcome):\s*(.*)$`)
var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

func parseDecisions(raw string) ([]domain.Decision, []string) {
	var out []domain.Decision
	var problems []string
	var cur *domain.Decision
	var body []string
	inMeta := false
	flush := func() {
		if cur != nil {
			cur.Body = strings.TrimSpace(strings.Join(body, "\n"))
			if cur.Date == "" {
				problems = append(problems, fmt.Sprintf("decision %s has no date", cur.ID))
			}
			out = append(out, *cur)
		}
		cur, body = nil, nil
	}
	var fence domain.Fence
	for _, l := range strings.Split(normalize(raw), "\n") {
		if in, _ := fence.Line(l); !in && strings.HasPrefix(l, "## ") {
			flush()
			m := decisionHeadRe.FindStringSubmatch(l)
			if m == nil {
				problems = append(problems, fmt.Sprintf("heading %q is not '## <id>: <title>'", l))
				continue
			}
			cur = &domain.Decision{ID: m[1], Title: strings.TrimSpace(m[2])}
			inMeta = true
			continue
		}
		if cur == nil {
			continue
		}
		if inMeta {
			if strings.TrimSpace(l) == "" {
				continue // blank lines before the metadata are tolerated
			}
			if m := decisionKVRe.FindStringSubmatch(l); m != nil {
				v := strings.TrimSpace(m[2])
				switch m[1] {
				case "date":
					if !dateRe.MatchString(v) {
						problems = append(problems, fmt.Sprintf("decision %s: date %q is not YYYY-MM-DD", cur.ID, v))
					}
					cur.Date = v
				case "supersedes":
					cur.Supersedes = v
				case "outcome":
					cur.Outcome = v == "true" || v == "yes"
				}
				continue
			}
			inMeta = false
		}
		body = append(body, l)
	}
	flush()
	return out, problems
}

// canonicalDecisions renders decisions.md canonically: blank lines between
// a heading and its metadata lines are removed.
func canonicalDecisions(raw string) string {
	lines := strings.Split(normalize(raw), "\n")
	var out []string
	inMeta := false
	var fence domain.Fence
	for k, l := range lines {
		in, _ := fence.Line(l)
		switch {
		case in:
			inMeta = false
		case strings.HasPrefix(l, "## "):
			inMeta = true
		case inMeta && strings.TrimSpace(l) == "":
			next := ""
			for _, n := range lines[k+1:] {
				if strings.TrimSpace(n) != "" {
					next = n
					break
				}
			}
			if decisionKVRe.MatchString(next) {
				continue
			}
			inMeta = false
		case inMeta && !decisionKVRe.MatchString(l):
			inMeta = false
		}
		out = append(out, l)
	}
	return fileText(strings.Join(out, "\n"))
}

// --- baselines, ready, claim, resolution ---

func parseBaseline(name, raw string) (domain.Baseline, error) {
	fm, body, ok, err := splitFrontmatter(raw)
	if err == nil && !ok {
		err = fmt.Errorf("missing frontmatter")
	}
	if err != nil {
		return domain.Baseline{}, err
	}
	var f baselineFM
	if err := decodeStrict(fm, &f); err != nil {
		return domain.Baseline{}, err
	}
	at, err := parseTime(f.At)
	if err != nil {
		return domain.Baseline{}, err
	}
	return domain.Baseline{Name: name, By: f.By, At: at, Kind: domain.Kind(f.Kind), Ack: f.Ack, Requirement: normalize(body)}, nil
}

func renderBaseline(b *domain.Baseline) string {
	return withFrontmatter(encodeYAML(baselineFM{By: b.By, At: fmtTime(b.At), Kind: string(b.Kind), Ack: b.Ack}), b.Requirement)
}

func parseReady(raw string) (*domain.Ready, error) {
	fm, body, ok, err := splitFrontmatter(raw)
	if err == nil && !ok {
		err = fmt.Errorf("missing frontmatter")
	}
	if err != nil {
		return nil, err
	}
	var f readyFM
	if err := decodeStrict(fm, &f); err != nil {
		return nil, err
	}
	at, err := parseTime(f.At)
	if err != nil {
		return nil, err
	}
	return &domain.Ready{By: f.By, At: at, Baseline: f.Baseline, Note: normalize(body)}, nil
}

func renderReady(r *domain.Ready) string {
	return withFrontmatter(encodeYAML(readyFM{By: r.By, At: fmtTime(r.At), Baseline: r.Baseline}), r.Note)
}

func parseClaim(raw string) (*domain.Claim, error) {
	fm, body, ok, err := splitFrontmatter(raw)
	if err == nil && !ok {
		err = fmt.Errorf("missing frontmatter")
	}
	if err != nil {
		return nil, err
	}
	var f claimFM
	if err := decodeStrict(fm, &f); err != nil {
		return nil, err
	}
	at, err := parseTime(f.At)
	if err != nil {
		return nil, err
	}
	return &domain.Claim{By: f.By, At: at, Note: normalize(body)}, nil
}

func renderClaim(c *domain.Claim) string {
	return withFrontmatter(encodeYAML(claimFM{By: c.By, At: fmtTime(c.At)}), c.Note)
}

func parseResolution(raw string) (*domain.Resolution, error) {
	fm, body, ok, err := splitFrontmatter(raw)
	if err == nil && !ok {
		err = fmt.Errorf("missing frontmatter")
	}
	if err != nil {
		return nil, err
	}
	var f resolutionFM
	if err := decodeStrict(fm, &f); err != nil {
		return nil, err
	}
	return resolutionFromFM(f, body)
}

func resolutionFromFM(f resolutionFM, body string) (*domain.Resolution, error) {
	at, err := parseTime(f.At)
	if err != nil {
		return nil, err
	}
	r := &domain.Resolution{Outcome: f.Outcome, By: f.By, At: at, Reason: f.Reason, Evidence: f.Evidence, Note: normalize(body)}
	if f.Documentation != nil {
		r.Documentation = &domain.Documentation{NoImpact: f.Documentation.NoImpact}
		for _, e := range f.Documentation.Entries {
			r.Documentation.Entries = append(r.Documentation.Entries, domain.NormalizeEntryPath(e))
		}
	}
	return r, nil
}

func renderResolution(r *domain.Resolution) string {
	return withFrontmatter(encodeYAML(resolutionToFM(r)), r.Note)
}

func resolutionToFM(r *domain.Resolution) resolutionFM {
	f := resolutionFM{Outcome: r.Outcome, By: r.By, At: fmtTime(r.At), Reason: r.Reason, Evidence: r.Evidence}
	if r.Documentation != nil {
		f.Documentation = &docFM{Entries: r.Documentation.Entries, NoImpact: r.Documentation.NoImpact}
	}
	return f
}

// --- project.md ---

func parseProject(raw string) (domain.Project, error) {
	var p domain.Project
	fm, body, ok, err := splitFrontmatter(raw)
	if err == nil && !ok {
		err = fmt.Errorf("project.md must start with frontmatter containing schema")
	}
	if err != nil {
		return p, err
	}
	var f projectFM
	if err := decodeStrict(fm, &f); err != nil {
		return p, err
	}
	if f.Schema == 0 {
		return p, fmt.Errorf("schema is missing")
	}
	p = f.project()
	inDoD := false
	for _, l := range strings.Split(normalize(body), "\n") {
		if h2Re.MatchString(l) {
			inDoD = dodHeadingRe.MatchString(l)
			continue
		}
		if inDoD {
			if m := bulletRe.FindStringSubmatch(l); m != nil && strings.TrimSpace(m[1]) != "" {
				p.DoD = append(p.DoD, strings.TrimSpace(m[1]))
			}
		}
	}
	return p, nil
}

// renderProject writes p's frontmatter facts over body.
func renderProject(p domain.Project, body string) string {
	return withFrontmatter(encodeYAML(projectFMOf(p)), normalize(body))
}

// ReadProject reads the project facts in root's project.md (not the config),
// for callers that only need a project's identity, such as the registry.
func ReadProject(root string) (domain.Project, error) {
	b, err := os.ReadFile(filepath.Join(root, Dir, "project.md"))
	if err != nil {
		return domain.Project{}, err
	}
	return parseProject(string(b))
}

// replaceProjectDoD preserves all project prose outside the guidance section.
func replaceProjectDoD(body string, items []string) string {
	var out []string
	inDoD, found := false, false
	section := func() {
		out = append(out, "## Definition of Done", "")
		for _, item := range items {
			out = append(out, "- "+item)
		}
		out = append(out, "")
	}
	for _, line := range strings.Split(normalize(body), "\n") {
		if h2Re.MatchString(line) {
			inDoD = dodHeadingRe.MatchString(line)
			if inDoD {
				if !found {
					section()
					found = true
				}
				continue
			}
		}
		if !inDoD {
			out = append(out, line)
		}
	}
	if !found {
		out = append(out, "")
		section()
	}
	return normalize(strings.Join(out, "\n"))
}
