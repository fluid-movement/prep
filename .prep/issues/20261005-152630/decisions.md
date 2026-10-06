## D1: Follow OKF v0.2 closely
date: 2026-10-06
outcome: true

Decided by the user: the knowledge directory should be usable by OKF tools directly. prep generates index.md files with okf_version at the root, treats log.md as reserved, and accepts deprecated when reading; the convention of deleting stale entries stays. The overview remains an entry of type overview describing the project; navigation comes from the index files. Implemented in 20261006-064320. Alternative: overview as the only map, no index files (simpler, but OKF tools would see no listings).
