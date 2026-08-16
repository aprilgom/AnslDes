import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtemp, readFile, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import {
  createManifest,
  fingerprint,
  serialize,
} from "./release-manifest.mjs";

test("keeps the checked-in release manifest byte-for-byte fresh", async () => {
  const expected = serialize(await createManifest());
  const actual = await readFile(
    new URL("../release/ansldes-release.json", import.meta.url),
    "utf8",
  );
  assert.equal(actual, expected);
});

test("fingerprints sorted paths and changes when artifact content changes", async () => {
  const root = await mkdtemp(path.join(os.tmpdir(), "ansldes-release-"));
  await writeFile(path.join(root, "a.ts"), "export const a = 1;\n", "utf8");
  await writeFile(path.join(root, "b.ts"), "export const b = 2;\n", "utf8");

  const forward = await fingerprint(["a.ts", "b.ts"], root);
  const reverse = await fingerprint(["b.ts", "a.ts"], root);
  assert.deepEqual(reverse, forward);

  await writeFile(path.join(root, "b.ts"), "export const b = 3;\n", "utf8");
  const changed = await fingerprint(["a.ts", "b.ts"], root);
  assert.notEqual(changed.sha256, forward.sha256);
});

test("accepts only the exact release tag from the manifest", () => {
  const script = fileURLToPath(
    new URL("./check-release-tag.mjs", import.meta.url),
  );
  const accepted = spawnSync(process.execPath, [script, "v0.2.0"], {
    encoding: "utf8",
  });
  assert.equal(accepted.status, 0, accepted.stderr);

  const rejected = spawnSync(process.execPath, [script, "v0.1.1"], {
    encoding: "utf8",
  });
  assert.notEqual(rejected.status, 0);
  assert.match(rejected.stderr, /does not match manifest/);
});

test("package exports support import and CommonJS-aware toolchains", async () => {
  const corePackage = JSON.parse(
    await readFile(new URL("../packages/core/package.json", import.meta.url), "utf8"),
  );
  const reactNativePackage = JSON.parse(
    await readFile(
      new URL("../packages/react-native/package.json", import.meta.url),
      "utf8",
    ),
  );

  for (const packageJson of [corePackage, reactNativePackage]) {
    assert.equal(packageJson.exports["."].import, "./dist/index.js");
    assert.equal(packageJson.exports["."].require, "./dist/index.js");
    assert.equal(packageJson.exports["."].default, "./dist/index.js");
  }
});
