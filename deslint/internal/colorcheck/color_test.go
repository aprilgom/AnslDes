package colorcheck

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
)

func TestLiteralTokenAndComputedFixtureMapsAllNineRules(t *testing.T) {
	contents := read(t, "../../../packages/schema/testdata/color-negative-light.json")
	evidence, findings, err := Analyze("light.json", contents, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Theme != "light" || evidence.ScreenshotPath == evidence.ComputedColorPath {
		t.Fatalf("evidence = %#v", evidence)
	}
	for _, ruleID := range RuleIDs() {
		if !hasRule(findings, ruleID) {
			t.Fatalf("missing %s in %#v", ruleID, findings)
		}
	}
	hallmark := false
	for _, finding := range findings {
		if !slices.Contains(finding.SourceRuleIDs, strings.TrimPrefix(finding.RuleID, "color/")) {
			t.Fatalf("mapping = %#v", finding)
		}
		if finding.RuleID == "color/ai-color-palette" && slices.Contains(finding.SourceRuleIDs, "hallmark-eight-01") {
			hallmark = true
		}
	}
	if !hallmark {
		t.Fatal("missing Hallmark palette provenance")
	}
}

func TestPermissionsDoNotExemptTextContrast(t *testing.T) {
	contents := read(t, "../../../packages/schema/testdata/color-permissions-dark.json")
	_, findings, err := Analyze("dark.json", contents, Config{ApprovedPalettes: map[string]PalettePermission{
		"brand-accent": {Contexts: []string{"brand"}, Themes: []string{"dark"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].RuleID != "color/low-contrast" || findings[0].Owner != "example-brand" {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestPalettePermissionRequiresExactContextAndTheme(t *testing.T) {
	contents := read(t, "../../../packages/schema/testdata/color-permissions-dark.json")
	_, findings, err := Analyze("dark.json", contents, Config{ApprovedPalettes: map[string]PalettePermission{
		"brand-accent": {Contexts: []string{"hero"}, Themes: []string{"light"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, "color/ai-color-palette") {
		t.Fatalf("wrong-scope palette approval was accepted: %#v", findings)
	}
}

func TestContrastUsesSRGBRelativeLuminance(t *testing.T) {
	if ratio, ok := contrast("#000000", "#FFFFFF"); !ok || ratio != 21 {
		t.Fatalf("contrast = %v %v", ratio, ok)
	}
	if ratio, ok := contrast("#777777", "#FFFFFF"); !ok || ratio >= 4.5 {
		t.Fatalf("contrast = %v %v", ratio, ok)
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
