// Package treesitter implements source analysis with the official Tree-sitter
// Go bindings and TypeScript/TSX grammar.
package treesitter

import (
	"fmt"
	"sort"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"

	"github.com/aprilgom/AnslDes/deslint/internal/source"
)

// Analyzer parses TypeScript and TSX into a deterministic syntax summary.
type Analyzer struct{}

// NewAnalyzer returns a Tree-sitter source analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Analyze parses one source file and reports syntax structure. Syntax errors
// remain explicit in Summary.HasError and are never treated as a clean result.
func (a *Analyzer) Analyze(path string, contents []byte, language source.Language) (source.Summary, error) {
	parser := sitter.NewParser()
	defer parser.Close()

	grammar, err := grammarFor(language)
	if err != nil {
		return source.Summary{}, err
	}
	if err := parser.SetLanguage(grammar); err != nil {
		return source.Summary{}, fmt.Errorf("set %s grammar: %w", language, err)
	}

	tree := parser.Parse(contents, nil)
	defer tree.Close()
	root := tree.RootNode()
	counts := make(map[string]int)
	namedNodes := collectNamedNodeKinds(root, counts)
	propertyLiterals := make([]source.PropertyLiteral, 0)
	collectPropertyLiterals(root, contents, &propertyLiterals)
	sort.SliceStable(propertyLiterals, func(i, j int) bool {
		left, right := propertyLiterals[i], propertyLiterals[j]
		if left.Range.Start.Line != right.Range.Start.Line {
			return left.Range.Start.Line < right.Range.Start.Line
		}
		if left.Range.Start.Column != right.Range.Start.Column {
			return left.Range.Start.Column < right.Range.Start.Column
		}
		return left.Property < right.Property
	})

	return source.Summary{
		Path:             path,
		Language:         language,
		RootKind:         root.Kind(),
		HasError:         root.HasError(),
		NamedNodes:       namedNodes,
		NodeKindUses:     counts,
		PropertyLiterals: propertyLiterals,
	}, nil
}

func collectPropertyLiterals(node *sitter.Node, contents []byte, result *[]source.PropertyLiteral) {
	if literal, ok := propertyLiteral(node, contents); ok {
		*result = append(*result, literal)
	}
	for index := uint(0); index < node.ChildCount(); index++ {
		child := node.Child(index)
		if child != nil {
			collectPropertyLiterals(child, contents, result)
		}
	}
}

func propertyLiteral(node *sitter.Node, contents []byte) (source.PropertyLiteral, bool) {
	var propertyNode, valueNode *sitter.Node
	switch node.Kind() {
	case "pair":
		propertyNode = node.ChildByFieldName("key")
		valueNode = node.ChildByFieldName("value")
	case "jsx_attribute":
		propertyNode = node.ChildByFieldName("name")
		valueNode = node.ChildByFieldName("value")
		if propertyNode == nil {
			propertyNode = node.NamedChild(0)
		}
		if valueNode == nil {
			valueNode = node.NamedChild(1)
		}
	default:
		return source.PropertyLiteral{}, false
	}
	if propertyNode == nil || valueNode == nil {
		return source.PropertyLiteral{}, false
	}
	if valueNode.Kind() == "jsx_expression" {
		valueNode = valueNode.NamedChild(0)
		if valueNode == nil {
			return source.PropertyLiteral{}, false
		}
	}
	kind, value, ok := normalizedLiteral(valueNode, contents)
	if !ok {
		return source.PropertyLiteral{}, false
	}
	return source.PropertyLiteral{
		Property: trimQuotes(propertyNode.Utf8Text(contents)),
		Kind:     kind,
		Value:    value,
		Range: source.Range{
			Start: source.Position{Line: int(valueNode.StartPosition().Row) + 1, Column: int(valueNode.StartPosition().Column) + 1},
			End:   source.Position{Line: int(valueNode.EndPosition().Row) + 1, Column: int(valueNode.EndPosition().Column) + 1},
		},
	}, true
}

func normalizedLiteral(node *sitter.Node, contents []byte) (string, string, bool) {
	text := node.Utf8Text(contents)
	switch node.Kind() {
	case "string", "template_string":
		return "string", trimQuotes(text), true
	case "number":
		return "number", text, true
	case "unary_expression":
		if strings.HasPrefix(text, "-") {
			return "number", text, true
		}
	}
	return "", "", false
}

func trimQuotes(value string) string {
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '\'' && last == '\'') || (first == '"' && last == '"') || (first == '`' && last == '`') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func grammarFor(language source.Language) (*sitter.Language, error) {
	switch language {
	case source.LanguageTypeScript:
		return sitter.NewLanguage(typescript.LanguageTypescript()), nil
	case source.LanguageTSX:
		return sitter.NewLanguage(typescript.LanguageTSX()), nil
	default:
		return nil, fmt.Errorf("unsupported source language %q", language)
	}
}

func collectNamedNodeKinds(node *sitter.Node, counts map[string]int) int {
	total := 0
	if node.IsNamed() {
		counts[node.Kind()]++
		total++
	}
	for index := uint(0); index < node.ChildCount(); index++ {
		child := node.Child(index)
		if child != nil {
			total += collectNamedNodeKinds(child, counts)
		}
	}
	return total
}
