// Code generated from https://ansldes.dev/schema/consumer-conformance.v1.json; DO NOT EDIT.
// consumer conformance schema SHA-256: 28b1b1e232e7ed4f61d66a95839b225c493701f0584fab74baca7e94153c0975

export const consumerConformanceSchemaVersion = 1 as const;
export const consumerConformanceSchemaSha256 =
  "28b1b1e232e7ed4f61d66a95839b225c493701f0584fab74baca7e94153c0975" as const;

export type ConsumerPlatform =
  | "web"
  | "react-native"
  | "ios"
  | "android"
  | "design-document";
export type ControlState =
  | "default"
  | "pressed"
  | "focused"
  | "disabled"
  | "loading"
  | "error"
  | "selected";

export interface ConsumerControl {
  id: string;
  actionId: string;
  role: string;
  component: string;
  label: string;
  shapeToken: string;
  icon?: string;
  feedback: string;
  states: ControlState[];
  contractStatus: "matched" | "mismatched";
  affordanceSource: "design-system" | "platform" | "consumer-exception" | "invented";
  motionPurpose: "none" | "state-transition" | "continuity" | "decorative";
  motionRecipeStatus: "none" | "approved" | "unapproved";
  reduceMotionFallback: boolean;
  rawDurationMs?: number;
  prominence: "standard" | "emphasized" | "oversized";
  nativePrimitive: boolean;
  exceptionId?: string;
}

export interface ConsumerConformanceEvidence {
  $schema?: string;
  schemaVersion: typeof consumerConformanceSchemaVersion;
  profileId: string;
  platform: ConsumerPlatform;
  surfaceId: string;
  controls: ConsumerControl[];
}
