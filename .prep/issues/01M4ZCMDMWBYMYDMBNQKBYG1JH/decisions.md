## D1: Keep --project and the registry; remove links and cross-project dependencies
date: 2026-10-16

User decision, 2026-10-16. Talking about another prep project from inside one is very useful (the fpa api, events and profiles projects work in tandem), and --project as a plain flag keeps browsing and filing consistent across projects. Keep the project registry, prep projects, --project and the session briefing's line naming other projects; the registry forgets entries whose directory no longer exists (scratch directories end up in it today). Remove links (prep link and prep links, the links section in project.md), cross-project dependencies with their satisfied/unsatisfied/unknown status, cross-project knowledge links and their TUI display. Usage on this machine: zero links, zero cross-project dependencies and knowledge links across 8 projects, although the fpa projects had real dependencies that nobody recorded. A dependency on another project is stated in the requirement and checked with --project; the human says when to wait. Rebuild a simple version only if real cases demand it.

## D2: Verdicts carried into release 0.3.0
date: 2026-10-17
outcome: true

Verdicts recorded in this issue's decisions, 2026-10-16/17, are carried into release 0.3.0 (01M52ZNHXFXHNDATT9WWQ3C87E) by 01M4QY3CPZ25AEGV7R1CC2GPT6, 01M52ZQGH8HH6DAK9HEGZ28KDE, 01M52ZS6090E8W7APD2DWVDSBX. Open questions raised here were settled by the release's decision issues.
