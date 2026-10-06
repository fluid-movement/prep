## D1: Inline editing, overlay dialogs, direct keys, editable settings
date: 2026-10-06

Decided with the user: the flows that must feel best are creating an issue with its requirement and editing the project settings. Longer text gets an inline multi-line editor with ctrl+e handing off to $EDITOR; dialogs overlay the screen; frequent actions get direct keys and the action menu lists only what applies. Alternatives: keeping $EDITOR as the only path for long text; dialogs replacing the screen; the action menu as the single entry point.

## D2: Key assignments for direct actions
date: 2026-10-06

Existing keys keep their meaning (c check screen, s settings, t tree, p parent, y copy, r reload). New direct keys: E edit context, k check criteria, i set priority, > the issue's next transition (as prep guide orders transitions; input-free transitions ask for a second > in the footer, complete and drop open their dialogs); n and e stay. Alternatives: reassigning c to context (breaks the check screen key users know); running transitions without confirmation (define and ready write records that cannot be undone).

## D3: Config writes go through the domain like issue writes
date: 2026-10-06

Settings edits are a domain Change (PlanConfig) validated by CheckWrite and written by the markdown store, so the TUI uses the same pipeline (stage or commit via record) as every other write, and an invalid view query is rejected before anything is written. Alternative: the TUI writing config.yaml directly (bypasses validation and the commit mode).

## D4: Letter sequences over menu navigation
date: 2026-10-06

The user likes the melody of a i l for setting a low priority: every action stays reachable as a short, stable sequence of letters that turns into muscle memory, and arrows plus enter are only the fallback. So menus and pickers keep fixed mnemonic letters (the priority picker c h m l, the action menu's letters), direct keys shorten frequent sequences without changing them (i l works as well as a i l), and new dialogs (settings, the > confirmation) follow the same pattern: no list that can only be scrolled. Alternative: arrow-driven menus and pickers, which cannot become muscle memory.
