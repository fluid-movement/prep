---
title: Publish docs/ as a GitHub Pages site
kind: code
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - docs
  - cloud
---

Publish the `docs/` pages as a free GitHub Pages site, with navigation and search, built from the repository on every push to main. The pages stay the single source; there is no copy in the GitHub wiki.

Open choices to settle when defining: the generator (the built-in Jekyll with a docs theme such as just-the-docs needs only front matter and a `_config.yml`; MkDocs Material built in a GitHub Actions workflow gives better search and navigation and needs a small `mkdocs.yml`), and whether the README links to the published site. Enabling Pages publishes the docs on the public internet, which is the user's call, not an implementer's: the issue ends with the repository setting change described for the user to make.

Follows the decision "User docs are Markdown pages in docs/, command reference generated" (`prep knowledge find user docs`), which left a site for later.

## Open questions
