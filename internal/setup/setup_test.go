package setup

import (
	"errors"
	"strings"
	"testing"
)

type fake struct {
	name, version string
	detected      bool
	failInstall   bool
	calls         []string
}

func (f *fake) Name() string  { return f.name }
func (f *fake) Title() string { return strings.ToUpper(f.name) }
func (f *fake) Detect() bool  { return f.detected }
func (f *fake) Installed() (string, bool, error) {
	return f.version, f.version != "", nil
}
func (f *fake) Install(v string) ([]string, error) {
	f.calls = append(f.calls, "install "+v)
	if f.failInstall {
		return nil, errors.New("boom")
	}
	f.version = v
	return nil, nil
}
func (f *fake) Remove() error {
	f.calls = append(f.calls, "remove")
	f.version = ""
	return nil
}

func register(t *testing.T, fs ...*fake) {
	for _, f := range fs {
		Register(f)
		name := f.name
		t.Cleanup(func() { Unregister(name) })
	}
}

func TestApplyInstallsUpdatesAndRemoves(t *testing.T) {
	a := &fake{name: "a", detected: true}
	b := &fake{name: "b", detected: true, version: "v0.1.0"}
	c := &fake{name: "c", detected: true, version: "v0.2.0"}
	register(t, a, b, c)

	got := Apply([]string{"a", "c"}, "v0.2.0")
	want := map[string]string{"a": "installed", "b": "removed", "c": "up to date"}
	for _, r := range got {
		if want[r.Name] != r.Action {
			t.Errorf("%s: %s, want %s", r.Name, r.Action, want[r.Name])
		}
	}
	if a.version != "v0.2.0" || b.version != "" || len(c.calls) != 0 {
		t.Fatalf("a=%q b=%q c calls=%v", a.version, b.version, c.calls)
	}

	a.version = "v0.1.0"
	if r := Apply([]string{"a"}, "v0.2.0"); r[0].Action != "updated" {
		t.Fatalf("update: %+v", r)
	}
}

func TestApplyKeepsGoingAfterAFailure(t *testing.T) {
	a := &fake{name: "a", detected: true, failInstall: true}
	b := &fake{name: "b", detected: true}
	register(t, a, b)
	got := Apply([]string{"a", "b"}, "v1.0.0")
	if got[0].Action != "failed" || got[0].Err != "boom" || got[1].Action != "installed" {
		t.Fatalf("results = %+v", got)
	}
}

func TestRefresh(t *testing.T) {
	a := &fake{name: "a", detected: true, version: "v0.1.0"}
	gone := &fake{name: "gone", detected: false, version: "v0.1.0"}
	register(t, a, gone)
	got := Refresh([]string{"a", "gone", "unknown"}, "v0.2.0")
	if got[0].Action != "updated" || got[1].Action != "skipped" || got[2].Action != "skipped" || !strings.Contains(got[2].Err, "not supported") {
		t.Fatalf("results = %+v", got)
	}
	if gone.version != "v0.1.0" {
		t.Fatal("refresh touched a harness that is not on the machine")
	}
}

func TestValidateAndSurvey(t *testing.T) {
	register(t, &fake{name: "a", detected: true, version: "v1.0.0"}, &fake{name: "b"})
	if err := Validate([]string{"a", "zz"}); err == nil || !strings.Contains(err.Error(), "unknown harness zz; available: a, b") {
		t.Fatalf("validate: %v", err)
	}
	s := Survey([]string{"b"})
	if len(s) != 2 || !s[0].Detected || s[0].Installed != "v1.0.0" || s[0].Chosen || !s[1].Chosen {
		t.Fatalf("survey = %+v", s)
	}
}
