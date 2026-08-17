// Package diagnostic defines the stable external finding contract.
package diagnostic

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"sort"
	"strconv"
)

// Severity is the configured impact of a diagnostic.
type Severity string

const (
	// SeverityError fails the lint run.
	SeverityError Severity = "error"
	// SeverityWarning is counted separately and may fail a zero budget.
	SeverityWarning Severity = "warning"
)

// EvidenceKind identifies which independently acquired input produced a finding.
type EvidenceKind string

const (
	// EvidenceDefinition is the product design-system definition.
	EvidenceDefinition EvidenceKind = "definition"
	// EvidenceWebSource is static Web source evidence.
	EvidenceWebSource EvidenceKind = "web-source"
	// EvidenceWebRendered is rendered browser evidence.
	EvidenceWebRendered EvidenceKind = "web-rendered"
	// EvidenceNativeSource is native or React Native source evidence.
	EvidenceNativeSource EvidenceKind = "native-source"
	// EvidenceDesignDocumentSource is serialized design-document evidence.
	EvidenceDesignDocumentSource EvidenceKind = "design-document-source"
	// EvidenceDesignDocumentComputed is computed design-document layout evidence.
	EvidenceDesignDocumentComputed EvidenceKind = "design-document-computed"
	// EvidenceSimulator is iOS simulator runtime evidence.
	EvidenceSimulator EvidenceKind = "simulator"
	// EvidenceEmulator is Android emulator runtime evidence.
	EvidenceEmulator EvidenceKind = "emulator"
	// EvidencePhysicalDevice is physical-device runtime evidence.
	EvidencePhysicalDevice EvidenceKind = "physical-device"
	// EvidenceConsumerConformance is a consumer-owned component conformance inventory.
	EvidenceConsumerConformance EvidenceKind = "consumer-conformance"
	// EvidenceConsumerContentRegistry is the independently executed claim and phrase registry.
	EvidenceConsumerContentRegistry EvidenceKind = "consumer-content-registry"
	// EvidenceExecution is a missing, malformed, or stale input.
	EvidenceExecution EvidenceKind = "execution"

	// EvidenceSource is the legacy Go name for native source evidence.
	EvidenceSource = EvidenceNativeSource
	// EvidencePencil is the legacy Go name for design-document source evidence.
	EvidencePencil = EvidenceDesignDocumentSource
	// EvidenceLayout is the legacy Go name for computed design-document evidence.
	EvidenceLayout = EvidenceDesignDocumentComputed
)

// FindingStatus separates blocking findings from advisory findings.
type FindingStatus string

const (
	// FindingFail is a blocking deterministic finding.
	FindingFail FindingStatus = "fail"
	// FindingAdvisory is a non-blocking deterministic finding.
	FindingAdvisory FindingStatus = "advisory"
)

// Position is a one-based source position.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Range is an optional source span.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Diagnostic is stable across text, JSON, and SARIF reports.
type Diagnostic struct {
	RuleID        string        `json:"ruleId"`
	SourceRuleIDs []string      `json:"sourceRuleIds"`
	Status        FindingStatus `json:"status"`
	Severity      Severity      `json:"severity"`
	Message       string        `json:"message"`
	Path          string        `json:"path"`
	Range         *Range        `json:"range,omitempty"`
	EvidenceKind  EvidenceKind  `json:"evidenceKind"`
	Platform      string        `json:"platform"`
	Viewport      string        `json:"viewport,omitempty"`
	Owner         string        `json:"owner"`
	Fingerprint   string        `json:"fingerprint"`
	Category      string        `json:"category,omitempty"`
}

// WithViewport adds exact rendered viewport attribution and refreshes the fingerprint.
func WithViewport(value Diagnostic, viewport string) Diagnostic {
	value.Viewport = viewport
	value.Fingerprint = fingerprint(value)
	return value
}

// New returns a diagnostic with a deterministic fingerprint.
func New(ruleID string, severity Severity, message, path string, sourceRange *Range, evidence EvidenceKind, platform, owner, category string) Diagnostic {
	return NewWithSources(ruleID, nil, severity, message, path, sourceRange, evidence, platform, owner, category)
}

// NewWithSources returns a canonical diagnostic with separately tracked upstream rule IDs.
func NewWithSources(ruleID string, sourceRuleIDs []string, severity Severity, message, path string, sourceRange *Range, evidence EvidenceKind, platform, owner, category string) Diagnostic {
	status := FindingFail
	if severity == SeverityWarning {
		status = FindingAdvisory
	}
	d := Diagnostic{
		RuleID:        ruleID,
		SourceRuleIDs: uniqueSorted(sourceRuleIDs),
		Status:        status,
		Severity:      severity,
		Message:       message,
		Path:          path,
		Range:         sourceRange,
		EvidenceKind:  evidence,
		Platform:      platform,
		Owner:         owner,
		Category:      category,
	}
	d.Fingerprint = fingerprint(d)
	return d
}

// MergeCanonical merges duplicate canonical findings and preserves sorted upstream provenance.
func MergeCanonical(diagnostics []Diagnostic) []Diagnostic {
	byFingerprint := make(map[string]Diagnostic, len(diagnostics))
	for _, finding := range diagnostics {
		finding.SourceRuleIDs = uniqueSorted(finding.SourceRuleIDs)
		if existing, ok := byFingerprint[finding.Fingerprint]; ok {
			existing.SourceRuleIDs = uniqueSorted(append(existing.SourceRuleIDs, finding.SourceRuleIDs...))
			byFingerprint[finding.Fingerprint] = existing
			continue
		}
		byFingerprint[finding.Fingerprint] = finding
	}
	result := make([]Diagnostic, 0, len(byFingerprint))
	for _, finding := range byFingerprint {
		result = append(result, finding)
	}
	Sort(result)
	return result
}

// Sort orders diagnostics independently of filesystem and map iteration order.
func Sort(diagnostics []Diagnostic) {
	sort.SliceStable(diagnostics, func(i, j int) bool {
		left, right := diagnostics[i], diagnostics[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		leftLine, leftColumn := position(left.Range)
		rightLine, rightColumn := position(right.Range)
		if leftLine != rightLine {
			return leftLine < rightLine
		}
		if leftColumn != rightColumn {
			return leftColumn < rightColumn
		}
		if left.RuleID != right.RuleID {
			return left.RuleID < right.RuleID
		}
		if left.Message != right.Message {
			return left.Message < right.Message
		}
		return left.Fingerprint < right.Fingerprint
	})
}

func position(sourceRange *Range) (int, int) {
	if sourceRange == nil {
		return 0, 0
	}
	return sourceRange.Start.Line, sourceRange.Start.Column
}

func fingerprint(d Diagnostic) string {
	line, column := position(d.Range)
	content := d.RuleID + "\x00" + string(d.Severity) + "\x00" + d.Message + "\x00" + d.Path + "\x00" +
		string(d.EvidenceKind) + "\x00" + d.Platform + "\x00" + d.Viewport + "\x00" + d.Owner + "\x00" + d.Category + "\x00" +
		strconv.Itoa(line) + "\x00" + strconv.Itoa(column)
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func uniqueSorted(values []string) []string {
	result := append([]string(nil), values...)
	if result == nil {
		result = []string{}
	}
	sort.Strings(result)
	return slices.Compact(result)
}
