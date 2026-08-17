package report_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/report"
)

func TestWriteTextJSONAndSARIFDeterministically(t *testing.T) {
	t.Parallel()
	finding := diagnostic.New(
		"source/raw-value", diagnostic.SeverityError, "raw color", "src/A.tsx",
		&diagnostic.Range{Start: diagnostic.Position{Line: 2, Column: 3}, End: diagnostic.Position{Line: 2, Column: 12}},
		diagnostic.EvidenceNativeSource, "react-native", "ansldes/source", "raw",
	)
	finding = diagnostic.WithViewport(finding, "mobile:390x844")
	value := newReport(t, report.Input{
		DefinitionID: "example-product",
		Evidence: []report.EvidenceStatus{{
			Kind: diagnostic.EvidenceNativeSource, Platform: "react-native", Status: report.EvidenceStatusPass,
		}},
		Diagnostics: []diagnostic.Diagnostic{finding},
		Failed:      true,
	})

	for _, format := range []report.Format{report.FormatText, report.FormatJSON, report.FormatSARIF} {
		var first, second bytes.Buffer
		if err := report.Write(&first, value, format); err != nil {
			t.Fatalf("Write(%s) error = %v", format, err)
		}
		if err := report.Write(&second, value, format); err != nil {
			t.Fatalf("Write(%s) second error = %v", format, err)
		}
		if first.String() != second.String() {
			t.Fatalf("Write(%s) is not deterministic", format)
		}
	}

	var jsonOutput bytes.Buffer
	if err := report.Write(&jsonOutput, value, report.FormatJSON); err != nil {
		t.Fatal(err)
	}
	var decoded report.Report
	if err := json.Unmarshal(jsonOutput.Bytes(), &decoded); err != nil || decoded.SchemaVersion != report.SchemaVersion {
		t.Fatalf("JSON report decode error = %v, value = %#v", err, decoded)
	}
	if decoded.Evidence[0].Status != report.EvidenceStatusFail {
		t.Fatalf("evidence status = %q", decoded.Evidence[0].Status)
	}
	if decoded.Diagnostics[0].Viewport != "mobile:390x844" {
		t.Fatalf("JSON viewport = %q", decoded.Diagnostics[0].Viewport)
	}
	var textOutput bytes.Buffer
	if err := report.Write(&textOutput, value, report.FormatText); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(textOutput.String(), "RULESET "+value.RuleSet.FingerprintSHA256) ||
		!strings.Contains(textOutput.String(), "PACK ansldes-core@1.0.0") ||
		!strings.Contains(textOutput.String(), `viewport="mobile:390x844" owner="ansldes/source"`) {
		t.Fatalf("text report = %s", textOutput.String())
	}
	var sarif bytes.Buffer
	if err := report.Write(&sarif, value, report.FormatSARIF); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sarif.String(), `"version": "2.1.0"`) ||
		!strings.Contains(sarif.String(), `"ruleId": "source/raw-value"`) ||
		!strings.Contains(sarif.String(), `"viewport": "mobile:390x844"`) ||
		!strings.Contains(sarif.String(), `"fingerprintSha256": "`+value.RuleSet.FingerprintSHA256+`"`) {
		t.Fatalf("SARIF = %s", sarif.String())
	}
}

func TestNewMergesCanonicalProvenanceAndPreservesFalsePositive(t *testing.T) {
	t.Parallel()
	first := diagnostic.NewWithSources(
		"visual/excessive-pill", []string{"impeccable/excessive-pill"}, diagnostic.SeverityWarning,
		"pill treatment", "src/A.tsx", nil, diagnostic.EvidenceWebSource, "web", "ansldes/web", "visual",
	)
	second := diagnostic.NewWithSources(
		"visual/excessive-pill", []string{"hallmark/rounded-ui"}, diagnostic.SeverityWarning,
		"pill treatment", "src/A.tsx", nil, diagnostic.EvidenceWebSource, "web", "ansldes/web", "visual",
	)
	falsePositive := report.NewFalsePositive(first, "example-owner", "approved exact component treatment")
	value := newReport(t, report.Input{
		DefinitionID: "example-product",
		Evidence: []report.EvidenceStatus{{
			Kind: diagnostic.EvidenceWebSource, Platform: "web", Status: report.EvidenceStatusPass,
		}},
		Diagnostics:    []diagnostic.Diagnostic{second, first},
		FalsePositives: []report.FalsePositive{falsePositive},
	})
	if len(value.Diagnostics) != 1 || strings.Join(value.Diagnostics[0].SourceRuleIDs, ",") != "hallmark/rounded-ui,impeccable/excessive-pill" {
		t.Fatalf("merged diagnostics = %#v", value.Diagnostics)
	}
	if len(value.FalsePositives) != 1 || value.FalsePositives[0].OwnerFingerprint == "" {
		t.Fatalf("false positives = %#v", value.FalsePositives)
	}
}

func TestOptionalJudgmentDoesNotChangeDeterministicFingerprintOrVerdict(t *testing.T) {
	t.Parallel()
	base := report.Input{
		DefinitionID: "example-product",
		Evidence: []report.EvidenceStatus{{
			Kind: diagnostic.EvidenceWebRendered, Platform: "web", Status: report.EvidenceStatusPass,
		}},
	}
	withoutReview := newReport(t, base)
	base.VisualJudgments = []report.VisualJudgment{{
		ID: "glassmorphism-everywhere", Status: report.JudgmentFail,
		EvidenceKind: diagnostic.EvidenceWebRendered, Platform: "web", Reviewer: "reviewer",
	}}
	withReview := newReport(t, base)
	if withReview.FingerprintSHA256 != withoutReview.FingerprintSHA256 || withReview.Status != withoutReview.Status {
		t.Fatalf("optional judgment changed deterministic result: %#v %#v", withoutReview, withReview)
	}
	if len(withReview.VisualJudgments) != 1 {
		t.Fatalf("visual judgments = %#v", withReview.VisualJudgments)
	}
}

func TestNewRejectsWebEvidenceSerializedAsNativePlatform(t *testing.T) {
	t.Parallel()
	ruleSet, err := report.NewActiveRuleSet("ansldes-core", "1.0.0", []string{"source/raw-value"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = report.New(report.Input{
		DefinitionID: "example-product",
		RuleSet:      ruleSet,
		Evidence: []report.EvidenceStatus{{
			Kind: diagnostic.EvidenceWebRendered, Platform: "react-native", Status: report.EvidenceStatusPass,
		}},
	})
	if err == nil {
		t.Fatal("New(web evidence on native platform) error = nil")
	}
}

func TestRuleSetFingerprintIgnoresInputOrdering(t *testing.T) {
	t.Parallel()
	left, err := report.NewActiveRuleSet("ansldes-core", "1.0.0", []string{"source/raw-value", "layout/problem"})
	if err != nil {
		t.Fatal(err)
	}
	right, err := report.NewActiveRuleSet("ansldes-core", "1.0.0", []string{"layout/problem", "source/raw-value"})
	if err != nil {
		t.Fatal(err)
	}
	if left.FingerprintSHA256 != right.FingerprintSHA256 {
		t.Fatalf("rule-set fingerprints differ: %s != %s", left.FingerprintSHA256, right.FingerprintSHA256)
	}
}

func TestRuleSetCanonicalizesEveryActivationStatus(t *testing.T) {
	t.Parallel()
	pack := report.NewRulePack("example-pack", "1.0.0", []string{"a/rule", "b/rule", "c/rule", "d/rule"})
	left, err := report.NewRuleSet([]report.RulePack{pack}, []report.RuleActivation{
		{RuleID: "d/rule", Status: report.RuleUnsupported, Reason: "provider unavailable"},
		{RuleID: "b/rule", Status: report.RuleNotApplicable, Reason: "evidence absent"},
		{RuleID: "a/rule", Status: report.RuleActive},
		{RuleID: "c/rule", Status: report.RuleDisabled, Reason: "governed override"},
	})
	if err != nil {
		t.Fatal(err)
	}
	right, err := report.NewRuleSet([]report.RulePack{pack}, []report.RuleActivation{
		{RuleID: "c/rule", Status: report.RuleDisabled, Reason: "governed override"},
		{RuleID: "a/rule", Status: report.RuleActive},
		{RuleID: "b/rule", Status: report.RuleNotApplicable, Reason: "evidence absent"},
		{RuleID: "d/rule", Status: report.RuleUnsupported, Reason: "provider unavailable"},
	})
	if err != nil || left.FingerprintSHA256 != right.FingerprintSHA256 || !reflect.DeepEqual(left, right) {
		t.Fatalf("rule sets differ: %#v %#v, %v", left, right, err)
	}
}

func TestInactiveActivationCannotReplaceFindingWithFalsePositive(t *testing.T) {
	t.Parallel()
	pack := report.NewRulePack("example-pack", "1.0.0", []string{"example/rule"})
	ruleSet, err := report.NewRuleSet([]report.RulePack{pack}, []report.RuleActivation{{RuleID: "example/rule", Status: report.RuleUnsupported, Reason: "provider unavailable"}})
	if err != nil {
		t.Fatal(err)
	}
	finding := diagnostic.New("example/rule", diagnostic.SeverityError, "finding", "src/example.tsx", nil, diagnostic.EvidenceNativeSource, "react-native", "example-owner", "example")
	_, err = report.New(report.Input{
		DefinitionID: "example-product", RuleSet: ruleSet,
		Evidence:       []report.EvidenceStatus{{Kind: diagnostic.EvidenceNativeSource, Platform: "react-native", Status: report.EvidenceStatusFalsePositive, Path: "src/example.tsx"}},
		FalsePositives: []report.FalsePositive{report.NewFalsePositive(finding, "example-owner", "approved exact exception")},
	})
	if err == nil {
		t.Fatal("New(inactive false positive) error = nil")
	}
}

func newReport(t *testing.T, input report.Input) report.Report {
	t.Helper()
	if len(input.RuleSet.Packs) == 0 {
		ruleIDs := []string{"source/raw-value"}
		seen := map[string]bool{"source/raw-value": true}
		for _, finding := range input.Diagnostics {
			if !seen[finding.RuleID] {
				ruleIDs = append(ruleIDs, finding.RuleID)
				seen[finding.RuleID] = true
			}
		}
		ruleSet, err := report.NewActiveRuleSet("ansldes-core", "1.0.0", ruleIDs)
		if err != nil {
			t.Fatal(err)
		}
		input.RuleSet = ruleSet
	}
	value, err := report.New(input)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
