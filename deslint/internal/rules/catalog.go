package rules

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// AntiSlopCatalogRule is one machine-readable Impeccable/Hallmark rule registration.
type AntiSlopCatalogRule struct {
	ID                    string   `json:"id"`
	ImplementationVersion string   `json:"implementationVersion"`
	Category              string   `json:"category"`
	Scopes                []string `json:"scopes"`
	Platforms             []string `json:"platforms"`
	EvidenceKinds         []string `json:"evidenceKinds"`
	DefaultSeverity       string   `json:"defaultSeverity"`
	UpstreamAdvisory      bool     `json:"upstreamAdvisory"`
	UpstreamSeverity      string   `json:"upstreamSeverity"`
	Provenance            []string `json:"provenance"`
	SourceRuleIDs         []string `json:"sourceRuleIds"`
	Dependencies          []string `json:"dependencies"`
	Providers             []string `json:"providers"`
}

type antiSlopCatalogFile struct {
	SchemaVersion int `json:"schemaVersion"`
	MigrationNote struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"migrationNote"`
	Impeccable struct {
		Name               string `json:"name"`
		Version            string `json:"version"`
		Integrity          string `json:"integrity"`
		Commit             string `json:"commit"`
		SourceTreeSHA256   string `json:"sourceTreeSha256"`
		TreeSHA256         string `json:"treeSha256"`
		ComputedTreeSHA256 string `json:"computedTreeSha256"`
	} `json:"impeccable"`
	Hallmark struct {
		Commit       string `json:"commit"`
		SourceSHA256 string `json:"sourceSha256"`
		Mappings     []struct {
			SourceID          string `json:"sourceId"`
			CanonicalSourceID string `json:"canonicalSourceId"`
		} `json:"mappings"`
	} `json:"hallmark"`
	Pack struct {
		ID                string `json:"id"`
		Version           string `json:"version"`
		FingerprintSHA256 string `json:"fingerprintSha256"`
	} `json:"pack"`
	Rules []AntiSlopCatalogRule `json:"rules"`
}

//go:embed anti_slop_catalog.json
var antiSlopCatalogJSON []byte

var loadedAntiSlopCatalog = mustLoadAntiSlopCatalog()

// AntiSlopCatalog returns a defensive copy of the exact 63-rule catalog.
func AntiSlopCatalog() []AntiSlopCatalogRule {
	result := make([]AntiSlopCatalogRule, 0, len(loadedAntiSlopCatalog.Rules))
	for _, rule := range loadedAntiSlopCatalog.Rules {
		result = append(result, cloneCatalogRule(rule))
	}
	return result
}

// LookupSourceRule resolves one exact namespaced upstream source ID.
func LookupSourceRule(sourceRuleID string) (AntiSlopCatalogRule, bool) {
	for _, rule := range loadedAntiSlopCatalog.Rules {
		if slices.Contains(rule.SourceRuleIDs, sourceRuleID) {
			return cloneCatalogRule(rule), true
		}
	}
	return AntiSlopCatalogRule{}, false
}

func mustLoadAntiSlopCatalog() antiSlopCatalogFile {
	var catalog antiSlopCatalogFile
	if err := json.Unmarshal(antiSlopCatalogJSON, &catalog); err != nil {
		panic(fmt.Sprintf("decode embedded anti-slop catalog: %v", err))
	}
	if catalog.SchemaVersion != 1 || catalog.MigrationNote.Path == "" || len(catalog.MigrationNote.SHA256) != 64 || catalog.Impeccable.Name != "impeccable" || catalog.Impeccable.Version != "3.6.0" || catalog.Impeccable.Integrity == "" || catalog.Impeccable.TreeSHA256 != catalog.Impeccable.ComputedTreeSHA256 || len(catalog.Rules) != 63 || len(catalog.Hallmark.Mappings) != 8 {
		panic("embedded anti-slop catalog provenance or exact counts are invalid")
	}
	seenIDs := map[string]bool{}
	seenSources := map[string]bool{}
	for _, rule := range catalog.Rules {
		if rule.ID == "" || rule.ImplementationVersion == "" || rule.Category == "" || rule.DefaultSeverity == "" || len(rule.Platforms) == 0 || len(rule.EvidenceKinds) == 0 || len(rule.Provenance) == 0 || len(rule.SourceRuleIDs) == 0 || len(rule.Providers) == 0 || seenIDs[rule.ID] {
			panic("embedded anti-slop catalog has incomplete or duplicate rules")
		}
		seenIDs[rule.ID] = true
		for _, sourceID := range rule.SourceRuleIDs {
			if strings.HasPrefix(sourceID, "impeccable/") {
				if seenSources[sourceID] {
					panic("embedded anti-slop catalog has duplicate Impeccable source identity")
				}
				seenSources[sourceID] = true
			}
		}
	}
	if len(seenSources) != 59 {
		panic("embedded anti-slop catalog must map exactly 59 Impeccable source rules")
	}
	members := make([]string, 0, len(catalog.Rules))
	for _, rule := range catalog.Rules {
		members = append(members, rule.ID)
	}
	sort.Strings(members)
	pack := RulePackSpec{ID: catalog.Pack.ID, Version: catalog.Pack.Version, Rules: make([]RuleSpec, 0, len(members))}
	for _, id := range members {
		pack.Rules = append(pack.Rules, RuleSpec{ID: id})
	}
	if PackFingerprint(pack) != catalog.Pack.FingerprintSHA256 {
		panic("embedded anti-slop pack fingerprint is stale")
	}
	return catalog
}

func cloneCatalogRule(value AntiSlopCatalogRule) AntiSlopCatalogRule {
	value.Scopes = append([]string(nil), value.Scopes...)
	value.Platforms = append([]string(nil), value.Platforms...)
	value.EvidenceKinds = append([]string(nil), value.EvidenceKinds...)
	value.Provenance = append([]string(nil), value.Provenance...)
	value.SourceRuleIDs = append([]string(nil), value.SourceRuleIDs...)
	value.Dependencies = append([]string(nil), value.Dependencies...)
	value.Providers = append([]string(nil), value.Providers...)
	return value
}
