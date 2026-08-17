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
const reportSchema = JSON.parse(
  await readFile(
    new URL("../packages/schema/deslint-report.schema.json", import.meta.url),
  ),
);
const conformanceSchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/consumer-conformance.schema.json",
      import.meta.url,
    ),
  ),
);
const consumerReleaseLockSchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/consumer-release-lock.schema.json",
      import.meta.url,
    ),
  ),
);
const designContextSchema = JSON.parse(
  await readFile(
    new URL("../packages/schema/design-context.schema.json", import.meta.url),
  ),
);
const visualDetailSchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/visual-detail-evidence.schema.json",
      import.meta.url,
    ),
  ),
);
const typographySchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/typography-evidence.schema.json",
      import.meta.url,
    ),
  ),
);
const colorEvidenceSchema = JSON.parse(
  await readFile(
    new URL("../packages/schema/color-evidence.schema.json", import.meta.url),
  ),
);
const layoutEvidenceSchema = JSON.parse(
  await readFile(
    new URL("../packages/schema/layout-evidence.schema.json", import.meta.url),
  ),
);
const motionEvidenceSchema = JSON.parse(
  await readFile(
    new URL("../packages/schema/motion-evidence.schema.json", import.meta.url),
  ),
);
const copyEvidenceSchema = JSON.parse(
  await readFile(
    new URL("../packages/schema/copy-evidence.schema.json", import.meta.url),
  ),
);
const imageryEvidenceSchema = JSON.parse(
  await readFile(
    new URL("../packages/schema/imagery-evidence.schema.json", import.meta.url),
  ),
);
const runtimeEvidenceSchema = JSON.parse(
  await readFile(
    new URL("../packages/schema/runtime-evidence.schema.json", import.meta.url),
  ),
);
const nativeSourceEvidenceSchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/native-source-evidence.schema.json",
      import.meta.url,
    ),
  ),
);
const nativeRuntimeEvidenceSchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/native-runtime-evidence.schema.json",
      import.meta.url,
    ),
  ),
);
const webProviderEvidenceSchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/web-provider-evidence.schema.json",
      import.meta.url,
    ),
  ),
);
const stageExecutionEvidenceSchema = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/stage-execution-evidence.schema.json",
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
const validateReport = ajv.compile(reportSchema);
const validateConformance = ajv.compile(conformanceSchema);
const validateConsumerReleaseLock = ajv.compile(consumerReleaseLockSchema);
const validateDesignContext = ajv.compile(designContextSchema);
const validateVisualDetail = ajv.compile(visualDetailSchema);
const validateTypography = ajv.compile(typographySchema);
const validateColorEvidence = ajv.compile(colorEvidenceSchema);
const validateLayoutEvidence = ajv.compile(layoutEvidenceSchema);
const validateMotionEvidence = ajv.compile(motionEvidenceSchema);
const validateCopyEvidence = ajv.compile(copyEvidenceSchema);
const validateImageryEvidence = ajv.compile(imageryEvidenceSchema);
const validateRuntimeEvidence = ajv.compile(runtimeEvidenceSchema);
const validateNativeSourceEvidence = ajv.compile(nativeSourceEvidenceSchema);
const validateNativeRuntimeEvidence = ajv.compile(nativeRuntimeEvidenceSchema);
const validateWebProviderEvidence = ajv.compile(webProviderEvidenceSchema);
const validateStageExecutionEvidence = ajv.compile(
  stageExecutionEvidenceSchema,
);
const examplePolicy = JSON.parse(
  await readFile(
    new URL("../packages/schema/testdata/example-policy.json", import.meta.url),
  ),
);
const exampleReport = JSON.parse(
  await readFile(
    new URL("../packages/schema/testdata/example-report.json", import.meta.url),
  ),
);
const conformanceExample = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/testdata/operate-conformance.json",
      import.meta.url,
    ),
  ),
);
const designContextExample = JSON.parse(
  await readFile(
    new URL(
      "../packages/schema/testdata/generated-design-context/.impeccable/design.json",
      import.meta.url,
    ),
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

test("accepts a deterministic report with separate evidence and activation status", () => {
  assert.equal(
    validateReport(exampleReport),
    true,
    JSON.stringify(validateReport.errors),
  );
});

test("rejects legacy evidence kinds and unknown report status", () => {
  const legacyEvidence = structuredClone(exampleReport);
  legacyEvidence.evidence[0].kind = "source";
  assert.equal(validateReport(legacyEvidence), false);

  const unknownStatus = structuredClone(exampleReport);
  unknownStatus.ruleSet.rules[0].status = "ignored";
  assert.equal(validateReport(unknownStatus), false);

  const fabricatedPass = structuredClone(exampleReport);
  fabricatedPass.diagnostics.push({
    ruleId: "source/raw-value",
    sourceRuleIds: [],
    status: "pass",
    severity: "error",
    message: "fabricated pass",
    path: "src/Example.tsx",
    evidenceKind: "native-source",
    platform: "react-native",
    owner: "example-owner",
    fingerprint: "4".repeat(64),
  });
  assert.equal(validateReport(fabricatedPass), false);
});

test("accepts product-neutral consumer conformance evidence", () => {
  assert.equal(
    validateConformance(conformanceExample),
    true,
    JSON.stringify(validateConformance.errors),
  );
});

test("rejects unknown conformance fields and duplicate states", () => {
  const unknown = structuredClone(conformanceExample);
  unknown.controls[0].consumerPath = "src/Product.tsx";
  assert.equal(validateConformance(unknown), false);

  const duplicateState = structuredClone(conformanceExample);
  duplicateState.controls[0].states.push("default");
  assert.equal(validateConformance(duplicateState), false);
});

test("accepts generated design context and rejects stale-shaped metadata", () => {
  assert.equal(
    validateDesignContext(designContextExample),
    true,
    JSON.stringify(validateDesignContext.errors),
  );
  const unknown = structuredClone(designContextExample);
  unknown.generator.manualTokens = true;
  assert.equal(validateDesignContext(unknown), false);
});

test("accepts strict visual-detail provider evidence", async () => {
  for (const fixture of [
    "visual-detail-web.json",
    "visual-detail-native.json",
    "visual-detail-design-document.json",
    "visual-detail-permissions.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateVisualDetail(value),
      true,
      JSON.stringify(validateVisualDetail.errors),
    );
  }
});

test("accepts profile-scoped typography evidence and font-scale fixtures", async () => {
  for (const fixture of [
    "typography-negative.json",
    "typography-positive-100.json",
    "typography-positive-235.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateTypography(value),
      true,
      JSON.stringify(validateTypography.errors),
    );
  }
});

test("accepts theme-specific screenshot and computed color evidence", async () => {
  for (const fixture of [
    "color-negative-light.json",
    "color-permissions-dark.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateColorEvidence(value),
      true,
      JSON.stringify(validateColorEvidence.errors),
    );
  }
});

test("accepts browser, native, and design-document layout evidence", async () => {
  for (const fixture of [
    "layout-negative-web.json",
    "layout-permissions-native.json",
    "layout-design-document.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateLayoutEvidence(value),
      true,
      JSON.stringify(validateLayoutEvidence.errors),
    );
  }
});

test("accepts source, reduced-runtime, and design-document motion evidence", async () => {
  for (const fixture of [
    "motion-negative-source.json",
    "motion-reduced-simulator.json",
    "motion-design-document.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateMotionEvidence(value),
      true,
      JSON.stringify(validateMotionEvidence.errors),
    );
  }
});

test("accepts Korean, English advisory, and not-run copy evidence", async () => {
  for (const fixture of [
    "copy-ko-negative.json",
    "copy-ko-positive.json",
    "copy-en-advisory.json",
    "copy-registry-not-run.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateCopyEvidence(value),
      true,
      JSON.stringify(validateCopyEvidence.errors),
    );
  }
});

test("accepts Web, native, and permission-scoped imagery evidence", async () => {
  for (const fixture of [
    "imagery-negative-web.json",
    "imagery-negative-native.json",
    "imagery-permissions.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateImageryEvidence(value),
      true,
      JSON.stringify(validateImageryEvidence.errors),
    );
  }
});

test("accepts completed Web/native and failed detector runtime evidence", async () => {
  for (const fixture of [
    "runtime-negative-web.json",
    "runtime-negative-native.json",
    "runtime-permissions.json",
    "runtime-detector-failure-web.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateRuntimeEvidence(value),
      true,
      JSON.stringify(validateRuntimeEvidence.errors),
    );
  }
});

test("accepts separate React Native source and platform runtime conformance evidence", async () => {
  for (const fixture of [
    "native-source-negative.json",
    "native-source-positive.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateNativeSourceEvidence(value),
      true,
      JSON.stringify(validateNativeSourceEvidence.errors),
    );
  }
  for (const fixture of [
    "native-runtime-negative-ios.json",
    "native-runtime-negative-android.json",
    "native-runtime-positive-ios.json",
    "native-runtime-positive-android.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateNativeRuntimeEvidence(value),
      true,
      JSON.stringify(validateNativeRuntimeEvidence.errors),
    );
  }
});

test("accepts completed, excluded, not-run, and failed Web provider evidence", async () => {
  for (const fixture of [
    "web-provider-regex-negative.json",
    "web-provider-regex-positive.json",
    "web-provider-artifact-excluded.json",
    "web-provider-static-positive.json",
    "web-provider-static-negative.json",
    "web-provider-browser-negative.json",
    "web-provider-browser-positive.json",
    "web-provider-visual-negative.json",
    "web-provider-visual-positive.json",
    "web-provider-fallback-not-run.json",
    "web-provider-browser-error.json",
  ]) {
    const value = JSON.parse(
      await readFile(
        new URL(`../packages/schema/testdata/${fixture}`, import.meta.url),
      ),
    );
    assert.equal(
      validateWebProviderEvidence(value),
      true,
      `${fixture}: ${JSON.stringify(validateWebProviderEvidence.errors)}`,
    );
  }
});

test("accepts an exact consumer release lock and rejects checksum drift", async () => {
  const value = JSON.parse(
    await readFile(
      new URL(
        "../packages/schema/testdata/example-consumer-lock.json",
        import.meta.url,
      ),
    ),
  );
  assert.equal(
    validateConsumerReleaseLock(value),
    true,
    JSON.stringify(validateConsumerReleaseLock.errors),
  );
  value.release.manifestSha256 = "not-a-checksum";
  assert.equal(validateConsumerReleaseLock(value), false);
});

test("accepts exact stage execution evidence and rejects exit-code rewriting", async () => {
  const value = JSON.parse(
    await readFile(
      new URL(
        "../packages/schema/testdata/stage-execution-positive.json",
        import.meta.url,
      ),
    ),
  );
  assert.equal(validateStageExecutionEvidence(value), true);
  value.exitCode = 2;
  assert.equal(validateStageExecutionEvidence(value), false);
});
