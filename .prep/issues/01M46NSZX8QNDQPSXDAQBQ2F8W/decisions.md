## D1: Error for a duplicated Open questions section, warning otherwise
date: 2026-10-06

One code, I028, for a `##` heading that appears more than once in `issue.md`, `acceptance.md`, `context.md` or `findings.md`. Headings compare case-insensitively, ignoring surrounding space and anything inside code fences. A duplicated `## Open questions` in `issue.md` is an error, because prep reads only the first copy and questions under the others would not block define. Every other duplicate is a warning, since nothing prep reads is lost. The class is fixable when at most one copy has content: `prep fix` keeps that copy (or the first, when all are empty) and drops the empty ones. Otherwise the class is guided, and someone moves the content under one heading.

Alternatives considered: making every duplicate an error. That would block unrelated writes on harmless free-form notes. Merging copies that have content automatically: the order and intent of two non-empty sections is a judgement call.
