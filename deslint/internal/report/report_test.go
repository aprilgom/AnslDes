package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/report"
)

func TestWriteTextJSONAndSARIFDeterministically(t *testing.T) {
	t.Parallel()
	finding := diagnostic.New(
		"source/raw-value", diagnostic.SeverityError, "raw color", "src/A.tsx",
		&diagnostic.Range{Start: diagnostic.Position{Line: 2, Column: 3}, End: diagnostic.Position{Line: 2, Column: 12}},
		diagnostic.EvidenceSource, "react-native", "ansldes/source", "raw",
	)
	value := report.New("example-product", []report.EvidenceStatus{{Kind: diagnostic.EvidenceSource, Status: "acquired"}}, []diagnostic.Diagnostic{finding}, true)

	for _, format := range []report.Format{report.FormatText, report.FormatJSON, report.FormatSARIF} {
		var first, second bytes.Buffer
		if err := report.Write(&first, value, format); err != nil {
			t.Fatalf("Write(%s) error = %v", format, err)
		}
		if err := report.Write(&second, value, format); err != nil {
			t.Fatalf("Write(%s) second error = %v", format, err)
		}
		if first.String() != second.String() {
			t.Fatalf("Write(%s) is not deterministic", format)
		}
	}

	var jsonOutput bytes.Buffer
	if err := report.Write(&jsonOutput, value, report.FormatJSON); err != nil {
		t.Fatal(err)
	}
	var decoded report.Report
	if err := json.Unmarshal(jsonOutput.Bytes(), &decoded); err != nil || decoded.SchemaVersion != 1 {
		t.Fatalf("JSON report decode error = %v, value = %#v", err, decoded)
	}
	var sarif bytes.Buffer
	if err := report.Write(&sarif, value, report.FormatSARIF); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sarif.String(), `"version": "2.1.0"`) || !strings.Contains(sarif.String(), `"ruleId": "source/raw-value"`) {
		t.Fatalf("SARIF = %s", sarif.String())
	}
}
