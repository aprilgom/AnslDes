// Package copycheck evaluates locale-aware structural copy evidence.
package copycheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Evidence is one locale-specific copy inventory.
type Evidence struct {
	Schema                string                  `json:"$schema,omitempty"`
	SchemaVersion         int                     `json:"schemaVersion"`
	EvidenceKind          diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform              string                  `json:"platform"`
	ProfileID             string                  `json:"profileId"`
	SurfaceID             string                  `json:"surfaceId"`
	Locale                string                  `json:"locale"`
	ContentRegistryStatus string                  `json:"contentRegistryStatus"`
	Nodes                 []Node                  `json:"nodes"`
}

// Node is one normalized copy literal and its consumer intent.
type Node struct {
	ID                 string `json:"id"`
	Owner              string `json:"owner"`
	ContainerID        string `json:"containerId"`
	ContentRole        string `json:"contentRole"`
	Intent             string `json:"intent"`
	Text               string `json:"text"`
	CadencePattern     string `json:"cadencePattern"`
	FeatureReference   string `json:"featureReference,omitempty"`
	RationaleReference string `json:"rationaleReference,omitempty"`
	SourceReference    string `json:"sourceReference,omitempty"`
	RecoveryCopyID     string `json:"recoveryCopyId,omitempty"`
}

// LocalePolicy is the versioned consumer phrase and recovery registry.
type LocalePolicy struct {
	PhraseRegistryVersion string
	MarketingBuzzwords    []string
	TheaterPhrases        []string
	ProtectedTerms        []string
	RecoveryCopyIDs       []string
}

// Config injects consumer policy without embedding product wording in the engine.
type Config struct {
	ProfileID        string
	RegistryVersion  string
	LocalePolicy     LocalePolicy
	LocalePolicies   map[string]LocalePolicy
	SourceReferences []string
	Severity         func(string) diagnostic.Severity
	Active           func(string) bool
}

var ruleIDs = []string{
	rules.RuleCopyEmDashOveruse,
	rules.RuleCopyMarketingBuzzword,
	rules.RuleCopyAphoristicCadence,
	rules.RuleCopyRepeatedContainerText,
	rules.RuleCopyTheaterSlopPhrase,
	rules.RuleCopyUnverifiedSocialProof,
}

// Analyze strictly parses and evaluates copy evidence.
func Analyze(path string, contents []byte, config Config) (Evidence, []diagnostic.Diagnostic, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, fmt.Errorf("copy evidence has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, nil, err
	}
	if evidence.SchemaVersion != 1 || evidence.ProfileID == "" || evidence.SurfaceID == "" || evidence.Locale == "" || len(evidence.Nodes) == 0 {
		return Evidence{}, nil, fmt.Errorf("copy evidence identity, locale, and nodes are required")
	}
	if !platformMatches(evidence.EvidenceKind, evidence.Platform) {
		return Evidence{}, nil, fmt.Errorf("copy evidence kind %q is incompatible with platform %q", evidence.EvidenceKind, evidence.Platform)
	}
	if evidence.ContentRegistryStatus != "executed" && evidence.ContentRegistryStatus != "not-run" {
		return Evidence{}, nil, fmt.Errorf("content registry status %q is invalid", evidence.ContentRegistryStatus)
	}
	if config.ProfileID != "" && evidence.ProfileID != config.ProfileID {
		return Evidence{}, nil, fmt.Errorf("copy evidence profile %q does not match policy profile %q", evidence.ProfileID, config.ProfileID)
	}
	localePolicy := config.LocalePolicy
	if configured, ok := config.LocalePolicies[evidence.Locale]; ok {
		localePolicy = configured
	}
	if config.RegistryVersion == "" || localePolicy.PhraseRegistryVersion == "" {
		return Evidence{}, nil, fmt.Errorf("copy evidence requires a versioned consumer content policy for locale %s", evidence.Locale)
	}
	if config.Severity == nil {
		config.Severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if config.Active == nil {
		config.Active = func(string) bool { return true }
	}

	nodes := append([]Node(nil), evidence.Nodes...)
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	seen := map[string]bool{}
	for _, node := range nodes {
		if err := validateNode(node); err != nil {
			return Evidence{}, nil, fmt.Errorf("copy node %q: %w", node.ID, err)
		}
		if seen[node.ID] {
			return Evidence{}, nil, fmt.Errorf("copy evidence has duplicate node identity %q", node.ID)
		}
		seen[node.ID] = true
	}

	findings := []diagnostic.Diagnostic{}
	if strings.HasPrefix(strings.ToLower(evidence.Locale), "en") {
		totalChars, totalDashes, first := 0, 0, Node{}
		for _, node := range nodes {
			if node.ContentRole != "body" {
				continue
			}
			if first.ID == "" {
				first = node
			}
			totalChars += utf8.RuneCountInString(node.Text)
			totalDashes += strings.Count(node.Text, "—") + strings.Count(node.Text, "--")
		}
		if totalDashes >= 8 && totalChars > 0 && totalChars <= totalDashes*500 && config.Active(rules.RuleCopyEmDashOveruse) {
			findings = append(findings, newFinding(rules.RuleCopyEmDashOveruse, fmt.Sprintf("English body copy uses %d em dashes across %d characters", totalDashes, totalChars), path+"#/nodes/"+first.ID, first, diagnostic.SeverityWarning, evidence.EvidenceKind, evidence.Platform, nil))
		}
	}

	byCadence := map[string][]Node{}
	byLiteral := map[string][]Node{}
	for _, node := range nodes {
		nodePath := path + "#/nodes/" + node.ID
		normalized := normalize(node.Text)
		if node.CadencePattern != "none" {
			byCadence[node.ContainerID] = append(byCadence[node.ContainerID], node)
		}
		byLiteral[node.ContainerID+"\x00"+normalized] = append(byLiteral[node.ContainerID+"\x00"+normalized], node)
		buzzword := containsAny(normalized, localePolicy.MarketingBuzzwords)
		protected := containsAny(normalized, localePolicy.ProtectedTerms) && node.RationaleReference != ""
		findings = add(findings, buzzword && node.Intent == "market" && node.FeatureReference == "" && node.RationaleReference == "" && !protected && node.ContentRole != "legal" && node.ContentRole != "domain-term", rules.RuleCopyMarketingBuzzword, "marketing phrase has no concrete feature or rationale reference", nodePath, node, evidence, config, nil)
		findings = add(findings, containsAny(normalized, localePolicy.TheaterPhrases), rules.RuleCopyTheaterSlopPhrase, "copy matches the versioned locale theater-phrase registry", nodePath, node, evidence, config, nil)
		claim := node.ContentRole == "metric" || node.ContentRole == "testimonial" || node.ContentRole == "logo-claim"
		verified := node.SourceReference != "" && slices.Contains(config.SourceReferences, node.SourceReference)
		if evidence.ContentRegistryStatus == "executed" {
			findings = add(findings, claim && !verified, rules.RuleCopyUnverifiedSocialProof, "claim has no stable source reference in the consumer content registry", nodePath, node, evidence, config, []string{"hallmark-eight-08"})
		}
	}

	for _, container := range sortedMapKeys(byCadence) {
		values := byCadence[container]
		if len(values) >= 3 {
			node := values[0]
			findings = add(findings, true, rules.RuleCopyAphoristicCadence, "container repeats manufactured-contrast or short-rebuttal cadence three or more times", path+"#/nodes/"+node.ID, node, evidence, config, nil)
		}
	}
	for _, key := range sortedMapKeys(byLiteral) {
		values := byLiteral[key]
		if len(values) < 2 || approvedRecovery(values, localePolicy.RecoveryCopyIDs) {
			continue
		}
		node := values[0]
		findings = add(findings, true, rules.RuleCopyRepeatedContainerText, "container repeats the same normalized literal across structural roles", path+"#/nodes/"+node.ID, node, evidence, config, nil)
	}
	diagnostic.Sort(findings)
	return evidence, diagnostic.MergeCanonical(findings), nil
}

// RuleIDs returns the exact six-rule copy membership.
func RuleIDs() []string { return append([]string(nil), ruleIDs...) }

func platformMatches(kind diagnostic.EvidenceKind, platform string) bool {
	switch kind {
	case diagnostic.EvidenceWebSource:
		return platform == "web"
	case diagnostic.EvidenceNativeSource:
		return platform == "react-native"
	case diagnostic.EvidenceDesignDocumentSource:
		return platform == "design-document"
	case diagnostic.EvidenceDefinition, diagnostic.EvidenceWebRendered, diagnostic.EvidenceDesignDocumentComputed,
		diagnostic.EvidenceSimulator, diagnostic.EvidenceEmulator, diagnostic.EvidencePhysicalDevice,
		diagnostic.EvidenceConsumerConformance, diagnostic.EvidenceConsumerContentRegistry, diagnostic.EvidenceExecution:
		return false
	}
	return false
}

func validateNode(node Node) error {
	if node.ID == "" || node.Owner == "" || node.ContainerID == "" || strings.TrimSpace(node.Text) == "" {
		return fmt.Errorf("id, owner, containerId, and text are required")
	}
	if !slices.Contains([]string{"heading", "body", "label", "question", "reason", "risk", "result", "next-action", "metric", "testimonial", "logo-claim", "recovery", "legal", "domain-term"}, node.ContentRole) ||
		!slices.Contains([]string{"inform", "instruct", "recover", "market", "claim"}, node.Intent) ||
		!slices.Contains([]string{"none", "manufactured-contrast", "short-rebuttal"}, node.CadencePattern) {
		return fmt.Errorf("contains an unknown enum value")
	}
	return nil
}

func normalize(value string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	}), " ")
}

func containsAny(normalized string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(normalized, normalize(phrase)) {
			return true
		}
	}
	return false
}

func approvedRecovery(nodes []Node, approved []string) bool {
	if len(nodes) == 0 || nodes[0].RecoveryCopyID == "" || !slices.Contains(approved, nodes[0].RecoveryCopyID) {
		return false
	}
	for _, node := range nodes[1:] {
		if node.RecoveryCopyID != nodes[0].RecoveryCopyID {
			return false
		}
	}
	return true
}

func sortedMapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func add(findings []diagnostic.Diagnostic, condition bool, ruleID, message, path string, node Node, evidence Evidence, config Config, extraSources []string) []diagnostic.Diagnostic {
	if !condition || !config.Active(ruleID) {
		return findings
	}
	evidenceKind, platform := evidence.EvidenceKind, evidence.Platform
	if ruleID == rules.RuleCopyUnverifiedSocialProof {
		evidenceKind, platform = diagnostic.EvidenceConsumerContentRegistry, "all"
	}
	return append(findings, newFinding(ruleID, message, path, node, config.Severity(ruleID), evidenceKind, platform, extraSources))
}

func newFinding(ruleID, message, path string, node Node, severity diagnostic.Severity, evidenceKind diagnostic.EvidenceKind, platform string, extraSources []string) diagnostic.Diagnostic {
	sources := append([]string{strings.TrimPrefix(ruleID, "copy/")}, extraSources...)
	return diagnostic.NewWithSources(ruleID, sources, severity, message, path, nil, evidenceKind, platform, node.Owner, "copy")
}
