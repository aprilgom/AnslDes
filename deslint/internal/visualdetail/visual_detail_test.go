package visualdetail

import (
	"os"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

func TestFiveRulesPreserveProviderEvidenceAndOwner(t *testing.T) {
	for _, fixture := range []string{"visual-detail-web.json", "visual-detail-native.json", "visual-detail-design-document.json"} {
		// #nosec G304 -- repository-owned fixture names.
		contents, err := os.ReadFile("../../../packages/schema/testdata/" + fixture)
		if err != nil {
			t.Fatal(err)
		}
		evidence, findings, err := Analyze(fixture, contents, func(string) diagnostic.Severity { return diagnostic.SeverityError }, nil)
		if err != nil || len(findings) != 5 {
			t.Fatalf("%s: %#v %v", fixture, findings, err)
		}
		for _, finding := range findings {
			if finding.EvidenceKind != evidence.EvidenceKind || finding.Owner == "" || len(finding.SourceRuleIDs) != 1 {
				t.Fatalf("finding = %#v", finding)
			}
		}
	}
}

func TestPermissionsAreExactAndStructuralRegressionRemains(t *testing.T) {
	contents, err := os.ReadFile("../../../packages/schema/testdata/visual-detail-permissions.json")
	if err != nil {
		t.Fatal(err)
	}
	_, findings, err := Analyze("permissions.json", contents, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, finding := range findings {
		counts[finding.RuleID]++
	}
	if len(findings) != 3 || counts[rules.RuleVisualSideTab] != 1 || counts[rules.RuleVisualBorderAccentRounded] != 1 || counts[rules.RuleNativeListRowAccessoryWrapper] != 1 {
		t.Fatalf("findings = %#v", findings)
	}
}
