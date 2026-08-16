// Package report renders deterministic text, JSON, and SARIF output.
package report

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
)

const schemaVersion = 1

// Format is an external report encoding.
type Format string

const (
	// FormatText emits one stable human-readable diagnostic per line.
	FormatText Format = "text"
	// FormatJSON emits the native deslint report contract.
	FormatJSON Format = "json"
	// FormatSARIF emits SARIF 2.1.0 for code-scanning consumers.
	FormatSARIF Format = "sarif"
)

// EvidenceStatus records acquisition independently for each evidence kind.
type EvidenceStatus struct {
	Kind   diagnostic.EvidenceKind `json:"kind"`
	Status string                  `json:"status"`
	Path   string                  `json:"path,omitempty"`
}

// Summary contains exact budget counters.
type Summary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Raw      int `json:"raw"`
	Overflow int `json:"overflow"`
	Overlap  int `json:"overlap"`
}

// Report is the deterministic native JSON contract.
type Report struct {
	SchemaVersion     int                     `json:"schemaVersion"`
	Status            string                  `json:"status"`
	DefinitionID      string                  `json:"definitionId"`
	FingerprintSHA256 string                  `json:"fingerprintSha256"`
	Evidence          []EvidenceStatus        `json:"evidence"`
	Summary           Summary                 `json:"summary"`
	Diagnostics       []diagnostic.Diagnostic `json:"diagnostics"`
}

// New creates a sorted report with a content fingerprint.
func New(definitionID string, evidence []EvidenceStatus, diagnostics []diagnostic.Diagnostic, failed bool) Report {
	diagnostic.Sort(diagnostics)
	sort.SliceStable(evidence, func(i, j int) bool { return evidence[i].Kind < evidence[j].Kind })
	status := "pass"
	if failed {
		status = "fail"
	}
	result := Report{
		SchemaVersion: schemaVersion,
		Status:        status,
		DefinitionID:  definitionID,
		Evidence:      evidence,
		Summary:       summarize(diagnostics),
		Diagnostics:   diagnostics,
	}
	result.FingerprintSHA256 = fingerprint(result)
	return result
}

// Write emits a report in the requested format.
func Write(writer io.Writer, value Report, format Format) error {
	switch format {
	case FormatText:
		return writeText(writer, value)
	case FormatJSON:
		return writeJSON(writer, value)
	case FormatSARIF:
		return writeSARIF(writer, value)
	default:
		return fmt.Errorf("unsupported report format %q", format)
	}
}

func summarize(diagnostics []diagnostic.Diagnostic) Summary {
	var summary Summary
	for _, finding := range diagnostics {
		if finding.Severity == diagnostic.SeverityError {
			summary.Errors++
		} else {
			summary.Warnings++
		}
		if finding.Category == "raw" {
			summary.Raw++
		}
		if finding.Category == "overflow" || finding.Category == "clipping" {
			summary.Overflow++
		}
		if finding.Category == "overlap" {
			summary.Overlap++
		}
	}
	return summary
}

func writeText(writer io.Writer, value Report) error {
	for _, finding := range value.Diagnostics {
		line, column := 0, 0
		if finding.Range != nil {
			line, column = finding.Range.Start.Line, finding.Range.Start.Column
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s:%d:%d %s %s %s [%s]\n",
			finding.Path,
			line,
			column,
			finding.Severity,
			finding.RuleID,
			finding.Message,
			finding.EvidenceKind,
		); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(
		writer,
		"%s errors=%d warnings=%d raw=%d overflow=%d overlap=%d fingerprint=%s\n",
		strings.ToUpper(value.Status),
		value.Summary.Errors,
		value.Summary.Warnings,
		value.Summary.Raw,
		value.Summary.Overflow,
		value.Summary.Overlap,
		value.FingerprintSHA256,
	)
	return err
}

func writeJSON(writer io.Writer, value Report) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name  string      `json:"name"`
	Rules []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID string `json:"id"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
	Partial   sarifPartial    `json:"partialFingerprints"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	Physical sarifPhysical `json:"physicalLocation"`
}

type sarifPhysical struct {
	Artifact sarifArtifact `json:"artifactLocation"`
	Region   *sarifRegion  `json:"region,omitempty"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
	EndLine     int `json:"endLine"`
	EndColumn   int `json:"endColumn"`
}

type sarifPartial struct {
	Primary string `json:"primaryLocationLineHash"`
}

func writeSARIF(writer io.Writer, value Report) error {
	ruleSet := make(map[string]struct{})
	results := make([]sarifResult, 0, len(value.Diagnostics))
	for _, finding := range value.Diagnostics {
		ruleSet[finding.RuleID] = struct{}{}
		level := "error"
		if finding.Severity == diagnostic.SeverityWarning {
			level = "warning"
		}
		var region *sarifRegion
		if finding.Range != nil {
			region = &sarifRegion{
				StartLine: finding.Range.Start.Line, StartColumn: finding.Range.Start.Column,
				EndLine: finding.Range.End.Line, EndColumn: finding.Range.End.Column,
			}
		}
		results = append(results, sarifResult{
			RuleID:  finding.RuleID,
			Level:   level,
			Message: sarifMessage{Text: finding.Message},
			Locations: []sarifLocation{{Physical: sarifPhysical{
				Artifact: sarifArtifact{URI: finding.Path}, Region: region,
			}}},
			Partial: sarifPartial{Primary: finding.Fingerprint},
		})
	}
	ruleIDs := make([]string, 0, len(ruleSet))
	for ruleID := range ruleSet {
		ruleIDs = append(ruleIDs, ruleID)
	}
	sort.Strings(ruleIDs)
	rules := make([]sarifRule, 0, len(ruleIDs))
	for _, ruleID := range ruleIDs {
		rules = append(rules, sarifRule{ID: ruleID})
	}
	log := sarifLog{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []sarifRun{{
			Tool:    sarifTool{Driver: sarifDriver{Name: "deslint", Rules: rules}},
			Results: results,
		}},
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(log)
}

func fingerprint(value Report) string {
	clone := value
	clone.FingerprintSHA256 = ""
	contents, _ := json.Marshal(clone)
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
