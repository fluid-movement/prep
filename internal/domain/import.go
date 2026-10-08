package domain

import (
	"io"
	"strings"
	"time"
)

// ImportTag marks the issue that imports existing work items.
const ImportTag = "import"

// ImportOpen returns the unresolved import issue, empty when there is none.
func (t *Tree) ImportOpen() string {
	for _, id := range t.IDs() {
		if contains(t.Issues[id].Tags, ImportTag) && !t.State(id).Terminal() {
			return id
		}
	}
	return ""
}

const importRequirement = `Bring the work this project already tracks elsewhere into prep, once. The sources are whatever the project uses: TODO and FIXME comments, GitHub issues, other trackers, task lists in documents. The findings list the candidates found, the user reviews them, and each approved candidate becomes an open issue that names its source. Candidates already imported are skipped, so the import can run again. Afterwards prep is the source of truth for this work; nothing is kept in sync with the sources.`

const importContext = `Follow these steps; each prep command validates before writing.

1. Ask the user which sources hold the project's work, and look for the usual ones yourself: TODO, FIXME and similar comments in the code (search the repository), GitHub issues (gh issue list --state open --limit 200 --json number,title,url,labels,body, when the project is on GitHub), task lists and roadmaps in the documentation, exports or APIs of other trackers the user names. Use only the sources the user confirms. With many sources or a large codebase, search each with a subagent that returns only the candidate lines.
2. Write the candidates as findings: prep findings <this issue> --body-file -. Group them by source; one line per candidate with a title, a kind (code, manual, research or decision) and its source as a link or a path:line, such as https://github.com/owner/repo/issues/12 or internal/export/csv.go:48. Note a tag or priority when the source carries one. Ask the user to review the list and strike what should not come in; update the findings until they agree.
3. For each approved candidate, look for an earlier import of it: prep list --text "Source: <source>". When that finds an issue whose Source: line is exactly this source (a path:line can be the prefix of a longer one), skip the candidate. Otherwise create it: prep new --title "<title>" --kind <kind> [--tag <tag>] [--priority <level>] --body-file -, with a requirement that states the work in a few sentences, puts what the source leaves unresolved under ## Open questions, and ends in the line Source: <source>. Leave the new issues open: they go through prep define like any other issue.
4. Add to the findings which issues were created and which candidates were skipped, then complete this issue: prep complete <this issue> --no-impact "imported work items; no knowledge changed".

Nothing is kept in sync: after the import, change the work in prep, and close or annotate the sources however the project prefers.`

// PlanImport returns the changes that create the import issue: a research
// issue tagged import, with a context that guides the agent through the
// import and its criteria, left open for the user to define. It refuses
// while an import issue is unresolved. entropy feeds the issue ID (nil
// means crypto/rand).
func (t *Tree) PlanImport(actor string, now time.Time, entropy io.Reader) ([]*Change, error) {
	if id := t.ImportOpen(); id != "" {
		return nil, &Error{Code: ErrInvalid, Message: "an import issue is unresolved: prep guide " + id}
	}
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
	issue, err := step(cur.PlanNew(NewIssueInput{Title: "Import existing work", Kind: KindResearch, Tags: []string{ImportTag}, Body: importRequirement, Entropy: entropy}, now))
	if err != nil {
		return nil, err
	}
	id := issue.IssueID
	ctx := strings.ReplaceAll(importContext, "<this issue>", id)
	if _, err := step(cur.PlanRecord(id, OpContext, RecordInput{Actor: actor, Now: now, Text: ctx})); err != nil {
		return nil, err
	}
	if _, err := step(cur.PlanRecord(id, OpCriterion, RecordInput{Actor: actor, Now: now, Acceptance: []AcceptanceOp{
		{Op: "add", Text: "The user named the sources, and the candidates from them are written as findings the user reviewed"},
		{Op: "add", Text: "Each approved candidate is an open issue whose requirement ends in a Source: line, or was skipped because an issue already names its source"},
		{Op: "add", Text: "The findings list the created issues and the skipped candidates"},
	}})); err != nil {
		return nil, err
	}
	return out, nil
}
