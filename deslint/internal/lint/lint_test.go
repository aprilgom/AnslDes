package lint_test

import (
	"os"
	"testing"
	"time"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/lint"
	"github.com/aprilgom/AnslDes/deslint/internal/policy"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
	"github.com/aprilgom/AnslDes/deslint/internal/source/treesitter"
)

func TestRunnerKeepsEvidenceKindsIndependent(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	result, err := (lint.Runner{SourceAnalyzer: treesitter.NewAnalyzer()}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "pass" || len(result.Diagnostics) != 0 || len(result.Evidence) != 4 {
		t.Fatalf("Run() = %#v", result)
	}
	for _, evidence := range result.Evidence {
		if evidence.Status != "acquired" {
			t.Fatalf("evidence = %#v", result.Evidence)
		}
	}
}

func TestRunnerReportsRawSyntaxLayoutAndBudgetFailures(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Sources = []lint.Input{input(t, "../../testdata/negative/Raw.tsx")}
	request.Pencil = new(input(t, "../../testdata/negative/raw.pen.json"))
	request.Layout = new(input(t, "../../testdata/negative/layout.json"))
	result, err := (lint.Runner{SourceAnalyzer: treesitter.NewAnalyzer()}).Run(request)
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
	result, err := (lint.Runner{SourceAnalyzer: treesitter.NewAnalyzer()}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "fail" || !hasRule(result.Diagnostics, rules.RuleEvidenceMissing) || !hasRule(result.Diagnostics, rules.RuleSourceSyntaxError) {
		t.Fatalf("Run() = %#v", result)
	}
}

func TestRunnerRejectsStaleLayoutEvidence(t *testing.T) {
	t.Parallel()
	request := positiveRequest(t)
	request.Policy.Evidence.LayoutDocumentSHA256 = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	result, err := (lint.Runner{SourceAnalyzer: treesitter.NewAnalyzer()}).Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "fail" || !hasRule(result.Diagnostics, rules.RuleEvidenceStale) {
		t.Fatalf("Run() = %#v", result)
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
		Definition: input(t, "../../../packages/schema/testdata/example-product.json"),
		Policy:     productPolicy,
		Sources:    []lint.Input{input(t, "../../testdata/positive/Example.tsx")},
		Pencil:     new(input(t, "../../testdata/positive/document.pen.json")),
		Layout:     new(input(t, "../../testdata/positive/layout.json")),
		Now:        time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC),
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
