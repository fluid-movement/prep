## D1: Edit is refused on done and dropped issues
date: 2026-10-05

A resolved issue is history: its requirement, kind and relations describe what was done. Changing them afterwards rewrites the record the knowledge base and dependents point to. Alternative: allow title-only edits on resolved issues (cosmetic, but one rule is simpler; revisit if needed).

## D2: Edits append a history line
date: 2026-10-05

Every write leaves a record without relying on git history (design principle 4). The line names the time, the actor and the changed fields, not old and new values; requirement changes are visible through baselines and staleness. Alternative: no record (loses who changed relations and titles).

## D3: --depends-on replaces the list
date: 2026-10-05

One flag with replace semantics, decided with the user: the given values become the whole list, a single empty value clears it, and an empty value mixed with IDs is a usage error. Alternative: separate add and remove flags (two flags for one field).
