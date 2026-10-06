- 2026-10-06T06:48:44Z edited by claude-code/2.1.289: tags

- 2026-10-06T10:40:28Z edited by claude-code/2.1.289: title, depends_on, requirement

- 2026-10-06T10:40:41Z released by claude-code/2.1.289 (claimed by claude-code/2.1.289 at 2026-10-06T10:40:29Z): user asked to stop after enrichment; implementation not started

- 2026-10-06T10:50:23Z claude-code/2.1.291: Moved the fsnotify watcher to internal/watch, added prep watch and history in prep show --json, with Go tests

- 2026-10-06T10:50:23Z claude-code/2.1.291: Wrote the panel as the prep plugin's hooks module (hooks/register.tsx, infer.ts, panel.tsx, types/index.d.ts, tests/panel.test.tsx); /prep:focus and /prep:pane are commands/*.md answered by command.run hooks, since registered commands cannot carry the prep: namespace. Developed in the plugin folder directly: the installed plugin is read from this repository, so /reload-plugins loads it

- 2026-10-06T10:51:56Z claude-code/2.1.291: Live try: the pane opened at session start, /prep:pane closed and reopened it, prep show did not move the panel and prep guide did
