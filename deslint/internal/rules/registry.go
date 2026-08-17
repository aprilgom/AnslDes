package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"sort"
)

const (
	// CorePackID is the stable identity of the built-in product-neutral rule pack.
	CorePackID = "ansldes-core"
	// CorePackVersion changes when the exact built-in membership or behavior changes.
	CorePackVersion = "1.12.0"
	// AntiSlopPackID owns the exact 59 Impeccable plus four Hallmark canonical rules.
	AntiSlopPackID = "ansldes-anti-slop"
	// AntiSlopPackVersion changes only with catalog membership or behavior.
	AntiSlopPackVersion = "1.1.0"
)

// RuleSpec is the single registration unit for deterministic rule metadata.
type RuleSpec struct {
	ID                    string
	ImplementationVersion string
	Category              string
	Scopes                []string
	EvidenceKinds         []string
	Platforms             []string
	DefaultSeverity       string
	Provenance            []string
	Dependencies          []string
	Providers             []string
	RequiredInputs        []string
	Applicability         []RuleApplicability
}

// RuleApplicability records one exact provider-family mapping for a canonical rule.
type RuleApplicability struct {
	Target              string
	Status              string
	EvidenceKinds       []string
	Reason              string
	AlternativeEvidence []string
	Dependencies        []string
	SupplementRuleIDs   []string
}

// RulePackSpec is one versioned manifest in the engine registry.
type RulePackSpec struct {
	ID      string
	Version string
	Rules   []RuleSpec
}

var allRuleSpecs = []RuleSpec{
	{ID: RuleDefinitionSchemaVersion, EvidenceKinds: []string{"definition"}, Platforms: []string{"all"}},
	{ID: RuleDefinitionInvalidRef, EvidenceKinds: []string{"definition"}, Platforms: []string{"all"}},
	{ID: RuleDefinitionUnknownToken, EvidenceKinds: []string{"definition"}, Platforms: []string{"all"}},
	{ID: RuleSourceSyntaxError, EvidenceKinds: []string{"native-source", "web-source"}, Platforms: []string{"react-native", "web"}},
	{ID: RuleSourceRawValue, EvidenceKinds: []string{"native-source", "web-source"}, Platforms: []string{"react-native", "web"}},
	{ID: RulePencilRawValue, EvidenceKinds: []string{"design-document-source"}, Platforms: []string{"design-document"}},
	{ID: RuleLayoutProblem, EvidenceKinds: []string{"design-document-computed"}, Platforms: []string{"design-document"}},
	{ID: RuleEvidenceMissing, EvidenceKinds: []string{"execution"}, Platforms: []string{"all"}},
	{ID: RuleEvidenceStale, EvidenceKinds: []string{"execution"}, Platforms: []string{"all"}},
	{ID: RulePolicyDefinitionMismatch, EvidenceKinds: []string{"execution"}, Platforms: []string{"all"}},
	{ID: RulePolicyBudgetExceeded, EvidenceKinds: []string{"execution"}, Platforms: []string{"all"}},
	{ID: RulePolicyExceptionExpired, EvidenceKinds: []string{"execution"}, Platforms: []string{"all"}},
	{ID: RuleProfileExaggeratedButton, EvidenceKinds: []string{"consumer-conformance"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}},
	{ID: RuleProfileMismatchedControl, EvidenceKinds: []string{"consumer-conformance"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}},
	{ID: RuleProfileGratuitousMotion, EvidenceKinds: []string{"consumer-conformance"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}},
	{ID: RuleProfileInventedAffordance, EvidenceKinds: []string{"consumer-conformance"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}},
	{ID: RuleProfileInconsistentAction, EvidenceKinds: []string{"consumer-conformance"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}},
	{ID: RuleProfileMissingState, EvidenceKinds: []string{"consumer-conformance"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}},
	{ID: RuleDesignSystemFont, EvidenceKinds: []string{"native-source", "web-source"}, Platforms: []string{"react-native", "web"}, RequiredInputs: []string{"design-context"}},
	{ID: RuleDesignSystemColor, EvidenceKinds: []string{"native-source", "web-source"}, Platforms: []string{"react-native", "web"}, RequiredInputs: []string{"design-context"}},
	{ID: RuleDesignSystemRadius, EvidenceKinds: []string{"native-source", "web-source"}, Platforms: []string{"react-native", "web"}, RequiredInputs: []string{"design-context"}},
	{ID: RuleDesignSystemFontSize, EvidenceKinds: []string{"native-source", "web-source"}, Platforms: []string{"react-native", "web"}, RequiredInputs: []string{"design-context"}},
	{ID: RuleVisualSideTab, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"visual-detail"}},
	{ID: RuleVisualBorderAccentRounded, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"visual-detail"}},
	{ID: RuleVisualThinBorderWideShadow, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"visual-detail"}},
	{ID: RuleVisualRepeatingStripes, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"visual-detail"}},
	{ID: RuleVisualGridBackground, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"visual-detail"}},
	{ID: RuleNativeListRowAccessoryWrapper, EvidenceKinds: []string{"native-source", "design-document-source"}, Platforms: []string{"react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"visual-detail"}},
	{ID: RuleTypographyOverusedFont, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyFlatTypeHierarchy, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyIconTileStack, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyItalicSerifDisplay, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyHeroEyebrowChip, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyKickerAboveHeading, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyOversizedH1, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyExtremeNegativeTracking, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyTightLeading, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyTinyText, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyUndersizedUIText, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyAllCapsBody, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographyWideTracking, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleTypographySkippedHeading, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android", "design-document"}, RequiredInputs: []string{"typography"}},
	{ID: RuleColorGradientText, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleColorAiColorPalette, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleColorCreamPalette, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleColorDarkGlow, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleColorRadialHalo, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleColorRadialSpotlightGlow, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleColorGrayOnColor, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleColorLowContrast, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleColorPureExtremeSurface, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"color"}},
	{ID: RuleLayoutNestedCards, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutMonotonousSpacing, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutNumberedSectionLabels, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutEdgeFlushCards, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutTextOcclusion, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutFirstViewportColumnOverflow, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutHeadingRhythm, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutLineLength, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutCrampedPadding, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutBodyTextViewportEdge, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutTextOverflow, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutClippedOverflowContainer, EvidenceKinds: []string{"web-rendered", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutEqualIconFeatureColumns, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleLayoutFullViewportCenteredHero, EvidenceKinds: []string{"web-rendered", "native-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"layout-detail"}},
	{ID: RuleMotionBounceEasing, EvidenceKinds: []string{"web-source", "web-rendered", "native-source", "design-document-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"motion"}},
	{ID: RuleMotionPulsingDot, EvidenceKinds: []string{"web-source", "web-rendered", "native-source", "design-document-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"motion"}},
	{ID: RuleMotionBlinkingCursor, EvidenceKinds: []string{"web-source", "web-rendered", "native-source", "design-document-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"motion"}},
	{ID: RuleMotionMarquee, EvidenceKinds: []string{"web-source", "web-rendered", "native-source", "design-document-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"motion"}},
	{ID: RuleMotionLayoutTransition, EvidenceKinds: []string{"web-source", "web-rendered", "native-source", "design-document-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"motion"}},
	{ID: RuleMotionImageHoverTransform, EvidenceKinds: []string{"web-source", "web-rendered", "native-source", "design-document-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"motion"}},
	{ID: RuleCopyEmDashOveruse, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "design-document"}, RequiredInputs: []string{"copy"}},
	{ID: RuleCopyMarketingBuzzword, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "design-document"}, RequiredInputs: []string{"copy"}},
	{ID: RuleCopyAphoristicCadence, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "design-document"}, RequiredInputs: []string{"copy"}},
	{ID: RuleCopyRepeatedContainerText, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "design-document"}, RequiredInputs: []string{"copy"}},
	{ID: RuleCopyTheaterSlopPhrase, EvidenceKinds: []string{"web-source", "native-source", "design-document-source"}, Platforms: []string{"web", "react-native", "design-document"}, RequiredInputs: []string{"copy"}},
	{ID: RuleCopyUnverifiedSocialProof, EvidenceKinds: []string{"consumer-content-registry"}, Platforms: []string{"all"}, RequiredInputs: []string{"copy"}},
	{ID: RuleImageryShapeAssembledIllustration, EvidenceKinds: []string{"web-source", "web-rendered", "native-source", "design-document-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"imagery"}},
	{ID: RuleImageryBrokenImage, EvidenceKinds: []string{"web-source", "web-rendered", "native-source", "design-document-source", "design-document-computed", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "design-document", "ios", "android"}, RequiredInputs: []string{"imagery"}},
	{ID: RuleRuntimeScriptError, EvidenceKinds: []string{"web-rendered", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "ios", "android"}, RequiredInputs: []string{"runtime"}},
	{ID: RuleRuntimeContentHiddenAtRest, EvidenceKinds: []string{"web-rendered", "native-source", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android"}},
	{ID: RuleRuntimeJustifiedText, EvidenceKinds: []string{"web-rendered", "native-source", "simulator", "emulator", "physical-device"}, Platforms: []string{"web", "react-native", "ios", "android"}},
	{ID: RuleNativeEnvironmentContract, EvidenceKinds: []string{"simulator", "emulator", "physical-device"}, Platforms: []string{"ios", "android"}, RequiredInputs: []string{"native-runtime"}},
	{ID: RuleNativeAccessibilityContract, EvidenceKinds: []string{"native-source", "simulator", "emulator", "physical-device"}, Platforms: []string{"react-native", "ios", "android"}, RequiredInputs: []string{"native-conformance"}},
	{ID: RuleNativeFontScalingLayout, EvidenceKinds: []string{"simulator", "emulator", "physical-device"}, Platforms: []string{"ios", "android"}, RequiredInputs: []string{"native-runtime"}},
	{ID: RuleNativeTouchTarget, EvidenceKinds: []string{"simulator", "emulator", "physical-device"}, Platforms: []string{"ios", "android"}, RequiredInputs: []string{"native-runtime"}},
	{ID: RuleNativeReducedMotionContrast, EvidenceKinds: []string{"simulator", "emulator", "physical-device"}, Platforms: []string{"ios", "android"}, RequiredInputs: []string{"native-runtime"}},
	{ID: RuleNativeStartupWork, EvidenceKinds: []string{"simulator", "emulator", "physical-device"}, Platforms: []string{"ios", "android"}, RequiredInputs: []string{"native-runtime"}},
	{ID: RuleNativeListVirtualization, EvidenceKinds: []string{"native-source"}, Platforms: []string{"react-native"}, RequiredInputs: []string{"native-source-conformance"}},
	{ID: RuleNativeFrameBudget, EvidenceKinds: []string{"simulator", "emulator", "physical-device"}, Platforms: []string{"ios", "android"}, RequiredInputs: []string{"native-runtime"}},
	{ID: RuleNativeRenderStability, EvidenceKinds: []string{"native-source"}, Platforms: []string{"react-native"}, RequiredInputs: []string{"native-source-conformance"}},
	{ID: RuleNativeImageEfficiency, EvidenceKinds: []string{"native-source"}, Platforms: []string{"react-native"}, RequiredInputs: []string{"native-source-conformance"}},
	{ID: RuleNativeBundleWeight, EvidenceKinds: []string{"native-source", "simulator", "emulator", "physical-device"}, Platforms: []string{"react-native", "ios", "android"}, RequiredInputs: []string{"native-conformance"}},
	{ID: RuleNativeSemanticAppearance, EvidenceKinds: []string{"native-source", "simulator", "emulator", "physical-device"}, Platforms: []string{"react-native", "ios", "android"}, RequiredInputs: []string{"native-conformance"}},
	{ID: RuleNativeAdaptiveLayout, EvidenceKinds: []string{"native-source", "simulator", "emulator", "physical-device"}, Platforms: []string{"react-native", "ios", "android"}, RequiredInputs: []string{"native-conformance"}},
	{ID: RuleNativeNavigationContract, EvidenceKinds: []string{"native-source", "simulator", "emulator", "physical-device"}, Platforms: []string{"react-native", "ios", "android"}, RequiredInputs: []string{"native-conformance"}},
	{ID: RuleNativeRuntimeMatrix, EvidenceKinds: []string{"simulator", "emulator", "physical-device"}, Platforms: []string{"ios", "android"}, RequiredInputs: []string{"native-runtime"}},
}

var coreRuleSpecs, antiSlopRuleSpecs = partitionRuleSpecs()

var registeredPacks = []RulePackSpec{
	{ID: AntiSlopPackID, Version: AntiSlopPackVersion, Rules: antiSlopRuleSpecs},
	{ID: CorePackID, Version: CorePackVersion, Rules: coreRuleSpecs},
}

// AllRuleIDs is the exact sorted union of registered rule IDs.
var AllRuleIDs = registeredRuleIDs()

// RegisteredPacks returns a defensive copy of the rule-pack registry.
func RegisteredPacks() []RulePackSpec {
	result := make([]RulePackSpec, len(registeredPacks))
	for index, pack := range registeredPacks {
		result[index] = RulePackSpec{ID: pack.ID, Version: pack.Version, Rules: cloneRuleSpecs(pack.Rules)}
	}
	return result
}

// LookupPack returns one exact id/version match.
func LookupPack(id, version string) (RulePackSpec, bool) {
	for _, pack := range registeredPacks {
		if pack.ID == id && pack.Version == version {
			return RulePackSpec{ID: pack.ID, Version: pack.Version, Rules: cloneRuleSpecs(pack.Rules)}, true
		}
	}
	return RulePackSpec{}, false
}

// Lookup returns one canonical rule registration.
func Lookup(ruleID string) (RuleSpec, bool) {
	for _, pack := range registeredPacks {
		for _, spec := range pack.Rules {
			if spec.ID == ruleID {
				return cloneRuleSpec(spec), true
			}
		}
	}
	return RuleSpec{}, false
}

// PackFingerprint returns the manifest fingerprint over exact sorted membership.
func PackFingerprint(pack RulePackSpec) string {
	members := make([]string, 0, len(pack.Rules))
	for _, spec := range pack.Rules {
		members = append(members, spec.ID)
	}
	sort.Strings(members)
	contents, _ := json.Marshal(struct {
		ID      string   `json:"id"`
		Version string   `json:"version"`
		Members []string `json:"members"`
	}{ID: pack.ID, Version: pack.Version, Members: members})
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}

// CorePackFingerprint returns the current built-in core manifest fingerprint.
func CorePackFingerprint() string {
	pack, _ := LookupPack(CorePackID, CorePackVersion)
	return PackFingerprint(pack)
}

// AntiSlopPackFingerprint returns the exact canonical 63-member manifest fingerprint.
func AntiSlopPackFingerprint() string {
	pack, _ := LookupPack(AntiSlopPackID, AntiSlopPackVersion)
	return PackFingerprint(pack)
}

func partitionRuleSpecs() ([]RuleSpec, []RuleSpec) {
	catalog := AntiSlopCatalog()
	catalogByID := make(map[string]AntiSlopCatalogRule, len(catalog))
	for _, entry := range catalog {
		catalogByID[entry.ID] = entry
	}
	core := make([]RuleSpec, 0, len(allRuleSpecs)-len(catalog))
	antiSlop := make([]RuleSpec, 0, len(catalog))
	seenCatalog := map[string]bool{}
	for _, spec := range allRuleSpecs {
		entry, found := catalogByID[spec.ID]
		if !found {
			spec.ImplementationVersion = "ansldes@1"
			spec.Category = "infrastructure"
			spec.DefaultSeverity = "error"
			spec.Provenance = []string{"ansldes/" + spec.ID}
			core = append(core, spec)
			continue
		}
		seenCatalog[spec.ID] = true
		spec.ImplementationVersion = entry.ImplementationVersion
		spec.Category = entry.Category
		spec.Scopes = append([]string(nil), entry.Scopes...)
		spec.DefaultSeverity = entry.DefaultSeverity
		spec.Provenance = append([]string(nil), entry.Provenance...)
		spec.Dependencies = append([]string(nil), entry.Dependencies...)
		spec.Providers = append([]string(nil), entry.Providers...)
		spec.Applicability = applicabilityFor(spec)
		antiSlop = append(antiSlop, spec)
	}
	if len(antiSlop) != len(catalog) || len(seenCatalog) != len(catalog) {
		panic("anti-slop catalog and implementation registry exact sets differ")
	}
	sort.SliceStable(core, func(i, j int) bool { return core[i].ID < core[j].ID })
	sort.SliceStable(antiSlop, func(i, j int) bool { return antiSlop[i].ID < antiSlop[j].ID })
	return core, antiSlop
}

func applicabilityFor(spec RuleSpec) []RuleApplicability {
	webKinds := matchingEvidence(spec.EvidenceKinds, []string{"web-source", "web-rendered"})
	if len(webKinds) == 0 && slices.Contains(spec.Platforms, "all") {
		webKinds = append([]string(nil), spec.EvidenceKinds...)
	}
	nativeKinds := matchingEvidence(spec.EvidenceKinds, []string{"native-source", "simulator", "emulator", "physical-device"})
	documentKinds := matchingEvidence(spec.EvidenceKinds, []string{"design-document-source", "design-document-computed"})
	result := []RuleApplicability{{
		Target: "web", Status: "supported", EvidenceKinds: webKinds,
		Reason:       "pinned Web providers implement the upstream detector intent",
		Dependencies: append([]string(nil), spec.RequiredInputs...),
	}}
	nativeStatus := "shared-intent"
	supplements := []string{}
	if spec.ID == RuleTypographyIconTileStack {
		nativeStatus = "native-supplement"
		supplements = []string{RuleNativeListRowAccessoryWrapper}
	}
	if len(nativeKinds) == 0 {
		result = append(result, RuleApplicability{
			Target: "react-native", Status: "unsupported", Reason: "no deterministic native provider can reproduce the required rendered Web semantics",
			AlternativeEvidence: webKinds,
		})
	} else {
		result = append(result, RuleApplicability{
			Target: "react-native", Status: nativeStatus, EvidenceKinds: nativeKinds,
			Reason:       "provider-neutral intent is evaluated from native source or runtime evidence",
			Dependencies: append([]string(nil), spec.RequiredInputs...), SupplementRuleIDs: supplements,
		})
	}
	documentStatus := "shared-intent"
	if slices.Contains(documentKinds, "design-document-computed") {
		documentStatus = "design-document-computed"
	}
	if len(documentKinds) == 0 {
		result = append(result, RuleApplicability{
			Target: "design-document", Status: "unsupported", Reason: "the design-document visitor cannot establish this runtime or content provenance fact",
			AlternativeEvidence: append(webKinds, nativeKinds...),
		})
	} else {
		result = append(result, RuleApplicability{
			Target: "design-document", Status: documentStatus, EvidenceKinds: documentKinds,
			Reason:       "the document source or computed-layout visitor exposes the required normalized evidence",
			Dependencies: append([]string(nil), spec.RequiredInputs...),
		})
	}
	return result
}

func matchingEvidence(actual, candidates []string) []string {
	result := []string{}
	for _, kind := range actual {
		if slices.Contains(candidates, kind) {
			result = append(result, kind)
		}
	}
	sort.Strings(result)
	return result
}

func registeredRuleIDs() []string {
	result := make([]string, 0)
	for _, pack := range registeredPacks {
		for _, spec := range pack.Rules {
			result = append(result, spec.ID)
		}
	}
	sort.Strings(result)
	return slices.Compact(result)
}

func cloneRuleSpecs(values []RuleSpec) []RuleSpec {
	result := make([]RuleSpec, 0, len(values))
	for _, value := range values {
		result = append(result, cloneRuleSpec(value))
	}
	return result
}

func cloneRuleSpec(value RuleSpec) RuleSpec {
	value.Scopes = append([]string(nil), value.Scopes...)
	value.EvidenceKinds = append([]string(nil), value.EvidenceKinds...)
	value.Platforms = append([]string(nil), value.Platforms...)
	value.Provenance = append([]string(nil), value.Provenance...)
	value.Dependencies = append([]string(nil), value.Dependencies...)
	value.Providers = append([]string(nil), value.Providers...)
	value.RequiredInputs = append([]string(nil), value.RequiredInputs...)
	applicability := value.Applicability
	value.Applicability = make([]RuleApplicability, len(applicability))
	for index, mapping := range applicability {
		mapping.EvidenceKinds = append([]string(nil), mapping.EvidenceKinds...)
		mapping.AlternativeEvidence = append([]string(nil), mapping.AlternativeEvidence...)
		mapping.Dependencies = append([]string(nil), mapping.Dependencies...)
		mapping.SupplementRuleIDs = append([]string(nil), mapping.SupplementRuleIDs...)
		value.Applicability[index] = mapping
	}
	return value
}
