// Package webcheck normalizes pinned Impeccable provider output into canonical diagnostics.
package webcheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Evidence is one source, static HTML, browser, or visual-contrast provider result.
type Evidence struct {
	Schema           string                  `json:"$schema,omitempty"`
	SchemaVersion    int                     `json:"schemaVersion"`
	EvidenceKind     diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform         string                  `json:"platform"`
	Provider         string                  `json:"provider"`
	SurfaceID        string                  `json:"surfaceId"`
	WebPolicyVersion string                  `json:"webPolicyVersion"`
	CaptureID        string                  `json:"captureId"`
	RouteID          string                  `json:"routeId"`
	Owner            string                  `json:"owner"`
	Viewport         Viewport                `json:"viewport"`
	Theme            string                  `json:"theme"`
	FontScale        float64                 `json:"fontScale"`
	ReduceMotion     bool                    `json:"reduceMotion"`
	Execution        Execution               `json:"execution"`
	Findings         []ProviderFinding       `json:"findings"`
}

// Viewport is one exact consumer policy viewport.
type Viewport struct {
	ID     string `json:"id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Execution separates completion, not-run, fallback, and provider failure.
type Execution struct {
	Status     string `json:"status"`
	Capability string `json:"capability"`
	Reason     string `json:"reason"`
}

// ProviderFinding is one pinned upstream rule observation.
type ProviderFinding struct {
	ID             string    `json:"id"`
	UpstreamRuleID string    `json:"upstreamRuleId"`
	Owner          string    `json:"owner"`
	Path           string    `json:"path"`
	Message        string    `json:"message"`
	Artifact       *Artifact `json:"artifact"`
}

// Artifact identifies one generated output byte-for-byte.
type Artifact struct {
	Path              string `json:"path"`
	FingerprintSHA256 string `json:"fingerprintSha256"`
	Generated         bool   `json:"generated"`
}

// Route is one consumer-owned provider target.
type Route struct {
	Owner  string
	Target string
}

// CaptureRequirement is one exact provider/route/render-axis requirement.
type CaptureRequirement struct {
	ID           string
	Provider     string
	RouteID      string
	ViewportID   string
	Theme        string
	FontScale    float64
	ReduceMotion bool
}

// ArtifactExclusion is one exact generated output exception with reproduction evidence.
type ArtifactExclusion struct {
	Path                string
	FingerprintSHA256   string
	Owner               string
	Rationale           string
	ReproductionCommand string
}

// ExcludedFinding preserves an exact artifact match for false-positive reporting.
type ExcludedFinding struct {
	Finding   diagnostic.Diagnostic
	Exclusion ArtifactExclusion
}

// ProviderExecutionError distinguishes provider failure from an application finding.
type ProviderExecutionError struct {
	Provider  string
	CaptureID string
	Reason    string
}

func (e *ProviderExecutionError) Error() string {
	return fmt.Sprintf("web provider %q failed for capture %q: %s", e.Provider, e.CaptureID, e.Reason)
}

// Config injects consumer-owned routes, axes, and exact artifact exclusions.
type Config struct {
	RegistryVersion    string
	Routes             map[string]Route
	Viewports          map[string]Viewport
	Themes             []string
	FontScales         []float64
	ReduceMotion       []bool
	RequiredCaptures   []CaptureRequirement
	ArtifactExclusions []ArtifactExclusion
	Severity           func(string) diagnostic.Severity
	Active             func(string) bool
}

// Analyze parses one provider result and returns canonical findings plus exact artifact exclusions.
func Analyze(_ string, contents []byte, config Config) (Evidence, []diagnostic.Diagnostic, []ExcludedFinding, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, nil, err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, nil, fmt.Errorf("web provider evidence has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, nil, nil, err
	}
	if config.Severity == nil {
		config.Severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if config.Active == nil {
		config.Active = func(string) bool { return true }
	}
	if err := validateEvidence(evidence, config); err != nil {
		return Evidence{}, nil, nil, err
	}
	if evidence.Execution.Status == "failed" {
		return evidence, nil, nil, &ProviderExecutionError{Provider: evidence.Provider, CaptureID: evidence.CaptureID, Reason: evidence.Execution.Reason}
	}
	if evidence.Execution.Status == "not-run" {
		return evidence, []diagnostic.Diagnostic{}, []ExcludedFinding{}, nil
	}

	providerFindings := append([]ProviderFinding(nil), evidence.Findings...)
	sort.SliceStable(providerFindings, func(i, j int) bool { return providerFindings[i].ID < providerFindings[j].ID })
	seen := map[string]bool{}
	findings := []diagnostic.Diagnostic{}
	excluded := []ExcludedFinding{}
	for _, providerFinding := range providerFindings {
		if providerFinding.ID == "" || providerFinding.UpstreamRuleID == "" || providerFinding.Owner == "" || providerFinding.Path == "" || providerFinding.Message == "" || seen[providerFinding.ID] {
			return Evidence{}, nil, nil, fmt.Errorf("web provider finding identity is incomplete or duplicated")
		}
		seen[providerFinding.ID] = true
		catalogRule, found := rules.LookupSourceRule("impeccable/" + providerFinding.UpstreamRuleID)
		if !found {
			return Evidence{}, nil, nil, fmt.Errorf("web provider emitted unknown upstream rule %q", providerFinding.UpstreamRuleID)
		}
		if !slices.Contains(catalogRule.Providers, evidence.Provider) || !slices.Contains(catalogRule.EvidenceKinds, string(evidence.EvidenceKind)) {
			return Evidence{}, nil, nil, fmt.Errorf("web provider %q is incompatible with rule %q", evidence.Provider, providerFinding.UpstreamRuleID)
		}
		if !config.Active(catalogRule.ID) {
			continue
		}
		severity := config.Severity(catalogRule.ID)
		if catalogRule.DefaultSeverity == "advisory" {
			severity = diagnostic.SeverityWarning
		}
		finding := diagnostic.NewWithSources(catalogRule.ID, catalogRule.SourceRuleIDs, severity, providerFinding.Message, providerFinding.Path, nil, evidence.EvidenceKind, evidence.Platform, providerFinding.Owner, catalogRule.Category)
		finding = diagnostic.WithViewport(finding, fmt.Sprintf("%s:%dx%d", evidence.Viewport.ID, evidence.Viewport.Width, evidence.Viewport.Height))
		if exclusion, matched := exactArtifactExclusion(providerFinding.Artifact, config.ArtifactExclusions); matched {
			excluded = append(excluded, ExcludedFinding{Finding: finding, Exclusion: exclusion})
			continue
		}
		findings = append(findings, finding)
	}
	diagnostic.Sort(findings)
	sort.SliceStable(excluded, func(i, j int) bool { return excluded[i].Finding.Fingerprint < excluded[j].Finding.Fingerprint })
	return evidence, diagnostic.MergeCanonical(findings), excluded, nil
}

// CoverageFindings returns missing-evidence diagnostics for required captures not completed on exact axes.
func CoverageFindings(evidences []Evidence, config Config) []diagnostic.Diagnostic {
	if config.Severity == nil {
		config.Severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if config.Active == nil {
		config.Active = func(string) bool { return true }
	}
	if !config.Active(rules.RuleEvidenceMissing) {
		return []diagnostic.Diagnostic{}
	}
	requirements := append([]CaptureRequirement(nil), config.RequiredCaptures...)
	sort.SliceStable(requirements, func(i, j int) bool { return requirements[i].ID < requirements[j].ID })
	findings := []diagnostic.Diagnostic{}
	for _, requirement := range requirements {
		matched := false
		for _, evidence := range evidences {
			if evidence.CaptureID == requirement.ID && evidence.Provider == requirement.Provider && evidence.RouteID == requirement.RouteID && evidence.Viewport.ID == requirement.ViewportID && evidence.Theme == requirement.Theme && evidence.FontScale == requirement.FontScale && evidence.ReduceMotion == requirement.ReduceMotion && evidence.Execution.Status == "completed" && evidence.Execution.Capability == "full" {
				matched = true
				break
			}
		}
		if !matched {
			kind := diagnostic.EvidenceWebSource
			if requirement.Provider == "browser" || requirement.Provider == "visual-contrast" {
				kind = diagnostic.EvidenceWebRendered
			}
			findings = append(findings, diagnostic.New(rules.RuleEvidenceMissing, config.Severity(rules.RuleEvidenceMissing), fmt.Sprintf("required Web provider capture %s was not completed with full capability", requirement.ID), "<web-provider>#/"+requirement.ID, nil, kind, "web", "ansldes/evidence", "missing"))
		}
	}
	diagnostic.Sort(findings)
	return findings
}

func validateEvidence(evidence Evidence, config Config) error {
	if evidence.SchemaVersion != 1 || evidence.Platform != "web" || evidence.SurfaceID == "" || evidence.CaptureID == "" || evidence.RouteID == "" || evidence.Owner == "" {
		return fmt.Errorf("web provider evidence identity is invalid")
	}
	if config.RegistryVersion == "" || evidence.WebPolicyVersion != config.RegistryVersion {
		return fmt.Errorf("web evidence policy version %q does not match consumer policy %q", evidence.WebPolicyVersion, config.RegistryVersion)
	}
	if !providerMatchesKind(evidence.Provider, evidence.EvidenceKind) {
		return fmt.Errorf("web provider %q is incompatible with evidence kind %q", evidence.Provider, evidence.EvidenceKind)
	}
	route, found := config.Routes[evidence.RouteID]
	if !found || route.Owner != evidence.Owner || route.Target == "" {
		return fmt.Errorf("web evidence route %q does not exact-match consumer policy", evidence.RouteID)
	}
	viewport, found := config.Viewports[evidence.Viewport.ID]
	if !found || viewport != evidence.Viewport || !slices.Contains(config.Themes, evidence.Theme) || !slices.Contains(config.FontScales, evidence.FontScale) || !slices.Contains(config.ReduceMotion, evidence.ReduceMotion) {
		return fmt.Errorf("web evidence viewport, theme, font scale, or Reduce Motion axis drifted")
	}
	if !slices.Contains([]string{"completed", "not-run", "failed"}, evidence.Execution.Status) || !slices.Contains([]string{"full", "regex-fallback"}, evidence.Execution.Capability) {
		return fmt.Errorf("web provider execution status or capability is invalid")
	}
	if evidence.Execution.Status == "completed" && (evidence.Execution.Capability != "full" || evidence.Execution.Reason != "") {
		return fmt.Errorf("completed Web provider evidence requires full capability and no failure reason")
	}
	if evidence.Execution.Status != "completed" && (evidence.Execution.Reason == "" || len(evidence.Findings) != 0) {
		return fmt.Errorf("non-completed Web provider evidence requires a reason and no findings")
	}
	return nil
}

func providerMatchesKind(provider string, kind diagnostic.EvidenceKind) bool {
	return (provider == "regex-source" || provider == "static-html") && kind == diagnostic.EvidenceWebSource || (provider == "browser" || provider == "visual-contrast") && kind == diagnostic.EvidenceWebRendered
}

func exactArtifactExclusion(artifact *Artifact, exclusions []ArtifactExclusion) (ArtifactExclusion, bool) {
	if artifact == nil || !artifact.Generated || len(artifact.FingerprintSHA256) != 64 {
		return ArtifactExclusion{}, false
	}
	artifactPath := filepath.ToSlash(filepath.Clean(artifact.Path))
	for _, exclusion := range exclusions {
		if filepath.ToSlash(filepath.Clean(exclusion.Path)) == artifactPath && exclusion.FingerprintSHA256 == artifact.FingerprintSHA256 && exclusion.Owner != "" && exclusion.Rationale != "" && exclusion.ReproductionCommand != "" {
			return exclusion, true
		}
	}
	return ArtifactExclusion{}, false
}
