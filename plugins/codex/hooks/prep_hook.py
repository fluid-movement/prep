#!/usr/bin/env python3
"""Codex hooks for prep context, protected records, and activity."""

import json
import os
import re
import shlex
import subprocess
import sys
from pathlib import Path

ACTOR = "codex"
ISSUE = re.compile(r"(?<![0-9A-Z])[0-9A-Z]{26}(?![0-9A-Z])")


def event_input():
    try:
        return json.load(sys.stdin)
    except Exception:
        return {}


def project_root(event):
    cwd = Path(str(event.get("cwd") or os.getcwd())).resolve()
    for path in (cwd, *cwd.parents):
        if (path / ".prep").is_dir():
            return path
    return None


def run_prep(args, root, timeout=5):
    try:
        return subprocess.run(["prep", *args], cwd=root, capture_output=True,
                              text=True, timeout=timeout, check=False)
    except Exception:
        return None


def output_context(event_name, message):
    if not message:
        return
    print(json.dumps({"hookSpecificOutput": {
        "hookEventName": event_name, "additionalContext": message
    }}))


def context(mode, event, root):
    prompt = str(event.get("prompt") or "")
    parts = []
    if mode == "session":
        result = run_prep(["prime", "--hook"], root)
        if result and result.returncode == 0 and result.stdout.strip():
            parts.append(result.stdout.strip())
    # A valid ID is authoritative enough to fetch its guide. For all other
    # prompts, the instruction asks Codex to search the tracker before coding.
    match = ISSUE.search(prompt.upper())
    if match:
        result = run_prep(["guide", match.group(0)], root)
        if result and result.returncode == 0 and result.stdout.strip():
            parts.append(result.stdout.strip())
    if mode == "prompt":
        parts.append(
            "Prep project detected. Before code or research work, check whether "
            "an existing issue covers the request (`prep list --text <distinctive words>`), "
            "and check actionable work with `prep next`. For a matching issue, run "
            "`prep guide <id>` and follow its current step and gates before editing. "
            "If no issue covers authorized work, create one with `prep new`. Use prep "
            "commands for all .prep changes; never edit lifecycle records directly. "
            "For a non-workflow question or a task outside this prep project, proceed normally."
        )
    output_context("SessionStart" if mode == "session" else "UserPromptSubmit", "\n\n".join(parts))


def protected_edit(event, root):
    tool = str(event.get("tool_name") or "")
    payload = event.get("tool_input") or {}
    command = str(payload.get("command") or "") if isinstance(payload, dict) else ""
    if re.search(r"(?:^|\s)(?:prep|\S*/prep)\s+", command):
        return False
    if tool == "apply_patch":
        touches = ".prep/" in command or ".prep\\" in command
    else:
        # Only treat shell constructs which can write as edits; reads such as
        # git diff or cat on .prep remain available.
        writes = re.search(r"(?:>|\btee\b|\bsed\s+-i\b|\bperl\s+-i\b|\btruncate\b|\brm\b|\bmv\b|\bcp\b|\bpython\w*\b|\bruby\b)", command)
        touches = bool(writes and re.search(r"\.prep(?:/|\\)", command))
    return touches and (root / ".prep").is_dir()


def deny(event):
    print(json.dumps({"hookSpecificOutput": {
        "hookEventName": "PreToolUse",
        "permissionDecision": "deny",
        "permissionDecisionReason": "Use prep CLI commands to change .prep records; direct edits bypass lifecycle validation."
    }}))


def allow_with_session(event):
    payload = event.get("tool_input") or {}
    command = str(payload.get("command") or "") if isinstance(payload, dict) else ""
    session = str(event.get("session_id") or "")
    if not command or not session or "PREP_SESSION=" in command:
        return
    command = (f"export PREP_SESSION={shlex.quote(session)}; "
               f"export PREP_ACTOR={ACTOR}; {command}")
    print(json.dumps({"hookSpecificOutput": {
        "hookEventName": "PreToolUse", "permissionDecision": "allow",
        "updatedInput": {"command": command}
    }}))


def activity(event, root):
    tool = str(event.get("tool_name") or "tool")
    inp = event.get("tool_input") or {}
    response = event.get("tool_response")
    if tool == "Bash" and isinstance(inp, dict) and re.search(r"(?:^|\s)(?:prep|\S*/prep)\s+", str(inp.get("command") or "")):
        return  # prep records its own CLI events
    target = ""
    area = "other"
    if isinstance(inp, dict):
        target = str(inp.get("file_path") or inp.get("path") or inp.get("command") or inp.get("pattern") or "")
    if ".prep/knowledge/" in target:
        area = "knowledge"
    elif ".prep/" in target:
        area = "issue"
    elif tool == "Bash" or target:
        area = "code"
    output = json.dumps(response, ensure_ascii=False) if response is not None else ""
    body = json.dumps({
        "kind": "tool", "op": tool, "verb": tool.lower(), "target": target[:300],
        "area": area, "chars": len(output),
        "failed": bool(event.get("error") or (isinstance(response, dict) and response.get("isError"))),
        "actor": ACTOR, "session": event.get("session_id") or "",
    }) + "\n"
    try:
        subprocess.run(["prep", "activity", "add"], cwd=root, input=body,
                       text=True, stdout=subprocess.DEVNULL,
                       stderr=subprocess.DEVNULL, timeout=3, check=False)
    except Exception:
        pass


def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else ""
    event = event_input()
    root = project_root(event)
    if root is None:
        return 0
    if mode in ("session", "prompt"):
        context(mode, event, root)
    elif mode == "before-tool":
        if protected_edit(event, root):
            deny(event)
        elif event.get("tool_name") == "Bash":
            allow_with_session(event)
    elif mode == "after-tool":
        activity(event, root)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:
        # A prep integration error must never break an ordinary Codex action.
        sys.exit(0)
