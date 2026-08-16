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

// Summary is the bootstrap parser output. It will be replaced by normalized
// import, binding, JSX and expression IR during migration.
type Summary struct {
	Path             string            `json:"path"`
	Language         Language          `json:"language"`
	RootKind         string            `json:"rootKind"`
	HasError         bool              `json:"hasError"`
	NamedNodes       int               `json:"namedNodes"`
	NodeKindUses     map[string]int    `json:"nodeKindUses"`
	PropertyLiterals []PropertyLiteral `json:"propertyLiterals"`
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
	Range    Range  `json:"range"`
}

// Analyzer parses a source file without coupling rules to a concrete parser.
type Analyzer interface {
	Analyze(path string, contents []byte, language Language) (Summary, error)
}
