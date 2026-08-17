package designcontext

import (
	"os"
	"testing"
)

func TestGeneratedFixtureMatchesCanonicalDefinition(t *testing.T) {
	// #nosec G304 -- repository-owned fixtures.
	sidecar, err := os.ReadFile("../../../packages/schema/testdata/generated-design-context/.impeccable/design.json")
	if err != nil {
		t.Fatal(err)
	}
	context, err := Parse(sidecar)
	if err != nil {
		t.Fatal(err)
	}
	// #nosec G304 -- repository-owned fixtures.
	definition, err := os.ReadFile("../../../packages/schema/testdata/example-product.json")
	if err != nil {
		t.Fatal(err)
	}
	fingerprint, err := ContractFingerprint(definition)
	if err != nil || fingerprint != context.Source.ContractSHA256 {
		t.Fatalf("fingerprint = %s, context = %s, error = %v", fingerprint, context.Source.ContractSHA256, err)
	}
}
