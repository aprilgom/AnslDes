// Package designcontext parses generated design-system awareness sidecars.
package designcontext

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
)

// GeneratorID is the only accepted generated-context producer.
const GeneratorID = "ansldes-design-context"

// GeneratorVersion is the supported exact generator contract version.
const GeneratorVersion = "1.0.0"

// Context is the strict generated sidecar consumed by design-system rules.
type Context struct {
	Schema        string         `json:"$schema,omitempty"`
	SchemaVersion int            `json:"schemaVersion"`
	Generator     Generator      `json:"generator"`
	Source        Source         `json:"source"`
	Typography    Typography     `json:"typography"`
	Colors        ColorContract  `json:"colors"`
	Radii         RadiusContract `json:"radii"`
}

// Generator identifies the sidecar producer.
type Generator struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// Source binds the sidecar to one canonical definition fingerprint.
type Source struct {
	SchemaVersion  int    `json:"schemaVersion"`
	ID             string `json:"id"`
	Version        string `json:"version"`
	ContractSHA256 string `json:"contractSha256"`
}

// Typography contains canonical family, weight, token, and role contracts.
type Typography struct {
	FamilyName string               `json:"familyName"`
	Weights    map[string]Weight    `json:"weights"`
	Tokens     map[string]TypeToken `json:"tokens"`
	Roles      map[string]TypeRole  `json:"roles"`
}

// Weight binds a logical weight to a physical font and numeric weight.
type Weight struct {
	FontFamily string `json:"fontFamily"`
	FontWeight int    `json:"fontWeight"`
}

// TypeToken is one semantic typography ramp entry.
type TypeToken struct {
	FontSize              float64  `json:"fontSize"`
	LineHeight            float64  `json:"lineHeight"`
	MaxFontSizeMultiplier float64  `json:"maxFontSizeMultiplier"`
	AllowedWeights        []string `json:"allowedWeights"`
	Category              string   `json:"category"`
	FunctionalTextAllowed bool     `json:"functionalTextAllowed"`
}

// TypeRole selects one visual token and logical weight.
type TypeRole struct {
	VisualToken string `json:"visualToken"`
	Weight      string `json:"weight"`
}

// ColorContract preserves the canonical color layers.
type ColorContract struct {
	Primitive map[string]string            `json:"primitive"`
	Semantic  map[string]map[string]string `json:"semantic"`
	Component map[string]string            `json:"component"`
	Asset     map[string]string            `json:"asset"`
}

// RadiusContract preserves primitive, semantic, and component radii.
type RadiusContract struct {
	Primitive map[string]float64 `json:"primitive"`
	Semantic  map[string]string  `json:"semantic"`
	Component map[string]string  `json:"component"`
}

// Parse rejects malformed, unknown, duplicate, and unsupported sidecar data.
func Parse(contents []byte) (Context, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Context{}, fmt.Errorf("parse design context JSON: %w", err)
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Context{}, fmt.Errorf("design context has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var result Context
	if err := decoder.Decode(&result); err != nil {
		return Context{}, fmt.Errorf("decode design context: %w", err)
	}
	if result.SchemaVersion != 1 || result.Generator.ID != GeneratorID || result.Generator.Version != GeneratorVersion || len(result.Source.ContractSHA256) != 64 {
		return Context{}, fmt.Errorf("design context generator or source contract metadata is invalid")
	}
	return result, nil
}

// ContractFingerprint matches the compiler's canonical JSON SHA-256.
func ContractFingerprint(contents []byte) (string, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	canonical, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	canonical = append(canonical, '\n')
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}
