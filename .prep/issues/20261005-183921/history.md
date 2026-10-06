- 2026-10-06T06:48:43Z edited by claude-code/2.1.289: tags

- 2026-10-06T11:24:35Z edited by claude-code/2.1.291: parent

- 2026-10-06T13:53:28Z claude-code/2.1.291: I028 from loadIssue via duplicateHeadings (sections.go: ## sections outside code fences, case-insensitive). Error only for Open questions in issue.md, because splitOpenQuestions takes the first copy; acceptance.md re-enters the DoD section at every DoD heading, so its duplicates are warnings. Fix merges when at most one copy has content. Corpus: invalid/duplicate-open-questions, valid/duplicate-heading. This repo has no duplicates today (the import leftover was removed by hand).
