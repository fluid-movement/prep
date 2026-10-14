package mdstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestDependencyOverrideRecordRoundTripAndFmt(t *testing.T) {
	o := domain.DependencyOverride{Ref: "api:01M4BBBBBBBBBBBBBBBBBBBBBB", SourceDigest: strings.Repeat("a", 64), LocalDigest: strings.Repeat("b", 64), Reason: "Proceed while checkout is unavailable", By: "human:alex", At: time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)}
	o.ID = o.ContentID()
	raw := renderDependencyOverrides([]domain.DependencyOverride{o})
	parsed, err := parseDependencyOverrides(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 || parsed[0] != o {
		t.Fatalf("round trip = %#v", parsed)
	}
	store := Open(t.TempDir())
	issueDir := store.abs(IssueDir("01M4AAAAAAAAAAAAAAAAAAAAAA"))
	if err := os.MkdirAll(issueDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issueDir, "dependency-overrides.md"), []byte(strings.ReplaceAll(raw, "\n", "\r\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := store.Fmt(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) != 1 || changed[0] != IssueDir("01M4AAAAAAAAAAAAAAAAAAAAAA")+"/dependency-overrides.md" {
		t.Fatalf("Fmt changes = %v", changed)
	}
}
