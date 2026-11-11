package domain

import "strings"

// The hand-off report is the text an implementer or checker returns after
// working from a brief. Its format is defined here once: the brief prints it
// (ReportFormat) and the hand-back parser reads it by the same names, so the
// two cannot drift.

// Report section names, written as Markdown headings (ReportHeading).
const (
	ReportSummary   = "Summary"
	ReportCriteria  = "Criteria"
	ReportDecisions = "Decisions"
	ReportFindings  = "Findings"
	ReportProblems  = "Problems"
)

// ReportSections lists every section a report may hold, in order.
var ReportSections = []string{ReportSummary, ReportCriteria, ReportDecisions, ReportFindings, ReportProblems}

const (
	// ReportHeadingPrefix starts a section's heading line.
	ReportHeadingPrefix = "## "
	// ReportItemPrefix starts the title line of one decision or finding.
	ReportItemPrefix = "### "
	// ReportScopePrefix starts the optional scope line of a decision or finding.
	ReportScopePrefix = "Scope: "
	// ReportMet and ReportNotMet are the verdicts of a criteria line, which
	// reads "<number>. <verdict> — <evidence>".
	ReportMet    = "met"
	ReportNotMet = "not met"
	// ReportVerdictSep separates a criteria line's verdict from its evidence.
	ReportVerdictSep = " — "
)

// ReportHeading is the heading line of a report section.
func ReportHeading(section string) string { return ReportHeadingPrefix + section }

// reportSectionHelp says what goes under each section.
var reportSectionHelp = map[string]string{
	ReportSummary:   "One or two sentences on what you did and where the change is (branch or directory).",
	ReportCriteria:  "One line per acceptance criterion, by its number: `<number>. " + ReportMet + ReportVerdictSep + "<evidence>` or `<number>. " + ReportNotMet + ReportVerdictSep + "<what is missing>`. Evidence is a test name, a file and line, or a command and its result.",
	ReportDecisions: "Each decision you made that someone changing this code later should know: a line `" + ReportItemPrefix + "<title>`, then the reason below it. Optionally a line `" + ReportScopePrefix + "<path or glob>` naming the code it constrains.",
	ReportFindings:  "Each thing you found out that is worth keeping: a line `" + ReportItemPrefix + "<title>`, then what you found and how you know.",
	ReportProblems:  "Anything that stopped you, conflicts with the brief, or is left open.",
}

// ReportSectionsFor lists the sections asked of a role: an implementer of a
// research issue reports findings instead of a criteria verdict, and a
// checker reports a verdict and problems only.
func ReportSectionsFor(kind Kind, check, hasCriteria bool) []string {
	switch {
	case check && (kind != KindResearch || hasCriteria):
		return []string{ReportSummary, ReportCriteria, ReportProblems}
	case check:
		return []string{ReportSummary, ReportProblems}
	case kind == KindResearch:
		return []string{ReportSummary, ReportFindings, ReportDecisions, ReportProblems}
	}
	return []string{ReportSummary, ReportCriteria, ReportDecisions, ReportFindings, ReportProblems}
}

// ReportFormat describes the report for the given sections, as the brief
// prints it.
func ReportFormat(sections []string) string {
	var b strings.Builder
	b.WriteString("Write the report as Markdown with these headings, in this order. Leave out a section that would be empty, except " + ReportHeading(sections[0]) + ".\n")
	for _, s := range sections {
		b.WriteString("\n" + ReportHeading(s) + "\n" + reportSectionHelp[s] + "\n")
	}
	return b.String()
}
