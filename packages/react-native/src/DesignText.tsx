import { forwardRef } from "react";
import { Text, type TextProps } from "react-native";

import { useDesignSystem, useNativeFontWeightMode } from "./context";
import { resolveNativeTextRecipe } from "./textRecipe";

export interface DesignTextProps extends TextProps {
  typographyRole: string;
}

export const DesignText = forwardRef<Text, DesignTextProps>(function DesignText(
  {
    allowFontScaling = true,
    maxFontSizeMultiplier,
    style,
    typographyRole,
    ...props
  },
  ref,
) {
  const runtime = useDesignSystem();
  const fontWeightMode = useNativeFontWeightMode();
  const recipe = resolveNativeTextRecipe(
    runtime,
    typographyRole,
    fontWeightMode,
  );
  return (
    <Text
      ref={ref}
      allowFontScaling={allowFontScaling}
      maxFontSizeMultiplier={
        maxFontSizeMultiplier ?? recipe.maxFontSizeMultiplier
      }
      style={[
        style,
        {
          fontFamily: recipe.fontFamily,
          fontSize: recipe.fontSize,
          fontWeight: recipe.fontWeight,
          lineHeight: recipe.lineHeight,
        },
      ]}
      {...props}
    />
  );
});
