package governance

import (
	"errors"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/policy"
)

func TestVerifyRejectsEachConfigAndReportBypass(t *testing.T) {
	config := policy.GovernancePolicy{
		ForbiddenFlags:   []string{"--no-config", "--no-design-system", "--no-inline-ignores", "--no-advisory"},
		RequireExitCode2: true, RequireUnmodifiedReport: true, PassingReportsOnly: true,
	}
	valid := Execution{
		Arguments: []string{"lint"}, ToolExitCode: 0, WrapperExitCode: 0, ReportStatus: "pass", ReportStored: true,
		OriginalReportSHA256: "same", StoredReportSHA256: "same", OriginalFindingCount: 0, StoredFindingCount: 0,
	}
	cases := map[string]struct {
		mutate func(*Execution)
		code   string
	}{
		"forbidden flag": {func(value *Execution) { value.Arguments = append(value.Arguments, "--no-config") }, "forbidden-flag"},
		"exit wrapper":   {func(value *Execution) { value.ToolExitCode, value.WrapperExitCode = 2, 0 }, "exit-code-rewrite"},
		"JSON deletion":  {func(value *Execution) { value.StoredReportSHA256 = "changed" }, "report-mutated"},
		"count rewrite":  {func(value *Execution) { value.StoredFindingCount = 1 }, "finding-count-rewritten"},
		"forced storage": {func(value *Execution) { value.ReportStatus = "fail" }, "forced-pass-storage"},
	}
	for name, item := range cases {
		t.Run(name, func(t *testing.T) {
			value := valid
			item.mutate(&value)
			var violation *ViolationError
			if err := Verify(config, value); !errors.As(err, &violation) || violation.Code != item.code {
				t.Fatalf("Verify() = %v", err)
			}
		})
	}
	if err := Verify(config, valid); err != nil {
		t.Fatalf("Verify(valid) = %v", err)
	}
}
