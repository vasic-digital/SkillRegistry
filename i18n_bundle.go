// i18n_bundle.go provides a self-contained, dependency-light
// Translator backed by the YAML locale bundles shipped in
// i18n/bundles/ (CONST-046, round-333 §11.4 sweep, 2026-05-19).
//
// It exists so SkillRegistry is usable end-to-end WITHOUT forcing a
// consuming application to pull in a heavyweight go-i18n stack: a
// caller may either inject its own Translator (CONST-051(B)
// configuration injection) or load this BundleTranslator off the
// in-tree bundle. The challenge runner and unit tests use the latter
// so migrated user-facing strings render as real locale text rather
// than echoed message IDs.
//
// Template interpolation uses Go's text/template with go-i18n style
// {{.Var}} placeholders — the same placeholder syntax the YAML
// bundle declares — so the bundle format stays compatible with a
// full go-i18n adapter a consumer might swap in later.
package agents

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// bundleEntry is one message in a YAML locale bundle. The go-i18n
// "other" plural form is the only form SkillRegistry's migrated
// strings use today.
type bundleEntry struct {
	Other string `yaml:"other"`
}

// BundleTranslator resolves message IDs against an in-memory map
// parsed from a YAML locale bundle. It satisfies Translator and is
// safe for concurrent reads after construction (the message map is
// never mutated after NewBundleTranslator returns).
type BundleTranslator struct {
	locale   string
	messages map[string]string
}

// NewBundleTranslatorFromFile parses the YAML bundle at path and
// returns a BundleTranslator. The bundle maps message IDs to entries
// with an "other" form (go-i18n compatible). locale is a label used
// for diagnostics only.
func NewBundleTranslatorFromFile(locale, path string) (*BundleTranslator, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("skillregistry i18n: read bundle %q: %w", path, err)
	}
	return NewBundleTranslatorFromBytes(locale, raw)
}

// NewBundleTranslatorFromBytes parses raw YAML bundle bytes. Useful
// for embedding the bundle or loading it from a non-file source.
func NewBundleTranslatorFromBytes(locale string, raw []byte) (*BundleTranslator, error) {
	var doc map[string]bundleEntry
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("skillregistry i18n: parse bundle: %w", err)
	}
	msgs := make(map[string]string, len(doc))
	for id, entry := range doc {
		msgs[id] = entry.Other
	}
	return &BundleTranslator{locale: locale, messages: msgs}, nil
}

// Locale returns the locale label this translator was built with.
func (b *BundleTranslator) Locale() string { return b.locale }

// MessageIDs returns the sorted set of message IDs the bundle
// defines. Used by the round-333 paired-mutation test to assert
// every migrated tr() call has a backing bundle entry.
func (b *BundleTranslator) MessageIDs() []string {
	ids := make([]string, 0, len(b.messages))
	for id := range b.messages {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// T resolves messageID against the bundle, interpolating templateData
// via text/template. An unknown messageID returns the messageID
// itself (loud echo) plus an error so a misconfigured bundle degrades
// visibly — never a silent empty string (§11.4 PASS-bluff guard).
func (b *BundleTranslator) T(_ context.Context, messageID string, templateData map[string]any) (string, error) {
	pattern, ok := b.messages[messageID]
	if !ok {
		return messageID, fmt.Errorf("skillregistry i18n: unknown message ID %q (locale %s)", messageID, b.locale)
	}
	if !strings.Contains(pattern, "{{") || len(templateData) == 0 {
		return pattern, nil
	}
	tmpl, err := template.New(messageID).Parse(pattern)
	if err != nil {
		return messageID, fmt.Errorf("skillregistry i18n: parse template %q: %w", messageID, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return messageID, fmt.Errorf("skillregistry i18n: render %q: %w", messageID, err)
	}
	return buf.String(), nil
}
