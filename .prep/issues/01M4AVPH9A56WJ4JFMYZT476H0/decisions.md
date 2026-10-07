## D1: Drift counts from the entry's last change
date: 2026-10-07

An entry is current as of the later of its confirmed_commit and the last commit that changed the entry file; while the entry file has uncommitted changes, uncommitted changes in its scope count as reviewed with it.

Rationale: writing or confirming an entry is the review; tying drift to the entry's own history lets code and entry land in one commit. confirm stays the way to record a review without editing the body.

Alternatives: a pending confirmed_commit filled in later (needs a later write, the second commit again); ignoring uncommitted changes in drift (hides drift while working).
