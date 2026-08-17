// Package governance verifies CI execution and report anti-bypass evidence.
package governance

import (
	"fmt"
	"slices"

	"github.com/aprilgom/AnslDes/deslint/internal/policy"
)

// Execution is one immutable wrapper and report-storage observation.
type Execution struct {
	Arguments            []string
	ToolExitCode         int
	WrapperExitCode      int
	ReportStatus         string
	ReportStored         bool
	OriginalReportSHA256 string
	StoredReportSHA256   string
	OriginalFindingCount int
	StoredFindingCount   int
}

// ViolationError carries a stable governance failure code.
type ViolationError struct {
	Code string
}

func (e *ViolationError) Error() string { return fmt.Sprintf("governance violation: %s", e.Code) }

// Verify rejects CLI, exit-code, report mutation, and forced-pass bypasses.
func Verify(config policy.GovernancePolicy, execution Execution) error {
	for _, argument := range execution.Arguments {
		if slices.Contains(config.ForbiddenFlags, argument) {
			return &ViolationError{Code: "forbidden-flag"}
		}
	}
	if config.RequireExitCode2 && execution.ToolExitCode == 2 && execution.WrapperExitCode != 2 {
		return &ViolationError{Code: "exit-code-rewrite"}
	}
	if config.RequireUnmodifiedReport && (execution.OriginalReportSHA256 == "" || execution.OriginalReportSHA256 != execution.StoredReportSHA256) {
		return &ViolationError{Code: "report-mutated"}
	}
	if execution.OriginalFindingCount < 0 || execution.OriginalFindingCount != execution.StoredFindingCount {
		return &ViolationError{Code: "finding-count-rewritten"}
	}
	if config.PassingReportsOnly && execution.ReportStored && execution.ReportStatus != "pass" {
		return &ViolationError{Code: "forced-pass-storage"}
	}
	return nil
}
