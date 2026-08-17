package rules

import (
	"reflect"
	"testing"
)

func TestRegistryReturnsStableDefensiveCopies(t *testing.T) {
	first := RegisteredPacks()
	second := RegisteredPacks()
	if !reflect.DeepEqual(first, second) || len(first) != 2 {
		t.Fatalf("registry differs: %#v %#v", first, second)
	}
	first[0].Rules[0].ID = "changed"
	third := RegisteredPacks()
	if third[0].Rules[0].ID == "changed" {
		t.Fatal("registry returned mutable shared rule specs")
	}
	if PackFingerprint(third[0]) != AntiSlopPackFingerprint() {
		t.Fatalf("fingerprints differ: %s %s", PackFingerprint(third[0]), AntiSlopPackFingerprint())
	}
}

func TestEveryRegisteredRuleHasApplicabilityMetadata(t *testing.T) {
	for _, ruleID := range AllRuleIDs {
		spec, found := Lookup(ruleID)
		if !found || len(spec.EvidenceKinds) == 0 || len(spec.Platforms) == 0 || spec.ImplementationVersion == "" || spec.Category == "" || spec.DefaultSeverity == "" || len(spec.Provenance) == 0 {
			t.Fatalf("incomplete RuleSpec for %q: %#v", ruleID, spec)
		}
	}
}

func TestAntiSlopCatalogHasExactProvenanceAndPackMembership(t *testing.T) {
	catalog := AntiSlopCatalog()
	pack, found := LookupPack(AntiSlopPackID, AntiSlopPackVersion)
	if !found || len(catalog) != 63 || len(pack.Rules) != 63 || AntiSlopPackFingerprint() != "df05187c79c254c8b7f1159b375ae2017e5d55b20af512aabf7f7b522173159f" {
		t.Fatalf("catalog/pack = %d %#v %s", len(catalog), pack, AntiSlopPackFingerprint())
	}
	upstream := 0
	for _, rule := range catalog {
		for _, sourceID := range rule.SourceRuleIDs {
			if len(sourceID) > len("impeccable/") && sourceID[:len("impeccable/")] == "impeccable/" {
				upstream++
			}
		}
	}
	if upstream != 59 {
		t.Fatalf("upstream source count = %d", upstream)
	}
}

func TestCanonicalRulesHaveExactNativeAndDesignDocumentMappings(t *testing.T) {
	pack, _ := LookupPack(AntiSlopPackID, AntiSlopPackVersion)
	if len(pack.Rules) != 63 {
		t.Fatalf("anti-slop rule count = %d", len(pack.Rules))
	}
	for _, spec := range pack.Rules {
		if len(spec.Applicability) != 3 {
			t.Fatalf("%s applicability = %#v", spec.ID, spec.Applicability)
		}
		seen := map[string]bool{}
		for _, mapping := range spec.Applicability {
			if seen[mapping.Target] || mapping.Reason == "" {
				t.Fatalf("%s mapping is incomplete or duplicated: %#v", spec.ID, mapping)
			}
			seen[mapping.Target] = true
			if mapping.Status == "unsupported" && len(mapping.AlternativeEvidence) == 0 {
				t.Fatalf("%s unsupported mapping lacks alternative evidence: %#v", spec.ID, mapping)
			}
		}
		for _, target := range []string{"web", "react-native", "design-document"} {
			if !seen[target] {
				t.Fatalf("%s mapping lacks %s", spec.ID, target)
			}
		}
	}
	spec, _ := Lookup(RuleTypographyIconTileStack)
	foundSupplement := false
	for _, mapping := range spec.Applicability {
		if mapping.Target == "react-native" && mapping.Status == "native-supplement" && reflect.DeepEqual(mapping.SupplementRuleIDs, []string{RuleNativeListRowAccessoryWrapper}) {
			foundSupplement = true
		}
	}
	if !foundSupplement {
		t.Fatalf("native supplement mapping = %#v", spec.Applicability)
	}
}

func TestComposerAddsRemovesAndReplacesRulesWithoutDispatchChanges(t *testing.T) {
	implementations := map[string]RuleSpec{
		"synthetic/alpha": {ID: "synthetic/alpha", EvidenceKinds: []string{"web-source"}, Platforms: []string{"web"}, Providers: []string{"synthetic"}, Dependencies: []string{"parser"}},
		"synthetic/beta":  {ID: "synthetic/beta", EvidenceKinds: []string{"web-rendered"}, Platforms: []string{"web"}, Providers: []string{"synthetic"}},
	}
	providers := []ProviderCapability{{ID: "synthetic", EvidenceKinds: []string{"web-source", "web-rendered"}, Platforms: []string{"web"}}}
	manifest := manifestFor("synthetic-pack", "1.0.0", []string{"synthetic/alpha"})
	first, err := ComposeRulePacks([]RulePackManifest{manifest}, implementations, providers, []string{"parser"})
	if err != nil || len(first) != 1 || first[0].ID != "synthetic/alpha" {
		t.Fatalf("first composition = %#v %v", first, err)
	}
	replaced := manifestFor("synthetic-pack", "1.1.0", []string{"synthetic/beta"})
	second, err := ComposeRulePacks([]RulePackManifest{replaced}, implementations, providers, []string{"parser"})
	if err != nil || len(second) != 1 || second[0].ID != "synthetic/beta" {
		t.Fatalf("replacement = %#v %v", second, err)
	}
}

func TestComposerRejectsDuplicateUnknownDependencyAndProviderDrift(t *testing.T) {
	spec := RuleSpec{ID: "synthetic/alpha", EvidenceKinds: []string{"web-source"}, Platforms: []string{"web"}, Providers: []string{"source"}, Dependencies: []string{"parser"}}
	implementations := map[string]RuleSpec{spec.ID: spec}
	manifest := manifestFor("one", "1.0.0", []string{spec.ID})
	providers := []ProviderCapability{{ID: "source", EvidenceKinds: []string{"web-source"}, Platforms: []string{"web"}}}
	if _, err := ComposeRulePacks([]RulePackManifest{manifest}, implementations, providers, nil); err == nil {
		t.Fatal("missing dependency error = nil")
	}
	unknown := manifestFor("unknown", "1.0.0", []string{"synthetic/missing"})
	if _, err := ComposeRulePacks([]RulePackManifest{unknown}, implementations, providers, []string{"parser"}); err == nil {
		t.Fatal("unknown member error = nil")
	}
	wrongProvider := []ProviderCapability{{ID: "source", EvidenceKinds: []string{"native-source"}, Platforms: []string{"react-native"}}}
	if _, err := ComposeRulePacks([]RulePackManifest{manifest}, implementations, wrongProvider, []string{"parser"}); err == nil {
		t.Fatal("incompatible provider error = nil")
	}
	other := manifestFor("two", "1.0.0", []string{spec.ID})
	if _, err := ComposeRulePacks([]RulePackManifest{manifest, other}, implementations, providers, []string{"parser"}); err == nil {
		t.Fatal("duplicate rule error = nil")
	}
}

func manifestFor(id, version string, members []string) RulePackManifest {
	pack := RulePackSpec{ID: id, Version: version, Rules: make([]RuleSpec, 0, len(members))}
	for _, member := range members {
		pack.Rules = append(pack.Rules, RuleSpec{ID: member})
	}
	return RulePackManifest{ID: id, Version: version, Members: members, FingerprintSHA256: PackFingerprint(pack)}
}
