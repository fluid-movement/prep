## D1: The git remote is a project's identity; names are handles mapped onto remotes
date: 2026-10-11

The user proposed the git repository as the project's ID, so a cross-project link resolves on every machine that has the other project checked out. Records use short names (`api:01M4…`) instead of URLs so they stay readable; each project's `project.md` maps the names it uses to remotes (`links:`), and its own `name:` is initialized from the repository name. Remotes are normalized to `host/owner/repo`. A project that is not checked out resolves to "unknown", never blocking. Rejected: directory basenames as identity (differ per clone, collide).

## D2: A permanent project ID is the identity; the remote is only a clone hint
date: 2026-10-11
supersedes: D1

The remote changes when a repository is renamed, transferred or moved to another host, and nothing in the referring project can notice, so it cannot be the identity. Each project gets a ULID `id:` in project.md, written once by `prep init` and committed; links store the target's ID plus its last known remote as a clone hint, under the referring project's own local name. Records use `<local name>:<full issue ID>`. Renaming the target (`name:`), changing its remote or moving its directory never breaks a link; renaming a local name goes through `prep link rename`, which rewrites every reference. A fork that keeps project.md is a clone of the same project; `prep init --new-id` splits it off. The user asked for links "as solid as possible".
