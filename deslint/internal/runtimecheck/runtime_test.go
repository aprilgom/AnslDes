package runtimecheck

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

func TestWebFixtureReportsThreeRulesWithExactRouteAndOwner(t *testing.T) {
	evidence, findings, err := Analyze("web.json", read(t, "../../../packages/schema/testdata/runtime-negative-web.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Platform != "web" {
		t.Fatalf("evidence = %#v", evidence)
	}
	for _, ruleID := range RuleIDs() {
		finding := findRule(t, findings, ruleID)
		if !strings.Contains(finding.Path, "#/routes/example-route/") {
			t.Fatalf("finding route = %#v", finding)
		}
		if finding.Owner != "example-route-owner" && finding.Owner != "example-content-owner" {
			t.Fatalf("finding owner = %#v", finding)
		}
	}
}

func TestNativeFailureKindsRemainSeparateFromWeb(t *testing.T) {
	_, findings, err := Analyze("native.json", read(t, "../../../packages/schema/testdata/runtime-negative-native.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	if countRule(findings, rules.RuleRuntimeScriptError) != 3 || !hasRule(findings, rules.RuleRuntimeJustifiedText) {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestExactExportOwnerPermissionSuppressesJustification(t *testing.T) {
	_, findings, err := Analyze("export.json", read(t, "../../../packages/schema/testdata/runtime-permissions.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %#v", findings)
	}
	drifted := config()
	drifted.JustifiedTextExceptions[0].Owner = "different-owner"
	_, findings, err = Analyze("export.json", read(t, "../../../packages/schema/testdata/runtime-permissions.json"), drifted)
	if err != nil || !hasRule(findings, rules.RuleRuntimeJustifiedText) {
		t.Fatalf("drifted permission = %#v, %v", findings, err)
	}
}

func TestDetectorFailureIsExecutionErrorNotScriptFinding(t *testing.T) {
	_, findings, err := Analyze("failed.json", read(t, "../../../packages/schema/testdata/runtime-detector-failure-web.json"), config())
	var processError *DetectorProcessError
	if !errors.As(err, &processError) || len(findings) != 0 || processError.Stage != "navigation" {
		t.Fatalf("error/findings = %v %#v", err, findings)
	}
}

func config() Config {
	return Config{
		RegistryVersion: "1.0.0",
		JustifiedTextExceptions: []JustifiedTextException{{
			Platform: "web", SurfaceID: "example-export-surface", RouteID: "example-export-route",
			NodeID: "export-body-copy", Owner: "example-export-owner", Context: "export",
		}},
	}
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

func findRule(t *testing.T, findings []diagnostic.Diagnostic, ruleID string) diagnostic.Diagnostic {
	t.Helper()
	for _, finding := range findings {
		if finding.RuleID == ruleID {
			return finding
		}
	}
	t.Fatalf("missing %s in %#v", ruleID, findings)
	return diagnostic.Diagnostic{}
}

func hasRule(findings []diagnostic.Diagnostic, ruleID string) bool {
	return countRule(findings, ruleID) > 0
}

func countRule(findings []diagnostic.Diagnostic, ruleID string) int {
	count := 0
	for _, finding := range findings {
		if finding.RuleID == ruleID {
			count++
		}
	}
	return count
}
