---
type: decision
title: Activity archives are read only on request
description: Past 1 MiB the stream moves to a dated archive and a new stream starts; only prep activity --all reads the archives, every other reader sees the current stream.
date: "2026-11-11"
issue: 01M72AV3TQPT1H6GQ6RY46R17T
scope:
  - internal/activity/**
supersedes:
  - 01M5722AWW3PAZEMQD04NAZDT4
  - 01M5722J2FD04MQCPYRK8G7THS
---

Past 1 MiB the stream is archived to `.prep/local/activity-<date>.jsonl` (`-<n>` added when the name exists) and a new stream starts, instead of being pruned: pruning kept two to three days and the usage analysis needs weeks. The rename is atomic, so concurrent writers at worst archive once. prep never prunes the archives. New on 2026-11-11: `prep activity --all` (`activity.ReadAll`) reads the archives with the current stream, because the usage research had to fall back to harness transcripts when the stream of the day held only that day. The TUI, prime and plain `prep activity` still read only the current stream, so their cost is unchanged. `--all` rather than `--since <date>`: analysis wants everything and filters itself, and one flag was enough. Files are ordered by their first event, not by name: an archive is named after the day it was closed and the second of a day (`-2`) sorts before the first.
