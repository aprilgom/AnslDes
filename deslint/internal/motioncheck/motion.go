// Package motioncheck evaluates source and runtime motion evidence.
package motioncheck

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

// Evidence records resolved motion behavior under one platform preference.
type Evidence struct {
	Schema                          string                  `json:"$schema,omitempty"`
	SchemaVersion                   int                     `json:"schemaVersion"`
	EvidenceKind                    diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform                        string                  `json:"platform"`
	ProfileID                       string                  `json:"profileId"`
	SurfaceID                       string                  `json:"surfaceId"`
	Preference                      string                  `json:"preference"`
	CapturePath                     string                  `json:"capturePath"`
	PreferenceResolvedBeforeEffects bool                    `json:"preferenceResolvedBeforeEffects"`
	EffectsReplayedAfterResolution  bool                    `json:"effectsReplayedAfterResolution"`
	Nodes                           []Node                  `json:"nodes"`
}

// Node is one resolved animation or side effect.
type Node struct {
	ID                           string    `json:"id"`
	Owner                        string    `json:"owner"`
	TransitionID                 string    `json:"transitionId,omitempty"`
	Purpose                      string    `json:"purpose"`
	MotionKind                   string    `json:"motionKind"`
	DurationMS                   float64   `json:"durationMs"`
	Easing                       []float64 `json:"easing"`
	EasingBehavior               string    `json:"easingBehavior"`
	AnimatedProperties           []string  `json:"animatedProperties"`
	Perpetual                    bool      `json:"perpetual"`
	UserControlled               bool      `json:"userControlled"`
	ChangingDataSource           bool      `json:"changingDataSource"`
	AccessibleLabel              string    `json:"accessibleLabel"`
	EditableInput                bool      `json:"editableInput"`
	HoverPurpose                 string    `json:"hoverPurpose"`
	ReducedMotionFallbackApplied bool      `json:"reducedMotionFallbackApplied"`
	ReducedStateUnderstandable   bool      `json:"reducedStateUnderstandable"`
	NextActionUnderstandable     bool      `json:"nextActionUnderstandable"`
}

// Transition is one resolved canonical-definition motion recipe.
type Transition struct {
	Owner             string
	Purpose           string
	DurationMS        float64
	ReducedDurationMS float64
	Easing            []float64
	ReducedFallback   string
}

// Config injects consumer registry ownership and governed activation.
type Config struct {
	ProfileID string
	Registry  map[string]Transition
	Severity  func(string) diagnostic.Severity
	Active    func(string) bool
}

var ruleIDs = []string{
	rules.RuleMotionBounceEasing,
	rules.RuleMotionPulsingDot,
	rules.RuleMotionBlinkingCursor,
	rules.RuleMotionMarquee,
	rules.RuleMotionLayoutTransition,
	rules.RuleMotionImageHoverTransform,
}

// Analyze strictly parses and evaluates one motion evidence payload.
func Analyze(path string, contents []byte, config Config) (Evidence, []diagnostic.Diagnostic, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, fmt.Errorf("motion evidence has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, nil, err
	}
	if evidence.SchemaVersion != 1 || evidence.ProfileID == "" || evidence.SurfaceID == "" || evidence.CapturePath == "" || len(evidence.Nodes) == 0 {
		return Evidence{}, nil, fmt.Errorf("motion evidence identity, capture, and nodes are required")
	}
	if !platformMatches(evidence.EvidenceKind, evidence.Platform) {
		return Evidence{}, nil, fmt.Errorf("motion evidence kind %q is incompatible with platform %q", evidence.EvidenceKind, evidence.Platform)
	}
	if evidence.Preference != "no-preference" && evidence.Preference != "reduce" {
		return Evidence{}, nil, fmt.Errorf("motion preference %q is invalid", evidence.Preference)
	}
	if !evidence.PreferenceResolvedBeforeEffects || evidence.EffectsReplayedAfterResolution {
		return Evidence{}, nil, fmt.Errorf("motion preference must resolve before effects and must not replay completed effects")
	}
	if config.ProfileID != "" && evidence.ProfileID != config.ProfileID {
		return Evidence{}, nil, fmt.Errorf("motion evidence profile %q does not match policy profile %q", evidence.ProfileID, config.ProfileID)
	}
	if config.Severity == nil {
		config.Severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if config.Active == nil {
		config.Active = func(string) bool { return true }
	}

	nodes := append([]Node(nil), evidence.Nodes...)
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	seen := map[string]bool{}
	findings := []diagnostic.Diagnostic{}
	for _, node := range nodes {
		if err := validateNode(node); err != nil {
			return Evidence{}, nil, fmt.Errorf("motion node %q: %w", node.ID, err)
		}
		if seen[node.ID] {
			return Evidence{}, nil, fmt.Errorf("motion evidence has duplicate node identity %q", node.ID)
		}
		seen[node.ID] = true
		if err := validateRegistry(node, evidence.Preference, config.Registry); err != nil {
			return Evidence{}, nil, fmt.Errorf("motion node %q: %w", node.ID, err)
		}
		if evidence.Preference == "reduce" && (!node.ReducedMotionFallbackApplied || !node.ReducedStateUnderstandable || !node.NextActionUnderstandable) {
			return Evidence{}, nil, fmt.Errorf("motion node %q lacks an understandable reduced-motion fallback", node.ID)
		}

		nodePath := path + "#/nodes/" + node.ID
		bounce := node.EasingBehavior == "bounce" || node.EasingBehavior == "elastic" || node.EasingBehavior == "overshoot"
		pulse := node.MotionKind == "status-pulse" && (!node.ChangingDataSource || strings.TrimSpace(node.AccessibleLabel) == "")
		caret := node.MotionKind == "fake-caret" && !node.EditableInput
		marquee := node.MotionKind == "marquee" && node.Perpetual && !node.UserControlled
		layout := node.MotionKind == "layout-transition" && node.DurationMS > 0 && hasLayoutProperty(node.AnimatedProperties)
		hover := node.MotionKind == "image-hover-transform" && node.HoverPurpose == "decorative" && slices.Contains(node.AnimatedProperties, "transform")
		findings = add(findings, bounce, rules.RuleMotionBounceEasing, "animation uses bounce, elastic, or overshoot easing", nodePath, node, evidence, config)
		findings = add(findings, pulse, rules.RuleMotionPulsingDot, "status pulse lacks both a changing-data source and accessible label", nodePath, node, evidence, config)
		findings = add(findings, caret, rules.RuleMotionBlinkingCursor, "blinking fake caret is outside an editable input", nodePath, node, evidence, config)
		findings = add(findings, marquee, rules.RuleMotionMarquee, "perpetual marquee has no user control", nodePath, node, evidence, config)
		findings = add(findings, layout, rules.RuleMotionLayoutTransition, "animation changes width, height, padding, or margin", nodePath, node, evidence, config)
		findings = add(findings, hover, rules.RuleMotionImageHoverTransform, "image hover transform has no functional purpose", nodePath, node, evidence, config)
		if node.Purpose == "decorative" && !bounce && !pulse && !caret && !marquee && !layout && !hover && !allowedDecorativeException(node) {
			return Evidence{}, nil, fmt.Errorf("motion node %q has decorative motion outside a deterministic permission or rule", node.ID)
		}
	}
	diagnostic.Sort(findings)
	return evidence, diagnostic.MergeCanonical(findings), nil
}

// RuleIDs returns the exact six-rule motion membership.
func RuleIDs() []string { return append([]string(nil), ruleIDs...) }

func platformMatches(kind diagnostic.EvidenceKind, platform string) bool {
	switch kind {
	case diagnostic.EvidenceWebSource, diagnostic.EvidenceWebRendered:
		return platform == "web"
	case diagnostic.EvidenceNativeSource:
		return platform == "react-native"
	case diagnostic.EvidenceDesignDocumentSource, diagnostic.EvidenceDesignDocumentComputed:
		return platform == "design-document"
	case diagnostic.EvidenceSimulator:
		return platform == "ios"
	case diagnostic.EvidenceEmulator:
		return platform == "android"
	case diagnostic.EvidencePhysicalDevice:
		return platform == "ios" || platform == "android"
	case diagnostic.EvidenceDefinition, diagnostic.EvidenceConsumerConformance,
		diagnostic.EvidenceConsumerContentRegistry, diagnostic.EvidenceExecution:
		return false
	}
	return false
}

func validateNode(node Node) error {
	if node.ID == "" || node.Owner == "" || len(node.Easing) != 4 {
		return fmt.Errorf("id, owner, and four easing coordinates are required")
	}
	if !slices.Contains([]string{"state-change", "feedback", "loading", "reveal", "decorative"}, node.Purpose) ||
		!slices.Contains([]string{"state-change", "feedback", "loading", "reveal", "status-pulse", "fake-caret", "marquee", "layout-transition", "image-hover-transform"}, node.MotionKind) ||
		!slices.Contains([]string{"standard", "bounce", "elastic", "overshoot"}, node.EasingBehavior) ||
		!slices.Contains([]string{"none", "functional", "decorative"}, node.HoverPurpose) {
		return fmt.Errorf("contains an unknown enum value")
	}
	if node.DurationMS < 0 {
		return fmt.Errorf("durationMs must be non-negative")
	}
	for _, property := range node.AnimatedProperties {
		if !slices.Contains([]string{"opacity", "transform", "width", "height", "padding", "margin", "scroll-offset"}, property) {
			return fmt.Errorf("animated property %q is invalid", property)
		}
	}
	return nil
}

func validateRegistry(node Node, preference string, registry map[string]Transition) error {
	if node.TransitionID == "" {
		return nil
	}
	transition, ok := registry[node.TransitionID]
	if !ok {
		return fmt.Errorf("transitionId %q is not in the consumer motion registry", node.TransitionID)
	}
	if transition.Owner == "" || transition.Purpose == "" || transition.Owner != node.Owner || transition.Purpose != node.Purpose {
		return fmt.Errorf("transitionId %q owner or purpose does not exactly match the consumer registry", node.TransitionID)
	}
	expectedDuration := transition.DurationMS
	if preference == "reduce" {
		expectedDuration = transition.ReducedDurationMS
		if transition.ReducedFallback == "" {
			return fmt.Errorf("transitionId %q has no reduced-motion fallback", node.TransitionID)
		}
	}
	if node.DurationMS != expectedDuration || !slices.Equal(node.Easing, transition.Easing) {
		return fmt.Errorf("transitionId %q duration or easing does not match the resolved consumer registry", node.TransitionID)
	}
	return nil
}

func hasLayoutProperty(properties []string) bool {
	for _, property := range properties {
		if property == "width" || property == "height" || property == "padding" || property == "margin" {
			return true
		}
	}
	return false
}

func allowedDecorativeException(node Node) bool {
	switch node.MotionKind {
	case "status-pulse":
		return node.ChangingDataSource && strings.TrimSpace(node.AccessibleLabel) != ""
	case "fake-caret":
		return node.EditableInput
	case "marquee":
		return !node.Perpetual || node.UserControlled
	case "image-hover-transform":
		return node.HoverPurpose == "functional"
	default:
		return false
	}
}

func add(findings []diagnostic.Diagnostic, condition bool, ruleID, message, path string, node Node, evidence Evidence, config Config) []diagnostic.Diagnostic {
	if !condition || !config.Active(ruleID) {
		return findings
	}
	return append(findings, diagnostic.NewWithSources(ruleID, []string{strings.TrimPrefix(ruleID, "motion/")}, config.Severity(ruleID), message, path, nil, evidence.EvidenceKind, evidence.Platform, node.Owner, "motion"))
}
