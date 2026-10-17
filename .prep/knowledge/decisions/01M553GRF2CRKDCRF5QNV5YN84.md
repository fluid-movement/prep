---
type: decision
title: install.sh never edits shell files and needs no root
description: Installs into ~/.local/bin and prints a ready-to-paste PATH line instead of editing rc files.
date: "2026-10-17"
scope:
  - install.sh
---

XDG user executable location, no root needed; (implied) editing user shell files is intrusive. Also: downloads are checksum-verified, and update refuses on mismatch.
