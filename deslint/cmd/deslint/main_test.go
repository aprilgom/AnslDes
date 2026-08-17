package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/report"
)

func TestDevelopmentVersionMatchesReleaseCandidate(t *testing.T) {
	t.Parallel()
	if version != "0.1.0-dev" {
		t.Fatalf("version = %q", version)
	}
}

func TestExitCodeSeparatesFindingsFromExecutionErrors(t *testing.T) {
	t.Parallel()
	if exitCode(nil) != 0 || exitCode(errLintFailed) != 2 || exitCode(errors.New("execution")) != 1 {
		t.Fatalf("exit codes = success:%d finding:%d execution:%d", exitCode(nil), exitCode(errLintFailed), exitCode(errors.New("execution")))
	}
}

func TestLintDoesNotOverwriteReportWhenInputValidationFails(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	output := filepath.Join(directory, "report.json")
	if err := os.WriteFile(output, []byte("preserve-me"), 0o600); err != nil {
		t.Fatal(err)
	}
	invalidPolicy := filepath.Join(directory, "invalid-policy.json")
	if err := os.WriteFile(invalidPolicy, []byte(`{"schemaVersion": 99}`), 0o600); err != nil {
		t.Fatal(err)
	}
	err := run([]string{
		"lint",
		"--definition", "../../../packages/schema/testdata/example-product.json",
		"--policy", invalidPolicy,
		"--format", "json",
		"--out", output,
	})
	if err == nil {
		t.Fatal("run() error = nil")
	}
	// #nosec G304 -- output is a test-owned path under t.TempDir.
	contents, readErr := os.ReadFile(output)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(contents) != "preserve-me" {
		t.Fatalf("output = %q", contents)
	}
}

func TestLintWritesPassingJSONAndFailingReports(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	output := filepath.Join(directory, "report.json")
	base := []string{
		"lint",
		"--definition", "../../../packages/schema/testdata/example-product.json",
		"--policy", "../../../packages/schema/testdata/example-policy.json",
		"--source", "../../testdata/positive/Example.tsx",
		"--pencil", "../../testdata/positive/document.pen.json",
		"--layout", "../../testdata/positive/layout.json",
		"--conformance", "../../../packages/schema/testdata/operate-conformance.json",
		"--design-context", "../../../packages/schema/testdata/generated-design-context/.impeccable/design.json",
		"--format", "json",
		"--out", output,
	}
	if err := run(base); err != nil {
		t.Fatalf("run(pass) error = %v", err)
	}
	if value := readReport(t, output); value.Status != "pass" {
		t.Fatalf("pass report = %#v", value)
	}

	failing := append([]string(nil), base...)
	for index := range failing {
		if failing[index] == "../../testdata/positive/Example.tsx" {
			failing[index] = "../../testdata/negative/Raw.tsx"
		}
	}
	err := run(failing)
	if !errors.Is(err, errLintFailed) {
		t.Fatalf("run(fail) error = %v", err)
	}
	if value := readReport(t, output); value.Status != "fail" {
		t.Fatalf("fail report = %#v", value)
	}
}

func readReport(t *testing.T, path string) report.Report {
	t.Helper()
	// #nosec G304 -- callers provide test-owned fixture and temporary paths.
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value report.Report
	if err := json.Unmarshal(contents, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
