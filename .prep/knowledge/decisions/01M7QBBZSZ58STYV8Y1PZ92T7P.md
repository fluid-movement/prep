---
type: decision
title: 'Docs site: MkDocs Material on GitHub Pages, built by Actions'
description: Public site at fluid-movement.github.io/prep from docs/, MkDocs Material, deployed by a workflow; no wiki copy.
date: "2026-11-18"
issue: 01M7Q8KHZWNTTMCAD9PP38K8Q9
---

The user asked to set Pages up and to publish. MkDocs Material needs no front matter in the pages (the Jekyll route needs it and would disturb the generated commands page and the link checks), gives search and navigation, and `--strict` validates links. Built by a GitHub Actions workflow so the Pages source is "GitHub Actions", not a branch. The GitHub wiki was rejected: a separate repository, a second copy, no pull requests. A custom domain is not used; the default `fluid-movement.github.io/prep` applies because the owner is a user account.
