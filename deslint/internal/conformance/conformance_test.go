package conformance

import (
	"os"
	"reflect"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
)

func TestAnalyzeOperateFixtureDeterministically(t *testing.T) {
	contents := readFixture(t, "../../../packages/schema/testdata/operate-conformance.json")
	config := Config{ProfileID: "operate", MaxOversizedActions: 1, MaxInconsistentActions: 0}
	first, err := Analyze("operate.json", contents, config)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Analyze("operate.json", contents, config)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Diagnostics) != 0 || !reflect.DeepEqual(first, second) {
		t.Fatalf("unexpected deterministic result: %#v %#v", first, second)
	}
}

func TestAnalyzeFindsSystemicConformanceProblems(t *testing.T) {
	contents := readFixture(t, "../../../packages/schema/testdata/operate-conformance-invalid.json")
	result, err := Analyze("invalid.json", contents, Config{
		ProfileID: "operate", MaxOversizedActions: 0, MaxInconsistentActions: 0,
		Severity: func(string) diagnostic.Severity { return diagnostic.SeverityError },
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(result.Diagnostics), 8; got != want {
		for _, finding := range result.Diagnostics {
			t.Log(finding.RuleID, finding.Message)
		}
		t.Fatalf("diagnostic count = %d, want %d", got, want)
	}
	for _, finding := range result.Diagnostics {
		if finding.EvidenceKind != diagnostic.EvidenceConsumerConformance {
			t.Fatalf("evidence kind = %q", finding.EvidenceKind)
		}
	}
}

func TestProfileThresholdsOnlyAdjustCounts(t *testing.T) {
	contents := readFixture(t, "../../../packages/schema/testdata/operate-conformance-invalid.json")
	strict, err := Analyze("invalid.json", contents, Config{ProfileID: "operate"})
	if err != nil {
		t.Fatal(err)
	}
	permissive, err := Analyze("invalid.json", contents, Config{
		ProfileID: "operate", MaxOversizedActions: 10, MaxInconsistentActions: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(strict.Diagnostics) != 8 || len(permissive.Diagnostics) != 5 {
		t.Fatalf("threshold verdict counts = strict:%d permissive:%d", len(strict.Diagnostics), len(permissive.Diagnostics))
	}
}

func TestAnalyzeRejectsUnknownAndDuplicateFields(t *testing.T) {
	for _, contents := range [][]byte{
		[]byte(`{"schemaVersion":1,"profileId":"operate","platform":"web","surfaceId":"x","controls":[],"consumerPath":"x"}`),
		[]byte(`{"schemaVersion":1,"profileId":"operate","profileId":"read","platform":"web","surfaceId":"x","controls":[]}`),
	} {
		if _, err := Analyze("bad.json", contents, Config{}); err == nil {
			t.Fatal("expected strict parse error")
		}
	}
}

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	// #nosec G304 -- callers provide repository-owned fixture paths.
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}
