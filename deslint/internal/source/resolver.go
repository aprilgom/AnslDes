package source

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
)

func canonicalSummary(summary Summary) ([]byte, error) { return json.Marshal(summary) }

// ModuleResolver deterministically resolves relative, package, and tsconfig.paths imports against an exact inventory.
type ModuleResolver struct {
	BaseURL        string
	Paths          map[string][]string
	AvailableFiles []string
	Extensions     []string
}

// Resolve returns one exact normalized module path or an error for missing/ambiguous matches.
func (r ModuleResolver) Resolve(importer, specifier string) (string, error) {
	if specifier == "" || strings.Contains(specifier, "\\") {
		return "", fmt.Errorf("invalid module specifier %q", specifier)
	}
	available := make(map[string]bool, len(r.AvailableFiles))
	for _, file := range r.AvailableFiles {
		available[path.Clean(file)] = true
	}
	candidates := []string{}
	if strings.HasPrefix(specifier, ".") {
		candidates = append(candidates, path.Join(path.Dir(importer), specifier))
	} else {
		for pattern, targets := range r.Paths {
			wildcard, matched := matchAlias(pattern, specifier)
			if !matched {
				continue
			}
			for _, target := range targets {
				candidates = append(candidates, path.Join(r.BaseURL, strings.Replace(target, "*", wildcard, 1)))
			}
		}
		if len(candidates) == 0 {
			return "package:" + specifier, nil
		}
	}
	extensions := append([]string(nil), r.Extensions...)
	if len(extensions) == 0 {
		extensions = []string{".ts", ".tsx", ".mts", ".js", ".jsx"}
	}
	matches := []string{}
	for _, candidate := range candidates {
		candidate = path.Clean(candidate)
		if available[candidate] {
			matches = append(matches, candidate)
		}
		for _, extension := range extensions {
			if available[candidate+extension] {
				matches = append(matches, candidate+extension)
			}
			if available[path.Join(candidate, "index"+extension)] {
				matches = append(matches, path.Join(candidate, "index"+extension))
			}
		}
	}
	sort.Strings(matches)
	matches = compact(matches)
	if len(matches) != 1 {
		return "", fmt.Errorf("module %q from %q resolved to %d files", specifier, importer, len(matches))
	}
	return matches[0], nil
}

func matchAlias(pattern, specifier string) (string, bool) {
	if !strings.Contains(pattern, "*") {
		return "", pattern == specifier
	}
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 || !strings.HasPrefix(specifier, parts[0]) || !strings.HasSuffix(specifier, parts[1]) {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(specifier, parts[0]), parts[1]), true
}

func compact(values []string) []string {
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

// SemanticProcess is the process transport boundary for optional type-aware analysis.
type SemanticProcess interface {
	Exchange(request []byte) (response []byte, err error)
}

// ProcessSemanticProvider delegates type-aware work without exposing engine internals to syntax rules.
type ProcessSemanticProvider struct{ Process SemanticProcess }

// Analyze sends canonical JSON IR to the configured process boundary.
func (p ProcessSemanticProvider) Analyze(summary Summary) ([]byte, error) {
	if p.Process == nil {
		return nil, fmt.Errorf("semantic process is required")
	}
	request, err := canonicalSummary(summary)
	if err != nil {
		return nil, err
	}
	return p.Process.Exchange(request)
}
