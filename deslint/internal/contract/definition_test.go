package contract_test

import (
	"os"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/contract"
	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

func TestAnalyzeDefinitionReferences(t *testing.T) {
	t.Parallel()
	valid := readDefinition(t)
	severity := func(string) diagnostic.Severity { return diagnostic.SeverityError }

	analysis, err := contract.Analyze("definition.json", valid, severity)
	if err != nil {
		t.Fatalf("Analyze(valid) error = %v", err)
	}
	if analysis.DefinitionID != "example-product" || len(analysis.Diagnostics) != 0 {
		t.Fatalf("Analyze(valid) = %#v", analysis)
	}

	unknown := []byte(strings.Replace(string(valid), "{radius.primitive.medium}", "{radius.primitive.missing}", 1))
	analysis, err = contract.Analyze("definition.json", unknown, severity)
	if err != nil {
		t.Fatalf("Analyze(unknown) error = %v", err)
	}
	if !hasRule(analysis.Diagnostics, rules.RuleDefinitionUnknownToken) {
		t.Fatalf("unknown diagnostics = %#v", analysis.Diagnostics)
	}

	invalid := []byte(strings.Replace(string(valid), "{radius.primitive.medium}", "{product.brand.radius}", 1))
	analysis, err = contract.Analyze("definition.json", invalid, severity)
	if err != nil {
		t.Fatalf("Analyze(invalid) error = %v", err)
	}
	if !hasRule(analysis.Diagnostics, rules.RuleDefinitionInvalidRef) {
		t.Fatalf("invalid diagnostics = %#v", analysis.Diagnostics)
	}
}

func TestAnalyzeDefinitionRejectsDuplicateAndVersionDrift(t *testing.T) {
	t.Parallel()
	valid := string(readDefinition(t))
	severity := func(string) diagnostic.Severity { return diagnostic.SeverityError }
	duplicate := []byte(strings.Replace(valid, `"id": "example-product",`, `"id": "first", "id": "example-product",`, 1))
	analysis, err := contract.Analyze("definition.json", duplicate, severity)
	if err != nil {
		t.Fatalf("Analyze(duplicate) error = %v", err)
	}
	if !hasRule(analysis.Diagnostics, rules.RuleDefinitionSchemaVersion) {
		t.Fatalf("duplicate diagnostics = %#v", analysis.Diagnostics)
	}

	drift := []byte(strings.Replace(valid, `"schemaVersion": 1`, `"schemaVersion": 2`, 1))
	analysis, err = contract.Analyze("definition.json", drift, severity)
	if err != nil {
		t.Fatalf("Analyze(drift) error = %v", err)
	}
	if !hasRule(analysis.Diagnostics, rules.RuleDefinitionSchemaVersion) {
		t.Fatalf("version diagnostics = %#v", analysis.Diagnostics)
	}
}

func readDefinition(t *testing.T) []byte {
	t.Helper()
	contents, err := os.ReadFile("../../../packages/schema/testdata/example-product.json")
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
