package rules

import (
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/designcontext"
	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/source"
)

func TestDesignSystemRulesMapExactUpstreamIDs(t *testing.T) {
	context := testDesignContext()
	summary := source.Summary{Path: "Example.tsx", PropertyLiterals: []source.PropertyLiteral{
		literal("fontFamily", "string", "Unknown Sans", "font", 1),
		literal("backgroundColor", "string", "#123456", "color", 2),
		literal("borderRadius", "number", "99", "radius", 3),
		literal("fontSize", "number", "15", "size", 4),
	}}
	findings := AnalyzeDesignSystem(summary, context, func(string) diagnostic.Severity { return diagnostic.SeverityError })
	if len(findings) != 4 {
		t.Fatalf("findings = %#v", findings)
	}
	expected := map[string]string{
		RuleDesignSystemFont: "design-system-font", RuleDesignSystemColor: "design-system-color",
		RuleDesignSystemRadius: "design-system-radius", RuleDesignSystemFontSize: "design-system-font-size",
	}
	for _, finding := range findings {
		if len(finding.SourceRuleIDs) != 1 || finding.SourceRuleIDs[0] != expected[finding.RuleID] {
			t.Fatalf("mapping = %#v", finding)
		}
	}
}

func TestDesignSystemAllowsCanonicalValuesAndDetectsRoleDrift(t *testing.T) {
	context := testDesignContext()
	allowed := source.Summary{Path: "Allowed.tsx", PropertyLiterals: []source.PropertyLiteral{
		literal("fontFamily", "string", "ExampleSans-Bold", "type", 1), literal("fontWeight", "number", "700", "type", 1),
		literal("typographyRole", "string", "button.label", "type", 1), literal("fontSize", "number", "16", "type", 1),
		literal("color", "string", "#FFFFFF", "color", 2), literal("borderRadius", "number", "12", "radius", 3),
	}}
	if findings := AnalyzeDesignSystem(allowed, context, func(string) diagnostic.Severity { return diagnostic.SeverityError }); len(findings) != 0 {
		t.Fatalf("allowed findings = %#v", findings)
	}
	drift := allowed
	drift.PropertyLiterals[1].Value = "400"
	if findings := AnalyzeDesignSystem(drift, context, func(string) diagnostic.Severity { return diagnostic.SeverityError }); len(findings) != 1 || findings[0].RuleID != RuleDesignSystemFont {
		t.Fatalf("drift findings = %#v", findings)
	}
}

func testDesignContext() designcontext.Context {
	return designcontext.Context{Typography: designcontext.Typography{FamilyName: "Example Sans", Weights: map[string]designcontext.Weight{"regular": {FontFamily: "ExampleSans-Regular", FontWeight: 400}, "bold": {FontFamily: "ExampleSans-Bold", FontWeight: 700}}, Tokens: map[string]designcontext.TypeToken{"body.medium": {FontSize: 16}}, Roles: map[string]designcontext.TypeRole{"button.label": {VisualToken: "body.medium", Weight: "bold"}}}, Colors: designcontext.ColorContract{Primitive: map[string]string{"neutral.0": "#FFFFFF"}}, Radii: designcontext.RadiusContract{Primitive: map[string]float64{"medium": 12}}}
}

func literal(property, kind, value, group string, line int) source.PropertyLiteral {
	return source.PropertyLiteral{Property: property, Kind: kind, Value: value, Group: group, Range: source.Range{Start: source.Position{Line: line, Column: 1}, End: source.Position{Line: line, Column: 2}}}
}
