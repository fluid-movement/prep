---
title: Finish hook asks for decisions; redirects match paths
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZRB3VY0MSZ0HZWJ2SEC24
tags:
  - integration
---

After a turn that changed code, the finish hook asks once whether a decision was made and should be recorded; nothing else. The record redirect hook matches the paths a tool touches, not words in command text (it blocked a prep log whose text named the records directory). Needs an eval case. Decided in 01M4ZCM7TTPASZ2EN8MQN6EPVZ and 01M4ZCM5XQ3J6TM1N4AZ330ZH7.

## Open questions
