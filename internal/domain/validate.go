package domain

import (
	"fmt"
	"strings"
)

// KnowledgeSizeLimit is the entry size (bytes) above which lint warns.
const KnowledgeSizeLimit = 8 * 1024

// Validate checks the whole tree: relations, records and the knowledge base.
// Format-level diagnostics from the adapters are included.
func Validate(t *Tree) []Diagnostic {
	ds := append([]Diagnostic(nil), t.ParseDiags...)
	add := func(d Diagnostic) { ds = append(ds, d) }

	if t.Project.Schema != SchemaVersion && t.Project.Schema != 0 {
		add(Diagnostic{Code: CodeSchemaMismatch, Severity: SevError, Class: ClassGuided, File: "project.md",
			Message: fmt.Sprintf("schema %d, this binary uses %d", t.Project.Schema, SchemaVersion),
			Fix:     "run prep migrate, or update prep if the project is newer"})
	}

	for _, id := range t.IDs() {
		i := t.Issues[id]
		e := func(code string, class Class, file, fix, format string, a ...any) {
			add(Diagnostic{Code: code, Severity: SevError, Class: class, Issue: id, File: file, Message: fmt.Sprintf(format, a...), Fix: fix})
		}
		w := func(code string, class Class, file, fix, format string, a ...any) {
			add(Diagnostic{Code: code, Severity: SevWarning, Class: class, Issue: id, File: file, Message: fmt.Sprintf(format, a...), Fix: fix})
		}

		if strings.TrimSpace(i.Title) == "" {
			e(CodeTitleMissing, ClassManual, "issue.md", "add a title to the frontmatter", "title is missing")
		}
		if !i.Kind.Valid() {
			e(CodeKindInvalid, ClassManual, "issue.md", "set kind to one of "+joinKinds(), "kind %q is invalid", i.Kind)
		}

		for _, tag := range i.Tags {
			if !ValidTag(tag) {
				e(CodeTagInvalid, ClassManual, "issue.md", "use lowercase letters, digits and . _ - / (prep edit --tag)", "tag %q is malformed", tag)
			}
		}

		// Parent relation.
		if i.Parent != "" {
			if t.Issues[i.Parent] == nil {
				e(CodeParentMissing, ClassManual, "issue.md", "fix or remove the parent field", "parent %s does not exist", i.Parent)
			} else if inParentCycle(t, id) {
				e(CodeParentCycle, ClassManual, "issue.md", "change a parent field so the chain ends at a top-level issue", "parent chain contains a cycle")
			}
		}

		// Dependencies.
		seen := map[string]bool{}
		anc := map[string]bool{}
		for _, a := range t.Ancestors(id) {
			anc[a] = true
		}
		for _, d := range i.DependsOn {
			if seen[d] {
				w(CodeDuplicateDep, ClassFixable, "issue.md", "prep fix removes the duplicate", "dependency %s listed twice", d)
				continue
			}
			seen[d] = true
			switch {
			case d == id:
				e(CodeSelfDependency, ClassManual, "issue.md", "remove the issue's own ID from depends_on", "issue depends on itself")
			case t.Issues[d] == nil:
				e(CodeDepMissing, ClassManual, "issue.md", "fix or remove the dependency", "dependency %s does not exist", d)
			case anc[d]:
				e(CodeDepOnAncestor, ClassManual, "issue.md", "remove the dependency; a parent completes only after its children", "depends on its own ancestor %s (deadlock)", d)
			case t.State(d) == StateDropped && !t.State(id).Terminal():
				w(CodeDepDropped, ClassGuided, "issue.md", "remove the dependency or replace it with the issue that supersedes it", "depends on dropped issue %s", d)
			}
		}
		if inDepCycle(t, id) {
			e(CodeDepCycle, ClassManual, "issue.md", "remove a depends_on entry to break the cycle", "dependency cycle through this issue")
		}

		// Kind-specific files.
		if i.HasFile("findings.md") && i.Kind != KindResearch && i.Kind.Valid() {
			e(CodeFindingsKind, ClassGuided, "findings.md", "change kind to research, or move the findings into context.md or history.md", "findings.md is only allowed on research issues")
		}

		// Records.
		if i.Ready != nil && !t.ReadyValid(i) && i.Resolution == nil {
			w(CodeReadyOutdated, ClassGuided, "ready.md", "enrich against the current requirement and run prep ready, or prep ack if the change was trivial", "ready.md references baseline %s, not the newest", i.Ready.Baseline)
		}
		if i.Claim != nil && !t.ReadyValid(i) && i.Resolution == nil {
			e(CodeOrphanClaim, ClassGuided, "claim.md", "delete claim.md, or restore a valid ready sign-off", "claim.md exists without a valid ready sign-off")
		}
		if i.Ready != nil {
			found := false
			for _, b := range i.Baselines {
				if b.Name == i.Ready.Baseline {
					found = true
				}
			}
			if !found {
				e(CodeRecordInvalid, ClassManual, "ready.md", "point baseline at an existing baselines/ file", "ready.md references unknown baseline %q", i.Ready.Baseline)
			}
		}
		if r := i.Resolution; r != nil {
			switch r.Outcome {
			case OutcomeDone:
				if r.Documentation == nil || (len(r.Documentation.Entries) == 0 && strings.TrimSpace(r.Documentation.NoImpact) == "") {
					e(CodeRecordInvalid, ClassManual, "resolution.md", "add documentation entries or a no_impact reason", "resolution has no documentation decision")
				}
				if r.Documentation != nil {
					for _, p := range r.Documentation.Entries {
						if t.Knowledge[p] == nil {
							e(CodeDocEntryMissing, ClassManual, "resolution.md", "restore the entry or fix the reference", "documentation entry %s does not exist", p)
						}
					}
				}
				if t.IsParent(id) {
					for _, c := range t.Children(id) {
						if !t.State(c).Terminal() {
							w(CodeResolvedChildOpen, ClassGuided, "resolution.md", "complete or drop the child, or move it to another parent", "done parent has unresolved child %s", c)
						}
					}
				}
			case OutcomeDropped:
				if strings.TrimSpace(r.Reason) == "" {
					e(CodeRecordInvalid, ClassManual, "resolution.md", "add a reason", "dropped resolution has no reason")
				}
			default:
				e(CodeRecordInvalid, ClassManual, "resolution.md", "set outcome to done or dropped", "invalid outcome %q", r.Outcome)
			}
		}
		if t.State(id) == StateInProgress && t.Drifted(i) {
			w(CodeStaleInProgress, ClassGuided, "issue.md", "prep ack for a trivial change; otherwise prep release, then prep define and re-enrich", "requirement changed while in progress")
		}

		// Decisions.
		ids := map[string]bool{}
		for _, d := range i.Decisions {
			if ids[d.ID] {
				e(CodeDecisionInvalid, ClassManual, "decisions.md", "give each entry a unique ID", "duplicate decision id %s", d.ID)
			}
			if d.Supersedes != "" && !ids[d.Supersedes] {
				e(CodeDecisionInvalid, ClassManual, "decisions.md", "supersedes must name an earlier entry", "decision %s supersedes unknown or later entry %s", d.ID, d.Supersedes)
			}
			ids[d.ID] = true
		}

		// Context links into the knowledge base.
		for _, l := range i.ContextLinks {
			if t.Knowledge[l] == nil {
				w(CodeContextLinkMissing, ClassManual, "context.md", "fix the link or create the entry", "linked knowledge entry %s does not exist", l)
			}
		}
	}

	// Opt-outs that match no inherited item.
	for _, id := range t.IDs() {
		i := t.Issues[id]
		if len(i.DoDOptOuts) == 0 {
			continue
		}
		inherited := map[string]bool{}
		for _, it := range t.Project.DoD {
			inherited[it] = true
		}
		for _, a := range t.Ancestors(id) {
			if ai := t.Issues[a]; ai != nil {
				for _, it := range ai.DoDAdd {
					inherited[it] = true
				}
			}
		}
		for _, o := range i.DoDOptOuts {
			if !inherited[o.Item] {
				add(Diagnostic{Code: CodeDoDOptOutUnused, Severity: SevWarning, Class: ClassManual, Issue: id, File: "acceptance.md",
					Message: fmt.Sprintf("opt-out %q matches no inherited Definition of Done item", o.Item), Fix: "use the exact item text, or remove the opt-out"})
			}
		}
	}

	ds = append(ds, validateKnowledge(t)...)
	SortDiagnostics(ds)
	return ds
}

func validateKnowledge(t *Tree) []Diagnostic {
	var ds []Diagnostic
	for _, p := range sortedKeys(t.Knowledge) {
		k := t.Knowledge[p]
		file := ".prep/knowledge" + p
		e := func(code string, sev Severity, class Class, fix, format string, a ...any) {
			ds = append(ds, Diagnostic{Code: code, Severity: sev, Class: class, File: file, Message: fmt.Sprintf(format, a...), Fix: fix})
		}
		if k.Type == "" {
			e(CodeKnowledgeField, SevError, ClassManual, "add type: one of "+strings.Join(KnowledgeTypes, ", "), "type is missing")
		} else if !contains(KnowledgeTypes, k.Type) {
			e(CodeKnowledgeField, SevError, ClassManual, "use one of "+strings.Join(KnowledgeTypes, ", "), "type %q is not allowed", k.Type)
		}
		if strings.TrimSpace(k.Title) == "" {
			e(CodeKnowledgeField, SevError, ClassManual, "add a title", "title is missing")
		}
		if strings.TrimSpace(k.Description) == "" {
			e(CodeKnowledgeField, SevError, ClassManual, "add a one-line description; prep guide lists it for triage", "description is missing")
		}
		if k.Status != "" && k.Status != "draft" && k.Status != "stable" {
			e(CodeKnowledgeField, SevError, ClassManual, "use draft or stable; stale entries are deleted, not deprecated", "status %q is not allowed", k.Status)
		}
		for _, l := range k.Links {
			if t.Knowledge[l] == nil {
				e(CodeKnowledgeLink, SevError, ClassManual, "fix the link; entries are never renamed", "broken link to %s", l)
			}
		}
		if k.Size > KnowledgeSizeLimit {
			e(CodeKnowledgeSize, SevWarning, ClassManual, "split the entry: one concept per entry", "entry is %d bytes, above %d", k.Size, KnowledgeSizeLimit)
		}
		if k.DriftErr != "" {
			e(CodeKnowledgeCommit, SevWarning, ClassManual, "set confirmed_commit to a commit in this repository", "%s", k.DriftErr)
		}
		if len(k.Drifted) > 0 {
			e(CodeKnowledgeDrift, SevWarning, ClassGuided, "re-check the entry against the code, update it and confirmed_commit", "scoped paths changed since %s: %s", k.ConfirmedCommit, strings.Join(k.Drifted, ", "))
		}
	}
	return ds
}

func inParentCycle(t *Tree, id string) bool {
	seen := map[string]bool{}
	cur := id
	for {
		i := t.Issues[cur]
		if i == nil || i.Parent == "" {
			return false
		}
		if i.Parent == id {
			return true
		}
		if seen[i.Parent] {
			return false // cycle exists above, reported on its members
		}
		seen[i.Parent] = true
		cur = i.Parent
	}
}

func inDepCycle(t *Tree, id string) bool {
	seen := map[string]bool{}
	var visit func(string) bool
	visit = func(cur string) bool {
		i := t.Issues[cur]
		if i == nil {
			return false
		}
		for _, d := range i.DependsOn {
			if d == id && cur != id {
				return true
			}
			if d == id {
				continue // self-dependency is reported separately
			}
			if !seen[d] {
				seen[d] = true
				if visit(d) {
					return true
				}
			}
		}
		return false
	}
	return visit(id)
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
