- 2026-10-07T07:49:48Z edited by claude-code/opus-5.5: title, requirement

- 2026-10-07T07:49:55Z released by claude-code/opus-5.5 (claimed by claude-code/opus-5.5 at 2026-10-07T07:47:40Z): requirement extended with not- and in-progress

- 2026-10-07T07:53:19Z claude-code/opus-5.5: Done: Filter Not* fields, NegationPrefixes (not-, !), Query excludes; ValidTag rejects not- tags; StateInProgress is in-progress (underscore rejected), StateLabel splits on hyphens; picker keeps a typed not-/! and hides counts; prep flags and list help show not-. Converted: fixtures, config.yaml.dist, DefaultConfig, plugin panel and project view (claude plugin test: 21 pass), README, goldens (only the spelling changed). A bulk sed briefly rewrote issue records and two knowledge entries by hand (BSD grep has no \| alternation in the exclude filter); they were restored from the index and the knowledge entries were updated through prep knowledge update.
