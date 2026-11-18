package cli

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"

	"github.com/charmbracelet/x/ansi"
	"github.com/creack/pty"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

const cloudPrompt = `Run prep guide %[1]s and follow it: claim the issue, implement exactly its requirement and context, and run its checks. Log "Implemented, ready for check". Do not complete the issue or check criteria. For this cloud handoff, commit the work AND prep records and push your branch, overriding the guide's local-only no-commit/no-push instruction.
Before committing, write your final report in prep handback format below into a versioned finding using prep finding %[1]s --title "Cloud handback %[1]s" --description "Cloud implementation report" --body-file <report-file>. Use prep commands, never edit record files. Do not use prep log for the full report, since it flattens Markdown, or prep handback to import a research report, since that would complete the issue. The report must travel in the records on the pushed branch; a final chat answer or local report file alone does not come back.

%[2]s`

func cmdHandoff(a *app, args []string) error {
	if len(args) == 0 {
		return usageErr("usage: prep handoff start <id>... [--yes] | fetch [--merge]")
	}
	switch args[0] {
	case "start":
		return handoffStart(a, args[1:])
	case "fetch":
		return handoffFetch(a, args[1:])
	default:
		return usageErr("unknown handoff subcommand %q; use start or fetch", args[0])
	}
}

func handoffStart(a *app, args []string) error {
	fs := flag.NewFlagSet("handoff start", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "start without confirmation")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) == 0 {
		return usageErr("usage: prep handoff start <id>... [--yes]")
	}
	unlock, err := a.lockWrite()
	if err != nil {
		return err
	}
	defer unlock()
	t, err := a.load()
	if err != nil {
		return err
	}
	var ids, prompts []string
	seen := map[string]bool{}
	for _, ref := range pos {
		id, err := t.Resolve(ref)
		if err != nil {
			return err
		}
		if err := t.BriefGate(id); err != nil {
			return err
		}
		if t.State(id) != domain.StateReady {
			return &domain.Error{Code: domain.ErrGate, Message: id + " must be ready and unclaimed; it is " + string(t.State(id))}
		}
		if seen[id] {
			return usageErr("duplicate issue %s", id)
		}
		seen[id] = true
		ids = append(ids, id)
		i := t.Issues[id]
		prompts = append(prompts, fmt.Sprintf(cloudPrompt, id, domain.ReportFormat(domain.ReportSectionsFor(i.Kind, false, len(i.Criteria) > 0))))
	}
	if err := gitx.RequirePushed(a.store.Root); err != nil {
		return err
	}
	input, terminal := a.stdin.(*os.File)
	if !terminal || !term.IsTerminal(int(input.Fd())) {
		var commands []string
		for _, prompt := range prompts {
			commands = append(commands, "claude --cloud "+shellQuote(prompt))
		}
		return &domain.Error{Code: domain.ErrGate, Message: "handoff start requires an interactive terminal; run from the repository:\n" + strings.Join(commands, "\n")}
	}
	// Confirmation goes to stderr so --json stdout remains machine-readable.
	fmt.Fprintf(a.errw, "About to start %d cloud sessions.\n", len(ids))
	if !*yes {
		fmt.Fprint(a.errw, "Continue? [y/N] ")
		answer, _ := bufio.NewReader(input).ReadString('\n')
		if answer = strings.ToLower(strings.TrimSpace(answer)); answer != "y" && answer != "yes" {
			return &domain.Error{Code: domain.ErrGate, Message: "handoff cancelled; no sessions started"}
		}
	}
	type session struct {
		Issue string `json:"issue"`
		ID    string `json:"session_id"`
		Link  string `json:"link"`
	}
	sessions := []session{}
	for n, id := range ids {
		sid, link, err := startCloud(a.store.Root, prompts[n])
		if err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
		c, err := t.PlanRecord(id, domain.OpLog, domain.RecordInput{Actor: a.actor, Now: a.now(), Text: "Cloud session: " + sid + " " + link})
		if err != nil {
			return err
		}
		if _, err := a.applyAll(t, []*domain.Change{c}); err != nil {
			return fmt.Errorf("session %s %s started but could not log it: %w", sid, link, err)
		}
		t = t.Apply(c)
		sessions = append(sessions, session{id, sid, link})
		if !a.json {
			a.printf("%s %s %s\n", id, sid, link)
		}
	}
	if a.json {
		a.emit(map[string]any{"ok": true, "sessions": sessions})
	}
	return nil
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }

func startCloud(root, prompt string) (string, string, error) {
	cmd := exec.Command("claude", "--cloud", prompt)
	cmd.Dir = root
	f, err := pty.Start(cmd)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	var output bytes.Buffer
	_, readErr := io.Copy(&output, f)
	waitErr := cmd.Wait()
	if waitErr != nil {
		return "", "", fmt.Errorf("claude --cloud: %w: %s", waitErr, strings.TrimSpace(output.String()))
	}
	// Linux PTYs signal EOF with EIO; macOS returns io.EOF.
	if readErr != nil && !errors.Is(readErr, syscall.EIO) {
		return "", "", readErr
	}
	var id, link string
	for _, line := range strings.Split(ansi.Strip(output.String()), "\n") {
		line = strings.TrimSpace(line)
		if _, value, ok := strings.Cut(line, "Created cloud session:"); ok {
			id = strings.TrimSpace(value)
		}
		if _, value, ok := strings.Cut(line, "View:"); ok {
			link = strings.TrimSpace(value)
		}
	}
	if id == "" || link == "" {
		return "", "", fmt.Errorf("claude --cloud did not print a session id and View link: %s", strings.TrimSpace(output.String()))
	}
	return id, link, nil
}

type cloudIssue struct {
	ID     string `json:"id"`
	Report bool   `json:"report"`
}

type cloudBranch struct {
	Branch string       `json:"branch"`
	Issues []cloudIssue `json:"issues"`
	Merged bool         `json:"merged"`
}

func handoffFetch(a *app, args []string) error {
	fs := flag.NewFlagSet("handoff fetch", flag.ContinueOnError)
	merge := fs.Bool("merge", false, "merge branches sequentially, format and check each")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 0 {
		return usageErr("usage: prep handoff fetch [--merge]")
	}
	// fmt takes its own store lock; do not hold another Store's lock across it.
	if err := a.open(); err != nil {
		return err
	}
	refs, err := gitx.CloudBranches(a.store.Root)
	if err != nil {
		return err
	}
	branches := []cloudBranch{}
	for _, ref := range refs {
		files, err := gitx.BranchFiles(a.store.Root, ref)
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		reports := map[string]bool{}
		for _, file := range files {
			parts := strings.Split(file, "/")
			if strings.HasPrefix(file, ".prep/issues/") && len(parts) >= 4 {
				seen[parts[2]] = true
			}
			if strings.HasPrefix(file, ".prep/knowledge/findings/") && strings.HasSuffix(file, ".md") {
				raw, err := gitx.Run(a.store.Root, "show", ref+":"+file)
				if err != nil {
					continue // A deleted finding cannot contain a report.
				}
				sections := strings.SplitN(raw, "---", 3)
				if len(sections) != 3 {
					continue
				}
				var metadata struct{ Issue, Title string }
				if yaml.Unmarshal([]byte(sections[1]), &metadata) == nil &&
					domain.ValidID(metadata.Issue) && metadata.Title == "Cloud handback "+metadata.Issue {
					seen[metadata.Issue] = true
					reports[metadata.Issue] = strings.Contains(sections[2], "\n## Summary\n")
				}
			}
		}
		ids := []string{}
		for id := range seen {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		b := cloudBranch{Branch: strings.TrimPrefix(ref, "refs/remotes/origin/"), Issues: []cloudIssue{}}
		for _, id := range ids {
			b.Issues = append(b.Issues, cloudIssue{id, reports[id]})
		}
		if !a.json {
			a.printf("%s", b.Branch)
			for _, i := range b.Issues {
				status := "missing"
				if i.Report {
					status = "present"
				}
				a.printf(" %s (report %s)", i.ID, status)
			}
			if len(b.Issues) == 0 {
				a.printf(" (no issue records)")
			}
			a.printf("\n")
		}
		branches = append(branches, b)
	}
	if *merge {
		for n, ref := range refs {
			if err := gitx.Merge(a.store.Root, ref); err != nil {
				return fmt.Errorf("%s: merge failed, stopped: %w", ref, err)
			}
			// Keep subordinate output out of the top-level JSON result.
			sub := *a
			var output bytes.Buffer
			sub.out = &output
			sub.json = false
			if err := cmdFmt(&sub, nil); err != nil {
				return fmt.Errorf("%s: prep fmt failed, stopped: %v\n%s", ref, err, output.String())
			}
			output.Reset()
			if err := cmdCheck(&sub, nil); err != nil {
				return fmt.Errorf("%s: prep check failed, stopped: %v\n%s", ref, err, output.String())
			}
			branches[n].Merged = true
		}
	}
	if a.json {
		a.emit(map[string]any{"ok": true, "branches": branches})
	}
	return nil
}
