import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { compileDesignSystem } from "../../compiler/src/index.mjs";
import {
  createDesignSystem,
  resolveScaledControlMinHeight,
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
