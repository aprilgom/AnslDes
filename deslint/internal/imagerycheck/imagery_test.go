package imagerycheck

import (
	"os"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

func TestWebFixtureMapsBothImageryRules(t *testing.T) {
	evidence, findings, err := Analyze("web.json", read(t, "../../../packages/schema/testdata/imagery-negative-web.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Platform != "web" || !hasRule(findings, rules.RuleImageryShapeAssembledIllustration) || !hasRule(findings, rules.RuleImageryBrokenImage) {
		t.Fatalf("evidence/findings = %#v %#v", evidence, findings)
	}
}

func TestNativeOwnerSourceFingerprintAndAccessibilityDrift(t *testing.T) {
	_, findings, err := Analyze("native.json", read(t, "../../../packages/schema/testdata/imagery-negative-native.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, rules.RuleImageryBrokenImage) {
		t.Fatalf("findings = %#v", findings)
	}
	messages := ""
	for _, finding := range findings {
		messages += finding.Message + "\n"
	}
	if !strings.Contains(messages, "drifted") || !strings.Contains(messages, "accessible label") || !strings.Contains(messages, "placeholder") {
		t.Fatalf("messages = %s", messages)
	}
}

func TestIconAndIntentionalOmissionAreNegativeFixtures(t *testing.T) {
	_, findings, err := Analyze("permissions.json", read(t, "../../../packages/schema/testdata/imagery-permissions.json"), config())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %#v", findings)
	}
}

func config() Config {
	return Config{RegistryVersion: "1.0.0", Assets: map[string]Asset{
		"hero-art": {
			Owner: "example-asset-owner", Role: "hero-illustration", ImplementationSource: "asset://example-hero",
			Consumers: []string{"example-hero"}, FingerprintSHA256: strings.Repeat("1", 64), Decorative: false,
		},
		"optional-art": {
			Owner: "example-asset-owner", Role: "hero-illustration", ImplementationSource: "none",
			Consumers: []string{"example-optional-hero"}, FingerprintSHA256: strings.Repeat("0", 64), IntentionallyOmitted: true, Decorative: true,
		},
	}}
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
