// Code generated from https://ansldes.dev/schema/deslint-report.v1.json; DO NOT EDIT.
// report schema SHA-256: 911458d971f5fea50173a4b0d3f522c6a2de974662fec5f8c1ef5934dcfd4822

package report

const SchemaVersion = 1
const SchemaSHA256 = "911458d971f5fea50173a4b0d3f522c6a2de974662fec5f8c1ef5934dcfd4822"

type Status string
type EvidenceStatusValue string
type RuleActivationStatus string
type JudgmentStatus string

const (
	StatusPass Status = "pass"
	StatusFail Status = "fail"

	EvidenceStatusPass          EvidenceStatusValue = "pass"
	EvidenceStatusFail          EvidenceStatusValue = "fail"
	EvidenceStatusAdvisory      EvidenceStatusValue = "advisory"
	EvidenceStatusFalsePositive EvidenceStatusValue = "false-positive"
	EvidenceStatusNotRun        EvidenceStatusValue = "not-run"
	EvidenceStatusDeferred      EvidenceStatusValue = "deferred"

	RuleActive        RuleActivationStatus = "active"
	RuleNotApplicable RuleActivationStatus = "not-applicable"
	RuleDisabled      RuleActivationStatus = "disabled"
	RuleUnsupported   RuleActivationStatus = "unsupported"

	JudgmentPass        JudgmentStatus = "pass"
	JudgmentFail        JudgmentStatus = "fail"
	JudgmentNotReviewed JudgmentStatus = "not-reviewed"
)
