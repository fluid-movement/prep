package cli

import (
	"encoding/json"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestLifecycleJSONGuidanceIsAdditiveAndActionable(t *testing.T) {
	policyHarness := newHarness(t)
	policyHarness.dir = t.TempDir()
	policyHarness.ok("init", "--no-bootstrap")
	var projectPolicy struct {
		OK         bool   `json:"ok"`
		Configured bool   `json:"configured"`
		Next       string `json:"next"`
	}
	policyHarness.jsonOf(&projectPolicy, "policy", "show")
	if !projectPolicy.OK || projectPolicy.Configured || projectPolicy.Next != "prep policy set --verification <method> --documentation <method>" {
		t.Fatalf("unconfigured policy guidance: %+v", projectPolicy)
	}
	var policyWrite writeResult
	policyHarness.jsonOf(&policyWrite, "policy", "set", "--verification", "Manual inspection", "--documentation", "Review documentation impact.")
	if !policyWrite.OK || policyWrite.Next != "prep policy show" {
		t.Fatalf("policy set response: %+v", policyWrite)
	}

	h, id, contract := approvalFixture(t)
	h.verify(id, "No documentation impact.")

	var policy map[string]any
	h.jsonOf(&policy, "policy", "show", id)
	if policy["ok"] != true || policy["id"] != id || policy["bound"] == nil || policy["current"] == nil || policy["next"] != "prep guide "+id {
		t.Fatalf("policy show contract: %#v", policy)
	}
	var evidence map[string]any
	h.jsonOf(&evidence, "evidence", "show", id)
	if evidence["ok"] != true || evidence["id"] != id || evidence["records"] == nil || evidence["next"] != "prep guide "+id {
		t.Fatalf("evidence show contract: %#v", evidence)
	}
	var knowledge map[string]any
	h.jsonOf(&knowledge, "knowledge", "review", "show", id)
	if knowledge["ok"] != true || knowledge["id"] != id || knowledge["reviews"] == nil || knowledge["next"] != "prep guide "+id {
		t.Fatalf("knowledge review show contract: %#v", knowledge)
	}

	code, output := h.run("complete", id, "--no-impact", "mismatched reason", "--json")
	if code == 0 {
		t.Fatalf("completion with a mismatched review unexpectedly passed: %s", output)
	}
	var failure struct {
		OK    bool `json:"ok"`
		Error struct {
			Code  string         `json:"code"`
			Next  string         `json:"next"`
			Unmet []domain.Unmet `json:"unmet"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &failure); err != nil {
		t.Fatalf("decode contract failure: %v\n%s", err, output)
	}
	if failure.OK || failure.Error.Code != domain.ErrGate || failure.Error.Next != "prep guide "+id || len(failure.Error.Unmet) == 0 {
		t.Fatalf("completion error contract: %+v", failure)
	}

	var write writeResult
	h.jsonOf(&write, "evidence", "record", id, "--requirement", contract.Requirements[0].ID,
		"--result", "pass", "--method", "manual", "--observation", "Rechecked the scoped work.", "--scope", ".prep/knowledge")
	if !write.OK || write.ID != id || write.Next != "prep evidence show "+id {
		t.Fatalf("evidence write response lost/omitted fields: %+v", write)
	}
	h.jsonOf(&write, "knowledge", "review", "record", id, "--no-impact", "--reason", "No documentation impact.", "--scope", ".prep/knowledge")
	if !write.OK || write.ID != id || write.Next != "prep knowledge review show "+id {
		t.Fatalf("knowledge review write response: %+v", write)
	}

	h.jsonOf(&write, "complete", id, "--no-impact", "No documentation impact.")
	if !write.OK || write.ID != id || write.Next != "prep guide "+id {
		t.Fatalf("lifecycle completion omitted bounded guidance: %+v", write)
	}
}
