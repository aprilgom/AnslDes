// Code generated from https://ansldes.dev/schema/design-system-definition.v1.json; DO NOT EDIT.
// definition schema SHA-256: 944c5b2ee33ae7e502509000bcd83bc7daa64009d71c8c99d52feebb43622708

package contract

import "encoding/json"

const DefinitionSchemaVersion = 1
const DefinitionSchemaSHA256 = "944c5b2ee33ae7e502509000bcd83bc7daa64009d71c8c99d52feebb43622708"

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
