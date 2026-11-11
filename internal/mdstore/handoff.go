package mdstore

import (
	"path/filepath"

	"github.com/fluid-movement/prep/internal/activity"
)

// The hand-off files of an issue: the brief an orchestrator hands to an
// implementer or checker, and the report that agent writes back. They live
// in the per-machine, unversioned area, one directory per issue.
const (
	HandoffBrief       = "brief.md"
	HandoffReport      = "report.md"
	HandoffCheckBrief  = "check-brief.md"
	HandoffCheckReport = "check-report.md"
)

// HandoffDir returns the repository-relative hand-off directory of an issue,
// .prep/local/handoff/<id>.
func HandoffDir(id string) string { return Dir + "/" + activity.Dir + "/handoff/" + id }

// HandoffFiles names the brief and the report of a hand-off: the
// implementer's, or with check the checker's.
func HandoffFiles(check bool) (brief, report string) {
	if check {
		return HandoffCheckBrief, HandoffCheckReport
	}
	return HandoffBrief, HandoffReport
}

// HandoffPath returns the absolute path of a hand-off file of an issue in
// this checkout. It is absolute so that an agent working in another
// worktree writes where the orchestrator reads.
func (s *Store) HandoffPath(id, name string) string {
	p := s.abs(HandoffDir(id) + "/" + name)
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}
