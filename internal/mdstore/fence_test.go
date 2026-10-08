package mdstore

import "testing"

func TestOpenQuestionsIgnoreHeadingsInFences(t *testing.T) {
	body := "Prose.\n\n## Open questions\n- How to run?\n\n```sh\n# install first\nmake\n```\n\n- And then?\n\n## Notes\nAfter."
	prose, questions := splitOpenQuestions(body)
	want := "- How to run?\n\n```sh\n# install first\nmake\n```\n\n- And then?"
	if questions != want {
		t.Fatalf("questions =\n%q\nwant\n%q", questions, want)
	}
	if prose != "Prose.\n\n## Notes\nAfter." {
		t.Fatalf("prose = %q", prose)
	}
	if p, q := splitOpenQuestions("```md\n## Open questions\n- not one\n```"); q != "" || p == "" {
		t.Fatalf("a heading in a fence opened the section: %q %q", p, q)
	}
	if hasOpenQuestions("```\n## Open questions\n```") {
		t.Fatal("hasOpenQuestions counts a heading in a fence")
	}
}

func TestDecisionBodiesKeepFencedHeadings(t *testing.T) {
	raw := "## D1: Template\ndate: 2026-10-05\n\nThe file:\n\n```md\n## inner\n```\n\n## D2: Next\ndate: 2026-10-06\n\nRows.\n"
	ds, problems := parseDecisions(raw)
	if len(problems) > 0 || len(ds) != 2 || ds[0].Body != "The file:\n\n```md\n## inner\n```" {
		t.Fatalf("parsed %+v problems %v", ds, problems)
	}
	if got := canonicalDecisions(raw); got != raw {
		t.Fatalf("canonical =\n%q\nwant\n%q", got, raw)
	}
}

func TestNormalizeTracksFenceType(t *testing.T) {
	in := "````md\n```go\nx  \n\n\n```\n````\ntail  "
	want := "````md\n```go\nx  \n\n\n```\n````\ntail"
	if got := normalize(in); got != want {
		t.Fatalf("normalize =\n%q\nwant\n%q", got, want)
	}
	in = "```\na\n~~~\nb  \n\n\n```"
	if got := normalize(in); got != in {
		t.Fatalf("a ~~~ line closed a ``` fence: %q", got)
	}
}
