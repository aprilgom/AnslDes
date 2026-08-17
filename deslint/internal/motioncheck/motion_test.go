package motioncheck

import (
	"os"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
)

func TestSourceFixtureMapsAllSixRules(t *testing.T) {
	evidence, findings, err := Analyze("source.json", read(t, "../../../packages/schema/testdata/motion-negative-source.json"), Config{ProfileID: "operate"})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.EvidenceKind != diagnostic.EvidenceNativeSource {
		t.Fatalf("evidence = %#v", evidence)
	}
	for _, ruleID := range RuleIDs() {
		if !hasRule(findings, ruleID) {
			t.Fatalf("missing %s in %#v", ruleID, findings)
		}
	}
}

func TestReducedMotionUsesExactRegistryAndKeepsStateUnderstandable(t *testing.T) {
	evidence, findings, err := Analyze("reduced.json", read(t, "../../../packages/schema/testdata/motion-reduced-simulator.json"), Config{ProfileID: "operate", Registry: registry()})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Preference != "reduce" || len(findings) != 0 {
		t.Fatalf("evidence/findings = %#v %#v", evidence, findings)
	}
}

func TestDesignDocumentUsesResolvedNormalTransition(t *testing.T) {
	evidence, findings, err := Analyze("document.json", read(t, "../../../packages/schema/testdata/motion-design-document.json"), Config{ProfileID: "operate", Registry: registry()})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Platform != "design-document" || len(findings) != 0 {
		t.Fatalf("evidence/findings = %#v %#v", evidence, findings)
	}
}

func TestRegistryOwnerAndPreferenceSequencingAreStrict(t *testing.T) {
	contents := read(t, "../../../packages/schema/testdata/motion-reduced-simulator.json")
	wrongOwner := []byte(strings.Replace(string(contents), `"owner": "example-control"`, `"owner": "other-owner"`, 1))
	if _, _, err := Analyze("owner.json", wrongOwner, Config{ProfileID: "operate", Registry: registry()}); err == nil || !strings.Contains(err.Error(), "owner or purpose") {
		t.Fatalf("owner error = %v", err)
	}
	replayed := []byte(strings.Replace(string(contents), `"effectsReplayedAfterResolution": false`, `"effectsReplayedAfterResolution": true`, 1))
	if _, _, err := Analyze("replayed.json", replayed, Config{ProfileID: "operate", Registry: registry()}); err == nil || !strings.Contains(err.Error(), "must not replay") {
		t.Fatalf("replay error = %v", err)
	}
}

func registry() map[string]Transition {
	return map[string]Transition{
		"control.press": {
			Owner:             "example-control",
			Purpose:           "state-change",
			DurationMS:        160,
			ReducedDurationMS: 0,
			Easing:            []float64{0.2, 0, 0, 1},
			ReducedFallback:   "instant",
		},
	}
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	// #nosec G304 -- callers provide repository-owned fixture paths.
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func hasRule(findings []diagnostic.Diagnostic, ruleID string) bool {
	for _, finding := range findings {
		if finding.RuleID == ruleID {
			return true
		}
	}
	return false
}
