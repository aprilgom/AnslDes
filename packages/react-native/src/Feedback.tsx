import type { ReactNode } from "react";
import { StyleSheet, View } from "react-native";

import { useDesignSystem } from "./context";
import { DesignText } from "./DesignText";
import { number, recipe, string } from "./recipe";

export interface FeedbackProps {
  action?: ReactNode;
  description: string;
  title: string;
  tone?: string;
}

export function Feedback({
  action,
  description,
  title,
  tone = "info",
}: FeedbackProps) {
  const runtime = useDesignSystem();
  const component = runtime.component("feedback");
  const variant = recipe(component.variants[tone], `feedback tone ${tone}`);
  const size = recipe(component.sizes.medium, "feedback size medium");
  const urgent = variant.liveRegion === "assertive";
  return (
    <View
      accessibilityLiveRegion={urgent ? "assertive" : "polite"}
      accessibilityRole={urgent ? "alert" : undefined}
      style={[
        styles.root,
        {
          backgroundColor: string(variant.container, "feedback container"),
          borderColor: string(variant.border, "feedback border"),
          borderRadius: number(variant.radius, "feedback radius"),
          borderWidth: number(
            component.slots.borderWidth,
            "feedback borderWidth",
          ),
          gap: number(size.contentGap, "feedback contentGap"),
          paddingHorizontal: number(
            size.horizontalPadding,
            "feedback horizontalPadding",
          ),
          paddingVertical: number(
            size.verticalPadding,
            "feedback verticalPadding",
          ),
        },
      ]}
    >
      <DesignText
        typographyRole={string(size.titleRole, "feedback titleRole")}
        style={{ color: string(variant.title, "feedback title") }}
      >
        {title}
      </DesignText>
      <DesignText
        typographyRole={string(
          size.descriptionRole,
          "feedback descriptionRole",
        )}
        style={{ color: string(variant.description, "feedback description") }}
      >
        {description}
      </DesignText>
      {action}
    </View>
  );
}

const styles = StyleSheet.create({
  root: { minWidth: 0 },
});
