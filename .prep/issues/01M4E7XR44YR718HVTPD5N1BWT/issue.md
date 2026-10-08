---
title: 'Markdown store: code fences, legacy config, temp files and conflicts'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - storage
---

Found in the 0.2.0 codebase audit. The markdown store misreads valid files in four ways:

- Fences: a `#` or `##` line inside a fenced code block ended the Open questions section (questions after it escaped the define gate) and started a new decision in decisions.md (a valid file reported invalid). `normalize`, `Sections` and `Blocks` let a `~~~` line close a ``` block.
- Legacy config: converting a 0.1.0 config replaced every `in_progress` in the file, so a view like `--tag in_progress` or a name containing it was rewritten, and `prep fix` wrote that to disk.
- Temp files: an atomic write's `.prep-tmp-*` file in an issue directory or `.prep` was reported as an unknown file (a crash leaves one; the TUI's reload can see one mid-write).
- Conflicts: a change touching several files wrote them one by one, so a compare-and-swap conflict on a later file left the earlier ones written.

All fence tracking goes through one `domain.Fence` that matches the closing fence's character and length; the conversion touches only `--state` values; loading skips temp files; `Apply` renders and checks every file before writing the first.

## Open questions
