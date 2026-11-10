---
title: record why prep commands fail and fix the top failing gates
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
---

Research finding: 8.7% of prep commands failed in fpa-events (complete 34, claim 32, list 27, ready 19), but the activity stream only stores failed:true. Record the error code per failed command, as the earlier hook-outcomes decision intended, then fix the most common causes. Related bug 01M5A68QTP1ZHEEXW9NY8BEBYB. Not yet defined.

## Open questions
