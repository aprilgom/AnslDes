// Package source defines the language-neutral source-analysis boundary.
package source

// Language identifies the source grammar used to build the normalized IR.
type Language string

const (
	// LanguageTypeScript selects the TypeScript grammar.
	LanguageTypeScript Language = "typescript"
	// LanguageTSX selects the TSX grammar.
	LanguageTSX Language = "tsx"
)

// Summary is the parser-neutral source IR consumed by rules and semantic providers.
type Summary struct {
	Path             string            `json:"path"`
	Language         Language          `json:"language"`
	RootKind         string            `json:"rootKind"`
	HasError         bool              `json:"hasError"`
	NamedNodes       int               `json:"namedNodes"`
	NodeKindUses     map[string]int    `json:"nodeKindUses"`
	PropertyLiterals []PropertyLiteral `json:"propertyLiterals"`
	SyntaxNodes      []SyntaxNode      `json:"syntaxNodes"`
	Bindings         []Binding         `json:"bindings"`
	Scopes           []Scope           `json:"scopes"`
}

// SyntaxNode normalizes import, export, declaration, JSX, attribute, expression, and spread constructs.
type SyntaxNode struct {
	Kind  string `json:"kind"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Range Range  `json:"range"`
}

// Binding is a lexical declaration with alias, duplicate, and shadow resolution.
type Binding struct {
	Name      string `json:"name"`
	AliasOf   string `json:"aliasOf,omitempty"`
	Kind      string `json:"kind"`
	ScopeID   string `json:"scopeId"`
	Duplicate bool   `json:"duplicate"`
	Shadows   bool   `json:"shadows"`
	Range     Range  `json:"range"`
}

// Scope is one stable lexical scope and its parent relationship.
type Scope struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	Kind     string `json:"kind"`
	Range    Range  `json:"range"`
}

// SemanticProvider keeps type-aware analysis behind an explicit process boundary.
type SemanticProvider interface {
	Analyze(summary Summary) ([]byte, error)
}

// Position is a one-based source coordinate.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Range is a literal source span.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// PropertyLiteral is a normalized object or JSX property with a literal value.
type PropertyLiteral struct {
	Property string `json:"property"`
	Kind     string `json:"kind"`
	Value    string `json:"value"`
	Group    string `json:"group"`
	Range    Range  `json:"range"`
}

// Analyzer parses a source file without coupling rules to a concrete parser.
type Analyzer interface {
	Analyze(path string, contents []byte, language Language) (Summary, error)
}
