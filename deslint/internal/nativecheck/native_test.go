package nativecheck

import (
	"os"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

func TestNegativeSourceMapsEverySourceOwnedNativeRule(t *testing.T) {
	_, findings, err := AnalyzeSource("source.json", read(t, "../../../packages/schema/testdata/native-source-negative.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	for _, ruleID := range []string{
		rules.RuleNativeAccessibilityContract, rules.RuleNativeListVirtualization, rules.RuleNativeRenderStability,
		rules.RuleNativeImageEfficiency, rules.RuleNativeBundleWeight, rules.RuleNativeSemanticAppearance,
		rules.RuleNativeAdaptiveLayout, rules.RuleNativeNavigationContract,
		rules.RuleRuntimeJustifiedText, rules.RuleRuntimeContentHiddenAtRest, rules.RuleMotionLayoutTransition,
	} {
		if !hasRule(findings, ruleID) {
			t.Fatalf("missing %s in %#v", ruleID, findings)
		}
	}
}

func TestNegativeRuntimeKeepsIOSAndAndroidTargetRulesPlatformSpecific(t *testing.T) {
	_, iosFindings, err := AnalyzeRuntime("ios.json", read(t, "../../../packages/schema/testdata/native-runtime-negative-ios.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	for _, ruleID := range []string{
		rules.RuleNativeEnvironmentContract, rules.RuleNativeAccessibilityContract, rules.RuleNativeFontScalingLayout,
		rules.RuleNativeTouchTarget, rules.RuleNativeReducedMotionContrast, rules.RuleNativeStartupWork,
		rules.RuleNativeFrameBudget, rules.RuleNativeBundleWeight, rules.RuleNativeSemanticAppearance,
		rules.RuleNativeAdaptiveLayout, rules.RuleNativeNavigationContract,
	} {
		if !hasRule(iosFindings, ruleID) {
			t.Fatalf("missing %s in %#v", ruleID, iosFindings)
		}
	}
	_, androidFindings, err := AnalyzeRuntime("android.json", read(t, "../../../packages/schema/testdata/native-runtime-negative-android.json"), config())
	if err != nil || !hasRule(androidFindings, rules.RuleNativeTouchTarget) || !hasRule(androidFindings, rules.RuleNativeFontScalingLayout) {
		t.Fatalf("android findings/error = %#v %v", androidFindings, err)
	}
}

func TestPositiveSourceAndFullRuntimeMatrixAreClean(t *testing.T) {
	_, sourceFindings, err := AnalyzeSource("source.json", read(t, "../../../packages/schema/testdata/native-source-positive.json"), config())
	if err != nil || len(sourceFindings) != 0 {
		t.Fatalf("source = %#v %v", sourceFindings, err)
	}
	ios, iosFindings, err := AnalyzeRuntime("ios.json", read(t, "../../../packages/schema/testdata/native-runtime-positive-ios.json"), config())
	if err != nil || len(iosFindings) != 0 {
		t.Fatalf("ios = %#v %v", iosFindings, err)
	}
	android, androidFindings, err := AnalyzeRuntime("android.json", read(t, "../../../packages/schema/testdata/native-runtime-positive-android.json"), config())
	if err != nil || len(androidFindings) != 0 {
		t.Fatalf("android = %#v %v", androidFindings, err)
	}
	coverage := CoverageFindings("<native-runtime>", []RuntimeEvidence{ios, android}, config())
	if len(coverage) != 0 {
		t.Fatalf("coverage = %#v", coverage)
	}
}

func TestMissingRuntimeAxisIsFindingNotPass(t *testing.T) {
	ios, _, err := AnalyzeRuntime("ios.json", read(t, "../../../packages/schema/testdata/native-runtime-positive-ios.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	coverage := CoverageFindings("<native-runtime>", []RuntimeEvidence{ios}, config())
	if !hasRule(coverage, rules.RuleNativeRuntimeMatrix) {
		t.Fatalf("coverage = %#v", coverage)
	}
	for _, finding := range coverage {
		if finding.Platform != "android" || finding.EvidenceKind != diagnostic.EvidenceEmulator {
			t.Fatalf("coverage finding = %#v", finding)
		}
	}
}

func config() Config {
	return Config{
		RegistryVersion:          "1.0.0",
		IOSAdjacentTargetSpacing: 8,
		Thresholds: Thresholds{
			MaxSynchronousStartupMS: 20, MaxInitializationMS: 800, MaxMainThreadWorkMS: 8,
			MaxFrameDropRatio: 0.02, MaxThumbnailDecodeRatio: 4,
			MaxJSBundleRegressionBytes: 1024, MaxAppBinaryRegressionBytes: 2048,
		},
		RequiredRuntimeCaptures: []RuntimeRequirement{
			{ID: "ios-phone-light", Platform: "ios", EvidenceKind: diagnostic.EvidenceSimulator, FormFactor: "phone", Orientation: "portrait", WindowMode: "fullscreen", FoldPosture: "not-applicable", Theme: "light", MinimumFontScale: 1, ReduceMotion: false},
			{ID: "ios-tablet-dark-large", Platform: "ios", EvidenceKind: diagnostic.EvidenceSimulator, FormFactor: "tablet", Orientation: "landscape", WindowMode: "split", FoldPosture: "not-applicable", Theme: "dark", MinimumFontScale: 2.35, ReduceMotion: true},
			{ID: "android-phone-light", Platform: "android", EvidenceKind: diagnostic.EvidenceEmulator, FormFactor: "phone", Orientation: "portrait", WindowMode: "fullscreen", FoldPosture: "not-applicable", Theme: "light", MinimumFontScale: 1, ReduceMotion: false},
			{ID: "android-tablet-dark-large", Platform: "android", EvidenceKind: diagnostic.EvidenceEmulator, FormFactor: "tablet", Orientation: "landscape", WindowMode: "multi-window", FoldPosture: "not-applicable", Theme: "dark", MinimumFontScale: 2.35, ReduceMotion: true},
			{ID: "android-foldable-half-open", Platform: "android", EvidenceKind: diagnostic.EvidenceEmulator, FormFactor: "foldable", Orientation: "portrait", WindowMode: "multi-window", FoldPosture: "half-open", Theme: "dark", MinimumFontScale: 1.6, ReduceMotion: true},
		},
	}
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	// #nosec G304 -- callers provide repository-owned fixture paths.
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func hasRule(findings []diagnostic.Diagnostic, ruleID string) bool {
	for _, finding := range findings {
		if finding.RuleID == ruleID {
			return true
		}
	}
	return false
}
