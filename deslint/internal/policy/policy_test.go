package policy_test

import (
	"encoding/json"
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

func TestEvidenceKindsKeepLegacyAliasesAndRejectRequiredDeferredOverlap(t *testing.T) {
	t.Parallel()
	valid := string(readPolicy(t))
	legacy := strings.Replace(
		valid,
		`"native-source",\n      "design-document-source",\n      "design-document-computed"`,
		`"source",\n      "pencil",\n      "computed-layout"`,
		1,
	)
	legacyPolicy, err := policy.Parse([]byte(legacy))
	if err != nil {
		t.Fatalf("Parse(legacy evidence aliases) error = %v", err)
	}
	if !legacyPolicy.Requires(diagnostic.EvidenceNativeSource) ||
		!legacyPolicy.Requires(diagnostic.EvidenceDesignDocumentSource) ||
		!legacyPolicy.Requires(diagnostic.EvidenceDesignDocumentComputed) {
		t.Fatalf("legacy evidence aliases were not canonicalized: %#v", legacyPolicy.Evidence)
	}

	overlap := strings.Replace(valid, `"deferredKinds": []`, `"deferredKinds": ["native-source"]`, 1)
	if _, err := policy.Parse([]byte(overlap)); err == nil {
		t.Fatal("Parse(required/deferred overlap) error = nil")
	}
}

func TestApplyExceptionsRequiresExactActiveMatch(t *testing.T) {
	t.Parallel()
	productPolicy, err := policy.Parse(readPolicy(t))
	if err != nil {
		t.Fatal(err)
	}
	productPolicy.Exceptions = []policy.Exception{{
		RuleID: rules.RuleSourceRawValue, Engine: string(diagnostic.EvidenceNativeSource), Platform: "react-native",
		Path: "src/Exact.tsx", Owner: "owner", Rationale: "approved temporary migration",
		Reviewer: "example-reviewer", ExpiresAt: "2026-12-31", ReviewTrigger: "source owner or implementation changes",
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
	_, matches := productPolicy.ClassifyExceptions(findings, time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC))
	if len(matches) != 1 || matches[0].Finding.Path != "src/Exact.tsx" || matches[0].Exception.Owner != "owner" {
		t.Fatalf("ClassifyExceptions() = %#v", matches)
	}
	productPolicy.Exceptions[0].Owner = "other-owner"
	remaining, matches := productPolicy.ClassifyExceptions(findings, time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC))
	if len(remaining) != 2 || len(matches) != 0 {
		t.Fatalf("owner drift was suppressed: %#v %#v", remaining, matches)
	}
	productPolicy.Exceptions[0].Owner = "owner"
	productPolicy.Exceptions[0].Engine = string(diagnostic.EvidenceWebSource)
	remaining, matches = productPolicy.ClassifyExceptions(findings, time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC))
	if len(remaining) != 2 || len(matches) != 0 {
		t.Fatalf("provider drift was suppressed: %#v %#v", remaining, matches)
	}
}

func TestParseRejectsUnknownRuleOverrideAndPackFingerprint(t *testing.T) {
	t.Parallel()
	valid := string(readPolicy(t))
	unknown := strings.Replace(valid, `"ruleOverrides": []`, `"ruleOverrides": [{"ruleId":"profile/*","status":"disabled","owner":"owner","rationale":"broad override is invalid","expiresAt":"2026-12-31","reviewTrigger":"profile contract changes"}]`, 1)
	if _, err := policy.Parse([]byte(unknown)); err == nil {
		t.Fatal("Parse(wildcard rule override) error = nil")
	}
	wrongFingerprint := strings.Replace(valid, rules.CorePackFingerprint(), strings.Repeat("0", 64), 1)
	if _, err := policy.Parse([]byte(wrongFingerprint)); err == nil {
		t.Fatal("Parse(wrong rule pack fingerprint) error = nil")
	}
}

func TestParseRequiresExactPackAndReviewForDisabledRule(t *testing.T) {
	t.Parallel()
	var value map[string]any
	if err := json.Unmarshal(readPolicy(t), &value); err != nil {
		t.Fatal(err)
	}
	override := map[string]any{
		"ruleId": rules.RuleVisualSideTab, "packId": rules.AntiSlopPackID, "packVersion": rules.AntiSlopPackVersion,
		"status": "disabled", "owner": "example-owner", "rationale": "temporary exact migration",
		"reviewer": "example-reviewer", "expiresAt": "2026-12-31", "reviewTrigger": "provider implementation changes",
	}
	value["ruleOverrides"] = []any{override}
	if _, err := policy.Parse(marshal(t, value)); err != nil {
		t.Fatalf("Parse(exact override) = %v", err)
	}
	override["packId"] = rules.CorePackID
	override["packVersion"] = rules.CorePackVersion
	if _, err := policy.Parse(marshal(t, value)); err == nil {
		t.Fatal("Parse(cross-pack override) error = nil")
	}
}

func TestParseRequiresCustomProfileGovernance(t *testing.T) {
	t.Parallel()
	valid := string(readPolicy(t))
	custom := strings.Replace(valid, `"id": "operate"`, `"id": "specialized"`, 1)
	if _, err := policy.Parse([]byte(custom)); err == nil {
		t.Fatal("Parse(ungoverned custom profile) error = nil")
	}
}

func TestParsePreservesDetectorDefaultWithoutProfile(t *testing.T) {
	t.Parallel()
	productPolicy, err := policy.Parse(readPolicy(t))
	if err != nil {
		t.Fatal(err)
	}
	productPolicy.Profile = nil
	contents, err := json.Marshal(productPolicy)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := policy.Parse(contents)
	if err != nil || parsed.ProfileID() != "" {
		t.Fatalf("Parse(default detector profile) = %#v, %v", parsed.Profile, err)
	}
}

func TestParseValidatesWebRegistryReferencesAndExactArtifacts(t *testing.T) {
	t.Parallel()
	var value map[string]any
	if err := json.Unmarshal(readPolicy(t), &value); err != nil {
		t.Fatal(err)
	}
	web := value["web"].(map[string]any)
	captures := web["requiredCaptures"].([]any)
	captures[0].(map[string]any)["viewportId"] = "undeclared"
	if _, err := policy.Parse(marshal(t, value)); err == nil {
		t.Fatal("Parse(undeclared Web viewport) error = nil")
	}

	if err := json.Unmarshal(readPolicy(t), &value); err != nil {
		t.Fatal(err)
	}
	web = value["web"].(map[string]any)
	exclusions := web["artifactExclusions"].([]any)
	exclusions[0].(map[string]any)["path"] = "generated/**"
	if _, err := policy.Parse(marshal(t, value)); err == nil {
		t.Fatal("Parse(broad Web artifact exclusion) error = nil")
	}
}

func TestParseRejectsGovernanceBypassBroadIgnoreAndPackOmission(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(map[string]any){
		"missing forbidden flag": func(value map[string]any) {
			governance := value["governance"].(map[string]any)
			governance["forbiddenFlags"] = []any{"--no-config"}
		},
		"broad ignore": func(value map[string]any) {
			governance := value["governance"].(map[string]any)
			governance["ignores"] = []any{map[string]any{
				"kind": "file", "engine": "native-source", "platform": "react-native", "path": "src/**", "owner": "example-owner",
				"rationale": "temporary exact migration", "reviewer": "example-reviewer", "expiresAt": "2026-12-31", "reviewTrigger": "source ownership changes",
			}}
		},
		"pack omission": func(value map[string]any) {
			packs := value["rulePacks"].([]any)
			value["rulePacks"] = packs[1:]
		},
	} {
		t.Run(name, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(readPolicy(t), &value); err != nil {
				t.Fatal(err)
			}
			mutate(value)
			if _, err := policy.Parse(marshal(t, value)); err == nil {
				t.Fatalf("Parse(%s) error = nil", name)
			}
		})
	}
}

func TestIgnoreAllowanceDoesNotCrossProviderOrOwner(t *testing.T) {
	t.Parallel()
	productPolicy, err := policy.Parse(readPolicy(t))
	if err != nil {
		t.Fatal(err)
	}
	productPolicy.Governance.Ignores = []policy.IgnoreAllowance{{
		Kind: "inline", RuleID: rules.RuleVisualSideTab, Engine: string(diagnostic.EvidenceNativeSource), Platform: "react-native",
		Path: "src/Example.tsx", Owner: "example-owner", Rationale: "temporary exact source allowance", Reviewer: "example-reviewer",
		ExpiresAt: "2026-12-31", ReviewTrigger: "source implementation changes",
	}}
	request := policy.IgnoreRequest{
		Kind: "inline", RuleID: rules.RuleVisualSideTab, Engine: string(diagnostic.EvidenceNativeSource), Platform: "react-native",
		Path: "src/Example.tsx", Owner: "example-owner",
	}
	if !productPolicy.AuthorizesIgnore(request, time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("exact ignore was not authorized")
	}
	request.Engine = string(diagnostic.EvidenceWebSource)
	if productPolicy.AuthorizesIgnore(request, time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("native ignore crossed into Web provider")
	}
	request.Engine = string(diagnostic.EvidenceNativeSource)
	request.Owner = "drifted-owner"
	if productPolicy.AuthorizesIgnore(request, time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("ignore authorized a drifted owner")
	}
}

func marshal(t *testing.T, value any) []byte {
	t.Helper()
	contents, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func readPolicy(t *testing.T) []byte {
	t.Helper()
	contents, err := os.ReadFile("../../../packages/schema/testdata/example-policy.json")
	if err != nil {
		t.Fatal(err)
	}
	return contents
}
