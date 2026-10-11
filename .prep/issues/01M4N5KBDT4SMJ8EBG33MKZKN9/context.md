Code paths:
- `internal/domain/links.go`: `ProjectLink`, `Project.Links`, `SplitRef`, `XKnowledgeLinks`, `Tree.LinkRefs` (dependencies and `](name:/…)` links in requirement, context, findings, decisions, criteria, work log and knowledge entries), `validateLinks` (X001–X003), `PlanLinks`, `PlanLinkRename` (with `Change.LinkRename`), `PlanLinkRemove` (refused while referenced, listing the references), `RenameLinkInText` (raw-text rename used by the store and by `Tree.Apply`).
- `internal/domain/diag.go`: the X family for links and cross-project references (X001–X007; X004–X007 come from the registry, not `Validate`).
- `internal/mdstore`: `projectFM.Links` (name → `{id, remote}`), rendering `Change.Links` into project.md and `Change.LinkRename` over every record and knowledge file (`linkRefFiles`).
- `internal/projects/links.go`: `ResolveLinks`, `Check` (X004 clone hint differs from the checkout's remote), `FixRemotes`.
- `internal/cli/links.go`: `prep links`, `prep link add|rename|remove`, `validate` (domain plus registry checks, used by `check`, `prime` and the TUI), `fixLinkRemotes` in `prep fix`.

Knowledge: [Cross-project work](/features/cross-project.md), [Storage format](/conventions/storage-format.md), [Domain package](/components/domain.md), [Project registry](/components/projects.md).
