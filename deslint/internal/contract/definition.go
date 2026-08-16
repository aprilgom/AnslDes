// Package contract validates language-neutral design-system definitions.
package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

var referencePattern = regexp.MustCompile(`^\{(color|spacing|radius|size|typography|motion|elevation|layer)\.([a-z]+)\.([a-z0-9][a-z0-9._-]*)\}$`)

// Analysis is the validated identity and deterministic diagnostics for a definition.
type Analysis struct {
	DefinitionID string
	Diagnostics  []diagnostic.Diagnostic
}

// Analyze checks schema version, structural keys, reference syntax, and reference existence.
func Analyze(path string, contents []byte, severity func(string) diagnostic.Severity) (Analysis, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Analysis{}, fmt.Errorf("parse definition JSON: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil {
		return Analysis{}, fmt.Errorf("decode definition: %w", err)
	}
	var typed Definition
	if err := json.Unmarshal(contents, &typed); err != nil {
		return Analysis{}, fmt.Errorf("decode definition model: %w", err)
	}

	diagnostics := make([]diagnostic.Diagnostic, 0)
	for _, duplicate := range duplicates {
		diagnostics = append(diagnostics, schemaDiagnostic(path, severity, "duplicate key "+duplicate))
	}
	validateExactKeys(root, []string{"$schema", "schemaVersion", "id", "version", "themes", "foundations", "components"}, path, severity, &diagnostics)
	if typed.SchemaVersion != DefinitionSchemaVersion {
		diagnostics = append(diagnostics, schemaDiagnostic(
			path,
			severity,
			fmt.Sprintf("schemaVersion = %d, want %d", typed.SchemaVersion, DefinitionSchemaVersion),
		))
	}
	if typed.ID == "" {
		diagnostics = append(diagnostics, schemaDiagnostic(path, severity, "id is required"))
	}
	foundations, ok := object(root["foundations"])
	if !ok {
		diagnostics = append(diagnostics, schemaDiagnostic(path, severity, "foundations must be an object"))
	} else {
		validateExactKeysWithOptional(
			foundations,
			[]string{"color", "spacing", "radius", "size", "typography", "motion", "elevation", "layer"},
			[]string{"icon"},
			path,
			severity,
			&diagnostics,
		)
	}
	if components, ok := object(root["components"]); !ok || len(components) == 0 {
		diagnostics = append(diagnostics, schemaDiagnostic(path, severity, "components must be a non-empty object"))
	}

	registry := referenceRegistry(foundations)
	validateStrictTokenLayers(foundations, root["components"], path, severity, &diagnostics)
	walkReferences(root, "$", registry, path, severity, &diagnostics)
	diagnostic.Sort(diagnostics)
	return Analysis{DefinitionID: typed.ID, Diagnostics: diagnostics}, nil
}

func validateStrictTokenLayers(foundations map[string]any, componentsValue any, path string, severity func(string) diagnostic.Severity, diagnostics *[]diagnostic.Diagnostic) {
	for _, collection := range []string{"color", "spacing", "radius", "size"} {
		group, _ := object(foundations[collection])
		primitive, _ := object(group["primitive"])
		for _, name := range sortedKeys(primitive) {
			value := primitive[name]
			if collection == "color" {
				literal, ok := value.(string)
				if !ok || strings.HasPrefix(literal, "{") {
					appendLayerDiagnostic(path, severity, fmt.Sprintf("foundations.%s.primitive.%s must be a raw color", collection, name), diagnostics)
				}
				continue
			}
			number, ok := value.(json.Number)
			if !ok {
				appendLayerDiagnostic(path, severity, fmt.Sprintf("foundations.%s.primitive.%s must be a raw number", collection, name), diagnostics)
				continue
			}
			if collection != "spacing" {
				numeric, numberErr := number.Float64()
				if numberErr != nil || numeric < 0 {
					appendLayerDiagnostic(path, severity, fmt.Sprintf("foundations.%s.primitive.%s must be non-negative", collection, name), diagnostics)
				}
			}
		}
		semantic, _ := object(group["semantic"])
		for _, name := range sortedKeys(semantic) {
			if collection == "color" {
				themes, ok := object(semantic[name])
				if !ok {
					appendLayerDiagnostic(path, severity, fmt.Sprintf("%s.semantic.%s must map themes to %s.primitive references", collection, name, collection), diagnostics)
					continue
				}
				for _, theme := range sortedKeys(themes) {
					validateLayerReference(themes[theme], collection, "primitive", fmt.Sprintf("foundations.%s.semantic.%s.%s", collection, name, theme), path, severity, diagnostics)
				}
				continue
			}
			validateLayerReference(semantic[name], collection, "primitive", fmt.Sprintf("foundations.%s.semantic.%s", collection, name), path, severity, diagnostics)
		}

		component, _ := object(group["component"])
		for _, name := range sortedKeys(component) {
			validateLayerReference(component[name], collection, "semantic", fmt.Sprintf("foundations.%s.component.%s", collection, name), path, severity, diagnostics)
		}
	}

	walkComponentLayerReferences(componentsValue, "components", path, severity, diagnostics)
}

func validateLayerReference(value any, collection string, expectedLayer string, jsonPath string, path string, severity func(string) diagnostic.Severity, diagnostics *[]diagnostic.Diagnostic) {
	reference, ok := value.(string)
	if !ok {
		appendLayerDiagnostic(path, severity, fmt.Sprintf("%s must reference %s.%s", jsonPath, collection, expectedLayer), diagnostics)
		return
	}
	match := referencePattern.FindStringSubmatch(reference)
	if match == nil || match[1] != collection || match[2] != expectedLayer {
		appendLayerDiagnostic(path, severity, fmt.Sprintf("%s must reference %s.%s", jsonPath, collection, expectedLayer), diagnostics)
	}
}

func walkComponentLayerReferences(value any, jsonPath string, path string, severity func(string) diagnostic.Severity, diagnostics *[]diagnostic.Diagnostic) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range sortedKeys(typed) {
			walkComponentLayerReferences(typed[key], jsonPath+"."+key, path, severity, diagnostics)
		}
	case []any:
		for index, item := range typed {
			walkComponentLayerReferences(item, fmt.Sprintf("%s[%d]", jsonPath, index), path, severity, diagnostics)
		}
	case string:
		match := referencePattern.FindStringSubmatch(typed)
		if match == nil {
			return
		}
		if (match[1] == "color" || match[1] == "spacing" || match[1] == "radius" || match[1] == "size") && match[2] != "component" {
			appendLayerDiagnostic(path, severity, fmt.Sprintf("%s must reference %s.component instead of %s.%s", jsonPath, match[1], match[1], match[2]), diagnostics)
		}
	}
}

func appendLayerDiagnostic(path string, severity func(string) diagnostic.Severity, message string, diagnostics *[]diagnostic.Diagnostic) {
	*diagnostics = append(*diagnostics, diagnostic.New(
		rules.RuleDefinitionInvalidRef,
		severity(rules.RuleDefinitionInvalidRef),
		message,
		path,
		nil,
		diagnostic.EvidenceDefinition,
		"all",
		"ansldes/contract",
		"token-layer",
	))
}

func sortedKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func walkReferences(value any, jsonPath string, registry map[string]struct{}, path string, severity func(string) diagnostic.Severity, diagnostics *[]diagnostic.Diagnostic) {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			walkReferences(typed[key], jsonPath+"."+key, registry, path, severity, diagnostics)
		}
	case []any:
		for index, item := range typed {
			walkReferences(item, fmt.Sprintf("%s[%d]", jsonPath, index), registry, path, severity, diagnostics)
		}
	case string:
		if !strings.HasPrefix(typed, "{") {
			return
		}
		match := referencePattern.FindStringSubmatch(typed)
		if match == nil {
			*diagnostics = append(*diagnostics, diagnostic.New(
				rules.RuleDefinitionInvalidRef,
				severity(rules.RuleDefinitionInvalidRef),
				fmt.Sprintf("invalid reference %q at %s", typed, jsonPath),
				path,
				nil,
				diagnostic.EvidenceDefinition,
				"all",
				"ansldes/contract",
				"reference",
			))
			return
		}
		identity := match[1] + "." + match[2] + "." + match[3]
		if _, exists := registry[identity]; !exists {
			*diagnostics = append(*diagnostics, diagnostic.New(
				rules.RuleDefinitionUnknownToken,
				severity(rules.RuleDefinitionUnknownToken),
				fmt.Sprintf("unknown token %s at %s", typed, jsonPath),
				path,
				nil,
				diagnostic.EvidenceDefinition,
				"all",
				"ansldes/contract",
				"reference",
			))
		}
	}
}

func referenceRegistry(foundations map[string]any) map[string]struct{} {
	registry := make(map[string]struct{})
	for _, collection := range []string{"color", "spacing", "radius", "size"} {
		group, _ := object(foundations[collection])
		for _, layer := range []string{"primitive", "semantic", "component", "asset"} {
			values, _ := object(group[layer])
			for name := range values {
				registry[collection+"."+layer+"."+name] = struct{}{}
			}
		}
	}
	typography, _ := object(foundations["typography"])
	for source, layer := range map[string]string{"roles": "role", "tokens": "token", "weights": "weight"} {
		values, _ := object(typography[source])
		for name := range values {
			registry["typography."+layer+"."+name] = struct{}{}
		}
	}
	motion, _ := object(foundations["motion"])
	for source, layer := range map[string]string{"durations": "duration", "easings": "easing", "transitions": "transition"} {
		values, _ := object(motion[source])
		for name := range values {
			registry["motion."+layer+"."+name] = struct{}{}
		}
	}
	elevation, _ := object(foundations["elevation"])
	for name := range elevation {
		registry["elevation.recipe."+name] = struct{}{}
	}
	layers, _ := object(foundations["layer"])
	for _, layer := range []string{"semantic", "component"} {
		values, _ := object(layers[layer])
		for name := range values {
			registry["layer."+layer+"."+name] = struct{}{}
		}
	}
	return registry
}

func validateExactKeys(value map[string]any, allowed []string, path string, severity func(string) diagnostic.Severity, diagnostics *[]diagnostic.Diagnostic) {
	required := make([]string, 0, len(allowed))
	for _, key := range allowed {
		if key != "$schema" {
			required = append(required, key)
		}
	}
	validateExactKeysWithOptional(value, required, []string{"$schema"}, path, severity, diagnostics)
}

func validateExactKeysWithOptional(value map[string]any, required []string, optional []string, path string, severity func(string) diagnostic.Severity, diagnostics *[]diagnostic.Diagnostic) {
	allowed := append(append([]string{}, required...), optional...)
	allowedSet := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = true
	}
	for key := range value {
		if !allowedSet[key] {
			*diagnostics = append(*diagnostics, schemaDiagnostic(path, severity, "unknown key "+key))
		}
	}
	for _, key := range required {
		if _, exists := value[key]; !exists {
			*diagnostics = append(*diagnostics, schemaDiagnostic(path, severity, "missing key "+key))
		}
	}
}

func schemaDiagnostic(path string, severity func(string) diagnostic.Severity, message string) diagnostic.Diagnostic {
	return diagnostic.New(
		rules.RuleDefinitionSchemaVersion,
		severity(rules.RuleDefinitionSchemaVersion),
		message,
		path,
		nil,
		diagnostic.EvidenceDefinition,
		"all",
		"ansldes/contract",
		"schema",
	)
}

func object(value any) (map[string]any, bool) {
	result, ok := value.(map[string]any)
	return result, ok
}
