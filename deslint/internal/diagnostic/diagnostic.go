// Package diagnostic defines the stable external finding contract.
package diagnostic

import (
	"crypto/sha256"
	"encoding/hex"
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
	// EvidenceSource is TypeScript or TSX syntax evidence.
	EvidenceSource EvidenceKind = "source"
	// EvidencePencil is serialized Pencil document evidence.
	EvidencePencil EvidenceKind = "pencil"
	// EvidenceLayout is computed layout-engine evidence.
	EvidenceLayout EvidenceKind = "computed-layout"
	// EvidenceExecution is a missing, malformed, or stale input.
	EvidenceExecution EvidenceKind = "execution"
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
	RuleID       string       `json:"ruleId"`
	Severity     Severity     `json:"severity"`
	Message      string       `json:"message"`
	Path         string       `json:"path"`
	Range        *Range       `json:"range,omitempty"`
	EvidenceKind EvidenceKind `json:"evidenceKind"`
	Platform     string       `json:"platform"`
	Owner        string       `json:"owner"`
	Fingerprint  string       `json:"fingerprint"`
	Category     string       `json:"category,omitempty"`
}

// New returns a diagnostic with a deterministic fingerprint.
func New(ruleID string, severity Severity, message, path string, sourceRange *Range, evidence EvidenceKind, platform, owner, category string) Diagnostic {
	d := Diagnostic{
		RuleID:       ruleID,
		Severity:     severity,
		Message:      message,
		Path:         path,
		Range:        sourceRange,
		EvidenceKind: evidence,
		Platform:     platform,
		Owner:        owner,
		Category:     category,
	}
	d.Fingerprint = fingerprint(d)
	return d
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
		string(d.EvidenceKind) + "\x00" + d.Platform + "\x00" + d.Owner + "\x00" + d.Category + "\x00" +
		strconv.Itoa(line) + "\x00" + strconv.Itoa(column)
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
