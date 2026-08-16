import type { DesignSystemRuntime } from "@ansldes/core";

import type { FontWeightMode } from "./context";

export type NativeFontWeight =
  | "normal"
  | "100"
  | "200"
  | "300"
  | "400"
  | "500"
  | "600"
  | "700"
  | "800"
  | "900";

export interface NativeTextRecipe {
  fontFamily: string;
  fontSize: number;
  fontWeight: NativeFontWeight;
  lineHeight: number;
  maxFontSizeMultiplier: number;
}

export function resolveNativeTextRecipe(
  runtime: DesignSystemRuntime,
  role: string,
  fontWeightMode: FontWeightMode,
): NativeTextRecipe {
  const resolved = runtime.typography(role);
  return {
    fontFamily: resolved.fontFamily,
    fontSize: resolved.fontSize,
    fontWeight:
      fontWeightMode === "normal"
        ? "normal"
        : requireNativeFontWeight(resolved.fontWeight),
    lineHeight: resolved.lineHeight,
    maxFontSizeMultiplier: resolved.maxFontSizeMultiplier,
  };
}

function requireNativeFontWeight(value: number): NativeFontWeight {
  const weight = String(value);
  if (!/^[1-9]00$/u.test(weight)) {
    throw new Error(`fontWeight ${weight} is not supported by React Native`);
  }
  return weight as Exclude<NativeFontWeight, "normal">;
}
