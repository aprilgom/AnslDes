import { forwardRef, useState } from "react";
import { StyleSheet, TextInput, type TextInputProps, View } from "react-native";

import { useDesignSystem, useNativeFontWeightMode } from "./context";
import { DesignText } from "./DesignText";
import { number, recipe, string } from "./recipe";
import { resolveNativeTextRecipe } from "./textRecipe";

export interface TextFieldProps
  extends Omit<
    TextInputProps,
    | "accessibilityLabel"
    | "allowFontScaling"
    | "editable"
    | "maxFontSizeMultiplier"
    | "placeholderTextColor"
    | "style"
  > {
  disabled?: boolean;
  error?: string;
  helperText?: string;
  label: string;
  size?: string;
  variant?: string;
}

export const TextField = forwardRef<TextInput, TextFieldProps>(
  function TextField(
    {
      disabled = false,
      error,
      helperText,
      label,
      onBlur,
      onFocus,
      placeholder,
      size = "medium",
      variant = "default",
      ...props
    },
    ref,
  ) {
    const runtime = useDesignSystem();
    const fontWeightMode = useNativeFontWeightMode();
    const component = runtime.component("textField");
    const sizeRecipe = recipe(component.sizes[size], `textField size ${size}`);
    const variantRecipe = recipe(
      component.variants[variant],
      `textField variant ${variant}`,
    );
    const slots = component.slots;
    const [focused, setFocused] = useState(false);
    const borderColor = error
      ? string(slots.errorBorder, "textField errorBorder")
      : focused
        ? string(slots.focusBorder, "textField focusBorder")
        : string(slots.idleBorder, "textField idleBorder");

    return (
      <View
        style={{ gap: number(sizeRecipe.contentGap, "textField contentGap") }}
      >
        <DesignText
          typographyRole={string(sizeRecipe.labelRole, "textField labelRole")}
          style={{ color: string(variantRecipe.label, "textField label") }}
        >
          {label}
        </DesignText>
        <TextInput
          ref={ref}
          accessibilityLabel={label}
          allowFontScaling
          editable={!disabled}
          maxFontSizeMultiplier={
            resolveNativeTextRecipe(
              runtime,
              string(sizeRecipe.inputRole, "textField inputRole"),
              fontWeightMode,
            ).maxFontSizeMultiplier
          }
          onBlur={(event) => {
            setFocused(false);
            onBlur?.(event);
          }}
          onFocus={(event) => {
            setFocused(true);
            onFocus?.(event);
          }}
          placeholder={placeholder}
          placeholderTextColor={string(
            slots.placeholder,
            "textField placeholder",
          )}
          style={[
            styles.input,
            resolveNativeTextRecipe(
              runtime,
              string(sizeRecipe.inputRole, "textField inputRole"),
              fontWeightMode,
            ),
            {
              backgroundColor: string(
                disabled ? slots.disabledContainer : variantRecipe.container,
                "textField container",
              ),
              borderColor,
              borderRadius: number(variantRecipe.radius, "textField radius"),
              borderWidth: number(
                focused ? slots.focusBorderWidth : slots.idleBorderWidth,
                "textField borderWidth",
              ),
              color: string(
                disabled ? slots.disabledText : variantRecipe.text,
                "textField text",
              ),
              minHeight: number(sizeRecipe.minHeight, "textField minHeight"),
              paddingHorizontal: number(
                sizeRecipe.horizontalPadding,
                "textField horizontalPadding",
              ),
              paddingVertical: number(
                sizeRecipe.verticalPadding,
                "textField verticalPadding",
              ),
            },
          ]}
          {...props}
        />
        {error ? (
          <DesignText
            accessibilityLiveRegion="polite"
            typographyRole={string(
              sizeRecipe.messageRole,
              "textField messageRole",
            )}
            style={{ color: string(slots.errorText, "textField errorText") }}
          >
            {error}
          </DesignText>
        ) : helperText ? (
          <DesignText
            typographyRole={string(
              sizeRecipe.messageRole,
              "textField messageRole",
            )}
            style={{ color: string(variantRecipe.helper, "textField helper") }}
          >
            {helperText}
          </DesignText>
        ) : null}
      </View>
    );
  },
);

const styles = StyleSheet.create({
  input: {
    minWidth: 0,
  },
});
