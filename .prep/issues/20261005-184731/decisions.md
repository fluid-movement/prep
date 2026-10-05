## D1: Granular commands per record
date: 2026-10-05

Decided with the user: each record gets a command that matches its semantics (decisions appended with generated IDs and dates, criteria added and checked one by one, history appended) rather than one generic write per file. Granular commands let the domain validate meaning, not just format, and keep the API independent of the markdown layout.

## D2: Acceptance changes are operations, not a rewrite
date: 2026-10-05

The change carries operations (check 2, add text, opt out of an item) instead of a full new criteria list, so the markdown adapter can apply them line by line and keep any prose a human wrote in acceptance.md. Alternative: re-render the file from parsed criteria (simpler, but silently drops text).

## D3: Cloud sessions keep file editing for now
date: 2026-10-05

Decided with the user: the skill's fallback without the binary stays file editing. A better answer is tracked in its own research issue.
