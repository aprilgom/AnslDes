// Package stage validates independently captured provider process execution evidence.
package stage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.com/aprilgom/AnslDes/deslint/internal/report"
)

var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type envelope struct {
	SchemaVersion            int      `json:"schemaVersion"`
	StageID                  string   `json:"stageId"`
	Owner                    string   `json:"owner"`
	Platform                 string   `json:"platform"`
	Command                  []string `json:"command"`
	Status                   string   `json:"status"`
	ExitCode                 int      `json:"exitCode"`
	Stdout                   string   `json:"stdout"`
	Stderr                   string   `json:"stderr"`
	DependencySHA256         string   `json:"dependencySha256"`
	ObservedDependencySHA256 string   `json:"observedDependencySha256"`
	Schema                   string   `json:"$schema"`
}

// Parse rejects unknown fields, status/exit rewrites, and incomplete ownership or dependency evidence.
func Parse(contents []byte) (report.StageExecution, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var value envelope
	if err := decoder.Decode(&value); err != nil {
		return report.StageExecution{}, fmt.Errorf("parse stage execution: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return report.StageExecution{}, fmt.Errorf("stage execution must contain exactly one JSON value")
	}
	if value.SchemaVersion != 1 || value.StageID == "" || value.Owner == "" || value.Platform == "" || len(value.Command) == 0 ||
		!sha256Pattern.MatchString(value.DependencySHA256) || !sha256Pattern.MatchString(value.ObservedDependencySHA256) ||
		(value.Status != "pass" && value.Status != "fail") || (value.Status == "pass" && value.ExitCode != 0) ||
		(value.Status == "fail" && value.ExitCode == 0) {
		return report.StageExecution{}, fmt.Errorf("stage execution %q is invalid", value.StageID)
	}
	return report.StageExecution{
		StageID: value.StageID, Owner: value.Owner, Platform: value.Platform, Command: value.Command,
		Status: value.Status, ExitCode: value.ExitCode, Stdout: value.Stdout, Stderr: value.Stderr,
		DependencySHA256: value.DependencySHA256, ObservedDependencySHA256: value.ObservedDependencySHA256,
	}, nil
}

// Fresh reports whether the provider ran against the exact locked dependency.
func Fresh(value report.StageExecution) bool {
	return value.DependencySHA256 == value.ObservedDependencySHA256
}
