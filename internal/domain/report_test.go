package domain

import (
	"strings"
	"testing"
)

func TestParseReport(t *testing.T) {
	issue := &Issue{Kind: KindCode, Criteria: []Criterion{{Text: "one"}, {Text: "two"}}}
	for _, tt := range []struct {
		name, text, fault string
		check             bool
	}{
		{"implementer", "## Summary\nBuilt it.\n## Decisions\n### Stream\nScope: internal/**, cmd/**\nMemory stays flat. Buffers grow too large.\n## Findings\n### Library\nIt works. Tested locally.", "", false},
		{"checker", "## Summary\nChecked.\n## Criteria\n1. met — tests\n2. not met — broken", "", true},
		{"cosmetic", "# Report\n## sUMMARY\n\nChecked.\n\n## CRITERIA\n1) **MET**: tests\n\n2: (not met), broken\n## Extra\nIgnore this.\n## Problems\nNone.", "", true},
		{"unknown sections", "## Notes\nIgnored.\n## Summary\nBuilt it.\n## Extra\n### Not a decision\nSome text.", "", false},
		{"no heading", "Built it.", "no known heading", false},
		{"unknown only", "## Notes\nBuilt it.", "no known heading", false},
		{"missing summary", "## Problems\nBroken.", "missing Summary", false},
		{"empty summary", "## Summary\n\n## Criteria\n1. met", "missing Summary", false},
		{"unknown number", "## Summary\nChecked.\n## Criteria\n3. met", "no criterion 3", true},
		{"zero", "## Summary\nChecked.\n## Criteria\n0. met", "no criterion 0", true},
		{"no verdict", "## Summary\nChecked.\n## Criteria\n1. tests pass", "criterion 1: missing", true},
		{"missing criteria", "## Summary\nChecked.", "missing verdict for criterion 1", true},
		{"missing one verdict", "## Summary\nChecked.\n## Criteria\n1. met", "missing verdict for criterion 2", true},
		{"duplicate", "## Summary\nChecked.\n## Criteria\n1. met\n1. not met", "duplicate verdict", true},
		{"no reason", "## Summary\nBuilt.\n## Decisions\n### Stream", "has no reason", false},
		{"scope not reason", "## Summary\nBuilt.\n## Decisions\n### Stream\nScope: internal/**", "has no reason", false},
		{"empty scope", "## Summary\nBuilt.\n## Decisions\n### Stream\nScope: \nMemory.", "empty Scope", false},
		{"no title", "## Summary\nBuilt.\n## Decisions\nUse streaming because memory.", "missing ### title", false},
		{"checker role", "## Summary\nChecked.\n## Criteria\n1. met\n2. met\n## Decisions\n### Stream\nMemory.", "checker report cannot", true},
		{"implementer role", "## Summary\nBuilt.\n## Criteria\n1. met", "implementer report cannot", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseReport(tt.text, issue, tt.check)
			if tt.fault != "" {
				if err == nil || !strings.Contains(err.Error(), tt.fault) {
					t.Fatalf("got %v, want %q", err, tt.fault)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.name == "implementer" {
				d := r.Decisions[0]
				if d.Title != "Stream" || d.Description != "Memory stays flat." || d.Body != "Buffers grow too large." || strings.Join(d.Scope, ",") != "internal/**,cmd/**" {
					t.Fatalf("decision: %+v", d)
				}
			}
			if tt.check && (len(r.Criteria) != 2 || !r.Criteria[0].Met || r.Criteria[1].Met) {
				t.Fatalf("verdicts: %+v", r.Criteria)
			}
		})
	}
}

func TestReportSectionsForRoles(t *testing.T) {
	for _, kind := range []Kind{KindCode, KindResearch, KindManual, KindDecision} {
		for _, hasCriteria := range []bool{false, true} {
			for _, check := range []bool{false, true} {
				sections := ReportSectionsFor(kind, check, hasCriteria)
				if got := strings.Contains(strings.Join(sections, ","), ReportCriteria); got != (check || kind == KindResearch) {
					t.Fatalf("kind=%s check=%v hasCriteria=%v: %v", kind, check, hasCriteria, sections)
				}
			}
		}
	}
}

func TestCheckerRequiresCriteriaForResearch(t *testing.T) {
	for _, criteria := range [][]Criterion{nil, {{Text: "answer"}}} {
		i := &Issue{Kind: KindResearch, Criteria: criteria}
		if _, err := ParseReport("## Summary\nChecked.", i, true); err == nil {
			t.Fatal("accepted checker report without Criteria")
		}
		text := "## Summary\nChecked.\n## Criteria\n"
		if len(criteria) > 0 {
			text += "1. met — evidence"
		}
		if _, err := ParseReport(text, i, true); err != nil {
			t.Fatal(err)
		}
		if _, err := ParseReport(text+"\n## Findings\n### Result\nQuoting works.", i, false); err != nil {
			t.Fatal(err)
		}
		if _, err := ParseReport("## Summary\nInvestigated.\n## Findings\n### Result\nQuoting works.", i, false); err == nil {
			t.Fatal("accepted research report without Criteria")
		}
	}
}

func TestParseResearchReport(t *testing.T) {
	i := &Issue{Kind: KindResearch}
	if _, err := ParseReport("## Summary\nInvestigated.\n## Criteria", i, false); err == nil || !strings.Contains(err.Error(), "Findings") {
		t.Fatalf("missing findings: %v", err)
	}
	r, err := ParseReport("## Summary\nInvestigated.\n## Criteria\n## Findings\n### Result\nThe library works.\n\nTested quoting and newlines.", i, false)
	if err != nil || len(r.Findings) != 1 || r.Findings[0].Body != "Tested quoting and newlines." {
		t.Fatalf("report: %+v, %v", r, err)
	}
}

func TestResearchRequiresEveryCriterionVerdict(t *testing.T) {
	i := &Issue{Kind: KindResearch, Criteria: []Criterion{{Text: "one"}, {Text: "two"}}}
	for _, tc := range []struct{ criteria, fault string }{
		{"1. met", "missing verdict for criterion 2"},
		{"1. met\n2. evidence only", "criterion 2: missing met/not met verdict"},
		{"1. met\n3. met", "no criterion 3"},
	} {
		_, err := ParseReport("## Summary\nInvestigated.\n## Criteria\n"+tc.criteria+"\n## Findings\n### Result\nEvidence.", i, false)
		if err == nil || !strings.Contains(err.Error(), tc.fault) {
			t.Fatalf("got %v, want %q", err, tc.fault)
		}
	}
}
