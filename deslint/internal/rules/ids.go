// Package rules contains stable rule IDs and platform-neutral evaluation.
package rules

const (
	// RuleDefinitionSchemaVersion identifies unsupported definition schema versions.
	RuleDefinitionSchemaVersion = "definition/schema-version"
	// RuleDefinitionInvalidRef identifies invalid definition references.
	RuleDefinitionInvalidRef = "definition/invalid-reference"
	// RuleDefinitionUnknownToken identifies definition references to unknown tokens.
	RuleDefinitionUnknownToken = "definition/unknown-token"
	// RuleSourceSyntaxError identifies source syntax errors.
	RuleSourceSyntaxError = "source/syntax-error"
	// RuleSourceRawValue identifies disallowed raw values in source files.
	RuleSourceRawValue = "source/raw-value"
	// RulePencilRawValue identifies disallowed raw values in Pencil documents.
	RulePencilRawValue = "pencil/raw-value"
	// RuleLayoutProblem identifies computed-layout violations.
	RuleLayoutProblem = "layout/problem"
	// RuleEvidenceMissing identifies required evidence that was not provided.
	RuleEvidenceMissing = "evidence/missing"
	// RuleEvidenceStale identifies evidence that no longer matches its source.
	RuleEvidenceStale = "evidence/stale"
	// RulePolicyDefinitionMismatch identifies incompatible policy and definition versions.
	RulePolicyDefinitionMismatch = "policy/definition-mismatch"
	// RulePolicyBudgetExceeded identifies diagnostic counts above a configured budget.
	RulePolicyBudgetExceeded = "policy/budget-exceeded"
	// RulePolicyExceptionExpired identifies expired policy exceptions.
	RulePolicyExceptionExpired = "policy/exception-expired"
)

// ConfigurableRuleIDs is the exact v1 severity registry.
var ConfigurableRuleIDs = []string{
	RuleDefinitionSchemaVersion,
	RuleDefinitionInvalidRef,
	RuleDefinitionUnknownToken,
	RuleSourceSyntaxError,
	RuleSourceRawValue,
	RulePencilRawValue,
	RuleLayoutProblem,
	RuleEvidenceMissing,
	RuleEvidenceStale,
}
