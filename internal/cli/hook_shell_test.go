package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestHookShellFromInsideRecords(t *testing.T) {
	root := t.TempDir()
	for _, cwd := range []string{".prep", ".prep/issues", ".prep/issues/x", ".prep/knowledge"} {
		for _, tc := range []struct {
			command string
			hint    string
		}{
			{`cd ..`, ""},
			{`pwd`, ""},
			{`git status`, ""},
			{`echo hi`, ""},
			{`prep check`, ""},
			{`cd .. && pwd`, ""},
			{`cat {f}`, "prep "},
			{`cat ./{f}`, "prep "},
			{`ls .`, "prep "},
			{`echo x > {f}`, "write commands"},
			{`rm {f}`, "write commands"},
		} {
			dir := filepath.Join(root, cwd)
			// A record file by a path relative to this directory.
			file := map[string]string{".prep": "project.md", ".prep/issues": "x/issue.md", ".prep/issues/x": "issue.md", ".prep/knowledge": "x.md"}[cwd]
			command := strings.ReplaceAll(tc.command, "{f}", file)
			t.Run(cwd+" "+command, func(t *testing.T) {
				got := hookRedirect(hookEvent{Cwd: dir, Tool: "Bash", Input: map[string]any{"command": command}}, root)
				if tc.hint == "" && got != "" || tc.hint != "" && !strings.Contains(got, tc.hint) {
					t.Fatalf("redirect = %q; want hint %q", got, tc.hint)
				}
			})
		}
	}
}

func TestHookRootRedirectUnchanged(t *testing.T) {
	root := t.TempDir()
	for _, command := range []string{`cat .prep/issues/x/issue.md`, `echo x > .prep/project.md`} {
		got := hookRedirect(hookEvent{Cwd: root, Tool: "Bash", Input: map[string]any{"command": command}}, root)
		if got == "" {
			t.Fatalf("%q was allowed from the project root", command)
		}
	}
}

func TestHookShellRecordPaths(t *testing.T) {
	root := t.TempDir()
	const issue = ".prep/issues/x/issue.md"
	for _, tc := range []struct {
		command string
		hint    string
	}{
		{`prep log x "text naming .prep/issues/x"`, ""},
		{`PREP_ACTOR=codex PREP_SESSION='a b' /bin/prep edit x --body "cat .prep/issues/x/issue.md"`, ""},
		{`git add -A .prep`, ""},
		{`git rm .prep/issues/x/issue.md`, ""},
		{`git checkout --ours .prep/knowledge/x.md`, ""},
		{`git diff -- .prep/issues/x/issue.md`, ""},
		{`git ls-files | grep prep-tmp`, ""},
		{"git commit -F - <<'EOF'\ncat .prep/issues/x/issue.md; rm .prep/project.md\nEOF", ""},
		{"prep new --body-file - <<'EOF'\nSee .prep/issues/x/issue.md\nEOF", ""},
		{"python3 - <<EOF\ntext = '.prep/issues/x/issue.md'\nopen('/tmp/scratch', 'w').write(text)\nEOF", ""},
		{`echo "see .prep/issues/x; cat .prep/issues/x/issue.md" > /tmp/message`, ""},
		{`cat 'see .prep/issues/x/issue.md'`, ""},
		{`cat "see .prep/issues/x/issue.md"`, ""},
		{`echo '.prep/issues/x/issue.md'`, ""},
		{`printf '%s' "cat .prep/issues/x/issue.md"`, ""},
		{`cat README.md # cat .prep/issues/x/issue.md`, ""},
		{`cat ` + issue, "prep show"},
		{`cat "` + issue + `"`, "prep show"},
		{`cat '` + issue + `'`, "prep show"},
		{`cat ./.prep/issues/x/../x/issue.md`, "prep show"},
		{`cat ` + root + `/` + issue, "prep show"},
		{`/bin/cat ` + issue, "prep show"},
		{`head ` + issue, "prep show"},
		{`tail ` + issue, "prep show"},
		{`sed -n '1,10p' ` + issue, "prep show"},
		{`awk '{print $0}' ` + issue, "prep show"},
		{`grep -r foo .prep/issues`, "prep show"},
		{`cat .prep/knowledge/x.md`, "prep knowledge"},
		{`sudo cat ` + issue, "prep show"},
		{`env FOO=1 cat ` + issue, "prep show"},
		{`timeout 5 cat ` + issue, "prep show"},
		{`nice -n 10 head ` + issue, "prep show"},
		{`echo ` + issue + ` | xargs cat ` + issue, "prep show"},
		{`sudo prep show x`, ""},
		{`sed -i s/a/b/ .prep/issues/x/context.md`, "write commands"},
		{`sed -i '' s/a/b/ ` + issue, "write commands"},
		{`sed -i.bak s/a/b/ ` + issue, "write commands"},
		{`echo x > .prep/issues/x/history.md`, "write commands"},
		{`echo x>>".prep/issues/x/history.md"`, "write commands"},
		{`cp a .prep/project.md`, "write commands"},
		{`mv a ` + issue, "write commands"},
		{`rm ` + issue, "write commands"},
		{`echo x | tee ` + issue, "write commands"},
		{`wc -l < ` + issue, "prep show"},
		{`cat ` + issue + ` > /tmp/scratch`, "prep show"},
		{`prep next && cat ` + issue, "prep show"},
		{`git diff; rm ` + issue, "write commands"},
		{"python3 - <<'EOF'\ncat .prep/issues/x/issue.md\nEOF\ncat " + issue, "prep show"},
		{"cat <<'EOF' > .prep/issues/x/history.md\ntext\nEOF", "write commands"},
		{"cat <<ONE <<'TWO'\ncat .prep/issues/x/issue.md\nONE\nrm .prep/project.md\nTWO", ""},
	} {
		t.Run(tc.command, func(t *testing.T) {
			got := hookRedirect(hookEvent{Cwd: root, Tool: "Bash", Input: map[string]any{"command": tc.command}}, root)
			if tc.hint == "" && got != "" || tc.hint != "" && !strings.Contains(got, tc.hint) {
				t.Fatalf("redirect = %q; want hint %q", got, tc.hint)
			}
		})
	}
}
