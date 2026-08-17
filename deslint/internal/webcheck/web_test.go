package webcheck

import (
	"errors"
	"os"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

func TestEachCompletedProviderMapsCanonicalFindings(t *testing.T) {
	cases := []struct {
		fixture string
		ruleID  string
	}{
		{"web-provider-regex-negative.json", rules.RuleVisualSideTab},
		{"web-provider-static-negative.json", rules.RuleVisualSideTab},
		{"web-provider-browser-negative.json", rules.RuleRuntimeContentHiddenAtRest},
		{"web-provider-visual-negative.json", rules.RuleColorLowContrast},
	}
	for _, item := range cases {
		evidence, findings, excluded, err := Analyze(item.fixture, read(t, "../../../packages/schema/testdata/"+item.fixture), config())
		if err != nil || evidence.Execution.Status != "completed" || len(excluded) != 0 || !hasRule(findings, item.ruleID) {
			t.Fatalf("%s = %#v %#v %#v %v", item.fixture, evidence, findings, excluded, err)
		}
	}
	for _, fixture := range []string{"web-provider-regex-positive.json", "web-provider-static-positive.json"} {
		_, findings, excluded, err := Analyze(fixture, read(t, "../../../packages/schema/testdata/"+fixture), config())
		if err != nil || len(findings) != 0 || len(excluded) != 0 {
			t.Fatalf("%s positive = %#v %#v %v", fixture, findings, excluded, err)
		}
	}
}

func TestGeneratedArtifactRequiresExactFingerprint(t *testing.T) {
	_, findings, excluded, err := Analyze("artifact.json", read(t, "../../../packages/schema/testdata/web-provider-artifact-excluded.json"), config())
	if err != nil || len(findings) != 0 || len(excluded) != 1 || excluded[0].Exclusion.ReproductionCommand == "" {
		t.Fatalf("exact exclusion = %#v %#v %v", findings, excluded, err)
	}
	drifted := config()
	drifted.ArtifactExclusions[0].FingerprintSHA256 = "2222222222222222222222222222222222222222222222222222222222222222"
	_, findings, excluded, err = Analyze("artifact.json", read(t, "../../../packages/schema/testdata/web-provider-artifact-excluded.json"), drifted)
	if err != nil || !hasRule(findings, rules.RuleVisualSideTab) || len(excluded) != 0 {
		t.Fatalf("drifted exclusion = %#v %#v %v", findings, excluded, err)
	}
}

func TestFallbackIsNotRunAndBrowserFailureIsExecutionError(t *testing.T) {
	fallback, findings, _, err := Analyze("fallback.json", read(t, "../../../packages/schema/testdata/web-provider-fallback-not-run.json"), config())
	if err != nil || fallback.Execution.Status != "not-run" || len(findings) != 0 {
		t.Fatalf("fallback = %#v %#v %v", fallback, findings, err)
	}
	_, _, _, err = Analyze("error.json", read(t, "../../../packages/schema/testdata/web-provider-browser-error.json"), config())
	var providerError *ProviderExecutionError
	if !errors.As(err, &providerError) || providerError.Provider != "browser" {
		t.Fatalf("provider error = %v", err)
	}
}

func TestFullProviderMatrixIsExact(t *testing.T) {
	evidences := []Evidence{}
	for _, fixture := range []string{
		"web-provider-artifact-excluded.json",
		"web-provider-static-positive.json",
		"web-provider-browser-positive.json",
		"web-provider-visual-positive.json",
	} {
		evidence, _, _, err := Analyze(fixture, read(t, "../../../packages/schema/testdata/"+fixture), config())
		if err != nil {
			t.Fatal(err)
		}
		evidences = append(evidences, evidence)
	}
	if findings := CoverageFindings(evidences, config()); len(findings) != 0 {
		t.Fatalf("coverage = %#v", findings)
	}
	if findings := CoverageFindings(evidences[:2], config()); !hasRule(findings, rules.RuleEvidenceMissing) {
		t.Fatalf("missing coverage = %#v", findings)
	}
}

func config() Config {
	return Config{
		RegistryVersion: "1.0.0",
		Routes:          map[string]Route{"example-home": {Owner: "example-web-owner", Target: "http://127.0.0.1:4173/example"}},
		Viewports: map[string]Viewport{
			"desktop": {ID: "desktop", Width: 1280, Height: 800},
			"mobile":  {ID: "mobile", Width: 390, Height: 844},
		},
		Themes: []string{"light", "dark"}, FontScales: []float64{1, 1.6}, ReduceMotion: []bool{false, true},
		RequiredCaptures: []CaptureRequirement{
			{ID: "regex-source-light", Provider: "regex-source", RouteID: "example-home", ViewportID: "desktop", Theme: "light", FontScale: 1, ReduceMotion: false},
			{ID: "static-html-light", Provider: "static-html", RouteID: "example-home", ViewportID: "desktop", Theme: "light", FontScale: 1, ReduceMotion: false},
			{ID: "browser-dark-large", Provider: "browser", RouteID: "example-home", ViewportID: "mobile", Theme: "dark", FontScale: 1.6, ReduceMotion: true},
			{ID: "visual-dark-large", Provider: "visual-contrast", RouteID: "example-home", ViewportID: "mobile", Theme: "dark", FontScale: 1.6, ReduceMotion: true},
		},
		ArtifactExclusions: []ArtifactExclusion{{Path: "generated/example.css", FingerprintSHA256: "1111111111111111111111111111111111111111111111111111111111111111", Owner: "example-build-owner", Rationale: "generated vendor compatibility artifact", ReproductionCommand: "npm run build:example"}},
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
