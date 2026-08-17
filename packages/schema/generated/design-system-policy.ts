// Code generated from https://ansldes.dev/schema/design-system-policy.v1.json; DO NOT EDIT.
// policy schema SHA-256: b0af57833aa55b97ccd9f7f628a4734fc05ecaa61c12c09492b2a41531ef2295

export const policySchemaVersion = 1 as const;
export const policySchemaSha256 =
  "b0af57833aa55b97ccd9f7f628a4734fc05ecaa61c12c09492b2a41531ef2295" as const;

export type Severity = "error" | "warning";
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
  | "source"
  | "pencil"
  | "computed-layout";

export interface DesignSystemPolicy {
  schemaVersion: typeof policySchemaVersion;
  definitionId: string;
  profile?: ConsumerProfile;
  severities: Record<string, Severity>;
  source: {
    rawProperties: { color: string[]; number: string[]; motion: string[] };
    exactExcludes: string[];
  };
  content?: ContentPolicy;
  assets?: AssetPolicy;
  runtime?: RuntimePolicy;
  native?: NativePolicy;
  web?: WebPolicy;
  evidence: {
    requiredKinds: EvidenceKind[];
    deferredKinds?: Exclude<EvidenceKind, "definition" | "source" | "pencil" | "computed-layout">[];
    layoutDocumentSha256?: string;
  };
  budgets: {
    error: number;
    warning: number;
    raw: number;
    overflow: number;
    overlap: number;
    blocking: number;
    advisory: number;
    exception: number;
    notRun: number;
    deferred: number;
  };
  exceptions: Array<{
    ruleId: string;
    engine: EvidenceKind;
    platform: string;
    path: string;
    owner: string;
    rationale: string;
    reviewer: string;
    expiresAt: string;
    reviewTrigger: string;
  }>;
  rulePacks: RulePackRequirement[];
  ruleOverrides: RuleOverride[];
  governance: GovernancePolicy;
}

export interface ContentPolicy {
  registryVersion: string;
  locales: Record<string, LocaleContentPolicy>;
  sourceReferences: string[];
}

export interface LocaleContentPolicy {
  phraseRegistryVersion: string;
  marketingBuzzwords: string[];
  theaterPhrases: string[];
  protectedTerms: string[];
  recoveryCopyIds: string[];
}

export interface AssetPolicy {
  registryVersion: string;
  entries: Record<string, AssetEntry>;
}

export interface AssetEntry {
  owner: string;
  role: "icon" | "logo" | "data-diagram" | "hero-illustration" | "photo" | "video-poster";
  implementationSource: string;
  consumers: string[];
  fingerprintSha256: string;
  intentionallyOmitted: boolean;
  decorative: boolean;
}

export interface RuntimePolicy {
  registryVersion: string;
  justifiedTextExceptions: JustifiedTextException[];
}

export interface JustifiedTextException {
  platform: "web" | "ios" | "android";
  surfaceId: string;
  routeId: string;
  nodeId: string;
  owner: string;
  context: "print" | "export";
}

export interface NativePolicy {
  registryVersion: string;
  iosAdjacentTargetSpacing: number;
  thresholds: NativeThresholds;
  requiredRuntimeCaptures: NativeRuntimeRequirement[];
}

export interface NativeThresholds {
  maxSynchronousStartupMs: number;
  maxInitializationMs: number;
  maxMainThreadWorkMs: number;
  maxFrameDropRatio: number;
  maxThumbnailDecodeRatio: number;
  maxJsBundleRegressionBytes: number;
  maxAppBinaryRegressionBytes: number;
}

export interface NativeRuntimeRequirement {
  id: string;
  platform: "ios" | "android";
  evidenceKind: "simulator" | "emulator" | "physical-device";
  formFactor: "phone" | "tablet" | "foldable";
  orientation: "portrait" | "landscape";
  windowMode: "fullscreen" | "split" | "multi-window";
  foldPosture: "not-applicable" | "flat" | "half-open";
  theme: "light" | "dark";
  minimumFontScale: number;
  reduceMotion: boolean;
}

export interface WebPolicy {
  registryVersion: string;
  buildCommand: string;
  routes: WebRoute[];
  viewports: WebViewport[];
  themes: string[];
  fontScales: number[];
  reduceMotion: boolean[];
  requiredCaptures: WebCaptureRequirement[];
  artifactExclusions: WebArtifactExclusion[];
}

export interface WebRoute {
  id: string;
  owner: string;
  target: string;
}

export interface WebViewport {
  id: string;
  width: number;
  height: number;
}

export interface WebCaptureRequirement {
  id: string;
  provider: "regex-source" | "static-html" | "browser" | "visual-contrast";
  routeId: string;
  viewportId: string;
  theme: string;
  fontScale: number;
  reduceMotion: boolean;
}

export interface WebArtifactExclusion {
  path: string;
  fingerprintSha256: string;
  owner: string;
  rationale: string;
  reproductionCommand: string;
}

export interface ConsumerProfile {
  id: "operate" | "read" | "browse" | "create" | string;
  primaryUserGoal: string;
  density: "compact" | "comfortable" | "spacious";
  noveltyTolerance: "low" | "medium" | "high";
  nativeAffordancePriority: "required" | "preferred" | "neutral";
  severityOverrides: Record<string, Severity>;
  thresholds: { maxOversizedActions: number; maxInconsistentActions: number };
  requiredEvidence: EvidenceKind[];
  rationale?: string;
  reviewer?: string;
  evidenceOwner?: string;
}

export interface RulePackRequirement {
  id: string;
  version: string;
  fingerprintSha256: string;
}

export interface RuleOverride {
  ruleId: string;
  packId: string;
  packVersion: string;
  status: "disabled";
  owner: string;
  rationale: string;
  reviewer: string;
  expiresAt: string;
  reviewTrigger: string;
}

export interface GovernancePolicy {
  reviewedAt: string;
  reviewIntervalDays: number;
  reviewer: string;
  reviewSubjects: Array<"impeccable-version" | "hallmark-commit" | "rule-mapping-drift" | "exceptions">;
  advisoryMode: "report";
  forbiddenFlags: Array<"--no-config" | "--no-design-system" | "--no-inline-ignores" | "--no-advisory">;
  requireExitCode2: true;
  requireUnmodifiedReport: true;
  passingReportsOnly: true;
  ignores: IgnoreAllowance[];
}

export interface IgnoreAllowance {
  kind: "rule" | "file" | "value" | "inline";
  ruleId?: string;
  engine: EvidenceKind;
  platform: string;
  path: string;
  property?: string;
  value?: string;
  owner: string;
  rationale: string;
  reviewer: string;
  expiresAt: string;
  reviewTrigger: string;
}
