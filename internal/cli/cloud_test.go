package cli

import (
	"bytes"
	"strings"
	"testing"
)

func runCloud(t *testing.T, version string, args ...string) (int, string, string) {
	t.Helper()
	old := Version
	Version = version
	t.Cleanup(func() { Version = old })
	var out, errb bytes.Buffer
	code := Main(append([]string{"cloud"}, args...), strings.NewReader(""), &out, &errb)
	return code, out.String(), errb.String()
}

func TestCloudScriptRelease(t *testing.T) {
	code, out, errs := runCloud(t, "v0.3.0", "script")
	if code != 0 || errs != "" {
		t.Fatalf("exit %d stderr %q", code, errs)
	}
	if !strings.Contains(out, "export PREP_VERSION=v0.3.0\n") {
		t.Errorf("no pin:\n%s", out)
	}
	checkCloudScript(t, out)
}

func TestCloudScriptDevelopment(t *testing.T) {
	code, out, errs := runCloud(t, "6449764-dirty", "script")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out, "PREP_VERSION") {
		t.Errorf("development build pins a version:\n%s", out)
	}
	if !strings.Contains(errs, "development build: the script installs the latest release") {
		t.Errorf("stderr: %q", errs)
	}
	checkCloudScript(t, out)
}

func TestCloudUsage(t *testing.T) {
	if code, _, _ := runCloud(t, "v0.3.0"); code != 2 {
		t.Errorf("no subcommand: exit %d, want 2", code)
	}
	if code, _, _ := runCloud(t, "v0.3.0", "bogus"); code != 2 {
		t.Errorf("bad subcommand: exit %d, want 2", code)
	}
}

func checkCloudScript(t *testing.T, out string) {
	t.Helper()
	var setupLine bool
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "curl ") || strings.Contains(l, " setup ") {
			if !strings.HasSuffix(l, "|| true") {
				t.Errorf("line can fail without || true: %s", l)
			}
		}
		if strings.Contains(l, " setup ") {
			setupLine = true
			if !strings.Contains(l, "setup --harness claude-code") {
				t.Errorf("setup line: %s", l)
			}
		}
	}
	if !setupLine {
		t.Error("script lacks the prep setup line")
	}
}
