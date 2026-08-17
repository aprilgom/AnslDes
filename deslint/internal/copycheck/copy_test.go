package copycheck

import (
	"os"
	"slices"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

func TestKoreanFixtureMapsFiveLocaleAwareRulesWithoutEnglishPunctuation(t *testing.T) {
	evidence, findings, err := Analyze("ko.json", read(t, "../../../packages/schema/testdata/copy-ko-negative.json"), koreanConfig())
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Locale != "ko-KR" || hasRule(findings, rules.RuleCopyEmDashOveruse) {
		t.Fatalf("evidence/findings = %#v %#v", evidence, findings)
	}
	for _, ruleID := range []string{rules.RuleCopyMarketingBuzzword, rules.RuleCopyAphoristicCadence, rules.RuleCopyRepeatedContainerText, rules.RuleCopyTheaterSlopPhrase, rules.RuleCopyUnverifiedSocialProof} {
		if !hasRule(findings, ruleID) {
			t.Fatalf("missing %s in %#v", ruleID, findings)
		}
	}
	for _, finding := range findings {
		if finding.RuleID == rules.RuleCopyUnverifiedSocialProof && (!slices.Contains(finding.SourceRuleIDs, "hallmark-eight-08") || finding.EvidenceKind != diagnostic.EvidenceConsumerContentRegistry) {
			t.Fatalf("social proof provenance = %#v", finding)
		}
	}
}

func TestEnglishEmDashSaturationIsAdvisory(t *testing.T) {
	_, findings, err := Analyze("en.json", read(t, "../../../packages/schema/testdata/copy-en-advisory.json"), englishConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].RuleID != rules.RuleCopyEmDashOveruse || findings[0].Status != diagnostic.FindingAdvisory || findings[0].Severity != diagnostic.SeverityWarning {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestProtectedRecoveryAndVerifiedClaimAreNegativeFixtures(t *testing.T) {
	_, findings, err := Analyze("positive.json", read(t, "../../../packages/schema/testdata/copy-ko-positive.json"), koreanConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestMissingContentRegistryIsNotAClaimPass(t *testing.T) {
	evidence, findings, err := Analyze("not-run.json", read(t, "../../../packages/schema/testdata/copy-registry-not-run.json"), koreanConfig())
	if err != nil {
		t.Fatal(err)
	}
	if evidence.ContentRegistryStatus != "not-run" || hasRule(findings, rules.RuleCopyUnverifiedSocialProof) {
		t.Fatalf("evidence/findings = %#v %#v", evidence, findings)
	}
}

func koreanConfig() Config {
	return Config{
		ProfileID:       "operate",
		RegistryVersion: "1.0.0",
		LocalePolicy: LocalePolicy{
			PhraseRegistryVersion: "1.0.0",
			MarketingBuzzwords:    []string{"혁신적인", "원활한"},
			TheaterPhrases:        []string{"보안 연극"},
			ProtectedTerms:        []string{"복구 지점"},
			RecoveryCopyIDs:       []string{"recovery.retry"},
		},
		SourceReferences: []string{"source.metrics.active-users"},
	}
}

func englishConfig() Config {
	config := koreanConfig()
	config.ProfileID = "operate"
	config.LocalePolicy = LocalePolicy{
		PhraseRegistryVersion: "1.0.0",
		MarketingBuzzwords:    []string{"seamless", "revolutionary"},
		TheaterPhrases:        []string{"security theater"},
		ProtectedTerms:        []string{"recovery point"},
		RecoveryCopyIDs:       []string{"recovery.retry"},
	}
	return config
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
