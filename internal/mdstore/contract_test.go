package mdstore_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
)

// The contract corpus is a set of valid and deliberately broken trees. Valid
// trees must produce no errors (warnings only when listed in expect); broken
// trees must produce every code listed in expect, and at least one error.
// The corpus doubles as the format specification's examples and as the test
// suite for any future storage adapter.
func TestContract(t *testing.T) {
	for _, kind := range []string{"valid", "invalid"} {
		dirs, err := os.ReadDir(filepath.Join("..", "..", "testdata", "contract", kind))
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range dirs {
			root := filepath.Join("..", "..", "testdata", "contract", kind, d.Name())
			t.Run(kind+"/"+d.Name(), func(t *testing.T) {
				tree, err := domain.Load(mdstore.Open(root), &okf.Store{Root: root}, false)
				if err != nil {
					t.Fatal(err)
				}
				got := map[string]domain.Severity{}
				var lines []string
				for _, dg := range domain.Validate(tree) {
					got[dg.Code] = dg.Severity
					lines = append(lines, string(dg.Severity)+" "+dg.Code+" "+dg.Issue+" "+dg.File+": "+dg.Message)
				}
				expect := readExpect(t, root)
				for _, code := range expect {
					if _, ok := got[code]; !ok {
						t.Errorf("expected %s, got:\n%s", code, strings.Join(lines, "\n"))
					}
				}
				listed := map[string]bool{}
				for _, c := range expect {
					listed[c] = true
				}
				hasErr := false
				for code, sev := range got {
					if sev == domain.SevError {
						hasErr = true
					}
					if kind == "valid" && !listed[code] {
						t.Errorf("unexpected %s %s:\n%s", sev, code, strings.Join(lines, "\n"))
					}
				}
				if kind == "invalid" && !hasErr {
					t.Errorf("broken tree produced no error:\n%s", strings.Join(lines, "\n"))
				}
			})
		}
	}
}

func readExpect(t *testing.T, root string) []string {
	b, err := os.ReadFile(filepath.Join(root, "expect"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	codes := strings.Fields(string(b))
	sort.Strings(codes)
	return codes
}
