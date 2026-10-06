## D1: prep import plans a guided research issue, like the knowledge bootstrap
date: 2026-10-06

`prep import` creates one research issue, "Import existing work", tagged `import`, with a context that guides the agent:
1. Ask the user which sources the project uses, and look for the usual ones.
2. Write the candidates as findings, one line each with title, kind and source; the user reviews and strikes what should not come in.
3. For each approved candidate, skip it when `prep list --text "<source>"` finds it. Otherwise create it with `prep new` (title, kind, tags, priority when the source has one) and a requirement that ends in `Source: <ref>` and lists unresolved points under Open questions.
4. Complete the issue with the findings.
The command refuses to create a second unresolved import issue, and prints its ID and `prep guide <id>`.

Alternatives considered: importers per source, which are project-specific and which the user ruled out; text in the skill only, which does not show up in prime and guide and leaves no record of what was imported.
