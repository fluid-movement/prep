package cli

import (
	"flag"
	"fmt"
	"time"

	"github.com/fluid-movement/prep/internal/activity"
)

// activityReport is prep activity report: the recorded sessions summed up.
// Without --since it reads the current stream; with it the archives too,
// and counts the events from that date on.
func activityReport(a *app, args []string) error {
	fs := flag.NewFlagSet("activity report", flag.ContinueOnError)
	sinceArg := fs.String("since", "", "count events from this date on (YYYY-MM-DD or RFC 3339), archives included")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) > 0 {
		return usageErr("usage: prep activity report [--since <date>]")
	}
	var since time.Time
	if *sinceArg != "" {
		if since, err = time.ParseInLocation("2006-01-02", *sinceArg, time.Local); err != nil {
			if since, err = time.Parse(time.RFC3339, *sinceArg); err != nil {
				return usageErr("--since takes a date as YYYY-MM-DD or RFC 3339, not %q", *sinceArg)
			}
		}
	}
	if _, err := a.load(); err != nil {
		return err
	}
	dir := a.prepDir()
	var events []activity.Event
	if since.IsZero() {
		events, err = activity.Read(dir, 0)
	} else {
		events, err = activity.ReadAll(dir)
	}
	if err != nil {
		return err
	}
	r := activity.BuildReport(events, since)
	if a.json {
		a.emit(r)
		return nil
	}
	if r.Events == 0 {
		a.printf("no activity recorded in this period\n")
		return nil
	}
	a.printf("Activity %s to %s: %d events, %d sessions, %d turns", r.From.Local().Format("2006-01-02 15:04"), r.To.Local().Format("2006-01-02 15:04"), r.Events, r.Sessions, r.Turns)
	if r.Unattributed > 0 {
		a.printf(" (%d events outside a session count under commands and hand-off only)", r.Unattributed)
	}
	a.printf("\n")

	h := r.Hooks
	a.printf("\nHooks\n")
	a.printf("  prompt text: sent %d, skipped %d\n", h.PromptsSent, h.PromptsSkipped)
	a.printf("  Stop question: asked %d, skipped %d, followed by a decision or finding %s\n", h.Stop.Asked, h.Stop.Skipped, share(h.Stop.Followed, h.Stop.Asked))
	for _, b := range h.StopByFiles {
		if b.Asked+b.Skipped == 0 {
			continue
		}
		label := b.Files + " changed files:"
		if b.Files == "unknown" {
			label = "file count not recorded:"
		}
		a.printf("    %-24s asked %d, skipped %d, followed %s\n", label, b.Asked, b.Skipped, share(b.Followed, b.Asked))
	}
	a.printf("  before-tool denials: %d (%d reads, %d writes)\n", h.DeniedReads+h.DeniedWrites, h.DeniedReads, h.DeniedWrites)
	a.printf("  edit-time decision notices: %d\n", h.EditNotices)
	a.printf("  hook errors: %d%s\n", h.Errors, nameList(h.ErrorsByHook))
	a.printf("  text added to the context by hooks: %d characters\n", h.InjectedChars)

	ru := r.Rules
	a.printf("\nRule following\n")
	a.printf("  code edits: %d, with no issue named earlier in the session %s\n", ru.Edits, share(ru.EditsNoIssue, ru.Edits))
	t := ru.Turns
	a.printf("  per turn: %d edited code; no issue named earlier in the session %s; issue named in the turn before the edit %s; knowledge lookup in the turn before the edit %s; no prep command %s\n",
		t.Edited, share(t.NoIssue, t.Edited), share(t.IssueInTurn, t.Edited), share(t.KnowledgeInTurn, t.Edited), share(t.NoPrep, t.Edited))
	s := ru.Sessions
	a.printf("  per session: %d edited code; an edit before any issue was named %s; knowledge lookup before the first edit %s; guide then show in %s\n",
		s.Edited, share(s.NoIssue, s.Edited), share(s.Knowledge, s.Edited), share(s.GuidesThenShow, s.Guides))
	i := ru.Issues
	a.printf("  per issue: %d had code edits; knowledge lookup before the first edit %s; guide then show for %s\n",
		i.Edited, share(i.Knowledge, i.Edited), share(i.GuidesThenShow, i.Guided))
	a.printf("  guides followed by show of the same issue: %s\n", share(ru.GuidesThenShow, ru.Guides))

	ho := r.Handoff
	a.printf("\nHand-off\n")
	a.printf("  briefs: %d\n", ho.Briefs)
	a.printf("  hand-backs: accepted %d, refused %d%s\n", ho.Accepted, ho.Refused, nameList(ho.RefusedBy))
	a.printf("  failed checks (accepted hand-backs with no completion after): %d\n", ho.FailedChecks)

	a.printf("\nCommands\n")
	if len(r.Commands) == 0 {
		a.printf("  none recorded\n")
	}
	for _, c := range r.Commands {
		line := fmt.Sprintf("  %-16s %5d", c.Command, c.Count)
		if c.Failed > 0 {
			line += fmt.Sprintf("  %d failed", c.Failed)
		}
		a.printf("%s\n", line)
	}
	if len(r.Failures) > 0 {
		a.printf("  failures by command and error code:\n")
		for _, f := range r.Failures {
			a.printf("    %-16s %-22s %d\n", f.Command, f.Code, f.Count)
		}
	}

	c := r.Cost
	a.printf("\nCost\n")
	a.printf("  request events in %d of %d sessions", c.SessionsWithRequests, c.Sessions)
	if c.Subagents > 0 {
		a.printf(" and %d of %d subagents", c.SubagentsWithRequests, c.Subagents)
	}
	a.printf("\n")
	if c.SessionsWithRequests > 0 {
		tk := c.Tokens
		a.printf("  reported: $%.2f, %d requests, tokens in %d, out %d, cache read %d, cache write %d\n", c.CostUSD, c.Requests, tk.Input, tk.Output, tk.CacheRead, tk.CacheWrite)
	}
	if missing := c.Sessions - c.SessionsWithRequests; missing > 0 || c.Subagents > c.SubagentsWithRequests {
		a.printf("  no figure for sessions and subagents without request events (%d and %d): only the Claude Code main session reports requests, and nothing is estimated\n", missing, c.Subagents-c.SubagentsWithRequests)
	}
	for _, ac := range c.ByActor {
		a.printf("    %-22s request events in %d of %d sessions\n", ac.Actor, ac.SessionsWithRequests, ac.Sessions)
	}
	return nil
}

// share renders n of total with its percentage.
func share(n, total int) string {
	if total == 0 {
		return fmt.Sprintf("%d of 0", n)
	}
	return fmt.Sprintf("%d of %d (%d%%)", n, total, (n*100+total/2)/total)
}

// nameList renders counts by name as a parenthesized list, or nothing.
func nameList(counts []activity.NameCount) string {
	if len(counts) == 0 {
		return ""
	}
	out := " ("
	for k, c := range counts {
		if k > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%s %d", c.Name, c.Count)
	}
	return out + ")"
}
