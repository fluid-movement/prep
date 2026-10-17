---
type: decision
title: "Refresh through the new binary"
description: "Refresh through the new binary"
date: "2026-10-06"
issue: 01M48721G8Y21HT0ZGAXB5QZF2
---

prep update runs prep setup --refresh with the freshly installed binary instead of refreshing from the old process, because the new version knows the current plugin layout and harness list. Alternative: refresh from the running (old) binary, which may install integrations the new binary no longer matches.
