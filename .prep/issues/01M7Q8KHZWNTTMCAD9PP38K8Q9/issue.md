---
title: Publish docs/ as a GitHub Pages site
kind: code
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - docs
  - cloud
---

Publish the `docs/` pages as a GitHub Pages site at https://fluid-movement.github.io/prep/ (the default domain GitHub gives a repository owned by a user account; no custom domain). The pages stay the single source; there is no copy in the GitHub wiki.

Settled with the user: publish publicly, and use MkDocs Material built by a GitHub Actions workflow (navigation, search, dark mode, and `mkdocs build --strict` validates links and anchors). A workflow `.github/workflows/docs.yml` builds on every push to main that changes `docs/`, `mkdocs.yml` or the workflow, and deploys through the GitHub Pages actions. `mkdocs.yml` carries the navigation in the same order as the README's page list; `docs/index.md` is the landing page; the docs dependencies are pinned in `requirements-docs.txt`. The README links to the published site. The repository's Pages source is set to GitHub Actions, which is a repository setting and is done through the API once the workflow exists.

Follows the decision "User docs are Markdown pages in docs/, command reference generated" (`prep knowledge find user docs`), which left a site for later; the generated `docs/commands.md` is published like any page.

## Open questions
