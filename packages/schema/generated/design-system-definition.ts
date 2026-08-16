// Code generated from https://ansldes.dev/schema/design-system-definition.v1.json; DO NOT EDIT.
// definition schema SHA-256: 307ea7d484ff09fa3235c5bb56531af5765731e6f8bdee09cf5b09c7a17f9024

export const definitionSchemaVersion = 1 as const;
export const definitionSchemaSha256 =
  "307ea7d484ff09fa3235c5bb56531af5765731e6f8bdee09cf5b09c7a17f9024" as const;

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
