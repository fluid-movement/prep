package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

// Turn snapshots are local UI state, never issue or knowledge records. Only
// dirty files are hashed; HEAD also lets finish notice commits made in a turn.
type hookTurn struct {
	Head  string            `json:"head"`
	Dirty map[string]string `json:"dirty"`
}

func hookGit(root string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	b, err := cmd.Output()
	return string(b), err
}

func hookTurnPath(a *app, e hookEvent) string {
	return hookStatePath(a, e, ".json")
}

// hookStatePath names this session's local hook state file. The turn snapshot
// and the last-sent prompt text are both keyed by harness and session.
func hookStatePath(a *app, e hookEvent, ext string) string {
	key := sha256.Sum256([]byte(a.actor + "\x00" + e.Session))
	return filepath.Join(a.store.Root, mdstore.Dir, "local", "hooks", hex.EncodeToString(key[:])+ext)
}

func hookSnapshot(root string) (hookTurn, error) {
	state := hookTurn{Dirty: map[string]string{}}
	head, _ := hookGit(root, "rev-parse", "--verify", "HEAD")
	state.Head = strings.TrimSpace(head)
	args := []string{"diff", "--relative", "--name-only", "--no-renames", "-z"}
	if state.Head != "" {
		args = append(args, "HEAD")
	}
	dirty, err := hookGit(root, args...)
	if err != nil {
		return state, err
	}
	extra, err := hookGit(root, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return state, err
	}
	if state.Head == "" {
		cached, err := hookGit(root, "ls-files", "--cached", "-z")
		if err != nil {
			return state, err
		}
		dirty += cached
	}
	for _, p := range strings.Split(dirty+extra, "\x00") {
		if p == "" || strings.HasPrefix(p, mdstore.Dir+"/local/") {
			continue
		}
		fingerprint, err := hookFingerprint(filepath.Join(root, p))
		if err != nil {
			return state, err
		}
		if fingerprint != "" {
			state.Dirty[p] = fingerprint
		}

	}
	return state, nil
}

func hookFingerprint(abs string) (string, error) {
	st, err := os.Lstat(abs)
	if os.IsNotExist(err) {
		return "deleted", nil
	}
	if err != nil {
		return "", err
	}
	if st.IsDir() {
		return "", nil
	}
	if st.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(abs)
		return "link:" + link, err
	}
	if st.Size() > 8<<20 {
		return fmt.Sprintf("large:%d:%d:%s", st.Size(), st.ModTime().UnixNano(), st.Mode()), nil
	}
	f, err := os.Open(abs)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)) + ":" + st.Mode().String(), nil
}

func hookSaveTurn(path string, state hookTurn) {
	if os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	b, err := json.Marshal(state)
	if err != nil {
		return
	}
	f, err := os.CreateTemp(filepath.Dir(path), "turn-*")
	if err != nil {
		return
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return
	}
	if f.Close() != nil {
		return
	}
	_ = os.Rename(f.Name(), path)
}

func hookMarkTurn(a *app, e hookEvent) {
	if e.Session == "" || e.StopActive {
		return
	}
	state, err := hookSnapshot(a.store.Root)
	if err == nil {
		hookSaveTurn(hookTurnPath(a, e), state)
	}
}

func hookFinish(a *app, e hookEvent) {
	if e.Session == "" || e.StopActive {
		return
	}
	path := hookTurnPath(a, e)
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var before hookTurn
	if json.Unmarshal(b, &before) != nil {
		return
	}
	after, err := hookSnapshot(a.store.Root)
	if err != nil {
		return
	}
	changed := map[string]bool{}
	for p, h := range after.Dirty {
		if before.Dirty[p] != h {
			changed[p] = true
		}
	}
	for p, h := range before.Dirty {
		if after.Dirty[p] != h {
			changed[p] = true
		}
	}
	if before.Head != "" && after.Head != "" && before.Head != after.Head {
		delta, err := hookGit(a.store.Root, "diff", "--relative", "--name-only", "--no-renames", "-z", before.Head, after.Head)
		if err != nil {
			return
		}
		for _, p := range strings.Split(delta, "\x00") {
			if p != "" {
				changed[p] = true
			}
		}
	}
	// Committing pre-existing dirty work changes Git status, not turn content.
	for p := range changed {
		if baseline, ok := before.Dirty[p]; ok {
			current, err := hookFingerprint(filepath.Join(a.store.Root, p))
			if err != nil {
				return
			}
			if current == baseline {
				delete(changed, p)
			}
		}
	}

	var code []string
	for p := range changed {
		if p != mdstore.Dir && !strings.HasPrefix(p, mdstore.Dir+"/") {
			code = append(code, p)
		}
	}
	sort.Strings(code)
	if len(code) == 0 {
		hookSaveTurn(path, after)
		return
	}
	entries, _, err := a.kstore.Load()
	if err != nil {
		return
	}
	tree := &domain.Tree{Knowledge: map[string]*domain.Entry{}}
	for _, entry := range entries {
		tree.Knowledge[entry.Path] = entry
	}
	tree.DeriveSuperseded()
	var decisions []string
	for _, entry := range entries {
		if entry.Type != "decision" || entry.Superseded {
			continue
		}
		matched := false
		for _, scope := range entry.Scope {
			for _, p := range code {
				if domain.ScopeMatch(scope, p) {
					matched = true
					break
				}
			}
		}
		if matched {
			decisions = append(decisions, fmt.Sprintf("%s (%s)", entry.Title, entry.ID))
		}
	}
	sort.Strings(decisions)
	if len(decisions) > 5 {
		decisions = decisions[:5]
	}
	hookSaveTurn(path, after)
	message := "Before finishing, did this turn make a decision (a choice between options, with a reason) that later sessions need? If so, record it with prep decide [<issue>] --title <title> --description <description> --body-file -; otherwise, say in one line that no decision needs recording. This check grants no additional permission: preserve the original request's scope and explicit no-tracking/no-record-write restrictions. If records cannot be changed within that scope, report the needed follow-up without writing them. Changed files: " + hookFinishList(code) + "."
	if len(decisions) > 0 {
		message += " Decisions scoped to these files: " + strings.Join(decisions, ", ") + ". Superseding one needs prep decide --supersedes <id>."
	}
	if a.actor == "claude-code" {
		hookContext(a, "Stop", message)
	} else {
		a.emit(map[string]string{"decision": "block", "reason": message})
	}
}

func hookFinishList(items []string) string {
	if len(items) > 8 {
		return strings.Join(items[:8], ", ") + ", and more"
	}
	return strings.Join(items, ", ")
}
