package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

// hookScopedMax is how many decisions one before-tool call names.
const hookScopedMax = 5

var hookPatchFiles = regexp.MustCompile(`(?m)^\*\*\* (?:Add|Update|Delete) File: (.+)$|^\*\*\* Move to: (.+)$`)

// hookScopedContext names, before a file edit, the current decisions scoped
// to the edited files that this session has not been shown yet. It adds
// context only: the edit goes through the normal permission flow.
func hookScopedContext(a *app, e hookEvent, event string) {
	paths := hookEditPaths(e, a.store.Root)
	if len(paths) == 0 {
		return
	}
	decisions, err := hookScopedDecisions(a, paths)
	if err != nil {
		hookFail(a, e, "before-tool", "knowledge", err)
		return
	}
	shown := hookShown(a, e)
	var fresh []*domain.Entry
	for _, d := range decisions {
		if !shown[d.ID] {
			fresh = append(fresh, d)
		}
	}
	if len(fresh) == 0 {
		return
	}
	more := 0
	if len(fresh) > hookScopedMax {
		more = len(fresh) - hookScopedMax
		fresh = fresh[:hookScopedMax]
	}
	var names, ids []string
	for _, d := range fresh {
		names = append(names, fmt.Sprintf("%s (%s)", d.Title, d.ID))
		ids = append(ids, d.ID)
	}
	message := "Decisions scoped to " + hookFinishList(paths) + ": " + strings.Join(names, "; ")
	if more > 0 {
		message += fmt.Sprintf("; %d more match (prep knowledge list --type decision --scope %s)", more, paths[0])
	}
	message = hookBound(message+". Read one with prep knowledge show <id> before changing what it covers; replacing one needs prep decide --supersedes <id>.", 2000)
	hookMarkShown(a, e, ids)
	hookRecordChars(a, e, "before-tool", "context", activity.AreaCode, hookFinishList(paths), utf8.RuneCountInString(message))
	hookContext(a, event, message)
}

// hookEditPaths lists the files a file-editing tool call targets, as
// root-relative slash paths. Paths outside the project and under .prep are
// left out; other tools have none.
func hookEditPaths(e hookEvent, root string) []string {
	if e.Tool != "Edit" && e.Tool != "Write" && e.Tool != "NotebookEdit" && e.Tool != "apply_patch" {
		return nil
	}
	var targets []string
	for _, key := range []string{"file_path", "notebook_path", "path"} {
		if s := hookString(e.Input, key); s != "" {
			targets = append(targets, s)
			break
		}
	}
	// apply_patch carries a patch, not a path.
	for _, match := range hookPatchFiles.FindAllStringSubmatch(hookString(e.Input, "command"), -1) {
		targets = append(targets, strings.TrimSpace(match[1]+match[2]))
	}
	cwd := e.Cwd
	if cwd == "" {
		cwd = root
	}
	var paths []string
	seen := map[string]bool{}
	for _, p := range targets {
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, p)
		}
		rel, err := filepath.Rel(root, filepath.Clean(p))
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") || rel == mdstore.Dir || strings.HasPrefix(rel, mdstore.Dir+"/") || seen[rel] {
			continue
		}
		seen[rel] = true
		paths = append(paths, rel)
	}
	return paths
}

// hookScopedDecisions returns the current decisions whose scope matches one
// of the root-relative paths: the most specific scope first, then the newest.
func hookScopedDecisions(a *app, paths []string) ([]*domain.Entry, error) {
	entries, _, err := a.kstore.Load()
	if err != nil {
		return nil, err
	}
	tree := &domain.Tree{Knowledge: map[string]*domain.Entry{}}
	for _, entry := range entries {
		tree.Knowledge[entry.Path] = entry
	}
	tree.DeriveSuperseded()
	var decisions []*domain.Entry
	specific := map[*domain.Entry]int{}
	for _, entry := range entries {
		if entry.Type != "decision" || entry.Superseded {
			continue
		}
		best := -1
		for _, scope := range entry.Scope {
			for _, p := range paths {
				if n := hookScopeSpecificity(scope); n > best && domain.ScopeMatch(scope, p) {
					best = n
				}
			}
		}
		if best >= 0 {
			decisions = append(decisions, entry)
			specific[entry] = best
		}
	}
	sort.SliceStable(decisions, func(i, j int) bool {
		x, y := decisions[i], decisions[j]
		if specific[x] != specific[y] {
			return specific[x] > specific[y]
		}
		// Record IDs are ULIDs, so the larger one is the newer one.
		return x.ID > y.ID
	})
	return decisions, nil
}

// hookScopeSpecificity is the number of literal characters in a scope
// pattern: src/cart.go is more specific than src/*.go, and ** the least.
func hookScopeSpecificity(scope string) int {
	scope = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(scope), "./"), "/")
	return len(scope) - strings.Count(scope, "*") - strings.Count(scope, "?") - strings.Count(scope, "[") - strings.Count(scope, "]")
}

// hookShownPath names the file listing the decisions this context has been
// shown. A subagent has its own context, so its agent id is part of the key.
func hookShownPath(a *app, e hookEvent) string {
	if e.AgentID != "" {
		e.Session += "\x00" + e.AgentID
	}
	return hookStatePath(a, e, ".shown")
}

// hookShown reads the shown decision IDs. Without a session nothing can be
// keyed, and a missing or unreadable record means nothing was shown.
func hookShown(a *app, e hookEvent) map[string]bool {
	shown := map[string]bool{}
	if e.Session == "" {
		return shown
	}
	b, _ := os.ReadFile(hookShownPath(a, e))
	for _, id := range strings.Fields(string(b)) {
		shown[id] = true
	}
	return shown
}

func hookMarkShown(a *app, e hookEvent, ids []string) {
	if e.Session == "" {
		return
	}
	path := hookShownPath(a, e)
	if os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(strings.Join(ids, "\n") + "\n")
}

// hookForgetShown makes the next edit name its decisions again, for a context
// that starts over (session start, compaction).
func hookForgetShown(a *app, e hookEvent) {
	if e.Session != "" {
		_ = os.Remove(hookShownPath(a, e))
	}
}
