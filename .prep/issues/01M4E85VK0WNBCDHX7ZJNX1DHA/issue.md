---
title: Knowledge drift for unconfirmed entries; safe knowledge writes; exact git paths
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - knowledge
  - okf
  - git
---

Found in the 0.2.0 codebase audit.

- Drift was computed only for entries with `confirmed_commit`, although the docs say the baseline is the later of it and the entry's last commit. Scoped entries never confirmed (agent-activity, tui-agent) were never checked.
- `prep knowledge new /x/index.md` or `/log.md` was accepted (index.md was then overwritten by the generated index), and creating an entry over a file that failed to load overwrote it silently.
- An entry's frontmatter ended at any line starting with `---`, such as `----` in a block scalar.
- git path lists were parsed line by line, so git's quoting of non-ASCII names made drift lists and the ignore check in staging miss those files; read-only git calls could take `index.lock` from a commit running alongside.

## Open questions
