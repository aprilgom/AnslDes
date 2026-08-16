import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { compileDesignSystem } from "../../compiler/src/index.mjs";
import { createDesignSystem } from "../../core/dist/index.js";
import {
  assertButtonContract,
  resolveButtonPresentation,
} from "../dist/buttonRecipe.js";

const definition = JSON.parse(
  await readFile(
    new URL("../../schema/testdata/example-product.json", import.meta.url),
  ),
);
const runtime = createDesignSystem(compileDesignSystem(definition));

test("resolves button visuals only from product slots, variant, and size recipes", () => {
  assertButtonContract(runtime.component("button"));
  const presentation = resolveButtonPresentation(runtime, {
    disabled: false,
    focused: true,
    fontScale: 1,
    label: "Continue",
    loading: false,
    pressed: true,
    size: "medium",
    variant: "primary",
  });

  assert.equal(presentation.container.backgroundColor, "#18202A");
  assert.equal(presentation.container.borderColor, "#4263EB");
  assert.equal(presentation.container.borderRadius, 12);
  assert.equal(presentation.container.minHeight, 52);
  assert.deepEqual(presentation.accessibilityState, {
    busy: false,
    disabled: false,
  });
});

test("keeps loading unavailable and expands at 235 percent text", () => {
  const presentation = resolveButtonPresentation(runtime, {
    accessibilityContext: "Document",
    disabled: false,
    focused: false,
    fontScale: 2.35,
    label: "Remove",
    loading: true,
    pressed: false,
    size: "medium",
    variant: "primary",
  });

  assert.equal(presentation.accessibilityLabel, "Document. Remove");
  assert.deepEqual(presentation.accessibilityState, {
    busy: true,
    disabled: true,
  });
  assert.ok(presentation.container.minHeight > 52);
});
