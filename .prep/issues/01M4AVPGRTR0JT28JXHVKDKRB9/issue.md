---
title: Config errors repeat the file name
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - config
priority: low
---

`prep check` prints config errors as `config.yaml: config.yaml: unmarshal errors…`: the diagnostic names the file and its message names it again. The message drops the prefix; errors from commands that load the config (which have no file column) keep naming the file.

## Open questions
