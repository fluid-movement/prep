---
type: decision
title: "Update from GitHub Releases directly"
description: "Update from GitHub Releases directly"
date: "2026-10-06"
issue: 01M46ARP808AXT8MABAYMNHSY3
---

prep update reads the GitHub releases API and checksums.txt itself instead of depending on a self-update library: a few dozen lines of net/http, archive/tar, archive/zip and crypto/sha256, all standard library, keep the binary small and the behaviour visible. Alternative: a library such as go-selfupdate (more features, another dependency tree).
