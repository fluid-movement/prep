package mutate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Caller sends a prompt to a model and returns its reply (see judge.Caller).
type Caller func(ctx context.Context, prompt string) (string, error)

const proposePrompt = `You improve the instructions an AI coding agent gets for using a project tracker called prep. Below is the skill text (SKILL.md) the agent is given. Propose %d DIFFERENT rewrites of the whole skill, each testing one clear hypothesis about why agents might follow it better or at lower cost (for example: shorter, reordered, a rule made more concrete, a rule merged into another).

Rules for every rewrite:
- Keep the YAML frontmatter (name and description) exactly.
- Keep every command and flag accurate; do not invent commands.
- Change one idea per rewrite, so a result can be traced to it.
- Do not make the text more than twice as long.

%s
Reply with a single JSON object and nothing else:
{"mutants": [{"id": "short-kebab-id", "hypothesis": "one sentence", "description": "what changed", "skill_md": "the full rewritten SKILL.md"}]}

## Current SKILL.md

%s`

// Propose asks a model for n rewrites of the skill, each with a stated
// hypothesis. failures, when not empty, is a summary of what the agents got
// wrong in the latest report. Invalid proposals (bad id, lost frontmatter,
// unchanged, too long) are dropped.
func Propose(ctx context.Context, call Caller, p Plugin, failures string, n int) ([]Mutant, error) {
	text, err := p.read(SkillPath)
	if err != nil {
		return nil, err
	}
	hint := ""
	if strings.TrimSpace(failures) != "" {
		hint = "In the latest run the agents got these things wrong; favour rewrites that address them:\n" + failures + "\n"
	}
	reply, err := call(ctx, fmt.Sprintf(proposePrompt, n, hint, text))
	if err != nil {
		return nil, err
	}
	start, end := strings.Index(reply, "{"), strings.LastIndex(reply, "}")
	if start < 0 || end < start {
		return nil, fmt.Errorf("the proposal has no JSON object")
	}
	var r struct {
		Mutants []struct {
			ID          string `json:"id"`
			Hypothesis  string `json:"hypothesis"`
			Description string `json:"description"`
			SkillMD     string `json:"skill_md"`
		} `json:"mutants"`
	}
	if err := json.Unmarshal([]byte(reply[start:end+1]), &r); err != nil {
		return nil, fmt.Errorf("the proposal is not valid JSON: %w", err)
	}
	var out []Mutant
	seen := map[string]bool{}
	for _, m := range r.Mutants {
		id := "ai-" + slug(m.ID)
		if id == "ai-" || seen[id] || m.SkillMD == "" || m.SkillMD == text || len(m.SkillMD) > 2*len(text) {
			continue
		}
		if !strings.HasPrefix(m.SkillMD, "---\n") {
			continue
		}
		seen[id] = true
		out = append(out, Mutant{ID: id, Description: "Agent-proposed: " + strings.TrimSpace(m.Description), Hypothesis: m.Hypothesis,
			Files: map[string]string{SkillPath: strings.TrimRight(m.SkillMD, "\n") + "\n"}})
		if len(out) == n {
			break
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no usable mutant in the proposal")
	}
	return out, nil
}
