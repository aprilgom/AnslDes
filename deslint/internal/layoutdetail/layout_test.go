package layoutdetail

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
)

func TestWebFixtureMapsAllFourteenRules(t *testing.T) {
	evidence, findings, err := Analyze("web.json", read(t, "../../../packages/schema/testdata/layout-negative-web.json"), Config{ProfileID: "operate", Density: "comfortable"})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.EvidenceKind != diagnostic.EvidenceWebRendered || evidence.CapturePath == evidence.ComputedBoundsPath {
		t.Fatalf("evidence = %#v", evidence)
	}
	for _, ruleID := range RuleIDs() {
		if !hasRule(findings, ruleID) {
			t.Fatalf("missing %s in %#v", ruleID, findings)
		}
	}
	for _, finding := range findings {
		if !slices.Contains(finding.SourceRuleIDs, strings.TrimPrefix(finding.RuleID, "layout/")) {
			t.Fatalf("source mapping = %#v", finding)
		}
	}
	assertSource(t, findings, "layout/nested-cards", "hallmark-eight-05")
	assertSource(t, findings, "layout/equal-icon-feature-columns", "hallmark-eight-03")
	assertSource(t, findings, "layout/full-viewport-centered-hero", "hallmark-eight-04")
}

func TestDataGridCanvasAndSemanticBoundaryAreNegativeFixtures(t *testing.T) {
	_, findings, err := Analyze("native.json", read(t, "../../../packages/schema/testdata/layout-permissions-native.json"), Config{ProfileID: "operate", Density: "compact"})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestDesignDocumentComputedEvidenceIsIndependent(t *testing.T) {
	evidence, findings, err := Analyze("document.json", read(t, "../../../packages/schema/testdata/layout-design-document.json"), Config{ProfileID: "operate", Density: "comfortable"})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.EvidenceKind != diagnostic.EvidenceDesignDocumentComputed || evidence.Platform != "design-document" || evidence.DocumentReport == nil || evidence.DocumentReport.NodeCount != len(evidence.Nodes) || len(findings) != 0 {
		t.Fatalf("evidence/findings = %#v %#v", evidence, findings)
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

func hasRule(findings []diagnostic.Diagnostic, ruleID string) bool {
	for _, finding := range findings {
		if finding.RuleID == ruleID {
			return true
		}
	}
	return false
}

func assertSource(t *testing.T, findings []diagnostic.Diagnostic, ruleID, sourceID string) {
	t.Helper()
	for _, finding := range findings {
		if finding.RuleID == ruleID && slices.Contains(finding.SourceRuleIDs, sourceID) {
			return
		}
	}
	t.Fatalf("%s missing source %s", ruleID, sourceID)
}
