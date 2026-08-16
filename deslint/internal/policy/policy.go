package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Parse validates a product policy without importing product paths into the engine.
func Parse(contents []byte) (Policy, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Policy{}, fmt.Errorf("parse policy JSON: %w", err)
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Policy{}, fmt.Errorf("policy has duplicate keys: %s", strings.Join(duplicates, ", "))
	}

	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var result Policy
	if err := decoder.Decode(&result); err != nil {
		return Policy{}, fmt.Errorf("decode policy: %w", err)
	}
	if err := validate(result); err != nil {
		return Policy{}, err
	}
	return result, nil
}

// Severity returns the exact configured severity for a rule.
func (p Policy) Severity(ruleID string) diagnostic.Severity {
	if p.Severities[ruleID] == string(diagnostic.SeverityWarning) {
		return diagnostic.SeverityWarning
	}
	return diagnostic.SeverityError
}

// RawPropertyKinds maps each configured style property to its raw-value category.
func (p Policy) RawPropertyKinds() map[string]string {
	result := make(map[string]string)
	for _, property := range p.Source.RawProperties.Color {
		result[property] = "color"
	}
	for _, property := range p.Source.RawProperties.Number {
		result[property] = "number"
	}
	for _, property := range p.Source.RawProperties.Motion {
		result[property] = "motion"
	}
	return result
}

// IsExcluded returns true only for an exact normalized path match.
func (p Policy) IsExcluded(path string) bool {
	normalized := filepath.ToSlash(filepath.Clean(path))
	for _, excluded := range p.Source.ExactExcludes {
		if normalized == filepath.ToSlash(filepath.Clean(excluded)) {
			return true
		}
	}
	return false
}

// Requires reports whether an independently acquired evidence kind is mandatory.
func (p Policy) Requires(kind diagnostic.EvidenceKind) bool {
	for _, required := range p.Evidence.RequiredKinds {
		if required == string(kind) {
			return true
		}
	}
	return false
}

// ApplyExceptions removes only active, exact rule/path matches.
func (p Policy) ApplyExceptions(diagnostics []diagnostic.Diagnostic, now time.Time) []diagnostic.Diagnostic {
	result := make([]diagnostic.Diagnostic, 0, len(diagnostics))
	for _, finding := range diagnostics {
		if !p.excepted(finding, now) {
			result = append(result, finding)
		}
	}
	return result
}

// ExpiredExceptions returns deterministic policy entries that can no longer suppress findings.
func (p Policy) ExpiredExceptions(now time.Time) []Exception {
	result := make([]Exception, 0)
	for _, exception := range p.Exceptions {
		expiresAt, err := time.Parse("2006-01-02", exception.ExpiresAt)
		if err == nil && expiresAt.Before(midnightUTC(now)) {
			result = append(result, exception)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		return result[i].RuleID < result[j].RuleID
	})
	return result
}

func (p Policy) excepted(finding diagnostic.Diagnostic, now time.Time) bool {
	for _, exception := range p.Exceptions {
		expiresAt, err := time.Parse("2006-01-02", exception.ExpiresAt)
		if err != nil || expiresAt.Before(midnightUTC(now)) {
			continue
		}
		if exception.RuleID == finding.RuleID && filepath.ToSlash(filepath.Clean(exception.Path)) == filepath.ToSlash(filepath.Clean(finding.Path)) {
			return true
		}
	}
	return false
}

func validate(value Policy) error {
	if value.SchemaVersion != SchemaVersion {
		return fmt.Errorf("policy schemaVersion = %d, want %d", value.SchemaVersion, SchemaVersion)
	}
	if value.DefinitionID == "" {
		return fmt.Errorf("policy definitionId is required")
	}
	expectedRules := append([]string(nil), rules.ConfigurableRuleIDs...)
	sort.Strings(expectedRules)
	actualRules := make([]string, 0, len(value.Severities))
	for ruleID, severity := range value.Severities {
		if severity != string(diagnostic.SeverityError) && severity != string(diagnostic.SeverityWarning) {
			return fmt.Errorf("policy severity %s = %q", ruleID, severity)
		}
		actualRules = append(actualRules, ruleID)
	}
	sort.Strings(actualRules)
	if strings.Join(expectedRules, "\x00") != strings.Join(actualRules, "\x00") {
		return fmt.Errorf("policy severities must exactly match the v1 rule registry")
	}
	for _, excluded := range value.Source.ExactExcludes {
		cleaned := filepath.ToSlash(filepath.Clean(excluded))
		if excluded == "" || filepath.IsAbs(excluded) || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.ContainsAny(excluded, "*?[]") {
			return fmt.Errorf("policy exactExcludes entry %q is not an exact relative path", excluded)
		}
	}
	seenProperties := make(map[string]string)
	for category, properties := range map[string][]string{
		"color":  value.Source.RawProperties.Color,
		"number": value.Source.RawProperties.Number,
		"motion": value.Source.RawProperties.Motion,
	} {
		for _, property := range properties {
			if property == "" {
				return fmt.Errorf("policy raw property names must not be empty")
			}
			if previous, exists := seenProperties[property]; exists {
				return fmt.Errorf("policy raw property %q is duplicated across %s and %s", property, previous, category)
			}
			seenProperties[property] = category
		}
	}
	if value.Budgets.Error < 0 || value.Budgets.Warning < 0 || value.Budgets.Raw < 0 || value.Budgets.Overflow < 0 || value.Budgets.Overlap < 0 {
		return fmt.Errorf("policy budgets must be non-negative")
	}
	validEvidence := map[string]bool{
		string(diagnostic.EvidenceDefinition): true,
		string(diagnostic.EvidenceSource):     true,
		string(diagnostic.EvidencePencil):     true,
		string(diagnostic.EvidenceLayout):     true,
	}
	seenEvidence := make(map[string]bool)
	for _, kind := range value.Evidence.RequiredKinds {
		if !validEvidence[kind] || seenEvidence[kind] {
			return fmt.Errorf("policy required evidence %q is invalid or duplicated", kind)
		}
		seenEvidence[kind] = true
	}
	for _, exception := range value.Exceptions {
		if exception.RuleID == "" || exception.Path == "" || exception.Owner == "" || len(exception.Rationale) < 8 {
			return fmt.Errorf("policy exception must have ruleId, path, owner, and rationale")
		}
		if _, err := time.Parse("2006-01-02", exception.ExpiresAt); err != nil {
			return fmt.Errorf("policy exception expiry %q is invalid", exception.ExpiresAt)
		}
		if !contains(expectedRules, exception.RuleID) {
			return fmt.Errorf("policy exception ruleId %q is not configurable", exception.RuleID)
		}
	}
	return nil
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func midnightUTC(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
