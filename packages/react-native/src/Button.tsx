import { forwardRef, type ReactNode, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  StyleSheet,
  useWindowDimensions,
  View,
} from "react-native";

import { resolveButtonPresentation } from "./buttonRecipe";
import { useDesignSystem } from "./context";
import { DesignText } from "./DesignText";

export interface ButtonProps {
  accessibilityContext?: string;
  accessibilityHint?: string;
  disabled?: boolean;
  label: string;
  leadingIcon?: ReactNode;
  loading?: boolean;
  onPress: () => void;
  size?: string;
  testID?: string;
  trailingIcon?: ReactNode;
  variant?: string;
}

export const Button = forwardRef<View, ButtonProps>(function Button(
  {
    accessibilityContext,
    accessibilityHint,
    disabled = false,
    label,
    leadingIcon,
    loading = false,
    onPress,
    size = "medium",
    testID,
    trailingIcon,
    variant = "primary",
  },
  ref,
) {
  const runtime = useDesignSystem();
  const [focused, setFocused] = useState(false);
  const { fontScale } = useWindowDimensions();

  return (
    <Pressable
      ref={ref}
      accessibilityHint={accessibilityHint}
      accessibilityRole="button"
      disabled={disabled || loading}
      onBlur={() => setFocused(false)}
      onFocus={() => setFocused(true)}
      onPress={onPress}
      testID={testID}
      style={({ pressed }) => {
        const presentation = resolveButtonPresentation(runtime, {
          accessibilityContext,
          disabled,
          focused,
          fontScale,
          label,
          loading,
          pressed,
          size,
          variant,
        });
        return [styles.root, presentation.container];
      }}
      {...resolveAccessibility(runtime, {
        accessibilityContext,
        disabled,
        focused,
        fontScale,
        label,
        loading,
        size,
        variant,
      })}
    >
      {({ pressed }) => {
        const presentation = resolveButtonPresentation(runtime, {
          accessibilityContext,
          disabled,
          focused,
          fontScale,
          label,
          loading,
          pressed,
          size,
          variant,
        });
        return (
          <View style={[styles.content, { gap: presentation.container.gap }]}>
            {loading ? (
              <View
                accessibilityElementsHidden
                importantForAccessibility="no-hide-descendants"
              >
                <ActivityIndicator
                  color={presentation.labelColor}
                  size="small"
                />
              </View>
            ) : leadingIcon ? (
              <View
                accessibilityElementsHidden
                importantForAccessibility="no-hide-descendants"
              >
                {leadingIcon}
              </View>
            ) : null}
            <DesignText
              typographyRole={presentation.typography.semanticRole}
              style={[styles.label, { color: presentation.labelColor }]}
            >
              {label}
            </DesignText>
            {!loading && trailingIcon ? (
              <View
                accessibilityElementsHidden
                importantForAccessibility="no-hide-descendants"
              >
                {trailingIcon}
              </View>
            ) : null}
          </View>
        );
      }}
    </Pressable>
  );
});

function resolveAccessibility(
  runtime: ReturnType<typeof useDesignSystem>,
  options: Omit<Parameters<typeof resolveButtonPresentation>[1], "pressed">,
) {
  const presentation = resolveButtonPresentation(runtime, {
    ...options,
    pressed: false,
  });
  return {
    accessibilityLabel: presentation.accessibilityLabel,
    accessibilityState: presentation.accessibilityState,
  };
}

const styles = StyleSheet.create({
  content: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "center",
    minWidth: 0,
  },
  label: {
    flexShrink: 1,
    textAlign: "center",
  },
  root: {
    alignItems: "center",
    justifyContent: "center",
  },
});
