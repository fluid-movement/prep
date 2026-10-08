---
title: Pi pane uses the prep theme
kind: code
parent: 01M46ARM9G3492V298TM82KD45
depends_on:
  - 01M4DQMG9ZVZT172V0SC967R2C
  - 01M4AWZVXW06G512BPA4Z7G6EC
tags:
  - integration
  - pi
---

As the Claude Code panel does after 01M4AWZVXW06G512BPA4Z7G6EC, the Pi pane paints with the colors `prep theme colors --json` prints, following prep's configured theme or a theme chosen for the pane, with an option to keep Pi's own theme colors. It reads the colors at session start and when `prep watch` reports a change, and falls back to Pi's theme when the command fails.

## Open questions
