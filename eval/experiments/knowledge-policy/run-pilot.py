#!/usr/bin/env python3
"""Run the bounded Codex knowledge-policy pilot against fresh fixture copies."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shlex
import shutil
import subprocess
import time


CONDITIONS = ("code-only", "current", "lean", "stale")
PROMPT = (
    "the headline on the player page gets cut off after 80 characters, i think "
    "the database column is too small. fix it.\n\n"
    "Use the already-tracked issue. Follow the project's current prep workflow "
    "and completion gates, including truthful evidence and an explicit "
    "knowledge review report. Do not invent approval or claim checks you did not "
    "perform. Read and follow `.agents/skills/prep/SKILL.md`, which is identical "
    "in every condition. Treat knowledge entries as guidance and inspect the "
    "current code before changing it. Run a relevant check and report what you "
    "actually ran."
)


def command(args, *, cwd, env=None, check=True):
    return subprocess.run(args, cwd=cwd, env=env, text=True, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, check=check)


def error_category(stderr):
    text = (stderr or "").lower()
    if any(term in text for term in ("not logged in", "not authenticated", "unauthorized", "authentication")):
        return "authentication unavailable"
    if "approval" in text or "approve-for-me" in text:
        return "approval routing unavailable"
    if "model" in text and any(term in text for term in ("unsupported", "not found", "unknown")):
        return "model unavailable"
    if "sandbox" in text or "permission denied" in text or "operation not permitted" in text:
        return "sandbox permission failure"
    if not text.strip():
        return "no stderr details"
    return "other launch/runtime error"


def knowledge_manifest(project, binary, prep_session):
    prefix = ["env", "PREP_ACTOR=codex", f"PREP_SESSION={prep_session}", str(binary)]
    listed = command(prefix + ["knowledge", "list", "--json"], cwd=project, check=False)
    if listed.returncode != 0:
        return []
    data = json.loads(listed.stdout)
    entries = []
    for item in data.get("entries", []):
        shown = command(prefix + ["knowledge", "show", item["path"]], cwd=project, check=False)
        content = shown.stdout if shown.returncode == 0 else ""
        entries.append({"path": item["path"], "sha256_of_cli_show": hashlib.sha256(content.encode()).hexdigest(),
                        "bytes": len(content.encode()), "read_ok": shown.returncode == 0})
    return entries


def command_trace(stdout, project):
    counts = {}
    reads = []
    executions = []

    def category(segment):
        segment = re.sub(
            r"^(?:(?:PREP_ACTOR|PREP_SESSION)=(?:'[^']*'|\"[^\"]*\"|\S+)\s+)*",
            "", segment.strip())
        for label, prefix in (
                ("claim", "prep claim"),
                ("knowledge_find", "prep knowledge find"),
                ("knowledge_show", "prep knowledge show"),
                ("knowledge_review_record", "prep knowledge review record"),
                ("evidence_record", "prep evidence record"),
                ("complete", "prep complete"),
                ("prep_check", "prep check")):
            if segment == prefix or segment.startswith(prefix + " "):
                return label, segment
        if segment == "./test.sh" or segment.startswith("./test.sh "):
            return "test_script", segment
        if segment == "sh test.sh" or segment.startswith("sh test.sh "):
            return "test_script", segment
        if segment.startswith("node --test ") or segment.startswith("node tests/"):
            return "node_test", segment
        if segment.startswith("node --input-type=module") or segment.startswith("node <<"):
            return "node_runtime_check", segment
        if segment.startswith("node ") and "--version" not in segment and " -v" not in segment:
            return "node_runtime_check", segment
        if segment == "node":
            return "node_runtime_check", segment
        if segment.startswith("python3 ") or segment == "python3":
            return "python_check", segment
        if segment.startswith("sqlite3 "):
            return "sqlite_check", segment
        if segment.startswith("git diff --check"):
            return "diff_check", segment
        return None, None

    def single_shell_command(script):
        # Heredoc bodies are program text, not shell separators. The fixture
        # checks use a single node command with a JS heredoc.
        heredoc = re.match(r"\s*(?:node|python3|sqlite3)\s+[^\n]*<<\s*(['\"]?)([A-Za-z_][A-Za-z0-9_]*)\1", script)
        if heredoc:
            return True
        # Treat multiline shell scripts conservatively: shlex otherwise
        # discards newlines that can separate distinct shell commands.
        if "\n" in script.strip():
            return False
        lexer = shlex.shlex(script, posix=True, punctuation_chars=";&|")
        try:
            return not any(token in {";", "&&", "||", "&", "|"} for token in lexer)
        except ValueError:
            return False

    for line in stdout.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        item = event.get("item") or {}
        if event.get("type") != "item.completed" or item.get("type") != "command_execution":
            continue
        command_text = item.get("command", "")
        try:
            outer = shlex.split(command_text)
            script = outer[outer.index("-lc") + 1] if "-lc" in outer else command_text
        except (ValueError, IndexError):
            script = command_text
        normalized_script = re.sub(
            r"^(?:(?:PREP_ACTOR|PREP_SESSION)=(?:'[^']*'|\"[^\"]*\"|\S+)\s+)*",
            "", script.strip())
        if not single_shell_command(script):
            continue
        label, normalized = category(normalized_script)
        if not label:
            continue
        key = f"{label}:{item.get('exit_code')}"
        counts[key] = counts.get(key, 0) + 1
        if label == "knowledge_show":
            try:
                tokens = shlex.split(normalized)
                reads.append(tokens[3])
            except (ValueError, IndexError):
                pass
        if label in {"test_script", "node_test", "node_runtime_check", "python_check", "sqlite_check", "prep_check", "diff_check"}:
            executions.append({
                "kind": label,
                "command": normalized,
                "shell_command": re.sub(
                    r"(PREP_SESSION=)(?:'[^']*'|\"[^\"]*\"|\S+)",
                    r"\1[redacted]", command_text),
                "outer_shell_exit_code": item.get("exit_code"),
                "output_excerpt": item.get("aggregated_output", "")[:500],
            })
    return {"completed_command_counts": counts,
            "knowledge_show_paths": sorted(set(reads)),
            "observed_check_commands": executions}


def outcome(project, binary, issue_id, prep_session):
    env = os.environ.copy()
    env["PATH"] = str(project / ".trial-bin") + os.pathsep + env.get("PATH", "")
    attribution = ["env", "PREP_ACTOR=codex", f"PREP_SESSION={prep_session}"]
    show = command(attribution + [str(binary), "show", issue_id, "--json"], cwd=project, env=env, check=False)
    review = command(attribution + [str(binary), "knowledge", "review", "show", issue_id, "--json"],
                     cwd=project, env=env, check=False)
    show_json = json.loads(show.stdout) if show.returncode == 0 else {}
    review_json = json.loads(review.stdout) if review.returncode == 0 else {}
    schema = (project / "db/schema.sql").read_text(errors="replace")
    profile = (project / "src/lib/profile.ts").read_text(errors="replace")
    headline_line = next((line for line in schema.splitlines()
                          if re.match(r"\s*headline\b", line, re.I)), "")
    width = max((int(v) for v in re.findall(r"VARCHAR\s*\(\s*(\d+)\s*\)", headline_line, re.I)), default=0)
    storage_ok = width >= 200 or bool(re.search(r"\bheadline\s+TEXT\b", headline_line, re.I))
    slices = [int(v) for v in re.findall(r"\.slice\(\s*0\s*,\s*(\d+)\s*\)", profile)]
    truncation = bool(re.search(r"\.(?:slice|substring)\(\s*0\s*,\s*(\d+)\s*\)", profile))
    paths = command(["git", "status", "--short"], cwd=project, check=False).stdout.splitlines()
    knowledge_diff = command(["git", "diff", "HEAD", "--numstat", "--", ".prep/knowledge"],
                             cwd=project, check=False).stdout.splitlines()
    knowledge_reviews = [{"entry": r.get("review", {}).get("entry"),
                          "kind": r.get("review", {}).get("kind"),
                          "applicable": r.get("applicable", False)}
                         for r in review_json.get("reviews", [])]
    evidence = command(attribution + [str(binary), "evidence", "show", issue_id],
                       cwd=project, env=env, check=False)
    evidence_records = []
    for result, method, applicable in re.findall(
            r"\b(pass|fail)\s+\[([a-z-]+),\s+reported\]\s+applicable=(true|false)", evidence.stdout):
        evidence_records.append({"result": result, "method": method, "applicable": applicable == "true"})
    return {
        "issue_state": show_json.get("state"),
        "stale": show_json.get("stale"),
        "evidence_pending": show_json.get("evidence_pending"),
        "resolution_review_status": (show_json.get("resolution_review") or {}).get("status"),
        "knowledge_reviews": knowledge_reviews,
        "evidence_reports": evidence_records,
        "evidence_report_query_ok": evidence.returncode == 0,
        "criterion_proxy_pass": storage_ok and not truncation,
        "criterion_proxy_kind": "static schema/formatter inspection; no browser render",
        "schema_width": width,
        "render_slice_limits": slices,
        "changed_paths": [p[3:] for p in paths if len(p) >= 4],
        "knowledge_diff_numstat": knowledge_diff,
        "queries_ok": show.returncode == 0 and review.returncode == 0,
    }


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--source", type=Path, required=True, help="clean pinned source extraction")
    ap.add_argument("--prep-binary", type=Path, required=True)
    ap.add_argument("--plugin-skill", type=Path, required=True)
    ap.add_argument("--out", type=Path, required=True, help="private temporary result root")
    ap.add_argument("--repetitions", type=int, default=3)
    ap.add_argument("--conditions", default=",".join(CONDITIONS),
                    help="comma-separated condition subset, default all four")
    ap.add_argument("--timeout", type=int, default=900)
    ap.add_argument("--codex", default="codex")
    ap.add_argument("--model", default=None,
                    help="model to request explicitly (otherwise the configured default)")
    ap.add_argument("--skill-revision", default="65067e922a4962c7a8d44863843c1ac8a62e43ff")
    opt = ap.parse_args()
    source, binary, skill, out = map(lambda p: p.resolve(),
                                     (opt.source, opt.prep_binary, opt.plugin_skill, opt.out))
    out.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env["GOCACHE"] = str(out / "go-cache")
    version = command([opt.codex, "--version"], cwd=source, env=env).stdout.strip()
    doctor = json.loads(command([opt.codex, "doctor", "--json"], cwd=source, env=env).stdout)
    checks = doctor.get("checks", {})
    config_check = checks.get("config.load", {}) if isinstance(checks, dict) else next(
        (check for check in checks if isinstance(check, dict) and check.get("name") == "config.load"), {})
    configured_model = (config_check.get("details") or {}).get("model")
    if not configured_model:
        raise RuntimeError("Codex doctor did not report a configured model; refusing unlabelled runs")
    model_requested = opt.model or configured_model
    summary = {"source_commit": "bacc2245668c64831389dced173514b2db6000df",
               "codex_cli_version": version, "model_requested": model_requested,
               "skill_revision": opt.skill_revision,
               "skill_sha256": hashlib.sha256(skill.read_bytes()).hexdigest(),
               "prompt": PROMPT, "runs": []}
    selected = tuple(c.strip() for c in opt.conditions.split(",") if c.strip())
    if not selected or any(c not in CONDITIONS for c in selected):
        ap.error("--conditions must name code-only,current,lean,stale")
    for condition in selected:
        for rep in range(1, opt.repetitions + 1):
            run_id = f"{condition}-{rep}"
            workspace_root = out / "workspaces" / run_id
            proc = command(["go", "run", "./eval/experiments/knowledge-policy/build-fixture.go",
                            "--condition", condition, "--out", str(workspace_root),
                            "--binary", str(binary), "--spec", "eval/fixtures/webapp.yaml"],
                           cwd=source, env=env)
            match = re.search(r"headline_issue=([A-Z0-9]+)", proc.stdout)
            if not match:
                raise RuntimeError(f"fixture builder did not return issue id for {run_id}")
            issue_id = match.group(1)
            project = workspace_root / "webapp"
            (project / ".trial-bin").mkdir(exist_ok=True)
            prep_session = f"knowledge-policy-{run_id}"
            wrapper = project / ".trial-bin/prep"
            wrapper.write_text("#!/bin/sh\nexec env PREP_ACTOR=codex PREP_SESSION=" + prep_session +
                               " " + str(binary) + " \"$@\"\n")
            wrapper.chmod(0o755)
            skill_dir = project / ".agents/skills/prep"
            skill_dir.mkdir(parents=True, exist_ok=True)
            shutil.copy2(skill, skill_dir / "SKILL.md")
            initial_knowledge = knowledge_manifest(project, binary, prep_session)
            run_dir = out / "sessions" / run_id
            run_dir.mkdir(parents=True, exist_ok=True)
            event_path = run_dir / "events.jsonl"
            answer_path = run_dir / "last-message.txt"
            run_env = os.environ.copy()
            run_env["PATH"] = str(project / ".trial-bin") + os.pathsep + run_env.get("PATH", "")
            cmd = [opt.codex, "exec", "--ephemeral", "--json", "--model", model_requested,
                   "--sandbox", "workspace-write", "--cd", str(project),
                   "--output-last-message", str(answer_path), PROMPT]
            started = time.monotonic()
            model = None
            usage = None
            timed_out = False
            with event_path.open("w") as events:
                child = subprocess.Popen(cmd, cwd=project, env=run_env, text=True,
                                         stdout=subprocess.PIPE, stderr=subprocess.PIPE, bufsize=1)
                try:
                    stdout, stderr = child.communicate(timeout=opt.timeout)
                except subprocess.TimeoutExpired:
                    timed_out = True
                    child.kill()
                    stdout, stderr = child.communicate()
                code = child.returncode
                events.write(stdout)
                for line in stdout.splitlines():
                    try:
                        event = json.loads(line)
                    except json.JSONDecodeError:
                        continue
                    for key, value in event.items():
                        if key == "model" and isinstance(value, str):
                            model = value
                    if event.get("type") == "turn.completed" and isinstance(event.get("usage"), dict):
                        usage = event["usage"]
            elapsed = round(time.monotonic() - started, 2)
            result = outcome(project, binary, issue_id, prep_session)
            size = answer_path.stat().st_size if answer_path.exists() else 0
            summary["runs"].append({
                "run_id": run_id, "condition": condition, "repetition": rep,
                "issue_id": issue_id, "model_requested": model_requested,
                "model_reported": model, "exit_code": code,
                "error_category": error_category(stderr) if code != 0 else None,
                "timed_out": timed_out, "elapsed_seconds": elapsed,
                "usage_reported": usage, "last_message_bytes": size,
                "initial_knowledge": initial_knowledge, **command_trace(stdout, project), **result,
            })
            (out / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
            print(f"{run_id}: exit={code} timeout={timed_out} elapsed={elapsed}s model_requested={model_requested} model_event={model or 'unreported'} state={result['issue_state']} criterion_proxy={result['criterion_proxy_pass']} error={error_category(stderr) if code != 0 else 'none'}", flush=True)


if __name__ == "__main__":
    main()
