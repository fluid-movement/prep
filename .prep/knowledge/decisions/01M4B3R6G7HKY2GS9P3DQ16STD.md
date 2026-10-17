---
type: decision
title: "prep check validates both files"
description: "prep check validates both files"
date: "2026-10-07"
issue: 01M4AHPY6HZ0ZFMMH349V1V2TC
---

`prep check` parses `config.yaml.dist` and, when it exists, `config.yaml`, and reports errors in either with the file name.

Rationale: the dist file is what every fresh clone uses, so a broken dist must surface even for a user whose own config is fine. The effective config is the one loaded into the tree.

Alternatives: validate only the effective file (a broken dist goes unnoticed by everyone who has their own config).
