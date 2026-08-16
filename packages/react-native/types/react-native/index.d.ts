declare module "react-native" {
  import type { Component, ComponentType, ReactNode, Ref } from "react";

  export interface AccessibilityState {
    busy?: boolean;
    checked?: boolean | "mixed";
    disabled?: boolean;
  }

  export interface TextStyle {
    color?: string;
    flexShrink?: number;
    fontFamily?: string;
    fontSize?: number;
    fontWeight?:
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
    lineHeight?: number;
    textAlign?: "auto" | "left" | "right" | "center" | "justify";
  }

  export interface ViewStyle {
    alignItems?: "flex-start" | "flex-end" | "center" | "stretch";
    backgroundColor?: string;
    borderColor?: string;
    borderRadius?: number;
    borderWidth?: number;
    flexDirection?: "row" | "column";
    flexShrink?: number;
    gap?: number;
    height?: number;
    justifyContent?: "flex-start" | "flex-end" | "center" | "space-between";
    minHeight?: number;
    minWidth?: number;
    opacity?: number;
    paddingHorizontal?: number;
    paddingVertical?: number;
    width?: number;
  }

  export type StyleProp<T> =
    | T
    | readonly StyleProp<T>[]
    | false
    | null
    | undefined;

  export interface TextProps {
    accessibilityLiveRegion?: "none" | "polite" | "assertive";
    allowFontScaling?: boolean;
    children?: ReactNode;
    maxFontSizeMultiplier?: number;
    ref?: Ref<Text>;
    role?: string;
    style?: StyleProp<TextStyle>;
  }

  export class Text extends Component<TextProps> {}

  export interface ViewProps {
    accessibilityLiveRegion?: "none" | "polite" | "assertive";
    accessibilityRole?: "alert" | "button" | "checkbox" | "radio";
    accessibilityElementsHidden?: boolean;
    children?: ReactNode;
    importantForAccessibility?: "auto" | "yes" | "no" | "no-hide-descendants";
    ref?: Ref<View>;
    style?: StyleProp<ViewStyle>;
    testID?: string;
  }

  export class View extends Component<ViewProps> {}

  export interface PressableStateCallbackType {
    pressed: boolean;
  }

  export interface PressableProps
    extends Omit<ViewProps, "children" | "style"> {
    accessibilityHint?: string;
    accessibilityLabel?: string;
    accessibilityRole?: "button" | "checkbox" | "radio";
    accessibilityState?: AccessibilityState;
    children?: ReactNode | ((state: PressableStateCallbackType) => ReactNode);
    disabled?: boolean;
    onBlur?: () => void;
    onFocus?: () => void;
    onPress?: () => void;
    style?:
      | StyleProp<ViewStyle>
      | ((state: PressableStateCallbackType) => StyleProp<ViewStyle>);
    testID?: string;
  }

  export const Pressable: ComponentType<PressableProps & { ref?: Ref<View> }>;
  export const ActivityIndicator: ComponentType<{
    color?: string;
    size?: "small" | "large";
  }>;

  export interface NativeSyntheticEvent<T> {
    nativeEvent: T;
  }

  export interface TextInputFocusEventData {}

  export interface TextInputProps {
    accessibilityLabel?: string;
    allowFontScaling?: boolean;
    editable?: boolean;
    maxFontSizeMultiplier?: number;
    onBlur?: (event: NativeSyntheticEvent<TextInputFocusEventData>) => void;
    onChangeText?: (text: string) => void;
    onFocus?: (event: NativeSyntheticEvent<TextInputFocusEventData>) => void;
    placeholder?: string;
    placeholderTextColor?: string;
    ref?: Ref<TextInput>;
    style?: StyleProp<TextStyle | ViewStyle>;
    testID?: string;
    value?: string;
  }

  export class TextInput extends Component<TextInputProps> {}

  export const StyleSheet: {
    create<T extends Record<string, TextStyle | ViewStyle>>(styles: T): T;
  };

  export function useWindowDimensions(): {
    fontScale: number;
    height: number;
    width: number;
  };
}
