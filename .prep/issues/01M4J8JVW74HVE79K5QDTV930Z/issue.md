---
title: prep guide points to prep commands, not record files
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - cli
  - agent-context
---

The `Read:` section of `prep guide` lists record files (`.prep/issues/<id>/issue.md`, `acceptance.md`, `context.md`, `decisions.md`), so it tells the agent to read `.prep` directly, against the rule that records are read through prep. The ledger analysis (01M4J8780KZRKWHJR4NN8AD85M) found 67 direct reads of issue records with `cat`, `sed`, `head` and `ls`, most of them the files the guide names.

The guide names the prep reads instead (`prep show <id>`, a section of it where one exists), and `prep show` carries everything an agent would otherwise read from those files.

## Open questions

- Does `prep show` need a section selector (requirement, context, decisions, criteria) so an agent can read one record narrowly?
