// Package layout validates independently generated computed-layout evidence.
package layout

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Report is the stable input boundary for a headless layout engine.
type Report struct {
	DocumentSHA256 string  `json:"documentSha256"`
	NodeCount      int     `json:"nodeCount"`
	Issues         []Issue `json:"issues"`
}

// Issue is one computed geometry problem.
type Issue struct {
	Kind    string `json:"kind"`
	NodeID  string `json:"nodeId"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// Analyze rejects stale reports and converts engine issues into layout diagnostics.
func Analyze(path string, contents []byte, expectedSHA string, severity func(string) diagnostic.Severity) ([]diagnostic.Diagnostic, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var report Report
	if err := decoder.Decode(&report); err != nil {
		return nil, fmt.Errorf("decode layout report: %w", err)
	}
	if report.DocumentSHA256 == "" || report.NodeCount < 0 {
		return nil, fmt.Errorf("layout report requires documentSha256 and non-negative nodeCount")
	}
	diagnostics := make([]diagnostic.Diagnostic, 0, len(report.Issues)+1)
	if expectedSHA != "" && report.DocumentSHA256 != expectedSHA {
		diagnostics = append(diagnostics, diagnostic.New(
			rules.RuleEvidenceStale,
			severity(rules.RuleEvidenceStale),
			fmt.Sprintf("layout document SHA %s does not match expected %s", report.DocumentSHA256, expectedSHA),
			path,
			nil,
			diagnostic.EvidenceLayout,
			"pencil",
			"ansldes/layout",
			"stale",
		))
	}
	for _, issue := range report.Issues {
		if issue.Kind != "overflow" && issue.Kind != "overlap" && issue.Kind != "clipping" {
			return nil, fmt.Errorf("unsupported layout issue kind %q", issue.Kind)
		}
		message := issue.Message
		if message == "" {
			message = fmt.Sprintf("%s at node %s", issue.Kind, issue.NodeID)
		}
		issuePath := issue.Path
		if issuePath == "" {
			issuePath = path
		}
		diagnostics = append(diagnostics, diagnostic.New(
			rules.RuleLayoutProblem,
			severity(rules.RuleLayoutProblem),
			message,
			issuePath,
			nil,
			diagnostic.EvidenceLayout,
			"pencil",
			"ansldes/layout",
			issue.Kind,
		))
	}
	diagnostic.Sort(diagnostics)
	return diagnostics, nil
}
