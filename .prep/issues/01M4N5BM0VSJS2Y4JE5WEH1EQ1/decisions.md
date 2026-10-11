## D1: A dependency on a missing issue in a checked-out project is a warning
date: 2026-10-11

The requirement called X005 an error. The other checkout may simply be behind (not pulled), and an error would fail `prep check` (and hooks or CI running it) in this project for a state of another repository it cannot fix. So X005 is a warning whose fix says to pull that checkout first; the dependency counts as unknown meanwhile and does not block.
