import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { compileDesignSystem } from "../../compiler/src/index.mjs";
import {
  createDesignSystem,
  mapDigitSelectionThroughFormat,
  resolveScaledControlMinHeight,
  resolveToggleThumbTravel,
} from "../dist/index.js";

const definition = JSON.parse(
  await readFile(
    new URL("../../schema/testdata/example-product.json", import.meta.url),
  ),
);
const bundle = compileDesignSystem(definition);

test("selects a product theme without embedding product token names in core", () => {
  const light = createDesignSystem(bundle, { theme: "light" });
  const dark = light.withTheme("dark");

  assert.equal(light.color("semantic", "text.primary"), "#18202A");
  assert.equal(dark.color("semantic", "text.primary"), "#FFFFFF");
  assert.deepEqual(light.availableThemes, ["dark", "light"]);
});

test("keeps framework-neutral typography values resolved", () => {
  const typography = createDesignSystem(bundle).typography("button.label");

  assert.equal(typography.fontWeight, 700);
  assert.equal(typography.fontFamily, "ExampleSans-Bold");
});

test("resolves framework-neutral icon, motion, and elevation recipes", () => {
  const runtime = createDesignSystem(bundle);

  assert.equal(runtime.iconSize("medium"), 20);
  assert.equal(runtime.iconStroke("regular"), 2);
  assert.equal(runtime.icon("arrow-right").geometry[0].kind, "path");
  assert.equal(runtime.iconUsage("navigation.next").icon, "arrow-right");
  assert.equal(runtime.iconAction("navigation.forward").label, "Continue");
  assert.deepEqual(runtime.motionTransition("control.press", false), {
    duration: 160,
    easing: [0.2, 0, 0, 1],
    fallback: null,
  });
  assert.deepEqual(runtime.motionTransition("control.press", true), {
    duration: 0,
    easing: [0.2, 0, 0, 1],
    fallback: "instant",
  });
  assert.deepEqual(runtime.elevation("raised"), { level: 1 });
});

test("keeps interaction geometry and formatted selection framework-neutral", () => {
  assert.equal(
    resolveToggleThumbTravel({ trackWidth: 46, thumbSize: 22, inset: 3 }),
    18,
  );
  assert.deepEqual(
    mapDigitSelectionThroughFormat("1234", "1,234", { start: 2, end: 4 }),
    {
      start: 3,
      end: 5,
    },
  );
});

test("scales control height through the role maximum without shrinking the base", () => {
  assert.equal(
    resolveScaledControlMinHeight({
      baseHeight: 52,
      fontScale: 2.35,
      lineHeight: 24,
      maximumFontScale: 2.35,
      verticalPadding: 8,
    }),
    73,
  );
  assert.equal(
    resolveScaledControlMinHeight({
      baseHeight: 52,
      fontScale: 0.5,
      lineHeight: 24,
      maximumFontScale: 2.35,
      verticalPadding: 8,
    }),
    52,
  );
});

test("rejects unknown themes and missing token names", () => {
  assert.throws(
    () => createDesignSystem(bundle, { theme: "unknown" }),
    /unknown theme/u,
  );
  assert.throws(
    () => createDesignSystem(bundle).color("semantic", "missing"),
    /must be a string/u,
  );
});

test("keeps definition v1 bundles without the optional icon group compatible", () => {
  const withoutIcon = structuredClone(definition);
  delete withoutIcon.foundations.icon;
  const runtime = createDesignSystem(compileDesignSystem(withoutIcon));

  assert.equal(runtime.color("semantic", "text.primary"), "#18202A");
  assert.throws(() => runtime.icon("arrow-right"), /foundations.icon/u);
});
