## D1: Fixed priority levels, default medium, no inheritance
date: 2026-10-06

Decided with the user: four fixed levels (critical, high, medium, low) in issue.md frontmatter, unset meaning medium and not written; each issue has only its own priority; priority orders, filters and shows, and is editable in any state like tags. Alternatives: a ranked order between issues (every reorder touches neighbours and conflicts in merges); three levels; unset sorting last; children inheriting their parent's priority.

## D2: Sort at the query layer, not in Tree.IDs
date: 2026-10-06

Tree.IDs stays chronological (validation, indexing and stable diagnostics rely on it). A domain helper orders IDs by priority then ID, applied where work is listed: Tree.Query results, the roots and children in Tree.TreeOrder, and prep prime's lists. Alternative: make IDs priority-ordered (reorders unrelated output such as diagnostics and changes the children index).

## D3: No schema bump for the optional field
date: 2026-10-06

priority is an optional, additive frontmatter field like tags, which also arrived within schema 1; absent means medium, so every existing file stays valid. An older binary rejects only issue.md files that set a priority (KnownFields), and prep update resolves that. Alternative: schema 2 with a no-op migration, which would make older binaries refuse the whole project instead of the few prioritized issues.
