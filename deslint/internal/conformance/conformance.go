// Package conformance evaluates product-neutral consumer component inventories.
package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

var requiredStates = []string{"default", "disabled", "error", "focused", "loading", "pressed"}

// Config contains the narrow profile adjustments allowed during conformance analysis.
type Config struct {
	ProfileID              string
	MaxOversizedActions    int
	MaxInconsistentActions int
	Severity               func(string) diagnostic.Severity
	Active                 func(string) bool
}

// Result contains the parsed evidence identity and deterministic findings.
type Result struct {
	ProfileID   string
	Platform    string
	SurfaceID   string
	Diagnostics []diagnostic.Diagnostic
}

// Analyze strictly parses and evaluates one consumer conformance payload.
func Analyze(path string, contents []byte, config Config) (Result, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Result{}, fmt.Errorf("parse consumer conformance JSON: %w", err)
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Result{}, fmt.Errorf("consumer conformance has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Result{}, fmt.Errorf("decode consumer conformance: %w", err)
	}
	if err := validate(evidence, config.ProfileID); err != nil {
		return Result{}, err
	}
	severity := config.Severity
	if severity == nil {
		severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	active := config.Active
	if active == nil {
		active = func(string) bool { return true }
	}
	controls := append([]Control(nil), evidence.Controls...)
	sort.SliceStable(controls, func(i, j int) bool { return controls[i].ID < controls[j].ID })
	findings := make([]diagnostic.Diagnostic, 0)
	oversized := make([]Control, 0)
	byAction := make(map[string][]Control)
	for _, control := range controls {
		byAction[control.ActionID] = append(byAction[control.ActionID], control)
		controlPath := path + "#/controls/" + control.ID
		if control.Prominence == "oversized" {
			oversized = append(oversized, control)
		}
		findings = appendIf(findings, active, control.ContractStatus == "mismatched", rules.RuleProfileMismatchedControl,
			fmt.Sprintf("control %s does not match component contract %s", control.ID, control.Component), controlPath, evidence.Platform, severity)
		approvedMotionPermission := control.MotionRecipeStatus == "approved" && control.ReduceMotionFallback
		findings = appendIf(findings, active, control.MotionPurpose == "decorative" && !approvedMotionPermission, rules.RuleProfileGratuitousMotion,
			fmt.Sprintf("control %s uses motion without state or continuity purpose", control.ID), controlPath, evidence.Platform, severity)
		findings = appendIf(findings, active, control.AffordanceSource == "invented", rules.RuleProfileInventedAffordance,
			fmt.Sprintf("control %s uses an invented affordance", control.ID), controlPath, evidence.Platform, severity)
		missing := missingStates(control.States)
		findings = appendIf(findings, active, len(missing) > 0, rules.RuleProfileMissingState,
			fmt.Sprintf("control %s is missing states: %s", control.ID, strings.Join(missing, ", ")), controlPath, evidence.Platform, severity)
	}
	if active(rules.RuleProfileExaggeratedButton) && len(oversized) > config.MaxOversizedActions {
		for _, control := range oversized[config.MaxOversizedActions:] {
			findings = append(findings, newFinding(rules.RuleProfileExaggeratedButton,
				fmt.Sprintf("oversized action %s exceeds profile threshold %d", control.ID, config.MaxOversizedActions),
				path+"#/controls/"+control.ID, evidence.Platform, severity))
		}
	}
	actionIDs := make([]string, 0, len(byAction))
	for actionID := range byAction {
		actionIDs = append(actionIDs, actionID)
	}
	sort.Strings(actionIDs)
	inconsistent := make([]string, 0)
	for _, actionID := range actionIDs {
		if !consistent(byAction[actionID]) {
			inconsistent = append(inconsistent, actionID)
		}
	}
	if active(rules.RuleProfileInconsistentAction) && len(inconsistent) > config.MaxInconsistentActions {
		for _, actionID := range inconsistent[config.MaxInconsistentActions:] {
			findings = append(findings, newFinding(rules.RuleProfileInconsistentAction,
				fmt.Sprintf("action %s has inconsistent shape, label, icon, or feedback", actionID),
				path+"#/actions/"+actionID, evidence.Platform, severity))
		}
	}
	diagnostic.Sort(findings)
	return Result{ProfileID: evidence.ProfileID, Platform: evidence.Platform, SurfaceID: evidence.SurfaceID, Diagnostics: findings}, nil
}

func appendIf(findings []diagnostic.Diagnostic, active func(string) bool, condition bool, ruleID, message, path, platform string, severity func(string) diagnostic.Severity) []diagnostic.Diagnostic {
	if condition && active(ruleID) {
		return append(findings, newFinding(ruleID, message, path, platform, severity))
	}
	return findings
}

func newFinding(ruleID, message, path, platform string, severity func(string) diagnostic.Severity) diagnostic.Diagnostic {
	return diagnostic.New(ruleID, severity(ruleID), message, path, nil, diagnostic.EvidenceConsumerConformance, platform, "ansldes/profile", "conformance")
}

func missingStates(states []string) []string {
	result := make([]string, 0)
	for _, required := range requiredStates {
		if !slices.Contains(states, required) {
			result = append(result, required)
		}
	}
	return result
}

func consistent(controls []Control) bool {
	if len(controls) < 2 {
		return true
	}
	first := controls[0]
	for _, control := range controls[1:] {
		if control.ShapeToken != first.ShapeToken || control.Label != first.Label ||
			control.Icon != first.Icon || control.Feedback != first.Feedback {
			return false
		}
	}
	return true
}

func validate(value Evidence, configuredProfile string) error {
	if value.SchemaVersion != SchemaVersion {
		return fmt.Errorf("consumer conformance schemaVersion = %d, want %d", value.SchemaVersion, SchemaVersion)
	}
	if value.ProfileID == "" || value.Platform == "" || value.SurfaceID == "" || len(value.Controls) == 0 {
		return fmt.Errorf("consumer conformance profileId, platform, surfaceId, and controls are required")
	}
	if configuredProfile != "" && value.ProfileID != configuredProfile {
		return fmt.Errorf("consumer conformance profileId %q does not match policy profile %q", value.ProfileID, configuredProfile)
	}
	validPlatforms := []string{"web", "react-native", "ios", "android", "design-document"}
	if !slices.Contains(validPlatforms, value.Platform) {
		return fmt.Errorf("consumer conformance platform %q is invalid", value.Platform)
	}
	seen := make(map[string]bool)
	for _, control := range value.Controls {
		if control.ID == "" || control.ActionID == "" || control.Role == "" || control.Component == "" ||
			control.Label == "" || control.ShapeToken == "" || control.Feedback == "" {
			return fmt.Errorf("consumer conformance control identities and contract fields are required")
		}
		if seen[control.ID] {
			return fmt.Errorf("consumer conformance control id %q is duplicated", control.ID)
		}
		seen[control.ID] = true
		if !slices.Contains([]string{"primary-action", "secondary-action", "input", "selection", "navigation", "feedback", "overlay"}, control.Role) ||
			!slices.Contains([]string{"matched", "mismatched"}, control.ContractStatus) ||
			!slices.Contains([]string{"design-system", "platform", "consumer-exception", "invented"}, control.AffordanceSource) ||
			!slices.Contains([]string{"none", "state-transition", "continuity", "decorative"}, control.MotionPurpose) ||
			!slices.Contains([]string{"none", "approved", "unapproved"}, control.MotionRecipeStatus) ||
			!slices.Contains([]string{"standard", "emphasized", "oversized"}, control.Prominence) {
			return fmt.Errorf("consumer conformance control %q has an invalid contract enum", control.ID)
		}
		if control.AffordanceSource == "consumer-exception" && control.ExceptionID == "" {
			return fmt.Errorf("consumer exception control %q requires exceptionId", control.ID)
		}
		stateSeen := make(map[string]bool)
		for _, state := range control.States {
			if !slices.Contains([]string{"default", "pressed", "focused", "disabled", "loading", "error", "selected"}, state) {
				return fmt.Errorf("consumer conformance control %q has invalid state %q", control.ID, state)
			}
			if stateSeen[state] {
				return fmt.Errorf("consumer conformance control %q duplicates state %q", control.ID, state)
			}
			stateSeen[state] = true
		}
	}
	return nil
}
