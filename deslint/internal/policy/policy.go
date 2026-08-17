// Package policy parses and evaluates product-owned lint configuration.
package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

const (
	// ProfileOperate prioritizes task completion and familiar controls.
	ProfileOperate = "operate"
	// ProfileRead prioritizes sustained prose consumption.
	ProfileRead = "read"
	// ProfileBrowse prioritizes scanability and discovery.
	ProfileBrowse = "browse"
	// ProfileCreate prioritizes authoring and reversible editing.
	ProfileCreate = "create"
)

// BuiltinProfileIDs is the stable set of product-neutral profile identities.
var BuiltinProfileIDs = []string{ProfileOperate, ProfileRead, ProfileBrowse, ProfileCreate}

// Parse validates a product policy without importing product paths into the engine.
func Parse(contents []byte) (Policy, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Policy{}, fmt.Errorf("parse policy JSON: %w", err)
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Policy{}, fmt.Errorf("policy has duplicate keys: %s", strings.Join(duplicates, ", "))
	}

	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var result Policy
	if err := decoder.Decode(&result); err != nil {
		return Policy{}, fmt.Errorf("decode policy: %w", err)
	}
	if err := validate(result); err != nil {
		return Policy{}, err
	}
	return result, nil
}

// Severity returns the exact configured severity for a rule.
func (p Policy) Severity(ruleID string) diagnostic.Severity {
	if p.Profile != nil && p.Profile.SeverityOverrides[ruleID] == string(diagnostic.SeverityWarning) {
		return diagnostic.SeverityWarning
	}
	if p.Severities[ruleID] == string(diagnostic.SeverityWarning) {
		return diagnostic.SeverityWarning
	}
	return diagnostic.SeverityError
}

// RawPropertyKinds maps each configured style property to its raw-value category.
func (p Policy) RawPropertyKinds() map[string]string {
	result := make(map[string]string)
	for _, property := range p.Source.RawProperties.Color {
		result[property] = "color"
	}
	for _, property := range p.Source.RawProperties.Number {
		result[property] = "number"
	}
	for _, property := range p.Source.RawProperties.Motion {
		result[property] = "motion"
	}
	return result
}

// IsExcluded returns true only for an exact normalized path match.
func (p Policy) IsExcluded(path string) bool {
	normalized := filepath.ToSlash(filepath.Clean(path))
	for _, excluded := range p.Source.ExactExcludes {
		if normalized == filepath.ToSlash(filepath.Clean(excluded)) {
			return true
		}
	}
	return false
}

// Requires reports whether an independently acquired evidence kind is mandatory.
func (p Policy) Requires(kind diagnostic.EvidenceKind) bool {
	for _, configured := range p.Evidence.RequiredKinds {
		if canonicalEvidenceKind(configured) == string(kind) {
			return true
		}
	}
	if p.Profile != nil {
		for _, configured := range p.Profile.RequiredEvidence {
			if canonicalEvidenceKind(configured) == string(kind) {
				return true
			}
		}
	}
	return false
}

// RuleActive reports whether an exact canonical rule is enabled by governance policy.
func (p Policy) RuleActive(ruleID string) bool {
	return p.RuleActiveAt(ruleID, time.Time{})
}

// RuleActiveAt applies expiry when determining whether an exact rule is disabled.
func (p Policy) RuleActiveAt(ruleID string, now time.Time) bool {
	for _, override := range p.RuleOverrides {
		if override.RuleID == ruleID && override.Status == "disabled" {
			if !now.IsZero() {
				expiresAt, err := time.Parse("2006-01-02", override.ExpiresAt)
				if err == nil && expiresAt.Before(midnightUTC(now)) {
					return true
				}
			}
			return false
		}
	}
	return true
}

// RuleOverride returns the exact governance record for one canonical rule.
func (p Policy) RuleOverride(ruleID string) (RuleOverride, bool) {
	for _, override := range p.RuleOverrides {
		if override.RuleID == ruleID {
			return override, true
		}
	}
	return RuleOverride{}, false
}

// ProfileID returns the selected profile, or empty for detector-default behavior.
func (p Policy) ProfileID() string {
	if p.Profile == nil {
		return ""
	}
	return p.Profile.ID
}

// Deferred reports whether a runtime evidence kind was explicitly postponed.
func (p Policy) Deferred(kind diagnostic.EvidenceKind) bool {
	for _, configured := range p.Evidence.DeferredKinds {
		if canonicalEvidenceKind(configured) == string(kind) {
			return true
		}
	}
	return false
}

// ExceptionMatch preserves the exact policy record that classified a finding.
type ExceptionMatch struct {
	Finding   diagnostic.Diagnostic
	Exception Exception
}

// IgnoreRequest is one provider observation requesting an exact governed allowance.
type IgnoreRequest struct {
	Kind, RuleID, Engine, Platform, Path, Property, Value, Owner string
}

// AuthorizesIgnore only matches the same provider, platform, path, owner, and optional rule/value tuple.
func (p Policy) AuthorizesIgnore(request IgnoreRequest, now time.Time) bool {
	requestPath := filepath.ToSlash(filepath.Clean(request.Path))
	for _, allowance := range p.Governance.Ignores {
		expiresAt, err := time.Parse("2006-01-02", allowance.ExpiresAt)
		if err != nil || expiresAt.Before(midnightUTC(now)) {
			continue
		}
		if allowance.Kind == request.Kind && allowance.RuleID == request.RuleID && allowance.Engine == request.Engine && allowance.Platform == request.Platform && filepath.ToSlash(filepath.Clean(allowance.Path)) == requestPath && allowance.Property == request.Property && allowance.Value == request.Value && allowance.Owner == request.Owner {
			return true
		}
	}
	return false
}

// ClassifyExceptions separates exact active matches without discarding their evidence.
func (p Policy) ClassifyExceptions(diagnostics []diagnostic.Diagnostic, now time.Time) ([]diagnostic.Diagnostic, []ExceptionMatch) {
	result := make([]diagnostic.Diagnostic, 0, len(diagnostics))
	matches := make([]ExceptionMatch, 0)
	for _, finding := range diagnostics {
		exception, matched := p.matchingException(finding, now)
		if matched {
			matches = append(matches, ExceptionMatch{Finding: finding, Exception: exception})
			continue
		}
		result = append(result, finding)
	}
	diagnostic.Sort(result)
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Finding.Path != matches[j].Finding.Path {
			return matches[i].Finding.Path < matches[j].Finding.Path
		}
		if matches[i].Finding.RuleID != matches[j].Finding.RuleID {
			return matches[i].Finding.RuleID < matches[j].Finding.RuleID
		}
		return matches[i].Finding.Fingerprint < matches[j].Finding.Fingerprint
	})
	return result, matches
}

// ApplyExceptions returns only active findings for callers that do not render reports.
func (p Policy) ApplyExceptions(diagnostics []diagnostic.Diagnostic, now time.Time) []diagnostic.Diagnostic {
	result, _ := p.ClassifyExceptions(diagnostics, now)
	return result
}

// ExpiredExceptions returns deterministic policy entries that can no longer suppress findings.
func (p Policy) ExpiredExceptions(now time.Time) []Exception {
	result := make([]Exception, 0)
	for _, exception := range p.Exceptions {
		expiresAt, err := time.Parse("2006-01-02", exception.ExpiresAt)
		if err == nil && expiresAt.Before(midnightUTC(now)) {
			result = append(result, exception)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		return result[i].RuleID < result[j].RuleID
	})
	return result
}

// GovernanceReviewOverdue reports whether the pinned review interval elapsed.
func (p Policy) GovernanceReviewOverdue(now time.Time) bool {
	reviewedAt, err := time.Parse("2006-01-02", p.Governance.ReviewedAt)
	if err != nil || p.Governance.ReviewIntervalDays < 1 {
		return true
	}
	return reviewedAt.AddDate(0, 0, p.Governance.ReviewIntervalDays).Before(midnightUTC(now))
}

// ExpiredIgnores returns deterministic governance allowances that must fail review.
func (p Policy) ExpiredIgnores(now time.Time) []IgnoreAllowance {
	result := []IgnoreAllowance{}
	for _, allowance := range p.Governance.Ignores {
		expiresAt, err := time.Parse("2006-01-02", allowance.ExpiresAt)
		if err == nil && expiresAt.Before(midnightUTC(now)) {
			result = append(result, allowance)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		return result[i].Kind < result[j].Kind
	})
	return result
}

func (p Policy) matchingException(finding diagnostic.Diagnostic, now time.Time) (Exception, bool) {
	for _, exception := range p.Exceptions {
		expiresAt, err := time.Parse("2006-01-02", exception.ExpiresAt)
		if err != nil || expiresAt.Before(midnightUTC(now)) {
			continue
		}
		if exception.RuleID == finding.RuleID && exception.Engine == string(finding.EvidenceKind) && exception.Platform == finding.Platform && exception.Owner == finding.Owner && filepath.ToSlash(filepath.Clean(exception.Path)) == filepath.ToSlash(filepath.Clean(finding.Path)) {
			return exception, true
		}
	}
	return Exception{}, false
}

func validate(value Policy) error {
	if value.SchemaVersion != SchemaVersion {
		return fmt.Errorf("policy schemaVersion = %d, want %d", value.SchemaVersion, SchemaVersion)
	}
	if value.DefinitionID == "" {
		return fmt.Errorf("policy definitionId is required")
	}
	expectedRules := append([]string(nil), rules.ConfigurableRuleIDs...)
	sort.Strings(expectedRules)
	actualRules := make([]string, 0, len(value.Severities))
	for ruleID, severity := range value.Severities {
		if severity != string(diagnostic.SeverityError) && severity != string(diagnostic.SeverityWarning) {
			return fmt.Errorf("policy severity %s = %q", ruleID, severity)
		}
		actualRules = append(actualRules, ruleID)
	}
	sort.Strings(actualRules)
	if strings.Join(expectedRules, "\x00") != strings.Join(actualRules, "\x00") {
		return fmt.Errorf("policy severities must exactly match the v1 rule registry")
	}
	for _, excluded := range value.Source.ExactExcludes {
		cleaned := filepath.ToSlash(filepath.Clean(excluded))
		if excluded == "" || filepath.IsAbs(excluded) || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.ContainsAny(excluded, "*?[]") {
			return fmt.Errorf("policy exactExcludes entry %q is not an exact relative path", excluded)
		}
	}
	seenProperties := make(map[string]string)
	for category, properties := range map[string][]string{
		"color":  value.Source.RawProperties.Color,
		"number": value.Source.RawProperties.Number,
		"motion": value.Source.RawProperties.Motion,
	} {
		for _, property := range properties {
			if property == "" {
				return fmt.Errorf("policy raw property names must not be empty")
			}
			if previous, exists := seenProperties[property]; exists {
				return fmt.Errorf("policy raw property %q is duplicated across %s and %s", property, previous, category)
			}
			seenProperties[property] = category
		}
	}
	if value.Budgets.Error < 0 || value.Budgets.Warning < 0 || value.Budgets.Raw < 0 || value.Budgets.Overflow < 0 || value.Budgets.Overlap < 0 || value.Budgets.Blocking < 0 || value.Budgets.Advisory < 0 || value.Budgets.Exception < 0 || value.Budgets.NotRun < 0 || value.Budgets.Deferred < 0 {
		return fmt.Errorf("policy budgets must be non-negative")
	}
	validEvidence := map[string]bool{
		string(diagnostic.EvidenceDefinition):              true,
		string(diagnostic.EvidenceWebSource):               true,
		string(diagnostic.EvidenceWebRendered):             true,
		string(diagnostic.EvidenceNativeSource):            true,
		string(diagnostic.EvidenceDesignDocumentSource):    true,
		string(diagnostic.EvidenceDesignDocumentComputed):  true,
		string(diagnostic.EvidenceSimulator):               true,
		string(diagnostic.EvidenceEmulator):                true,
		string(diagnostic.EvidencePhysicalDevice):          true,
		string(diagnostic.EvidenceConsumerConformance):     true,
		string(diagnostic.EvidenceConsumerContentRegistry): true,
		"source":          true,
		"pencil":          true,
		"computed-layout": true,
	}
	seenEvidence := make(map[string]bool)
	for _, kind := range value.Evidence.RequiredKinds {
		canonical := canonicalEvidenceKind(kind)
		if !validEvidence[kind] || seenEvidence[canonical] {
			return fmt.Errorf("policy required evidence %q is invalid or duplicated", kind)
		}
		seenEvidence[canonical] = true
	}
	for _, kind := range value.Evidence.DeferredKinds {
		if !validEvidence[kind] || kind == string(diagnostic.EvidenceDefinition) ||
			kind == "source" || kind == "pencil" || kind == "computed-layout" {
			return fmt.Errorf("policy deferred evidence %q is invalid", kind)
		}
		if seenEvidence[kind] {
			return fmt.Errorf("policy evidence %q cannot be both required and deferred", kind)
		}
		seenEvidence[kind] = true
	}
	if value.Profile != nil {
		if err := validateProfile(*value.Profile, validEvidence); err != nil {
			return err
		}
		for _, kind := range value.Profile.RequiredEvidence {
			if value.Deferred(diagnostic.EvidenceKind(canonicalEvidenceKind(kind))) {
				return fmt.Errorf("policy evidence %q cannot be both profile-required and deferred", kind)
			}
		}
	}
	if value.Content != nil {
		if value.Content.RegistryVersion == "" || len(value.Content.Locales) == 0 {
			return fmt.Errorf("policy content registryVersion and locales are required")
		}
		if err := validateUniqueContentValues("sourceReferences", value.Content.SourceReferences); err != nil {
			return err
		}
		for locale, content := range value.Content.Locales {
			if locale == "" || content.PhraseRegistryVersion == "" {
				return fmt.Errorf("policy content locale and phraseRegistryVersion are required")
			}
			for name, values := range map[string][]string{
				"marketingBuzzwords": content.MarketingBuzzwords,
				"theaterPhrases":     content.TheaterPhrases,
				"protectedTerms":     content.ProtectedTerms,
				"recoveryCopyIds":    content.RecoveryCopyIDs,
			} {
				if err := validateUniqueContentValues(locale+"."+name, values); err != nil {
					return err
				}
			}
		}
	}
	if value.Assets != nil {
		if value.Assets.RegistryVersion == "" {
			return fmt.Errorf("policy asset registryVersion is required")
		}
		for id, asset := range value.Assets.Entries {
			if id == "" || asset.Owner == "" || asset.ImplementationSource == "" || len(asset.FingerprintSHA256) != 64 ||
				!slices.Contains([]string{"icon", "logo", "data-diagram", "hero-illustration", "photo", "video-poster"}, asset.Role) {
				return fmt.Errorf("policy asset %q is incomplete", id)
			}
			if err := validateUniqueContentValues("asset "+id+" consumers", asset.Consumers); err != nil {
				return err
			}
		}
	}
	if value.Runtime != nil {
		if value.Runtime.RegistryVersion == "" {
			return fmt.Errorf("policy runtime registryVersion is required")
		}
		seenRuntimeExceptions := map[string]bool{}
		for _, exception := range value.Runtime.JustifiedTextExceptions {
			if exception.Platform == "" || exception.SurfaceID == "" || exception.RouteID == "" || exception.NodeID == "" || exception.Owner == "" ||
				!slices.Contains([]string{"web", "ios", "android"}, exception.Platform) || !slices.Contains([]string{"print", "export"}, exception.Context) {
				return fmt.Errorf("policy justified text exception is incomplete")
			}
			identity := strings.Join([]string{exception.Platform, exception.SurfaceID, exception.RouteID, exception.NodeID, exception.Owner, exception.Context}, "\x00")
			if seenRuntimeExceptions[identity] {
				return fmt.Errorf("policy justified text exception %q is duplicated", exception.NodeID)
			}
			seenRuntimeExceptions[identity] = true
		}
	}
	if value.Native != nil {
		thresholds := value.Native.Thresholds
		if value.Native.RegistryVersion == "" || value.Native.IOSAdjacentTargetSpacing < 0 || thresholds.MaxSynchronousStartupMS < 0 || thresholds.MaxInitializationMS < 0 || thresholds.MaxMainThreadWorkMS < 0 || thresholds.MaxFrameDropRatio < 0 || thresholds.MaxFrameDropRatio > 1 || thresholds.MaxThumbnailDecodeRatio < 1 || thresholds.MaxJSBundleRegressionBytes < 0 || thresholds.MaxAppBinaryRegressionBytes < 0 {
			return fmt.Errorf("policy native registry or thresholds are invalid")
		}
		seenNativeCaptures := map[string]bool{}
		for _, capture := range value.Native.RequiredRuntimeCaptures {
			kindMatches := capture.EvidenceKind == string(diagnostic.EvidenceSimulator) && capture.Platform == "ios" || capture.EvidenceKind == string(diagnostic.EvidenceEmulator) && capture.Platform == "android" || capture.EvidenceKind == string(diagnostic.EvidencePhysicalDevice) && (capture.Platform == "ios" || capture.Platform == "android")
			if capture.ID == "" || seenNativeCaptures[capture.ID] || !kindMatches || capture.MinimumFontScale < 1 ||
				!slices.Contains([]string{"phone", "tablet", "foldable"}, capture.FormFactor) || !slices.Contains([]string{"portrait", "landscape"}, capture.Orientation) || !slices.Contains([]string{"fullscreen", "split", "multi-window"}, capture.WindowMode) || !slices.Contains([]string{"not-applicable", "flat", "half-open"}, capture.FoldPosture) || !slices.Contains([]string{"light", "dark"}, capture.Theme) {
				return fmt.Errorf("policy native runtime capture %q is invalid or duplicated", capture.ID)
			}
			if capture.Platform == "ios" && capture.FormFactor == "foldable" || capture.FormFactor != "foldable" && capture.FoldPosture != "not-applicable" {
				return fmt.Errorf("policy native runtime capture %q has incompatible fold posture", capture.ID)
			}
			seenNativeCaptures[capture.ID] = true
		}
	}
	if value.Web != nil {
		if value.Web.RegistryVersion == "" || strings.TrimSpace(value.Web.BuildCommand) == "" {
			return fmt.Errorf("policy Web registryVersion and buildCommand are required")
		}
		routes := make(map[string]bool, len(value.Web.Routes))
		for _, route := range value.Web.Routes {
			if route.ID == "" || route.Owner == "" || strings.TrimSpace(route.Target) == "" || routes[route.ID] {
				return fmt.Errorf("policy Web route %q is incomplete or duplicated", route.ID)
			}
			routes[route.ID] = true
		}
		viewports := make(map[string]bool, len(value.Web.Viewports))
		for _, viewport := range value.Web.Viewports {
			if viewport.ID == "" || viewport.Width <= 0 || viewport.Height <= 0 || viewports[viewport.ID] {
				return fmt.Errorf("policy Web viewport %q is invalid or duplicated", viewport.ID)
			}
			viewports[viewport.ID] = true
		}
		if len(routes) == 0 || len(viewports) == 0 || len(value.Web.Themes) == 0 || len(value.Web.FontScales) == 0 || len(value.Web.ReduceMotion) == 0 {
			return fmt.Errorf("policy Web routes and render axes must not be empty")
		}
		if err := validateUniqueContentValues("Web themes", value.Web.Themes); err != nil {
			return err
		}
		fontScales := make(map[float64]bool, len(value.Web.FontScales))
		for _, scale := range value.Web.FontScales {
			if scale < 1 || fontScales[scale] {
				return fmt.Errorf("policy Web font scale %v is invalid or duplicated", scale)
			}
			fontScales[scale] = true
		}
		reduceMotion := make(map[bool]bool, len(value.Web.ReduceMotion))
		for _, reduced := range value.Web.ReduceMotion {
			if reduceMotion[reduced] {
				return fmt.Errorf("policy Web Reduce Motion axis is duplicated")
			}
			reduceMotion[reduced] = true
		}
		themes := make(map[string]bool, len(value.Web.Themes))
		for _, theme := range value.Web.Themes {
			themes[theme] = true
		}
		seenCaptures := make(map[string]bool, len(value.Web.RequiredCaptures))
		for _, capture := range value.Web.RequiredCaptures {
			if capture.ID == "" || seenCaptures[capture.ID] || !slices.Contains([]string{"regex-source", "static-html", "browser", "visual-contrast"}, capture.Provider) ||
				!routes[capture.RouteID] || !viewports[capture.ViewportID] || !themes[capture.Theme] || !fontScales[capture.FontScale] || !reduceMotion[capture.ReduceMotion] {
				return fmt.Errorf("policy Web capture %q is invalid, duplicated, or references an undeclared axis", capture.ID)
			}
			seenCaptures[capture.ID] = true
		}
		seenArtifactExclusions := make(map[string]bool, len(value.Web.ArtifactExclusions))
		for _, exclusion := range value.Web.ArtifactExclusions {
			cleaned := filepath.ToSlash(filepath.Clean(exclusion.Path))
			identity := cleaned + "\x00" + exclusion.FingerprintSHA256
			if !isExactRelativePath(exclusion.Path) || len(exclusion.FingerprintSHA256) != 64 || exclusion.Owner == "" || len(exclusion.Rationale) < 8 || strings.TrimSpace(exclusion.ReproductionCommand) == "" || seenArtifactExclusions[identity] {
				return fmt.Errorf("policy Web artifact exclusion %q is incomplete, broad, or duplicated", exclusion.Path)
			}
			seenArtifactExclusions[identity] = true
		}
	}
	for _, exception := range value.Exceptions {
		if exception.RuleID == "" || !validGovernanceEngine(exception.Engine) || exception.Platform == "" || !isExactRelativePath(exception.Path) || exception.Owner == "" || len(exception.Rationale) < 8 || exception.Reviewer == "" || len(exception.ReviewTrigger) < 8 {
			return fmt.Errorf("policy exception must have an exact rule, engine, platform, path, owner, rationale, reviewer, and reviewTrigger")
		}
		if _, err := time.Parse("2006-01-02", exception.ExpiresAt); err != nil {
			return fmt.Errorf("policy exception expiry %q is invalid", exception.ExpiresAt)
		}
		if !slices.Contains(expectedRules, exception.RuleID) {
			return fmt.Errorf("policy exception ruleId %q is not configurable", exception.RuleID)
		}
	}
	if len(value.RulePacks) == 0 {
		return fmt.Errorf("policy must declare at least one required rule pack")
	}
	seenPacks := make(map[string]bool)
	selectedRules := make(map[string]bool)
	selectedRulePacks := make(map[string]string)
	for _, requirement := range value.RulePacks {
		key := requirement.ID + "@" + requirement.Version
		pack, found := rules.LookupPack(requirement.ID, requirement.Version)
		if !found || requirement.FingerprintSHA256 != rules.PackFingerprint(pack) {
			return fmt.Errorf("policy rule pack %s does not match a registered exact version and fingerprint", key)
		}
		if seenPacks[key] {
			return fmt.Errorf("policy rule pack %s is duplicated", key)
		}
		seenPacks[key] = true
		for _, spec := range pack.Rules {
			selectedRules[spec.ID] = true
			selectedRulePacks[spec.ID] = key
		}
	}
	for _, required := range []string{rules.AntiSlopPackID + "@" + rules.AntiSlopPackVersion, rules.CorePackID + "@" + rules.CorePackVersion} {
		if !seenPacks[required] {
			return fmt.Errorf("policy governance requires exact built-in rule pack %s", required)
		}
	}
	seenOverrides := make(map[string]bool)
	for _, override := range value.RuleOverrides {
		packKey := override.PackID + "@" + override.PackVersion
		if !selectedRules[override.RuleID] || selectedRulePacks[override.RuleID] != packKey || override.Status != "disabled" {
			return fmt.Errorf("policy rule override %q must target one canonical rule with disabled status", override.RuleID)
		}
		if seenOverrides[override.RuleID] {
			return fmt.Errorf("policy rule override %q is duplicated", override.RuleID)
		}
		seenOverrides[override.RuleID] = true
		if override.Owner == "" || len(override.Rationale) < 8 || override.Reviewer == "" || len(override.ReviewTrigger) < 8 {
			return fmt.Errorf("policy rule override %q requires owner, rationale, reviewer, and reviewTrigger", override.RuleID)
		}
		if _, err := time.Parse("2006-01-02", override.ExpiresAt); err != nil {
			return fmt.Errorf("policy rule override expiry %q is invalid", override.ExpiresAt)
		}
	}
	if err := validateGovernance(value, selectedRules); err != nil {
		return err
	}
	return nil
}

func validateGovernance(value Policy, selectedRules map[string]bool) error {
	governance := value.Governance
	if _, err := time.Parse("2006-01-02", governance.ReviewedAt); err != nil || governance.ReviewIntervalDays < 1 || governance.ReviewIntervalDays > 90 || governance.Reviewer == "" || governance.AdvisoryMode != "report" || !governance.RequireExitCode2 || !governance.RequireUnmodifiedReport || !governance.PassingReportsOnly {
		return fmt.Errorf("policy governance review and anti-bypass contract is invalid")
	}
	expectedFlags := []string{"--no-advisory", "--no-config", "--no-design-system", "--no-inline-ignores"}
	actualFlags := append([]string(nil), governance.ForbiddenFlags...)
	sort.Strings(actualFlags)
	if !slices.Equal(actualFlags, expectedFlags) {
		return fmt.Errorf("policy governance forbidden flags must be the exact anti-bypass set")
	}
	expectedSubjects := []string{"exceptions", "hallmark-commit", "impeccable-version", "rule-mapping-drift"}
	actualSubjects := append([]string(nil), governance.ReviewSubjects...)
	sort.Strings(actualSubjects)
	if !slices.Equal(actualSubjects, expectedSubjects) {
		return fmt.Errorf("policy governance review subjects must cover upstream, mapping drift, and exceptions")
	}
	seen := map[string]bool{}
	fileAllowances := map[string]bool{}
	for _, allowance := range governance.Ignores {
		if !slices.Contains([]string{"rule", "file", "value", "inline"}, allowance.Kind) || !validGovernanceEngine(allowance.Engine) || allowance.Platform == "" || !isExactRelativePath(allowance.Path) || allowance.Owner == "" || len(allowance.Rationale) < 8 || allowance.Reviewer == "" || len(allowance.ReviewTrigger) < 8 {
			return fmt.Errorf("policy governance %s ignore is incomplete or broad", allowance.Kind)
		}
		if _, err := time.Parse("2006-01-02", allowance.ExpiresAt); err != nil {
			return fmt.Errorf("policy governance ignore expiry %q is invalid", allowance.ExpiresAt)
		}
		if (allowance.Kind == "rule" || allowance.Kind == "inline") && !selectedRules[allowance.RuleID] {
			return fmt.Errorf("policy governance ignore rule %q is unknown or unselected", allowance.RuleID)
		}
		if allowance.Kind == "value" && (allowance.Value == "" || allowance.Property == "") {
			return fmt.Errorf("policy governance value ignore requires an exact property, value, and owner")
		}
		if allowance.Kind != "value" && (allowance.Value != "" || allowance.Property != "") {
			return fmt.Errorf("policy governance %s ignore cannot carry value fields", allowance.Kind)
		}
		identity := strings.Join([]string{allowance.Kind, allowance.RuleID, allowance.Engine, allowance.Platform, filepath.ToSlash(filepath.Clean(allowance.Path)), allowance.Property, allowance.Value, allowance.Owner}, "\x00")
		if seen[identity] {
			return fmt.Errorf("policy governance ignore is duplicated")
		}
		seen[identity] = true
		if allowance.Kind == "file" {
			fileAllowances[filepath.ToSlash(filepath.Clean(allowance.Path))] = true
		}
	}
	for _, excluded := range value.Source.ExactExcludes {
		if !fileAllowances[filepath.ToSlash(filepath.Clean(excluded))] {
			return fmt.Errorf("policy exactExcludes entry %q lacks an owned governance file allowance", excluded)
		}
	}
	return nil
}

func validGovernanceEngine(value string) bool {
	return slices.Contains([]string{
		string(diagnostic.EvidenceDefinition), string(diagnostic.EvidenceWebSource), string(diagnostic.EvidenceWebRendered),
		string(diagnostic.EvidenceNativeSource), string(diagnostic.EvidenceDesignDocumentSource), string(diagnostic.EvidenceDesignDocumentComputed),
		string(diagnostic.EvidenceSimulator), string(diagnostic.EvidenceEmulator), string(diagnostic.EvidencePhysicalDevice),
		string(diagnostic.EvidenceConsumerConformance), string(diagnostic.EvidenceConsumerContentRegistry), string(diagnostic.EvidenceExecution),
	}, value)
}

func isExactRelativePath(path string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	return path != "" && !filepath.IsAbs(path) && cleaned != ".." && !strings.HasPrefix(cleaned, "../") && !strings.ContainsAny(path, "*?[]")
}

func validateUniqueContentValues(name string, values []string) error {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" || seen[value] {
			return fmt.Errorf("policy content %s contains an empty or duplicate value", name)
		}
		seen[value] = true
	}
	return nil
}

func validateProfile(profile ConsumerProfile, validEvidence map[string]bool) error {
	builtIn := slices.Contains(BuiltinProfileIDs, profile.ID)
	if profile.ID == "" || profile.PrimaryUserGoal == "" {
		return fmt.Errorf("policy profile id and primaryUserGoal are required")
	}
	if !builtIn && (len(profile.Rationale) < 8 || profile.Reviewer == "" || profile.EvidenceOwner == "") {
		return fmt.Errorf("custom policy profile %q requires rationale, reviewer, and evidenceOwner", profile.ID)
	}
	if !slices.Contains([]string{"compact", "comfortable", "spacious"}, profile.Density) ||
		!slices.Contains([]string{"low", "medium", "high"}, profile.NoveltyTolerance) ||
		!slices.Contains([]string{"required", "preferred", "neutral"}, profile.NativeAffordancePriority) {
		return fmt.Errorf("policy profile %q has invalid density, novelty tolerance, or native-affordance priority", profile.ID)
	}
	if profile.Thresholds.MaxOversizedActions < 0 || profile.Thresholds.MaxInconsistentActions < 0 {
		return fmt.Errorf("policy profile thresholds must be non-negative")
	}
	for ruleID, severity := range profile.SeverityOverrides {
		if !slices.Contains(rules.AllRuleIDs, ruleID) ||
			(severity != string(diagnostic.SeverityError) && severity != string(diagnostic.SeverityWarning)) {
			return fmt.Errorf("policy profile severity override %q is not a canonical rule or severity", ruleID)
		}
	}
	seen := make(map[string]bool)
	for _, kind := range profile.RequiredEvidence {
		canonical := canonicalEvidenceKind(kind)
		if !validEvidence[kind] || seen[canonical] {
			return fmt.Errorf("policy profile required evidence %q is invalid or duplicated", kind)
		}
		seen[canonical] = true
	}
	return nil
}

func midnightUTC(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func canonicalEvidenceKind(kind string) string {
	switch kind {
	case "source":
		return string(diagnostic.EvidenceNativeSource)
	case "pencil":
		return string(diagnostic.EvidenceDesignDocumentSource)
	case "computed-layout":
		return string(diagnostic.EvidenceDesignDocumentComputed)
	default:
		return kind
	}
}
