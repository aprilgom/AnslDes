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
	syntaxNodes, bindings, scopes := collectNormalizedIR(root, contents)

	return source.Summary{
		Path:             path,
		Language:         language,
		RootKind:         root.Kind(),
		HasError:         root.HasError(),
		NamedNodes:       namedNodes,
		NodeKindUses:     counts,
		PropertyLiterals: propertyLiterals,
		SyntaxNodes:      syntaxNodes,
		Bindings:         bindings,
		Scopes:           scopes,
	}, nil
}

func collectNormalizedIR(root *sitter.Node, contents []byte) ([]source.SyntaxNode, []source.Binding, []source.Scope) {
	nodes := make([]source.SyntaxNode, 0)
	bindings := make([]source.Binding, 0)
	scopes := []source.Scope{{ID: scopeID(root), Kind: "program", Range: nodeRange(root)}}
	declared := map[string]map[string]bool{scopes[0].ID: {}}
	parents := map[string]string{}

	var walk func(*sitter.Node, string)
	walk = func(node *sitter.Node, currentScope string) {
		if kind, ok := normalizedNodeKind(node.Kind()); ok {
			nodes = append(nodes, source.SyntaxNode{Kind: kind, Name: nodeName(node, contents), Value: normalizedNodeValue(node, contents), Range: nodeRange(node)})
		}
		if name, alias, kind, ok := bindingFor(node, contents); ok {
			duplicate := declared[currentScope][name]
			shadows := false
			for parent := parents[currentScope]; parent != ""; parent = parents[parent] {
				if declared[parent][name] {
					shadows = true
					break
				}
			}
			bindings = append(bindings, source.Binding{Name: name, AliasOf: alias, Kind: kind, ScopeID: currentScope, Duplicate: duplicate, Shadows: shadows, Range: nodeRange(node)})
			declared[currentScope][name] = true
		}
		childScope := currentScope
		if node != root && isScope(node.Kind()) {
			childScope = scopeID(node)
			parents[childScope] = currentScope
			declared[childScope] = map[string]bool{}
			scopes = append(scopes, source.Scope{ID: childScope, ParentID: currentScope, Kind: node.Kind(), Range: nodeRange(node)})
		}
		for index := uint(0); index < node.ChildCount(); index++ {
			if child := node.Child(index); child != nil {
				walk(child, childScope)
			}
		}
	}
	walk(root, scopes[0].ID)
	sort.SliceStable(nodes, func(i, j int) bool { return irLess(nodes[i].Range, nodes[j].Range, nodes[i].Kind, nodes[j].Kind) })
	sort.SliceStable(bindings, func(i, j int) bool {
		return irLess(bindings[i].Range, bindings[j].Range, bindings[i].Name, bindings[j].Name)
	})
	return nodes, bindings, scopes
}

func normalizedNodeKind(kind string) (string, bool) {
	switch kind {
	case "import_statement":
		return "import", true
	case "export_statement":
		return "export", true
	case "lexical_declaration", "variable_declaration", "function_declaration", "class_declaration", "interface_declaration", "type_alias_declaration":
		return "declaration", true
	case "jsx_element", "jsx_self_closing_element":
		return "jsx-element", true
	case "jsx_attribute":
		return "attribute", true
	case "call_expression", "member_expression", "binary_expression", "ternary_expression", "arrow_function":
		return "expression", true
	case "spread_element":
		return "spread", true
	default:
		return "", false
	}
}

func isScope(kind string) bool {
	switch kind {
	case "function_declaration", "function_expression", "arrow_function", "method_definition", "statement_block":
		return true
	default:
		return false
	}
}

func bindingFor(node *sitter.Node, contents []byte) (string, string, string, bool) {
	var nameNode, sourceNode *sitter.Node
	switch node.Kind() {
	case "variable_declarator", "function_declaration", "class_declaration", "interface_declaration", "type_alias_declaration":
		nameNode = node.ChildByFieldName("name")
	case "required_parameter", "optional_parameter":
		nameNode = node.ChildByFieldName("pattern")
		if nameNode == nil {
			nameNode = node.NamedChild(0)
		}
	case "import_specifier":
		sourceNode = node.ChildByFieldName("name")
		nameNode = node.ChildByFieldName("alias")
		if nameNode == nil {
			nameNode = sourceNode
		}
	default:
		return "", "", "", false
	}
	if nameNode == nil || nameNode.Kind() != "identifier" {
		return "", "", "", false
	}
	name := nameNode.Utf8Text(contents)
	alias := ""
	if sourceNode != nil && sourceNode.Utf8Text(contents) != name {
		alias = sourceNode.Utf8Text(contents)
	}
	return name, alias, node.Kind(), true
}

func nodeName(node *sitter.Node, contents []byte) string {
	for _, field := range []string{"name", "function", "property"} {
		if child := node.ChildByFieldName(field); child != nil {
			return child.Utf8Text(contents)
		}
	}
	return ""
}

func normalizedNodeValue(node *sitter.Node, contents []byte) string {
	if node.Kind() == "import_statement" {
		if child := node.ChildByFieldName("source"); child != nil {
			return trimQuotes(child.Utf8Text(contents))
		}
	}
	return ""
}

func scopeID(node *sitter.Node) string {
	return fmt.Sprintf("%s:%d:%d", node.Kind(), node.StartByte(), node.EndByte())
}

func nodeRange(node *sitter.Node) source.Range {
	return source.Range{
		Start: source.Position{Line: int(node.StartPosition().Row) + 1, Column: int(node.StartPosition().Column) + 1},
		End:   source.Position{Line: int(node.EndPosition().Row) + 1, Column: int(node.EndPosition().Column) + 1},
	}
}

func irLess(left, right source.Range, leftName, rightName string) bool {
	if left.Start.Line != right.Start.Line {
		return left.Start.Line < right.Start.Line
	}
	if left.Start.Column != right.Start.Column {
		return left.Start.Column < right.Start.Column
	}
	return leftName < rightName
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
		Group:    nodeGroup(node),
		Range: source.Range{
			Start: source.Position{Line: int(valueNode.StartPosition().Row) + 1, Column: int(valueNode.StartPosition().Column) + 1},
			End:   source.Position{Line: int(valueNode.EndPosition().Row) + 1, Column: int(valueNode.EndPosition().Column) + 1},
		},
	}, true
}

func nodeGroup(node *sitter.Node) string {
	parent := node.Parent()
	if parent == nil {
		return ""
	}
	return fmt.Sprintf("%d:%d", parent.StartByte(), parent.EndByte())
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
