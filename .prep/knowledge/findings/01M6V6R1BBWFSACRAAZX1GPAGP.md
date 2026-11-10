---
type: finding
title: Prep's token cost is extra turns, not hook text or command output
description: 'Measured 2026-11-07 from activity streams and transcripts incl. subagents (fpa-events 84 files, prep 50, daypack 3; Sonnet-ratio prices, ~3.5 chars/token, about ±30%). Hook injections are about 0.3-0.4% of session cost (SessionStart ~500-840 tok, prompt ~240, Stop ~250, skill ~480). Prep command output incl. re-reading it each turn is about 2%. Turns that only run prep commands are 18% of turns and cost in fpa-events and 22% in prep, because each re-reads a 100-500k-token context. Avoidable patterns: guide then show of the same issue (59 shows after 48 guides); 55-59% of show/guide calls re-read an issue already read in the session; 196 back-to-back criterion calls in fpa-events though --add repeats; 8.7% of commands fail in fpa-events (complete 34, claim 32, list 27, ready 19) against 0.7% in prep, reasons not recorded; prep help called 10 times. The Stop decision question led to a recorded decision 18 of 53 times in fpa-events (about 1.5% of cost) and 2 of 9 in prep: keep, maybe narrow. Daypack has one session, too little data. Do not trim hook text.'
date: "2026-11-07"
issue: 01M5FS7KX63E8KZARSXXAS44C4
---


