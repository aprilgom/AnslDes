package lint_test

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/lint"
	"github.com/aprilgom/AnslDes/deslint/internal/policy"
	"github.com/aprilgom/AnslDes/deslint/internal/report"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
	"github.com/aprilgom/AnslDes/deslint/internal/source"
	"github.com/aprilgom/AnslDes/deslint/internal/webcheck"
)

func TestRunnerKeepsEvidenceKindsIndependent(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "pass" || len(result.Diagnostics) != 0 || len(result.Evidence) != 5 {
		t.Fatalf("Run() = %#v", result)
	}
	expected := map[diagnostic.EvidenceKind]string{
		diagnostic.EvidenceDefinition:             "all",
		diagnostic.EvidenceNativeSource:           "react-native",
		diagnostic.EvidenceDesignDocumentSource:   "pencil",
		diagnostic.EvidenceDesignDocumentComputed: "pencil",
		diagnostic.EvidenceConsumerConformance:    "react-native",
	}
	for kind, platform := range expected {
		evidence := findEvidence(t, result.Evidence, kind)
		if evidence.Status != "pass" || evidence.Platform != platform {
			t.Fatalf("evidence = %#v", result.Evidence)
		}
	}
}

func TestRunnerRecordsProfileRulesAsNotApplicableWithoutEvidence(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Conformance = nil
	request.Policy.Profile.RequiredEvidence = nil
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, activation := range result.RuleSet.Rules {
		if strings.HasPrefix(activation.RuleID, "profile/") && activation.Status != report.RuleNotApplicable {
			t.Fatalf("activation = %#v", activation)
		}
	}
}

func TestRunnerRecordsExactDisabledRule(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Policy.RuleOverrides = []policy.RuleOverride{{
		RuleID: rules.RuleProfileGratuitousMotion, Status: "disabled", Owner: "example-owner",
		Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "consumer contract changes",
	}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, activation := range result.RuleSet.Rules {
		if activation.RuleID == rules.RuleProfileGratuitousMotion && activation.Status != report.RuleDisabled {
			t.Fatalf("activation = %#v", activation)
		}
	}
}

func TestRunnerRejectsStaleGeneratedDesignContext(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.DesignContext = new(input(t, "../../../packages/schema/testdata/generated-design-context/.impeccable/design.json"))
	stale := append([]byte(nil), request.DesignContext.Contents...)
	prefix := []byte(`"contractSha256": "`)
	start := bytes.Index(stale, prefix) + len(prefix)
	if start < len(prefix) || start >= len(stale) {
		t.Fatal("contractSha256 not found")
	}
	if stale[start] == '0' {
		stale[start] = '1'
	} else {
		stale[start] = '0'
	}
	request.DesignContext.Contents = stale
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "fail" || !hasRule(result.Diagnostics, rules.RuleEvidenceStale) {
		t.Fatalf("result = %#v", result)
	}
}

func TestRunnerHonorsExactDesignRuleDisables(t *testing.T) {
	t.Parallel()
	for _, item := range []struct{ ruleID, path string }{
		{rules.RuleDesignSystemFont, "DesignFontDrift.tsx"},
		{rules.RuleDesignSystemColor, "DesignColorDrift.tsx"},
		{rules.RuleDesignSystemRadius, "DesignRadiusDrift.tsx"},
		{rules.RuleDesignSystemFontSize, "DesignSizeDrift.tsx"},
	} {
		request := positiveRequest(t)
		request.Sources = []lint.Input{{Path: item.path, Contents: []byte("fixture")}}
		request.DesignContext = new(input(t, "../../../packages/schema/testdata/generated-design-context/.impeccable/design.json"))
		request.Policy.RuleOverrides = []policy.RuleOverride{{
			RuleID: item.ruleID, Status: "disabled", Owner: "example-owner",
			Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "design contract changes",
		}}
		result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
		if err != nil {
			t.Fatal(err)
		}
		if hasRule(result.Diagnostics, item.ruleID) {
			t.Fatalf("disabled rule %s emitted: %#v", item.ruleID, result.Diagnostics)
		}
	}
}

func TestRunnerHonorsExactVisualRuleDisable(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.VisualDetails = []lint.Input{input(t, "../../../packages/schema/testdata/visual-detail-web.json")}
	request.Policy.RuleOverrides = []policy.RuleOverride{{
		RuleID: rules.RuleVisualSideTab, Status: "disabled", Owner: "example-owner",
		Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "visual contract changes",
	}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleVisualSideTab) {
		t.Fatalf("disabled visual rule emitted: %#v", result.Diagnostics)
	}
	for _, activation := range result.RuleSet.Rules {
		if activation.RuleID == rules.RuleVisualSideTab && activation.Status != report.RuleDisabled {
			t.Fatalf("activation = %#v", activation)
		}
	}
}

func TestRunnerHonorsExactTypographyRuleDisable(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Typographies = []lint.Input{input(t, "../../../packages/schema/testdata/typography-negative.json")}
	request.Policy.RuleOverrides = []policy.RuleOverride{{RuleID: rules.RuleTypographyTinyText, Status: "disabled", Owner: "example-owner", Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "typography policy changes"}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleTypographyTinyText) {
		t.Fatalf("disabled typography rule emitted: %#v", result.Diagnostics)
	}
}

func TestRunnerHonorsExactColorRuleDisable(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Colors = []lint.Input{
		input(t, "../../../packages/schema/testdata/color-negative-light.json"),
		input(t, "../../../packages/schema/testdata/color-permissions-dark.json"),
	}
	request.Policy.RuleOverrides = []policy.RuleOverride{{RuleID: rules.RuleColorGradientText, Status: "disabled", Owner: "example-owner", Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "color contract changes"}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleColorGradientText) {
		t.Fatalf("disabled color rule emitted: %#v", result.Diagnostics)
	}
	for _, activation := range result.RuleSet.Rules {
		if activation.RuleID == rules.RuleColorLowContrast && activation.Status != report.RuleActive {
			t.Fatalf("color evidence activation = %#v", activation)
		}
	}
}

func TestRunnerRequiresSeparateColorEvidenceForEveryTheme(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Colors = []lint.Input{input(t, "../../../packages/schema/testdata/color-negative-light.json")}
	_, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err == nil || !strings.Contains(err.Error(), "color evidence themes must be exact") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunnerHonorsExactLayoutDetailRuleDisable(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.LayoutDetails = []lint.Input{input(t, "../../../packages/schema/testdata/layout-negative-web.json")}
	request.Policy.RuleOverrides = []policy.RuleOverride{{RuleID: rules.RuleLayoutNestedCards, Status: "disabled", Owner: "example-owner", Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "layout contract changes"}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleLayoutNestedCards) {
		t.Fatalf("disabled layout rule emitted: %#v", result.Diagnostics)
	}
	for _, activation := range result.RuleSet.Rules {
		if activation.RuleID == rules.RuleLayoutLineLength && activation.Status != report.RuleActive {
			t.Fatalf("layout evidence activation = %#v", activation)
		}
	}
}

func TestRunnerUsesDefinitionMotionRegistryAndExactDisable(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Motions = []lint.Input{
		input(t, "../../../packages/schema/testdata/motion-negative-source.json"),
		input(t, "../../../packages/schema/testdata/motion-reduced-simulator.json"),
	}
	request.Policy.RuleOverrides = []policy.RuleOverride{{RuleID: rules.RuleMotionBounceEasing, Status: "disabled", Owner: "example-owner", Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "motion contract changes"}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleMotionBounceEasing) {
		t.Fatalf("disabled motion rule emitted: %#v", result.Diagnostics)
	}
	if !hasRule(result.Diagnostics, rules.RuleMotionPulsingDot) {
		t.Fatalf("active motion rule missing: %#v", result.Diagnostics)
	}
}

func TestRunnerKeepsCopyAdvisoryNonBlocking(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Copies = []lint.Input{input(t, "../../../packages/schema/testdata/copy-en-advisory.json")}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != report.StatusPass || !hasRule(result.Diagnostics, rules.RuleCopyEmDashOveruse) {
		t.Fatalf("advisory result = %#v", result)
	}
	for _, finding := range result.Diagnostics {
		if finding.RuleID == rules.RuleCopyEmDashOveruse && finding.Status != diagnostic.FindingAdvisory {
			t.Fatalf("em-dash finding = %#v", finding)
		}
	}
}

func TestRunnerRecordsContentRegistryNotRun(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Copies = []lint.Input{input(t, "../../../packages/schema/testdata/copy-registry-not-run.json")}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	registry := findEvidence(t, result.Evidence, diagnostic.EvidenceConsumerContentRegistry)
	if registry.Status != report.EvidenceStatusNotRun || hasRule(result.Diagnostics, rules.RuleCopyUnverifiedSocialProof) {
		t.Fatalf("registry/result = %#v %#v", registry, result.Diagnostics)
	}
}

func TestRunnerUsesAssetRegistryAndExactImageryDisable(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Imagery = []lint.Input{input(t, "../../../packages/schema/testdata/imagery-negative-web.json")}
	request.Policy.RuleOverrides = []policy.RuleOverride{{RuleID: rules.RuleImageryShapeAssembledIllustration, Status: "disabled", Owner: "example-owner", Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "asset registry changes"}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleImageryShapeAssembledIllustration) || !hasRule(result.Diagnostics, rules.RuleImageryBrokenImage) {
		t.Fatalf("imagery diagnostics = %#v", result.Diagnostics)
	}
}

func TestRunnerUsesRuntimeRegistryAndExactRuleDisable(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Runtimes = []lint.Input{input(t, "../../../packages/schema/testdata/runtime-negative-web.json")}
	request.Policy.RuleOverrides = []policy.RuleOverride{{RuleID: rules.RuleRuntimeContentHiddenAtRest, Status: "disabled", Owner: "example-owner", Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "runtime capture changes"}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleRuntimeContentHiddenAtRest) || !hasRule(result.Diagnostics, rules.RuleRuntimeScriptError) || !hasRule(result.Diagnostics, rules.RuleRuntimeJustifiedText) {
		t.Fatalf("runtime diagnostics = %#v", result.Diagnostics)
	}
}

func TestRunnerDoesNotUseWebRuntimeAsNativeCompletion(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Runtimes = []lint.Input{input(t, "../../../packages/schema/testdata/runtime-permissions.json")}
	request.Policy.Evidence.RequiredKinds = append(request.Policy.Evidence.RequiredKinds, string(diagnostic.EvidenceSimulator))
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	web := findEvidence(t, result.Evidence, diagnostic.EvidenceWebRendered)
	native := findEvidence(t, result.Evidence, diagnostic.EvidenceSimulator)
	if web.Status != report.EvidenceStatusPass || native.Status != report.EvidenceStatusNotRun || result.Status != report.StatusFail {
		t.Fatalf("runtime evidence = %#v", result.Evidence)
	}
}

func TestRunnerKeepsNativeSourceAndRuntimeConformanceSeparate(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.NativeSources = []lint.Input{input(t, "../../../packages/schema/testdata/native-source-positive.json")}
	request.NativeRuntimes = []lint.Input{
		input(t, "../../../packages/schema/testdata/native-runtime-positive-ios.json"),
		input(t, "../../../packages/schema/testdata/native-runtime-positive-android.json"),
	}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleNativeRuntimeMatrix) {
		t.Fatalf("native matrix diagnostics = %#v", result.Diagnostics)
	}
	if findEvidence(t, result.Evidence, diagnostic.EvidenceSimulator).Platform != "ios" || findEvidence(t, result.Evidence, diagnostic.EvidenceEmulator).Platform != "android" {
		t.Fatalf("native evidence = %#v", result.Evidence)
	}
}

func TestRunnerHonorsExactNativeRuleDisable(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.NativeSources = []lint.Input{input(t, "../../../packages/schema/testdata/native-source-negative.json")}
	request.Policy.RuleOverrides = []policy.RuleOverride{{RuleID: rules.RuleNativeListVirtualization, Status: "disabled", Owner: "example-owner", Rationale: "temporary exact compatibility", ExpiresAt: "2026-12-31", ReviewTrigger: "native source contract changes"}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if hasRule(result.Diagnostics, rules.RuleNativeListVirtualization) || !hasRule(result.Diagnostics, rules.RuleNativeRenderStability) {
		t.Fatalf("native source diagnostics = %#v", result.Diagnostics)
	}
}

func TestRunnerReportsRawSyntaxLayoutAndBudgetFailures(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Sources = []lint.Input{input(t, "../../testdata/negative/Raw.tsx")}
	request.Pencil = new(input(t, "../../testdata/negative/raw.pen.json"))
	request.Layout = new(input(t, "../../testdata/negative/layout.json"))
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "fail" || result.Summary.Raw != 6 || result.Summary.Overflow != 1 || result.Summary.Overlap != 1 {
		t.Fatalf("Run() summary = %#v", result.Summary)
	}
	for _, ruleID := range []string{rules.RuleSourceRawValue, rules.RulePencilRawValue, rules.RuleLayoutProblem, rules.RulePolicyBudgetExceeded} {
		if !hasRule(result.Diagnostics, ruleID) {
			t.Fatalf("missing %s in %#v", ruleID, result.Diagnostics)
		}
	}
}

func TestRunnerDoesNotTurnMissingEvidenceOrSyntaxErrorsIntoPass(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Pencil = nil
	request.Sources = []lint.Input{input(t, "../../testdata/negative/Broken.tsx")}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "fail" || !hasRule(result.Diagnostics, rules.RuleEvidenceMissing) || !hasRule(result.Diagnostics, rules.RuleSourceSyntaxError) {
		t.Fatalf("Run() = %#v", result)
	}
}

func TestRunnerDoesNotBudgetRequiredEvidenceOrParserFailuresIntoPass(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Pencil = nil
	request.Sources = []lint.Input{input(t, "../../testdata/negative/Broken.tsx")}
	request.Policy.Budgets.Error = 100
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "fail" {
		t.Fatalf("Run() status = %q", result.Status)
	}
}

func TestRunnerRejectsStaleLayoutEvidence(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Policy.Evidence.LayoutDocumentSHA256 = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "fail" || !hasRule(result.Diagnostics, rules.RuleEvidenceStale) {
		t.Fatalf("Run() = %#v", result)
	}
}

func TestRunnerKeepsDeferredRuntimeEvidenceOutOfPass(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Policy.Evidence.DeferredKinds = []string{string(diagnostic.EvidencePhysicalDevice)}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	evidence := findEvidence(t, result.Evidence, diagnostic.EvidencePhysicalDevice)
	if evidence.Status != "deferred" || result.Status != "pass" {
		t.Fatalf("Run() = %#v", result)
	}
}

func TestRunnerPreservesExactFalsePositiveRecords(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	negative := input(t, "../../testdata/negative/Raw.tsx")
	request.Sources = []lint.Input{negative}
	request.Policy.Exceptions = []policy.Exception{{
		RuleID: rules.RuleSourceRawValue, Engine: string(diagnostic.EvidenceNativeSource), Platform: "react-native",
		Path: negative.Path, Owner: "ansldes/source", Rationale: "approved exact migration fixture",
		Reviewer: "example-reviewer", ExpiresAt: "2026-12-31", ReviewTrigger: "source implementation or owner changes",
	}}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "pass" || len(result.FalsePositives) == 0 || result.Summary.FalsePositives != len(result.FalsePositives) {
		t.Fatalf("Run() = %#v", result)
	}
	for _, falsePositive := range result.FalsePositives {
		if falsePositive.OwnerFingerprint == "" || falsePositive.Status != "false-positive" {
			t.Fatalf("false positive = %#v", falsePositive)
		}
	}
	if evidence := findEvidence(t, result.Evidence, diagnostic.EvidenceNativeSource); evidence.Status != "false-positive" {
		t.Fatalf("source evidence = %#v", evidence)
	}
}

func TestRunnerCompletesExactWebProviderMatrixAndPreservesArtifactExclusion(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	for _, fixture := range []string{
		"web-provider-artifact-excluded.json",
		"web-provider-static-positive.json",
		"web-provider-browser-positive.json",
		"web-provider-visual-positive.json",
	} {
		request.WebProviders = append(request.WebProviders, input(t, "../../../packages/schema/testdata/"+fixture))
	}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != report.StatusPass || len(result.Diagnostics) != 0 || len(result.FalsePositives) != 1 {
		t.Fatalf("Web matrix result = %#v", result)
	}
	if falsePositive := result.FalsePositives[0]; falsePositive.Status != report.EvidenceStatusFalsePositive || !strings.Contains(falsePositive.Rationale, "reproduce: npm run build:example") {
		t.Fatalf("Web artifact false positive = %#v", falsePositive)
	}
	webStatuses := 0
	for _, item := range result.Evidence {
		if item.Kind == diagnostic.EvidenceWebSource || item.Kind == diagnostic.EvidenceWebRendered {
			webStatuses++
			if item.Status != report.EvidenceStatusPass && item.Status != report.EvidenceStatusFalsePositive {
				t.Fatalf("Web evidence status = %#v", item)
			}
		}
	}
	if webStatuses != 4 {
		t.Fatalf("Web evidence count = %d, want 4: %#v", webStatuses, result.Evidence)
	}
	if len(result.RuleSet.Rules) != len(rules.AllRuleIDs) {
		t.Fatalf("effective rule count = %d, registry = %d", len(result.RuleSet.Rules), len(rules.AllRuleIDs))
	}
	for index, activation := range result.RuleSet.Rules {
		if activation.RuleID != rules.AllRuleIDs[index] {
			t.Fatalf("effective rule[%d] = %q, registry = %q", index, activation.RuleID, rules.AllRuleIDs[index])
		}
	}
}

func TestRunnerKeepsWebFallbackNotRunSeparateFromFullPass(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Policy.Web.RequiredCaptures = []policy.WebCaptureRequirement{{
		ID: "static-html-light", Provider: "static-html", RouteID: "example-home", ViewportID: "desktop", Theme: "light", FontScale: 1, ReduceMotion: false,
	}}
	request.WebProviders = []lint.Input{input(t, "../../../packages/schema/testdata/web-provider-fallback-not-run.json")}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != report.StatusFail || !hasRule(result.Diagnostics, rules.RuleEvidenceMissing) {
		t.Fatalf("fallback result = %#v", result)
	}
	seenNotRun := false
	for _, item := range result.Evidence {
		if item.Path == "../../../packages/schema/testdata/web-provider-fallback-not-run.json" && item.Status == report.EvidenceStatusNotRun {
			seenNotRun = true
		}
	}
	if !seenNotRun {
		t.Fatalf("fallback evidence = %#v", result.Evidence)
	}
}

func TestRunnerReturnsTypedWebProviderExecutionError(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.WebProviders = []lint.Input{input(t, "../../../packages/schema/testdata/web-provider-browser-error.json")}
	_, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	var providerError *webcheck.ProviderExecutionError
	if !errors.As(err, &providerError) || providerError.CaptureID != "browser-dark-large" {
		t.Fatalf("Web provider error = %v", err)
	}
}

func TestRunnerProducesTheSameNeutralMultiStageReportForAnyInputOrder(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.LayoutDetails = []lint.Input{
		input(t, "../../../packages/schema/testdata/layout-design-document.json"),
		input(t, "../../../packages/schema/testdata/layout-permissions-native.json"),
	}
	request.Motions = []lint.Input{input(t, "../../../packages/schema/testdata/motion-design-document.json")}
	request.Copies = []lint.Input{input(t, "../../../packages/schema/testdata/copy-ko-positive.json")}
	request.Imagery = []lint.Input{input(t, "../../../packages/schema/testdata/imagery-permissions.json")}
	request.Runtimes = []lint.Input{input(t, "../../../packages/schema/testdata/runtime-permissions.json")}
	request.NativeSources = []lint.Input{input(t, "../../../packages/schema/testdata/native-source-positive.json")}
	request.NativeRuntimes = []lint.Input{
		input(t, "../../../packages/schema/testdata/native-runtime-positive-ios.json"),
		input(t, "../../../packages/schema/testdata/native-runtime-positive-android.json"),
	}
	request.WebProviders = []lint.Input{
		input(t, "../../../packages/schema/testdata/web-provider-regex-positive.json"),
		input(t, "../../../packages/schema/testdata/web-provider-static-positive.json"),
		input(t, "../../../packages/schema/testdata/web-provider-browser-positive.json"),
		input(t, "../../../packages/schema/testdata/web-provider-visual-positive.json"),
	}

	runner := lint.Runner{SourceAnalyzer: testAnalyzer{}}
	first, err := runner.Run(request)
	if err != nil {
		t.Fatal(err)
	}
	reverseInputs(request.LayoutDetails)
	reverseInputs(request.NativeRuntimes)
	reverseInputs(request.WebProviders)
	second, err := runner.Run(request)
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != report.StatusPass || len(first.Diagnostics) != 0 || len(first.FalsePositives) != 0 {
		t.Fatalf("neutral multi-stage result = %#v", first)
	}
	if !reflect.DeepEqual(first, second) || first.FingerprintSHA256 != second.FingerprintSHA256 {
		t.Fatalf("input ordering changed report: %#v %#v", first, second)
	}
}

func TestRunnerReportsStageCommandOutputAndRejectsDependencyDrift(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	stageInput := input(t, "../../../packages/schema/testdata/stage-execution-positive.json")
	request.StageExecutions = []lint.Input{stageInput}
	result, err := (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil || result.Status != report.StatusPass || len(result.StageExecutions) != 1 || result.StageExecutions[0].Stdout == "" {
		t.Fatalf("stage result = %#v, %v", result, err)
	}
	request.StageExecutions[0].Contents = bytes.Replace(
		stageInput.Contents,
		[]byte(`"observedDependencySha256": "1111111111111111111111111111111111111111111111111111111111111111"`),
		[]byte(`"observedDependencySha256": "2222222222222222222222222222222222222222222222222222222222222222"`),
		1,
	)
	result, err = (lint.Runner{SourceAnalyzer: testAnalyzer{}}).Run(request)
	if err != nil || result.Status != report.StatusFail || !hasRule(result.Diagnostics, rules.RuleEvidenceStale) {
		t.Fatalf("stale stage result = %#v, %v", result, err)
	}
}

func reverseInputs(values []lint.Input) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func positiveRequest(t *testing.T) lint.Request {
	t.Helper()
	policyContents := read(t, "../../../packages/schema/testdata/example-policy.json")
	productPolicy, err := policy.Parse(policyContents)
	if err != nil {
		t.Fatal(err)
	}
	return lint.Request{
		Definition:  input(t, "../../../packages/schema/testdata/example-product.json"),
		Policy:      productPolicy,
		Sources:     []lint.Input{input(t, "../../testdata/positive/Example.tsx")},
		Pencil:      new(input(t, "../../testdata/positive/document.pen.json")),
		Layout:      new(input(t, "../../testdata/positive/layout.json")),
		Conformance: new(input(t, "../../../packages/schema/testdata/operate-conformance.json")),
		Now:         time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC),
	}
}

func input(t *testing.T, path string) lint.Input {
	t.Helper()
	return lint.Input{Path: path, Contents: read(t, path)}
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

func hasRule(diagnostics []diagnostic.Diagnostic, ruleID string) bool {
	for _, finding := range diagnostics {
		if finding.RuleID == ruleID {
			return true
		}
	}
	return false
}

func findEvidence(t *testing.T, evidence []report.EvidenceStatus, kind diagnostic.EvidenceKind) report.EvidenceStatus {
	t.Helper()
	for _, record := range evidence {
		if record.Kind == kind {
			return record
		}
	}
	t.Fatalf("missing evidence %q in %#v", kind, evidence)
	return report.EvidenceStatus{}
}

type testAnalyzer struct{}

func (testAnalyzer) Analyze(path string, _ []byte, language source.Language) (source.Summary, error) {
	summary := source.Summary{Path: path, Language: language, RootKind: "program"}
	if strings.Contains(path, "Broken.tsx") {
		summary.HasError = true
	}
	if strings.Contains(path, "Raw.tsx") {
		summary.PropertyLiterals = []source.PropertyLiteral{
			literal("backgroundColor", "string", "#ffffff", 2),
			literal("borderRadius", "number", "12", 3),
			literal("duration", "number", "160", 4),
			literal("gap", "number", "8", 5),
		}
	}
	if strings.Contains(path, "DesignFontDrift.tsx") {
		summary.PropertyLiterals = []source.PropertyLiteral{literal("fontFamily", "string", "Unknown Sans", 2)}
	}
	if strings.Contains(path, "DesignColorDrift.tsx") {
		summary.PropertyLiterals = []source.PropertyLiteral{literal("backgroundColor", "string", "#123456", 2)}
	}
	if strings.Contains(path, "DesignRadiusDrift.tsx") {
		summary.PropertyLiterals = []source.PropertyLiteral{literal("borderRadius", "number", "99", 2)}
	}
	if strings.Contains(path, "DesignSizeDrift.tsx") {
		summary.PropertyLiterals = []source.PropertyLiteral{literal("fontSize", "number", "15", 2)}
	}
	return summary, nil
}

func literal(property, kind, value string, line int) source.PropertyLiteral {
	return source.PropertyLiteral{
		Property: property,
		Kind:     kind,
		Value:    value,
		Range: source.Range{
			Start: source.Position{Line: line, Column: 3},
			End:   source.Position{Line: line, Column: 12},
		},
	}
}
