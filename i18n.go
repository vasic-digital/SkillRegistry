// i18n.go declares SkillRegistry's hardcoded-content abstraction per
// CONST-046 (round-333 §11.4 anti-bluff sweep, 2026-05-19). It mirrors
// the "consumer defines its own Translator interface" pattern of every
// prior CONST-046-migrated package in the HelixCode codebase (the
// reference is helix_code/internal/approval/i18n/translator.go).
//
// CONST-051(B) decoupling: this seam is fully project-not-aware — it
// declares only an interface plus a safe NoopTranslator default. A
// consuming binary (HelixCode, HelixAgent, or any third-party) builds
// a real Translator over its own i18n bundle and injects it via
// SetTranslator at boot. SkillRegistry itself never reaches into a
// parent project's tree.
//
// Wire-in path at boot: consuming binary loads its locale bundle (the
// reference bundle ships at i18n/bundles/active.en.yaml), wraps it in
// an adapter that satisfies Translator, and calls SetTranslator. The
// package-level tr() helper falls back to NoopTranslator{} when not
// wired — loud message-ID echo, never a silent swallow (which would be
// a §11.4 PASS-bluff at the i18n layer).
package agents

import (
	"context"
	"sync"
)

// Translator is the contract SkillRegistry uses for every
// CONST-046-migrated user-facing string. A consuming application
// supplies a real implementation (typically a go-i18n localizer
// adapter); unit tests within this package use NoopTranslator.
type Translator interface {
	// T resolves messageID against the active locale. templateData
	// supplies named placeholders for go-i18n style interpolation;
	// pass nil when the message has no placeholders.
	T(ctx context.Context, messageID string, templateData map[string]any) (string, error)
}

// NoopTranslator returns the messageID verbatim. SAFETY default for
// unit tests within this package + backward-compat for callers who
// have not yet wired a real Translator. Production paths SHOULD inject
// a real Translator via SetTranslator at boot.
type NoopTranslator struct{}

// T returns id unchanged (loud echo). Never returns an error.
func (NoopTranslator) T(_ context.Context, id string, _ map[string]any) (string, error) {
	return id, nil
}

var (
	translatorMu     sync.RWMutex
	activeTranslator Translator = NoopTranslator{}
)

// SetTranslator installs the process-wide Translator. Passing nil
// resets to NoopTranslator so a caller can never accidentally wire a
// nil dereference. Safe for concurrent use.
func SetTranslator(t Translator) {
	translatorMu.Lock()
	defer translatorMu.Unlock()
	if t == nil {
		activeTranslator = NoopTranslator{}
		return
	}
	activeTranslator = t
}

// currentTranslator returns the installed Translator under a read
// lock. Always non-nil.
func currentTranslator() Translator {
	translatorMu.RLock()
	defer translatorMu.RUnlock()
	return activeTranslator
}

// tr resolves messageID against the active Translator. On any
// translator error it falls back to the messageID itself (loud echo)
// so a misconfigured bundle degrades visibly instead of producing an
// empty user-facing string. data carries named placeholders; pass nil
// when the message has none.
func tr(messageID string, data map[string]any) string {
	out, err := currentTranslator().T(context.Background(), messageID, data)
	if err != nil || out == "" {
		return messageID
	}
	return out
}
