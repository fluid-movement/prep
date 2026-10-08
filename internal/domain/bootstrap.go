package domain

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// BootstrapTag marks the issues that bootstrap the knowledge base.
const BootstrapTag = "bootstrap"

// GateBootstrap blocks knowledge entries before the overview exists.
const GateBootstrap = "G_BOOTSTRAP"

// Bootstrapped reports whether the knowledge base has been bootstrapped:
// its overview entry exists.
func (t *Tree) Bootstrapped() bool { return t.Knowledge[OverviewEntry] != nil }

// BootstrapStart returns the first unresolved bootstrap-tagged issue without
// unresolved bootstrap children: the survey while it is open, then the
// areas. Empty when there is none.
func (t *Tree) BootstrapStart() string {
	for _, id := range t.TreeOrder(t.IDs()) {
		i := t.Issues[id]
		if t.State(id).Terminal() || !contains(i.Tags, BootstrapTag) {
			continue
		}
		open := false
		for _, c := range t.Children(id) {
			if !t.State(c).Terminal() {
				open = true
			}
		}
		if !open {
			return id
		}
	}
	return ""
}

// BootstrapAlert is the message prime and guide show while the knowledge
// base is not bootstrapped; empty once it is.
func (t *Tree) BootstrapAlert() string {
	if t.Bootstrapped() {
		return ""
	}
	if s := t.BootstrapStart(); s != "" {
		return fmt.Sprintf("The knowledge base is not bootstrapped yet (no %s). Continue the bootstrap: prep guide %s.", OverviewEntry, s)
	}
	return fmt.Sprintf("The knowledge base is not bootstrapped yet (no %s). Create the bootstrap issues with prep knowledge bootstrap.", OverviewEntry)
}

const bootstrapParent = `Map this project into the knowledge base so agents find the context they need for every issue. The survey child produces the map and the overview entry, which marks the knowledge base bootstrapped; each area of the map then becomes another child that writes its entries as drafts. The knowledge base is complete for now when every area is written and the user has reviewed the drafts.`

const bootstrapSurvey = `Survey the project: what it is, how it is built and run, and which knowledge entries agents need. The findings are the map: a short description of the project and the proposed entries, each with type, path, title, one-line description and scope, grouped into areas small enough for one agent session. Sources are the code and the existing documentation, such as the README, architecture or decision records, design documents, build and CI configuration. In a repository without code the map describes the project as it is meant to become.`

const bootstrapContext = `Follow these steps; each prep command validates before writing.

1. Read the existing documentation and configuration first (README, docs/, ADRs, design documents, build files, CI), then the code's top-level structure. Note what is already explained well and where it lives. In a large project, give each area to a subagent that returns a short summary, rather than reading it all yourself.
2. Write the map as findings: prep findings <this issue> --body-file -. Start with two or three paragraphs on what the project is, how it is organized and how it is built and tested. Then list the proposed entries as a table: type (overview, feature, component, decision, convention, pitfall), path such as /components/export.md, title, one-line description, and scope (paths or globs the entry describes). Group them into areas of three to eight entries.
3. Ask the user to review the map and adjust it until they agree. This is the important review; entries are cheap to write once the map is right.
4. Write the overview: prep knowledge new /overview.md --type overview --title "<project>" --description "<one line>" --status draft --body-file -. It describes the project in a few paragraphs; links to entries are added as the entries are written. Writing it marks the knowledge base bootstrapped.
5. Create one issue per area under the bootstrap parent: prep new --parent <parent> --kind code --tag bootstrap --title "Document <area>" --body-file -, naming the entries from the map. Each writes its entries with prep knowledge new --status draft --scope ..., links them from the overview with prep knowledge update, confirms them with prep knowledge confirm, and completes with the commit that added them.
6. Complete this issue: prep complete <this issue> --docs /overview.md.

Conventions for entries: one concept per entry, current state only, bundle-relative links such as /components/export.md, and descriptions that let an agent decide whether to open the entry.`

// PlanBootstrap returns the changes that create the bootstrap issues: a
// parent and its survey child, both tagged and left open for the user to
// define. Each change is planned on the tree with the previous ones
// applied; the caller writes them in order. entropy feeds the issue IDs
// (nil means crypto/rand).
func (t *Tree) PlanBootstrap(actor string, now time.Time, entropy io.Reader) ([]*Change, error) {
	var out []*Change
	cur := t
	step := func(c *Change, err error) (*Change, error) {
		if err != nil {
			return nil, err
		}
		out = append(out, c)
		cur = cur.Apply(c)
		return c, nil
	}
	parent, err := step(cur.PlanNew(NewIssueInput{Title: "Bootstrap the knowledge base", Kind: KindManual, Tags: []string{BootstrapTag}, Body: bootstrapParent, Entropy: entropy}, now))
	if err != nil {
		return nil, err
	}
	pid := parent.IssueID
	if _, err := step(cur.PlanRecord(pid, OpCriterion, RecordInput{Actor: actor, Now: now, Acceptance: []AcceptanceOp{
		{Op: "add", Text: "Every area of the map is written as entries"},
		{Op: "add", Text: "The user reviewed the drafts and the reviewed entries are stable"},
	}})); err != nil {
		return nil, err
	}
	survey, err := step(cur.PlanNew(NewIssueInput{Title: "Survey the project into a knowledge map", Kind: KindResearch, Parent: pid, Tags: []string{BootstrapTag}, Body: bootstrapSurvey, Entropy: entropy}, now.Add(time.Second)))
	if err != nil {
		return nil, err
	}
	sid := survey.IssueID
	ctx := strings.NewReplacer("<this issue>", sid, "<parent>", pid).Replace(bootstrapContext)
	if _, err := step(cur.PlanRecord(sid, OpContext, RecordInput{Actor: actor, Now: now, Text: ctx})); err != nil {
		return nil, err
	}
	if _, err := step(cur.PlanRecord(sid, OpCriterion, RecordInput{Actor: actor, Now: now, Acceptance: []AcceptanceOp{
		{Op: "add", Text: "The map is written as findings and the user reviewed it"},
		{Op: "add", Text: "/overview.md exists as a draft"},
		{Op: "add", Text: "Each area of the map has a child issue under the bootstrap parent"},
	}})); err != nil {
		return nil, err
	}
	return out, nil
}
