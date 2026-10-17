---
type: decision
title: "Hook runs prime only when the binary is present"
description: "Hook runs prime only when the binary is present"
date: "2026-10-05"
issue: 01M46ARJB0BYKFHTN456ZAS5SB
---

The hook is `command -v prep >/dev/null 2>&1 && prep prime || true`. Installing the binary in cloud sessions waits for the install flow; until then the skill's "Without the binary" section covers those sessions. Alternative: download a binary in the hook (no release artifacts exist yet).
