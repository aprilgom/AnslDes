package rules

import (
	"fmt"
	"slices"
	"sort"
)

// RulePackManifest declares exact rule membership without embedding evaluator dispatch.
type RulePackManifest struct {
	ID                string
	Version           string
	FingerprintSHA256 string
	Members           []string
}

// ProviderCapability declares the evidence and platform tuples one provider can supply.
type ProviderCapability struct {
	ID            string
	EvidenceKinds []string
	Platforms     []string
}

// ComposeRulePacks deterministically composes manifests against independent implementation and provider registries.
func ComposeRulePacks(manifests []RulePackManifest, implementations map[string]RuleSpec, providers []ProviderCapability, availableDependencies []string) ([]RuleSpec, error) {
	ordered := append([]RulePackManifest(nil), manifests...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].ID != ordered[j].ID {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].Version < ordered[j].Version
	})
	providerByID := make(map[string]ProviderCapability, len(providers))
	for _, provider := range providers {
		if provider.ID == "" || providerByID[provider.ID].ID != "" {
			return nil, fmt.Errorf("provider capability %q is empty or duplicated", provider.ID)
		}
		providerByID[provider.ID] = provider
	}
	dependencies := make(map[string]bool, len(availableDependencies))
	for _, dependency := range availableDependencies {
		dependencies[dependency] = true
	}
	seenPacks := map[string]bool{}
	seenRules := map[string]bool{}
	result := make([]RuleSpec, 0)
	for _, manifest := range ordered {
		packKey := manifest.ID + "@" + manifest.Version
		if manifest.ID == "" || manifest.Version == "" || seenPacks[packKey] {
			return nil, fmt.Errorf("rule pack %q is empty or duplicated", packKey)
		}
		seenPacks[packKey] = true
		members := append([]string(nil), manifest.Members...)
		sort.Strings(members)
		fingerprintPack := RulePackSpec{ID: manifest.ID, Version: manifest.Version, Rules: make([]RuleSpec, 0, len(members))}
		for _, member := range members {
			fingerprintPack.Rules = append(fingerprintPack.Rules, RuleSpec{ID: member})
		}
		if manifest.FingerprintSHA256 != PackFingerprint(fingerprintPack) {
			return nil, fmt.Errorf("rule pack %s fingerprint is stale", packKey)
		}
		for _, member := range members {
			spec, found := implementations[member]
			if !found || spec.ID != member {
				return nil, fmt.Errorf("rule pack %s contains unknown member %q", packKey, member)
			}
			if seenRules[member] {
				return nil, fmt.Errorf("rule %q is duplicated across composed packs", member)
			}
			seenRules[member] = true
			for _, dependency := range spec.Dependencies {
				if !dependencies[dependency] {
					return nil, fmt.Errorf("rule %q is missing dependency %q", member, dependency)
				}
			}
			if len(spec.Providers) > 0 && !hasCompatibleProvider(spec, providerByID) {
				return nil, fmt.Errorf("rule %q has no compatible provider", member)
			}
			result = append(result, cloneRuleSpec(spec))
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func hasCompatibleProvider(spec RuleSpec, providers map[string]ProviderCapability) bool {
	for _, providerID := range spec.Providers {
		provider, found := providers[providerID]
		if found && intersects(spec.EvidenceKinds, provider.EvidenceKinds) && intersects(spec.Platforms, provider.Platforms) {
			return true
		}
	}
	return false
}

func intersects(left, right []string) bool {
	for _, value := range left {
		if slices.Contains(right, value) {
			return true
		}
	}
	return false
}
