import { type ReactNode, useState } from "react";
import { Pressable, StyleSheet, View } from "react-native";

import { useDesignSystem } from "./context";
import { DesignText } from "./DesignText";
import { number, recipe, string } from "./recipe";

export interface ListItemProps {
  accessibilityHint?: string;
  accessory?: ReactNode;
  disabled?: boolean;
  label: string;
  onPress?: () => void;
  size?: string;
  supportingText?: string;
  testID?: string;
  variant?: string;
}

export function ListItem(props: ListItemProps) {
  return props.onPress ? (
    <InteractiveListItem {...props} onPress={props.onPress} />
  ) : (
    <ListContent {...props} />
  );
}

function InteractiveListItem(props: ListItemProps & { onPress: () => void }) {
  const [focused, setFocused] = useState(false);
  const resolved = useResolvedListItem(props, focused);
  return (
    <Pressable
      accessibilityHint={props.accessibilityHint}
      accessibilityLabel={resolved.accessibilityLabel}
      accessibilityRole="button"
      accessibilityState={{ disabled: props.disabled ?? false }}
      disabled={props.disabled}
      onBlur={() => setFocused(false)}
      onFocus={() => setFocused(true)}
      onPress={props.onPress}
      testID={props.testID}
      style={({ pressed }) => [
        resolved.container,
        { opacity: pressed ? resolved.pressedOpacity : 1 },
      ]}
    >
      <ListCopy {...props} resolved={resolved} />
    </Pressable>
  );
}

function ListContent(props: ListItemProps) {
  const resolved = useResolvedListItem(props, false);
  return (
    <View style={resolved.container} testID={props.testID}>
      <ListCopy {...props} resolved={resolved} />
    </View>
  );
}

function ListCopy({
  accessory,
  label,
  resolved,
  supportingText,
}: ListItemProps & { resolved: ReturnType<typeof useResolvedListItem> }) {
  return (
    <>
      <View style={styles.copy}>
        <DesignText
          typographyRole={resolved.labelRole}
          style={{ color: resolved.labelColor }}
        >
          {label}
        </DesignText>
        {supportingText ? (
          <DesignText
            typographyRole={resolved.supportingRole}
            style={{ color: resolved.supportingColor }}
          >
            {supportingText}
          </DesignText>
        ) : null}
      </View>
      {accessory ? (
        <View
          accessibilityElementsHidden
          importantForAccessibility="no-hide-descendants"
        >
          {accessory}
        </View>
      ) : null}
    </>
  );
}

function useResolvedListItem(props: ListItemProps, focused: boolean) {
  const runtime = useDesignSystem();
  const component = runtime.component("listItem");
  const sizeRecipe = recipe(
    component.sizes[props.size ?? "twoLine"],
    "listItem size",
  );
  const variantRecipe = recipe(
    component.variants[props.variant ?? "default"],
    "listItem variant",
  );
  return {
    accessibilityLabel: props.supportingText
      ? `${props.label}. ${props.supportingText}`
      : props.label,
    container: [
      styles.root,
      {
        backgroundColor: string(variantRecipe.container, "listItem container"),
        borderColor: string(
          focused ? component.slots.focusBorder : variantRecipe.border,
          "listItem border",
        ),
        borderRadius: number(variantRecipe.radius, "listItem radius"),
        borderWidth: number(
          component.slots.focusBorderWidth,
          "listItem borderWidth",
        ),
        gap: number(sizeRecipe.contentGap, "listItem contentGap"),
        minHeight: number(sizeRecipe.minHeight, "listItem minHeight"),
        paddingHorizontal: number(
          sizeRecipe.horizontalPadding,
          "listItem horizontalPadding",
        ),
        paddingVertical: number(
          sizeRecipe.verticalPadding,
          "listItem verticalPadding",
        ),
      },
    ],
    labelColor: string(variantRecipe.label, "listItem label"),
    labelRole: string(sizeRecipe.labelRole, "listItem labelRole"),
    pressedOpacity: number(
      component.slots.pressedOpacity,
      "listItem pressedOpacity",
    ),
    supportingColor: string(variantRecipe.supporting, "listItem supporting"),
    supportingRole: string(
      sizeRecipe.supportingRole,
      "listItem supportingRole",
    ),
  };
}

const styles = StyleSheet.create({
  copy: { flexShrink: 1 },
  root: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
});
