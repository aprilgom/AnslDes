// Code generated from https://ansldes.dev/schema/design-system-definition.v2.json; DO NOT EDIT.
// definition schema SHA-256: 1e1d8c99aa9dab3825860c297a2c6a9086c3aba78162552a20d4ef75fbae504e

export const definitionSchemaVersion = 2 as const;
export const definitionSchemaSha256 =
  "1e1d8c99aa9dab3825860c297a2c6a9086c3aba78162552a20d4ef75fbae504e" as const;

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
