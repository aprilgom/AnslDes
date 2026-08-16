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

export interface ResolvedMotionTransition {
  duration: number;
  easing: readonly [number, number, number, number];
  fallback: string | null;
}

export interface IconGeometryPart {
  readonly kind: string;
  readonly [name: string]: unknown;
}

export interface IconDefinition {
  viewBox: readonly [number, number, number, number];
  geometry: readonly IconGeometryPart[];
  defaultSize: string;
  allowedSizes: readonly string[];
  opticalAlignment: string;
  pathFingerprint?: string;
}

export interface TextSelection {
  start: number;
  end: number;
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
  elevation(name: string): Readonly<Record<string, unknown>>;
  motionTransition(
    name: string,
    reduceMotion: boolean,
  ): ResolvedMotionTransition;
  icon(name: string): IconDefinition;
  iconSize(name: string): number;
  iconStroke(name: string): number;
  iconUsage(name: string): Readonly<Record<string, unknown>>;
  iconAction(name: string): Readonly<Record<string, unknown>>;
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
      elevation(name: string): Readonly<Record<string, unknown>> {
        const elevation = requireRecord(
          foundations.elevation,
          "foundations.elevation",
        );
        return requireRecord(elevation[name], `elevation.${name}`);
      },
      motionTransition(
        name: string,
        reduceMotion: boolean,
      ): ResolvedMotionTransition {
        const motion = requireRecord(foundations.motion, "foundations.motion");
        const transitions = requireRecord(
          motion.transitions,
          "motion.transitions",
        );
        return requireMotionTransition(transitions[name], name, reduceMotion);
      },
      icon(name: string): IconDefinition {
        const icon = requireIconGroup(foundations);
        return requireIconDefinition(
          requireRecord(icon.icons, "icon.icons")[name],
          name,
        );
      },
      iconSize(name: string): number {
        const icon = requireIconGroup(foundations);
        return requireNumber(
          requireRecord(icon.sizes, "icon.sizes")[name],
          `icon.sizes.${name}`,
        );
      },
      iconStroke(name: string): number {
        const icon = requireIconGroup(foundations);
        return requireNumber(
          requireRecord(icon.strokes, "icon.strokes")[name],
          `icon.strokes.${name}`,
        );
      },
      iconUsage(name: string): Readonly<Record<string, unknown>> {
        const icon = requireIconGroup(foundations);
        return requireRecord(
          requireRecord(icon.usages, "icon.usages")[name],
          `icon.usages.${name}`,
        );
      },
      iconAction(name: string): Readonly<Record<string, unknown>> {
        const icon = requireIconGroup(foundations);
        return requireRecord(
          requireRecord(icon.actions, "icon.actions")[name],
          `icon.actions.${name}`,
        );
      },
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

export function resolveToggleThumbTravel({
  inset,
  thumbSize,
  trackWidth,
}: {
  inset: number;
  thumbSize: number;
  trackWidth: number;
}): number {
  return Math.max(0, trackWidth - thumbSize - Math.max(0, inset) * 2);
}

export function mapSelectionThroughFormat(
  editedText: string,
  formattedText: string,
  selection: TextSelection,
  isSemanticCharacter: (character: string) => boolean,
): TextSelection {
  const editedSelection = clampTextSelection(editedText, selection);
  const startCount = semanticCharacterCount(
    editedText,
    editedSelection.start,
    isSemanticCharacter,
  );
  const endCount = semanticCharacterCount(
    editedText,
    editedSelection.end,
    isSemanticCharacter,
  );
  return clampTextSelection(formattedText, {
    start: boundaryAfterSemanticCharacters(
      formattedText,
      startCount,
      isSemanticCharacter,
    ),
    end: boundaryAfterSemanticCharacters(
      formattedText,
      endCount,
      isSemanticCharacter,
    ),
  });
}

export function mapDigitSelectionThroughFormat(
  editedText: string,
  formattedText: string,
  selection: TextSelection,
): TextSelection {
  return mapSelectionThroughFormat(
    editedText,
    formattedText,
    selection,
    (character) => /[0-9]/u.test(character),
  );
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

function requireNumberTuple(
  value: unknown,
  label: string,
): readonly [number, number, number, number] {
  if (
    !Array.isArray(value) ||
    value.length !== 4 ||
    !value.every((item) => typeof item === "number" && Number.isFinite(item))
  ) {
    throw new Error(`${label} must contain four numbers`);
  }
  return [value[0], value[1], value[2], value[3]];
}

function requireIconGroup(foundations: Record<string, unknown>) {
  return requireRecord(foundations.icon, "foundations.icon");
}

function requireIconDefinition(value: unknown, name: string): IconDefinition {
  const record = requireRecord(value, `icon ${name}`);
  if (
    !Array.isArray(record.geometry) ||
    !record.geometry.every(
      (part) =>
        typeof part === "object" &&
        part !== null &&
        !Array.isArray(part) &&
        typeof (part as Record<string, unknown>).kind === "string",
    )
  ) {
    throw new Error(`icon ${name}.geometry must contain named parts`);
  }
  if (
    !Array.isArray(record.allowedSizes) ||
    !record.allowedSizes.every((item) => typeof item === "string")
  ) {
    throw new Error(`icon ${name}.allowedSizes must be a string array`);
  }
  return {
    viewBox: requireNumberTuple(record.viewBox, `${name}.viewBox`),
    geometry: record.geometry as IconGeometryPart[],
    defaultSize: requireString(record.defaultSize, `${name}.defaultSize`),
    allowedSizes: record.allowedSizes,
    opticalAlignment: requireString(
      record.opticalAlignment,
      `${name}.opticalAlignment`,
    ),
    ...(typeof record.pathFingerprint === "string"
      ? { pathFingerprint: record.pathFingerprint }
      : {}),
  };
}

function requireMotionTransition(
  value: unknown,
  name: string,
  reduceMotion: boolean,
): ResolvedMotionTransition {
  const recipe = requireRecord(value, `motion transition ${name}`);
  const reduced = requireRecord(
    recipe.reducedMotion,
    `${name}.reducedMotion`,
  );
  return {
    duration: requireNumber(
      reduceMotion ? reduced.duration : recipe.duration,
      `${name}.duration`,
    ),
    easing: requireNumberTuple(recipe.easing, `${name}.easing`),
    fallback: reduceMotion
      ? requireString(reduced.fallback, `${name}.reducedMotion.fallback`)
      : null,
  };
}

function clampTextSelection(
  text: string,
  selection: TextSelection,
): TextSelection {
  const start = Math.max(0, Math.min(selection.start, text.length));
  return { start, end: Math.max(start, Math.min(selection.end, text.length)) };
}

function semanticCharacterCount(
  text: string,
  boundary: number,
  isSemanticCharacter: (character: string) => boolean,
): number {
  return [...text.slice(0, boundary)].filter(isSemanticCharacter).length;
}

function boundaryAfterSemanticCharacters(
  text: string,
  count: number,
  isSemanticCharacter: (character: string) => boolean,
): number {
  if (count <= 0) return 0;
  let matched = 0;
  for (let index = 0; index < text.length; index += 1) {
    if (isSemanticCharacter(text[index] ?? "")) matched += 1;
    if (matched === count) return index + 1;
  }
  return text.length;
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
