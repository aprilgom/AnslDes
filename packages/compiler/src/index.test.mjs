import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";

import {
  canonicalStringify,
  compileDesignSystem,
  DesignSystemCompileError,
} from "./index.mjs";

const example = JSON.parse(
  await readFile(
    new URL("../../schema/testdata/example-product.json", import.meta.url),
  ),
);

test("compiles exact light and dark bundles with a stable fingerprint", () => {
  const first = compileDesignSystem(example);
  const reordered = JSON.parse(canonicalStringify(example));
  const second = compileDesignSystem(reordered);

  assert.equal(first.fingerprintSha256, second.fingerprintSha256);
  assert.equal(
    first.themes.light.foundations.color.semantic["text.primary"],
    "#18202A",
  );
  assert.equal(
    first.themes.dark.foundations.color.semantic["text.primary"],
    "#FFFFFF",
  );
  assert.equal(
    first.themes.light.components.button.variants.primary.container,
    "#4263EB",
  );
  assert.equal(
    first.themes.light.foundations.typography.roles["button.label"].fontWeight,
    700,
  );
  assert.equal(
    first.themes.light.foundations.icon.icons["arrow-right"].geometry[0].d,
    "M5 12h14M13 6l6 6-6 6",
  );
  assert.equal(
    first.themes.light.foundations.motion.transitions["control.press"]
      .reducedMotion.duration,
    0,
  );
});

test("rejects an unknown reference", () => {
  const invalid = structuredClone(example);
  invalid.foundations.radius.semantic.control = "{radius.primitive.unknown}";

  assert.throws(() => compileDesignSystem(invalid), DesignSystemCompileError);
});

test("rejects reference cycles", () => {
  const invalid = structuredClone(example);
  invalid.foundations.radius.semantic.control = "{radius.semantic.control}";

  assert.throws(
    () => compileDesignSystem(invalid),
    (error) =>
      error instanceof DesignSystemCompileError &&
      /reference cycle/u.test(error.message),
  );
});

test("rejects incomplete theme mappings", () => {
  const invalid = structuredClone(example);
  delete invalid.foundations.color.semantic["text.primary"].dark;

  assert.throws(
    () => compileDesignSystem(invalid),
    (error) =>
      error instanceof DesignSystemCompileError &&
      /themes must be exact/u.test(error.message),
  );
});

test("rejects typography roles that use a disallowed weight", () => {
  const invalid = structuredClone(example);
  invalid.foundations.typography.tokens["body.medium"].allowedWeights = [
    "regular",
  ];

  assert.throws(
    () => compileDesignSystem(invalid),
    (error) =>
      error instanceof DesignSystemCompileError &&
      /disallows weight/u.test(error.message),
  );
});

test("rejects icons that reference unknown sizes, strokes, or glyphs", () => {
  const invalid = structuredClone(example);
  invalid.foundations.icon.icons["arrow-right"].defaultSize = "missing";
  invalid.foundations.icon.icons["arrow-right"].geometry[0].strokeWidth =
    "missing";
  invalid.foundations.icon.usages["navigation.next"].icon = "missing";

  assert.throws(
    () => compileDesignSystem(invalid),
    (error) =>
      error instanceof DesignSystemCompileError &&
      /unknown default size/u.test(error.message) &&
      /unknown stroke/u.test(error.message) &&
      /unknown icon/u.test(error.message),
  );
});

test("does not overwrite an existing bundle when compilation fails", async () => {
  const directory = await mkdtemp(path.join(tmpdir(), "ansldes-compiler-"));
  const input = path.join(directory, "invalid.json");
  const output = path.join(directory, "bundle.json");
  const invalid = structuredClone(example);
  delete invalid.foundations.radius;
  await writeFile(input, JSON.stringify(invalid), "utf8");
  await writeFile(output, "sentinel\n", "utf8");

  try {
    const result = spawnSync(
      process.execPath,
      [
        new URL("./cli.mjs", import.meta.url).pathname,
        "compile",
        input,
        "--out",
        output,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(result.status, 0);
    assert.equal(await readFile(output, "utf8"), "sentinel\n");
  } finally {
    await rm(directory, { recursive: true });
  }
});
