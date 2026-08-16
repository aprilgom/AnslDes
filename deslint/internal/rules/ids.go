// Package rules contains stable rule IDs and platform-neutral evaluation.
package rules

const (
	RuleDefinitionSchemaVersion  = "definition/schema-version"
	RuleDefinitionInvalidRef     = "definition/invalid-reference"
	RuleDefinitionUnknownToken   = "definition/unknown-token"
	RuleSourceSyntaxError        = "source/syntax-error"
	RuleSourceRawValue           = "source/raw-value"
	RulePencilRawValue           = "pencil/raw-value"
	RuleLayoutProblem            = "layout/problem"
	RuleEvidenceMissing          = "evidence/missing"
	RuleEvidenceStale            = "evidence/stale"
	RulePolicyDefinitionMismatch = "policy/definition-mismatch"
	RulePolicyBudgetExceeded     = "policy/budget-exceeded"
	RulePolicyExceptionExpired   = "policy/exception-expired"
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
