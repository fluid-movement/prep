// Package outcome holds the assertions that read what the agent left
// behind: issues, knowledge entries, the check result and the repository,
// all taken from snapshots read through prep. Importing it registers its
// kinds.
package outcome

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/assert"
)

func init() {
	assert.Register(assert.Outcome, "issue", decodeIssue)
	assert.Register(assert.Outcome, "unchanged", decodeUnchanged)
	assert.Register(assert.Outcome, "knowledge", decodeKnowledge)
	assert.Register(assert.Outcome, "check", decodeCheck)
	assert.Register(assert.Outcome, "claimed", decodeClaimed)
	assert.Register(assert.Outcome, "repo", decodeRepo)
}

// project returns the snapshots of the named project (default: the
// primary) or an explanation.
func project(in *assert.Input, name string) (before, after *assert.ProjectSnapshot, err error) {
	if in.After == nil || in.Before == nil {
		return nil, nil, fmt.Errorf("no snapshots to check")
	}
	if name == "" {
		name = in.Project.Name
	}
	after = in.After.Projects[name]
	before = in.Before.Projects[name]
	if after == nil {
		return nil, nil, fmt.Errorf("no project %q in the workspace", name)
	}
	if before == nil {
		before = &assert.ProjectSnapshot{Issues: map[string]*assert.Issue{}, Knowledge: map[string]*assert.Entry{}}
	}
	return before, after, nil
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 100 {
		return s[:100] + "…"
	}
	return s
}

// bareID strips a cross-project prefix ("api:01ABC" -> "01ABC").
func bareID(s string) string {
	if i := strings.LastIndex(s, ":"); i >= 0 {
		return s[i+1:]
	}
	return s
}

// ---- issue

type issueCheck struct {
	Ref       string   `yaml:"ref"`   // an issue ID (usually {{alias}})
	Title     string   `yaml:"title"` // a substring of the title of an issue the agent created
	Project   string   `yaml:"project"`
	Exists    *bool    `yaml:"exists"`
	State     string   `yaml:"state"`
	IssueKind string   `yaml:"kind"`
	Parent    string   `yaml:"parent"` // an issue ID
	DependsOn []string `yaml:"depends_on"`
	Tags      []string `yaml:"tags"`
	Priority  string   `yaml:"priority"`
	Criteria  *struct {
		Count    *int     `yaml:"count"`
		Min      *int     `yaml:"min"`
		Contains []string `yaml:"contains"`
		Checked  *int     `yaml:"checked"`
	} `yaml:"criteria"`
	Requirement string `yaml:"requirement"` // text the requirement contains
	Context     string `yaml:"context"`     // text the context contains
}

func decodeIssue(n *yaml.Node) (assert.Assertion, error) {
	var c issueCheck
	if err := assert.DecodeStrict(n, &c); err != nil {
		return nil, err
	}
	if (c.Ref == "") == (c.Title == "") {
		return nil, fmt.Errorf("line %d: issue needs exactly one of ref and title", n.Line)
	}
	if c.Ref != "" && c.Exists != nil && !*c.Exists {
		return nil, fmt.Errorf("line %d: exists: false goes with title, not ref", n.Line)
	}
	return &c, nil
}

func (c *issueCheck) Kind() string { return "issue" }
func (c *issueCheck) Describe() string {
	who := "issue " + c.Ref
	if c.Title != "" {
		who = fmt.Sprintf("an issue titled like %q", c.Title)
	}
	if c.Exists != nil && !*c.Exists {
		return "no " + who
	}
	var want []string
	add := func(k, v string) {
		if v != "" {
			want = append(want, k+" "+v)
		}
	}
	add("state", c.State)
	add("kind", c.IssueKind)
	add("parent", c.Parent)
	add("priority", c.Priority)
	if len(c.DependsOn) > 0 {
		add("depends on", strings.Join(c.DependsOn, ","))
	}
	if len(c.Tags) > 0 {
		add("tags", strings.Join(c.Tags, ","))
	}
	if c.Requirement != "" {
		add("requirement containing", fmt.Sprintf("%q", c.Requirement))
	}
	if c.Context != "" {
		add("context containing", fmt.Sprintf("%q", c.Context))
	}
	if c.Criteria != nil {
		want = append(want, "criteria as described")
	}
	if len(want) == 0 {
		return who + " exists"
	}
	return who + " with " + strings.Join(want, ", ")
}

func (c *issueCheck) Check(in *assert.Input) assert.Result {
	before, after, err := project(in, c.Project)
	if err != nil {
		return assert.Fail(c, err.Error(), -1)
	}
	var is *assert.Issue
	if c.Ref != "" {
		is = after.Issues[c.Ref]
		if is == nil {
			return assert.Fail(c, fmt.Sprintf("expected issue %s; it does not exist", c.Ref), -1)
		}
	} else {
		var matches []*assert.Issue
		for id, i := range after.Issues {
			if _, existed := before.Issues[id]; existed {
				continue
			}
			if strings.Contains(strings.ToLower(i.Title), strings.ToLower(c.Title)) {
				matches = append(matches, i)
			}
		}
		sort.Slice(matches, func(a, b int) bool { return matches[a].ID < matches[b].ID })
		if c.Exists != nil && !*c.Exists {
			if len(matches) > 0 {
				return assert.Fail(c, fmt.Sprintf("expected no new issue titled like %q; found %s %q", c.Title, matches[0].ID, matches[0].Title), -1)
			}
			return assert.Pass(c, "no such issue", -1)
		}
		switch len(matches) {
		case 0:
			var titles []string
			for id, i := range after.Issues {
				if _, existed := before.Issues[id]; !existed {
					titles = append(titles, fmt.Sprintf("%q", i.Title))
				}
			}
			sort.Strings(titles)
			return assert.Fail(c, fmt.Sprintf("expected a new issue titled like %q; new issues: %s", c.Title, orNone(titles)), -1)
		case 1:
			is = matches[0]
		default:
			var ids []string
			for _, m := range matches {
				ids = append(ids, fmt.Sprintf("%s %q", m.ID, m.Title))
			}
			return assert.Fail(c, fmt.Sprintf("ambiguous: %d new issues are titled like %q: %s", len(matches), c.Title, strings.Join(ids, "; ")), -1)
		}
	}
	var problems []string
	cmp := func(what, want, got string) {
		if want != "" && want != got {
			problems = append(problems, fmt.Sprintf("%s is %q, want %q", what, got, want))
		}
	}
	cmp("state", c.State, is.State)
	cmp("kind", c.IssueKind, is.Kind)
	cmp("parent", c.Parent, is.Parent)
	cmp("priority", c.Priority, is.Priority)
	for _, d := range c.DependsOn {
		found := false
		for _, have := range is.DependsOn {
			if have == d || bareID(have) == bareID(d) {
				found = true
			}
		}
		if !found {
			problems = append(problems, fmt.Sprintf("does not depend on %s (depends on %v)", d, is.DependsOn))
		}
	}
	for _, t := range c.Tags {
		if !has(is.Tags, t) {
			problems = append(problems, fmt.Sprintf("lacks tag %q (has %v)", t, is.Tags))
		}
	}
	if c.Requirement != "" && !strings.Contains(strings.ToLower(is.Requirement), strings.ToLower(c.Requirement)) {
		problems = append(problems, fmt.Sprintf("requirement lacks %q", c.Requirement))
	}
	if c.Context != "" && !strings.Contains(strings.ToLower(is.Context), strings.ToLower(c.Context)) {
		problems = append(problems, fmt.Sprintf("context lacks %q", c.Context))
	}
	if cr := c.Criteria; cr != nil {
		n := len(is.Criteria)
		if cr.Count != nil && n != *cr.Count {
			problems = append(problems, fmt.Sprintf("has %d criteria, want %d", n, *cr.Count))
		}
		if cr.Min != nil && n < *cr.Min {
			problems = append(problems, fmt.Sprintf("has %d criteria, want at least %d", n, *cr.Min))
		}
		for _, want := range cr.Contains {
			found := false
			for _, k := range is.Criteria {
				if strings.Contains(strings.ToLower(k.Text), strings.ToLower(want)) {
					found = true
				}
			}
			if !found {
				problems = append(problems, fmt.Sprintf("no criterion mentions %q", want))
			}
		}
		if cr.Checked != nil {
			got := 0
			for _, k := range is.Criteria {
				if k.Checked {
					got++
				}
			}
			if got != *cr.Checked {
				problems = append(problems, fmt.Sprintf("%d criteria checked, want %d", got, *cr.Checked))
			}
		}
	}
	if len(problems) > 0 {
		return assert.Fail(c, fmt.Sprintf("issue %s %q: %s", is.ID, is.Title, strings.Join(problems, "; ")), -1)
	}
	return assert.Pass(c, fmt.Sprintf("issue %s %q matches", is.ID, is.Title), -1)
}

func orNone(s []string) string {
	if len(s) == 0 {
		return "none"
	}
	return strings.Join(s, ", ")
}

func has(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// ---- unchanged

type unchanged struct {
	Project string   `yaml:"project"`
	Issues  []string `yaml:"issues"`
	All     bool     `yaml:"all"`
	Except  []string `yaml:"except"`
	// Knowledge includes the knowledge base in "all".
	Knowledge bool `yaml:"knowledge"`
}

func decodeUnchanged(n *yaml.Node) (assert.Assertion, error) {
	var u unchanged
	switch n.Kind {
	case yaml.SequenceNode:
		if err := n.Decode(&u.Issues); err != nil {
			return nil, err
		}
	default:
		if err := assert.DecodeStrict(n, &u); err != nil {
			return nil, err
		}
	}
	if !u.All && len(u.Issues) == 0 {
		return nil, fmt.Errorf("line %d: unchanged needs a list of issues or all: true", n.Line)
	}
	return &u, nil
}

func (u *unchanged) Kind() string { return "unchanged" }
func (u *unchanged) Describe() string {
	if u.All {
		d := "nothing changed in the project's issues"
		if u.Project != "" {
			d = "nothing changed in project " + u.Project
		}
		if len(u.Except) > 0 {
			d += " except " + strings.Join(u.Except, ", ")
		}
		return d
	}
	return "unchanged: " + strings.Join(u.Issues, ", ")
}

func (u *unchanged) Check(in *assert.Input) assert.Result {
	_, after, err := project(in, u.Project)
	if err != nil {
		return assert.Fail(u, err.Error(), -1)
	}
	name := u.Project
	if name == "" {
		name = in.Project.Name
	}
	pd := in.Diff.Projects[name]
	if pd == nil {
		return assert.Fail(u, "no diff for project "+name, -1)
	}
	var bad []string
	report := func(id, what string) {
		if is := after.Issues[id]; is != nil {
			bad = append(bad, fmt.Sprintf("%s %q %s", id, is.Title, what))
		} else {
			bad = append(bad, fmt.Sprintf("%s %s", id, what))
		}
	}
	if u.All {
		for _, id := range pd.Created {
			if !has(u.Except, id) {
				report(id, "was created")
			}
		}
		for id, fields := range pd.Changed {
			if !has(u.Except, id) {
				report(id, "changed: "+strings.Join(fields, ", "))
			}
		}
		for _, id := range pd.Removed {
			if !has(u.Except, id) {
				report(id, "was removed")
			}
		}
		if u.Knowledge && (len(pd.EntriesCreated)+len(pd.EntriesChanged)+len(pd.EntriesRemoved)) > 0 {
			bad = append(bad, fmt.Sprintf("knowledge changed: %v", append(append(pd.EntriesCreated, pd.EntriesChanged...), pd.EntriesRemoved...)))
		}
	} else {
		for _, id := range u.Issues {
			if fields, ok := pd.Changed[id]; ok {
				report(id, "changed: "+strings.Join(fields, ", "))
			}
			for _, r := range pd.Removed {
				if r == id {
					report(id, "was removed")
				}
			}
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		return assert.Fail(u, "expected no change; "+strings.Join(bad, "; "), -1)
	}
	return assert.Pass(u, "nothing changed", -1)
}

// ---- knowledge

type knowledgeCheck struct {
	Path        string   `yaml:"path"`
	Project     string   `yaml:"project"`
	Exists      *bool    `yaml:"exists"`
	Type        string   `yaml:"type"`
	Status      string   `yaml:"status"`
	Contains    []string `yaml:"contains"`
	NotContains []string `yaml:"not_contains"`
	Created     *bool    `yaml:"created"` // the entry did not exist before the run
}

func decodeKnowledge(n *yaml.Node) (assert.Assertion, error) {
	var k knowledgeCheck
	if n.Kind == yaml.ScalarNode {
		k.Path = n.Value
	} else if err := assert.DecodeStrict(n, &k); err != nil {
		return nil, err
	}
	if k.Path == "" {
		return nil, fmt.Errorf("line %d: knowledge needs a path", n.Line)
	}
	return &k, nil
}

func (k *knowledgeCheck) Kind() string { return "knowledge" }
func (k *knowledgeCheck) Describe() string {
	if k.Exists != nil && !*k.Exists {
		return "no knowledge entry " + k.Path
	}
	d := "knowledge entry " + k.Path + " exists"
	if k.Type != "" {
		d += ", type " + k.Type
	}
	if len(k.Contains) > 0 {
		d += fmt.Sprintf(", containing %q", k.Contains)
	}
	return d
}

func (k *knowledgeCheck) Check(in *assert.Input) assert.Result {
	before, after, err := project(in, k.Project)
	if err != nil {
		return assert.Fail(k, err.Error(), -1)
	}
	e := after.Knowledge[k.Path]
	if k.Exists != nil && !*k.Exists {
		if e != nil {
			return assert.Fail(k, "expected no entry "+k.Path+"; it exists", -1)
		}
		return assert.Pass(k, "no such entry", -1)
	}
	if e == nil {
		var have []string
		for p := range after.Knowledge {
			have = append(have, p)
		}
		sort.Strings(have)
		return assert.Fail(k, fmt.Sprintf("expected entry %s; entries: %s", k.Path, orNone(have)), -1)
	}
	var problems []string
	if k.Type != "" && e.Type != k.Type {
		problems = append(problems, fmt.Sprintf("type is %q, want %q", e.Type, k.Type))
	}
	if k.Status != "" && e.Status != k.Status {
		problems = append(problems, fmt.Sprintf("status is %q, want %q", e.Status, k.Status))
	}
	for _, c := range k.Contains {
		if !strings.Contains(strings.ToLower(e.Text), strings.ToLower(c)) {
			problems = append(problems, fmt.Sprintf("text lacks %q", c))
		}
	}
	for _, c := range k.NotContains {
		if strings.Contains(strings.ToLower(e.Text), strings.ToLower(c)) {
			problems = append(problems, fmt.Sprintf("text contains %q", c))
		}
	}
	if k.Created != nil {
		_, existed := before.Knowledge[k.Path]
		if existed == *k.Created {
			problems = append(problems, fmt.Sprintf("created=%v but existed before=%v", *k.Created, existed))
		}
	}
	if len(problems) > 0 {
		return assert.Fail(k, fmt.Sprintf("entry %s: %s", k.Path, strings.Join(problems, "; ")), -1)
	}
	return assert.Pass(k, "entry matches", -1)
}

// ---- check

type checkResult struct {
	Project string   `yaml:"project"`
	Allow   []string `yaml:"allow"`
	// Warnings, when set, is the most warnings accepted.
	Warnings *int `yaml:"warnings"`
}

func decodeCheck(n *yaml.Node) (assert.Assertion, error) {
	var c checkResult
	if n.Kind == yaml.ScalarNode { // `check: clean`
		return &c, nil
	}
	if err := assert.DecodeStrict(n, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *checkResult) Kind() string { return "check" }
func (c *checkResult) Describe() string {
	d := "prep check reports no errors"
	if len(c.Allow) > 0 {
		d += " except " + strings.Join(c.Allow, ", ")
	}
	return d
}
func (c *checkResult) Check(in *assert.Input) assert.Result {
	_, after, err := project(in, c.Project)
	if err != nil {
		return assert.Fail(c, err.Error(), -1)
	}
	var errs []string
	for _, d := range after.Check.Diagnostics {
		if d.Severity == "error" && !has(c.Allow, d.Code) {
			errs = append(errs, fmt.Sprintf("%s %s: %s", d.Code, d.File, clip(d.Message)))
		}
	}
	if len(errs) > 0 {
		return assert.Fail(c, fmt.Sprintf("prep check reports %d error(s): %s", len(errs), strings.Join(errs, "; ")), -1)
	}
	if c.Warnings != nil && after.Check.Warnings > *c.Warnings {
		return assert.Fail(c, fmt.Sprintf("prep check reports %d warning(s), want at most %d", after.Check.Warnings, *c.Warnings), -1)
	}
	return assert.Pass(c, "no errors", -1)
}

// ---- claimed

type claimed struct {
	Ref     string `yaml:"ref"`
	Project string `yaml:"project"`
	By      string `yaml:"by"`      // regular expression on the claimer
	Claimed *bool  `yaml:"claimed"` // false: expects it not to be claimed
	re      *regexp.Regexp
}

func decodeClaimed(n *yaml.Node) (assert.Assertion, error) {
	var c claimed
	if n.Kind == yaml.ScalarNode {
		c.Ref = n.Value
	} else if err := assert.DecodeStrict(n, &c); err != nil {
		return nil, err
	}
	if c.Ref == "" {
		return nil, fmt.Errorf("line %d: claimed needs ref", n.Line)
	}
	if c.By != "" {
		re, err := regexp.Compile(c.By)
		if err != nil {
			return nil, err
		}
		c.re = re
	}
	return &c, nil
}

func (c *claimed) want() bool   { return c.Claimed == nil || *c.Claimed }
func (c *claimed) Kind() string { return "claimed" }
func (c *claimed) Describe() string {
	if !c.want() {
		return "issue " + c.Ref + " is not claimed"
	}
	return "issue " + c.Ref + " is claimed"
}
func (c *claimed) Check(in *assert.Input) assert.Result {
	_, after, err := project(in, c.Project)
	if err != nil {
		return assert.Fail(c, err.Error(), -1)
	}
	is := after.Issues[c.Ref]
	if is == nil {
		return assert.Fail(c, "no issue "+c.Ref, -1)
	}
	isClaimed := is.State == "in-progress"
	switch {
	case isClaimed != c.want():
		return assert.Fail(c, fmt.Sprintf("issue %s %q is %s (state %s)", is.ID, is.Title, map[bool]string{true: "claimed", false: "not claimed"}[isClaimed], is.State), -1)
	case isClaimed && c.re != nil && !c.re.MatchString(is.ClaimedBy):
		return assert.Fail(c, fmt.Sprintf("issue %s is claimed by %q, want /%s/", is.ID, is.ClaimedBy, c.By), -1)
	}
	return assert.Pass(c, "as expected", -1)
}

// ---- repo

type repoCheck struct {
	Project string `yaml:"project"`
	// Commit says a new commit exists (true) or does not (false).
	Commit *bool `yaml:"commit"`
	// Changed lists globs of files the agent changed (committed or not).
	Changed []string `yaml:"changed"`
	// Untouched lists globs of files it must not have changed.
	Untouched []string `yaml:"untouched"`
}

func decodeRepo(n *yaml.Node) (assert.Assertion, error) {
	var r repoCheck
	if err := assert.DecodeStrict(n, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *repoCheck) Kind() string     { return "repo" }
func (r *repoCheck) Describe() string { return "the repository changed as described" }
func (r *repoCheck) Check(in *assert.Input) assert.Result {
	_, after, err := project(in, r.Project)
	if err != nil {
		return assert.Fail(r, err.Error(), -1)
	}
	if after.Git == nil {
		return assert.Fail(r, "the project has no git repository", -1)
	}
	name := r.Project
	if name == "" {
		name = in.Project.Name
	}
	pd := in.Diff.Projects[name]
	var problems []string
	if r.Commit != nil && pd.HeadMoved != *r.Commit {
		problems = append(problems, fmt.Sprintf("new commit: %v, want %v", pd.HeadMoved, *r.Commit))
	}
	for _, g := range r.Changed {
		if !matchAny(pd.Files, g) {
			problems = append(problems, fmt.Sprintf("no change to %s (changed: %s)", g, orNone(pd.Files)))
		}
	}
	for _, g := range r.Untouched {
		if matchAny(pd.Files, g) {
			problems = append(problems, fmt.Sprintf("%s was changed", g))
		}
	}
	if len(problems) > 0 {
		return assert.Fail(r, strings.Join(problems, "; "), -1)
	}
	return assert.Pass(r, "as expected", -1)
}

func matchAny(files []string, glob string) bool {
	for _, f := range files {
		if ok, _ := path.Match(glob, f); ok {
			return true
		}
		if strings.HasSuffix(glob, "/**") && strings.HasPrefix(f, strings.TrimSuffix(glob, "**")) {
			return true
		}
	}
	return false
}
