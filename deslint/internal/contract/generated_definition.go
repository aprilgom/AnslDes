// Code generated from https://ansldes.dev/schema/design-system-definition.v2.json; DO NOT EDIT.
// definition schema SHA-256: 6d72afbbcbd876ec8bfbaf5c9451c27067ea411ccb4865da302a85b7ce9cc041

package contract

import "encoding/json"

const DefinitionSchemaVersion = 2
const DefinitionSchemaSHA256 = "6d72afbbcbd876ec8bfbaf5c9451c27067ea411ccb4865da302a85b7ce9cc041"

type Definition struct {
	Schema        string                         `json:"$schema,omitempty"`
	SchemaVersion int                            `json:"schemaVersion"`
	ID            string                         `json:"id"`
	Version       string                         `json:"version"`
	Themes        ThemeDefinition                `json:"themes"`
	ColorUsage    *ColorUsageDefinition          `json:"colorUsage,omitempty"`
	Foundations   map[string]json.RawMessage     `json:"foundations"`
	Components    map[string]ComponentDefinition `json:"components"`
}

type ColorUsageDefinition struct {
	Contrast         ContrastDefinition         `json:"contrast"`
	ApprovedPalettes map[string]ApprovedPalette `json:"approvedPalettes"`
}

type ContrastDefinition struct {
	Body  float64 `json:"body"`
	Large float64 `json:"large"`
}

type ApprovedPalette struct {
	Contexts []string `json:"contexts"`
	Themes   []string `json:"themes"`
}

type ThemeDefinition struct {
	Names   []string `json:"names"`
	Default string   `json:"default"`
}

type ComponentDefinition struct {
	Anatomy   []string                   `json:"anatomy"`
	Slots     map[string]json.RawMessage `json:"slots"`
	Variants  map[string]json.RawMessage `json:"variants"`
	Sizes     map[string]json.RawMessage `json:"sizes"`
	States    []string                   `json:"states"`
	Semantics map[string]json.RawMessage `json:"semantics"`
}
