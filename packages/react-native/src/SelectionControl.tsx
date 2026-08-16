import { useState } from "react";
import { Pressable, StyleSheet, View } from "react-native";

import { useDesignSystem } from "./context";
import { DesignText } from "./DesignText";
import { number, recipe, string } from "./recipe";

export interface SelectionControlProps {
  description?: string;
  disabled?: boolean;
  label: string;
  mode: "checkbox" | "radio";
  onPress: () => void;
  selected: boolean;
  size?: string;
  testID?: string;
  variant?: string;
}

export function SelectionControl({
  description,
  disabled = false,
  label,
  mode,
  onPress,
  selected,
  size = "medium",
  testID,
  variant = "default",
}: SelectionControlProps) {
  const runtime = useDesignSystem();
  const component = runtime.component("selection");
  const sizeRecipe = recipe(component.sizes[size], `selection size ${size}`);
  const variantRecipe = recipe(
    component.variants[variant],
    `selection variant ${variant}`,
  );
  const slots = component.slots;
  const [focused, setFocused] = useState(false);

  return (
    <Pressable
      accessibilityLabel={description ? `${label}. ${description}` : label}
      accessibilityRole={mode}
      accessibilityState={{ checked: selected, disabled }}
      disabled={disabled}
      onBlur={() => setFocused(false)}
      onFocus={() => setFocused(true)}
      onPress={onPress}
      testID={testID}
      style={({ pressed }) => [
        styles.root,
        {
          backgroundColor: string(
            selected
              ? variantRecipe.selectedContainer
              : variantRecipe.container,
            "selection container",
          ),
          borderColor: string(
            focused
              ? slots.focusBorder
              : selected
                ? variantRecipe.selectedBorder
                : variantRecipe.border,
            "selection border",
          ),
          borderRadius: number(variantRecipe.radius, "selection radius"),
          borderWidth: number(
            slots.focusBorderWidth,
            "selection focusBorderWidth",
          ),
          gap: number(sizeRecipe.contentGap, "selection contentGap"),
          minHeight: number(sizeRecipe.minHeight, "selection minHeight"),
          opacity: disabled
            ? number(slots.disabledOpacity, "selection disabledOpacity")
            : pressed
              ? number(slots.pressedOpacity, "selection pressedOpacity")
              : 1,
          paddingHorizontal: number(
            sizeRecipe.horizontalPadding,
            "selection horizontalPadding",
          ),
          paddingVertical: number(
            sizeRecipe.verticalPadding,
            "selection verticalPadding",
          ),
        },
      ]}
    >
      <View
        accessibilityElementsHidden
        importantForAccessibility="no-hide-descendants"
        style={[
          styles.indicator,
          {
            backgroundColor: string(
              selected ? slots.selectedIndicator : slots.idleIndicator,
              "selection indicator",
            ),
            borderRadius:
              mode === "radio"
                ? number(slots.indicatorSize, "selection indicatorSize") / 2
                : number(slots.indicatorRadius, "selection indicatorRadius"),
            height: number(slots.indicatorSize, "selection indicatorSize"),
            width: number(slots.indicatorSize, "selection indicatorSize"),
          },
        ]}
      />
      <View style={styles.copy}>
        <DesignText
          typographyRole={string(sizeRecipe.labelRole, "selection labelRole")}
          style={{ color: string(variantRecipe.label, "selection label") }}
        >
          {label}
        </DesignText>
        {description ? (
          <DesignText
            typographyRole={string(
              sizeRecipe.descriptionRole,
              "selection descriptionRole",
            )}
            style={{
              color: string(variantRecipe.description, "selection description"),
            }}
          >
            {description}
          </DesignText>
        ) : null}
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  copy: { flexShrink: 1 },
  indicator: { flexShrink: 0 },
  root: { alignItems: "center", flexDirection: "row" },
});
