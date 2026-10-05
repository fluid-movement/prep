- 2026-10-05T19:09:55Z claude-code/2.1.289: Snapshot tests caught the first layout fixes (column gaps, double blank lines, markdown padding) as golden diffs, as intended.

- 2026-10-05T19:09:55Z claude-code/2.1.289: Binary grew from 3.3 MB to 14.8 MB with Glamour; prep list startup went from about 5 ms to 10-20 ms.

- 2026-10-05T19:12:02Z claude-code/2.1.289: Verified prep tui --gallery in a pseudo-terminal at 100 and 80 columns: renders every section, scrolls, quits with q (exit 0). Startup waits for the terminal's answer to the background color and cursor queries (Lip Gloss/termenv); a terminal that never answers delays startup by the termenv timeout.
