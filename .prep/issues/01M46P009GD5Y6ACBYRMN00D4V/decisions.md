## D1: Pin the marketplace to the binary tag
date: 2026-10-06

Third-party marketplaces do not auto-update, and an unpinned marketplace follows main, so the plugin could run ahead of or behind the binary. Pinning the marketplace to the binary's tag makes the installed plugin exactly match the installed binary; setup moves the pin when the binary changes. Alternative: follow main with a compatibility range (weaker guarantee, more checks).

## D2: Prime does the version check
date: 2026-10-06

The hook passes the plugin root to prep prime, which reads plugin.json and compares versions. The comparison lives in Go, tested, and the hook stays one line. Alternative: compare in a shell script inside the plugin (untested, duplicated per harness).
