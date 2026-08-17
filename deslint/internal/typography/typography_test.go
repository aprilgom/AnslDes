package typography

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
)

func TestNegativeFixtureMapsAllFourteenRulesDeterministically(t *testing.T) {
	contents := read(t, "../../../packages/schema/testdata/typography-negative.json")
	_, first, err := Analyze("negative.json", contents, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, ruleID := range RuleIDs() {
		if !hasRule(first, ruleID) {
			t.Fatalf("missing %s in %#v", ruleID, first)
		}
	}
	hallmark := false
	for _, finding := range first {
		upstream := strings.TrimPrefix(finding.RuleID, "typography/")
		if !slices.Contains(finding.SourceRuleIDs, upstream) {
			t.Fatalf("missing upstream mapping in %#v", finding)
		}
		if finding.RuleID == "typography/overused-font" && slices.Contains(finding.SourceRuleIDs, "hallmark-eight-02") {
			hallmark = true
		}
	}
	if !hallmark {
		t.Fatal("missing Hallmark overused-font provenance")
	}
	var reordered Evidence
	if unmarshalErr := json.Unmarshal(contents, &reordered); unmarshalErr != nil {
		t.Fatal(unmarshalErr)
	}
	slices.Reverse(reordered.Nodes)
	reorderedContents, err := json.Marshal(reordered)
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := Analyze("negative.json", reorderedContents, nil, nil)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("ordering differs: %v\n%#v\n%#v", err, first, second)
	}
}

func TestFontScaleMatrixAndProfileThresholdsRemainScoped(t *testing.T) {
	for _, fixture := range []string{"typography-positive-100.json", "typography-positive-235.json"} {
		contents := read(t, "../../../packages/schema/testdata/"+fixture)
		evidence, findings, err := Analyze(fixture, contents, nil, nil)
		if err != nil || len(findings) != 0 || evidence.FontScale < 1 {
			t.Fatalf("%s = %#v %#v %v", fixture, evidence, findings, err)
		}
	}
	contents := read(t, "../../../packages/schema/testdata/typography-positive-100.json")
	var stricter Evidence
	if err := json.Unmarshal(contents, &stricter); err != nil {
		t.Fatal(err)
	}
	stricter.Policy.MinimumHeadingRatio = 2.5
	changed, _ := json.Marshal(stricter)
	_, findings, err := Analyze("strict.json", changed, nil, nil)
	if err != nil || !hasRule(findings, "typography/flat-type-hierarchy") {
		t.Fatalf("profile threshold was not applied: %#v %v", findings, err)
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
