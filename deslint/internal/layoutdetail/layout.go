// Package layoutdetail evaluates provider-neutral semantic spacing and computed bounds.
package layoutdetail

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

// Evidence is one immutable provider capture for a surface and viewport.
type Evidence struct {
	Schema             string                  `json:"$schema,omitempty"`
	SchemaVersion      int                     `json:"schemaVersion"`
	EvidenceKind       diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform           string                  `json:"platform"`
	ProfileID          string                  `json:"profileId"`
	SurfaceID          string                  `json:"surfaceId"`
	Viewport           Size                    `json:"viewport"`
	CapturePath        string                  `json:"capturePath"`
	ComputedBoundsPath string                  `json:"computedBoundsPath"`
	DocumentReport     *DocumentReport         `json:"documentReport,omitempty"`
	Nodes              []Node                  `json:"nodes"`
}

// DocumentReport preserves the exact design-document visitor execution metadata.
type DocumentReport struct {
	FingerprintSHA256 string         `json:"fingerprintSha256"`
	VisitorOptions    VisitorOptions `json:"visitorOptions"`
	NodeCount         int            `json:"nodeCount"`
	Issues            []VisitorIssue `json:"issues"`
}

// VisitorOptions records deterministic traversal and bounds options.
type VisitorOptions struct {
	ResolveInstances bool `json:"resolveInstances"`
	IncludeHidden    bool `json:"includeHidden"`
	ComputeBounds    bool `json:"computeBounds"`
}

// VisitorIssue is one stable computed overflow, overlap, or clipping observation.
type VisitorIssue struct {
	ID      string `json:"id"`
	NodeID  string `json:"nodeId"`
	Owner   string `json:"owner"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// Size is a positive viewport extent.
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// SpacingRelation records one semantic relationship and its resolved gap.
type SpacingRelation struct {
	Kind  string  `json:"kind"`
	Value float64 `json:"value"`
	Token string  `json:"token"`
}

// Padding records computed inner insets.
type Padding struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// Node is a normalized structural or computed-layout observation.
type Node struct {
	ID                         string            `json:"id"`
	Owner                      string            `json:"owner"`
	Kind                       string            `json:"kind"`
	ContentRole                string            `json:"contentRole"`
	ParentID                   string            `json:"parentId,omitempty"`
	CardDepth                  int               `json:"cardDepth,omitempty"`
	RepeatedCardSiblings       int               `json:"repeatedCardSiblings,omitempty"`
	AccessoryTile              bool              `json:"accessoryTile,omitempty"`
	SemanticBoundary           bool              `json:"semanticBoundary,omitempty"`
	SpacingRelations           []SpacingRelation `json:"spacingRelations,omitempty"`
	GuideSkeletonRepeated      bool              `json:"guideSkeletonRepeated,omitempty"`
	SectionLabelNumbered       bool              `json:"sectionLabelNumbered,omitempty"`
	SectionLabelMeaningful     bool              `json:"sectionLabelMeaningful,omitempty"`
	SectionLabelFontSize       float64           `json:"sectionLabelFontSize,omitempty"`
	HorizontalScroller         bool              `json:"horizontalScroller,omitempty"`
	LeadingGutter              float64           `json:"leadingGutter,omitempty"`
	TrailingGutter             float64           `json:"trailingGutter,omitempty"`
	OpaqueOverlapArea          float64           `json:"opaqueOverlapArea,omitempty"`
	OpeningViewport            bool              `json:"openingViewport,omitempty"`
	ColumnHeights              []float64         `json:"columnHeights,omitempty"`
	SpacingBefore              float64           `json:"spacingBefore,omitempty"`
	SpacingAfter               float64           `json:"spacingAfter,omitempty"`
	LineLengthChars            float64           `json:"lineLengthChars,omitempty"`
	BoundedText                bool              `json:"boundedText,omitempty"`
	Padding                    *Padding          `json:"padding,omitempty"`
	ViewportHorizontalInset    float64           `json:"viewportHorizontalInset,omitempty"`
	OverflowsContainer         bool              `json:"overflowsContainer,omitempty"`
	UnintendedHorizontalScroll bool              `json:"unintendedHorizontalScroll,omitempty"`
	PositionedOverlay          bool              `json:"positionedOverlay,omitempty"`
	ClippingAncestor           bool              `json:"clippingAncestor,omitempty"`
	FeatureColumnCount         int               `json:"featureColumnCount,omitempty"`
	EqualColumns               bool              `json:"equalColumns,omitempty"`
	IconAboveHeading           bool              `json:"iconAboveHeading,omitempty"`
	PageFeatureRegion          bool              `json:"pageFeatureRegion,omitempty"`
	MinHeightVH                float64           `json:"minHeightVh,omitempty"`
	CenterHorizontal           bool              `json:"centerHorizontal,omitempty"`
	CenterVertical             bool              `json:"centerVertical,omitempty"`
	BoundaryAlternatives       []string          `json:"boundaryAlternatives,omitempty"`
}

// Config supplies the selected consumer profile and governed rule controls.
type Config struct {
	ProfileID string
	Density   string
	Severity  func(string) diagnostic.Severity
	Active    func(string) bool
}

var ruleIDs = []string{
	rules.RuleLayoutNestedCards,
	rules.RuleLayoutMonotonousSpacing,
	rules.RuleLayoutNumberedSectionLabels,
	rules.RuleLayoutEdgeFlushCards,
	rules.RuleLayoutTextOcclusion,
	rules.RuleLayoutFirstViewportColumnOverflow,
	rules.RuleLayoutHeadingRhythm,
	rules.RuleLayoutLineLength,
	rules.RuleLayoutCrampedPadding,
	rules.RuleLayoutBodyTextViewportEdge,
	rules.RuleLayoutTextOverflow,
	rules.RuleLayoutClippedOverflowContainer,
	rules.RuleLayoutEqualIconFeatureColumns,
	rules.RuleLayoutFullViewportCenteredHero,
}

// Analyze strictly parses and evaluates one layout evidence payload.
func Analyze(path string, contents []byte, config Config) (Evidence, []diagnostic.Diagnostic, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, fmt.Errorf("layout evidence has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, nil, err
	}
	if evidence.SchemaVersion != 1 || evidence.ProfileID == "" || evidence.SurfaceID == "" || evidence.CapturePath == "" || evidence.ComputedBoundsPath == "" || evidence.Viewport.Width <= 0 || evidence.Viewport.Height <= 0 || len(evidence.Nodes) == 0 {
		return Evidence{}, nil, fmt.Errorf("layout evidence identity, viewport, capture, computed bounds, and nodes are required")
	}
	if !platformMatches(evidence.EvidenceKind, evidence.Platform) {
		return Evidence{}, nil, fmt.Errorf("layout evidence kind %q is incompatible with platform %q", evidence.EvidenceKind, evidence.Platform)
	}
	if evidence.EvidenceKind == diagnostic.EvidenceDesignDocumentComputed {
		report := evidence.DocumentReport
		if report == nil || len(report.FingerprintSHA256) != 64 || report.NodeCount != len(evidence.Nodes) || !report.VisitorOptions.ResolveInstances || !report.VisitorOptions.ComputeBounds {
			return Evidence{}, nil, fmt.Errorf("design-document evidence requires an exact visitor fingerprint, options, and node count")
		}
		seenIssues := map[string]bool{}
		for _, issue := range report.Issues {
			if issue.ID == "" || issue.NodeID == "" || issue.Owner == "" || issue.Message == "" || seenIssues[issue.ID] || !slices.Contains([]string{"overflow", "overlap", "clipping"}, issue.Kind) {
				return Evidence{}, nil, fmt.Errorf("design-document visitor issue %q is incomplete, duplicated, or unsupported", issue.ID)
			}
			seenIssues[issue.ID] = true
		}
	}
	if config.ProfileID != "" && evidence.ProfileID != config.ProfileID {
		return Evidence{}, nil, fmt.Errorf("layout evidence profile %q does not match policy profile %q", evidence.ProfileID, config.ProfileID)
	}
	if config.Severity == nil {
		config.Severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if config.Active == nil {
		config.Active = func(string) bool { return true }
	}
	paddingFloor := map[string]float64{"compact": 8, "comfortable": 12, "spacious": 16}[config.Density]
	if paddingFloor == 0 {
		paddingFloor = 12
	}

	nodes := append([]Node(nil), evidence.Nodes...)
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	seen := map[string]bool{}
	findings := []diagnostic.Diagnostic{}
	if evidence.DocumentReport != nil {
		issues := append([]VisitorIssue(nil), evidence.DocumentReport.Issues...)
		sort.SliceStable(issues, func(i, j int) bool { return issues[i].ID < issues[j].ID })
		for _, issue := range issues {
			if config.Active(rules.RuleLayoutProblem) {
				findings = append(findings, diagnostic.New(rules.RuleLayoutProblem, config.Severity(rules.RuleLayoutProblem), issue.Message, path+"#/documentReport/issues/"+issue.ID, nil, evidence.EvidenceKind, evidence.Platform, issue.Owner, issue.Kind))
			}
		}
	}
	for _, node := range nodes {
		if err := validateNode(node); err != nil {
			return Evidence{}, nil, fmt.Errorf("layout node %q: %w", node.ID, err)
		}
		if seen[node.ID] {
			return Evidence{}, nil, fmt.Errorf("layout evidence has duplicate node identity %q", node.ID)
		}
		seen[node.ID] = true
		nodePath := path + "#/nodes/" + node.ID
		cardLike := node.Kind == "card" || node.Kind == "accessory-tile"
		findings = add(findings, cardLike && node.CardDepth >= 2 && node.RepeatedCardSiblings >= 2 && !node.SemanticBoundary, rules.RuleLayoutNestedCards, "repeated nested card surface has no semantic boundary", nodePath, node, evidence, config, []string{"hallmark-eight-05"})
		findings = add(findings, monotonous(node), rules.RuleLayoutMonotonousSpacing, "distinct semantic spacing relationships resolve to one flat gap", nodePath, node, evidence, config, nil)
		findings = add(findings, node.SectionLabelNumbered && !node.SectionLabelMeaningful && node.SectionLabelFontSize > 0 && node.SectionLabelFontSize <= 12, rules.RuleLayoutNumberedSectionLabels, "tiny numbered section label has no semantic meaning", nodePath, node, evidence, config, nil)
		findings = add(findings, node.HorizontalScroller && math.Abs(node.LeadingGutter-node.TrailingGutter) >= 8, rules.RuleLayoutEdgeFlushCards, "horizontal scroller has asymmetric edge gutters", nodePath, node, evidence, config, nil)
		findings = add(findings, (node.Kind == "text" || node.Kind == "heading") && node.OpaqueOverlapArea > 0, rules.RuleLayoutTextOcclusion, "opaque layer overlaps text bounds", nodePath, node, evidence, config, nil)
		findings = add(findings, columnOverflow(node, evidence.Viewport.Height), rules.RuleLayoutFirstViewportColumnOverflow, "opening columns overflow the first viewport with a large height imbalance", nodePath, node, evidence, config, nil)
		findings = add(findings, node.Kind == "heading" && node.SpacingAfter > 0 && node.SpacingBefore <= node.SpacingAfter, rules.RuleLayoutHeadingRhythm, "heading spacing before is not greater than spacing after", nodePath, node, evidence, config, nil)
		findings = add(findings, node.ContentRole == "prose" && node.LineLengthChars > 80, rules.RuleLayoutLineLength, "prose line length exceeds the 80ch warning ceiling", nodePath, node, evidence, config, nil)
		findings = add(findings, node.BoundedText && node.Padding != nil && minimumPadding(*node.Padding) < paddingFloor, rules.RuleLayoutCrampedPadding, fmt.Sprintf("bounded text padding is below the %.0fpx %s profile floor", paddingFloor, densityName(config.Density)), nodePath, node, evidence, config, nil)
		findings = add(findings, node.Kind == "text" && node.ContentRole == "prose" && node.ViewportHorizontalInset > 0 && node.ViewportHorizontalInset < 16, rules.RuleLayoutBodyTextViewportEdge, "body text is too close to the viewport edge", nodePath, node, evidence, config, nil)
		findings = add(findings, (node.Kind == "text" || node.Kind == "heading") && (node.OverflowsContainer || node.UnintendedHorizontalScroll), rules.RuleLayoutTextOverflow, "text exceeds its container or creates horizontal scrolling", nodePath, node, evidence, config, nil)
		findings = add(findings, node.PositionedOverlay && node.ClippingAncestor, rules.RuleLayoutClippedOverflowContainer, "positioned overlay is trapped by a clipping ancestor", nodePath, node, evidence, config, nil)
		findings = add(findings, node.Kind == "feature-grid" && node.FeatureColumnCount == 3 && node.EqualColumns && node.IconAboveHeading && node.PageFeatureRegion, rules.RuleLayoutEqualIconFeatureColumns, "page feature region uses three identical icon-above-heading columns", nodePath, node, evidence, config, []string{"hallmark-eight-03"})
		findings = add(findings, node.Kind == "hero" && node.MinHeightVH >= 100 && node.CenterHorizontal && node.CenterVertical, rules.RuleLayoutFullViewportCenteredHero, "hero combines full viewport height with two-axis centering", nodePath, node, evidence, config, []string{"hallmark-eight-04"})
	}
	diagnostic.Sort(findings)
	return evidence, diagnostic.MergeCanonical(findings), nil
}

// RuleIDs returns the exact layout-detail rule membership.
func RuleIDs() []string { return append([]string(nil), ruleIDs...) }

func platformMatches(kind diagnostic.EvidenceKind, platform string) bool {
	switch kind {
	case diagnostic.EvidenceWebRendered:
		return platform == "web"
	case diagnostic.EvidenceNativeSource:
		return platform == "react-native"
	case diagnostic.EvidenceDesignDocumentComputed:
		return platform == "design-document"
	case diagnostic.EvidenceSimulator:
		return platform == "ios"
	case diagnostic.EvidenceEmulator:
		return platform == "android"
	case diagnostic.EvidencePhysicalDevice:
		return platform == "ios" || platform == "android"
	case diagnostic.EvidenceDefinition, diagnostic.EvidenceWebSource, diagnostic.EvidenceDesignDocumentSource,
		diagnostic.EvidenceConsumerConformance, diagnostic.EvidenceConsumerContentRegistry, diagnostic.EvidenceExecution:
		return false
	}
	return false
}

func validateNode(node Node) error {
	if node.ID == "" || node.Owner == "" {
		return fmt.Errorf("id and owner are required")
	}
	if !slices.Contains([]string{"page", "section", "card", "text", "heading", "scroller", "overlay", "feature-grid", "hero", "data-grid", "canvas", "accessory-tile", "guide"}, node.Kind) ||
		!slices.Contains([]string{"prose", "data", "table", "control", "decorative"}, node.ContentRole) {
		return fmt.Errorf("contains an unknown enum value")
	}
	for _, relation := range node.SpacingRelations {
		if relation.Token == "" || relation.Value < 0 || !slices.Contains([]string{"intra-component", "inter-component", "section", "heading-before", "heading-after"}, relation.Kind) {
			return fmt.Errorf("contains an invalid spacing relation")
		}
	}
	if node.CardDepth < 0 || node.RepeatedCardSiblings < 0 || node.SectionLabelFontSize < 0 || node.LeadingGutter < 0 || node.TrailingGutter < 0 || node.OpaqueOverlapArea < 0 || node.SpacingBefore < 0 || node.SpacingAfter < 0 || node.LineLengthChars < 0 || node.ViewportHorizontalInset < 0 || node.FeatureColumnCount < 0 || node.MinHeightVH < 0 {
		return fmt.Errorf("contains an out-of-range numeric value")
	}
	return nil
}

func monotonous(node Node) bool {
	if node.GuideSkeletonRepeated {
		return true
	}
	if node.ContentRole == "data" || node.ContentRole == "table" || len(node.SpacingRelations) < 3 {
		return false
	}
	kinds := map[string]bool{}
	values := map[float64]bool{}
	for _, relation := range node.SpacingRelations {
		kinds[relation.Kind] = true
		values[relation.Value] = true
	}
	return len(kinds) >= 3 && len(values) == 1
}

func columnOverflow(node Node, viewportHeight float64) bool {
	if !node.OpeningViewport || len(node.ColumnHeights) < 2 {
		return false
	}
	minimum, maximum := node.ColumnHeights[0], node.ColumnHeights[0]
	for _, height := range node.ColumnHeights[1:] {
		minimum = math.Min(minimum, height)
		maximum = math.Max(maximum, height)
	}
	return maximum > viewportHeight && maximum-minimum > viewportHeight*0.25
}

func minimumPadding(padding Padding) float64 {
	return math.Min(math.Min(padding.Top, padding.Right), math.Min(padding.Bottom, padding.Left))
}

func densityName(value string) string {
	if value == "compact" || value == "comfortable" || value == "spacious" {
		return value
	}
	return "comfortable"
}

func add(findings []diagnostic.Diagnostic, condition bool, ruleID, message, path string, node Node, evidence Evidence, config Config, extraSources []string) []diagnostic.Diagnostic {
	if !condition || !config.Active(ruleID) {
		return findings
	}
	sources := append([]string{strings.TrimPrefix(ruleID, "layout/")}, extraSources...)
	return append(findings, diagnostic.NewWithSources(ruleID, sources, config.Severity(ruleID), message, path, nil, evidence.EvidenceKind, evidence.Platform, node.Owner, "layout-detail"))
}
