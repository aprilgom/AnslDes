package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/designcontext"
	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/source"
)

var rawColorPattern = regexp.MustCompile(`(?i)^(?:#[0-9a-f]{3,8}|rgba?\(|hsla?\(|transparent$)`)

// AnalyzeSource evaluates normalized syntax evidence against injected product policy.
func AnalyzeSource(summary source.Summary, propertyKinds map[string]string, severity func(string) diagnostic.Severity) []diagnostic.Diagnostic {
	return AnalyzeSourceWithDesignContext(summary, propertyKinds, severity, nil)
}

// AnalyzeSourceWithDesignContext applies generated contract permissions before raw-value rules.
func AnalyzeSourceWithDesignContext(summary source.Summary, propertyKinds map[string]string, severity func(string) diagnostic.Severity, context *designcontext.Context) []diagnostic.Diagnostic {
	diagnostics := make([]diagnostic.Diagnostic, 0)
	if summary.HasError {
		diagnostics = append(diagnostics, diagnostic.New(
			RuleSourceSyntaxError,
			severity(RuleSourceSyntaxError),
			"source contains a syntax error",
			summary.Path,
			nil,
			diagnostic.EvidenceSource,
			"react-native",
			"ansldes/source",
			"syntax",
		))
	}
	for _, literal := range summary.PropertyLiterals {
		category, configured := propertyKinds[literal.Property]
		if !configured || !isRawLiteral(category, literal.Kind, literal.Value) || (context != nil && DesignLiteralAllowed(literal, *context)) {
			continue
		}
		sourceRange := &diagnostic.Range{
			Start: diagnostic.Position{Line: literal.Range.Start.Line, Column: literal.Range.Start.Column},
			End:   diagnostic.Position{Line: literal.Range.End.Line, Column: literal.Range.End.Column},
		}
		diagnostics = append(diagnostics, diagnostic.New(
			RuleSourceRawValue,
			severity(RuleSourceRawValue),
			fmt.Sprintf("raw %s value %q for property %s", category, literal.Value, literal.Property),
			summary.Path,
			sourceRange,
			diagnostic.EvidenceSource,
			"react-native",
			"ansldes/source",
			"raw",
		))
	}
	if context != nil {
		diagnostics = append(diagnostics, AnalyzeDesignSystem(summary, *context, severity)...)
	}
	diagnostic.Sort(diagnostics)
	return diagnostics
}

// IsRawLiteral is shared with Pencil JSON property analysis.
func IsRawLiteral(category, kind, value string) bool {
	return isRawLiteral(category, kind, value)
}

func isRawLiteral(category, kind, value string) bool {
	if strings.HasPrefix(value, "{") {
		return false
	}
	switch category {
	case "color":
		return kind == "string" && rawColorPattern.MatchString(value)
	case "number", "motion":
		return kind == "number"
	default:
		return false
	}
}
