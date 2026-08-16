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
const definitionSchema = await readSchema(definitionSchemaPath);
const policySchema = await readSchema(policySchemaPath);

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
    path.join(root, "deslint/internal/contract/generated_definition.go"),
    definitionGo(definitionSchema),
  ],
  [
    path.join(root, "deslint/internal/policy/generated_policy.go"),
    policyGo(policySchema),
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

async function readSchema(schemaPath) {
  const source = await readFile(schemaPath, "utf8");
  const schema = JSON.parse(source);
  const version = schema.properties?.schemaVersion?.const;
  if (version !== 1 || typeof schema.$id !== "string") {
    throw new Error(`${schemaPath} must declare schemaVersion const 1 and $id`);
  }
  return {
    id: schema.$id,
    sha256: createHash("sha256").update(source).digest("hex"),
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
  foundations: Record<string, unknown>;
  components: Record<string, ComponentDefinition>;
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
\tFoundations   map[string]json.RawMessage     \`json:"foundations"\`
\tComponents    map[string]ComponentDefinition \`json:"components"\`
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

function policyGo(schema) {
  return `${header("policy", schema)}
package policy

const SchemaVersion = ${schema.version}
const SchemaSHA256 = "${schema.sha256}"

type Policy struct {
\tSchema        string            \`json:"$schema,omitempty"\`
\tSchemaVersion int               \`json:"schemaVersion"\`
\tDefinitionID  string            \`json:"definitionId"\`
\tSeverities    map[string]string \`json:"severities"\`
\tSource        SourcePolicy      \`json:"source"\`
\tEvidence      EvidencePolicy    \`json:"evidence"\`
\tBudgets       Budgets           \`json:"budgets"\`
\tExceptions    []Exception       \`json:"exceptions"\`
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
\tLayoutDocumentSHA256 string   \`json:"layoutDocumentSha256,omitempty"\`
}

type Budgets struct {
\tError    int \`json:"error"\`
\tWarning  int \`json:"warning"\`
\tRaw      int \`json:"raw"\`
\tOverflow int \`json:"overflow"\`
\tOverlap  int \`json:"overlap"\`
}

type Exception struct {
\tRuleID    string \`json:"ruleId"\`
\tPath      string \`json:"path"\`
\tOwner     string \`json:"owner"\`
\tRationale string \`json:"rationale"\`
\tExpiresAt string \`json:"expiresAt"\`
}
`;
}
