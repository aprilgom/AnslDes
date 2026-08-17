// Code generated from https://ansldes.dev/schema/design-system-definition.v2.json; DO NOT EDIT.
// definition schema SHA-256: 6d72afbbcbd876ec8bfbaf5c9451c27067ea411ccb4865da302a85b7ce9cc041

export const definitionSchemaVersion = 2 as const;
export const definitionSchemaSha256 =
  "6d72afbbcbd876ec8bfbaf5c9451c27067ea411ccb4865da302a85b7ce9cc041" as const;

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
