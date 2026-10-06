## D1: Render through the adapter before validating
date: 2026-10-06

Links are extracted by the OKF parser, so the adapter renders the would-be entry and parses it back; the domain validates the tree with that entry before anything is written. The domain keeps the rules, the adapter keeps the format. Alternative: move link parsing into the domain (couples it to markdown).

## D2: Edit frontmatter in place
date: 2026-10-06

OKF entries may carry keys prep does not model (generated, verified, sources, resource). Updating the YAML node instead of re-rendering from the domain struct keeps them and their order. Alternative: model every OKF key in the domain (grows with the format).
