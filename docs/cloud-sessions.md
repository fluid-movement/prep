# Cloud sessions

[Documentation](../README.md#documentation)

Define issues on your machine, hand them to Claude Code cloud sessions, and merge what comes back. A cloud session clones the repository from GitHub, so its `.prep/` records arrive with the code, and its own records (claims, work log) come back through the branch it pushes.

The cloud VM needs the same prep installation you have locally. Give it that with a custom cloud environment whose setup script installs prep and runs `prep setup`:

1. Print the script for your version: `prep cloud script`. It pins the release you run, installs it with `install.sh`, and runs `prep setup --harness claude-code`, the non-interactive form of the setup above, so the session gets the same binary, plugin, hooks, skill and activity bridge as a local install.
2. At claude.ai/code, add an environment (network access **Limited** with package managers allowed covers the GitHub download) and paste the script into its setup script box.
3. In your terminal run `/remote-env` in Claude Code and choose that environment as the default for `claude --cloud`.
4. Push your branch, then hand ready, unclaimed issues off from an interactive terminal:
   ```sh
   prep handoff start <id>...       # confirms the session count; add --yes to skip
   ```
   prep checks every issue and that `HEAD` is pushed before it starts anything, then starts one `claude --cloud` session per issue on a pseudo-terminal and logs each session's ID and link on its issue. Without a terminal it prints the exact commands to run and exits non-zero. Each session implements its issue, leaves it ready for a separate check, writes its report in handback format into the records on its branch, commits and pushes. One session handles one issue.
5. Run `prep handoff fetch` to list unmerged `claude/*` branches on origin, the issue each touches and whether its report is present. `prep handoff fetch --merge` merges them one at a time, runs `prep fmt` then `prep check` after each merge, and stops at the first failure. Run the checker pass separately. Neither command pushes your local branch.

Limits to know:

- The setup script lives in your claude.ai environment, not in the repository, and is per account. Collaborators need their own environment.
- The script pins the version it was printed for; print it again and paste it after updating prep so the cloud binary and plugin move with yours. A cloud binary older than the repository's schema fails with a message that names the fix.
- The environment caches the result of the script, rebuilds when you change the script or its allowed hosts, and again after about seven days.
- A cloud session uses your plan's usage like any other session; one small issue took about two points of the five-hour window.

See [harness integrations](harness-integrations.md) for local setup and the [command reference](commands.md) for `prep cloud script`.
