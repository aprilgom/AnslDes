// Code generated from https://ansldes.dev/schema/design-system-definition.v2.json; DO NOT EDIT.
// definition schema SHA-256: 1e1d8c99aa9dab3825860c297a2c6a9086c3aba78162552a20d4ef75fbae504e

package contract

import "encoding/json"

const DefinitionSchemaVersion = 2
const DefinitionSchemaSHA256 = "1e1d8c99aa9dab3825860c297a2c6a9086c3aba78162552a20d4ef75fbae504e"

type Definition struct {
	Schema        string                         `json:"$schema,omitempty"`
	SchemaVersion int                            `json:"schemaVersion"`
	ID            string                         `json:"id"`
	Version       string                         `json:"version"`
	Themes        ThemeDefinition                `json:"themes"`
	Foundations   map[string]json.RawMessage     `json:"foundations"`
	Components    map[string]ComponentDefinition `json:"components"`
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
