## D1: Write injected like Load
date: 2026-10-06

The TUI receives a write function that runs the CLI's pipeline (plan, CheckWrite, Apply, stage), mirroring the injected loader. It keeps the TUI free of storage imports and guarantees TUI writes pass the same validation as CLI writes. Alternative: the TUI shells out to the prep binary (slower, stringly-typed errors, harder to test).

## D2: Human actor from the OS user
date: 2026-10-06

TUI writes are recorded as human:<user>, which the domain already treats as a person: code issues cannot be completed by it. Alternative: reuse PREP_ACTOR (meant for agents; would let the TUI complete code issues if set to an agent name).

## D3: Action menu shows unavailable actions with reasons
date: 2026-10-06

Listing every action with the gate that blocks it teaches the lifecycle where it is used (for example: define needs an empty Open questions section). Alternative: show only available actions (shorter, but the user cannot tell why something is missing).
