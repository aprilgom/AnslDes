package source_test

import (
	"bytes"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/source"
)

func TestModuleResolverHandlesRelativePackageAndPathAliases(t *testing.T) {
	t.Parallel()
	resolver := source.ModuleResolver{
		BaseURL: ".", Paths: map[string][]string{"@components/*": {"src/components/*"}},
		AvailableFiles: []string{"src/components/Button.tsx", "src/theme.ts"},
	}
	for _, item := range []struct{ importer, specifier, want string }{
		{"src/App.tsx", "./theme", "src/theme.ts"},
		{"src/App.tsx", "@components/Button", "src/components/Button.tsx"},
		{"src/App.tsx", "react", "package:react"},
	} {
		got, err := resolver.Resolve(item.importer, item.specifier)
		if err != nil || got != item.want {
			t.Fatalf("Resolve(%q) = %q, %v; want %q", item.specifier, got, err, item.want)
		}
	}
}

func TestProcessSemanticProviderUsesCanonicalBoundary(t *testing.T) {
	t.Parallel()
	process := &semanticProcess{}
	provider := source.ProcessSemanticProvider{Process: process}
	first, err := provider.Analyze(source.Summary{Path: "src/A.tsx", NodeKindUses: map[string]int{"jsx": 1, "import": 2}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := provider.Analyze(source.Summary{Path: "src/A.tsx", NodeKindUses: map[string]int{"import": 2, "jsx": 1}})
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("semantic requests differ: %s %s, %v", first, second, err)
	}
}

type semanticProcess struct{}

func (p *semanticProcess) Exchange(request []byte) ([]byte, error) { return request, nil }
