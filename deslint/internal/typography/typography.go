// Package typography evaluates profile-scoped provider typography evidence.
package typography

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Evidence contains one surface's profile-owned typography thresholds and nodes.
type Evidence struct {
	Schema        string                  `json:"$schema,omitempty"`
	SchemaVersion int                     `json:"schemaVersion"`
	EvidenceKind  diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform      string                  `json:"platform"`
	ProfileID     string                  `json:"profileId"`
	SurfaceID     string                  `json:"surfaceId"`
	FontScale     float64                 `json:"fontScale"`
	Policy        Thresholds              `json:"policy"`
	Nodes         []Node                  `json:"nodes"`
}

// Thresholds keeps surface-specific criteria out of the generic engine.
type Thresholds struct {
	MinimumHeadingRatio         float64 `json:"minimumHeadingRatio"`
	MinimumBodySize             float64 `json:"minimumBodySize"`
	MinimumFunctionalSize       float64 `json:"minimumFunctionalSize"`
	PlatformFunctionalFloor     float64 `json:"platformFunctionalFloor"`
	MinimumLineHeightRatio      float64 `json:"minimumLineHeightRatio"`
	MaximumBodyTrackingEM       float64 `json:"maximumBodyTrackingEm"`
	MaximumDisplayViewportRatio float64 `json:"maximumDisplayViewportRatio"`
}

// Node is one normalized semantic text run.
type Node struct {
	ID                   string  `json:"id"`
	Owner                string  `json:"owner"`
	Order                int     `json:"order"`
	SemanticRole         string  `json:"semanticRole"`
	FontFamily           string  `json:"fontFamily"`
	PhysicalVariant      string  `json:"physicalVariant"`
	FontWeight           int     `json:"fontWeight"`
	FontSize             float64 `json:"fontSize"`
	LineHeight           float64 `json:"lineHeight"`
	LetterSpacingEM      float64 `json:"letterSpacingEm"`
	Text                 string  `json:"text"`
	Lines                int     `json:"lines"`
	Italic               bool    `json:"italic"`
	Serif                bool    `json:"serif"`
	AllCaps              bool    `json:"allCaps"`
	HeadingLevel         int     `json:"headingLevel"`
	AccessibilityHeading bool    `json:"accessibilityHeading"`
	ViewportHeight       float64 `json:"viewportHeight"`
	RenderedHeight       float64 `json:"renderedHeight"`
	IconTileBefore       bool    `json:"iconTileBefore"`
	EyebrowChipBefore    bool    `json:"eyebrowChipBefore"`
	KickerBefore         bool    `json:"kickerBefore"`
	ApprovedFont         bool    `json:"approvedFont"`
	ApprovedVariant      bool    `json:"approvedVariant"`
	BodyReferenceSize    float64 `json:"bodyReferenceSize"`
}

var ruleIDs = []string{rules.RuleTypographyOverusedFont, rules.RuleTypographyFlatTypeHierarchy, rules.RuleTypographyIconTileStack, rules.RuleTypographyItalicSerifDisplay, rules.RuleTypographyHeroEyebrowChip, rules.RuleTypographyKickerAboveHeading, rules.RuleTypographyOversizedH1, rules.RuleTypographyExtremeNegativeTracking, rules.RuleTypographyTightLeading, rules.RuleTypographyTinyText, rules.RuleTypographyUndersizedUIText, rules.RuleTypographyAllCapsBody, rules.RuleTypographyWideTracking, rules.RuleTypographySkippedHeading}

// Analyze strictly parses and evaluates typography evidence.
func Analyze(path string, contents []byte, severity func(string) diagnostic.Severity, active func(string) bool) (Evidence, []diagnostic.Diagnostic, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, fmt.Errorf("typography evidence has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, nil, fmt.Errorf("decode typography evidence: %w", err)
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
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Order != nodes[j].Order {
			return nodes[i].Order < nodes[j].Order
		}
		return nodes[i].ID < nodes[j].ID
	})
	findings := make([]diagnostic.Diagnostic, 0)
	rolesByFamily := map[string]map[string]bool{}
	for _, n := range nodes {
		family := strings.ToLower(n.FontFamily)
		if rolesByFamily[family] == nil {
			rolesByFamily[family] = map[string]bool{}
		}
		rolesByFamily[family][n.SemanticRole] = true
	}
	previousHeading := 0
	for _, n := range nodes {
		p := path + "#/nodes/" + n.ID
		display := n.SemanticRole == "display" || n.SemanticRole == "heading"
		body := n.SemanticRole == "body"
		functional := slices.Contains([]string{"ui", "label", "metadata"}, n.SemanticRole)
		findings = add(findings, !n.ApprovedFont || !n.ApprovedVariant, rules.RuleTypographyOverusedFont, "font family or physical variant is outside the approved registry", p, n, evidence, severity, active, nil)
		if familyIsOverused(n.FontFamily) && display && rolesByFamily[strings.ToLower(n.FontFamily)]["body"] {
			findings = add(findings, true, rules.RuleTypographyOverusedFont, "overused family is combined across display and body roles", p, n, evidence, severity, active, []string{"hallmark-eight-02"})
		}
		findings = add(findings, display && n.FontSize/n.BodyReferenceSize < evidence.Policy.MinimumHeadingRatio, rules.RuleTypographyFlatTypeHierarchy, "heading-to-body ratio is below the surface profile", p, n, evidence, severity, active, nil)
		findings = add(findings, display && n.IconTileBefore, rules.RuleTypographyIconTileStack, "rounded icon tile is stacked above a heading", p, n, evidence, severity, active, nil)
		findings = add(findings, display && n.Italic && n.Serif && !n.ApprovedVariant, rules.RuleTypographyItalicSerifDisplay, "italic serif display is not an approved variant", p, n, evidence, severity, active, nil)
		findings = add(findings, display && n.EyebrowChipBefore, rules.RuleTypographyHeroEyebrowChip, "eyebrow chip appears above hero text", p, n, evidence, severity, active, nil)
		findings = add(findings, display && n.KickerBefore, rules.RuleTypographyKickerAboveHeading, "tracked kicker appears above heading", p, n, evidence, severity, active, nil)
		findings = add(findings, display && len([]rune(n.Text)) > 60 && n.ViewportHeight > 0 && n.RenderedHeight/n.ViewportHeight > evidence.Policy.MaximumDisplayViewportRatio, rules.RuleTypographyOversizedH1, "sentence-length display occupies too much of the first viewport", p, n, evidence, severity, active, nil)
		findings = add(findings, n.LetterSpacingEM < -0.08, rules.RuleTypographyExtremeNegativeTracking, "negative tracking exceeds glyph-safe threshold", p, n, evidence, severity, active, nil)
		findings = add(findings, n.Lines > 1 && n.LineHeight/n.FontSize < evidence.Policy.MinimumLineHeightRatio, rules.RuleTypographyTightLeading, "multi-line leading is below the role threshold", p, n, evidence, severity, active, nil)
		findings = add(findings, body && n.FontSize < evidence.Policy.MinimumBodySize, rules.RuleTypographyTinyText, "body text is below the profile minimum", p, n, evidence, severity, active, nil)
		floor := math.Max(evidence.Policy.MinimumFunctionalSize, evidence.Policy.PlatformFunctionalFloor)
		findings = add(findings, functional && n.FontSize < floor, rules.RuleTypographyUndersizedUIText, "functional text is below the stronger platform or consumer floor", p, n, evidence, severity, active, nil)
		findings = add(findings, body && n.AllCaps && len([]rune(strings.TrimSpace(n.Text))) >= 20, rules.RuleTypographyAllCapsBody, "long body copy is all caps", p, n, evidence, severity, active, nil)
		findings = add(findings, body && n.LetterSpacingEM > evidence.Policy.MaximumBodyTrackingEM, rules.RuleTypographyWideTracking, "body tracking exceeds the profile maximum", p, n, evidence, severity, active, nil)
		if n.HeadingLevel > 0 || n.AccessibilityHeading {
			if previousHeading > 0 && n.HeadingLevel > previousHeading+1 {
				findings = add(findings, true, rules.RuleTypographySkippedHeading, "heading order skips a semantic level", p, n, evidence, severity, active, nil)
			}
			if n.HeadingLevel > 0 {
				previousHeading = n.HeadingLevel
			}
		}
	}
	diagnostic.Sort(findings)
	return evidence, diagnostic.MergeCanonical(findings), nil
}

func add(values []diagnostic.Diagnostic, condition bool, ruleID, message, path string, node Node, e Evidence, severity func(string) diagnostic.Severity, active func(string) bool, extra []string) []diagnostic.Diagnostic {
	if !condition || !active(ruleID) {
		return values
	}
	source := strings.TrimPrefix(ruleID, "typography/")
	sources := append([]string{source}, extra...)
	return append(values, diagnostic.NewWithSources(ruleID, sources, severity(ruleID), message, path, nil, e.EvidenceKind, e.Platform, node.Owner, "typography"))
}
func familyIsOverused(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	return v == "inter" || v == "roboto"
}
func validate(e Evidence) error {
	if e.SchemaVersion != 1 || e.ProfileID == "" || e.SurfaceID == "" || e.FontScale < 1 || len(e.Nodes) == 0 {
		return fmt.Errorf("typography evidence identity, scale, policy, and nodes are required")
	}
	valid := []diagnostic.EvidenceKind{diagnostic.EvidenceWebRendered, diagnostic.EvidenceNativeSource, diagnostic.EvidenceDesignDocumentComputed, diagnostic.EvidenceSimulator, diagnostic.EvidenceEmulator, diagnostic.EvidencePhysicalDevice}
	if !slices.Contains(valid, e.EvidenceKind) {
		return fmt.Errorf("typography evidence kind %q is invalid", e.EvidenceKind)
	}
	seen := map[string]bool{}
	for _, n := range e.Nodes {
		if n.ID == "" || n.Owner == "" || seen[n.ID] || n.FontSize <= 0 || n.LineHeight <= 0 || n.BodyReferenceSize <= 0 {
			return fmt.Errorf("typography node identity and metrics are invalid")
		}
		seen[n.ID] = true
	}
	return nil
}

// RuleIDs returns the exact typography rule membership for tests and adapters.
func RuleIDs() []string { return append([]string(nil), ruleIDs...) }
