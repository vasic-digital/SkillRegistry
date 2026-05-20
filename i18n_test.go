package agents

import (
	"context"
	"os"
	"strings"
	"testing"
)

const testBundlePath = "i18n/bundles/active.en.yaml"

// TestMain wires the in-tree English bundle as the process Translator
// before any test runs, so migrated user-facing strings render as real
// English text. Without this, NoopTranslator echoes message IDs and
// substring assertions across the suite would fail — which is itself
// the round-333 CONST-046 anti-bluff guarantee: the seam MUST resolve
// real copy when wired.
func TestMain(m *testing.M) {
	bt, err := NewBundleTranslatorFromFile("en", testBundlePath)
	if err != nil {
		// Loud failure — a missing/broken bundle is a CONST-046 defect,
		// not something to silently tolerate.
		os.Stderr.WriteString("skillregistry i18n TestMain: " + err.Error() + "\n")
		os.Exit(1)
	}
	SetTranslator(bt)
	code := m.Run()
	SetTranslator(nil) // reset to NoopTranslator
	os.Exit(code)
}

// TestI18n_NoopTranslator_EchoesID asserts the safety default echoes
// the message ID verbatim (loud, never empty).
func TestI18n_NoopTranslator_EchoesID(t *testing.T) {
	got, err := NoopTranslator{}.T(context.Background(), "skillregistry_validation_id_required", nil)
	if err != nil {
		t.Fatalf("NoopTranslator.T returned error: %v", err)
	}
	if got != "skillregistry_validation_id_required" {
		t.Fatalf("NoopTranslator.T = %q, want verbatim message ID", got)
	}
}

// TestI18n_BundleTranslator_ResolvesRealCopy proves the bundle-backed
// translator returns real English text for a plain message and
// interpolates placeholders for a templated message.
func TestI18n_BundleTranslator_ResolvesRealCopy(t *testing.T) {
	bt, err := NewBundleTranslatorFromFile("en", testBundlePath)
	if err != nil {
		t.Fatalf("NewBundleTranslatorFromFile: %v", err)
	}

	plain, err := bt.T(context.Background(), "skillregistry_validation_id_required", nil)
	if err != nil {
		t.Fatalf("T(id_required): %v", err)
	}
	if plain != "skill ID is required" {
		t.Fatalf("T(id_required) = %q, want %q", plain, "skill ID is required")
	}

	tmpl, err := bt.T(context.Background(), "skillregistry_validation_category_invalid",
		map[string]any{"Category": "bogus"})
	if err != nil {
		t.Fatalf("T(category_invalid): %v", err)
	}
	if !strings.Contains(tmpl, "bogus") || !strings.Contains(tmpl, "invalid category") {
		t.Fatalf("T(category_invalid) = %q, want interpolated 'bogus'", tmpl)
	}
}

// TestI18n_UnknownID_LoudEcho asserts an unknown message ID surfaces
// an error AND echoes the ID — never a silent empty string. A silent
// swallow here would be a §11.4 PASS-bluff at the i18n layer.
func TestI18n_UnknownID_LoudEcho(t *testing.T) {
	bt, err := NewBundleTranslatorFromFile("en", testBundlePath)
	if err != nil {
		t.Fatalf("NewBundleTranslatorFromFile: %v", err)
	}
	got, terr := bt.T(context.Background(), "skillregistry_does_not_exist", nil)
	if terr == nil {
		t.Fatalf("T(unknown) returned nil error, want loud failure")
	}
	if got != "skillregistry_does_not_exist" {
		t.Fatalf("T(unknown) = %q, want verbatim echo of unknown ID", got)
	}
}

// migratedMessageIDs is the closed set of message IDs introduced by
// the round-333 CONST-046 migration. The paired-mutation test below
// asserts every one of them resolves to a non-empty, non-echo string
// from the shipped bundle — proving the bundle and the tr() callsites
// agree.
var migratedMessageIDs = []string{
	"skillregistry_validation_skill_nil",
	"skillregistry_validation_id_required",
	"skillregistry_validation_name_required",
	"skillregistry_validation_description_required",
	"skillregistry_validation_id_length",
	"skillregistry_validation_id_charset",
	"skillregistry_validation_name_length",
	"skillregistry_validation_description_length",
	"skillregistry_validation_version_semver",
	"skillregistry_validation_category_invalid",
	"skillregistry_validation_trigger_empty",
	"skillregistry_validation_trigger_too_long",
	"skillregistry_validation_tag_empty",
	"skillregistry_validation_tag_too_long",
	"skillregistry_validation_parameter_name_empty",
	"skillregistry_validation_parameter_name_duplicate",
	"skillregistry_validation_parameter_type_invalid",
	"skillregistry_executor_skill_not_active",
	"skillregistry_executor_pre_hook_failed",
	"skillregistry_executor_execution_failed",
	"skillregistry_executor_handler_nil_result",
	"skillregistry_executor_post_hook_failed",
	"skillregistry_executor_executing_skill",
	// round-378 §11.4 Phase 4 — CLI-agent registry descriptions
	"skillregistry_agent_desc_opencode",
	"skillregistry_agent_desc_crush",
	"skillregistry_agent_desc_helixcode",
	"skillregistry_agent_desc_kiro",
	"skillregistry_agent_desc_aider",
	"skillregistry_agent_desc_claudecode",
	"skillregistry_agent_desc_cline",
	"skillregistry_agent_desc_codenamegoose",
	"skillregistry_agent_desc_deepseekcli",
	"skillregistry_agent_desc_forge",
	"skillregistry_agent_desc_geminicli",
	"skillregistry_agent_desc_gptengineer",
	"skillregistry_agent_desc_kilocode",
	"skillregistry_agent_desc_mistralcode",
	"skillregistry_agent_desc_ollamacode",
	"skillregistry_agent_desc_plandex",
	"skillregistry_agent_desc_qwencode",
	"skillregistry_agent_desc_amazonq",
	"skillregistry_agent_desc_agentdeck",
	"skillregistry_agent_desc_bridle",
	"skillregistry_agent_desc_cheshirecat",
	"skillregistry_agent_desc_claudeplugins",
	"skillregistry_agent_desc_claudesquad",
	"skillregistry_agent_desc_codai",
	"skillregistry_agent_desc_codex",
	"skillregistry_agent_desc_codexskills",
	"skillregistry_agent_desc_conduit",
	"skillregistry_agent_desc_emdash",
	"skillregistry_agent_desc_fauxpilot",
	"skillregistry_agent_desc_getshitdone",
	"skillregistry_agent_desc_githubcopilotcli",
	"skillregistry_agent_desc_githubspeckit",
	"skillregistry_agent_desc_gitmcp",
	"skillregistry_agent_desc_gptme",
	"skillregistry_agent_desc_mobileagent",
	"skillregistry_agent_desc_multiagentcoding",
	"skillregistry_agent_desc_nanocoder",
	"skillregistry_agent_desc_noi",
	"skillregistry_agent_desc_octogen",
	"skillregistry_agent_desc_openhands",
	"skillregistry_agent_desc_postgresmcp",
	"skillregistry_agent_desc_shai",
	"skillregistry_agent_desc_snowcli",
	"skillregistry_agent_desc_taskweaver",
	"skillregistry_agent_desc_uiuxpromax",
	"skillregistry_agent_desc_vtcode",
	"skillregistry_agent_desc_warp",
	"skillregistry_agent_desc_continue",
}

// TestI18n_AgentLocalizedDescription_ResolvesRealCopy proves the
// round-378 CONST-046 migration: every CLI agent in CLIAgentRegistry
// carries a message-ID Description that LocalizedDescription resolves
// to real English copy via the wired bundle Translator — never a
// leaked raw ID, never an empty string. This is the paired-mutation
// guard: deleting an agent's skillregistry_agent_desc_* entry from
// active.en.yaml makes tr() echo the ID and this test FAILs.
func TestI18n_AgentLocalizedDescription_ResolvesRealCopy(t *testing.T) {
	if len(CLIAgentRegistry) == 0 {
		t.Fatal("CLIAgentRegistry is empty — registry not populated")
	}
	for name, agent := range CLIAgentRegistry {
		if !strings.HasPrefix(agent.Description, "skillregistry_agent_desc_") {
			t.Errorf("agent %q Description %q is not a CONST-046 message ID", name, agent.Description)
			continue
		}
		got := agent.LocalizedDescription()
		if got == "" {
			t.Errorf("agent %q LocalizedDescription returned empty string", name)
			continue
		}
		if got == agent.Description {
			t.Errorf("agent %q LocalizedDescription echoed raw message ID %q — bundle entry missing", name, agent.Description)
			continue
		}
		if strings.HasPrefix(got, "skillregistry_") {
			t.Errorf("agent %q LocalizedDescription %q leaked a raw message ID prefix", name, got)
		}
	}
}

// TestI18n_AgentLocalizedDescription_NilSafe asserts the accessor is
// nil-safe — a nil *CLIAgent returns "" rather than panicking.
func TestI18n_AgentLocalizedDescription_NilSafe(t *testing.T) {
	var a *CLIAgent
	if got := a.LocalizedDescription(); got != "" {
		t.Fatalf("nil CLIAgent.LocalizedDescription() = %q, want empty string", got)
	}
}

// TestI18n_AgentLocalizedDescription_KnownAgents spot-checks a few
// well-known agents resolve to their expected English copy — proving
// the message-ID → bundle-text mapping is correct, not just non-empty.
func TestI18n_AgentLocalizedDescription_KnownAgents(t *testing.T) {
	cases := map[string]string{
		"Aider":      "AI pair programming in your terminal",
		"ClaudeCode": "Anthropic's official CLI for Claude",
		"OpenCode":   "OpenCode AI coding assistant",
	}
	for name, want := range cases {
		agent, ok := GetAgent(name)
		if !ok {
			t.Errorf("GetAgent(%q) not found", name)
			continue
		}
		if got := agent.LocalizedDescription(); got != want {
			t.Errorf("agent %q LocalizedDescription = %q, want %q", name, got, want)
		}
	}
}

// TestI18n_AllMigratedIDsHaveBundleEntries is the round-333
// paired-mutation guard: every migrated message ID MUST resolve to
// real copy. Deleting an entry from active.en.yaml (the mutation)
// makes T return an error for that ID and this test FAILs — proving
// the test actually verifies the bundle, not a tautology.
func TestI18n_AllMigratedIDsHaveBundleEntries(t *testing.T) {
	bt, err := NewBundleTranslatorFromFile("en", testBundlePath)
	if err != nil {
		t.Fatalf("NewBundleTranslatorFromFile: %v", err)
	}
	for _, id := range migratedMessageIDs {
		out, terr := bt.T(context.Background(), id, nil)
		if terr != nil {
			t.Errorf("message ID %q missing from bundle: %v", id, terr)
			continue
		}
		if out == id || strings.TrimSpace(out) == "" {
			t.Errorf("message ID %q resolved to echo/empty %q — bundle entry missing or blank", id, out)
		}
	}
}

// TestI18n_ValidationErrorRendersRealCopy proves the end-to-end seam:
// a real ValidateSkill failure surfaces translated English text, not
// an echoed message ID — confirming validator.go's tr() callsites are
// wired to the bundle.
func TestI18n_ValidationErrorRendersRealCopy(t *testing.T) {
	v := NewSkillValidator()
	err := v.ValidateSkill(&Skill{ID: "", Name: "x", Description: "long enough description here"})
	if err == nil {
		t.Fatal("ValidateSkill(empty ID) returned nil, want error")
	}
	if !strings.Contains(err.Error(), "skill ID is required") {
		t.Fatalf("ValidateSkill error = %q, want translated 'skill ID is required'", err.Error())
	}
	if strings.Contains(err.Error(), "skillregistry_validation_") {
		t.Fatalf("ValidateSkill error %q leaked a raw message ID — translator not wired", err.Error())
	}
}
