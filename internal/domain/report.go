package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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
	ReportDecisions: "Each decision you made that someone changing this code later should know: a line `" + ReportItemPrefix + "<title>`, then the reason below it. Optionally a line `" + ReportScopePrefix + "<glob>[, <glob>]` directly under the title naming the code it constrains.",
	ReportFindings:  "Each thing you found out that is worth keeping: a line `" + ReportItemPrefix + "<title>`, then what you found and how you know.",
	ReportProblems:  "Anything that stopped you, conflicts with the brief, or is left open.",
}

// ReportSectionsFor lists the sections asked of a role, for every issue kind.
// Checkers and research implementers report verdicts, even with no criteria.
func ReportSectionsFor(kind Kind, check, _ bool) []string {
	if check {
		return []string{ReportSummary, ReportCriteria, ReportProblems}
	}
	if kind == KindResearch {
		return []string{ReportSummary, ReportCriteria, ReportDecisions, ReportFindings, ReportProblems}
	}
	return []string{ReportSummary, ReportDecisions, ReportFindings, ReportProblems}
}

// ReportFormat describes the report for the given sections, as the brief
// prints it.
func ReportFormat(sections []string) string {
	var b strings.Builder
	required := ReportHeading(ReportSummary)
	for _, s := range sections {
		if s == ReportCriteria {
			required += " and " + ReportHeading(ReportCriteria)
		}
	}
	b.WriteString("Write the report as Markdown with these headings, in this order. Leave out a section that would be empty, except " + required + ".\n")
	for _, s := range sections {
		b.WriteString("\n" + ReportHeading(s) + "\n" + reportSectionHelp[s] + "\n")
	}
	return b.String()
}

// Report is a fully validated hand-back, before any records are planned.
type Report struct {
	Summary   string
	Criteria  []ReportCriterion
	Decisions []ReportEntry
	Findings  []ReportEntry
	Problems  string
}

type ReportCriterion struct {
	Number   int
	Met      bool
	Evidence string
}

type ReportEntry struct {
	Title, Description, Body string
	Scope                    []string
}

var reportHeading = regexp.MustCompile(`^#{1,2}\s+(.+?)\s*#*\s*$`)
var reportCriterion = regexp.MustCompile(`^\s*[-*]?\s*(\d+)\s*[.)\]:-]?\s*(.*)$`)
var reportVerdict = regexp.MustCompile(`(?i)^\W*(not\s+met|met)\b(.*)$`)

// ParseReport accepts cosmetic differences, but refuses content that could
// write a wrong record. Empty optional sections may be omitted, as in a brief.
func ParseReport(text string, issue *Issue, check bool) (*Report, error) {
	fault := func(format string, args ...any) error {
		return &Error{Code: ErrUsage, Message: "report: " + fmt.Sprintf(format, args...)}
	}
	sections := map[string]string{}
	section := ""
	known := false
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if m := reportHeading.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			section = ""
			for _, name := range ReportSections {
				if strings.EqualFold(m[1], name) {
					section, known = name, true
					if _, exists := sections[name]; !exists {
						sections[name] = ""
					}
					break
				}
			}
			continue
		}
		if section != "" {
			sections[section] += line + "\n"
		}
	}
	if !known {
		return nil, fault("no known heading (expected ## Summary)")
	}
	r := &Report{Summary: strings.TrimSpace(sections[ReportSummary]), Problems: strings.TrimSpace(sections[ReportProblems])}
	if r.Summary == "" {
		return nil, fault("missing Summary section or summary text")
	}
	switch strings.ToLower(strings.Trim(r.Problems, " .\n\t-*")) {
	case "none", "n/a", "no problems":
		r.Problems = ""
	}
	seen := map[int]bool{}
	needsCriteria := check || issue.Kind == KindResearch
	if !needsCriteria && strings.TrimSpace(sections[ReportCriteria]) != "" {
		return nil, fault("implementer report cannot contain Criteria verdicts; use --check for a checker's report")
	}
	for _, line := range strings.Split(strings.TrimSpace(sections[ReportCriteria]), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		m := reportCriterion.FindStringSubmatch(line)
		if m == nil {
			return nil, fault("Criteria: expected a numbered verdict, got %q", line)
		}
		n, _ := strconv.Atoi(m[1])
		if n < 1 || n > len(issue.Criteria) {
			return nil, fault("Criteria: no criterion %d (issue has %d)", n, len(issue.Criteria))
		}
		v := reportVerdict.FindStringSubmatch(m[2])
		if v == nil {
			return nil, fault("criterion %d: missing met/not met verdict", n)
		}
		if seen[n] {
			return nil, fault("criterion %d: duplicate verdict", n)
		}
		seen[n] = true
		r.Criteria = append(r.Criteria, ReportCriterion{Number: n, Met: strings.EqualFold(v[1], ReportMet), Evidence: strings.TrimSpace(strings.TrimLeft(v[2], " *_:;.—–-)\t"))})
	}
	if needsCriteria {
		if _, exists := sections[ReportCriteria]; !exists && len(issue.Criteria) == 0 {
			return nil, fault("missing Criteria section for checker or research report")
		}
		for n := range issue.Criteria {
			if !seen[n+1] {
				return nil, fault("Criteria: missing verdict for criterion %d", n+1)
			}
		}
	}
	if check {
		if strings.TrimSpace(sections[ReportDecisions])+strings.TrimSpace(sections[ReportFindings]) != "" {
			return nil, fault("checker report cannot contain Decisions or Findings; omit --check for an implementer's report")
		}
	}
	for _, name := range []string{ReportDecisions, ReportFindings} {
		entries, err := parseReportEntries(sections[name], name)
		if err != nil {
			return nil, err
		}
		if name == ReportDecisions {
			r.Decisions = entries
		} else {
			r.Findings = entries
		}
	}
	if !check && issue.Kind == KindResearch && len(r.Findings) == 0 {
		return nil, fault("research report needs a Findings section with a finding")
	}
	return r, nil
}

func parseReportEntries(text, section string) ([]ReportEntry, error) {
	var entries []ReportEntry
	var lines []string
	title := ""
	flush := func() error {
		if title == "" && strings.TrimSpace(strings.Join(lines, "\n")) == "" {
			return nil
		}
		if title == "" {
			return &Error{Code: ErrUsage, Message: "report: " + section + ": missing ### title"}
		}
		e := ReportEntry{Title: title}
		body := strings.TrimSpace(strings.Join(lines, "\n"))
		first, rest, _ := strings.Cut(body, "\n")
		if strings.HasPrefix(strings.ToLower(first), "scope:") {
			for _, scope := range strings.Split(first[len("scope:"):], ",") {
				scope = strings.TrimSpace(scope)
				if scope == "" {
					return &Error{Code: ErrUsage, Message: "report: " + section + ": empty Scope for " + title}
				}
				e.Scope = append(e.Scope, scope)
			}
			body = strings.TrimSpace(rest)
		}
		if body == "" {
			return &Error{Code: ErrUsage, Message: "report: " + section + ": " + title + " has no reason or finding text"}
		}
		// A sentence ends at punctuation followed by whitespace (not at a
		// dot inside a code path). The remainder preserves paragraphs.
		cut := len(body)
		for n := 0; n+1 < len(body); n++ {
			if strings.ContainsRune(".!?", rune(body[n])) && strings.ContainsRune(" \t\n\r", rune(body[n+1])) {
				cut = n + 1
				break
			}
		}
		e.Description = strings.Join(strings.Fields(body[:cut]), " ")
		e.Body = strings.TrimSpace(body[cut:])
		entries = append(entries, e)
		return nil
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), ReportItemPrefix) {
			if err := flush(); err != nil {
				return nil, err
			}
			title, lines = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ReportItemPrefix)), nil
			if title == "" {
				return nil, &Error{Code: ErrUsage, Message: "report: " + section + ": empty title"}
			}
		} else {
			lines = append(lines, line)
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return entries, nil
}
