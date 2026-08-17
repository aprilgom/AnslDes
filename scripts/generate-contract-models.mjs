import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const check = process.argv.includes("--check");
const definitionSchemaPath = path.join(
  root,
  "packages/schema/design-system-definition.schema.json",
);
const policySchemaPath = path.join(
  root,
  "packages/schema/design-system-policy.schema.json",
);
const reportSchemaPath = path.join(
  root,
  "packages/schema/deslint-report.schema.json",
);
const conformanceSchemaPath = path.join(
  root,
  "packages/schema/consumer-conformance.schema.json",
);
const definitionSchema = await readSchema(definitionSchemaPath, 2);
const policySchema = await readSchema(policySchemaPath, 1);
const reportSchema = await readSchema(reportSchemaPath, 1);
const conformanceSchema = await readSchema(conformanceSchemaPath, 1);

const outputs = new Map([
  [
    path.join(root, "packages/schema/generated/design-system-definition.ts"),
    definitionTypeScript(definitionSchema),
  ],
  [
    path.join(root, "packages/schema/generated/design-system-policy.ts"),
    policyTypeScript(policySchema),
  ],
  [
    path.join(root, "packages/schema/generated/deslint-report.ts"),
    reportTypeScript(reportSchema),
  ],
  [
    path.join(root, "packages/schema/generated/consumer-conformance.ts"),
    conformanceTypeScript(conformanceSchema),
  ],
  [
    path.join(root, "deslint/internal/contract/generated_definition.go"),
    definitionGo(definitionSchema),
  ],
  [
    path.join(root, "deslint/internal/policy/generated_policy.go"),
    policyGo(policySchema),
  ],
  [
    path.join(root, "deslint/internal/report/generated_report.go"),
    reportGo(reportSchema),
  ],
  [
    path.join(root, "deslint/internal/conformance/generated_conformance.go"),
    conformanceGo(conformanceSchema),
  ],
]);

for (const [outputPath, contents] of outputs) {
  if (check) {
    const actual = await readFile(outputPath, "utf8").catch(() => "");
    if (actual !== contents) {
      throw new Error(
        `${path.relative(root, outputPath)} is stale; run npm run generate:models`,
      );
    }
  } else {
    await mkdir(path.dirname(outputPath), { recursive: true });
    await writeFile(outputPath, contents);
  }
}

async function readSchema(schemaPath, expectedVersion) {
  const source = await readFile(schemaPath, "utf8");
  const canonicalSource = source.replaceAll("\r\n", "\n");
  const schema = JSON.parse(canonicalSource);
  const version = schema.properties?.schemaVersion?.const;
  if (version !== expectedVersion || typeof schema.$id !== "string") {
    throw new Error(
      `${schemaPath} must declare schemaVersion const ${expectedVersion} and $id`,
    );
  }
  return {
    id: schema.$id,
    sha256: createHash("sha256").update(canonicalSource).digest("hex"),
    version,
  };
}

function header(kind, schema) {
  return `// Code generated from ${schema.id}; DO NOT EDIT.\n// ${kind} schema SHA-256: ${schema.sha256}\n`;
}

function definitionTypeScript(schema) {
  return `${header("definition", schema)}
export const definitionSchemaVersion = ${schema.version} as const;
export const definitionSchemaSha256 =
  "${schema.sha256}" as const;

export type RecipeValue =
  | string
  | number
  | boolean
  | null
  | RecipeValue[]
  | { [key: string]: RecipeValue };

export interface ComponentDefinition {
  anatomy: string[];
  slots: Record<string, RecipeValue>;
  variants: Record<string, Record<string, RecipeValue>>;
  sizes: Record<string, Record<string, RecipeValue>>;
  states: string[];
  semantics: Record<string, RecipeValue>;
}

export interface DesignSystemDefinition {
  $schema?: string;
  schemaVersion: typeof definitionSchemaVersion;
  id: string;
  version: string;
  themes: { names: string[]; default: string };
  colorUsage?: ColorUsageDefinition;
  foundations: Record<string, unknown>;
  components: Record<string, ComponentDefinition>;
}

export interface ColorUsageDefinition {
  contrast: { body: number; large: number };
  approvedPalettes: Record<
    string,
    { contexts: string[]; themes: string[] }
  >;
}
`;
}

function policyTypeScript(schema) {
  return `${header("policy", schema)}
export const policySchemaVersion = ${schema.version} as const;
export const policySchemaSha256 =
  "${schema.sha256}" as const;

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
`;
}

function definitionGo(schema) {
  return `${header("definition", schema)}
package contract

import "encoding/json"

const DefinitionSchemaVersion = ${schema.version}
const DefinitionSchemaSHA256 = "${schema.sha256}"

type Definition struct {
\tSchema        string                         \`json:"$schema,omitempty"\`
\tSchemaVersion int                            \`json:"schemaVersion"\`
\tID            string                         \`json:"id"\`
\tVersion       string                         \`json:"version"\`
\tThemes        ThemeDefinition                \`json:"themes"\`
\tColorUsage    *ColorUsageDefinition          \`json:"colorUsage,omitempty"\`
\tFoundations   map[string]json.RawMessage     \`json:"foundations"\`
\tComponents    map[string]ComponentDefinition \`json:"components"\`
}

type ColorUsageDefinition struct {
\tContrast         ContrastDefinition         \`json:"contrast"\`
\tApprovedPalettes map[string]ApprovedPalette \`json:"approvedPalettes"\`
}

type ContrastDefinition struct {
\tBody  float64 \`json:"body"\`
\tLarge float64 \`json:"large"\`
}

type ApprovedPalette struct {
\tContexts []string \`json:"contexts"\`
\tThemes   []string \`json:"themes"\`
}

type ThemeDefinition struct {
\tNames   []string \`json:"names"\`
\tDefault string   \`json:"default"\`
}

type ComponentDefinition struct {
\tAnatomy   []string                   \`json:"anatomy"\`
\tSlots     map[string]json.RawMessage \`json:"slots"\`
\tVariants  map[string]json.RawMessage \`json:"variants"\`
\tSizes     map[string]json.RawMessage \`json:"sizes"\`
\tStates    []string                   \`json:"states"\`
\tSemantics map[string]json.RawMessage \`json:"semantics"\`
}
`;
}

function reportTypeScript(schema) {
  return `${header("report", schema)}
export const reportSchemaVersion = ${schema.version} as const;
export const reportSchemaSha256 =
  "${schema.sha256}" as const;

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
`;
}

function reportGo(schema) {
  return `${header("report", schema)}
package report

const SchemaVersion = ${schema.version}
const SchemaSHA256 = "${schema.sha256}"

type Status string
type EvidenceStatusValue string
type RuleActivationStatus string
type JudgmentStatus string

const (
\tStatusPass Status = "pass"
\tStatusFail Status = "fail"

\tEvidenceStatusPass          EvidenceStatusValue = "pass"
\tEvidenceStatusFail          EvidenceStatusValue = "fail"
\tEvidenceStatusAdvisory      EvidenceStatusValue = "advisory"
\tEvidenceStatusFalsePositive EvidenceStatusValue = "false-positive"
\tEvidenceStatusNotRun        EvidenceStatusValue = "not-run"
\tEvidenceStatusDeferred      EvidenceStatusValue = "deferred"

\tRuleActive        RuleActivationStatus = "active"
\tRuleNotApplicable RuleActivationStatus = "not-applicable"
\tRuleDisabled      RuleActivationStatus = "disabled"
\tRuleUnsupported   RuleActivationStatus = "unsupported"

\tJudgmentPass        JudgmentStatus = "pass"
\tJudgmentFail        JudgmentStatus = "fail"
\tJudgmentNotReviewed JudgmentStatus = "not-reviewed"
)
`;
}

function policyGo(schema) {
  return `${header("policy", schema)}
package policy

const SchemaVersion = ${schema.version}
const SchemaSHA256 = "${schema.sha256}"

type Policy struct {
\tSchema        string                \`json:"$schema,omitempty"\`
\tSchemaVersion int                   \`json:"schemaVersion"\`
\tDefinitionID  string                \`json:"definitionId"\`
\tProfile       *ConsumerProfile      \`json:"profile,omitempty"\`
\tSeverities    map[string]string     \`json:"severities"\`
\tSource        SourcePolicy          \`json:"source"\`
\tContent       *ContentPolicy        \`json:"content,omitempty"\`
\tAssets        *AssetPolicy          \`json:"assets,omitempty"\`
\tRuntime       *RuntimePolicy        \`json:"runtime,omitempty"\`
\tNative        *NativePolicy         \`json:"native,omitempty"\`
\tWeb           *WebPolicy            \`json:"web,omitempty"\`
\tEvidence      EvidencePolicy        \`json:"evidence"\`
\tBudgets       Budgets               \`json:"budgets"\`
\tExceptions    []Exception           \`json:"exceptions"\`
\tRulePacks     []RulePackRequirement \`json:"rulePacks"\`
\tRuleOverrides []RuleOverride        \`json:"ruleOverrides"\`
\tGovernance    GovernancePolicy      \`json:"governance"\`
}

type ContentPolicy struct {
\tRegistryVersion  string                         \`json:"registryVersion"\`
\tLocales          map[string]LocaleContentPolicy \`json:"locales"\`
\tSourceReferences []string                       \`json:"sourceReferences"\`
}

type LocaleContentPolicy struct {
\tPhraseRegistryVersion string   \`json:"phraseRegistryVersion"\`
\tMarketingBuzzwords    []string \`json:"marketingBuzzwords"\`
\tTheaterPhrases        []string \`json:"theaterPhrases"\`
\tProtectedTerms        []string \`json:"protectedTerms"\`
\tRecoveryCopyIDs       []string \`json:"recoveryCopyIds"\`
}

type AssetPolicy struct {
\tRegistryVersion string                \`json:"registryVersion"\`
\tEntries         map[string]AssetEntry \`json:"entries"\`
}

type AssetEntry struct {
\tOwner                string   \`json:"owner"\`
\tRole                 string   \`json:"role"\`
\tImplementationSource string   \`json:"implementationSource"\`
\tConsumers            []string \`json:"consumers"\`
\tFingerprintSHA256    string   \`json:"fingerprintSha256"\`
\tIntentionallyOmitted bool     \`json:"intentionallyOmitted"\`
\tDecorative           bool     \`json:"decorative"\`
}

type RuntimePolicy struct {
\tRegistryVersion         string                   \`json:"registryVersion"\`
\tJustifiedTextExceptions []JustifiedTextException \`json:"justifiedTextExceptions"\`
}

type JustifiedTextException struct {
\tPlatform  string \`json:"platform"\`
\tSurfaceID string \`json:"surfaceId"\`
\tRouteID   string \`json:"routeId"\`
\tNodeID    string \`json:"nodeId"\`
\tOwner     string \`json:"owner"\`
\tContext   string \`json:"context"\`
}

type NativePolicy struct {
\tRegistryVersion          string                     \`json:"registryVersion"\`
\tIOSAdjacentTargetSpacing float64                    \`json:"iosAdjacentTargetSpacing"\`
\tThresholds               NativeThresholds           \`json:"thresholds"\`
\tRequiredRuntimeCaptures  []NativeRuntimeRequirement \`json:"requiredRuntimeCaptures"\`
}

type NativeThresholds struct {
\tMaxSynchronousStartupMS     float64 \`json:"maxSynchronousStartupMs"\`
\tMaxInitializationMS         float64 \`json:"maxInitializationMs"\`
\tMaxMainThreadWorkMS         float64 \`json:"maxMainThreadWorkMs"\`
\tMaxFrameDropRatio           float64 \`json:"maxFrameDropRatio"\`
\tMaxThumbnailDecodeRatio     float64 \`json:"maxThumbnailDecodeRatio"\`
\tMaxJSBundleRegressionBytes  int     \`json:"maxJsBundleRegressionBytes"\`
\tMaxAppBinaryRegressionBytes int     \`json:"maxAppBinaryRegressionBytes"\`
}

type NativeRuntimeRequirement struct {
\tID               string  \`json:"id"\`
\tPlatform         string  \`json:"platform"\`
\tEvidenceKind     string  \`json:"evidenceKind"\`
\tFormFactor       string  \`json:"formFactor"\`
\tOrientation      string  \`json:"orientation"\`
\tWindowMode       string  \`json:"windowMode"\`
\tFoldPosture      string  \`json:"foldPosture"\`
\tTheme            string  \`json:"theme"\`
\tMinimumFontScale float64 \`json:"minimumFontScale"\`
\tReduceMotion     bool    \`json:"reduceMotion"\`
}

type WebPolicy struct {
\tRegistryVersion    string                  \`json:"registryVersion"\`
\tBuildCommand       string                  \`json:"buildCommand"\`
\tRoutes             []WebRoute              \`json:"routes"\`
\tViewports          []WebViewport           \`json:"viewports"\`
\tThemes             []string                \`json:"themes"\`
\tFontScales         []float64               \`json:"fontScales"\`
\tReduceMotion       []bool                  \`json:"reduceMotion"\`
\tRequiredCaptures   []WebCaptureRequirement \`json:"requiredCaptures"\`
\tArtifactExclusions []WebArtifactExclusion  \`json:"artifactExclusions"\`
}

type WebRoute struct {
\tID     string \`json:"id"\`
\tOwner  string \`json:"owner"\`
\tTarget string \`json:"target"\`
}

type WebViewport struct {
\tID     string \`json:"id"\`
\tWidth  int    \`json:"width"\`
\tHeight int    \`json:"height"\`
}

type WebCaptureRequirement struct {
\tID           string  \`json:"id"\`
\tProvider     string  \`json:"provider"\`
\tRouteID      string  \`json:"routeId"\`
\tViewportID   string  \`json:"viewportId"\`
\tTheme        string  \`json:"theme"\`
\tFontScale    float64 \`json:"fontScale"\`
\tReduceMotion bool    \`json:"reduceMotion"\`
}

type WebArtifactExclusion struct {
\tPath                string \`json:"path"\`
\tFingerprintSHA256   string \`json:"fingerprintSha256"\`
\tOwner               string \`json:"owner"\`
\tRationale           string \`json:"rationale"\`
\tReproductionCommand string \`json:"reproductionCommand"\`
}

type ConsumerProfile struct {
\tID                       string            \`json:"id"\`
\tPrimaryUserGoal          string            \`json:"primaryUserGoal"\`
\tDensity                  string            \`json:"density"\`
\tNoveltyTolerance         string            \`json:"noveltyTolerance"\`
\tNativeAffordancePriority string            \`json:"nativeAffordancePriority"\`
\tSeverityOverrides        map[string]string \`json:"severityOverrides"\`
\tThresholds               ProfileThresholds \`json:"thresholds"\`
\tRequiredEvidence         []string          \`json:"requiredEvidence"\`
\tRationale                string            \`json:"rationale,omitempty"\`
\tReviewer                 string            \`json:"reviewer,omitempty"\`
\tEvidenceOwner            string            \`json:"evidenceOwner,omitempty"\`
}

type ProfileThresholds struct {
\tMaxOversizedActions    int \`json:"maxOversizedActions"\`
\tMaxInconsistentActions int \`json:"maxInconsistentActions"\`
}

type RulePackRequirement struct {
\tID                string \`json:"id"\`
\tVersion           string \`json:"version"\`
\tFingerprintSHA256 string \`json:"fingerprintSha256"\`
}

type RuleOverride struct {
\tRuleID        string \`json:"ruleId"\`
\tPackID        string \`json:"packId"\`
\tPackVersion   string \`json:"packVersion"\`
\tStatus        string \`json:"status"\`
\tOwner         string \`json:"owner"\`
\tRationale     string \`json:"rationale"\`
\tReviewer      string \`json:"reviewer"\`
\tExpiresAt     string \`json:"expiresAt"\`
\tReviewTrigger string \`json:"reviewTrigger"\`
}

type GovernancePolicy struct {
\tReviewedAt              string            \`json:"reviewedAt"\`
\tReviewIntervalDays      int               \`json:"reviewIntervalDays"\`
\tReviewer                string            \`json:"reviewer"\`
\tReviewSubjects          []string          \`json:"reviewSubjects"\`
\tAdvisoryMode            string            \`json:"advisoryMode"\`
\tForbiddenFlags          []string          \`json:"forbiddenFlags"\`
\tRequireExitCode2        bool              \`json:"requireExitCode2"\`
\tRequireUnmodifiedReport bool              \`json:"requireUnmodifiedReport"\`
\tPassingReportsOnly      bool              \`json:"passingReportsOnly"\`
\tIgnores                 []IgnoreAllowance \`json:"ignores"\`
}

type IgnoreAllowance struct {
\tKind          string \`json:"kind"\`
\tRuleID        string \`json:"ruleId,omitempty"\`
\tEngine        string \`json:"engine"\`
\tPlatform      string \`json:"platform"\`
\tPath          string \`json:"path"\`
\tProperty      string \`json:"property,omitempty"\`
\tValue         string \`json:"value,omitempty"\`
\tOwner         string \`json:"owner"\`
\tRationale     string \`json:"rationale"\`
\tReviewer      string \`json:"reviewer"\`
\tExpiresAt     string \`json:"expiresAt"\`
\tReviewTrigger string \`json:"reviewTrigger"\`
}

type SourcePolicy struct {
\tRawProperties RawProperties \`json:"rawProperties"\`
\tExactExcludes []string      \`json:"exactExcludes"\`
}

type RawProperties struct {
\tColor  []string \`json:"color"\`
\tNumber []string \`json:"number"\`
\tMotion []string \`json:"motion"\`
}

type EvidencePolicy struct {
\tRequiredKinds        []string \`json:"requiredKinds"\`
\tDeferredKinds        []string \`json:"deferredKinds,omitempty"\`
\tLayoutDocumentSHA256 string   \`json:"layoutDocumentSha256,omitempty"\`
}

type Budgets struct {
\tError     int \`json:"error"\`
\tWarning   int \`json:"warning"\`
\tRaw       int \`json:"raw"\`
\tOverflow  int \`json:"overflow"\`
\tOverlap   int \`json:"overlap"\`
\tBlocking  int \`json:"blocking"\`
\tAdvisory  int \`json:"advisory"\`
\tException int \`json:"exception"\`
\tNotRun    int \`json:"notRun"\`
\tDeferred  int \`json:"deferred"\`
}

type Exception struct {
\tRuleID        string \`json:"ruleId"\`
\tEngine        string \`json:"engine"\`
\tPlatform      string \`json:"platform"\`
\tPath          string \`json:"path"\`
\tOwner         string \`json:"owner"\`
\tRationale     string \`json:"rationale"\`
\tReviewer      string \`json:"reviewer"\`
\tExpiresAt     string \`json:"expiresAt"\`
\tReviewTrigger string \`json:"reviewTrigger"\`
}
`;
}

function conformanceTypeScript(schema) {
  return `${header("consumer conformance", schema)}
export const consumerConformanceSchemaVersion = ${schema.version} as const;
export const consumerConformanceSchemaSha256 =
  "${schema.sha256}" as const;

export type ConsumerPlatform =
  | "web"
  | "react-native"
  | "ios"
  | "android"
  | "design-document";
export type ControlState =
  | "default"
  | "pressed"
  | "focused"
  | "disabled"
  | "loading"
  | "error"
  | "selected";

export interface ConsumerControl {
  id: string;
  actionId: string;
  role: string;
  component: string;
  label: string;
  shapeToken: string;
  icon?: string;
  feedback: string;
  states: ControlState[];
  contractStatus: "matched" | "mismatched";
  affordanceSource: "design-system" | "platform" | "consumer-exception" | "invented";
  motionPurpose: "none" | "state-transition" | "continuity" | "decorative";
  motionRecipeStatus: "none" | "approved" | "unapproved";
  reduceMotionFallback: boolean;
  rawDurationMs?: number;
  prominence: "standard" | "emphasized" | "oversized";
  nativePrimitive: boolean;
  exceptionId?: string;
}

export interface ConsumerConformanceEvidence {
  $schema?: string;
  schemaVersion: typeof consumerConformanceSchemaVersion;
  profileId: string;
  platform: ConsumerPlatform;
  surfaceId: string;
  controls: ConsumerControl[];
}
`;
}

function conformanceGo(schema) {
  return `${header("consumer conformance", schema)}
package conformance

const SchemaVersion = ${schema.version}
const SchemaSHA256 = "${schema.sha256}"

type Evidence struct {
\tSchema        string    \`json:"$schema,omitempty"\`
\tSchemaVersion int       \`json:"schemaVersion"\`
\tProfileID     string    \`json:"profileId"\`
\tPlatform      string    \`json:"platform"\`
\tSurfaceID     string    \`json:"surfaceId"\`
\tControls      []Control \`json:"controls"\`
}

type Control struct {
\tID                   string   \`json:"id"\`
\tActionID             string   \`json:"actionId"\`
\tRole                 string   \`json:"role"\`
\tComponent            string   \`json:"component"\`
\tLabel                string   \`json:"label"\`
\tShapeToken           string   \`json:"shapeToken"\`
\tIcon                 string   \`json:"icon,omitempty"\`
\tFeedback             string   \`json:"feedback"\`
\tStates               []string \`json:"states"\`
\tContractStatus       string   \`json:"contractStatus"\`
\tAffordanceSource     string   \`json:"affordanceSource"\`
\tMotionPurpose        string   \`json:"motionPurpose"\`
\tMotionRecipeStatus   string   \`json:"motionRecipeStatus"\`
\tReduceMotionFallback bool     \`json:"reduceMotionFallback"\`
\tRawDurationMS        float64  \`json:"rawDurationMs,omitempty"\`
\tProminence           string   \`json:"prominence"\`
\tNativePrimitive      bool     \`json:"nativePrimitive"\`
\tExceptionID          string   \`json:"exceptionId,omitempty"\`
}
`;
}
