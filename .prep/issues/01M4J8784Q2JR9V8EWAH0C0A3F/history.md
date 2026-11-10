- 2026-10-10T06:32:48Z edited by claude-code/opus-5.5: requirement

- 2026-10-10T06:32:55Z edited by claude-code/opus-5.5: requirement

- 2026-10-13T06:02:04Z edited by codex: requirement

- 2026-10-13T06:12:47Z edited by codex: parent

- 2026-10-16T18:33:37Z edited by claude-code: parent

- 2026-10-18T12:16:47Z edited by claude-code: requirement

- 2026-10-18T12:16:47Z claude-code: Local model candidate waits for the hardware upgrade: this machine has 8 GB of RAM, too little to run the model next to the harness.

- 2026-10-18T12:16:59Z edited by claude-code: requirement

- 2026-11-10T08:50:40Z claude-code: 2026-11-10: finding 01M72A9TBJ5JQXT2VGT15N2HW1 bears on the prompt-classification part. Replaying real prompts against issue titles found the issue the agent then opened in under 10% of first touches; prompts are conversational, so a classifier that sees only the prompt (rules, Jev or a local model) lacks the information and would need the conversation tail. The tool-call part is covered by rules; their problem is false denials (01M5A68QTP1ZHEEXW9NY8BEBYB), not missed calls.
