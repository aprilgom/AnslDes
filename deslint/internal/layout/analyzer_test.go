package layout

import (
	"os"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
)

func TestVisitorReportPreservesOptionsFingerprintNodeCountAndIssues(t *testing.T) {
	contents, err := os.ReadFile("../../testdata/negative/layout.json")
	if err != nil {
		t.Fatal(err)
	}
	findings, err := Analyze("layout.json", contents, strings.Repeat("0", 64), func(string) diagnostic.Severity { return diagnostic.SeverityError })
	if err != nil || len(findings) != 2 {
		t.Fatalf("visitor report = %#v %v", findings, err)
	}
	invalid := []byte(strings.Replace(string(contents), `"computeBounds": true`, `"computeBounds": false`, 1))
	if _, err := Analyze("layout.json", invalid, strings.Repeat("0", 64), func(string) diagnostic.Severity { return diagnostic.SeverityError }); err == nil {
		t.Fatal("Analyze(visitor without computed bounds) error = nil")
	}
}
