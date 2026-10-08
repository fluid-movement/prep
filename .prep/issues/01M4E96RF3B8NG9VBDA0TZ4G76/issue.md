---
title: 'Domain edge cases: ID case, dropped dependencies, ack after a kind change, dropped parents'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - cli
---

Found in the 0.2.0 codebase audit.

- `Resolve` was case-sensitive, so `prep show shdm3tm` found nothing, though ULIDs are Crockford base32, which ignores case.
- `prep new --depends-on <dropped>` succeeded (prep edit refuses it), leaving an issue that can never be claimed; repeated dependencies were kept.
- `prep ack` after a kind change (say manual to code) kept the sign-off without the context check ready would make.
- A dropped parent with unresolved children raised no warning; a done one does (I023).
- The Agent screen built the issue guide, which validates the whole tree, on every frame.

## Open questions
