package treesitter_test

import (
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/source"
	"github.com/aprilgom/AnslDes/deslint/internal/source/treesitter"
)

func TestAnalyzeTSX(t *testing.T) {
	t.Parallel()

	contents := []byte(`
import { Button as CanonicalButton } from "@components/Button";

export function Example({ disabled }: { disabled: boolean }) {
  return <CanonicalButton disabled={disabled} onPress={() => undefined} label="계속" />;
}
`)

	summary, err := treesitter.NewAnalyzer().Analyze("Example.tsx", contents, source.LanguageTSX)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if summary.HasError {
		t.Fatal("Analyze() reported an unexpected syntax error")
	}
	if summary.RootKind != "program" {
		t.Fatalf("RootKind = %q, want program", summary.RootKind)
	}
	if summary.NodeKindUses["import_statement"] != 1 {
		t.Fatalf("import_statement count = %d, want 1", summary.NodeKindUses["import_statement"])
	}
	if summary.NodeKindUses["jsx_self_closing_element"] != 1 {
		t.Fatalf("jsx_self_closing_element count = %d, want 1", summary.NodeKindUses["jsx_self_closing_element"])
	}
	kinds := map[string]bool{}
	for _, node := range summary.SyntaxNodes {
		kinds[node.Kind] = true
	}
	for _, kind := range []string{"import", "export", "declaration", "jsx-element", "attribute", "expression"} {
		if !kinds[kind] {
			t.Fatalf("normalized IR is missing %q: %#v", kind, summary.SyntaxNodes)
		}
	}
	if len(summary.Bindings) == 0 || summary.Bindings[0].AliasOf != "Button" {
		t.Fatalf("bindings = %#v", summary.Bindings)
	}
}

func TestAnalyzeResolvesDuplicateAndShadowedBindings(t *testing.T) {
	t.Parallel()
	summary, err := treesitter.NewAnalyzer().Analyze("Scope.ts", []byte(`
const value = 1;
function outer(value: number) {
  const local = value;
  { const local = 2; const local = 3; }
}
`), source.LanguageTypeScript)
	if err != nil {
		t.Fatal(err)
	}
	seenShadow, seenDuplicate := false, false
	for _, binding := range summary.Bindings {
		seenShadow = seenShadow || binding.Shadows
		seenDuplicate = seenDuplicate || binding.Duplicate
	}
	if !seenShadow || !seenDuplicate {
		t.Fatalf("bindings = %#v", summary.Bindings)
	}
}

func TestAnalyzeKeepsSyntaxErrorsVisible(t *testing.T) {
	t.Parallel()

	summary, err := treesitter.NewAnalyzer().Analyze(
		"Broken.tsx",
		[]byte(`export function Broken( { return <View /> }`),
		source.LanguageTSX,
	)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if !summary.HasError {
		t.Fatal("Analyze() HasError = false, want true")
	}
}

func TestAnalyzeNormalizesObjectAndJSXPropertyLiterals(t *testing.T) {
	t.Parallel()

	contents := []byte(`
const style = { backgroundColor: "#ffffff", gap: 12, dynamic: tokens.gap };
export const Example = () => <Box borderRadius={8} color={tokens.text} />;
`)
	summary, err := treesitter.NewAnalyzer().Analyze("Raw.tsx", contents, source.LanguageTSX)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(summary.PropertyLiterals) != 3 {
		t.Fatalf("PropertyLiterals count = %d, want 3: %#v", len(summary.PropertyLiterals), summary.PropertyLiterals)
	}
	if summary.PropertyLiterals[0].Property != "backgroundColor" || summary.PropertyLiterals[0].Value != "#ffffff" {
		t.Fatalf("first literal = %#v", summary.PropertyLiterals[0])
	}
	if summary.PropertyLiterals[2].Property != "borderRadius" || summary.PropertyLiterals[2].Range.Start.Line != 3 {
		t.Fatalf("JSX literal = %#v", summary.PropertyLiterals[2])
	}
}
