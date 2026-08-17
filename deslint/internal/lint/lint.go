// Package lint orchestrates independently acquired evidence through one rule model.
package lint

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/aprilgom/AnslDes/deslint/internal/colorcheck"
	"github.com/aprilgom/AnslDes/deslint/internal/conformance"
	"github.com/aprilgom/AnslDes/deslint/internal/contract"
	"github.com/aprilgom/AnslDes/deslint/internal/copycheck"
	"github.com/aprilgom/AnslDes/deslint/internal/designcontext"
	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/imagerycheck"
	"github.com/aprilgom/AnslDes/deslint/internal/layout"
	"github.com/aprilgom/AnslDes/deslint/internal/layoutdetail"
	"github.com/aprilgom/AnslDes/deslint/internal/motioncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/nativecheck"
	"github.com/aprilgom/AnslDes/deslint/internal/pen"
	"github.com/aprilgom/AnslDes/deslint/internal/policy"
	"github.com/aprilgom/AnslDes/deslint/internal/report"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
	"github.com/aprilgom/AnslDes/deslint/internal/runtimecheck"
	"github.com/aprilgom/AnslDes/deslint/internal/source"
	"github.com/aprilgom/AnslDes/deslint/internal/stage"
	"github.com/aprilgom/AnslDes/deslint/internal/typography"
	"github.com/aprilgom/AnslDes/deslint/internal/visualdetail"
	"github.com/aprilgom/AnslDes/deslint/internal/webcheck"
)

// Input is one named immutable evidence payload.
type Input struct {
	Path     string
	Contents []byte
}

// Request contains all inputs without filesystem or process coupling.
type Request struct {
	Definition      Input
	Policy          policy.Policy
	Sources         []Input
	Pencil          *Input
	Layout          *Input
	Conformance     *Input
	DesignContext   *Input
	VisualDetails   []Input
	Typographies    []Input
	Colors          []Input
	LayoutDetails   []Input
	Motions         []Input
	Copies          []Input
	Imagery         []Input
	Runtimes        []Input
	NativeSources   []Input
	NativeRuntimes  []Input
	WebProviders    []Input
	StageExecutions []Input
	Now             time.Time
}

// Runner evaluates evidence with a parser-independent source analyzer.
type Runner struct {
	SourceAnalyzer source.Analyzer
}

// Run returns a deterministic report. Malformed inputs return an execution error.
func (r Runner) Run(request Request) (report.Report, error) {
	if r.SourceAnalyzer == nil {
		return report.Report{}, fmt.Errorf("source analyzer is required")
	}
	severity := request.Policy.Severity
	definition, err := contract.Analyze(request.Definition.Path, request.Definition.Contents, severity)
	if err != nil {
		return report.Report{}, err
	}
	diagnostics := append([]diagnostic.Diagnostic(nil), definition.Diagnostics...)
	providerFalsePositives := []report.FalsePositive{}
	var generatedContext *designcontext.Context
	if request.DesignContext != nil {
		parsed, parseErr := designcontext.Parse(request.DesignContext.Contents)
		if parseErr != nil {
			return report.Report{}, parseErr
		}
		contractFingerprint, fingerprintErr := designcontext.ContractFingerprint(request.Definition.Contents)
		if fingerprintErr != nil {
			return report.Report{}, fingerprintErr
		}
		if parsed.Source.ContractSHA256 != contractFingerprint {
			diagnostics = append(diagnostics, diagnostic.New(rules.RuleEvidenceStale, request.Policy.Severity(rules.RuleEvidenceStale), "generated design context does not match the canonical definition", request.DesignContext.Path, nil, diagnostic.EvidenceExecution, "all", "ansldes/design-system", "stale"))
		} else {
			generatedContext = &parsed
		}
	}
	evidence := []report.EvidenceStatus{{
		Kind: diagnostic.EvidenceDefinition, Platform: "all", Status: report.EvidenceStatusPass, Path: request.Definition.Path,
	}}
	if definition.DefinitionID != request.Policy.DefinitionID {
		diagnostics = append(diagnostics, diagnostic.New(
			rules.RulePolicyDefinitionMismatch,
			diagnostic.SeverityError,
			fmt.Sprintf("policy definitionId %s does not match definition %s", request.Policy.DefinitionID, definition.DefinitionID),
			request.Definition.Path,
			nil,
			diagnostic.EvidenceDefinition,
			"all",
			"ansldes/policy",
			"policy",
		))
	}

	sourceAcquired := false
	for _, input := range request.Sources {
		if request.Policy.IsExcluded(input.Path) {
			continue
		}
		language, languageErr := languageForPath(input.Path)
		if languageErr != nil {
			return report.Report{}, languageErr
		}
		summary, analyzeErr := r.SourceAnalyzer.Analyze(input.Path, input.Contents, language)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		sourceAcquired = true
		diagnostics = append(diagnostics, rules.AnalyzeSourceWithDesignContext(summary, request.Policy.RawPropertyKinds(), severity, generatedContext)...)
	}
	if sourceAcquired {
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceNativeSource, Platform: "react-native", Status: report.EvidenceStatusPass})
	} else {
		evidence = append(evidence, missingEvidence(request.Policy, diagnostic.EvidenceNativeSource, "react-native", &diagnostics))
	}

	if request.Pencil != nil {
		findings, analyzeErr := pen.Analyze(request.Pencil.Path, request.Pencil.Contents, request.Policy.RawPropertyKinds(), severity)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceDesignDocumentSource, Platform: "pencil", Status: report.EvidenceStatusPass, Path: request.Pencil.Path})
	} else {
		evidence = append(evidence, missingEvidence(request.Policy, diagnostic.EvidenceDesignDocumentSource, "pencil", &diagnostics))
	}

	if request.Layout != nil {
		findings, analyzeErr := layout.Analyze(request.Layout.Path, request.Layout.Contents, request.Policy.Evidence.LayoutDocumentSHA256, severity)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceDesignDocumentComputed, Platform: "pencil", Status: report.EvidenceStatusPass, Path: request.Layout.Path})
	} else {
		evidence = append(evidence, missingEvidence(request.Policy, diagnostic.EvidenceDesignDocumentComputed, "pencil", &diagnostics))
	}

	if request.Conformance != nil {
		profileID := ""
		maxOversized, maxInconsistent := 0, 0
		if request.Policy.Profile != nil {
			profileID = request.Policy.Profile.ID
			maxOversized = request.Policy.Profile.Thresholds.MaxOversizedActions
			maxInconsistent = request.Policy.Profile.Thresholds.MaxInconsistentActions
		}
		result, analyzeErr := conformance.Analyze(request.Conformance.Path, request.Conformance.Contents, conformance.Config{
			ProfileID: profileID, MaxOversizedActions: maxOversized, MaxInconsistentActions: maxInconsistent,
			Severity: request.Policy.Severity,
			Active:   func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) },
		})
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, result.Diagnostics...)
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceConsumerConformance, Platform: result.Platform, Status: report.EvidenceStatusPass, Path: request.Conformance.Path})
	} else {
		evidence = append(evidence, missingEvidence(request.Policy, diagnostic.EvidenceConsumerConformance, "all", &diagnostics))
	}
	for _, input := range request.VisualDetails {
		visualEvidence, findings, analyzeErr := visualdetail.Analyze(input.Path, input.Contents, severity, func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) })
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: visualEvidence.EvidenceKind, Platform: visualEvidence.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	for _, input := range request.Typographies {
		typeEvidence, findings, analyzeErr := typography.Analyze(input.Path, input.Contents, severity, func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) })
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: typeEvidence.EvidenceKind, Platform: typeEvidence.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	colorThemes := map[string]bool{}
	colorConfig := colorcheck.Config{
		Severity: severity,
		Active:   func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) },
	}
	if len(request.Colors) > 0 {
		if definition.ColorUsage == nil {
			return report.Report{}, fmt.Errorf("color evidence requires colorUsage in the canonical definition")
		}
		colorConfig.BodyContrast = definition.ColorUsage.Contrast.Body
		colorConfig.LargeContrast = definition.ColorUsage.Contrast.Large
		colorConfig.ApprovedPalettes = make(map[string]colorcheck.PalettePermission, len(definition.ColorUsage.ApprovedPalettes))
		for id, permission := range definition.ColorUsage.ApprovedPalettes {
			colorConfig.ApprovedPalettes[id] = colorcheck.PalettePermission{Contexts: append([]string(nil), permission.Contexts...), Themes: append([]string(nil), permission.Themes...)}
		}
	}
	layoutProfileID, layoutDensity := "", "comfortable"
	if request.Policy.Profile != nil {
		layoutProfileID = request.Policy.Profile.ID
		layoutDensity = request.Policy.Profile.Density
	}
	for _, input := range request.LayoutDetails {
		layoutEvidence, findings, analyzeErr := layoutdetail.Analyze(input.Path, input.Contents, layoutdetail.Config{
			ProfileID: layoutProfileID,
			Density:   layoutDensity,
			Severity:  severity,
			Active:    func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) },
		})
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: layoutEvidence.EvidenceKind, Platform: layoutEvidence.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	motionRegistry := make(map[string]motioncheck.Transition, len(definition.Motion))
	for id, transition := range definition.Motion {
		motionRegistry[id] = motioncheck.Transition{
			Owner: transition.Owner, Purpose: transition.Purpose, DurationMS: transition.DurationMS,
			ReducedDurationMS: transition.ReducedDurationMS, Easing: append([]float64(nil), transition.Easing...), ReducedFallback: transition.ReducedFallback,
		}
	}
	for _, input := range request.Motions {
		motionEvidence, findings, analyzeErr := motioncheck.Analyze(input.Path, input.Contents, motioncheck.Config{
			ProfileID: layoutProfileID,
			Registry:  motionRegistry,
			Severity:  severity,
			Active:    func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) },
		})
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: motionEvidence.EvidenceKind, Platform: motionEvidence.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	copyConfig := copycheck.Config{
		ProfileID: layoutProfileID,
		Severity:  severity,
		Active:    func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) },
	}
	if len(request.Copies) > 0 {
		if request.Policy.Content == nil {
			return report.Report{}, fmt.Errorf("copy evidence requires a versioned content registry in consumer policy")
		}
		copyConfig.RegistryVersion = request.Policy.Content.RegistryVersion
		copyConfig.SourceReferences = append([]string(nil), request.Policy.Content.SourceReferences...)
		copyConfig.LocalePolicies = make(map[string]copycheck.LocalePolicy, len(request.Policy.Content.Locales))
		for locale, content := range request.Policy.Content.Locales {
			copyConfig.LocalePolicies[locale] = copycheck.LocalePolicy{
				PhraseRegistryVersion: content.PhraseRegistryVersion,
				MarketingBuzzwords:    append([]string(nil), content.MarketingBuzzwords...), TheaterPhrases: append([]string(nil), content.TheaterPhrases...),
				ProtectedTerms: append([]string(nil), content.ProtectedTerms...), RecoveryCopyIDs: append([]string(nil), content.RecoveryCopyIDs...),
			}
		}
	}
	for _, input := range request.Copies {
		copyEvidence, findings, analyzeErr := copycheck.Analyze(input.Path, input.Contents, copyConfig)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: copyEvidence.EvidenceKind, Platform: copyEvidence.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
		registryStatus := report.EvidenceStatusPass
		if copyEvidence.ContentRegistryStatus == "not-run" {
			registryStatus = report.EvidenceStatusNotRun
		}
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceConsumerContentRegistry, Platform: "all", Status: registryStatus, Path: input.Path})
	}
	imageryConfig := imagerycheck.Config{Severity: severity, Active: func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) }}
	if len(request.Imagery) > 0 {
		if request.Policy.Assets == nil {
			return report.Report{}, fmt.Errorf("imagery evidence requires a versioned asset registry in consumer policy")
		}
		imageryConfig.RegistryVersion = request.Policy.Assets.RegistryVersion
		imageryConfig.Assets = make(map[string]imagerycheck.Asset, len(request.Policy.Assets.Entries))
		for id, asset := range request.Policy.Assets.Entries {
			imageryConfig.Assets[id] = imagerycheck.Asset{
				Owner: asset.Owner, Role: asset.Role, ImplementationSource: asset.ImplementationSource,
				Consumers: append([]string(nil), asset.Consumers...), FingerprintSHA256: asset.FingerprintSHA256,
				IntentionallyOmitted: asset.IntentionallyOmitted, Decorative: asset.Decorative,
			}
		}
	}
	for _, input := range request.Imagery {
		imageryEvidence, findings, analyzeErr := imagerycheck.Analyze(input.Path, input.Contents, imageryConfig)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: imageryEvidence.EvidenceKind, Platform: imageryEvidence.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	runtimeConfig := runtimecheck.Config{Severity: severity, Active: func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) }}
	if len(request.Runtimes) > 0 {
		if request.Policy.Runtime == nil {
			return report.Report{}, fmt.Errorf("runtime evidence requires a versioned runtime registry in consumer policy")
		}
		runtimeConfig.RegistryVersion = request.Policy.Runtime.RegistryVersion
		runtimeConfig.JustifiedTextExceptions = make([]runtimecheck.JustifiedTextException, 0, len(request.Policy.Runtime.JustifiedTextExceptions))
		for _, exception := range request.Policy.Runtime.JustifiedTextExceptions {
			runtimeConfig.JustifiedTextExceptions = append(runtimeConfig.JustifiedTextExceptions, runtimecheck.JustifiedTextException{
				Platform: exception.Platform, SurfaceID: exception.SurfaceID, RouteID: exception.RouteID,
				NodeID: exception.NodeID, Owner: exception.Owner, Context: exception.Context,
			})
		}
	}
	for _, input := range request.Runtimes {
		runtimeEvidence, findings, analyzeErr := runtimecheck.Analyze(input.Path, input.Contents, runtimeConfig)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: runtimeEvidence.EvidenceKind, Platform: runtimeEvidence.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	nativeConfig := nativecheck.Config{Severity: severity, Active: func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) }}
	if len(request.NativeSources) > 0 || len(request.NativeRuntimes) > 0 {
		if request.Policy.Native == nil {
			return report.Report{}, fmt.Errorf("native conformance evidence requires a versioned native registry in consumer policy")
		}
		nativeConfig.RegistryVersion = request.Policy.Native.RegistryVersion
		nativeConfig.IOSAdjacentTargetSpacing = request.Policy.Native.IOSAdjacentTargetSpacing
		thresholds := request.Policy.Native.Thresholds
		nativeConfig.Thresholds = nativecheck.Thresholds{
			MaxSynchronousStartupMS: thresholds.MaxSynchronousStartupMS, MaxInitializationMS: thresholds.MaxInitializationMS,
			MaxMainThreadWorkMS: thresholds.MaxMainThreadWorkMS, MaxFrameDropRatio: thresholds.MaxFrameDropRatio,
			MaxThumbnailDecodeRatio: thresholds.MaxThumbnailDecodeRatio, MaxJSBundleRegressionBytes: thresholds.MaxJSBundleRegressionBytes,
			MaxAppBinaryRegressionBytes: thresholds.MaxAppBinaryRegressionBytes,
		}
		nativeConfig.RequiredRuntimeCaptures = make([]nativecheck.RuntimeRequirement, 0, len(request.Policy.Native.RequiredRuntimeCaptures))
		for _, requirement := range request.Policy.Native.RequiredRuntimeCaptures {
			nativeConfig.RequiredRuntimeCaptures = append(nativeConfig.RequiredRuntimeCaptures, nativecheck.RuntimeRequirement{
				ID: requirement.ID, Platform: requirement.Platform, EvidenceKind: diagnostic.EvidenceKind(requirement.EvidenceKind),
				FormFactor: requirement.FormFactor, Orientation: requirement.Orientation, WindowMode: requirement.WindowMode,
				FoldPosture: requirement.FoldPosture, Theme: requirement.Theme, MinimumFontScale: requirement.MinimumFontScale,
				ReduceMotion: requirement.ReduceMotion,
			})
		}
	}
	for _, input := range request.NativeSources {
		nativeSource, findings, analyzeErr := nativecheck.AnalyzeSource(input.Path, input.Contents, nativeConfig)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceNativeSource, Platform: nativeSource.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	nativeRuntimeEvidence := make([]nativecheck.RuntimeEvidence, 0, len(request.NativeRuntimes))
	for _, input := range request.NativeRuntimes {
		nativeRuntime, findings, analyzeErr := nativecheck.AnalyzeRuntime(input.Path, input.Contents, nativeConfig)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		nativeRuntimeEvidence = append(nativeRuntimeEvidence, nativeRuntime)
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: nativeRuntime.EvidenceKind, Platform: nativeRuntime.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	if len(request.NativeRuntimes) > 0 {
		diagnostics = append(diagnostics, nativecheck.CoverageFindings("<native-runtime>", nativeRuntimeEvidence, nativeConfig)...)
	}
	for _, input := range request.Colors {
		colorEvidence, findings, analyzeErr := colorcheck.Analyze(input.Path, input.Contents, colorConfig)
		if analyzeErr != nil {
			return report.Report{}, analyzeErr
		}
		colorThemes[colorEvidence.Theme] = true
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: colorEvidence.EvidenceKind, Platform: colorEvidence.Platform, Status: report.EvidenceStatusPass, Path: input.Path})
	}
	if len(request.Colors) > 0 {
		actualThemes := make([]string, 0, len(colorThemes))
		for theme := range colorThemes {
			actualThemes = append(actualThemes, theme)
		}
		expectedThemes := append([]string(nil), definition.Themes...)
		slices.Sort(actualThemes)
		slices.Sort(expectedThemes)
		if !slices.Equal(actualThemes, expectedThemes) {
			return report.Report{}, fmt.Errorf("color evidence themes must be exact: got %v, want %v", actualThemes, expectedThemes)
		}
	}
	webEvidence := make([]webcheck.Evidence, 0, len(request.WebProviders))
	if len(request.WebProviders) > 0 {
		if request.Policy.Web == nil {
			return report.Report{}, fmt.Errorf("web provider evidence requires a versioned Web registry in consumer policy")
		}
		webConfig := webConfigFromPolicy(request)
		for _, input := range request.WebProviders {
			providerEvidence, findings, excluded, analyzeErr := webcheck.Analyze(input.Path, input.Contents, webConfig)
			if analyzeErr != nil {
				return report.Report{}, analyzeErr
			}
			webEvidence = append(webEvidence, providerEvidence)
			diagnostics = append(diagnostics, findings...)
			status := report.EvidenceStatusPass
			if providerEvidence.Execution.Status == "not-run" {
				status = report.EvidenceStatusNotRun
			}
			evidence = append(evidence, report.EvidenceStatus{Kind: providerEvidence.EvidenceKind, Platform: providerEvidence.Platform, Status: status, Path: input.Path})
			for _, item := range excluded {
				providerFalsePositives = append(providerFalsePositives, report.NewFalsePositive(
					item.Finding,
					item.Exclusion.Owner,
					item.Exclusion.Rationale+"; reproduce: "+item.Exclusion.ReproductionCommand,
				))
			}
		}
		diagnostics = append(diagnostics, webcheck.CoverageFindings(webEvidence, webConfig)...)
	}
	stageExecutions := make([]report.StageExecution, 0, len(request.StageExecutions))
	for _, input := range request.StageExecutions {
		execution, parseErr := stage.Parse(input.Contents)
		if parseErr != nil {
			return report.Report{}, parseErr
		}
		stageExecutions = append(stageExecutions, execution)
		status := report.EvidenceStatusPass
		if execution.Status == "fail" || !stage.Fresh(execution) {
			status = report.EvidenceStatusFail
		}
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceExecution, Platform: "all", Status: status, Path: execution.StageID})
		if execution.Status == "fail" {
			diagnostics = append(diagnostics, diagnostic.New(
				rules.RuleEvidenceMissing, request.Policy.Severity(rules.RuleEvidenceMissing),
				fmt.Sprintf("stage %s command %q failed with exit code %d", execution.StageID, execution.Command, execution.ExitCode),
				input.Path, nil, diagnostic.EvidenceExecution, execution.Platform, execution.Owner, "stage-execution",
			))
		}
		if !stage.Fresh(execution) {
			diagnostics = append(diagnostics, diagnostic.New(
				rules.RuleEvidenceStale, request.Policy.Severity(rules.RuleEvidenceStale),
				fmt.Sprintf("stage %s dependency checksum is stale", execution.StageID),
				input.Path, nil, diagnostic.EvidenceExecution, execution.Platform, execution.Owner, "stale",
			))
		}
	}

	for _, configured := range []struct {
		kind     diagnostic.EvidenceKind
		platform string
	}{
		{diagnostic.EvidenceWebSource, "web"},
		{diagnostic.EvidenceWebRendered, "web"},
		{diagnostic.EvidenceSimulator, "ios"},
		{diagnostic.EvidenceEmulator, "android"},
		{diagnostic.EvidencePhysicalDevice, "native"},
		{diagnostic.EvidenceConsumerContentRegistry, "all"},
	} {
		if (request.Policy.Requires(configured.kind) || request.Policy.Deferred(configured.kind)) && !hasEvidenceKind(evidence, configured.kind) {
			evidence = append(evidence, missingEvidence(request.Policy, configured.kind, configured.platform, &diagnostics))
		}
	}

	for _, exception := range request.Policy.ExpiredExceptions(request.Now) {
		diagnostics = append(diagnostics, diagnostic.New(
			rules.RulePolicyExceptionExpired,
			diagnostic.SeverityError,
			fmt.Sprintf("exception for %s expired on %s", exception.RuleID, exception.ExpiresAt),
			exception.Path,
			nil,
			diagnostic.EvidenceExecution,
			"all",
			"ansldes/policy",
			"policy",
		))
	}
	for _, allowance := range request.Policy.ExpiredIgnores(request.Now) {
		diagnostics = append(diagnostics, diagnostic.New(
			rules.RulePolicyExceptionExpired,
			request.Policy.Severity(rules.RulePolicyExceptionExpired),
			fmt.Sprintf("%s ignore expired on %s", allowance.Kind, allowance.ExpiresAt),
			allowance.Path, nil, diagnostic.EvidenceExecution, "all", allowance.Owner, "expired-ignore",
		))
	}
	if request.Policy.GovernanceReviewOverdue(request.Now) {
		diagnostics = append(diagnostics, diagnostic.New(
			rules.RulePolicyExceptionExpired,
			request.Policy.Severity(rules.RulePolicyExceptionExpired),
			"90-day governance review is overdue",
			"<policy>#/governance/reviewedAt", nil, diagnostic.EvidenceExecution, "all", "ansldes/governance", "governance-review",
		))
	}
	for _, override := range request.Policy.RuleOverrides {
		expiresAt, parseErr := time.Parse("2006-01-02", override.ExpiresAt)
		if parseErr == nil && expiresAt.Before(midnightUTC(request.Now)) {
			diagnostics = append(diagnostics, diagnostic.New(
				rules.RulePolicyExceptionExpired,
				diagnostic.SeverityError,
				fmt.Sprintf("rule override for %s expired on %s", override.RuleID, override.ExpiresAt),
				"<policy>", nil, diagnostic.EvidenceExecution, "all", "ansldes/policy", "policy",
			))
		}
	}
	diagnostics = slices.DeleteFunc(diagnostics, func(finding diagnostic.Diagnostic) bool {
		return !request.Policy.RuleActiveAt(finding.RuleID, request.Now)
	})
	var exceptionMatches []policy.ExceptionMatch
	diagnostics, exceptionMatches = request.Policy.ClassifyExceptions(diagnostics, request.Now)
	falsePositives := append([]report.FalsePositive(nil), providerFalsePositives...)
	for _, match := range exceptionMatches {
		falsePositives = append(falsePositives, report.NewFalsePositive(
			match.Finding,
			match.Exception.Owner,
			match.Exception.Rationale,
		))
	}
	diagnostics, failed := enforceBudgets(diagnostics, falsePositives, evidence, request.Policy, request.Now)
	ruleSet, err := effectiveRuleSet(request)
	if err != nil {
		return report.Report{}, err
	}
	return report.New(report.Input{
		DefinitionID: definition.DefinitionID, RuleSet: ruleSet, Evidence: evidence,
		Diagnostics: diagnostics, FalsePositives: falsePositives, Failed: failed,
		StageExecutions: stageExecutions,
	})
}

func effectiveRuleSet(request Request) (report.RuleSet, error) {
	packs := make([]report.RulePack, 0, len(request.Policy.RulePacks))
	selectedRules := make(map[string]bool)
	for _, pack := range request.Policy.RulePacks {
		packs = append(packs, report.RulePack{ID: pack.ID, Version: pack.Version, FingerprintSHA256: pack.FingerprintSHA256})
		registered, found := rules.LookupPack(pack.ID, pack.Version)
		if found {
			for _, spec := range registered.Rules {
				selectedRules[spec.ID] = true
			}
		}
	}
	activations := make([]report.RuleActivation, 0, len(rules.AllRuleIDs))
	for _, ruleID := range rules.AllRuleIDs {
		activation := report.RuleActivation{RuleID: ruleID, Status: report.RuleActive}
		spec, _ := rules.Lookup(ruleID)
		if !selectedRules[ruleID] {
			activation.Status = report.RuleUnsupported
			activation.Reason = "owning rule pack is not required by policy"
		} else if override, ok := request.Policy.RuleOverride(ruleID); ok && !request.Policy.RuleActiveAt(ruleID, request.Now) {
			activation.Status = report.RuleDisabled
			activation.Reason = fmt.Sprintf("disabled by %s until %s; review trigger: %s", override.Owner, override.ExpiresAt, override.ReviewTrigger)
		} else if !ruleApplicable(spec, request) {
			activation.Status = report.RuleNotApplicable
			activation.Reason = "no applicable platform evidence was provided"
		}
		activations = append(activations, activation)
	}
	return report.NewRuleSet(packs, activations)
}

func ruleApplicable(spec rules.RuleSpec, request Request) bool {
	if len(request.WebProviders) > 0 && len(spec.Providers) > 0 {
		return true
	}
	if len(request.NativeSources) > 0 {
		for _, mapping := range spec.Applicability {
			if mapping.Target == "react-native" && mapping.Status != "unsupported" && slices.Contains(mapping.EvidenceKinds, string(diagnostic.EvidenceNativeSource)) {
				return true
			}
		}
	}
	directEvidence := false
	for _, required := range spec.RequiredInputs {
		if required == "design-context" && request.DesignContext == nil {
			return false
		}
		if required == "visual-detail" && len(request.VisualDetails) == 0 {
			return false
		} else if required == "visual-detail" {
			directEvidence = true
		}
		if required == "typography" && len(request.Typographies) == 0 {
			return false
		} else if required == "typography" {
			directEvidence = true
		}
		if required == "color" && len(request.Colors) == 0 {
			return false
		} else if required == "color" {
			directEvidence = true
		}
		if required == "layout-detail" && len(request.LayoutDetails) == 0 {
			return false
		} else if required == "layout-detail" {
			directEvidence = true
		}
		if required == "motion" && len(request.Motions) == 0 {
			return false
		} else if required == "motion" {
			directEvidence = true
		}
		if required == "copy" && len(request.Copies) == 0 {
			return false
		} else if required == "copy" {
			directEvidence = true
		}
		if required == "imagery" && len(request.Imagery) == 0 {
			return false
		} else if required == "imagery" {
			directEvidence = true
		}
		if required == "runtime" && len(request.Runtimes) == 0 {
			return false
		} else if required == "runtime" {
			directEvidence = true
		}
		if required == "native-source-conformance" && len(request.NativeSources) == 0 {
			return false
		} else if required == "native-source-conformance" {
			directEvidence = true
		}
		if required == "native-runtime" && len(request.NativeRuntimes) == 0 {
			return false
		} else if required == "native-runtime" {
			directEvidence = true
		}
		if required == "native-conformance" && len(request.NativeSources) == 0 && len(request.NativeRuntimes) == 0 {
			return false
		} else if required == "native-conformance" {
			directEvidence = true
		}
	}
	if directEvidence {
		return true
	}
	for _, kind := range spec.EvidenceKinds {
		switch kind {
		case "definition", "execution":
			return true
		case "native-source":
			if len(request.Sources) > 0 || len(request.NativeSources) > 0 {
				return true
			}
		case "design-document-source":
			if request.Pencil != nil {
				return true
			}
		case "design-document-computed":
			if request.Layout != nil {
				return true
			}
		case "consumer-conformance":
			if request.Conformance != nil {
				return true
			}
		}
	}
	return false
}

func webConfigFromPolicy(request Request) webcheck.Config {
	webPolicy := request.Policy.Web
	config := webcheck.Config{
		RegistryVersion: webPolicy.RegistryVersion,
		Routes:          make(map[string]webcheck.Route, len(webPolicy.Routes)),
		Viewports:       make(map[string]webcheck.Viewport, len(webPolicy.Viewports)),
		Themes:          append([]string(nil), webPolicy.Themes...),
		FontScales:      append([]float64(nil), webPolicy.FontScales...),
		ReduceMotion:    append([]bool(nil), webPolicy.ReduceMotion...),
		Severity:        request.Policy.Severity,
		Active:          func(ruleID string) bool { return request.Policy.RuleActiveAt(ruleID, request.Now) },
	}
	for _, route := range webPolicy.Routes {
		config.Routes[route.ID] = webcheck.Route{Owner: route.Owner, Target: route.Target}
	}
	for _, viewport := range webPolicy.Viewports {
		config.Viewports[viewport.ID] = webcheck.Viewport{ID: viewport.ID, Width: viewport.Width, Height: viewport.Height}
	}
	for _, capture := range webPolicy.RequiredCaptures {
		config.RequiredCaptures = append(config.RequiredCaptures, webcheck.CaptureRequirement{
			ID: capture.ID, Provider: capture.Provider, RouteID: capture.RouteID, ViewportID: capture.ViewportID,
			Theme: capture.Theme, FontScale: capture.FontScale, ReduceMotion: capture.ReduceMotion,
		})
	}
	for _, exclusion := range webPolicy.ArtifactExclusions {
		config.ArtifactExclusions = append(config.ArtifactExclusions, webcheck.ArtifactExclusion{
			Path: exclusion.Path, FingerprintSHA256: exclusion.FingerprintSHA256, Owner: exclusion.Owner,
			Rationale: exclusion.Rationale, ReproductionCommand: exclusion.ReproductionCommand,
		})
	}
	return config
}

func hasEvidenceKind(evidence []report.EvidenceStatus, kind diagnostic.EvidenceKind) bool {
	for _, record := range evidence {
		if record.Kind == kind {
			return true
		}
	}
	return false
}

func midnightUTC(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func missingEvidence(productPolicy policy.Policy, kind diagnostic.EvidenceKind, platform string, diagnostics *[]diagnostic.Diagnostic) report.EvidenceStatus {
	status := report.EvidenceStatusNotRun
	if productPolicy.Deferred(kind) {
		status = report.EvidenceStatusDeferred
	} else if productPolicy.Requires(kind) {
		*diagnostics = append(*diagnostics, diagnostic.New(
			rules.RuleEvidenceMissing,
			productPolicy.Severity(rules.RuleEvidenceMissing),
			fmt.Sprintf("required %s evidence was not provided", kind),
			"<evidence>",
			nil,
			diagnostic.EvidenceExecution,
			"all",
			"ansldes/evidence",
			"missing",
		))
	}
	return report.EvidenceStatus{Kind: kind, Platform: platform, Status: status}
}

func enforceBudgets(diagnostics []diagnostic.Diagnostic, falsePositives []report.FalsePositive, evidence []report.EvidenceStatus, productPolicy policy.Policy, now time.Time) ([]diagnostic.Diagnostic, bool) {
	counts := map[string]int{"error": 0, "warning": 0, "raw": 0, "overflow": 0, "overlap": 0, "blocking": 0, "advisory": 0, "exception": len(falsePositives), "not-run": 0, "deferred": 0}
	failed := false
	for _, finding := range diagnostics {
		if finding.RuleID != rules.RuleCopyEmDashOveruse {
			counts[string(finding.Severity)]++
		}
		switch finding.Status {
		case diagnostic.FindingFail:
			counts["blocking"]++
		case diagnostic.FindingAdvisory:
			counts["advisory"]++
		}
		if finding.RuleID == rules.RuleEvidenceMissing || finding.RuleID == rules.RuleEvidenceStale || finding.RuleID == rules.RuleSourceSyntaxError {
			failed = true
		}
		if finding.Category == "raw" {
			counts["raw"]++
		}
		if finding.Category == "overflow" || finding.Category == "clipping" {
			counts["overflow"]++
		}
		if finding.Category == "overlap" {
			counts["overlap"]++
		}
	}
	for _, item := range evidence {
		switch item.Status {
		case report.EvidenceStatusPass, report.EvidenceStatusFail, report.EvidenceStatusAdvisory,
			report.EvidenceStatusFalsePositive:
			continue
		case report.EvidenceStatusNotRun:
			counts["not-run"]++
		case report.EvidenceStatusDeferred:
			counts["deferred"]++
		}
	}
	budgets := map[string]int{
		"error": productPolicy.Budgets.Error, "warning": productPolicy.Budgets.Warning,
		"raw": productPolicy.Budgets.Raw, "overflow": productPolicy.Budgets.Overflow,
		"overlap":  productPolicy.Budgets.Overlap,
		"blocking": productPolicy.Budgets.Blocking, "advisory": productPolicy.Budgets.Advisory,
		"exception": productPolicy.Budgets.Exception, "not-run": productPolicy.Budgets.NotRun,
		"deferred": productPolicy.Budgets.Deferred,
	}
	for _, category := range []string{"error", "warning", "raw", "overflow", "overlap", "blocking", "advisory", "exception", "not-run", "deferred"} {
		if counts[category] <= budgets[category] {
			continue
		}
		if !productPolicy.RuleActiveAt(rules.RulePolicyBudgetExceeded, now) {
			continue
		}
		failed = true
		diagnostics = append(diagnostics, diagnostic.New(
			rules.RulePolicyBudgetExceeded,
			diagnostic.SeverityError,
			fmt.Sprintf("%s count %d exceeds budget %d", category, counts[category], budgets[category]),
			"<policy>",
			nil,
			diagnostic.EvidenceExecution,
			"all",
			"ansldes/policy",
			"budget",
		))
	}
	diagnostic.Sort(diagnostics)
	return diagnostics, failed
}

func languageForPath(path string) (source.Language, error) {
	switch filepath.Ext(path) {
	case ".tsx":
		return source.LanguageTSX, nil
	case ".ts":
		return source.LanguageTypeScript, nil
	default:
		return "", fmt.Errorf("unsupported source extension for %s", path)
	}
}
