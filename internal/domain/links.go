package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Links to other projects. A project lists the projects it refers to in
// project.md, each under a local name with the target's permanent ID and its
// last known remote as a clone hint. Records refer to them as
// <name>:<full issue ID> (dependencies) and <name>:/path.md (knowledge links).
// The local name belongs to this project, so renaming the target, changing
// its remote or moving it never breaks a reference; renaming the local name
// goes through PlanLinkRename, which rewrites every reference.

// OpLinks writes project.md's links (and, for a rename, every reference).
const OpLinks Op = "links"

// ProjectLink is one entry of project.md's links.
type ProjectLink struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Remote string `json:"remote,omitempty"`
}

// Link returns the link with the given local name.
func (p Project) Link(name string) (ProjectLink, bool) {
	for _, l := range p.Links {
		if l.Name == name {
			return l, true
		}
	}
	return ProjectLink{}, false
}

// LinkByID returns the link to the project with the given ID.
func (p Project) LinkByID(id string) (ProjectLink, bool) {
	for _, l := range p.Links {
		if l.ID == id {
			return l, true
		}
	}
	return ProjectLink{}, false
}

// SplitRef splits a cross-project reference "<name>:<rest>" whose name is a
// valid link name. Issue IDs never contain ':', so a local reference never
// splits.
func SplitRef(s string) (name, rest string, ok bool) {
	name, rest, ok = strings.Cut(s, ":")
	if !ok || !ValidProjectName(name) || rest == "" {
		return "", "", false
	}
	return name, rest, true
}

// xLinkRe finds cross-project knowledge links in markdown: [text](name:/path.md#section).
var xLinkRe = regexp.MustCompile(`\]\(([a-z0-9][a-z0-9._-]*):(/[^)\s]*)\)`)

var (
	fencedRe   = regexp.MustCompile("(?ms)^\\s*(```|~~~).*?^\\s*(```|~~~)\\s*$")
	codeSpanRe = regexp.MustCompile("`[^`\n]*`")
)

// StripCode removes fenced blocks and inline code so examples are not
// mistaken for links.
func StripCode(s string) string {
	return codeSpanRe.ReplaceAllString(fencedRe.ReplaceAllString(s, ""), "")
}

// XKnowledgeLinks returns the cross-project knowledge links in text as
// name and target (/path.md, without a #section), in order, once each.
func XKnowledgeLinks(text string) [][2]string {
	var out [][2]string
	seen := map[string]bool{}
	for _, m := range xLinkRe.FindAllStringSubmatch(StripCode(text), -1) {
		target := m[2]
		if k := strings.IndexByte(target, '#'); k >= 0 {
			target = target[:k]
		}
		if !ValidProjectName(m[1]) || seen[m[1]+":"+target] {
			continue
		}
		seen[m[1]+":"+target] = true
		out = append(out, [2]string{m[1], target})
	}
	return out
}

// LinkRef is one reference to a link: from an issue (a dependency, or a
// knowledge link in one of its records) or from a knowledge entry.
type LinkRef struct {
	Link   string `json:"link"`
	Target string `json:"target"` // an issue ID or /path.md
	Issue  string `json:"issue,omitempty"`
	Entry  string `json:"entry,omitempty"`
	File   string `json:"file"`
	Dep    bool   `json:"dependency,omitempty"`
}

// LinkRefs lists every reference to a link in the tree, issues in ID order
// first, then knowledge entries by path.
func (t *Tree) LinkRefs() []LinkRef {
	var out []LinkRef
	for _, id := range t.IDs() {
		i := t.Issues[id]
		for _, d := range i.DependsOn {
			if name, rest, ok := SplitRef(d); ok {
				out = append(out, LinkRef{Link: name, Target: rest, Issue: id, File: "issue.md", Dep: true})
			}
		}
		texts := []struct{ file, text string }{{"issue.md", i.Body}, {"context.md", i.Context}, {"history.md", i.History}}
		if i.Findings != nil {
			texts = append(texts, struct{ file, text string }{"findings.md", *i.Findings})
		}
		for _, d := range i.Decisions {
			texts = append(texts, struct{ file, text string }{"decisions.md", d.Body})
		}
		for _, c := range i.Criteria {
			texts = append(texts, struct{ file, text string }{"acceptance.md", c.Text})
		}
		for _, x := range texts {
			for _, l := range XKnowledgeLinks(x.text) {
				out = append(out, LinkRef{Link: l[0], Target: l[1], Issue: id, File: x.file})
			}
		}
	}
	var paths []string
	for p := range t.Knowledge {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		for _, l := range XKnowledgeLinks(t.Knowledge[p].Body) {
			out = append(out, LinkRef{Link: l[0], Target: l[1], Entry: p, File: p})
		}
	}
	return out
}

// Where names the place a reference is written, for messages.
func (r LinkRef) Where() string {
	if r.Issue != "" {
		return r.Issue + " " + r.File
	}
	return r.File
}

// validateLinks checks project.md's links and the references to them.
func validateLinks(t *Tree, add func(Diagnostic)) {
	p := t.Project
	byID := map[string]string{}
	for _, l := range p.Links {
		bad := func(format string, a ...any) {
			add(Diagnostic{Code: CodeLinkInvalid, Severity: SevError, Class: ClassManual, File: "project.md",
				Message: fmt.Sprintf("link %s: ", l.Name) + fmt.Sprintf(format, a...), Fix: "prep link remove " + l.Name + ", then prep link add"})
		}
		switch {
		case !ValidProjectName(l.Name):
			bad("the name is invalid (use %s)", ProjectNameRule)
		case !ValidID(l.ID):
			bad("id %q is not a project ID", l.ID)
		case p.ID != "" && l.ID == p.ID:
			bad("it links this project to itself")
		}
		if other, ok := byID[l.ID]; ok && l.ID != "" {
			add(Diagnostic{Code: CodeLinkDuplicate, Severity: SevWarning, Class: ClassManual, File: "project.md",
				Message: fmt.Sprintf("links %s and %s name the same project", other, l.Name), Fix: "prep link rename the references of one onto the other, then prep link remove it"})
		}
		byID[l.ID] = l.Name
	}
	for _, r := range t.LinkRefs() {
		if _, ok := p.Link(r.Link); ok {
			continue
		}
		add(Diagnostic{Code: CodeLinkUndefined, Severity: SevError, Class: ClassManual, Issue: r.Issue, File: r.File,
			Message: fmt.Sprintf("%s:%s names no link %q in project.md", r.Link, r.Target, r.Link),
			Fix:     "prep link add <project> --as " + r.Link + ", or correct the reference"})
	}
}

// PlanLinks plans writing project.md's links as given (sorted by name). It
// refuses invalid names, invalid IDs, the project itself, and two links with
// one name or one ID.
func (t *Tree) PlanLinks(links []ProjectLink) (*Change, error) {
	names, ids := map[string]bool{}, map[string]string{}
	var unmet []Unmet
	for _, l := range links {
		switch {
		case !ValidProjectName(l.Name):
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("link name %q is invalid: use %s", l.Name, ProjectNameRule)})
		case !ValidID(l.ID):
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("link %s: %q is not a project ID", l.Name, l.ID)})
		case t.Project.ID != "" && l.ID == t.Project.ID:
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("link %s: that is this project", l.Name)})
		case names[l.Name]:
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("link %s exists already", l.Name)})
		case ids[l.ID] != "":
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("that project is linked already as %s", ids[l.ID])})
		}
		names[l.Name], ids[l.ID] = true, l.Name
	}
	if len(unmet) > 0 {
		return nil, &Error{Code: ErrUsage, Message: "the links are invalid", Unmet: unmet}
	}
	sorted := append([]ProjectLink{}, links...)
	sort.Slice(sorted, func(a, b int) bool { return sorted[a].Name < sorted[b].Name })
	return &Change{Op: OpLinks, Links: &sorted}, nil
}

// PlanLinkRename plans renaming the link old to new: project.md's entry and
// every reference to it in issues and knowledge entries.
func (t *Tree) PlanLinkRename(old, new string) (*Change, error) {
	if _, ok := t.Project.Link(old); !ok {
		return nil, &Error{Code: ErrNotFound, Message: fmt.Sprintf("no link %q; prep links lists them", old)}
	}
	links := make([]ProjectLink, 0, len(t.Project.Links))
	for _, l := range t.Project.Links {
		if l.Name == old {
			l.Name = new
		}
		links = append(links, l)
	}
	c, err := t.PlanLinks(links)
	if err != nil {
		return nil, err
	}
	c.LinkRename = &[2]string{old, new}
	return c, nil
}

// PlanLinkRemove plans removing the link name; it is refused while anything
// refers to it.
func (t *Tree) PlanLinkRemove(name string) (*Change, error) {
	if _, ok := t.Project.Link(name); !ok {
		return nil, &Error{Code: ErrNotFound, Message: fmt.Sprintf("no link %q; prep links lists them", name)}
	}
	var unmet []Unmet
	for _, r := range t.LinkRefs() {
		if r.Link == name {
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("%s refers to %s:%s", r.Where(), r.Link, r.Target)})
		}
	}
	if len(unmet) > 0 {
		return nil, &Error{Code: ErrGate, Message: fmt.Sprintf("link %s is still referred to; change or remove those references first", name), Unmet: unmet}
	}
	var links []ProjectLink
	for _, l := range t.Project.Links {
		if l.Name != name {
			links = append(links, l)
		}
	}
	return t.PlanLinks(links)
}

// depRefRe matches a dependency list item naming a link in issue.md's
// frontmatter, quoted or not.
func depRefRe(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?m)^(\s*-\s*['"]?)` + regexp.QuoteMeta(name) + `:([0-9A-Z]{` + fmt.Sprint(ULIDLen) + `}['"]?\s*)$`)
}

// RenameLinkInText rewrites references to the link old in a record's or an
// entry's raw text to new: knowledge links ](old:/…) and dependency items
// "- old:<ID>". Everything else stays byte for byte.
func RenameLinkInText(text, old, new string) string {
	text = strings.ReplaceAll(text, "]("+old+":/", "]("+new+":/")
	return depRefRe(old).ReplaceAllString(text, "${1}"+new+":${2}")
}

// renameRef renames a dependency reference.
func renameRef(d, old, new string) string {
	if name, rest, ok := SplitRef(d); ok && name == old {
		return new + ":" + rest
	}
	return d
}

// applyLinkRename renames references in a copy of an issue.
func applyLinkRename(i *Issue, old, new string) *Issue {
	cp := *i
	cp.DependsOn = nil
	for _, d := range i.DependsOn {
		cp.DependsOn = append(cp.DependsOn, renameRef(d, old, new))
	}
	r := func(s string) string { return RenameLinkInText(s, old, new) }
	cp.Body, cp.Context, cp.History = r(i.Body), r(i.Context), r(i.History)
	if i.Findings != nil {
		f := r(*i.Findings)
		cp.Findings = &f
	}
	cp.Decisions = nil
	for _, d := range i.Decisions {
		d.Body = r(d.Body)
		cp.Decisions = append(cp.Decisions, d)
	}
	cp.Criteria = nil
	for _, c := range i.Criteria {
		c.Text = r(c.Text)
		cp.Criteria = append(cp.Criteria, c)
	}
	return &cp
}

// XIssue is another project's issue a dependency names, as the caller
// found it through the project registry (Tree.External).
type XIssue struct {
	Ref               string            `json:"ref"` // <link>:<issue ID>
	Link              string            `json:"link"`
	ID                string            `json:"id"`
	Status            XStatus           `json:"status"`
	Outcome           DependencyOutcome `json:"outcome"`
	State             State             `json:"state,omitempty"`
	Title             string            `json:"title,omitempty"`
	SourceProjectID   string            `json:"source_project_id,omitempty"`
	Revision          string            `json:"revision,omitempty"`
	TargetDigest      string            `json:"target_digest,omitempty"`
	Uncertainty       string            `json:"uncertainty,omitempty"`
	ProvenanceWarning string            `json:"provenance_warning,omitempty"`
	// Project is the project's own name there; Root its checkout.
	Project string `json:"project,omitempty"`
	Root    string `json:"root,omitempty"`
	// Remote is the link's clone hint, for a project not checked out.
	Remote string `json:"remote,omitempty"`
}

// XStatus says what became of a cross-project reference on this machine.
type XStatus string

const (
	XFound   XStatus = "found"   // the issue exists in the checked-out project
	XMissing XStatus = "missing" // the project is checked out, the issue is not there
	XUnknown XStatus = "unknown" // the project is not checked out here (or the link is undefined)
)

// DependencyOutcome describes whether a dependency is met independently of
// whether its source checkout could be read. Unknown outcomes fail closed at
// claim time.
type DependencyOutcome string

const (
	DependencySatisfied   DependencyOutcome = "satisfied"
	DependencyUnsatisfied DependencyOutcome = "unsatisfied"
	DependencyUnknown     DependencyOutcome = "unknown"
)

// DependencyTargetDigest hashes only issue content needed to reconcile an
// external prerequisite; its caller adds the checkout revision.
func DependencyTargetDigest(i *Issue) string {
	if i == nil {
		return ""
	}
	return contentID(struct {
		ID, Title, Kind, Body, Parent string
		DependsOn                     []string
		Criteria                      []Criterion
		Baselines                     []Baseline
		Ready                         *Ready
		Claim                         *Claim
		Resolution                    *Resolution
	}{i.ID, i.Title, string(i.Kind), i.Body, i.Parent, i.DependsOn, i.Criteria, i.Baselines, i.Ready, i.Claim, i.Resolution})
}

// XRefs lists the cross-project dependencies in the tree, once each, in
// issue order.
func (t *Tree) XRefs() []string {
	var out []string
	seen := map[string]bool{}
	for _, id := range t.IDs() {
		for _, d := range t.Issues[id].DependsOn {
			if _, _, ok := SplitRef(d); ok && !seen[d] {
				seen[d] = true
				out = append(out, d)
			}
		}
	}
	return out
}

// XIssueOf returns what is known about a cross-project dependency; a
// reference nobody resolved is unknown.
func (t *Tree) XIssueOf(ref string) XIssue {
	if x, ok := t.External[ref]; ok {
		return x
	}
	name, id, _ := SplitRef(ref)
	x := XIssue{Ref: ref, Link: name, ID: id, Status: XUnknown, Outcome: DependencyUnknown}
	if l, ok := t.Project.Link(name); ok {
		x.Remote = l.Remote
	}
	return x
}

// DepState returns a dependency's state: a local issue's, or a found
// cross-project issue's. known is false for a cross-project issue that
// is missing or not checked out; those outcomes block unless overridden.
func (t *Tree) DepState(d string) (s State, known bool) {
	if _, _, ok := SplitRef(d); !ok {
		return t.State(d), true
	}
	if x := t.XIssueOf(d); x.Status == XFound {
		return x.State, true
	}
	return "", false
}

func contentID(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
