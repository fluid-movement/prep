#!/usr/bin/env python3
"""Compatibility entry point for Codex sessions loaded before shared hooks.

Keep this file while installed sessions may still reference it. Workflow
behavior lives in the prep binary; launcher errors must never block a tool.
"""

import subprocess
import sys


def main():
    modes = {
        "session": "session-start",
        "prompt": "prompt",
        "before-tool": "before-tool",
        "after-tool": "after-tool",
    }
    mode = modes.get(sys.argv[1] if len(sys.argv) > 1 else "")
    if not mode:
        return
    try:
        result = subprocess.run(
            ["prep", "hook", mode, "--harness", "codex"],
            stdin=sys.stdin, capture_output=True, text=True, timeout=8,
            check=False,
        )
        if result.returncode == 0:
            sys.stdout.write(result.stdout)
    except Exception:
        pass


if __name__ == "__main__":
    main()
