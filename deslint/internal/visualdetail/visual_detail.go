// Package visualdetail evaluates provider-neutral visual-detail evidence.
package visualdetail

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

// Evidence is one provider-owned visual detail inventory.
type Evidence struct {
	Schema        string                  `json:"$schema,omitempty"`
	SchemaVersion int                     `json:"schemaVersion"`
	EvidenceKind  diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform      string                  `json:"platform"`
	SurfaceID     string                  `json:"surfaceId"`
	Nodes         []Node                  `json:"nodes"`
}

// Node is the normalized structural and visual state of one provider node.
type Node struct {
	ID                 string       `json:"id"`
	Owner              string       `json:"owner"`
	Kind               string       `json:"kind"`
	Radius             float64      `json:"radius"`
	BorderWidth        float64      `json:"borderWidth"`
	ShadowBlur         float64      `json:"shadowBlur"`
	ShadowSpread       float64      `json:"shadowSpread"`
	AccentEdges        []AccentEdge `json:"accentEdges"`
	BackgroundPattern  string       `json:"backgroundPattern"`
	PatternDataMeaning bool         `json:"patternDataMeaning"`
	DiagramOwner       string       `json:"diagramOwner,omitempty"`
	SemanticState      string       `json:"semanticState,omitempty"`
	ComponentOwner     string       `json:"componentOwner,omitempty"`
	AccessoryWrapper   bool         `json:"accessoryWrapper"`
	DeclaresBorder     bool         `json:"declaresBorder"`
	DeclaresElevation  bool         `json:"declaresElevation"`
}

// AccentEdge is one chromatic edge and its physical direction.
type AccentEdge struct {
	Side      string  `json:"side"`
	Thickness float64 `json:"thickness"`
	ColorRole string  `json:"colorRole"`
}

var upstreamIDs = map[string]string{
	rules.RuleVisualSideTab:              "side-tab",
	rules.RuleVisualBorderAccentRounded:  "border-accent-on-rounded",
	rules.RuleVisualThinBorderWideShadow: "gpt-thin-border-wide-shadow",
	rules.RuleVisualRepeatingStripes:     "repeating-stripes-gradient",
	rules.RuleVisualGridBackground:       "codex-grid-background",
}

// Analyze strictly parses one payload and returns deterministic findings.
func Analyze(path string, contents []byte, severity func(string) diagnostic.Severity, active func(string) bool) (Evidence, []diagnostic.Diagnostic, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, fmt.Errorf("parse visual detail JSON: %w", err)
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, fmt.Errorf("visual detail has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, nil, fmt.Errorf("decode visual detail: %w", err)
	}
	if err := validate(evidence); err != nil {
		return Evidence{}, nil, err
	}
	if severity == nil {
		severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if active == nil {
		active = func(string) bool { return true }
	}
	nodes := append([]Node(nil), evidence.Nodes...)
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	findings := make([]diagnostic.Diagnostic, 0)
	for _, node := range nodes {
		nodePath := path + "#/nodes/" + node.ID
		hasAccent := false
		for _, edge := range node.AccentEdges {
			if edge.Thickness >= 2 && edge.ColorRole != "transparent" {
				hasAccent = true
				break
			}
		}
		exemptState := (node.Kind == "status" || node.Kind == "selection") && node.SemanticState != "" && node.ComponentOwner == node.Owner
		findings = appendIf(findings, active, hasAccent && slices.Contains([]string{"card", "list", "callout"}, node.Kind) && !exemptState, rules.RuleVisualSideTab, "chromatic edge creates a side-tab treatment", nodePath, node.Owner, evidence, severity)
		findings = appendIf(findings, active, hasAccent && node.Radius > 0 && !exemptState, rules.RuleVisualBorderAccentRounded, "accent edge dominates a rounded surface", nodePath, node.Owner, evidence, severity)
		findings = appendIf(findings, active, node.DeclaresBorder && node.DeclaresElevation && node.BorderWidth > 0 && node.BorderWidth <= 1 && node.ShadowBlur >= 16, rules.RuleVisualThinBorderWideShadow, "hairline border is combined with wide diffuse elevation", nodePath, node.Owner, evidence, severity)
		patternPermission := slices.Contains([]string{"diagram", "canvas", "map", "measurement"}, node.Kind) && node.PatternDataMeaning && node.DiagramOwner == node.Owner
		findings = appendIf(findings, active, node.BackgroundPattern == "repeating-stripes" && !patternPermission, rules.RuleVisualRepeatingStripes, "repeating stripe background has no exact data meaning owner", nodePath, node.Owner, evidence, severity)
		findings = appendIf(findings, active, node.BackgroundPattern == "grid" && !patternPermission, rules.RuleVisualGridBackground, "grid background has no exact diagram or measurement owner", nodePath, node.Owner, evidence, severity)
		findings = appendIf(findings, active, node.Kind == "list-row" && node.AccessoryWrapper, rules.RuleNativeListRowAccessoryWrapper, "list-row accessory adds a separate tile or card wrapper", nodePath, node.Owner, evidence, severity)
	}
	diagnostic.Sort(findings)
	return evidence, findings, nil
}

func appendIf(findings []diagnostic.Diagnostic, active func(string) bool, condition bool, ruleID, message, path, owner string, evidence Evidence, severity func(string) diagnostic.Severity) []diagnostic.Diagnostic {
	if !condition || !active(ruleID) {
		return findings
	}
	sources := []string{}
	if upstream := upstreamIDs[ruleID]; upstream != "" {
		sources = []string{upstream}
	}
	return append(findings, diagnostic.NewWithSources(ruleID, sources, severity(ruleID), message, path, nil, evidence.EvidenceKind, evidence.Platform, owner, "visual-detail"))
}

func validate(evidence Evidence) error {
	if evidence.SchemaVersion != 1 || evidence.SurfaceID == "" || len(evidence.Nodes) == 0 {
		return fmt.Errorf("visual detail schemaVersion, surfaceId, and nodes are required")
	}
	validKinds := []diagnostic.EvidenceKind{diagnostic.EvidenceWebSource, diagnostic.EvidenceNativeSource, diagnostic.EvidenceDesignDocumentSource}
	if !slices.Contains(validKinds, evidence.EvidenceKind) {
		return fmt.Errorf("visual detail evidence kind %q is invalid", evidence.EvidenceKind)
	}
	if evidence.EvidenceKind == diagnostic.EvidenceWebSource && evidence.Platform != "web" {
		return fmt.Errorf("web visual detail evidence requires web platform")
	}
	if evidence.EvidenceKind == diagnostic.EvidenceDesignDocumentSource && evidence.Platform != "design-document" {
		return fmt.Errorf("design-document visual detail evidence requires design-document platform")
	}
	seen := make(map[string]bool)
	for _, node := range evidence.Nodes {
		if node.ID == "" || node.Owner == "" || seen[node.ID] {
			return fmt.Errorf("visual detail node identity and unique owner are required")
		}
		seen[node.ID] = true
		if !slices.Contains([]string{"card", "list", "callout", "status", "selection", "list-row", "diagram", "canvas", "map", "measurement", "surface"}, node.Kind) || !slices.Contains([]string{"none", "repeating-stripes", "grid"}, node.BackgroundPattern) {
			return fmt.Errorf("visual detail node %q has invalid kind or pattern", node.ID)
		}
	}
	return nil
}
