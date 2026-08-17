// Package rules contains stable rule IDs and platform-neutral evaluation.
package rules

const (
	// RuleDefinitionSchemaVersion identifies unsupported definition schema versions.
	RuleDefinitionSchemaVersion = "definition/schema-version"
	// RuleDefinitionInvalidRef identifies invalid definition references.
	RuleDefinitionInvalidRef = "definition/invalid-reference"
	// RuleDefinitionUnknownToken identifies definition references to unknown tokens.
	RuleDefinitionUnknownToken = "definition/unknown-token"
	// RuleSourceSyntaxError identifies source syntax errors.
	RuleSourceSyntaxError = "source/syntax-error"
	// RuleSourceRawValue identifies disallowed raw values in source files.
	RuleSourceRawValue = "source/raw-value"
	// RulePencilRawValue identifies disallowed raw values in Pencil documents.
	RulePencilRawValue = "pencil/raw-value"
	// RuleLayoutProblem identifies computed-layout violations.
	RuleLayoutProblem = "layout/problem"
	// RuleEvidenceMissing identifies required evidence that was not provided.
	RuleEvidenceMissing = "evidence/missing"
	// RuleEvidenceStale identifies evidence that no longer matches its source.
	RuleEvidenceStale = "evidence/stale"
	// RulePolicyDefinitionMismatch identifies incompatible policy and definition versions.
	RulePolicyDefinitionMismatch = "policy/definition-mismatch"
	// RulePolicyBudgetExceeded identifies diagnostic counts above a configured budget.
	RulePolicyBudgetExceeded = "policy/budget-exceeded"
	// RulePolicyExceptionExpired identifies expired policy exceptions.
	RulePolicyExceptionExpired = "policy/exception-expired"
	// RuleProfileExaggeratedButton identifies excessive oversized action prominence.
	RuleProfileExaggeratedButton = "profile/exaggerated-button"
	// RuleProfileMismatchedControl identifies controls outside their design-system contract.
	RuleProfileMismatchedControl = "profile/mismatched-form-control"
	// RuleProfileGratuitousMotion identifies decorative interaction motion.
	RuleProfileGratuitousMotion = "profile/gratuitous-motion"
	// RuleProfileInventedAffordance identifies affordances with no approved source.
	RuleProfileInventedAffordance = "profile/invented-affordance"
	// RuleProfileInconsistentAction identifies inconsistent representations of one action.
	RuleProfileInconsistentAction = "profile/inconsistent-action"
	// RuleProfileMissingState identifies incomplete interactive state contracts.
	RuleProfileMissingState = "profile/missing-state"
	// RuleDesignSystemFont maps the upstream design-system-font rule.
	RuleDesignSystemFont = "design-system/font"
	// RuleDesignSystemColor maps the upstream design-system-color rule.
	RuleDesignSystemColor = "design-system/color"
	// RuleDesignSystemRadius maps the upstream design-system-radius rule.
	RuleDesignSystemRadius = "design-system/radius"
	// RuleDesignSystemFontSize maps the upstream design-system-font-size rule.
	RuleDesignSystemFontSize = "design-system/font-size"
	// RuleVisualSideTab maps the upstream side-tab rule.
	RuleVisualSideTab = "visual/side-tab"
	// RuleVisualBorderAccentRounded maps the upstream border-accent-on-rounded rule.
	RuleVisualBorderAccentRounded = "visual/border-accent-on-rounded"
	// RuleVisualThinBorderWideShadow maps the upstream gpt-thin-border-wide-shadow rule.
	RuleVisualThinBorderWideShadow = "visual/gpt-thin-border-wide-shadow"
	// RuleVisualRepeatingStripes maps the upstream repeating-stripes-gradient rule.
	RuleVisualRepeatingStripes = "visual/repeating-stripes-gradient"
	// RuleVisualGridBackground maps the upstream codex-grid-background rule.
	RuleVisualGridBackground = "visual/codex-grid-background"
	// RuleNativeListRowAccessoryWrapper identifies an unnecessary accessory tile wrapper.
	RuleNativeListRowAccessoryWrapper = "native/list-row-accessory-wrapper"
	// RuleTypographyOverusedFont maps the upstream overused-font rule.
	RuleTypographyOverusedFont = "typography/overused-font"
	// RuleTypographyFlatTypeHierarchy maps the upstream flat-type-hierarchy rule.
	RuleTypographyFlatTypeHierarchy = "typography/flat-type-hierarchy"
	// RuleTypographyIconTileStack maps the upstream icon-tile-stack rule.
	RuleTypographyIconTileStack = "typography/icon-tile-stack"
	// RuleTypographyItalicSerifDisplay maps the upstream italic-serif-display rule.
	RuleTypographyItalicSerifDisplay = "typography/italic-serif-display"
	// RuleTypographyHeroEyebrowChip maps the upstream hero-eyebrow-chip rule.
	RuleTypographyHeroEyebrowChip = "typography/hero-eyebrow-chip"
	// RuleTypographyKickerAboveHeading maps the upstream kicker-above-heading rule.
	RuleTypographyKickerAboveHeading = "typography/kicker-above-heading"
	// RuleTypographyOversizedH1 maps the upstream oversized-h1 rule.
	RuleTypographyOversizedH1 = "typography/oversized-h1"
	// RuleTypographyExtremeNegativeTracking maps the upstream extreme-negative-tracking rule.
	RuleTypographyExtremeNegativeTracking = "typography/extreme-negative-tracking"
	// RuleTypographyTightLeading maps the upstream tight-leading rule.
	RuleTypographyTightLeading = "typography/tight-leading"
	// RuleTypographyTinyText maps the upstream tiny-text rule.
	RuleTypographyTinyText = "typography/tiny-text"
	// RuleTypographyUndersizedUIText maps the upstream undersized-ui-text rule.
	RuleTypographyUndersizedUIText = "typography/undersized-ui-text"
	// RuleTypographyAllCapsBody maps the upstream all-caps-body rule.
	RuleTypographyAllCapsBody = "typography/all-caps-body"
	// RuleTypographyWideTracking maps the upstream wide-tracking rule.
	RuleTypographyWideTracking = "typography/wide-tracking"
	// RuleTypographySkippedHeading maps the upstream skipped-heading rule.
	RuleTypographySkippedHeading = "typography/skipped-heading"
	// RuleColorGradientText maps the upstream gradient-text rule.
	RuleColorGradientText = "color/gradient-text"
	// RuleColorAiColorPalette maps the upstream ai-color-palette rule.
	RuleColorAiColorPalette = "color/ai-color-palette"
	// RuleColorCreamPalette maps the upstream cream-palette rule.
	RuleColorCreamPalette = "color/cream-palette"
	// RuleColorDarkGlow maps the upstream dark-glow rule.
	RuleColorDarkGlow = "color/dark-glow"
	// RuleColorRadialHalo maps the upstream radial-halo rule.
	RuleColorRadialHalo = "color/radial-halo"
	// RuleColorRadialSpotlightGlow maps the upstream radial-spotlight-glow rule.
	RuleColorRadialSpotlightGlow = "color/radial-spotlight-glow"
	// RuleColorGrayOnColor maps the upstream gray-on-color rule.
	RuleColorGrayOnColor = "color/gray-on-color"
	// RuleColorLowContrast maps the upstream low-contrast rule.
	RuleColorLowContrast = "color/low-contrast"
	// RuleColorPureExtremeSurface maps the upstream pure-extreme-surface rule.
	RuleColorPureExtremeSurface = "color/pure-extreme-surface"
	// RuleLayoutNestedCards maps the upstream nested-cards rule.
	RuleLayoutNestedCards = "layout/nested-cards"
	// RuleLayoutMonotonousSpacing maps the upstream monotonous-spacing rule.
	RuleLayoutMonotonousSpacing = "layout/monotonous-spacing"
	// RuleLayoutNumberedSectionLabels maps the upstream numbered-section-labels rule.
	RuleLayoutNumberedSectionLabels = "layout/numbered-section-labels"
	// RuleLayoutEdgeFlushCards maps the upstream edge-flush-cards rule.
	RuleLayoutEdgeFlushCards = "layout/edge-flush-cards"
	// RuleLayoutTextOcclusion maps the upstream text-occlusion rule.
	RuleLayoutTextOcclusion = "layout/text-occlusion"
	// RuleLayoutFirstViewportColumnOverflow maps the upstream first-viewport-column-overflow rule.
	RuleLayoutFirstViewportColumnOverflow = "layout/first-viewport-column-overflow"
	// RuleLayoutHeadingRhythm maps the upstream heading-rhythm rule.
	RuleLayoutHeadingRhythm = "layout/heading-rhythm"
	// RuleLayoutLineLength maps the upstream line-length rule.
	RuleLayoutLineLength = "layout/line-length"
	// RuleLayoutCrampedPadding maps the upstream cramped-padding rule.
	RuleLayoutCrampedPadding = "layout/cramped-padding"
	// RuleLayoutBodyTextViewportEdge maps the upstream body-text-viewport-edge rule.
	RuleLayoutBodyTextViewportEdge = "layout/body-text-viewport-edge"
	// RuleLayoutTextOverflow maps the upstream text-overflow rule.
	RuleLayoutTextOverflow = "layout/text-overflow"
	// RuleLayoutClippedOverflowContainer maps the upstream clipped-overflow-container rule.
	RuleLayoutClippedOverflowContainer = "layout/clipped-overflow-container"
	// RuleLayoutEqualIconFeatureColumns maps the Hallmark equal-icon-feature-columns rule.
	RuleLayoutEqualIconFeatureColumns = "layout/equal-icon-feature-columns"
	// RuleLayoutFullViewportCenteredHero maps the Hallmark full-viewport-centered-hero rule.
	RuleLayoutFullViewportCenteredHero = "layout/full-viewport-centered-hero"
	// RuleMotionBounceEasing maps the upstream bounce-easing rule.
	RuleMotionBounceEasing = "motion/bounce-easing"
	// RuleMotionPulsingDot maps the upstream pulsing-dot rule.
	RuleMotionPulsingDot = "motion/pulsing-dot"
	// RuleMotionBlinkingCursor maps the upstream blinking-cursor rule.
	RuleMotionBlinkingCursor = "motion/blinking-cursor"
	// RuleMotionMarquee maps the upstream marquee rule.
	RuleMotionMarquee = "motion/marquee"
	// RuleMotionLayoutTransition maps the upstream layout-transition rule.
	RuleMotionLayoutTransition = "motion/layout-transition"
	// RuleMotionImageHoverTransform maps the upstream image-hover-transform rule.
	RuleMotionImageHoverTransform = "motion/image-hover-transform"
	// RuleCopyEmDashOveruse maps the advisory upstream em-dash-overuse rule.
	RuleCopyEmDashOveruse = "copy/em-dash-overuse"
	// RuleCopyMarketingBuzzword maps the upstream marketing-buzzword rule.
	RuleCopyMarketingBuzzword = "copy/marketing-buzzword"
	// RuleCopyAphoristicCadence maps the upstream aphoristic-cadence rule.
	RuleCopyAphoristicCadence = "copy/aphoristic-cadence"
	// RuleCopyRepeatedContainerText maps the upstream repeated-container-text rule.
	RuleCopyRepeatedContainerText = "copy/repeated-container-text"
	// RuleCopyTheaterSlopPhrase maps the upstream theater-slop-phrase rule.
	RuleCopyTheaterSlopPhrase = "copy/theater-slop-phrase"
	// RuleCopyUnverifiedSocialProof maps the Hallmark unverified-social-proof rule.
	RuleCopyUnverifiedSocialProof = "copy/unverified-social-proof"
	// RuleImageryShapeAssembledIllustration maps the upstream shape-assembled-illustration rule.
	RuleImageryShapeAssembledIllustration = "imagery/shape-assembled-illustration"
	// RuleImageryBrokenImage maps the upstream broken-image rule.
	RuleImageryBrokenImage = "imagery/broken-image"
	// RuleRuntimeScriptError maps browser and native runtime failure signals.
	RuleRuntimeScriptError = "runtime/script-error"
	// RuleRuntimeContentHiddenAtRest maps the upstream content-hidden-at-rest rule.
	RuleRuntimeContentHiddenAtRest = "runtime/content-hidden-at-rest"
	// RuleRuntimeJustifiedText maps the upstream justified-text rule.
	RuleRuntimeJustifiedText = "runtime/justified-text"
	// RuleNativeEnvironmentContract checks system gesture, safe-area, and keyboard/IME outcomes.
	RuleNativeEnvironmentContract = "native/environment-contract"
	// RuleNativeAccessibilityContract checks labels, roles, announcements, and traversal order.
	RuleNativeAccessibilityContract = "native/accessibility-contract"
	// RuleNativeFontScalingLayout checks scaled text layout and platform type roles.
	RuleNativeFontScalingLayout = "native/font-scaling-layout"
	// RuleNativeTouchTarget checks platform target size and adjacent spacing.
	RuleNativeTouchTarget = "native/touch-target"
	// RuleNativeReducedMotionContrast checks reduced motion and appearance contrast together.
	RuleNativeReducedMotionContrast = "native/reduced-motion-contrast"
	// RuleNativeStartupWork checks first-frame synchronous work and initialization latency.
	RuleNativeStartupWork = "native/startup-work"
	// RuleNativeListVirtualization checks long-list virtualization and stable keys.
	RuleNativeListVirtualization = "native/list-virtualization"
	// RuleNativeFrameBudget checks scroll and gesture frame/main-thread budgets.
	RuleNativeFrameBudget = "native/frame-budget"
	// RuleNativeRenderStability checks rerenders, callbacks, keys, and memoization.
	RuleNativeRenderStability = "native/render-stability"
	// RuleNativeImageEfficiency checks thumbnail decode size, caching, and repeated loads.
	RuleNativeImageEfficiency = "native/image-efficiency"
	// RuleNativeBundleWeight checks unused dependencies and JS/app size regression.
	RuleNativeBundleWeight = "native/bundle-weight"
	// RuleNativeSemanticAppearance checks semantic platform roles and dark appearance.
	RuleNativeSemanticAppearance = "native/semantic-appearance"
	// RuleNativeAdaptiveLayout checks window-class, orientation, multi-window, and fold behavior.
	RuleNativeAdaptiveLayout = "native/adaptive-layout"
	// RuleNativeNavigationContract checks iOS and Android system navigation contracts.
	RuleNativeNavigationContract = "native/navigation-contract"
	// RuleNativeRuntimeMatrix checks the exact consumer-required runtime capture matrix.
	RuleNativeRuntimeMatrix = "native/runtime-matrix"
)

// ConfigurableRuleIDs is the exact v1 severity registry.
var ConfigurableRuleIDs = []string{
	RuleDefinitionSchemaVersion,
	RuleDefinitionInvalidRef,
	RuleDefinitionUnknownToken,
	RuleSourceSyntaxError,
	RuleSourceRawValue,
	RulePencilRawValue,
	RuleLayoutProblem,
	RuleEvidenceMissing,
	RuleEvidenceStale,
}
