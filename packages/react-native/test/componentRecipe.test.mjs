import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { compileDesignSystem } from "../../compiler/src/index.mjs";
import { createDesignSystem } from "../../core/dist/index.js";
import { resolveNativeTextRecipe } from "../dist/textRecipe.js";

const definition = JSON.parse(
  await readFile(
    new URL("../../schema/testdata/example-product.json", import.meta.url),
  ),
);
const runtime = createDesignSystem(compileDesignSystem(definition));

test("exposes product-owned anatomy for every generic React Native recipe", () => {
  assert.deepEqual(Object.keys(runtime.component("textField").variants), [
    "default",
  ]);
  assert.deepEqual(runtime.component("selection").semantics.roles, [
    "checkbox",
    "radio",
  ]);
  assert.equal(runtime.component("listItem").sizes.twoLine.minHeight, 52);
  assert.equal(
    runtime.component("feedback").variants.danger.liveRegion,
    "assertive",
  );
});

test("switches component recipe colors through the selected product theme", () => {
  const light = runtime.component("textField").variants.default;
  const dark = runtime.withTheme("dark").component("textField")
    .variants.default;

  assert.equal(light.container, "#FFFFFF");
  assert.equal(dark.container, "#18202A");
});

test("keeps React Native font-weight representation in the adapter", () => {
  assert.equal(
    resolveNativeTextRecipe(runtime, "button.label", "numeric").fontWeight,
    "700",
  );
  assert.equal(
    resolveNativeTextRecipe(runtime, "button.label", "normal").fontWeight,
    "normal",
  );
});
