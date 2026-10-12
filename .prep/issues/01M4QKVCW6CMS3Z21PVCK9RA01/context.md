New package `internal/eval/variant`; used by the [session runner](01M4QKVBYFPQT50HP41EHZ6F4X) through the [isolated environment](01M4QKVBFKQKCQP6N42BG2K8Z4). What a plugin contains: [Claude Code integration](/components/claude-code.md) (`.claude-plugin/plugin.json`, `skills/prep/SKILL.md`, `hooks/hooks.json`, `hooks/register.ts`, `hooks/classify.ts`).

- **variant.yaml**: `base: current|none|<other variant>`, `replace: {relative/path: file in the variant dir}`, `remove: [paths]`, `append_system_prompt: text`, `description`.
- **Materialize**: `Materialize(v Variant, dst string) (PluginDir, Hash, error)` copies the base plugin, applies the changes, and returns a content hash (sorted paths and contents). `none` yields no `--plugin-dir`.
- **Skill copy**: `internal/cli/skill.md` is identical to the plugin skill (`just sync-skill`); a variant replacing `skills/prep/SKILL.md` leaves `skill.md` alone. A `prep skill` text in a variant is not supported yet.
- **Hook environment**: the SessionStart hook runs `prep prime --hook --plugin "${CLAUDE_PLUGIN_ROOT}"`, which compares plugin and binary versions; a variant keeps `plugin.json` version as is so no warning appears unless a case wants it.
- **Tests**: materialize each kind, hash stability, and a variant that replaces the skill changes the hash; load error for an unknown base.
