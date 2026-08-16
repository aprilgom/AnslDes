import {
  type ComponentDefinition,
  type DesignSystemRuntime,
  type ResolvedTypographyRole,
  resolveScaledControlMinHeight,
} from "@ansldes/core";

export interface ButtonPresentation {
  accessibilityLabel: string;
  accessibilityState: { busy: boolean; disabled: boolean };
  container: {
    backgroundColor: string;
    borderColor: string;
    borderRadius: number;
    borderWidth: number;
    gap: number;
    minHeight: number;
    opacity: number;
    paddingHorizontal: number;
    paddingVertical: number;
  };
  labelColor: string;
  typography: ResolvedTypographyRole;
  unavailable: boolean;
}

export function resolveButtonPresentation(
  runtime: DesignSystemRuntime,
  options: {
    accessibilityContext?: string;
    disabled: boolean;
    focused: boolean;
    fontScale: number;
    label: string;
    loading: boolean;
    pressed: boolean;
    size: string;
    variant: string;
  },
): ButtonPresentation {
  const button = runtime.component("button");
  const variant = recipe(
    button.variants[options.variant],
    `button variant ${options.variant}`,
  );
  const size = recipe(
    button.sizes[options.size],
    `button size ${options.size}`,
  );
  const slots = button.slots;
  const typography = typographyRecipe(size.typographyRole);
  const unavailable = options.disabled || options.loading;
  const paddingVertical = number(
    size.verticalPadding,
    "button verticalPadding",
  );
  const backgroundColor = unavailable
    ? string(slots.disabledContainer, "button disabledContainer")
    : options.pressed
      ? string(variant.pressedContainer, "button pressedContainer")
      : string(variant.container, "button container");

  return {
    accessibilityLabel: options.accessibilityContext
      ? `${options.accessibilityContext}. ${options.label}`
      : options.label,
    accessibilityState: { busy: options.loading, disabled: unavailable },
    container: {
      backgroundColor,
      borderColor: string(
        options.focused ? slots.focusBorder : slots.idleBorder,
        "button border",
      ),
      borderRadius: number(variant.radius, "button radius"),
      borderWidth: number(slots.focusBorderWidth, "button focusBorderWidth"),
      gap: number(slots.contentGap, "button contentGap"),
      minHeight: resolveScaledControlMinHeight({
        baseHeight: number(size.minHeight, "button minHeight"),
        fontScale: options.fontScale,
        lineHeight: typography.lineHeight,
        maximumFontScale: typography.maxFontSizeMultiplier,
        verticalPadding: paddingVertical,
      }),
      opacity: options.pressed
        ? number(variant.pressedOpacity, "button pressedOpacity")
        : 1,
      paddingHorizontal: number(
        size.horizontalPadding,
        "button horizontalPadding",
      ),
      paddingVertical,
    },
    labelColor: string(
      unavailable ? slots.disabledLabel : variant.label,
      "button label color",
    ),
    typography,
    unavailable,
  };
}

function recipe(
  value: unknown,
  label: string,
): Readonly<Record<string, unknown>> {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new Error(`${label} is missing`);
  }
  return value as Readonly<Record<string, unknown>>;
}

function typographyRecipe(value: unknown): ResolvedTypographyRole {
  const result = recipe(value, "button typographyRole");
  return {
    semanticRole: string(result.semanticRole, "typography semanticRole"),
    visualToken: string(result.visualToken, "typography visualToken"),
    logicalWeight: string(result.logicalWeight, "typography logicalWeight"),
    fontFamily: string(result.fontFamily, "typography fontFamily"),
    fontWeight: number(result.fontWeight, "typography fontWeight"),
    fontSize: number(result.fontSize, "typography fontSize"),
    lineHeight: number(result.lineHeight, "typography lineHeight"),
    maxFontSizeMultiplier: number(
      result.maxFontSizeMultiplier,
      "typography maxFontSizeMultiplier",
    ),
  };
}

function string(value: unknown, label: string): string {
  if (typeof value !== "string") throw new Error(`${label} must be a string`);
  return value;
}

function number(value: unknown, label: string): number {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw new Error(`${label} must be numeric`);
  }
  return value;
}

export function assertButtonContract(button: ComponentDefinition): void {
  for (const slot of [
    "contentGap",
    "disabledContainer",
    "disabledLabel",
    "focusBorder",
    "focusBorderWidth",
    "idleBorder",
  ]) {
    if (!(slot in button.slots))
      throw new Error(`button slot ${slot} is required`);
  }
}
