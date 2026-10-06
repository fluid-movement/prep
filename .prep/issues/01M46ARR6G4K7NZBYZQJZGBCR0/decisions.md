## D1: Accept the lag; prep never pushes
date: 2026-10-06
outcome: true

Pushing is an outward action with network and permission side effects that belongs to the user or the harness, not to a local tool. Working on main, as this project does, has no lag, and parallel agents on branches can coordinate through the harness. Alternative: commit and push the claim to main on prep claim (surprising side effect, fails offline).
