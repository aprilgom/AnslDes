package rules

import "testing"

func TestPackEvolutionRequiresGovernedSemverAndExactDiff(t *testing.T) {
	previous := RulePackSpec{ID: "synthetic", Version: "1.0.0", Rules: []RuleSpec{{ID: "synthetic/one"}}}
	additive := RulePackSpec{ID: "synthetic", Version: "1.1.0", Rules: []RuleSpec{{ID: "synthetic/one"}, {ID: "synthetic/two"}}}
	addition := []RuleEvolution{{RuleID: "synthetic/two", Kind: "added", Owner: "example-owner", MigrationPlan: "enable after consumer evidence migration", DefaultActivation: "active"}}
	if err := ValidatePackEvolution(previous, additive, addition); err != nil {
		t.Fatalf("additive evolution = %v", err)
	}
	patch := additive
	patch.Version = "1.0.1"
	if err := ValidatePackEvolution(previous, patch, addition); err == nil {
		t.Fatal("patch addition error = nil")
	}
	removed := RulePackSpec{ID: "synthetic", Version: "2.0.0", Rules: []RuleSpec{{ID: "synthetic/replacement"}}}
	replacement := []RuleEvolution{
		{RuleID: "synthetic/one", Kind: "replaced", Owner: "example-owner", MigrationPlan: "migrate to the exact replacement rule", ReplacementRuleID: "synthetic/replacement"},
		{RuleID: "synthetic/replacement", Kind: "added", Owner: "example-owner", MigrationPlan: "activate with the replacement migration", DefaultActivation: "active"},
	}
	if err := ValidatePackEvolution(previous, removed, replacement); err != nil {
		t.Fatalf("replacement evolution = %v", err)
	}
	removed.Version = "1.1.0"
	if err := ValidatePackEvolution(previous, removed, replacement); err == nil {
		t.Fatal("minor removal error = nil")
	}
}
