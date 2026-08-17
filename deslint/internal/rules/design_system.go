package rules

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/designcontext"
	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/source"
)

var designRuleSources = map[string]string{
	RuleDesignSystemFont: "design-system-font", RuleDesignSystemColor: "design-system-color",
	RuleDesignSystemRadius: "design-system-radius", RuleDesignSystemFontSize: "design-system-font-size",
}

// DesignLiteralAllowed reports whether a raw value is an exact generated contract value.
func DesignLiteralAllowed(literal source.PropertyLiteral, context designcontext.Context) bool {
	switch literal.Property {
	case "color", "backgroundColor", "borderColor", "fill":
		return allowedColor(literal.Value, context)
	case "borderRadius":
		return allowedNumber(literal.Value, radiusValues(context))
	case "fontSize":
		return allowedNumber(literal.Value, fontSizes(context))
	default:
		return false
	}
}

// AnalyzeDesignSystem maps generated contract drift to exact upstream rule IDs.
func AnalyzeDesignSystem(summary source.Summary, context designcontext.Context, severity func(string) diagnostic.Severity) []diagnostic.Diagnostic {
	result := make([]diagnostic.Diagnostic, 0)
	groups := make(map[string]map[string]source.PropertyLiteral)
	for _, literal := range summary.PropertyLiterals {
		if groups[literal.Group] == nil {
			groups[literal.Group] = make(map[string]source.PropertyLiteral)
		}
		groups[literal.Group][literal.Property] = literal
		switch literal.Property {
		case "fontFamily":
			if !allowedFont(literal.Value, context) {
				result = append(result, designFinding(RuleDesignSystemFont, fmt.Sprintf("font family %q is outside the generated contract", literal.Value), summary.Path, literal, severity))
			}
		case "fontWeight":
			if !allowedWeight(literal.Value, context) {
				result = append(result, designFinding(RuleDesignSystemFont, fmt.Sprintf("font weight %q is outside the generated contract", literal.Value), summary.Path, literal, severity))
			}
		case "color", "backgroundColor", "borderColor", "fill":
			if isRawLiteral("color", literal.Kind, literal.Value) && !allowedColor(literal.Value, context) {
				result = append(result, designFinding(RuleDesignSystemColor, fmt.Sprintf("color %q is outside the generated contract", literal.Value), summary.Path, literal, severity))
			}
		case "borderRadius":
			if literal.Kind == "number" && !allowedNumber(literal.Value, radiusValues(context)) {
				result = append(result, designFinding(RuleDesignSystemRadius, fmt.Sprintf("radius %q is outside the generated contract", literal.Value), summary.Path, literal, severity))
			}
		case "fontSize":
			if literal.Kind == "number" && !allowedNumber(literal.Value, fontSizes(context)) {
				result = append(result, designFinding(RuleDesignSystemFontSize, fmt.Sprintf("font size %q is outside the generated typography ramp", literal.Value), summary.Path, literal, severity))
			}
		}
	}
	for _, properties := range groups {
		roleLiteral, hasRole := properties["typographyRole"]
		if !hasRole {
			continue
		}
		role, known := context.Typography.Roles[roleLiteral.Value]
		if !known {
			continue
		}
		token := context.Typography.Tokens[role.VisualToken]
		if size, ok := properties["fontSize"]; ok && !numberEqual(size.Value, token.FontSize) {
			result = append(result, designFinding(RuleDesignSystemFontSize, fmt.Sprintf("font size %q does not match typography role %s", size.Value, roleLiteral.Value), summary.Path, size, severity))
		}
		if weight, ok := properties["fontWeight"]; ok {
			expected := context.Typography.Weights[role.Weight]
			if !numberEqual(weight.Value, float64(expected.FontWeight)) && weight.Value != role.Weight {
				result = append(result, designFinding(RuleDesignSystemFont, fmt.Sprintf("font weight %q does not match typography role %s", weight.Value, roleLiteral.Value), summary.Path, weight, severity))
			}
		}
	}
	diagnostic.Sort(result)
	return diagnostic.MergeCanonical(result)
}

func designFinding(ruleID, message, path string, literal source.PropertyLiteral, severity func(string) diagnostic.Severity) diagnostic.Diagnostic {
	sourceRange := &diagnostic.Range{Start: diagnostic.Position{Line: literal.Range.Start.Line, Column: literal.Range.Start.Column}, End: diagnostic.Position{Line: literal.Range.End.Line, Column: literal.Range.End.Column}}
	return diagnostic.NewWithSources(ruleID, []string{designRuleSources[ruleID]}, severity(ruleID), message, path, sourceRange, diagnostic.EvidenceNativeSource, "react-native", "ansldes/design-system", "design-system")
}

func allowedFont(value string, context designcontext.Context) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == strings.ToLower(context.Typography.FamilyName) {
		return true
	}
	for name, weight := range context.Typography.Weights {
		if normalized == strings.ToLower(weight.FontFamily) || normalized == strings.ToLower(name) {
			return true
		}
	}
	return false
}

func allowedWeight(value string, context designcontext.Context) bool {
	for name, weight := range context.Typography.Weights {
		if value == name || numberEqual(value, float64(weight.FontWeight)) {
			return true
		}
	}
	return false
}

func allowedColor(value string, context designcontext.Context) bool {
	for _, colors := range []map[string]string{context.Colors.Primitive, context.Colors.Asset} {
		for _, allowed := range colors {
			if strings.EqualFold(value, allowed) {
				return true
			}
		}
	}
	return false
}

func radiusValues(context designcontext.Context) []float64 {
	result := make([]float64, 0, len(context.Radii.Primitive))
	for _, value := range context.Radii.Primitive {
		result = append(result, value)
	}
	return result
}
func fontSizes(context designcontext.Context) []float64 {
	result := make([]float64, 0, len(context.Typography.Tokens))
	for _, value := range context.Typography.Tokens {
		result = append(result, value.FontSize)
	}
	return result
}
func allowedNumber(value string, allowed []float64) bool {
	for _, candidate := range allowed {
		if numberEqual(value, candidate) {
			return true
		}
	}
	return false
}
func numberEqual(value string, expected float64) bool {
	parsed, err := strconv.ParseFloat(value, 64)
	return err == nil && parsed == expected
}
