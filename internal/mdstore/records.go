package mdstore

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/okf"
)

// Frontmatter shapes. Field order is the canonical key order.

type issueFM struct {
	Title     string   `yaml:"title"`
	Kind      string   `yaml:"kind"`
	Parent    string   `yaml:"parent,omitempty"`
	DependsOn []string `yaml:"depends_on,omitempty"`
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

type optOutFM struct {
	Item   string `yaml:"item"`
	Reason string `yaml:"reason,omitempty"`
}

type resolutionFM struct {
	Outcome       string     `yaml:"outcome"`
	By            string     `yaml:"by"`
	At            time.Time  `yaml:"at"`
	Reason        string     `yaml:"reason,omitempty"`
	Evidence      string     `yaml:"evidence,omitempty"`
	Documentation *docFM     `yaml:"documentation,omitempty"`
	DoD           []string   `yaml:"dod,omitempty"`
	DoDOptOuts    []optOutFM `yaml:"dod_opt_outs,omitempty"`
}

type projectFM struct {
	Schema int `yaml:"schema"`
}

func fmtTime(t time.Time) time.Time { return t.UTC().Truncate(time.Second) }

func parseTime(t time.Time) (time.Time, error) {
	if t.IsZero() {
		return t, fmt.Errorf("at is missing")
	}
	return t.UTC(), nil
}

// --- issue.md ---

var openQuestionsRe = regexp.MustCompile(`(?im)^##\s+open questions\s*$`)
var headingRe = regexp.MustCompile(`(?m)^#{1,2}\s`)

// splitOpenQuestions separates the Open questions section from the prose.
func splitOpenQuestions(body string) (prose, questions string) {
	loc := openQuestionsRe.FindStringIndex(body)
	if loc == nil {
		return body, ""
	}
	rest := body[loc[1]:]
	end := len(rest)
	if m := headingRe.FindStringIndex(rest); m != nil {
		end = m[0]
	}
	questions = strings.TrimSpace(rest[:end])
	prose = strings.TrimSpace(body[:loc[0]] + rest[end:])
	return prose, questions
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
	i.Body = normalize(body)
	i.Prose, i.OpenQuestions = splitOpenQuestions(i.Body)
	return nil
}

func renderIssue(i *domain.Issue) string {
	return withFrontmatter(encodeYAML(issueFM{Title: i.Title, Kind: string(i.Kind), Parent: i.Parent, DependsOn: i.DependsOn}), i.Body)
}

// requirementBody is the body of issue.md for a requirement given as text.
// It keeps the text's own Open questions section and adds an empty one only
// when the text has none.
func requirementBody(text string) string {
	text = normalize(text)
	if openQuestionsRe.MatchString(text) {
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
var optOutRe = regexp.MustCompile(`(?i)^opt-out:\s*(.+?)(?:\s+(?:—|--)\s+(.*))?$`)

func parseAcceptance(raw string, i *domain.Issue) {
	inDoD := false
	for _, l := range strings.Split(normalize(raw), "\n") {
		if h2Re.MatchString(l) {
			inDoD = dodHeadingRe.MatchString(l)
			continue
		}
		if !inDoD {
			if m := checkboxRe.FindStringSubmatch(l); m != nil {
				i.Criteria = append(i.Criteria, domain.Criterion{Text: strings.TrimSpace(m[3]), Checked: m[2] != " "})
			}
			continue
		}
		m := bulletRe.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		item := strings.TrimSpace(m[1])
		if o := optOutRe.FindStringSubmatch(item); o != nil {
			i.DoDOptOuts = append(i.DoDOptOuts, domain.OptOut{Item: strings.TrimSpace(o[1]), Reason: strings.TrimSpace(o[2])})
		} else if item != "" {
			i.DoDAdd = append(i.DoDAdd, item)
		}
	}
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
	for _, l := range strings.Split(normalize(raw), "\n") {
		if strings.HasPrefix(l, "## ") {
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
	at, err := parseTime(f.At)
	if err != nil {
		return nil, err
	}
	r := &domain.Resolution{Outcome: f.Outcome, By: f.By, At: at, Reason: f.Reason, Evidence: f.Evidence, DoD: f.DoD, Note: normalize(body)}
	if f.Documentation != nil {
		r.Documentation = &domain.Documentation{NoImpact: f.Documentation.NoImpact}
		for _, e := range f.Documentation.Entries {
			r.Documentation.Entries = append(r.Documentation.Entries, domain.NormalizeEntryPath(e))
		}
	}
	for _, o := range f.DoDOptOuts {
		r.DoDOptOuts = append(r.DoDOptOuts, domain.OptOut{Item: o.Item, Reason: o.Reason})
	}
	return r, nil
}

func renderResolution(r *domain.Resolution) string {
	f := resolutionFM{Outcome: r.Outcome, By: r.By, At: fmtTime(r.At), Reason: r.Reason, Evidence: r.Evidence, DoD: r.DoD}
	if r.Documentation != nil {
		f.Documentation = &docFM{Entries: r.Documentation.Entries, NoImpact: r.Documentation.NoImpact}
	}
	for _, o := range r.DoDOptOuts {
		f.DoDOptOuts = append(f.DoDOptOuts, optOutFM{Item: o.Item, Reason: o.Reason})
	}
	return withFrontmatter(encodeYAML(f), r.Note)
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
	p.Schema = f.Schema
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

func renderProject(schema int, body string) string {
	return withFrontmatter(encodeYAML(projectFM{Schema: schema}), normalize(body))
}
