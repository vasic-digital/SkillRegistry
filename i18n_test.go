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
