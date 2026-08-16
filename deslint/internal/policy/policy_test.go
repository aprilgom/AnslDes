package policy_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/policy"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

func TestParseRejectsBroadExcludesAndUnknownSeverity(t *testing.T) {
	t.Parallel()
	valid := string(readPolicy(t))
	if _, err := policy.Parse([]byte(valid)); err != nil {
		t.Fatalf("Parse(valid) error = %v", err)
	}
	broad := strings.Replace(valid, `"exactExcludes": []`, `"exactExcludes": ["src/**"]`, 1)
	if _, err := policy.Parse([]byte(broad)); err == nil {
		t.Fatal("Parse(broad exclude) error = nil")
	}
	unknown := strings.Replace(valid, `"source/raw-value": "error"`, `"source/raw-value": "off"`, 1)
	if _, err := policy.Parse([]byte(unknown)); err == nil {
		t.Fatal("Parse(unknown severity) error = nil")
	}
	escaping := strings.Replace(valid, `"exactExcludes": []`, `"exactExcludes": ["../outside.tsx"]`, 1)
	if _, err := policy.Parse([]byte(escaping)); err == nil {
		t.Fatal("Parse(escaping exclude) error = nil")
	}
	duplicateProperty := strings.Replace(valid, `"motion": ["duration"]`, `"motion": ["gap"]`, 1)
	if _, err := policy.Parse([]byte(duplicateProperty)); err == nil {
		t.Fatal("Parse(duplicate property) error = nil")
	}
}

func TestApplyExceptionsRequiresExactActiveMatch(t *testing.T) {
	t.Parallel()
	productPolicy, err := policy.Parse(readPolicy(t))
	if err != nil {
		t.Fatal(err)
	}
	productPolicy.Exceptions = []policy.Exception{{
		RuleID: rules.RuleSourceRawValue, Path: "src/Exact.tsx", Owner: "design-system",
		Rationale: "approved temporary migration", ExpiresAt: "2026-12-31",
	}}
	findings := []diagnostic.Diagnostic{
		diagnostic.New(rules.RuleSourceRawValue, diagnostic.SeverityError, "raw", "src/Exact.tsx", nil, diagnostic.EvidenceSource, "react-native", "owner", "raw"),
		diagnostic.New(rules.RuleSourceRawValue, diagnostic.SeverityError, "raw", "src/Other.tsx", nil, diagnostic.EvidenceSource, "react-native", "owner", "raw"),
	}
	filtered := productPolicy.ApplyExceptions(findings, time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC))
	if len(filtered) != 1 || filtered[0].Path != "src/Other.tsx" {
		t.Fatalf("ApplyExceptions() = %#v", filtered)
	}
	if expired := productPolicy.ExpiredExceptions(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)); len(expired) != 1 {
		t.Fatalf("ExpiredExceptions() = %#v", expired)
	}
}

func readPolicy(t *testing.T) []byte {
	t.Helper()
	contents, err := os.ReadFile("../../../packages/schema/testdata/example-policy.json")
	if err != nil {
		t.Fatal(err)
	}
	return contents
}
