// Package pen analyzes serialized Pencil document evidence independently of layout output.
package pen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Analyze reports configured raw properties in a Pencil JSON document.
func Analyze(path string, contents []byte, propertyKinds map[string]string, severity func(string) diagnostic.Severity) ([]diagnostic.Diagnostic, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	var root any
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("decode Pencil document: %w", err)
	}
	diagnostics := make([]diagnostic.Diagnostic, 0)
	walk(root, "$", path, propertyKinds, severity, &diagnostics)
	diagnostic.Sort(diagnostics)
	return diagnostics, nil
}

func walk(value any, jsonPath, path string, propertyKinds map[string]string, severity func(string) diagnostic.Severity, diagnostics *[]diagnostic.Diagnostic) {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child := typed[key]
			if category, configured := propertyKinds[key]; configured {
				kind, literal, ok := literalValue(child)
				if ok && rules.IsRawLiteral(category, kind, literal) {
					*diagnostics = append(*diagnostics, diagnostic.New(
						rules.RulePencilRawValue,
						severity(rules.RulePencilRawValue),
						fmt.Sprintf("raw %s value %q at %s.%s", category, literal, jsonPath, key),
						path,
						nil,
						diagnostic.EvidencePencil,
						"pencil",
						"ansldes/pencil",
						"raw",
					))
				}
			}
			walk(child, jsonPath+"."+key, path, propertyKinds, severity, diagnostics)
		}
	case []any:
		for index, child := range typed {
			walk(child, fmt.Sprintf("%s[%d]", jsonPath, index), path, propertyKinds, severity, diagnostics)
		}
	}
}

func literalValue(value any) (string, string, bool) {
	switch typed := value.(type) {
	case string:
		return "string", typed, true
	case json.Number:
		return "number", typed.String(), true
	default:
		return "", "", false
	}
}
