## D1: Own config replaces the dist file as a whole
date: 2026-10-07

prep uses `.prep/config.yaml` when it exists and otherwise `.prep/config.yaml.dist`; the two are never merged.

Rationale: the user asked for a fallback, and a whole-file rule is easy to predict: what is in your file is what applies. Views are ordered maps, so a merge would need rules for order and for removing a shared view.

Alternatives: key-by-key merge of own over dist (surprising when a user deletes a view and it comes back from the dist file).

## D2: prep check validates both files
date: 2026-10-07

`prep check` parses `config.yaml.dist` and, when it exists, `config.yaml`, and reports errors in either with the file name.

Rationale: the dist file is what every fresh clone uses, so a broken dist must surface even for a user whose own config is fine. The effective config is the one loaded into the tree.

Alternatives: validate only the effective file (a broken dist goes unnoticed by everyone who has their own config).

## D3: Ignored paths are never staged
date: 2026-10-07

`record` filters the written paths through `git check-ignore` before staging or committing, so writing the personal `config.yaml` does not run `git add` on it.

Rationale: `git add` on an ignored path fails, so without the filter every settings save warns, and commit mode all breaks. Filtering ignored paths generally is simpler than special-casing config.yaml, and it also respects any ignore rules a user adds.

Alternatives: drop `.prep/config.yaml` by name (works, but duplicates the ignore rule in code).
