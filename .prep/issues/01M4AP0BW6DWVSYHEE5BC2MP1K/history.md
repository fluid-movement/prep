- 2026-10-07T08:03:15Z edited by claude-code/opus-5.5: requirement

- 2026-10-07T08:04:52Z claude-code/opus-5.5: Done: filterChanged after every edit (keys, paste, suggestion) sets the filter when parseFilterText accepts the text; esc restores m.before; enter applies or errors. Found on the way: the first typed '-' parsed as a title search, so words starting with - are now flags, never text. Snapshots unchanged in text.
