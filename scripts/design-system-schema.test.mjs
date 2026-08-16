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

test("allows signed component spacing for product-owned optical alignment", () => {
  const valid = structuredClone(example);
  valid.foundations.spacing.component.overhang = -4;

  assert.equal(validate(valid), true, JSON.stringify(validate.errors));
});

test("continues to reject negative radius values", () => {
  const invalid = structuredClone(example);
  invalid.foundations.radius.component.button = -4;

  assert.equal(validate(invalid), false);
  assert.ok(validate.errors?.some((error) => error.keyword === "minimum"));
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
