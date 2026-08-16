// Code generated from https://ansldes.dev/schema/design-system-policy.v1.json; DO NOT EDIT.
// policy schema SHA-256: 2b962a65f488095a2b784713d4b7ba64c29ee52b4f098c2824dcaffe4527a36b

export const policySchemaVersion = 1 as const;
export const policySchemaSha256 =
  "2b962a65f488095a2b784713d4b7ba64c29ee52b4f098c2824dcaffe4527a36b" as const;

export type Severity = "error" | "warning";
export type EvidenceKind =
  | "definition"
  | "source"
  | "pencil"
  | "computed-layout";

export interface DesignSystemPolicy {
  schemaVersion: typeof policySchemaVersion;
  definitionId: string;
  severities: Record<string, Severity>;
  source: {
    rawProperties: { color: string[]; number: string[]; motion: string[] };
    exactExcludes: string[];
  };
  evidence: { requiredKinds: EvidenceKind[]; layoutDocumentSha256?: string };
  budgets: {
    error: number;
    warning: number;
    raw: number;
    overflow: number;
    overlap: number;
  };
  exceptions: Array<{
    ruleId: string;
    path: string;
    owner: string;
    rationale: string;
    expiresAt: string;
  }>;
}
