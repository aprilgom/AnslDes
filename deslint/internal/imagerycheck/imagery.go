// Package imagerycheck evaluates asset geometry, provenance, loading, and accessibility evidence.
package imagerycheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Evidence is one provider-specific imagery inventory.
type Evidence struct {
	Schema               string                  `json:"$schema,omitempty"`
	SchemaVersion        int                     `json:"schemaVersion"`
	EvidenceKind         diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform             string                  `json:"platform"`
	SurfaceID            string                  `json:"surfaceId"`
	AssetRegistryVersion string                  `json:"assetRegistryVersion"`
	Nodes                []Node                  `json:"nodes"`
}

// Node is one image, poster, native image, or inline SVG observation.
type Node struct {
	ID                         string  `json:"id"`
	Owner                      string  `json:"owner"`
	Consumer                   string  `json:"consumer"`
	AssetID                    string  `json:"assetId,omitempty"`
	Medium                     string  `json:"medium"`
	Role                       string  `json:"role"`
	Width                      float64 `json:"width"`
	Height                     float64 `json:"height"`
	PrimitiveShapeCount        int     `json:"primitiveShapeCount"`
	PictorialScene             bool    `json:"pictorialScene"`
	DeliberatelyDrawn          bool    `json:"deliberatelyDrawn"`
	SourceStatus               string  `json:"sourceStatus"`
	Source                     string  `json:"source"`
	FingerprintSHA256          string  `json:"fingerprintSha256"`
	Decorative                 bool    `json:"decorative"`
	ScreenReaderExcluded       bool    `json:"screenReaderExcluded"`
	AccessibleLabel            string  `json:"accessibleLabel"`
	GradientOrShapePlaceholder bool    `json:"gradientOrShapePlaceholder"`
}

// Asset is one exact consumer-owned registry entry.
type Asset struct {
	Owner                string
	Role                 string
	ImplementationSource string
	Consumers            []string
	FingerprintSHA256    string
	IntentionallyOmitted bool
	Decorative           bool
}

// Config injects the versioned consumer asset registry.
type Config struct {
	RegistryVersion string
	Assets          map[string]Asset
	Severity        func(string) diagnostic.Severity
	Active          func(string) bool
}

var ruleIDs = []string{rules.RuleImageryShapeAssembledIllustration, rules.RuleImageryBrokenImage}

// Analyze strictly parses and evaluates imagery evidence.
func Analyze(path string, contents []byte, config Config) (Evidence, []diagnostic.Diagnostic, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, fmt.Errorf("imagery evidence has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, nil, err
	}
	if evidence.SchemaVersion != 1 || evidence.SurfaceID == "" || evidence.AssetRegistryVersion == "" || len(evidence.Nodes) == 0 {
		return Evidence{}, nil, fmt.Errorf("imagery evidence identity, asset registry version, and nodes are required")
	}
	if !platformMatches(evidence.EvidenceKind, evidence.Platform) {
		return Evidence{}, nil, fmt.Errorf("imagery evidence kind %q is incompatible with platform %q", evidence.EvidenceKind, evidence.Platform)
	}
	if config.RegistryVersion == "" || evidence.AssetRegistryVersion != config.RegistryVersion {
		return Evidence{}, nil, fmt.Errorf("imagery evidence asset registry version %q does not match consumer policy %q", evidence.AssetRegistryVersion, config.RegistryVersion)
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
	findings := []diagnostic.Diagnostic{}
	for _, node := range nodes {
		if err := validateNode(node); err != nil {
			return Evidence{}, nil, fmt.Errorf("imagery node %q: %w", node.ID, err)
		}
		if seen[node.ID] {
			return Evidence{}, nil, fmt.Errorf("imagery evidence has duplicate node identity %q", node.ID)
		}
		seen[node.ID] = true
		nodePath := path + "#/nodes/" + node.ID
		shapeScene := node.Medium == "inline-svg" && node.Role == "hero-illustration" && node.Width >= 240 && node.Height >= 180 && node.PrimitiveShapeCount >= 8 && node.PictorialScene && !node.DeliberatelyDrawn
		findings = add(findings, shapeScene, rules.RuleImageryShapeAssembledIllustration, "hero-sized pictorial SVG is assembled from generic primitive shapes", nodePath, node, evidence, config)

		asset, registered := config.Assets[node.AssetID]
		exactOmission := registered && asset.IntentionallyOmitted && node.SourceStatus == "intentionally-omitted" && registryMatches(node, asset)
		brokenStatus := node.SourceStatus == "missing" || node.SourceStatus == "empty" || node.SourceStatus == "placeholder" || node.SourceStatus == "load-failed"
		findings = add(findings, brokenStatus && !exactOmission, rules.RuleImageryBrokenImage, "image source is missing, empty, placeholder, or failed to load", nodePath, node, evidence, config)
		findings = add(findings, node.SourceStatus == "intentionally-omitted" && !exactOmission, rules.RuleImageryBrokenImage, "omitted illustration has no exact consumer registry permission", nodePath, node, evidence, config)
		findings = add(findings, node.AssetID != "" && (!registered || !registryMatches(node, asset)), rules.RuleImageryBrokenImage, "asset owner, source, consumer, role, or fingerprint drifted from the registry", nodePath, node, evidence, config)
		findings = add(findings, node.GradientOrShapePlaceholder && (!registered || !asset.IntentionallyOmitted), rules.RuleImageryBrokenImage, "required asset is replaced by a gradient or shape placeholder", nodePath, node, evidence, config)
		findings = add(findings, node.Decorative && !node.ScreenReaderExcluded, rules.RuleImageryBrokenImage, "decorative asset is exposed to screen readers", nodePath, node, evidence, config)
		findings = add(findings, !node.Decorative && strings.TrimSpace(node.AccessibleLabel) == "", rules.RuleImageryBrokenImage, "functional image has no accessible label", nodePath, node, evidence, config)
	}
	diagnostic.Sort(findings)
	return evidence, diagnostic.MergeCanonical(findings), nil
}

// RuleIDs returns the exact two-rule imagery membership.
func RuleIDs() []string { return append([]string(nil), ruleIDs...) }

func registryMatches(node Node, asset Asset) bool {
	return node.Owner == asset.Owner && node.Role == asset.Role && node.Source == asset.ImplementationSource &&
		node.FingerprintSHA256 == asset.FingerprintSHA256 && node.Decorative == asset.Decorative && slices.Contains(asset.Consumers, node.Consumer)
}

func platformMatches(kind diagnostic.EvidenceKind, platform string) bool {
	switch kind {
	case diagnostic.EvidenceWebSource, diagnostic.EvidenceWebRendered:
		return platform == "web"
	case diagnostic.EvidenceNativeSource:
		return platform == "react-native"
	case diagnostic.EvidenceDesignDocumentSource, diagnostic.EvidenceDesignDocumentComputed:
		return platform == "design-document"
	case diagnostic.EvidenceSimulator:
		return platform == "ios"
	case diagnostic.EvidenceEmulator:
		return platform == "android"
	case diagnostic.EvidencePhysicalDevice:
		return platform == "ios" || platform == "android"
	case diagnostic.EvidenceDefinition, diagnostic.EvidenceConsumerConformance,
		diagnostic.EvidenceConsumerContentRegistry, diagnostic.EvidenceExecution:
		return false
	}
	return false
}

func validateNode(node Node) error {
	if node.ID == "" || node.Owner == "" || node.Consumer == "" {
		return fmt.Errorf("id, owner, and consumer are required")
	}
	if !slices.Contains([]string{"img", "video-poster", "native-image", "inline-svg"}, node.Medium) ||
		!slices.Contains([]string{"icon", "logo", "data-diagram", "hero-illustration", "photo", "video-poster"}, node.Role) ||
		!slices.Contains([]string{"missing", "empty", "placeholder", "loaded", "load-failed", "intentionally-omitted"}, node.SourceStatus) {
		return fmt.Errorf("contains an unknown enum value")
	}
	if node.Width < 0 || node.Height < 0 || node.PrimitiveShapeCount < 0 || (node.FingerprintSHA256 != "" && len(node.FingerprintSHA256) != 64) {
		return fmt.Errorf("contains invalid geometry or fingerprint")
	}
	return nil
}

func add(findings []diagnostic.Diagnostic, condition bool, ruleID, message, path string, node Node, evidence Evidence, config Config) []diagnostic.Diagnostic {
	if !condition || !config.Active(ruleID) {
		return findings
	}
	return append(findings, diagnostic.NewWithSources(ruleID, []string{strings.TrimPrefix(ruleID, "imagery/")}, config.Severity(ruleID), message, path, nil, evidence.EvidenceKind, evidence.Platform, node.Owner, "imagery"))
}
