- 2026-10-05T18:58:25Z edited by claude-code/2.1.289: depends_on, requirement

- 2026-10-05T19:02:24Z claude-code/2.1.289: Implemented the six record commands; TestRecordCommands covers each, including text a human wrote in acceptance.md surviving criterion edits.

- 2026-10-05T19:02:24Z claude-code/2.1.289: Context and findings now require --body or --body-file, so a bare call cannot silently clear the record.

- 2026-10-05T19:02:24Z claude-code/2.1.289: Found that history appends re-read history.md without compare-and-swap; all appends now check the hash first.
