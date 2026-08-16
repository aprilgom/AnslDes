import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import Ajv2020 from "ajv/dist/2020.js";

const schema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/design-system-definition.schema.json",
      import.meta.url,
    ),
  ),
);
const policySchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/design-system-policy.schema.json",
      import.meta.url,
    ),
  ),
);
const example = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/testdata/example-product.json",
      import.meta.url,
    ),
  ),
);
const ajv = new Ajv2020({ allErrors: true, strict: true });
const validate = ajv.compile(schema);
const validatePolicy = ajv.compile(policySchema);
const examplePolicy = JSON.parse(
  await readFile(
    new URL("../packages/schema/testdata/example-policy.json", import.meta.url),
  ),
);

test("accepts a product-owned definition without embedding product policy", () => {
  assert.equal(validate(example), true, JSON.stringify(validate.errors));
});

test("rejects a definition that omits a required foundation group", () => {
  const invalid = structuredClone(example);
  delete invalid.foundations.radius;

  assert.equal(validate(invalid), false);
  assert.ok(
    validate.errors?.some((error) => error.params.missingProperty === "radius"),
  );
});

test("rejects raw component references outside the generic reference grammar", () => {
  const invalid = structuredClone(example);
  invalid.components.button.variants.primary.containerColor =
    "{product.brand.primary}";

  assert.equal(validate(invalid), false);
  assert.ok(validate.errors?.some((error) => error.keyword === "pattern"));
});

test("rejects literals outside the primitive layer", () => {
  for (const [collection, layer, name, value] of [
    ["spacing", "semantic", "control.gap", 8],
    ["spacing", "component", "button.horizontal", -4],
    ["radius", "semantic", "control", 12],
    ["size", "component", "button", 52],
  ]) {
    const invalid = structuredClone(example);
    invalid.foundations[collection][layer][name] = value;

    assert.equal(validate(invalid), false, `${collection}.${layer}.${name}`);
    assert.ok(validate.errors?.some((error) => error.keyword === "type"));
  }
});

test("rejects negative primitive radius values", () => {
  const invalid = structuredClone(example);
  invalid.foundations.radius.primitive.medium = -4;

  assert.equal(validate(invalid), false);
  assert.ok(validate.errors?.some((error) => error.keyword === "minimum"));
});

test("rejects layer skipping and same-layer aliases", () => {
  const colorInvalid = structuredClone(example);
  colorInvalid.foundations.color.semantic["text.primary"].light =
    "{color.semantic.text.secondary}";
  assert.equal(validate(colorInvalid), false, "color.semantic.text.primary");
  assert.ok(validate.errors?.some((error) => error.keyword === "pattern"));

  for (const [collection, layer, name, reference] of [
    [
      "spacing",
      "semantic",
      "control.gap",
      "{spacing.semantic.control.vertical}",
    ],
    ["radius", "component", "button", "{radius.primitive.medium}"],
    ["size", "component", "button", "{size.primitive.control}"],
  ]) {
    const invalid = structuredClone(example);
    invalid.foundations[collection][layer][name] = reference;

    assert.equal(validate(invalid), false, `${collection}.${layer}.${name}`);
    assert.ok(validate.errors?.some((error) => error.keyword === "pattern"));
  }
});

test("accepts generic icon geometry and rejects unknown icon keys", () => {
  assert.equal(validate(example), true, JSON.stringify(validate.errors));

  const invalid = structuredClone(example);
  invalid.foundations.icon.icons["arrow-right"].consumerCount = 3;
  assert.equal(validate(invalid), false);
  assert.ok(
    validate.errors?.some(
      (error) =>
        error.keyword === "additionalProperties" &&
        error.params.additionalProperty === "consumerCount",
    ),
  );
});

test("keeps the icon foundation optional for definition v2 consumers", () => {
  const compatible = structuredClone(example);
  delete compatible.foundations.icon;

  assert.equal(validate(compatible), true, JSON.stringify(validate.errors));
});

test("accepts a product policy as a separate versioned input", () => {
  assert.equal(
    validatePolicy(examplePolicy),
    true,
    JSON.stringify(validatePolicy.errors),
  );
});

test("rejects unknown policy keys and incomplete exceptions", () => {
  const unknown = structuredClone(examplePolicy);
  unknown.source.ignore = ["**/*"];
  assert.equal(validatePolicy(unknown), false);

  const incomplete = structuredClone(examplePolicy);
  incomplete.exceptions.push({
    ruleId: "source/raw-value",
    path: "src/Example.tsx",
  });
  assert.equal(validatePolicy(incomplete), false);
});
