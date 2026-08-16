export type ThemeName = string;
export type TokenLayer = "primitive" | "semantic" | "component" | "asset";

export interface ResolvedTypographyRole {
  semanticRole: string;
  visualToken: string;
  logicalWeight: string;
  fontFamily: string;
  fontWeight: number;
  fontSize: number;
  lineHeight: number;
  maxFontSizeMultiplier: number;
}

export interface ComponentDefinition {
  anatomy: readonly string[];
  slots: Readonly<Record<string, unknown>>;
  variants: Readonly<Record<string, Readonly<Record<string, unknown>>>>;
  sizes: Readonly<Record<string, Readonly<Record<string, unknown>>>>;
  states: readonly string[];
  semantics: Readonly<Record<string, unknown>>;
}

export interface DesignSystemRuntime {
  readonly id: string;
  readonly version: string;
  readonly theme: ThemeName;
  readonly availableThemes: readonly ThemeName[];
  readonly fingerprintSha256: string;
  color(layer: TokenLayer, name: string): string;
  number(
    collection: "spacing" | "radius" | "size" | "layer",
    layer: "primitive" | "semantic" | "component",
    name: string,
  ): number;
  typography(role: string): ResolvedTypographyRole;
  component(name: string): ComponentDefinition;
  withTheme(theme: ThemeName): DesignSystemRuntime;
}

export interface CreateDesignSystemOptions {
  theme?: ThemeName;
}

export function createDesignSystem(
  bundle: unknown,
  options: CreateDesignSystemOptions = {},
): DesignSystemRuntime {
  const root = requireRecord(bundle, "compiled bundle");
  if (root.version !== 1)
    throw new Error("unsupported compiled bundle version");
  const definition = requireRecord(root.definition, "bundle.definition");
  const themes = requireRecord(root.themes, "bundle.themes");
  const availableThemes = Object.keys(themes).sort();
  const defaultTheme = requireString(
    definition.defaultTheme,
    "definition.defaultTheme",
  );
  const theme = options.theme ?? defaultTheme;
  if (!(theme in themes)) throw new Error(`unknown theme ${theme}`);

  function runtimeFor(selectedTheme: string): DesignSystemRuntime {
    const themeBundle = requireRecord(
      themes[selectedTheme],
      `themes.${selectedTheme}`,
    );
    const foundations = requireRecord(
      themeBundle.foundations,
      "theme.foundations",
    );
    const components = requireRecord(
      themeBundle.components,
      "theme.components",
    );

    function typography(role: string): ResolvedTypographyRole {
      const typographyGroup = requireRecord(
        foundations.typography,
        "foundations.typography",
      );
      const roles = requireRecord(typographyGroup.roles, "typography.roles");
      return requireTypographyRole(roles[role], role);
    }

    return Object.freeze({
      id: requireString(definition.id, "definition.id"),
      version: requireString(definition.version, "definition.version"),
      theme: selectedTheme,
      availableThemes,
      fingerprintSha256: requireString(
        root.fingerprintSha256,
        "fingerprintSha256",
      ),
      color(layer: TokenLayer, name: string): string {
        const color = requireRecord(foundations.color, "foundations.color");
        return requireString(
          requireRecord(color[layer], `color.${layer}`)[name],
          `color.${layer}.${name}`,
        );
      },
      number(
        collection: "spacing" | "radius" | "size" | "layer",
        layer: "primitive" | "semantic" | "component",
        name: string,
      ): number {
        const group = requireRecord(
          foundations[collection],
          `foundations.${collection}`,
        );
        const value = requireRecord(group[layer], `${collection}.${layer}`)[
          name
        ];
        if (typeof value !== "number")
          throw new Error(`${collection}.${layer}.${name} is not numeric`);
        return value;
      },
      typography,
      component(name: string): ComponentDefinition {
        return requireComponent(components[name], name);
      },
      withTheme(nextTheme: string): DesignSystemRuntime {
        if (!(nextTheme in themes))
          throw new Error(`unknown theme ${nextTheme}`);
        return runtimeFor(nextTheme);
      },
    });
  }

  return runtimeFor(theme);
}

export function resolveScaledControlMinHeight({
  baseHeight,
  fontScale,
  lineHeight,
  maximumFontScale,
  verticalPadding,
}: {
  baseHeight: number;
  fontScale: number;
  lineHeight: number;
  maximumFontScale: number;
  verticalPadding: number;
}): number {
  const normalizedScale = Number.isFinite(fontScale)
    ? Math.min(maximumFontScale, Math.max(1, fontScale))
    : 1;
  return Math.ceil(
    Math.max(baseHeight, lineHeight * normalizedScale + verticalPadding * 2),
  );
}

export function resolveScaledContentMinHeight({
  fontScale,
  gap,
  lineHeights,
  maximumFontScale,
  verticalPadding,
}: {
  fontScale: number;
  gap: number;
  lineHeights: readonly number[];
  maximumFontScale: number;
  verticalPadding: number;
}): number {
  const normalizedScale = Number.isFinite(fontScale)
    ? Math.min(maximumFontScale, Math.max(1, fontScale))
    : 1;
  const lineHeightTotal = lineHeights.reduce(
    (total, lineHeight) => total + lineHeight * normalizedScale,
    0,
  );
  const gapTotal = Math.max(0, lineHeights.length - 1) * gap;
  return Math.ceil(lineHeightTotal + gapTotal + verticalPadding * 2);
}

function requireRecord(value: unknown, label: string): Record<string, unknown> {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value as Record<string, unknown>;
}

function requireString(value: unknown, label: string): string {
  if (typeof value !== "string" || value.length === 0)
    throw new Error(`${label} must be a string`);
  return value;
}

function requireNumber(value: unknown, label: string): number {
  if (typeof value !== "number" || !Number.isFinite(value))
    throw new Error(`${label} must be numeric`);
  return value;
}

function requireTypographyRole(
  value: unknown,
  role: string,
): ResolvedTypographyRole {
  const record = requireRecord(value, `typography role ${role}`);
  return {
    semanticRole: requireString(record.semanticRole, `${role}.semanticRole`),
    visualToken: requireString(record.visualToken, `${role}.visualToken`),
    logicalWeight: requireString(record.logicalWeight, `${role}.logicalWeight`),
    fontFamily: requireString(record.fontFamily, `${role}.fontFamily`),
    fontWeight: requireNumber(record.fontWeight, `${role}.fontWeight`),
    fontSize: requireNumber(record.fontSize, `${role}.fontSize`),
    lineHeight: requireNumber(record.lineHeight, `${role}.lineHeight`),
    maxFontSizeMultiplier: requireNumber(
      record.maxFontSizeMultiplier,
      `${role}.maxFontSizeMultiplier`,
    ),
  };
}

function requireComponent(value: unknown, name: string): ComponentDefinition {
  const record = requireRecord(value, `component ${name}`);
  if (
    !Array.isArray(record.anatomy) ||
    !record.anatomy.every((item) => typeof item === "string")
  ) {
    throw new Error(`component ${name}.anatomy must be a string array`);
  }
  if (
    !Array.isArray(record.states) ||
    !record.states.every((item) => typeof item === "string")
  ) {
    throw new Error(`component ${name}.states must be a string array`);
  }
  return {
    anatomy: record.anatomy,
    slots: requireRecord(record.slots, `${name}.slots`),
    variants: requireRecord(
      record.variants,
      `${name}.variants`,
    ) as ComponentDefinition["variants"],
    sizes: requireRecord(
      record.sizes,
      `${name}.sizes`,
    ) as ComponentDefinition["sizes"],
    states: record.states,
    semantics: requireRecord(record.semantics, `${name}.semantics`),
  };
}
