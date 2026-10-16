## D1: Prep's job: clear issues and documented knowledge
date: 2026-10-16

User direction, 2026-10-16. Prep's job is to make sure issues are clear and knowledge about the project is documented. Getting agents to use prep is done with skills, hooks and maybe an MCP server. Prep currently has too much orchestration and too much friction. Every proposed change must answer: does this make the process better, or does it only add overhead? The user works on issues almost exclusively through agents; the TUI is for overview and following the agent.

## D2: Prune the tree: delete everything prep does not use
date: 2026-10-16

User direction, 2026-10-16. Prune the tree: delete all code and features prep is not using, not only those named in the area decisions. Anything without real use and without a decision to keep it goes, including issues that plan such features. Step 5 starts with removals and ends with a sweep for whatever is left unused: dead code found by tooling, commands and options with no use in the transcripts, records and tests that only serve removed features.

## D3: Review round passed; step 5 plan
date: 2026-10-16

User approved the review page (https://claude.ai/artifact/LBPDrhtuoq2yxHFFReuv1c), 2026-10-16: the decisions together drive prep in the right direction, and the recommendations on the eight tensions are accepted (solo sessions show the DoD commands run; decision eval before any knowledge gate; decision unification as a trial; OKF index merge driver; knowledge conversion as its own issue; 0.3.0 not released as it stands, the rethink becomes its content; parallel implementers use separate worktrees; step 5 removes first, one small issue at a time). Step 5 creates one release parent with ordered issues: the eval harness fix first, removals next, then the knowledge model, guidance, open decisions and a final prune sweep. Work goes one issue at a time with the user; done requires GitHub CI to be green.
