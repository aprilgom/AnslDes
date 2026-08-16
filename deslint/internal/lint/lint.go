// Package lint orchestrates independently acquired evidence through one rule model.
package lint

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/aprilgom/AnslDes/deslint/internal/contract"
	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/layout"
	"github.com/aprilgom/AnslDes/deslint/internal/pen"
	"github.com/aprilgom/AnslDes/deslint/internal/policy"
	"github.com/aprilgom/AnslDes/deslint/internal/report"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
	"github.com/aprilgom/AnslDes/deslint/internal/source"
)

// Input is one named immutable evidence payload.
type Input struct {
	Path     string
	Contents []byte
}

// Request contains all inputs without filesystem or process coupling.
type Request struct {
	Definition Input
	Policy     policy.Policy
	Sources    []Input
	Pencil     *Input
	Layout     *Input
	Now        time.Time
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
	evidence := []report.EvidenceStatus{{
		Kind: diagnostic.EvidenceDefinition, Status: "acquired", Path: request.Definition.Path,
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
		language, err := languageForPath(input.Path)
		if err != nil {
			return report.Report{}, err
		}
		summary, err := r.SourceAnalyzer.Analyze(input.Path, input.Contents, language)
		if err != nil {
			return report.Report{}, err
		}
		sourceAcquired = true
		diagnostics = append(diagnostics, rules.AnalyzeSource(summary, request.Policy.RawPropertyKinds(), severity)...)
	}
	if sourceAcquired {
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceSource, Status: "acquired"})
	} else {
		evidence = append(evidence, missingEvidence(request.Policy, diagnostic.EvidenceSource, &diagnostics))
	}

	if request.Pencil != nil {
		findings, err := pen.Analyze(request.Pencil.Path, request.Pencil.Contents, request.Policy.RawPropertyKinds(), severity)
		if err != nil {
			return report.Report{}, err
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidencePencil, Status: "acquired", Path: request.Pencil.Path})
	} else {
		evidence = append(evidence, missingEvidence(request.Policy, diagnostic.EvidencePencil, &diagnostics))
	}

	if request.Layout != nil {
		findings, err := layout.Analyze(request.Layout.Path, request.Layout.Contents, request.Policy.Evidence.LayoutDocumentSHA256, severity)
		if err != nil {
			return report.Report{}, err
		}
		diagnostics = append(diagnostics, findings...)
		evidence = append(evidence, report.EvidenceStatus{Kind: diagnostic.EvidenceLayout, Status: "acquired", Path: request.Layout.Path})
	} else {
		evidence = append(evidence, missingEvidence(request.Policy, diagnostic.EvidenceLayout, &diagnostics))
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
	diagnostics = request.Policy.ApplyExceptions(diagnostics, request.Now)
	diagnostics, failed := enforceBudgets(diagnostics, request.Policy)
	return report.New(definition.DefinitionID, evidence, diagnostics, failed), nil
}

func missingEvidence(productPolicy policy.Policy, kind diagnostic.EvidenceKind, diagnostics *[]diagnostic.Diagnostic) report.EvidenceStatus {
	status := "not-run"
	if productPolicy.Requires(kind) {
		status = "missing"
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
	return report.EvidenceStatus{Kind: kind, Status: status}
}

func enforceBudgets(diagnostics []diagnostic.Diagnostic, productPolicy policy.Policy) ([]diagnostic.Diagnostic, bool) {
	counts := map[string]int{"error": 0, "warning": 0, "raw": 0, "overflow": 0, "overlap": 0}
	for _, finding := range diagnostics {
		counts[string(finding.Severity)]++
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
	budgets := map[string]int{
		"error": productPolicy.Budgets.Error, "warning": productPolicy.Budgets.Warning,
		"raw": productPolicy.Budgets.Raw, "overflow": productPolicy.Budgets.Overflow,
		"overlap": productPolicy.Budgets.Overlap,
	}
	failed := false
	for _, category := range []string{"error", "warning", "raw", "overflow", "overlap"} {
		if counts[category] <= budgets[category] {
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
