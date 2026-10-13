package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/userconfig"
)

const hookReminder = "Prep project: before code or research, search `prep list --text <distinctive words>` for existing work, then `prep next`. Use `prep guide <id>` before acting; create an issue only when no existing one covers authorized work. Resolve ambiguous references before changing records. Read tracker facts with `prep show` or `prep list`, not guesses from titles. Read knowledge through `prep knowledge find|show`; update affected knowledge after code changes. Use prep commands for record changes. Follow the user's scope; ordinary questions and work outside this project can proceed normally."

type hookEvent struct {
	Cwd        string          `json:"cwd"`
	Session    string          `json:"session_id"`
	Prompt     string          `json:"prompt"`
	Tool       string          `json:"tool_name"`
	Input      map[string]any  `json:"tool_input"`
	Response   json.RawMessage `json:"tool_response"`
	Error      json.RawMessage `json:"error"`
	Event      string          `json:"hook_event_name"`
	StopActive bool            `json:"stop_hook_active"`
}

var hookIDs = regexp.MustCompile(`(?i)\b[0-7][0-9A-HJKMNP-TV-Z]{25}\b`)
var hookPaths = regexp.MustCompile(`"[^"]*"|'[^']*'|[^\s"'` + "`" + `;|&<>()]+`)
var hookRead = regexp.MustCompile(`(^|[\s;|&()])(?:\S*/)?(?:cat|sed|head|tail|grep|rg|ls|find|awk|less|more|bat)\b`)
var hookWrite = regexp.MustCompile(`>|(^|[\s;|&()])(?:\S*/)?(?:tee|rm|mv|cp|truncate|touch|mkdir|rmdir|python\w*|ruby|perl)\b|\bsed\s+[^;\n]*-i`)
var hookPrepCommand = regexp.MustCompile(`^(?:[A-Za-z_][A-Za-z0-9_]*=\S+\s+)*(?:[^\s]+/)?prep\s+`)

func cmdHook(a *app, args []string) error {
	fs := flag.NewFlagSet("hook", flag.ContinueOnError)
	harness := fs.String("harness", "", "claude-code or codex")
	plugin := fs.String("plugin", "", "plugin directory for version checks")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("usage: prep hook session-start|prompt|finish|before-tool|after-tool --harness claude-code|codex")
	}
	mode := pos[0]
	names := map[string]string{"session-start": "SessionStart", "prompt": "UserPromptSubmit", "finish": "Stop", "before-tool": "PreToolUse", "after-tool": "PostToolUse"}
	name, ok := names[mode]
	if !ok || (*harness != "claude-code" && *harness != "codex") {
		return usageErr("unknown hook mode or harness")
	}
	level := "small"
	if mode == "prompt" {
		cfg, err := userconfig.Load()
		if err != nil {
			return nil
		}
		if cfg.Hooks.Prompt != "" {
			level = cfg.Hooks.Prompt
		}
		if level == "off" {
			return nil
		}
		if level != "small" && level != "large" {
			return nil
		}
	}
	var e hookEvent
	// Bad or oversized payloads fail open. Reject trailing input as well.
	b, err := io.ReadAll(io.LimitReader(a.stdin, (1<<20)+1))
	if err != nil || len(b) > 1<<20 || json.Unmarshal(b, &e) != nil {
		return nil
	}
	if e.Cwd != "" {
		a.root = e.Cwd
	}
	if e.Cwd == "" {
		e.Cwd = a.root
		if e.Cwd == "" {
			e.Cwd, _ = os.Getwd()
		}
	}
	if err := a.open(); err != nil {
		return nil
	}
	a.actor, a.named, a.json = *harness, true, false
	a.errw = io.Discard
	switch mode {
	case "session-start":
		hookMarkTurn(a, e)
		message := hookCapture(a, func(c *app) error { return cmdPrime(c, []string{"--hook", "--max", "5", "--plugin", *plugin}) })
		if message == "" {
			return nil
		}
		if *harness == "codex" {
			message += "\nFor shell commands in this session, set PREP_ACTOR=codex and PREP_SESSION=" + hookQuote(e.Session) + " so prep CLI events carry attribution.\n"
		}
		hookContext(a, name, hookBound(message, 10000))
	case "prompt":
		hookMarkTurn(a, e)
		message := hookReminder
		if level == "large" {
			seen := map[string]bool{}
			for _, id := range hookIDs.FindAllString(strings.ToUpper(e.Prompt), -1) {
				if seen[id] {
					continue
				}
				seen[id] = true
				if len(seen) > 2 {
					break
				}
				guide := hookCapture(a, func(c *app) error { return cmdGuide(c, []string{id}) })
				if guide != "" {
					message += "\n\n" + hookBound(guide, 4000)
				}
			}
		}
		hookContext(a, name, hookBound(message, 10000))
	case "before-tool":
		if reason := hookRedirect(e, a.store.Root); reason != "" {
			a.emit(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": name, "permissionDecision": "deny", "permissionDecisionReason": reason}})
		}
	case "after-tool":
		hookActivity(a, e)
	case "finish":
		hookFinish(a, e)
	}
	return nil
}

func hookCapture(a *app, run func(*app) error) string {
	c := *a
	var buf bytes.Buffer
	c.out, c.errw, c.event = &buf, io.Discard, nil
	if run(&c) != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

func hookContext(a *app, event, message string) {
	if message != "" {
		a.emit(map[string]any{"hookSpecificOutput": map[string]string{"hookEventName": event, "additionalContext": message}})
	}
}

func hookBound(s string, max int) string {
	if len(s) <= max {
		return s
	}
	i := max - 100
	for i > 0 && !utf8.RuneStart(s[i]) {
		i--
	}
	return s[:i] + "\n[Hook context truncated; run prep guide <id> for the full current guide.]"
}

func hookQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }

func hookString(m map[string]any, key string) string { s, _ := m[key].(string); return s }

// Identify supported paths, including dot segments and absolute paths. This
// is a guardrail for common local tools, not a shell interpreter or sandbox.
func hookArea(target, cwd, root string) string {
	// File-tool inputs are a complete path, which can contain spaces.
	if area := hookRecordArea(target, cwd, root); area != "" {
		return area
	}
	for _, token := range hookPaths.FindAllString(strings.ReplaceAll(target, "\\", "/"), -1) {
		if area := hookRecordArea(token, cwd, root); area != "" {
			return area
		}
	}
	return ""
}

func hookRecordArea(p, cwd, root string) string {
	p = strings.Trim(p, "\"'")
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}
	rel, err := filepath.Rel(root, filepath.Clean(p))
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	if rel == mdstore.Dir+"/knowledge" || strings.HasPrefix(rel, mdstore.Dir+"/knowledge/") {
		return activity.AreaKnowledge
	}
	if rel == mdstore.Dir || rel == mdstore.Dir+"/issues" || strings.HasPrefix(rel, mdstore.Dir+"/issues/") || rel == mdstore.Dir+"/project.md" || rel == mdstore.Dir+"/config.yaml.dist" {
		return activity.AreaIssue
	}
	return ""
}

func hookRedirect(e hookEvent, root string) string {
	cwd := e.Cwd
	if cwd == "" {
		cwd = root
	}
	target := hookString(e.Input, "file_path")
	if target == "" {
		target = hookString(e.Input, "path")
	}
	if e.Tool == "Grep" || e.Tool == "Glob" {
		target += " " + hookString(e.Input, "pattern")
	}
	command := hookString(e.Input, "command")
	if e.Tool == "apply_patch" {
		target = ""
		for _, match := range regexp.MustCompile(`(?m)^\*\*\* (?:Add|Update|Delete) File: (.+)$|^\*\*\* Move to: (.+)$`).FindAllStringSubmatch(command, -1) {
			target += " " + match[1] + match[2]
		}
	}
	area := hookArea(target, cwd, root)
	write := e.Tool == "Edit" || e.Tool == "Write" || e.Tool == "NotebookEdit" || e.Tool == "apply_patch"
	if e.Tool == "Bash" {
		// Check each segment: a prep call must not exempt a later direct read
		// or write (the earlier Python integration did exempt the whole call).
		for _, segment := range strings.FieldsFunc(command, func(r rune) bool { return r == ';' || r == '|' || r == '&' || r == '\n' }) {
			found := hookArea(segment, cwd, root)
			if found == "" {
				continue
			}
			if hookWrite.MatchString(segment) {
				return "Use prep CLI write commands (prep edit, context, criterion, knowledge update, etc.) to change .prep records; direct edits bypass validation."
			}
			if hookRead.MatchString(segment) {
				area = found
			}
		}
	} else if !write && e.Tool != "Read" && e.Tool != "Grep" && e.Tool != "Glob" {
		return ""
	}
	if area == "" {
		return ""
	}
	if write {
		return "Use prep CLI write commands to change .prep records (prep edit, context, criterion, knowledge update, etc.); direct edits bypass validation."
	}
	if area == activity.AreaKnowledge {
		return "Read knowledge with prep knowledge find <words> or prep knowledge show <entry>[#section], never the files. Use --outline first for long entries."
	}
	return "Read issue records with prep show <id> and prep guide <id>; search with prep list --text <words>. Use prep links or prep views for project configuration."
}

func hookActivity(a *app, e hookEvent) {
	if e.Tool == "" {
		return
	}
	command := hookString(e.Input, "command")
	if e.Tool == "Bash" && hookOnlyPrep(command) {
		return
	}
	target := hookString(e.Input, "file_path")
	for _, key := range []string{"path", "command", "pattern", "url", "query"} {
		if target == "" {
			target = hookString(e.Input, key)
		}
	}
	area := hookArea(target, e.Cwd, a.store.Root)
	if area == "" {
		area = activity.AreaOther
		if e.Tool == "Bash" || target != "" {
			area = activity.AreaCode
		}
	}
	failed := len(e.Error) > 0 && string(e.Error) != "null" && string(e.Error) != "false" || e.Event == "PostToolUseFailure"
	var response struct {
		IsError  bool `json:"isError"`
		ExitCode int  `json:"exit_code"`
	}
	_ = json.Unmarshal(e.Response, &response)
	failed = failed || response.IsError || response.ExitCode != 0
	_ = activity.Append(filepath.Join(a.store.Root, mdstore.Dir), activity.Event{At: a.now().UTC(), Actor: a.actor, Session: e.Session, Kind: activity.KindTool, Op: e.Tool, Verb: strings.ToLower(e.Tool), Target: hookBound(target, 300), Area: area, Chars: len(e.Response), Failed: failed})
}

func hookOnlyPrep(command string) bool {
	segments := strings.FieldsFunc(command, func(r rune) bool { return r == ';' || r == '|' || r == '&' || r == '\n' })
	found := false
	for _, s := range segments {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "export PREP_") {
			continue
		}
		if !hookPrepCommand.MatchString(s) {
			return false
		}
		found = true
	}
	return found
}
