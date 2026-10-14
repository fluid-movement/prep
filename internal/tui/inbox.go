package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

type inboxState struct {
	items                    []domain.InboxItem
	id                       string
	cursor, offset, sequence int
	detail, loading          bool
	approving                bool
	vp                       viewport.Model
	snapshot                 *ReviewSnapshot
	err                      error
	extraScopes              []string
	pendingNotice            string
	docKey                   string
}
type inboxReviewMsg struct {
	id       string
	sequence int
	snapshot ReviewSnapshot
	err      error
}
type inboxApprovedMsg struct {
	id  string
	err error
}

func (m *Model) noteCompletions(t *domain.Tree) {
	if t == nil || m.tree == nil {
		return
	}
	count := 0
	for id, i := range t.Issues {
		if t.State(id) != domain.StateDone || i.Resolution == nil {
			continue
		}
		old := m.tree.Issues[id]
		if old == nil || old.Resolution == nil || old.Resolution.ContentID() != i.Resolution.ContentID() {
			count++
		}
	}
	if count > 0 {
		m.inboxS.pendingNotice = fmt.Sprintf("%d new completion(s) · u opens Inbox", count)
	}
}
func (m *Model) syncInbox() {
	s := &m.inboxS
	s.items = m.tree.Inbox()
	for n, item := range s.items {
		if item.IssueID == s.id {
			s.cursor = n
			break
		}
	}
	s.cursor = clamp(s.cursor, 0, len(s.items)-1)
	id := ""
	if len(s.items) > 0 {
		id = s.items[s.cursor].IssueID
	}
	if id != s.id {
		s.id = id
		s.snapshot = nil
		s.err = nil
		s.extraScopes = nil
		s.sequence++
		s.vp.GotoTop()
	}
}
func (m *Model) openInbox() tea.Cmd {
	m.screen = screenInbox
	m.syncInbox()
	return m.loadInboxReview()
}
func (m *Model) inboxItem() *domain.InboxItem {
	if len(m.inboxS.items) == 0 {
		return nil
	}
	return &m.inboxS.items[m.inboxS.cursor]
}
func (m *Model) loadInboxReview() tea.Cmd {
	s := &m.inboxS
	item := m.inboxItem()
	s.sequence++
	s.loading = false
	if item == nil || item.Category != domain.InboxCompletionReview {
		s.snapshot = nil
		return nil
	}
	id, seq := s.id, s.sequence
	if m.opts.Review == nil {
		s.err = fmt.Errorf("review adapter unavailable; open the issue with o")
		return nil
	}
	s.loading = true
	scopes := append([]string(nil), s.extraScopes...)
	return func() tea.Msg {
		snapshot, err := m.opts.Review(id, scopes)
		return inboxReviewMsg{id, seq, snapshot, err}
	}
}
func (m *Model) applyInboxReview(msg inboxReviewMsg) tea.Cmd {
	s := &m.inboxS
	if msg.id != s.id || msg.sequence != s.sequence {
		return nil
	}
	s.loading = false
	s.err = msg.err
	if msg.err != nil {
		s.snapshot = nil
	} else {
		s.snapshot = &msg.snapshot
	}
	return nil
}
func (m *Model) handleInboxApproved(msg inboxApprovedMsg) tea.Cmd {
	m.inboxS.approving = false
	matching := m.modal != nil && m.modal.kind == modalReviewApprove && m.modal.id == msg.id
	if msg.err != nil {
		if matching {
			m.modal.err = msg.err.Error()
			return nil
		}
		return m.flashErr(msg.err)
	}
	if matching {
		m.modal = nil
	}
	return tea.Batch(m.flash("Review approved · "+shortID(msg.id)), m.reload())
}
func (m *Model) inboxKey(key string, k tea.KeyPressMsg) tea.Cmd {
	s := &m.inboxS
	switch key {
	case "o":
		if s.id != "" {
			m.pushBack()
			m.screen = screenIssues
			return m.jump(s.id, false)
		}
	case "a":
		if s.approving {
			return m.flash("Review approval is being saved")
		}
		if s.loading || s.snapshot == nil || m.opts.ApproveReview == nil {
			return m.flash("Open a captured completion review before approving")
		}
		if s.snapshot.Fingerprint == "" {
			return m.flash("Choose work scopes with e, then inspect the refreshed bundle")
		}
		snapshot := *s.snapshot
		m.modal = &modal{kind: modalReviewApprove, id: s.id, reviewSnapshot: &snapshot, inputs: []textinput.Model{m.newInput("why you approve this completion", "")}}
		return m.modal.inputs[0].Focus()
	case "e":
		if item := m.inboxItem(); item != nil && item.Category == domain.InboxCompletionReview {
			m.modal = &modal{kind: modalReviewScopes, id: s.id, inputs: []textinput.Model{m.newInput("paths separated by commas", strings.Join(s.extraScopes, ", "))}}
			return m.modal.inputs[0].Focus()
		}
	case "esc", "left", "h":
		if s.detail {
			s.detail = false
			return nil
		}
		return m.goScreen(screenIssues)
	case "enter", "right", "l":
		s.detail = true
		return nil
	}
	if s.detail {
		var cmd tea.Cmd
		s.vp, cmd = s.vp.Update(k)
		return cmd
	}
	to := s.cursor
	switch key {
	case "j", "down":
		to++
	case "k", "up":
		to--
	case "g", "home":
		to = 0
	case "G", "end":
		to = len(s.items) - 1
	}
	to = clamp(to, 0, len(s.items)-1)
	if to != s.cursor && len(s.items) > 0 {
		s.cursor = to
		s.id = ""
		m.syncInbox()
		return m.loadInboxReview()
	}
	return nil
}

func (m *Model) inboxPane(w, h int) string {
	lw, rw := m.split(w, listRatio, minListW, minDetailW)
	if rw == 0 {
		if m.inboxS.detail {
			return m.inboxDetail(w, h)
		}
		return m.inboxList(w, h)
	}
	left := m.inboxList(lw, h)
	m.at.x += lw
	return lipgloss.JoinHorizontal(lipgloss.Top, left, m.inboxDetail(rw, h))
}
func (m *Model) inboxList(w, h int) string {
	s := &m.inboxS
	inner := ui.Inner(w)
	rows := max(1, (h-2)/2)
	s.offset = listOffset(s.offset, s.cursor, len(s.items), rows, !m.wheeled)
	var lines []string
	for n := s.offset; n < len(s.items) && n < s.offset+rows; n++ {
		item := s.items[n]
		i := m.tree.Issues[item.IssueID]
		title := ui.Fit(shortID(i.ID)+" "+i.Title, inner)
		reason := ui.Fit(string(item.Category)+" · "+item.Reason, inner)
		if n == s.cursor {
			title = ui.Selected(m.th, title, inner)
		}
		m.mark(fmt.Sprintf("inbox:%d", n), paneInner.x, paneInner.y+2*(n-s.offset), inner, 2)
		lines = append(lines, title, m.th.S.Muted.Render(reason))
	}
	body := strings.Join(lines, "\n")
	if len(s.items) == 0 {
		body = ui.Empty(m.th, "Nothing needs attention", "Ordinary backlog stays in Issues", inner, h-2)
	}
	m.pane("inbox", w, h)
	return ui.Pane{Title: fmt.Sprintf("Inbox · %d", len(s.items)), Body: body, Focused: !s.detail, Width: w, Height: h}.View(m.th)
}
func (m *Model) inboxDetail(w, h int) string {
	s := &m.inboxS
	inner := ui.Inner(w)
	doc := "Nothing selected."
	if item := m.inboxItem(); item != nil {
		i := m.tree.Issues[item.IssueID]
		doc = fmt.Sprintf("# %s\n\n`%s`\n\n%s\n\n%s\n\nPress o to open the issue and its lifecycle actions.", i.Title, i.ID, item.Reason, i.Prose)
		if i.OpenQuestions != "" {
			doc += "\n\n## Open questions\n\n" + i.OpenQuestions
		}
		if item.Category == domain.InboxCompletionReview {
			switch {
			case s.loading:
				doc = "Loading a fresh review bundle…"
			case s.err != nil:
				doc = "Review unavailable: " + s.err.Error()
			case s.snapshot != nil:
				doc = reviewDocument(*s.snapshot)
			}
		}
	}
	s.vp.SetWidth(inner)
	s.vp.SetHeight(max(1, h-2))
	key := fmt.Sprintf("%d:%s", inner, doc)
	if key != s.docKey {
		rendered, err := ui.Markdown(m.th, doc, inner)
		if err != nil {
			rendered = doc
		}
		s.vp.SetContent(rendered)
		s.docKey = key
	}
	m.pane("inbox-detail", w, h)
	return ui.Pane{Title: "Review / attention", Body: s.vp.View(), Focused: s.detail, Width: w, Height: h}.View(m.th)
}
func reviewDocument(s ReviewSnapshot) string {
	i := s.Issue
	if i == nil || i.Resolution == nil {
		return "No completion resolution available."
	}
	r := i.Resolution
	generation := r.Generation
	if generation == 0 {
		generation = 1
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n`%s`\n\nReview: **%s** · generation %d\n\nResolution `%s`\n\nCompleted by %s at %s\n\n%s\n\n", i.Title, i.ID, s.Status, generation, s.ResolutionID, r.By, r.At.Format("2006-01-02 15:04:05Z07:00"), r.Note)
	// The ready baseline preserves the requirement that was implemented.
	prose := i.Prose
	hasBaseline := false
	if i.Ready != nil {
		for _, base := range i.Baselines {
			if base.Name == i.Ready.Baseline {
				prose = base.Requirement
				hasBaseline = true
				break
			}
		}
	}
	if hasBaseline {
		b.WriteString("## Requirement at ready signoff\n\n")
	} else {
		b.WriteString("## Current requirement (ready snapshot unavailable)\n\n")
	}
	b.WriteString(prose + "\n\n")
	b.WriteString("## Current acceptance checklist\n\n")
	for _, c := range i.Criteria {
		mark := " "
		if c.Checked {
			mark = "x"
		}
		fmt.Fprintf(&b, "- [%s] %s\n", mark, c.Text)
	}
	b.WriteString("\n## Completion contract and Definition of Done\n\n")
	if r.Contract != nil {
		if s.CurrentContractID != "" && s.CurrentContractID != r.Contract.ID {
			b.WriteString("The current policy or requirements differ from the recorded completion contract.\n\n")
		}
		fmt.Fprintf(&b, "Verification policy: %s\n\nKnowledge policy: %s\n\n", r.Contract.Verification, r.Contract.Documentation)
		for _, req := range r.Contract.Requirements {
			fmt.Fprintf(&b, "- **%s**: %s\n", req.Kind, req.Text)
		}
		for _, opt := range r.Contract.OptOuts {
			fmt.Fprintf(&b, "- Opt out: %s — %s\n", opt.Item, opt.Reason)
		}
		b.WriteString("\n")
	} else {
		b.WriteString("Legacy completion: no bound contract.\n\n")
	}
	for _, d := range r.DoD {
		fmt.Fprintf(&b, "- %s\n", d)
	}
	for _, d := range r.DoDOptOuts {
		fmt.Fprintf(&b, "- Opt out: %s — %s\n", d.Item, d.Reason)
	}
	b.WriteString("\n## Verification evidence (reported)\n\n")
	if len(r.Verification) == 0 {
		b.WriteString("No captured verification reports.\n\n")
	}
	for _, v := range r.Verification {
		label := v.Requirement
		if r.Contract != nil {
			for _, req := range r.Contract.Requirements {
				if req.ID == v.Requirement {
					label = req.Kind + ": " + req.Text
					break
				}
			}
		}
		fmt.Fprintf(&b, "### %s · %s\n\n%s\n\nMethod: %s · command: `%s`\n\nBy %s at %s · %s · current scoped content: %t\n\nScopes: `%s`\n\n", label, v.Result, v.Observation, v.Method, v.Command, v.By, v.At.Format("2006-01-02 15:04Z07:00"), v.Provenance, i.EvidenceCurrent[v.ID], strings.Join(v.Scopes, "`, `"))
	}
	b.WriteString("## Knowledge enrichment (reported)\n\n")
	if len(r.KnowledgeReviews) == 0 {
		b.WriteString("No captured knowledge assessments.\n\n")
	}
	for _, v := range r.KnowledgeReviews {
		fmt.Fprintf(&b, "### %s %s\n\n%s\n\nRevision `%s` · by %s · current: %t\n\n", v.Kind, v.Entry, v.Observation, v.EntryRevision, v.By, s.KnowledgeCurrent[v.ID])
	}
	if r.Documentation != nil {
		b.WriteString("Documentation declaration:\n\n")
		for _, entry := range r.Documentation.Entries {
			fmt.Fprintf(&b, "- `%s`\n", entry)
		}
		if r.Documentation.NoImpact != "" {
			fmt.Fprintf(&b, "No knowledge impact declared: %s\n\n", r.Documentation.NoImpact)
		}
	}
	b.WriteString("## Exceptions and other approvals\n\n")
	if len(r.Approvals) == 0 {
		b.WriteString("No captured exception approvals.\n\n")
	}
	for _, a := range r.Approvals {
		fmt.Fprintf(&b, "- %s · %s · %s · by %s\n", a.Kind, a.Requirement, a.Reason, a.By)
	}
	fmt.Fprintf(&b, "\n## Captured work and changes\n\nScopes: `%s`\n\nConfiguration: `%s`\n\nFingerprint `%s`\n\n%s\n\n%s\n\n", strings.Join(s.WorkScopes, "`, `"), strings.Join(s.ConfigScopes, "`, `"), s.Fingerprint, s.ChangesNote, s.Evidence)
	if s.Changes != "" {
		fmt.Fprintf(&b, "```diff\n%s\n```\n\n", s.Changes)
	}
	b.WriteString("Press a to Approve review, e to select additional work scopes, y to copy the issue ID, or o to discuss it through the issue. Leaving keeps review pending. Reopening is an explicit agent action after your instruction.")
	return b.String()
}
