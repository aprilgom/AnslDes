package rules

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// RuleEvolution governs one exact rule addition, removal, or replacement.
type RuleEvolution struct {
	RuleID            string
	Kind              string
	Owner             string
	MigrationPlan     string
	Tombstone         bool
	ReplacementRuleID string
	DefaultActivation string
}

// ValidatePackEvolution enforces semantic-version and migration records over exact member diffs.
func ValidatePackEvolution(previous, next RulePackSpec, changes []RuleEvolution) error {
	previousVersion, err := semanticVersion(previous.Version)
	if err != nil {
		return err
	}
	nextVersion, err := semanticVersion(next.Version)
	if err != nil {
		return err
	}
	if previous.ID == "" || previous.ID != next.ID || nextVersion == previousVersion {
		return fmt.Errorf("rule pack evolution requires one identity and a changed version")
	}
	previousMembers := memberSet(previous)
	nextMembers := memberSet(next)
	expected := map[string]string{}
	for ruleID := range nextMembers {
		if !previousMembers[ruleID] {
			expected[ruleID] = "added"
		}
	}
	for ruleID := range previousMembers {
		if !nextMembers[ruleID] {
			expected[ruleID] = "removed"
		}
	}
	seen := map[string]bool{}
	for _, change := range changes {
		if change.RuleID == "" || seen[change.RuleID] || change.Owner == "" || len(change.MigrationPlan) < 8 {
			return fmt.Errorf("rule pack evolution record is incomplete or duplicated")
		}
		seen[change.RuleID] = true
		expectedKind := expected[change.RuleID]
		if change.Kind == "replaced" {
			expectedKind = "removed"
		}
		kindMatches := change.Kind == expectedKind || change.Kind == "replaced" && expectedKind == "removed"
		if expectedKind == "" || !kindMatches {
			return fmt.Errorf("rule pack evolution record for %q does not match the exact member diff", change.RuleID)
		}
		if change.Kind == "added" {
			if nextVersion[0] == previousVersion[0] && nextVersion[1] <= previousVersion[1] || !slices.Contains([]string{"active", "disabled"}, change.DefaultActivation) {
				return fmt.Errorf("additive rule %q requires a minor version and explicit default activation", change.RuleID)
			}
		} else if nextVersion[0] <= previousVersion[0] || !change.Tombstone && (change.ReplacementRuleID == "" || !nextMembers[change.ReplacementRuleID]) {
			return fmt.Errorf("removed rule %q requires a major version and tombstone or exact replacement", change.RuleID)
		}
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("rule pack evolution records do not cover the exact member diff")
	}
	return nil
}

func semanticVersion(value string) ([3]int, error) {
	var result [3]int
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return result, fmt.Errorf("invalid semantic version %q", value)
	}
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return result, fmt.Errorf("invalid semantic version %q", value)
		}
		result[index] = number
	}
	return result, nil
}

func memberSet(pack RulePackSpec) map[string]bool {
	result := make(map[string]bool, len(pack.Rules))
	for _, spec := range pack.Rules {
		if spec.ID != "" {
			result[spec.ID] = true
		}
	}
	return result
}
