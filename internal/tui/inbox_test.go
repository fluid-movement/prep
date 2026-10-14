package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/fluid-movement/prep/internal/domain"
)

func TestInboxApprovalSubmissionAndLateResponse(t *testing.T) {
	p, _ := sample(t)
	m := inboxModel(t, p, func(id string, _ []string) (ReviewSnapshot, error) {
		return snapshotFor(t, p, id, strings.Repeat("f", 64)), nil
	}, func(req ReviewApprovalRequest) (string, error) { return req.Snapshot.Issue.ID, nil })
	run(m, "a")
	typeIn(m, "Reviewed the captured work")
	id := m.inboxS.id
	first := m.submit()
	if first == nil || !m.inboxS.approving {
		t.Fatal("approval submission was not started")
	}
	if second := m.submit(); second != nil {
		t.Fatal("duplicate approval was submitted")
	}
	// A response from the submitted review must not close an unrelated dialog
	// opened after the person left that review.
	unrelated := &modal{kind: modalHelp}
	m.modal = unrelated
	m.handleInboxApproved(inboxApprovedMsg{id: id, err: errors.New("captured review changed")})
	if m.modal != unrelated || m.inboxS.approving {
		t.Fatal("late rejection changed the unrelated dialog or left approval locked")
	}
	if !strings.Contains(m.notice, "captured review changed") {
		t.Fatal("late rejection was not reported")
	}
	m.inboxS.approving = true
	m.handleInboxApproved(inboxApprovedMsg{id: id})
	if m.modal != unrelated || m.inboxS.approving {
		t.Fatal("late success closed unrelated dialog")
	}
}

func inboxModel(t *testing.T, p *project, review func(string, []string) (ReviewSnapshot, error), approve func(ReviewApprovalRequest) (string, error)) *Model {
	t.Helper()
	m := NewModel(testTheme(), Options{Load: p.load, Review: review, ApproveReview: approve, Actor: "human:tester"})
	m.after = func(_ time.Duration, _ func(time.Time) tea.Msg) tea.Cmd { return nil }
	m.Update(tea.WindowSizeMsg{Width: 110, Height: 28})
	run(m, "u")
	return m
}

func snapshotFor(t *testing.T, p *project, id, fingerprint string) ReviewSnapshot {
	t.Helper()
	tr, err := p.load()
	if err != nil {
		t.Fatal(err)
	}
	i := tr.Issues[id]
	if i == nil || i.Resolution == nil {
		t.Fatalf("fixture issue %s has no completion", id)
	}
	return ReviewSnapshot{Issue: i, Status: "pending", ResolutionID: i.Resolution.ContentID(), Fingerprint: fingerprint,
		WorkScopes: []string{"go.mod"}, Changes: "captured changes"}
}

func TestInboxNotificationsAreNewOnlyAndDoNotStealFocus(t *testing.T) {
	p, ids := sample(t)
	m := inboxModel(t, p, nil, nil)
	if m.notice != "" {
		t.Fatalf("initial load announced old completions: %q", m.notice)
	}
	selected := m.selected()
	tr, err := p.load()
	if err != nil {
		t.Fatal(err)
	}
	newIssue := tr.Issues[ids["format"]]
	copyIssue := *newIssue
	copyResolution := *newIssue.Resolution
	copyResolution.Note = "new completion revision"
	copyIssue.Resolution = &copyResolution
	tr.Issues[ids["format"]] = &copyIssue
	_, cmd := m.Update(loadedMsg{tree: tr})
	settle(m, cmd)
	if m.notice != "1 new completion(s) · u opens Inbox" {
		t.Fatalf("new completion notice = %q", m.notice)
	}
	if m.selected() != selected {
		t.Fatalf("completion notification changed focus from %s to %s", selected, m.selected())
	}
	m.notice = ""
	_, cmd = m.Update(loadedMsg{tree: tr})
	settle(m, cmd)
	if m.notice != "" {
		t.Fatalf("repeated load duplicated completion notice: %q", m.notice)
	}
}

func TestInboxApprovalUsesCapturedSnapshotAcrossReload(t *testing.T) {
	p, ids := sample(t)
	var approved ReviewApprovalRequest
	var calls int
	review := func(id string, _ []string) (ReviewSnapshot, error) {
		calls++
		fingerprint := strings.Repeat("a", 64)
		if calls > 1 {
			fingerprint = strings.Repeat("b", 64)
		}
		return snapshotFor(t, p, id, fingerprint), nil
	}
	m := inboxModel(t, p, review, func(req ReviewApprovalRequest) (string, error) { approved = req; return req.Snapshot.Issue.ID, nil })
	if m.screen != screenInbox || m.inboxS.snapshot == nil {
		t.Fatal("Inbox did not capture the completion review")
	}
	captured := *m.inboxS.snapshot
	run(m, "a")
	if m.modal == nil || m.modal.kind != modalReviewApprove {
		t.Fatal("a did not open approval for the captured review")
	}
	_, cmd := m.Update(m.reload()())
	settle(m, cmd)
	if m.inboxS.snapshot == nil || m.inboxS.snapshot.Fingerprint == captured.Fingerprint {
		t.Fatal("reload did not replace the current review snapshot")
	}
	if m.modal.reviewSnapshot == nil || m.modal.reviewSnapshot.Fingerprint != captured.Fingerprint {
		t.Fatal("open approval modal lost its captured snapshot during reload")
	}
	typeIn(m, "checked captured bundle")
	run(m, "enter")
	if approved.Snapshot.Fingerprint != captured.Fingerprint || approved.Reason != "checked captured bundle" {
		t.Fatalf("approval used wrong captured request: fingerprint=%q reason=%q", approved.Snapshot.Fingerprint, approved.Reason)
	}
	if m.inboxS.id != ids["format"] {
		t.Fatalf("unexpected selected issue %s", m.inboxS.id)
	}
}

func TestInboxIgnoresStaleReviewSequence(t *testing.T) {
	p, ids := sample(t)
	m := inboxModel(t, p, func(id string, _ []string) (ReviewSnapshot, error) {
		return snapshotFor(t, p, id, strings.Repeat("a", 64)), nil
	}, nil)
	id, oldSequence := m.inboxS.id, m.inboxS.sequence
	old := inboxReviewMsg{id: id, sequence: oldSequence, snapshot: snapshotFor(t, p, id, strings.Repeat("c", 64))}
	m.inboxS.sequence++
	current := snapshotFor(t, p, id, strings.Repeat("b", 64))
	m.inboxS.snapshot = &current
	_, cmd := m.Update(old)
	settle(m, cmd)
	if m.inboxS.snapshot.Fingerprint != strings.Repeat("b", 64) {
		t.Fatalf("stale sequence replaced current review for %s", ids["format"])
	}
}

func TestInboxLegacyReviewNeedsScopeAndRefreshBeforeApproval(t *testing.T) {
	p, _ := sample(t)
	var scopesSeen [][]string
	var approvals int
	review := func(id string, scopes []string) (ReviewSnapshot, error) {
		scopesSeen = append(scopesSeen, append([]string(nil), scopes...))
		fingerprint := ""
		if len(scopes) > 0 {
			fingerprint = strings.Repeat("d", 64)
		}
		return snapshotFor(t, p, id, fingerprint), nil
	}
	m := inboxModel(t, p, review, func(ReviewApprovalRequest) (string, error) { approvals++; return "ok", nil })
	run(m, "a")
	if m.modal != nil || !strings.Contains(m.notice, "Choose work scopes") {
		t.Fatal("legacy review was approvable without an explicit work scope")
	}
	run(m, "e")
	if m.modal == nil || m.modal.kind != modalReviewScopes {
		t.Fatal("e did not ask for a legacy work scope")
	}
	typeIn(m, "internal/tui/inbox.go")
	run(m, "enter")
	if len(scopesSeen) < 2 || len(scopesSeen[len(scopesSeen)-1]) != 1 || scopesSeen[len(scopesSeen)-1][0] != "internal/tui/inbox.go" {
		t.Fatalf("refreshed review did not use explicit scope: %v", scopesSeen)
	}
	if m.inboxS.snapshot == nil || m.inboxS.snapshot.Fingerprint == "" {
		t.Fatal("scope submission did not refresh the review")
	}
	run(m, "a")
	if m.modal == nil || m.modal.kind != modalReviewApprove {
		t.Fatal("refreshed scoped review did not allow opening approval")
	}
	if approvals != 0 {
		t.Fatal("opening approval implicitly approved the completion")
	}
}

func TestInboxKeysDoNotImplicitlyApprove(t *testing.T) {
	p, _ := sample(t)
	approvals := 0
	m := inboxModel(t, p, func(id string, _ []string) (ReviewSnapshot, error) {
		return snapshotFor(t, p, id, strings.Repeat("e", 64)), nil
	}, func(ReviewApprovalRequest) (string, error) { approvals++; return "ok", nil })
	run(m, "enter")
	run(m, "esc")
	run(m, "y")
	run(m, "o")
	if approvals != 0 || m.modal != nil {
		t.Fatal("enter, esc, y, or o implicitly approved a pending completion")
	}
}

func TestInboxRenderBoundsAndLayoutGoldens(t *testing.T) {
	p, _ := sample(t)
	for _, tc := range []struct {
		name   string
		w, h   int
		single bool
		golden []string
	}{{"narrow", 80, 24, false, []string{"Inbox · 3", "completion-review"}}, {"wide", 110, 28, false, []string{"Inbox · 3", "Review / attention"}}, {"one-pane", 80, 24, true, []string{"Inbox · 3", "Choose the default format"}}} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel(testTheme(), Options{Load: p.load, Single: tc.single})
			m.after = func(_ time.Duration, _ func(time.Time) tea.Msg) tea.Cmd { return nil }
			m.Update(tea.WindowSizeMsg{Width: tc.w, Height: tc.h})
			run(m, "u")
			v := ansi.Strip(m.View().Content)
			checkSize(t, v, tc.w, tc.h)
			for _, want := range tc.golden {
				if !strings.Contains(v, want) {
					t.Errorf("render lacks golden marker %q:\n%s", want, v)
				}
			}
		})
	}
}

func TestInboxReviewDocumentReportsCapturedLimitationsAndExceptions(t *testing.T) {
	p, ids := sample(t)
	snapshot := snapshotFor(t, p, ids["format"], strings.Repeat("f", 64))
	i := snapshot.Issue
	if i.Ready == nil || i.Resolution.Contract == nil || len(i.Resolution.Contract.Requirements) == 0 {
		t.Fatal("fixture is missing its ready contract")
	}
	baseline := i.Ready.Baseline
	foundBaseline := false
	for n := range i.Baselines {
		if i.Baselines[n].Name == baseline {
			i.Baselines[n].Requirement = "Original requirement at sign-off."
			foundBaseline = true
		}
	}
	if !foundBaseline {
		t.Fatal("fixture ready baseline was not found")
	}
	i.Prose = "Current edited requirement."
	i.Criteria = []domain.Criterion{{Text: "Current acceptance criterion.", Checked: false}}
	r := i.Resolution
	r.DoD = []string{"Document the change."}
	r.DoDOptOuts = []domain.OptOut{{Item: "Add release note", Reason: "Not user-facing."}}
	req := r.Contract.Requirements[0]
	r.Verification = []domain.VerificationEvidence{{ID: "verify-stale", Requirement: req.ID, Result: "fail", Method: "manual", Observation: "Observed a stale failure.", Provenance: "reported", By: "codex", At: time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC), Scopes: []string{"internal/export"}}}
	i.EvidenceCurrent = map[string]bool{"verify-stale": false}
	r.KnowledgeReviews = []domain.KnowledgeReview{{ID: "knowledge-stale", Kind: domain.KnowledgeUnchanged, Entry: "/components/cli.md", EntryRevision: strings.Repeat("a", 64), Observation: "Still accurate before the edit.", By: "codex"}}
	snapshot.KnowledgeCurrent = map[string]bool{"knowledge-stale": false}
	r.Approvals = []domain.CompletionApproval{{Kind: domain.ApprovalException, Requirement: req.ID, Reason: "Accepted this scoped exception.", By: "human:reviewer"}}
	doc := reviewDocument(snapshot)
	for _, want := range []string{
		"Original requirement at sign-off.", "Current acceptance checklist", "Current acceptance criterion.",
		"Verification policy:", "Knowledge policy:", "Document the change.", "Opt out: Add release note — Not user-facing.",
		"current scoped content: false", "Observed a stale failure.", "unchanged /components/cli.md",
		"Still accurate before the edit.", "current: false", "Exceptions and other approvals", "Accepted this scoped exception.",
		"Fingerprint `" + snapshot.Fingerprint + "`",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("captured review document lacks %q:\n%s", want, doc)
		}
	}
	baselineSection := strings.SplitN(strings.SplitN(doc, "## Current acceptance checklist", 2)[0], "## Requirement at ready signoff", 2)
	if len(baselineSection) != 2 || strings.Contains(baselineSection[1], "Current edited requirement.") {
		t.Errorf("sign-off section did not preserve the ready baseline separately from current edits:\n%s", doc)
	}
	missing := snapshot
	missing.Issue = &domain.Issue{ID: ids["format"], Title: "Legacy completion", Resolution: &domain.Resolution{Outcome: domain.OutcomeDone}}
	if got := reviewDocument(missing); !strings.Contains(got, "No captured verification reports.") || !strings.Contains(got, "No captured knowledge assessments.") {
		t.Fatalf("missing reports were not shown explicitly:\n%s", got)
	}
}

func TestInboxMouseNavigationAndWheel(t *testing.T) {
	p, ids := sample(t)
	for n := 0; n < 16; n++ {
		p.issue("Question backlog", domain.KindManual, "## Open questions\n\n- Needs an answer.", "")
	}
	m := inboxModel(t, p, func(id string, _ []string) (ReviewSnapshot, error) {
		return snapshotFor(t, p, id, strings.Repeat("9", 64)), nil
	}, nil)
	if len(m.inboxS.items) < 12 {
		t.Fatalf("fixture has only %d Inbox rows", len(m.inboxS.items))
	}
	m.View()
	_, cmd := m.Update(tea.MouseWheelMsg{X: 5, Y: m.headerH() + 5, Button: tea.MouseWheelDown})
	settle(m, cmd)
	if m.inboxS.offset == 0 {
		t.Fatal("wheel did not scroll the Inbox list")
	}
	m.inboxS.offset = 0
	m.View()
	selected := m.inboxS.id
	_, cmd = m.Update(tea.MouseClickMsg{X: 4, Y: m.headerH() + paneInner.y, Button: tea.MouseLeft})
	settle(m, cmd)
	if !m.inboxS.detail || m.inboxS.id != selected {
		t.Fatal("clicking the selected Inbox row did not open its detail")
	}
	run(m, "o")
	if m.screen != screenIssues || m.selected() != selected {
		t.Fatalf("o did not open selected issue %s", selected)
	}
	run(m, "backspace")
	if m.screen != screenInbox || m.inboxS.id != selected {
		t.Fatalf("back did not restore Inbox selection %s", selected)
	}
	if m.tree.Issues[selected] == nil || m.tree.Issues[ids["format"]] == nil {
		t.Fatal("mouse navigation lost issue data")
	}
}

func TestInboxVisualGoldens(t *testing.T) {
	p, ids := sample(t)
	for _, tc := range []struct{ width, height int }{{80, 24}, {110, 28}} {
		t.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func(t *testing.T) {
			review := func(id string, _ []string) (ReviewSnapshot, error) {
				return snapshotFor(t, p, id, strings.Repeat("1", 64)), nil
			}
			m := NewModel(testTheme(), Options{Load: p.load, Review: review})
			m.after = func(_ time.Duration, _ func(time.Time) tea.Msg) tea.Cmd { return nil }
			m.Update(tea.WindowSizeMsg{Width: tc.width, Height: tc.height})
			run(m, "u")
			inbox := ansi.Strip(m.View().Content)
			checkSize(t, inbox, tc.width, tc.height)
			golden(t, fmt.Sprintf("inbox-list-%dx%d", tc.width, tc.height), inbox)
			if m.inboxS.id != ids["format"] || m.inboxS.snapshot == nil {
				t.Fatal("review bundle was not captured for the selected completion")
			}
			m.inboxS.detail = true
			bundle := ansi.Strip(m.View().Content)
			checkSize(t, bundle, tc.width, tc.height)
			golden(t, fmt.Sprintf("inbox-bundle-%dx%d", tc.width, tc.height), bundle)
		})
	}
}
