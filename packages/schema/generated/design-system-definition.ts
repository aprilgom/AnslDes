// Code generated from https://ansldes.dev/schema/design-system-definition.v1.json; DO NOT EDIT.
// definition schema SHA-256: 944c5b2ee33ae7e502509000bcd83bc7daa64009d71c8c99d52feebb43622708

export const definitionSchemaVersion = 1 as const;
export const definitionSchemaSha256 =
  "944c5b2ee33ae7e502509000bcd83bc7daa64009d71c8c99d52feebb43622708" as const;

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
