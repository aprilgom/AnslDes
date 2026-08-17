// Code generated from https://ansldes.dev/schema/deslint-report.v1.json; DO NOT EDIT.
// report schema SHA-256: 911458d971f5fea50173a4b0d3f522c6a2de974662fec5f8c1ef5934dcfd4822

export const reportSchemaVersion = 1 as const;
export const reportSchemaSha256 =
  "911458d971f5fea50173a4b0d3f522c6a2de974662fec5f8c1ef5934dcfd4822" as const;

export type ReportStatus = "pass" | "fail";
export type EvidenceKind =
  | "definition"
  | "web-source"
  | "web-rendered"
  | "native-source"
  | "design-document-source"
  | "design-document-computed"
  | "simulator"
  | "emulator"
  | "physical-device"
  | "consumer-conformance"
  | "consumer-content-registry"
  | "execution";
export type EvidenceStatus =
  | "pass"
  | "fail"
  | "advisory"
  | "false-positive"
  | "not-run"
  | "deferred";
export type RuleActivationStatus =
  | "active"
  | "not-applicable"
  | "disabled"
  | "unsupported";

export interface DeslintReport {
  schemaVersion: typeof reportSchemaVersion;
  status: ReportStatus;
  definitionId: string;
  fingerprintSha256: string;
  ruleSet: {
    fingerprintSha256: string;
    packs: Array<{ id: string; version: string; fingerprintSha256: string }>;
    rules: Array<{ ruleId: string; status: RuleActivationStatus; reason?: string }>;
  };
  evidence: Array<{
    kind: EvidenceKind;
    platform: string;
    status: EvidenceStatus;
    path?: string;
  }>;
  summary: {
    errors: number;
    warnings: number;
    raw: number;
    overflow: number;
    overlap: number;
    falsePositives: number;
  };
  diagnostics: Array<{
    ruleId: string;
    sourceRuleIds: string[];
    status: "fail" | "advisory";
    severity: "error" | "warning";
    message: string;
    path: string;
    range?: {
      start: { line: number; column: number };
      end: { line: number; column: number };
    };
    evidenceKind: EvidenceKind;
    platform: string;
    viewport?: string;
    owner: string;
    fingerprint: string;
    category?: string;
  }>;
  falsePositives: Array<{
    ruleId: string;
    findingFingerprint: string;
    owner: string;
    ownerFingerprint: string;
    rationale: string;
    path: string;
    evidenceKind: EvidenceKind;
    platform: string;
    status: "false-positive";
  }>;
  visualJudgments: Array<{
    id: string;
    status: "pass" | "fail" | "not-reviewed";
    evidenceKind: EvidenceKind;
    platform: string;
    reviewer?: string;
    note?: string;
  }>;
  stageExecutions: Array<{
    stageId: string;
    owner: string;
    platform: string;
    command: string[];
    status: "pass" | "fail";
    exitCode: number;
    stdout: string;
    stderr: string;
    dependencySha256: string;
    observedDependencySha256: string;
  }>;
}
