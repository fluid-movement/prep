---
title: Align the knowledge bundle with OKF v0.2
kind: code
depends_on:
  - 20261005-152630
---

The knowledge directory follows the Open Knowledge Format v0.2 closely so OKF tools can read it directly. prep generates an index.md in every bundle directory listing its entries and subdirectories as * [Title](/path) - description, and the bundle-root index.md declares okf_version: "0.2"; the index files are rewritten on every knowledge write and by prep fmt, and prep check reports outdated ones as fixable. log.md is treated as the reserved chronological log, not as an entry. status: deprecated, which OKF allows, is accepted when reading, while the convention of deleting stale entries stays. The overview remains an ordinary entry of type overview that describes the project.

## Open questions
