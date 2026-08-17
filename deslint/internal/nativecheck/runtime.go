package nativecheck

import (
	"fmt"
	"slices"
	"sort"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// RuntimeEvidence is one simulator, emulator, or physical-device capture set.
type RuntimeEvidence struct {
	Schema              string                  `json:"$schema,omitempty"`
	SchemaVersion       int                     `json:"schemaVersion"`
	EvidenceKind        diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform            string                  `json:"platform"`
	SurfaceID           string                  `json:"surfaceId"`
	NativePolicyVersion string                  `json:"nativePolicyVersion"`
	Captures            []RuntimeCapture        `json:"captures"`
}

// RuntimeCapture keeps environment axes and each runtime concern independently observable.
type RuntimeCapture struct {
	ID            string               `json:"id"`
	Owner         string               `json:"owner"`
	Environment   RuntimeEnvironment   `json:"environment"`
	SystemGesture SystemGesture        `json:"systemGesture"`
	SafeArea      SafeArea             `json:"safeArea"`
	KeyboardIME   KeyboardIME          `json:"keyboardIme"`
	Accessibility RuntimeAccessibility `json:"accessibility"`
	FontScaling   FontScaling          `json:"fontScaling"`
	Motion        RuntimeMotion        `json:"motion"`
	Appearance    RuntimeAppearance    `json:"appearance"`
	Adaptivity    RuntimeAdaptivity    `json:"adaptivity"`
	Performance   RuntimePerformance   `json:"performance"`
	Bundle        RuntimeBundle        `json:"bundle"`
}

// RuntimeEnvironment is the exact device/window/theme/accessibility axis tuple.
type RuntimeEnvironment struct {
	FormFactor   string  `json:"formFactor"`
	Orientation  string  `json:"orientation"`
	WindowMode   string  `json:"windowMode"`
	FoldPosture  string  `json:"foldPosture"`
	Theme        string  `json:"theme"`
	FontScale    float64 `json:"fontScale"`
	RefreshRate  int     `json:"refreshRateHz"`
	ReduceMotion bool    `json:"reduceMotion"`
}

// SystemGesture records platform back gesture reachability and conflicts.
type SystemGesture struct {
	BackReachable bool `json:"backReachable"`
	ConflictFree  bool `json:"conflictFree"`
}

// SafeArea records safe-area and edge-to-edge inset handling.
type SafeArea struct {
	Respected               bool `json:"respected"`
	EdgeToEdgeInsetsApplied bool `json:"edgeToEdgeInsetsApplied"`
}

// KeyboardIME records an explicit keyboard/IME visibility test.
type KeyboardIME struct {
	Tested                 bool `json:"tested"`
	InputVisible           bool `json:"inputVisible"`
	PrimaryActionReachable bool `json:"primaryActionReachable"`
}

// RuntimeAccessibility records traversal and resolved interactive semantics.
type RuntimeAccessibility struct {
	ReadingOrder []string               `json:"readingOrder"`
	FocusOrder   []string               `json:"focusOrder"`
	Controls     []AccessibilityControl `json:"controls"`
}

// AccessibilityControl records resolved semantics and target geometry.
type AccessibilityControl struct {
	ID                     string  `json:"id"`
	Owner                  string  `json:"owner"`
	Interactive            bool    `json:"interactive"`
	Label                  string  `json:"label"`
	RoleOrTrait            string  `json:"roleOrTrait"`
	StateAnnounced         bool    `json:"stateAnnounced"`
	Width                  float64 `json:"width"`
	Height                 float64 `json:"height"`
	NearestAdjacentSpacing float64 `json:"nearestAdjacentSpacing"`
}

// FontScaling records scaled layout failures and resolved text units/roles.
type FontScaling struct {
	Clipped            bool          `json:"clipped"`
	Overlap            bool          `json:"overlap"`
	UnreachableControl bool          `json:"unreachableControl"`
	Text               []RuntimeText `json:"text"`
}

// RuntimeText records one resolved platform text role.
type RuntimeText struct {
	ID               string  `json:"id"`
	Owner            string  `json:"owner"`
	Size             float64 `json:"size"`
	UsesScaledUnits  bool    `json:"usesScaledUnits"`
	PlatformTypeRole bool    `json:"platformTypeRole"`
}

// RuntimeMotion records the resolved Reduce Motion outcome.
type RuntimeMotion struct {
	FallbackApplied bool `json:"fallbackApplied"`
}

// RuntimeAppearance records contrast, dark theme completeness, and quick-invert use.
type RuntimeAppearance struct {
	ContrastPass      bool `json:"contrastPass"`
	DarkThemeComplete bool `json:"darkThemeComplete"`
	QuickInvertUsed   bool `json:"quickInvertUsed"`
}

// RuntimeAdaptivity records applied window-class and posture behavior.
type RuntimeAdaptivity struct {
	SizeOrWindowClassApplied bool `json:"sizeOrWindowClassApplied"`
	NotSimplyScaled          bool `json:"notSimplyScaled"`
	WindowModeSupported      bool `json:"windowModeSupported"`
	FoldPostureSupported     bool `json:"foldPostureSupported"`
}

// RuntimePerformance records startup and frame-path measurements.
type RuntimePerformance struct {
	SynchronousStartupMS float64     `json:"synchronousStartupMs"`
	InitializationMS     float64     `json:"initializationMs"`
	Frames               []FramePath `json:"frames"`
}

// FramePath records one scroll or gesture frame sample.
type FramePath struct {
	ID                  string  `json:"id"`
	Owner               string  `json:"owner"`
	TotalFrames         int     `json:"totalFrames"`
	DroppedFrames       int     `json:"droppedFrames"`
	MaxMainThreadWorkMS float64 `json:"maxMainThreadWorkMs"`
}

// RuntimeBundle records current/baseline JS and binary weights.
type RuntimeBundle struct {
	JSBytes                int `json:"jsBytes"`
	JSBaselineBytes        int `json:"jsBaselineBytes"`
	AppBinaryBytes         int `json:"appBinaryBytes"`
	AppBinaryBaselineBytes int `json:"appBinaryBaselineBytes"`
	UnusedDependencyBytes  int `json:"unusedDependencyBytes"`
}

// AnalyzeRuntime strictly parses and evaluates one native runtime evidence payload.
func AnalyzeRuntime(path string, contents []byte, config Config) (RuntimeEvidence, []diagnostic.Diagnostic, error) {
	var evidence RuntimeEvidence
	if err := strictDecode(contents, "native runtime evidence", &evidence); err != nil {
		return RuntimeEvidence{}, nil, err
	}
	config = normalizeConfig(config)
	if evidence.SchemaVersion != 1 || evidence.SurfaceID == "" || len(evidence.Captures) == 0 || !runtimeKindMatches(evidence.EvidenceKind, evidence.Platform) {
		return RuntimeEvidence{}, nil, fmt.Errorf("native runtime evidence identity, kind, or platform is invalid")
	}
	if config.RegistryVersion == "" || evidence.NativePolicyVersion != config.RegistryVersion {
		return RuntimeEvidence{}, nil, fmt.Errorf("native runtime policy version %q does not match consumer policy %q", evidence.NativePolicyVersion, config.RegistryVersion)
	}
	captures := append([]RuntimeCapture(nil), evidence.Captures...)
	sort.SliceStable(captures, func(i, j int) bool { return captures[i].ID < captures[j].ID })
	if err := uniqueOwnedIDs("runtime capture", captures, func(value RuntimeCapture) (string, string) { return value.ID, value.Owner }); err != nil {
		return RuntimeEvidence{}, nil, err
	}

	findings := []diagnostic.Diagnostic{}
	for _, capture := range captures {
		if err := validateCapture(capture, evidence.Platform); err != nil {
			return RuntimeEvidence{}, nil, fmt.Errorf("native runtime capture %q: %w", capture.ID, err)
		}
		capturePath := path + "#/captures/" + capture.ID
		environmentInvalid := !capture.SystemGesture.BackReachable || !capture.SystemGesture.ConflictFree || !capture.SafeArea.Respected || !capture.SafeArea.EdgeToEdgeInsetsApplied || !capture.KeyboardIME.Tested || !capture.KeyboardIME.InputVisible || !capture.KeyboardIME.PrimaryActionReachable
		findings = add(findings, environmentInvalid, rules.RuleNativeEnvironmentContract, "native-environment-contract", "system gesture, safe-area/edge-to-edge, or keyboard/IME outcome failed", capturePath, capture.Owner, evidence.EvidenceKind, evidence.Platform, config)

		ordersDiffer := !slices.Equal(capture.Accessibility.ReadingOrder, capture.Accessibility.FocusOrder)
		findings = add(findings, ordersDiffer, rules.RuleNativeAccessibilityContract, "native-accessibility-contract", "reading order and focus order differ", capturePath+"/accessibility", capture.Owner, evidence.EvidenceKind, evidence.Platform, config)
		controls := sorted(capture.Accessibility.Controls, func(value AccessibilityControl) string { return value.ID })
		for _, control := range controls {
			semanticsInvalid := control.Interactive && (control.Label == "" || control.RoleOrTrait == "" || !control.StateAnnounced)
			findings = add(findings, semanticsInvalid, rules.RuleNativeAccessibilityContract, "native-accessibility-contract", "interactive control lacks a resolved label, role/trait, or state announcement", capturePath+"/accessibility/controls/"+control.ID, control.Owner, evidence.EvidenceKind, evidence.Platform, config)
			targetFloor, spacingFloor := 44.0, config.IOSAdjacentTargetSpacing
			if evidence.Platform == "android" {
				targetFloor, spacingFloor = 48, 8
			}
			targetInvalid := control.Interactive && (control.Width < targetFloor || control.Height < targetFloor || control.NearestAdjacentSpacing < spacingFloor)
			findings = add(findings, targetInvalid, rules.RuleNativeTouchTarget, "native-touch-target", "interactive target is below the platform size or adjacent-spacing floor", capturePath+"/accessibility/controls/"+control.ID, control.Owner, evidence.EvidenceKind, evidence.Platform, config)
		}

		fontInvalid := capture.FontScaling.Clipped || capture.FontScaling.Overlap || capture.FontScaling.UnreachableControl
		findings = add(findings, fontInvalid, rules.RuleNativeFontScalingLayout, "native-font-scaling-layout", "scaled text clips, overlaps, or makes a control unreachable", capturePath+"/fontScaling", capture.Owner, evidence.EvidenceKind, evidence.Platform, config)
		for _, text := range sorted(capture.FontScaling.Text, func(value RuntimeText) string { return value.ID }) {
			textInvalid := !text.UsesScaledUnits || !text.PlatformTypeRole || evidence.Platform == "ios" && text.Size < 11
			findings = add(findings, textInvalid, rules.RuleNativeFontScalingLayout, "native-font-scaling-layout", "text bypasses scaled units/platform type role or the iOS 11pt floor", capturePath+"/fontScaling/text/"+text.ID, text.Owner, evidence.EvidenceKind, evidence.Platform, config)
		}

		motionAppearanceInvalid := capture.Environment.ReduceMotion && !capture.Motion.FallbackApplied || !capture.Appearance.ContrastPass || capture.Environment.Theme == "dark" && !capture.Appearance.DarkThemeComplete || capture.Appearance.QuickInvertUsed
		findings = add(findings, motionAppearanceInvalid, rules.RuleNativeReducedMotionContrast, "native-reduced-motion-contrast", "Reduce Motion fallback, contrast, or dark appearance is incomplete", capturePath+"/appearance", capture.Owner, evidence.EvidenceKind, evidence.Platform, config)
		findings = add(findings, capture.Appearance.QuickInvertUsed || capture.Environment.Theme == "dark" && !capture.Appearance.DarkThemeComplete, rules.RuleNativeSemanticAppearance, "native-semantic-appearance", "runtime appearance uses quick invert or has a broken dark theme", capturePath+"/appearance", capture.Owner, evidence.EvidenceKind, evidence.Platform, config)

		adaptiveInvalid := !capture.Adaptivity.SizeOrWindowClassApplied || !capture.Adaptivity.NotSimplyScaled || !capture.Adaptivity.WindowModeSupported || capture.Environment.FormFactor == "foldable" && !capture.Adaptivity.FoldPostureSupported
		findings = add(findings, adaptiveInvalid, rules.RuleNativeAdaptiveLayout, "native-adaptive-layout", "runtime layout did not apply size/window class, window mode, or fold posture", capturePath+"/adaptivity", capture.Owner, evidence.EvidenceKind, evidence.Platform, config)
		findings = add(findings, !capture.SystemGesture.BackReachable || !capture.SystemGesture.ConflictFree, rules.RuleNativeNavigationContract, "native-navigation-contract", "system Back gesture is unreachable or conflicts with application gestures", capturePath+"/systemGesture", capture.Owner, evidence.EvidenceKind, evidence.Platform, config)

		startupInvalid := capture.Performance.SynchronousStartupMS > config.Thresholds.MaxSynchronousStartupMS || capture.Performance.InitializationMS > config.Thresholds.MaxInitializationMS
		findings = add(findings, startupInvalid, rules.RuleNativeStartupWork, "native-startup-work", "synchronous startup work or initialization exceeds consumer policy", capturePath+"/performance", capture.Owner, evidence.EvidenceKind, evidence.Platform, config)
		for _, frame := range sorted(capture.Performance.Frames, func(value FramePath) string { return value.ID }) {
			dropRatio := float64(frame.DroppedFrames) / float64(frame.TotalFrames)
			frameInvalid := dropRatio > config.Thresholds.MaxFrameDropRatio || frame.MaxMainThreadWorkMS > config.Thresholds.MaxMainThreadWorkMS
			findings = add(findings, frameInvalid, rules.RuleNativeFrameBudget, "native-frame-budget", "60/120Hz frame drop ratio or main-thread work exceeds consumer policy", capturePath+"/performance/frames/"+frame.ID, frame.Owner, evidence.EvidenceKind, evidence.Platform, config)
		}
		bundleInvalid := capture.Bundle.JSBytes-capture.Bundle.JSBaselineBytes > config.Thresholds.MaxJSBundleRegressionBytes || capture.Bundle.AppBinaryBytes-capture.Bundle.AppBinaryBaselineBytes > config.Thresholds.MaxAppBinaryRegressionBytes || capture.Bundle.UnusedDependencyBytes > 0
		findings = add(findings, bundleInvalid, rules.RuleNativeBundleWeight, "native-bundle-weight", "JS bundle/app binary regression or unused dependency bytes exceed consumer policy", capturePath+"/bundle", capture.Owner, evidence.EvidenceKind, evidence.Platform, config)
	}
	diagnostic.Sort(findings)
	return evidence, diagnostic.MergeCanonical(findings), nil
}

// CoverageFindings returns deterministic findings for exact policy capture requirements not satisfied by runtime evidence.
func CoverageFindings(path string, evidences []RuntimeEvidence, config Config) []diagnostic.Diagnostic {
	config = normalizeConfig(config)
	findings := []diagnostic.Diagnostic{}
	requirements := append([]RuntimeRequirement(nil), config.RequiredRuntimeCaptures...)
	sort.SliceStable(requirements, func(i, j int) bool { return requirements[i].ID < requirements[j].ID })
	for _, requirement := range requirements {
		matched := false
		for _, evidence := range evidences {
			if evidence.Platform != requirement.Platform || evidence.EvidenceKind != requirement.EvidenceKind {
				continue
			}
			for _, capture := range evidence.Captures {
				environment := capture.Environment
				if capture.ID == requirement.ID && environment.FormFactor == requirement.FormFactor && environment.Orientation == requirement.Orientation && environment.WindowMode == requirement.WindowMode && environment.FoldPosture == requirement.FoldPosture && environment.Theme == requirement.Theme && environment.FontScale >= requirement.MinimumFontScale && environment.ReduceMotion == requirement.ReduceMotion {
					matched = true
					break
				}
			}
		}
		findings = add(findings, !matched, rules.RuleNativeRuntimeMatrix, "native-runtime-matrix", "consumer-required native runtime capture is missing or its environment axes drifted", path+"#/"+requirement.ID, "ansldes/evidence", requirement.EvidenceKind, requirement.Platform, config)
	}
	diagnostic.Sort(findings)
	return diagnostic.MergeCanonical(findings)
}

// RuleIDs returns the exact native-conformance rule membership.
func RuleIDs() []string {
	return []string{
		rules.RuleNativeEnvironmentContract, rules.RuleNativeAccessibilityContract, rules.RuleNativeFontScalingLayout,
		rules.RuleNativeTouchTarget, rules.RuleNativeReducedMotionContrast, rules.RuleNativeStartupWork,
		rules.RuleNativeListVirtualization, rules.RuleNativeFrameBudget, rules.RuleNativeRenderStability,
		rules.RuleNativeImageEfficiency, rules.RuleNativeBundleWeight, rules.RuleNativeSemanticAppearance,
		rules.RuleNativeAdaptiveLayout, rules.RuleNativeNavigationContract, rules.RuleNativeRuntimeMatrix,
	}
}

func runtimeKindMatches(kind diagnostic.EvidenceKind, platform string) bool {
	return kind == diagnostic.EvidenceSimulator && platform == "ios" || kind == diagnostic.EvidenceEmulator && platform == "android" || kind == diagnostic.EvidencePhysicalDevice && (platform == "ios" || platform == "android")
}

func validateCapture(capture RuntimeCapture, platform string) error {
	environment := capture.Environment
	if !slices.Contains([]string{"phone", "tablet", "foldable"}, environment.FormFactor) || !slices.Contains([]string{"portrait", "landscape"}, environment.Orientation) || !slices.Contains([]string{"fullscreen", "split", "multi-window"}, environment.WindowMode) || !slices.Contains([]string{"not-applicable", "flat", "half-open"}, environment.FoldPosture) || !slices.Contains([]string{"light", "dark"}, environment.Theme) || environment.FontScale < 1 || !slices.Contains([]int{60, 120}, environment.RefreshRate) {
		return fmt.Errorf("environment axes are invalid")
	}
	if platform == "ios" && environment.FormFactor == "foldable" || environment.FormFactor != "foldable" && environment.FoldPosture != "not-applicable" {
		return fmt.Errorf("fold posture is incompatible with form factor or platform")
	}
	if err := uniqueOwnedIDs("accessibility control", capture.Accessibility.Controls, func(value AccessibilityControl) (string, string) { return value.ID, value.Owner }); err != nil {
		return err
	}
	if err := uniqueOwnedIDs("runtime text", capture.FontScaling.Text, func(value RuntimeText) (string, string) { return value.ID, value.Owner }); err != nil {
		return err
	}
	if err := uniqueOwnedIDs("frame path", capture.Performance.Frames, func(value FramePath) (string, string) { return value.ID, value.Owner }); err != nil {
		return err
	}
	controlIDs := make([]string, 0, len(capture.Accessibility.Controls))
	for _, control := range capture.Accessibility.Controls {
		if control.Width < 0 || control.Height < 0 || control.NearestAdjacentSpacing < 0 {
			return fmt.Errorf("control %q has invalid geometry", control.ID)
		}
		controlIDs = append(controlIDs, control.ID)
	}
	slices.Sort(controlIDs)
	reading := append([]string(nil), capture.Accessibility.ReadingOrder...)
	focus := append([]string(nil), capture.Accessibility.FocusOrder...)
	slices.Sort(reading)
	slices.Sort(focus)
	if !slices.Equal(controlIDs, reading) || !slices.Equal(controlIDs, focus) {
		return fmt.Errorf("reading/focus order must contain the exact control identity set")
	}
	for _, frame := range capture.Performance.Frames {
		if frame.TotalFrames <= 0 || frame.DroppedFrames < 0 || frame.DroppedFrames > frame.TotalFrames || frame.MaxMainThreadWorkMS < 0 {
			return fmt.Errorf("frame path %q has invalid measurements", frame.ID)
		}
	}
	return nil
}
