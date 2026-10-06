## D1: Keep the format; tolerate blank lines before the metadata
date: 2026-10-06
outcome: true

The format reads well, prep decide writes it, and it is easy to extend. Blank lines after a heading are common markdown style, so readers accept them and prep fmt normalizes them away; the strict form only produced a misleading error. Implemented in 01M47Z7JVGAR02VE3VRWM6PW01. Alternatives: strict with a better error (still trips humans); metadata in the heading (harder to parse and extend).
