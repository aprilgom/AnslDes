package nativecheck

import (
	"fmt"
	"slices"
	"sort"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// SourceEvidence is one normalized React Native source inventory.
type SourceEvidence struct {
	Schema              string             `json:"$schema,omitempty"`
	SchemaVersion       int                `json:"schemaVersion"`
	EvidenceKind        string             `json:"evidenceKind"`
	Platform            string             `json:"platform"`
	SurfaceID           string             `json:"surfaceId"`
	NativePolicyVersion string             `json:"nativePolicyVersion"`
	Controls            []SourceControl    `json:"controls"`
	Lists               []SourceList       `json:"lists"`
	RenderPaths         []RenderPath       `json:"renderPaths"`
	Images              []SourceImage      `json:"images"`
	Dependencies        []SourceDependency `json:"dependencies"`
	Appearance          []SourceAppearance `json:"appearance"`
	Patterns            []SourcePattern    `json:"patterns"`
	Adaptivity          SourceAdaptivity   `json:"adaptivity"`
	Navigation          SourceNavigation   `json:"navigation"`
}

// SourceControl is one statically resolved interactive accessibility contract.
type SourceControl struct {
	ID                string `json:"id"`
	Owner             string `json:"owner"`
	Interactive       bool   `json:"interactive"`
	Label             string `json:"label"`
	RoleOrTrait       string `json:"roleOrTrait"`
	StateAnnouncement bool   `json:"stateAnnouncement"`
	ReadingOrder      int    `json:"readingOrder"`
	FocusOrder        int    `json:"focusOrder"`
}

// SourceList records long-list implementation semantics.
type SourceList struct {
	ID          string `json:"id"`
	Owner       string `json:"owner"`
	LongList    bool   `json:"longList"`
	Virtualized bool   `json:"virtualized"`
	StableKey   bool   `json:"stableKey"`
}

// RenderPath records statically proven render-stability behavior.
type RenderPath struct {
	ID                  string `json:"id"`
	Owner               string `json:"owner"`
	UnnecessaryRerender bool   `json:"unnecessaryRerender"`
	UnstableCallback    bool   `json:"unstableCallback"`
	UnstableKey         bool   `json:"unstableKey"`
	MemoizationRequired bool   `json:"memoizationRequired"`
	Memoized            bool   `json:"memoized"`
}

// SourceImage records decoded and displayed image cost.
type SourceImage struct {
	ID            string `json:"id"`
	Owner         string `json:"owner"`
	Thumbnail     bool   `json:"thumbnail"`
	DisplayPixels int    `json:"displayPixels"`
	DecodedPixels int    `json:"decodedPixels"`
	Cached        bool   `json:"cached"`
	RepeatedLoad  bool   `json:"repeatedLoad"`
}

// SourceDependency records one dependency contribution and reachability result.
type SourceDependency struct {
	ID                string `json:"id"`
	Owner             string `json:"owner"`
	Used              bool   `json:"used"`
	ContributionBytes int    `json:"contributionBytes"`
}

// SourceAppearance records raw color and platform material usage.
type SourceAppearance struct {
	ID                 string `json:"id"`
	Owner              string `json:"owner"`
	RawColor           bool   `json:"rawColor"`
	SemanticSystemRole bool   `json:"semanticSystemRole"`
	HandRolledMaterial bool   `json:"handRolledMaterial"`
}

// SourcePattern records statically resolved text, reveal, and animation behavior.
type SourcePattern struct {
	ID                    string  `json:"id"`
	Owner                 string  `json:"owner"`
	TextAlignment         string  `json:"textAlignment"`
	HiddenAtRest          bool    `json:"hiddenAtRest"`
	RevealFallbackVisible bool    `json:"revealFallbackVisible"`
	RawAnimation          bool    `json:"rawAnimation"`
	TransitionID          *string `json:"transitionId"`
}

// SourceAdaptivity records declared window-class and navigation adaptations.
type SourceAdaptivity struct {
	Owner                      string `json:"owner"`
	UsesSizeOrWindowClass      bool   `json:"usesSizeOrWindowClass"`
	SimplePhoneScaleOnExpanded bool   `json:"simplePhoneScaleOnExpanded"`
	SupportsPortrait           bool   `json:"supportsPortrait"`
	SupportsLandscape          bool   `json:"supportsLandscape"`
	SupportsSplitView          bool   `json:"supportsSplitView"`
	SupportsMultiWindow        bool   `json:"supportsMultiWindow"`
	SupportsFoldPosture        bool   `json:"supportsFoldPosture"`
	CompactNavigation          string `json:"compactNavigation"`
	ExpandedNavigation         string `json:"expandedNavigation"`
}

// SourceNavigation records platform back and navigation container contracts.
type SourceNavigation struct {
	Owner                     string `json:"owner"`
	IOSNavigationStackOrSheet bool   `json:"iosNavigationStackOrSheet"`
	IOSEdgeSwipeBack          bool   `json:"iosEdgeSwipeBack"`
	AndroidPredictiveBack     bool   `json:"androidPredictiveBack"`
}

// AnalyzeSource strictly parses and evaluates one React Native source evidence payload.
func AnalyzeSource(path string, contents []byte, config Config) (SourceEvidence, []diagnostic.Diagnostic, error) {
	var evidence SourceEvidence
	if err := strictDecode(contents, "native source evidence", &evidence); err != nil {
		return SourceEvidence{}, nil, err
	}
	config = normalizeConfig(config)
	if evidence.SchemaVersion != 1 || evidence.EvidenceKind != string(diagnostic.EvidenceNativeSource) || evidence.Platform != "react-native" || evidence.SurfaceID == "" {
		return SourceEvidence{}, nil, fmt.Errorf("native source evidence identity is invalid")
	}
	if config.RegistryVersion == "" || evidence.NativePolicyVersion != config.RegistryVersion {
		return SourceEvidence{}, nil, fmt.Errorf("native source policy version %q does not match consumer policy %q", evidence.NativePolicyVersion, config.RegistryVersion)
	}
	if err := validateSourceIdentities(evidence); err != nil {
		return SourceEvidence{}, nil, err
	}

	findings := []diagnostic.Diagnostic{}
	controls := append([]SourceControl(nil), evidence.Controls...)
	sort.SliceStable(controls, func(i, j int) bool { return controls[i].ID < controls[j].ID })
	for _, control := range controls {
		invalid := control.Interactive && (control.Label == "" || control.RoleOrTrait == "" || !control.StateAnnouncement || control.ReadingOrder != control.FocusOrder)
		findings = add(findings, invalid, rules.RuleNativeAccessibilityContract, "native-accessibility-contract", "interactive control lacks an exact label, role/trait, state announcement, or traversal order", path+"#/controls/"+control.ID, control.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	}
	for _, list := range sorted(evidence.Lists, func(value SourceList) string { return value.ID }) {
		findings = add(findings, list.LongList && (!list.Virtualized || !list.StableKey), rules.RuleNativeListVirtualization, "native-list-virtualization", "long list is not virtualized with a stable key", path+"#/lists/"+list.ID, list.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	}
	for _, renderPath := range sorted(evidence.RenderPaths, func(value RenderPath) string { return value.ID }) {
		invalid := renderPath.UnnecessaryRerender || renderPath.UnstableCallback || renderPath.UnstableKey || (renderPath.MemoizationRequired && !renderPath.Memoized)
		findings = add(findings, invalid, rules.RuleNativeRenderStability, "native-render-stability", "render path has unnecessary rerenders, unstable identity, or missing required memoization", path+"#/renderPaths/"+renderPath.ID, renderPath.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	}
	for _, image := range sorted(evidence.Images, func(value SourceImage) string { return value.ID }) {
		ratio := float64(image.DecodedPixels) / float64(image.DisplayPixels)
		invalid := image.Thumbnail && ratio > config.Thresholds.MaxThumbnailDecodeRatio || image.RepeatedLoad && !image.Cached
		findings = add(findings, invalid, rules.RuleNativeImageEfficiency, "native-image-efficiency", "thumbnail decode ratio or repeated uncached load exceeds consumer policy", path+"#/images/"+image.ID, image.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	}
	for _, dependency := range sorted(evidence.Dependencies, func(value SourceDependency) string { return value.ID }) {
		findings = add(findings, !dependency.Used && dependency.ContributionBytes > 0, rules.RuleNativeBundleWeight, "native-bundle-weight", "unused dependency contributes bytes to the native artifact", path+"#/dependencies/"+dependency.ID, dependency.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	}
	for _, appearance := range sorted(evidence.Appearance, func(value SourceAppearance) string { return value.ID }) {
		invalid := appearance.RawColor || !appearance.SemanticSystemRole || appearance.HandRolledMaterial
		findings = add(findings, invalid, rules.RuleNativeSemanticAppearance, "native-semantic-appearance", "appearance bypasses semantic system/material roles or uses hand-rolled material", path+"#/appearance/"+appearance.ID, appearance.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	}
	for _, pattern := range sorted(evidence.Patterns, func(value SourcePattern) string { return value.ID }) {
		patternPath := path + "#/patterns/" + pattern.ID
		findings = add(findings, pattern.TextAlignment == "justify", rules.RuleRuntimeJustifiedText, "justified-text", "React Native body content uses justified alignment", patternPath, pattern.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
		findings = add(findings, pattern.HiddenAtRest && !pattern.RevealFallbackVisible, rules.RuleRuntimeContentHiddenAtRest, "content-hidden-at-rest", "React Native content is hidden at rest without a source-visible fallback", patternPath, pattern.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
		findings = add(findings, pattern.RawAnimation, rules.RuleMotionLayoutTransition, "layout-transition", "React Native source uses a raw animation outside the transition registry", patternPath, pattern.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	}
	adaptive := evidence.Adaptivity
	adaptiveInvalid := adaptive.Owner == "" || !adaptive.UsesSizeOrWindowClass || adaptive.SimplePhoneScaleOnExpanded || !adaptive.SupportsPortrait || !adaptive.SupportsLandscape || !adaptive.SupportsSplitView || !adaptive.SupportsMultiWindow || !adaptive.SupportsFoldPosture || adaptive.CompactNavigation == "none" || adaptive.ExpandedNavigation == "none"
	findings = add(findings, adaptiveInvalid, rules.RuleNativeAdaptiveLayout, "native-adaptive-layout", "source contract does not cover window classes, orientation, split/multi-window, fold posture, and navigation adaptation", path+"#/adaptivity", adaptive.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	navigation := evidence.Navigation
	navigationInvalid := navigation.Owner == "" || !navigation.IOSNavigationStackOrSheet || !navigation.IOSEdgeSwipeBack || !navigation.AndroidPredictiveBack
	findings = add(findings, navigationInvalid, rules.RuleNativeNavigationContract, "native-navigation-contract", "source contract does not preserve iOS stack/edge-swipe and Android predictive Back", path+"#/navigation", navigation.Owner, diagnostic.EvidenceNativeSource, evidence.Platform, config)
	diagnostic.Sort(findings)
	return evidence, diagnostic.MergeCanonical(findings), nil
}

func validateSourceIdentities(evidence SourceEvidence) error {
	checks := []error{
		uniqueOwnedIDs("control", evidence.Controls, func(value SourceControl) (string, string) { return value.ID, value.Owner }),
		uniqueOwnedIDs("list", evidence.Lists, func(value SourceList) (string, string) { return value.ID, value.Owner }),
		uniqueOwnedIDs("render path", evidence.RenderPaths, func(value RenderPath) (string, string) { return value.ID, value.Owner }),
		uniqueOwnedIDs("image", evidence.Images, func(value SourceImage) (string, string) { return value.ID, value.Owner }),
		uniqueOwnedIDs("dependency", evidence.Dependencies, func(value SourceDependency) (string, string) { return value.ID, value.Owner }),
		uniqueOwnedIDs("appearance", evidence.Appearance, func(value SourceAppearance) (string, string) { return value.ID, value.Owner }),
		uniqueOwnedIDs("pattern", evidence.Patterns, func(value SourcePattern) (string, string) { return value.ID, value.Owner }),
	}
	for _, err := range checks {
		if err != nil {
			return err
		}
	}
	for _, control := range evidence.Controls {
		if control.ReadingOrder < 0 || control.FocusOrder < 0 {
			return fmt.Errorf("native control %q has invalid traversal order", control.ID)
		}
	}
	for _, image := range evidence.Images {
		if image.DisplayPixels <= 0 || image.DecodedPixels <= 0 {
			return fmt.Errorf("native image %q has invalid pixel dimensions", image.ID)
		}
	}
	for _, pattern := range evidence.Patterns {
		if !slices.Contains([]string{"start", "center", "end", "justify"}, pattern.TextAlignment) || pattern.RawAnimation && pattern.TransitionID != nil {
			return fmt.Errorf("native source pattern %q has invalid alignment or raw-animation registry state", pattern.ID)
		}
	}
	if !slices.Contains([]string{"stack", "tabs", "bar", "none"}, evidence.Adaptivity.CompactNavigation) || !slices.Contains([]string{"sidebar", "rail", "drawer", "none"}, evidence.Adaptivity.ExpandedNavigation) {
		return fmt.Errorf("native adaptivity navigation enum is invalid")
	}
	return nil
}

func sorted[T any](values []T, identity func(T) string) []T {
	result := append([]T(nil), values...)
	sort.SliceStable(result, func(i, j int) bool { return identity(result[i]) < identity(result[j]) })
	return result
}
