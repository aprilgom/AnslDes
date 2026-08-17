// Package report renders deterministic text, JSON, and SARIF output.
package report

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
)

const schemaURL = "https://ansldes.dev/schema/deslint-report.v1.json"

// Format is an external report encoding.
type Format string

const (
	// FormatText emits one stable human-readable diagnostic per line.
	FormatText Format = "text"
	// FormatJSON emits the native deslint report contract.
	FormatJSON Format = "json"
	// FormatSARIF emits SARIF 2.1.0 for code-scanning consumers.
	FormatSARIF Format = "sarif"
)

// EvidenceStatus records a provider result independently for each evidence kind and platform.
type EvidenceStatus struct {
	Kind     diagnostic.EvidenceKind `json:"kind"`
	Platform string                  `json:"platform"`
	Status   EvidenceStatusValue     `json:"status"`
	Path     string                  `json:"path,omitempty"`
}

// RulePack identifies one exact versioned rule manifest.
type RulePack struct {
	ID                string `json:"id"`
	Version           string `json:"version"`
	FingerprintSHA256 string `json:"fingerprintSha256"`
}

// RuleActivation records whether one canonical rule participates in the run.
type RuleActivation struct {
	RuleID string               `json:"ruleId"`
	Status RuleActivationStatus `json:"status"`
	Reason string               `json:"reason,omitempty"`
}

// RuleSet records the exact effective pack and rule selection.
type RuleSet struct {
	FingerprintSHA256 string           `json:"fingerprintSha256"`
	Packs             []RulePack       `json:"packs"`
	Rules             []RuleActivation `json:"rules"`
}

// FalsePositive preserves an exact, policy-owned classification without hiding the source finding.
type FalsePositive struct {
	RuleID             string                  `json:"ruleId"`
	FindingFingerprint string                  `json:"findingFingerprint"`
	Owner              string                  `json:"owner"`
	OwnerFingerprint   string                  `json:"ownerFingerprint"`
	Rationale          string                  `json:"rationale"`
	Path               string                  `json:"path"`
	EvidenceKind       diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform           string                  `json:"platform"`
	Status             EvidenceStatusValue     `json:"status"`
}

// VisualJudgment is optional review evidence and never contributes to the deterministic fingerprint.
type VisualJudgment struct {
	ID           string                  `json:"id"`
	Status       JudgmentStatus          `json:"status"`
	EvidenceKind diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform     string                  `json:"platform"`
	Reviewer     string                  `json:"reviewer,omitempty"`
	Note         string                  `json:"note,omitempty"`
}

// Summary contains exact budget counters.
type Summary struct {
	Errors         int `json:"errors"`
	Warnings       int `json:"warnings"`
	Raw            int `json:"raw"`
	Overflow       int `json:"overflow"`
	Overlap        int `json:"overlap"`
	FalsePositives int `json:"falsePositives"`
}

// StageExecution preserves the exact provider command result and dependency freshness evidence.
type StageExecution struct {
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
}

// Input contains all deterministic and optional records used to construct a report.
type Input struct {
	DefinitionID    string
	RuleSet         RuleSet
	Evidence        []EvidenceStatus
	Diagnostics     []diagnostic.Diagnostic
	FalsePositives  []FalsePositive
	VisualJudgments []VisualJudgment
	StageExecutions []StageExecution
	Failed          bool
}

// Report is the deterministic native JSON contract.
type Report struct {
	Schema            string                  `json:"$schema,omitempty"`
	SchemaVersion     int                     `json:"schemaVersion"`
	Status            Status                  `json:"status"`
	DefinitionID      string                  `json:"definitionId"`
	FingerprintSHA256 string                  `json:"fingerprintSha256"`
	RuleSet           RuleSet                 `json:"ruleSet"`
	Evidence          []EvidenceStatus        `json:"evidence"`
	Summary           Summary                 `json:"summary"`
	Diagnostics       []diagnostic.Diagnostic `json:"diagnostics"`
	FalsePositives    []FalsePositive         `json:"falsePositives"`
	VisualJudgments   []VisualJudgment        `json:"visualJudgments"`
	StageExecutions   []StageExecution        `json:"stageExecutions"`
}

// NewRulePack returns a manifest fingerprint over an exact sorted member set.
func NewRulePack(id, version string, members []string) RulePack {
	sortedMembers := append([]string(nil), members...)
	sort.Strings(sortedMembers)
	contents, _ := json.Marshal(struct {
		ID      string   `json:"id"`
		Version string   `json:"version"`
		Members []string `json:"members"`
	}{ID: id, Version: version, Members: sortedMembers})
	return RulePack{ID: id, Version: version, FingerprintSHA256: hash(contents)}
}

// NewActiveRuleSet creates one pack whose members are all active.
func NewActiveRuleSet(packID, version string, ruleIDs []string) (RuleSet, error) {
	pack := NewRulePack(packID, version, ruleIDs)
	activations := make([]RuleActivation, 0, len(ruleIDs))
	for _, ruleID := range ruleIDs {
		activations = append(activations, RuleActivation{RuleID: ruleID, Status: RuleActive})
	}
	return normalizeRuleSet(RuleSet{Packs: []RulePack{pack}, Rules: activations})
}

// NewRuleSet creates a canonical effective rule set from explicit packs and activations.
func NewRuleSet(packs []RulePack, activations []RuleActivation) (RuleSet, error) {
	return normalizeRuleSet(RuleSet{Packs: packs, Rules: activations})
}

// NewFalsePositive creates an exact classification tied to one finding and owner.
func NewFalsePositive(finding diagnostic.Diagnostic, owner, rationale string) FalsePositive {
	ownerFingerprint := hash([]byte(owner + "\x00" + finding.Fingerprint))
	return FalsePositive{
		RuleID: finding.RuleID, FindingFingerprint: finding.Fingerprint,
		Owner: owner, OwnerFingerprint: ownerFingerprint, Rationale: rationale,
		Path: finding.Path, EvidenceKind: finding.EvidenceKind, Platform: finding.Platform,
		Status: EvidenceStatusFalsePositive,
	}
}

// New creates a sorted report with a deterministic fingerprint.
func New(input Input) (Report, error) {
	ruleSet, err := normalizeRuleSet(input.RuleSet)
	if err != nil {
		return Report{}, err
	}
	diagnostics := diagnostic.MergeCanonical(append([]diagnostic.Diagnostic(nil), input.Diagnostics...))
	if validationErr := validateDiagnostics(diagnostics, ruleSet); validationErr != nil {
		return Report{}, validationErr
	}
	falsePositives, err := normalizeFalsePositives(input.FalsePositives)
	if err != nil {
		return Report{}, err
	}
	if err := validateFalsePositiveActivations(falsePositives, ruleSet); err != nil {
		return Report{}, err
	}
	evidence, err := normalizeEvidence(input.Evidence, diagnostics, falsePositives)
	if err != nil {
		return Report{}, err
	}
	visualJudgments, err := normalizeVisualJudgments(input.VisualJudgments)
	if err != nil {
		return Report{}, err
	}
	stageExecutions, err := normalizeStageExecutions(input.StageExecutions)
	if err != nil {
		return Report{}, err
	}
	status := StatusPass
	if input.Failed {
		status = StatusFail
	}
	result := Report{
		Schema: schemaURL, SchemaVersion: SchemaVersion, Status: status,
		DefinitionID: input.DefinitionID, RuleSet: ruleSet, Evidence: evidence,
		Summary: summarize(diagnostics, falsePositives), Diagnostics: diagnostics,
		FalsePositives: falsePositives, VisualJudgments: visualJudgments,
		StageExecutions: stageExecutions,
	}
	result.FingerprintSHA256 = fingerprint(result)
	return result, nil
}

func normalizeStageExecutions(values []StageExecution) ([]StageExecution, error) {
	result := append([]StageExecution(nil), values...)
	if result == nil {
		result = []StageExecution{}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].StageID < result[j].StageID })
	for index, value := range result {
		if value.StageID == "" || value.Owner == "" || value.Platform == "" || len(value.Command) == 0 ||
			!validSHA256(value.DependencySHA256) || !validSHA256(value.ObservedDependencySHA256) ||
			(value.Status != "pass" && value.Status != "fail") || (value.Status == "pass" && value.ExitCode != 0) ||
			(value.Status == "fail" && value.ExitCode == 0) {
			return nil, fmt.Errorf("stage execution %q is invalid", value.StageID)
		}
		if index > 0 && result[index-1].StageID == value.StageID {
			return nil, fmt.Errorf("duplicate stage execution %q", value.StageID)
		}
	}
	return result, nil
}

func validateFalsePositiveActivations(values []FalsePositive, ruleSet RuleSet) error {
	activations := make(map[string]RuleActivationStatus, len(ruleSet.Rules))
	for _, rule := range ruleSet.Rules {
		activations[rule.RuleID] = rule.Status
	}
	for _, value := range values {
		if activations[value.RuleID] != RuleActive {
			return fmt.Errorf("false-positive rule %q is not active in the effective rule set", value.RuleID)
		}
	}
	return nil
}

// Write emits a report in the requested format.
func Write(writer io.Writer, value Report, format Format) error {
	switch format {
	case FormatText:
		return writeText(writer, value)
	case FormatJSON:
		return writeJSON(writer, value)
	case FormatSARIF:
		return writeSARIF(writer, value)
	default:
		return fmt.Errorf("unsupported report format %q", format)
	}
}

func normalizeRuleSet(value RuleSet) (RuleSet, error) {
	value.Packs = append([]RulePack(nil), value.Packs...)
	value.Rules = append([]RuleActivation(nil), value.Rules...)
	if value.Packs == nil {
		value.Packs = []RulePack{}
	}
	if value.Rules == nil {
		value.Rules = []RuleActivation{}
	}
	sort.SliceStable(value.Packs, func(i, j int) bool {
		if value.Packs[i].ID != value.Packs[j].ID {
			return value.Packs[i].ID < value.Packs[j].ID
		}
		return value.Packs[i].Version < value.Packs[j].Version
	})
	sort.SliceStable(value.Rules, func(i, j int) bool { return value.Rules[i].RuleID < value.Rules[j].RuleID })
	for index, pack := range value.Packs {
		if pack.ID == "" || pack.Version == "" || !validSHA256(pack.FingerprintSHA256) {
			return RuleSet{}, fmt.Errorf("rule pack id, version, and fingerprint are required")
		}
		if index > 0 && value.Packs[index-1].ID == pack.ID {
			return RuleSet{}, fmt.Errorf("duplicate rule pack %q", pack.ID)
		}
	}
	for index, rule := range value.Rules {
		if rule.RuleID == "" || !validActivation(rule.Status) {
			return RuleSet{}, fmt.Errorf("rule activation for %q is invalid", rule.RuleID)
		}
		if rule.Status != RuleActive && rule.Reason == "" {
			return RuleSet{}, fmt.Errorf("inactive rule %q requires a reason", rule.RuleID)
		}
		if index > 0 && value.Rules[index-1].RuleID == rule.RuleID {
			return RuleSet{}, fmt.Errorf("duplicate rule activation %q", rule.RuleID)
		}
	}
	value.FingerprintSHA256 = ""
	contents, _ := json.Marshal(value)
	value.FingerprintSHA256 = hash(contents)
	return value, nil
}

func normalizeEvidence(values []EvidenceStatus, diagnostics []diagnostic.Diagnostic, falsePositives []FalsePositive) ([]EvidenceStatus, error) {
	result := append([]EvidenceStatus(nil), values...)
	if result == nil {
		result = []EvidenceStatus{}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		if result[i].Platform != result[j].Platform {
			return result[i].Platform < result[j].Platform
		}
		return result[i].Path < result[j].Path
	})
	for index := range result {
		record := &result[index]
		if !validEvidence(record.Kind) || record.Platform == "" || !validEvidenceStatus(record.Status) {
			return nil, fmt.Errorf("evidence %q for platform %q is invalid", record.Kind, record.Platform)
		}
		if record.Kind == diagnostic.EvidenceWebSource || record.Kind == diagnostic.EvidenceWebRendered {
			if record.Platform != "web" {
				return nil, fmt.Errorf("web evidence %q cannot use platform %q", record.Kind, record.Platform)
			}
		} else if record.Platform == "web" {
			return nil, fmt.Errorf("non-Web evidence %q cannot use platform web", record.Kind)
		}
		if index > 0 && result[index-1].Kind == record.Kind && result[index-1].Platform == record.Platform && result[index-1].Path == record.Path {
			return nil, fmt.Errorf("duplicate evidence %q for platform %q", record.Kind, record.Platform)
		}
		if record.Status != EvidenceStatusPass {
			continue
		}
		hasAdvisory := false
		for _, finding := range diagnostics {
			if !sameEvidence(record.Kind, record.Platform, finding.EvidenceKind, finding.Platform) {
				continue
			}
			if finding.Status == diagnostic.FindingFail {
				record.Status = EvidenceStatusFail
				break
			}
			hasAdvisory = true
		}
		if record.Status == EvidenceStatusPass && hasAdvisory {
			record.Status = EvidenceStatusAdvisory
		}
		if record.Status == EvidenceStatusPass {
			for _, finding := range falsePositives {
				if sameEvidence(record.Kind, record.Platform, finding.EvidenceKind, finding.Platform) {
					record.Status = EvidenceStatusFalsePositive
					break
				}
			}
		}
	}
	return result, nil
}

func normalizeFalsePositives(values []FalsePositive) ([]FalsePositive, error) {
	result := append([]FalsePositive(nil), values...)
	if result == nil {
		result = []FalsePositive{}
	}
	for index := range result {
		result[index].Status = EvidenceStatusFalsePositive
		finding := result[index]
		if finding.RuleID == "" || !validSHA256(finding.FindingFingerprint) || finding.Owner == "" ||
			!validSHA256(finding.OwnerFingerprint) || len(finding.Rationale) < 8 || finding.Path == "" ||
			!validEvidence(finding.EvidenceKind) || finding.Platform == "" {
			return nil, fmt.Errorf("false-positive record for %q is incomplete", finding.RuleID)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		if result[i].RuleID != result[j].RuleID {
			return result[i].RuleID < result[j].RuleID
		}
		return result[i].FindingFingerprint < result[j].FindingFingerprint
	})
	return slices.CompactFunc(result, func(left, right FalsePositive) bool {
		return left.FindingFingerprint == right.FindingFingerprint && left.OwnerFingerprint == right.OwnerFingerprint
	}), nil
}

func normalizeVisualJudgments(values []VisualJudgment) ([]VisualJudgment, error) {
	result := append([]VisualJudgment(nil), values...)
	if result == nil {
		result = []VisualJudgment{}
	}
	for _, judgment := range result {
		if judgment.ID == "" || !validJudgmentStatus(judgment.Status) ||
			!validEvidence(judgment.EvidenceKind) || judgment.Platform == "" {
			return nil, fmt.Errorf("visual judgment %q is invalid", judgment.ID)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ID != result[j].ID {
			return result[i].ID < result[j].ID
		}
		if result[i].EvidenceKind != result[j].EvidenceKind {
			return result[i].EvidenceKind < result[j].EvidenceKind
		}
		return result[i].Platform < result[j].Platform
	})
	return result, nil
}

func validateDiagnostics(diagnostics []diagnostic.Diagnostic, ruleSet RuleSet) error {
	activations := make(map[string]RuleActivationStatus, len(ruleSet.Rules))
	for _, rule := range ruleSet.Rules {
		activations[rule.RuleID] = rule.Status
	}
	for _, finding := range diagnostics {
		if finding.RuleID == "" || finding.Message == "" || finding.Path == "" ||
			finding.Owner == "" || finding.Platform == "" || !validEvidence(finding.EvidenceKind) ||
			!validSHA256(finding.Fingerprint) ||
			(finding.Status != diagnostic.FindingFail && finding.Status != diagnostic.FindingAdvisory) {
			return fmt.Errorf("diagnostic for rule %q is incomplete", finding.RuleID)
		}
		if activations[finding.RuleID] != RuleActive {
			return fmt.Errorf("diagnostic rule %q is not active in the effective rule set", finding.RuleID)
		}
	}
	return nil
}

func summarize(diagnostics []diagnostic.Diagnostic, falsePositives []FalsePositive) Summary {
	summary := Summary{FalsePositives: len(falsePositives)}
	for _, finding := range diagnostics {
		if finding.Severity == diagnostic.SeverityError {
			summary.Errors++
		} else {
			summary.Warnings++
		}
		if finding.Category == "raw" {
			summary.Raw++
		}
		if finding.Category == "overflow" || finding.Category == "clipping" {
			summary.Overflow++
		}
		if finding.Category == "overlap" {
			summary.Overlap++
		}
	}
	return summary
}

func writeText(writer io.Writer, value Report) error {
	if _, err := fmt.Fprintf(writer, "RULESET %s\n", value.RuleSet.FingerprintSHA256); err != nil {
		return err
	}
	for _, pack := range value.RuleSet.Packs {
		if _, err := fmt.Fprintf(writer, "PACK %s@%s %s\n", pack.ID, pack.Version, pack.FingerprintSHA256); err != nil {
			return err
		}
	}
	for _, rule := range value.RuleSet.Rules {
		if _, err := fmt.Fprintf(writer, "RULE %s %s reason=%q\n", rule.RuleID, rule.Status, rule.Reason); err != nil {
			return err
		}
	}
	for _, finding := range value.Diagnostics {
		line, column := 0, 0
		if finding.Range != nil {
			line, column = finding.Range.Start.Line, finding.Range.Start.Column
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s:%d:%d %s %s %s [%s/%s viewport=%q owner=%q]\n",
			finding.Path,
			line,
			column,
			finding.Severity,
			finding.RuleID,
			finding.Message,
			finding.Platform,
			finding.EvidenceKind,
			finding.Viewport,
			finding.Owner,
		); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(
		writer,
		"%s errors=%d warnings=%d raw=%d overflow=%d overlap=%d false-positives=%d rules=%s fingerprint=%s\n",
		strings.ToUpper(string(value.Status)),
		value.Summary.Errors,
		value.Summary.Warnings,
		value.Summary.Raw,
		value.Summary.Overflow,
		value.Summary.Overlap,
		value.Summary.FalsePositives,
		value.RuleSet.FingerprintSHA256,
		value.FingerprintSHA256,
	)
	return err
}

func writeJSON(writer io.Writer, value Report) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool       sarifTool          `json:"tool"`
	Results    []sarifResult      `json:"results"`
	Properties sarifRunProperties `json:"properties"`
}

type sarifRunProperties struct {
	ReportFingerprintSHA256 string  `json:"reportFingerprintSha256"`
	RuleSet                 RuleSet `json:"ruleSet"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name  string      `json:"name"`
	Rules []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID string `json:"id"`
}

type sarifResult struct {
	RuleID     string                `json:"ruleId"`
	Level      string                `json:"level"`
	Message    sarifMessage          `json:"message"`
	Locations  []sarifLocation       `json:"locations"`
	Partial    sarifPartial          `json:"partialFingerprints"`
	Properties sarifResultProperties `json:"properties"`
}

type sarifResultProperties struct {
	Status       diagnostic.FindingStatus `json:"status"`
	EvidenceKind diagnostic.EvidenceKind  `json:"evidenceKind"`
	Platform     string                   `json:"platform"`
	Viewport     string                   `json:"viewport,omitempty"`
	Owner        string                   `json:"owner"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	Physical sarifPhysical `json:"physicalLocation"`
}

type sarifPhysical struct {
	Artifact sarifArtifact `json:"artifactLocation"`
	Region   *sarifRegion  `json:"region,omitempty"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
	EndLine     int `json:"endLine"`
	EndColumn   int `json:"endColumn"`
}

type sarifPartial struct {
	Primary string `json:"primaryLocationLineHash"`
}

func writeSARIF(writer io.Writer, value Report) error {
	ruleSet := make(map[string]struct{})
	results := make([]sarifResult, 0, len(value.Diagnostics))
	for _, finding := range value.Diagnostics {
		ruleSet[finding.RuleID] = struct{}{}
		level := "error"
		if finding.Severity == diagnostic.SeverityWarning {
			level = "warning"
		}
		var region *sarifRegion
		if finding.Range != nil {
			region = &sarifRegion{
				StartLine: finding.Range.Start.Line, StartColumn: finding.Range.Start.Column,
				EndLine: finding.Range.End.Line, EndColumn: finding.Range.End.Column,
			}
		}
		results = append(results, sarifResult{
			RuleID:  finding.RuleID,
			Level:   level,
			Message: sarifMessage{Text: finding.Message},
			Locations: []sarifLocation{{Physical: sarifPhysical{
				Artifact: sarifArtifact{URI: finding.Path}, Region: region,
			}}},
			Partial: sarifPartial{Primary: finding.Fingerprint},
			Properties: sarifResultProperties{
				Status: finding.Status, EvidenceKind: finding.EvidenceKind, Platform: finding.Platform,
				Viewport: finding.Viewport, Owner: finding.Owner,
			},
		})
	}
	ruleIDs := make([]string, 0, len(ruleSet))
	for ruleID := range ruleSet {
		ruleIDs = append(ruleIDs, ruleID)
	}
	sort.Strings(ruleIDs)
	rules := make([]sarifRule, 0, len(ruleIDs))
	for _, ruleID := range ruleIDs {
		rules = append(rules, sarifRule{ID: ruleID})
	}
	log := sarifLog{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []sarifRun{{
			Tool:       sarifTool{Driver: sarifDriver{Name: "deslint", Rules: rules}},
			Results:    results,
			Properties: sarifRunProperties{ReportFingerprintSHA256: value.FingerprintSHA256, RuleSet: value.RuleSet},
		}},
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(log)
}

func fingerprint(value Report) string {
	clone := value
	clone.FingerprintSHA256 = ""
	clone.VisualJudgments = []VisualJudgment{}
	contents, _ := json.Marshal(clone)
	return hash(contents)
}

func validEvidence(kind diagnostic.EvidenceKind) bool {
	switch kind {
	case diagnostic.EvidenceDefinition,
		diagnostic.EvidenceWebSource,
		diagnostic.EvidenceWebRendered,
		diagnostic.EvidenceNativeSource,
		diagnostic.EvidenceDesignDocumentSource,
		diagnostic.EvidenceDesignDocumentComputed,
		diagnostic.EvidenceSimulator,
		diagnostic.EvidenceEmulator,
		diagnostic.EvidencePhysicalDevice,
		diagnostic.EvidenceConsumerConformance,
		diagnostic.EvidenceConsumerContentRegistry,
		diagnostic.EvidenceExecution:
		return true
	default:
		return false
	}
}

func validEvidenceStatus(status EvidenceStatusValue) bool {
	switch status {
	case EvidenceStatusPass, EvidenceStatusFail, EvidenceStatusAdvisory,
		EvidenceStatusFalsePositive, EvidenceStatusNotRun, EvidenceStatusDeferred:
		return true
	default:
		return false
	}
}

func validActivation(status RuleActivationStatus) bool {
	switch status {
	case RuleActive, RuleNotApplicable, RuleDisabled, RuleUnsupported:
		return true
	default:
		return false
	}
}

func validJudgmentStatus(status JudgmentStatus) bool {
	switch status {
	case JudgmentPass, JudgmentFail, JudgmentNotReviewed:
		return true
	default:
		return false
	}
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sameEvidence(leftKind diagnostic.EvidenceKind, leftPlatform string, rightKind diagnostic.EvidenceKind, rightPlatform string) bool {
	return leftKind == rightKind && (leftPlatform == rightPlatform || leftPlatform == "all" || rightPlatform == "all")
}

func hash(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
