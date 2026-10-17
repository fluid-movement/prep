---
type: decision
title: Archive names are a contract
description: '`prep_<version>_<os>_<arch>.tar.gz` (zip on Windows) plus checksums.txt must match update.ArchiveName and install.sh.'
date: "2026-10-17"
scope:
  - .goreleaser.yaml
  - internal/update
  - install.sh
---

`prep_<version>_<os>_<arch>.tar.gz` (zip on Windows) plus checksums.txt must match update.ArchiveName and install.sh.

Both consumers derive the name.
