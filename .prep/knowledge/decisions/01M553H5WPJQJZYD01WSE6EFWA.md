---
type: decision
title: No journal, no rollback
description: An interrupted write leaves the previous or new state plus harmless extra files that `prep check` reports; the fix is to run the command again.
date: "2026-10-17"
scope:
  - internal/mdstore
---

(implied) ordered atomic file writes make recovery by rerun enough.
