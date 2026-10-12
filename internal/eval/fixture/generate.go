package fixture

import "fmt"

var (
	genVerbs = []string{"Add", "Fix", "Remove", "Rename", "Document", "Speed up", "Validate", "Cache", "Log", "Migrate", "Split", "Simplify"}
	genNouns = []string{"search box", "profile page", "export button", "login form", "email digest", "audit log", "settings tab", "image upload", "rate limiter",
		"invoice list", "session timeout", "dark mode", "csv import", "webhook retries", "password reset", "team invite", "api pagination", "health check"}
	genKinds  = []string{"code", "code", "code", "research", "decision", "manual"}
	genPrio   = []string{"", "", "low", "medium", "high"}
	genStates = []string{"open", "open", "open", "defined", "ready", "ready", "done", "done", "dropped", "in-progress"}
)

// generatedIssues returns n synthetic issues, the same every time. Every
// seventh is the child of an earlier parent group, and some depend on an
// earlier issue; dependencies only point at issues that end done or are not
// being worked on, so the build can always reach the requested states.
func generatedIssues(g *GenSpec) []IssueSpec {
	if g == nil || g.Issues == 0 {
		return nil
	}
	var out []IssueSpec
	for i := 0; i < g.Issues; i++ {
		verb, noun := genVerbs[i%len(genVerbs)], genNouns[(i*7+3)%len(genNouns)]
		is := IssueSpec{
			Alias:    fmt.Sprintf("gen%d", i),
			Title:    fmt.Sprintf("%s the %s (%d)", verb, noun, i+1),
			Kind:     genKinds[i%len(genKinds)],
			State:    genStates[(i*3)%len(genStates)],
			Priority: genPrio[(i*5)%len(genPrio)],
			Body:     fmt.Sprintf("%s the %s so that customers are not blocked. Generated issue %d.", verb, noun, i+1),
			Tags:     []string{[]string{"web", "api", "ops", "docs"}[i%4]},
		}
		if is.State == "in-progress" || is.State == "done" || is.State == "ready" {
			is.Criteria = []string{"Behaves as described"}
		}
		if is.Kind == "manual" {
			is.Context = ""
		}
		out = append(out, is)
	}
	return out
}

func generatedKnowledge(g *GenSpec) []KnowledgeSpec {
	if g == nil || g.Knowledge == 0 {
		return nil
	}
	var out []KnowledgeSpec
	for i := 0; i < g.Knowledge; i++ {
		noun := genNouns[(i*5+1)%len(genNouns)]
		out = append(out, KnowledgeSpec{
			Path:        fmt.Sprintf("/components/%s-%d.md", slug(noun), i+1),
			Type:        "component",
			Title:       fmt.Sprintf("The %s (%d)", noun, i+1),
			Description: fmt.Sprintf("How the %s works.", noun),
			Body:        fmt.Sprintf("The %s is implemented in src/%s. It is covered by tests and has no known issues. Generated entry %d.", noun, slug(noun), i+1),
		})
	}
	return out
}

func slug(s string) string {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			out = append(out, byte(r))
		} else if len(out) > 0 && out[len(out)-1] != '-' {
			out = append(out, '-')
		}
	}
	return string(out)
}
