import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";

const root = new URL("../", import.meta.url);

async function json(path) {
  return JSON.parse(await readFile(new URL(path, root), "utf8"));
}

const catalog = await json("deslint/internal/rules/anti_slop_catalog.json");
const policy = await json("packages/schema/testdata/example-policy.json");
const release = await json("release/ansldes-release.json");

assert.equal(
  catalog.rules.length,
  63,
  "canonical rule registry must be exact 63",
);
assert.equal(
  new Set(
    catalog.rules.flatMap((rule) =>
      rule.sourceRuleIds.filter((id) => id.startsWith("impeccable/")),
    ),
  ).size,
  59,
  "Impeccable source registry must be exact 59",
);
assert.equal(
  catalog.hallmark.mappings.length,
  8,
  "Hallmark source catalog must be exact 8",
);
assert.equal(new Set(catalog.rules.map((rule) => rule.id)).size, 63);
assert.deepEqual(
  policy.rulePacks.map(({ id, version, fingerprintSha256 }) => ({
    id,
    version,
    fingerprintSha256,
  })),
  [
    {
      id: "ansldes-anti-slop",
      version: catalog.pack.version,
      fingerprintSha256: catalog.pack.fingerprintSha256,
    },
    {
      id: "ansldes-core",
      version: "1.12.0",
      fingerprintSha256:
        "977d5b250fcb57cca5c1af87215f2d477cf9132eeebe9de12b3659a3587e4561",
    },
  ],
  "neutral policy must pin the exact built-in packs",
);

assert.equal(
  release.dependencies.antiSlopCatalog.pack.fingerprintSha256,
  catalog.pack.fingerprintSha256,
  "release must pin the current anti-slop pack",
);
assert.equal(
  release.dependencies.impeccable.treeSha256,
  catalog.impeccable.treeSha256,
);
assert.equal(
  release.dependencies.hallmark.sourceSha256,
  catalog.hallmark.sourceSha256,
);
assert.equal(release.compatibility.consumerReleaseLockSchemaVersion, 1);

for (const path of [
  "packages/schema/web-provider-evidence.schema.json",
  "packages/schema/native-source-evidence.schema.json",
  "packages/schema/native-runtime-evidence.schema.json",
  "packages/schema/layout-evidence.schema.json",
  "packages/schema/consumer-release-lock.schema.json",
  "packages/schema/stage-execution-evidence.schema.json",
]) {
  await readFile(new URL(path, root));
}

for (const id of [
  "01-evidence",
  "02-consumer-profile",
  "03-design-system-awareness",
  "04-visual-detail",
  "05-typography",
  "06-color",
  "07-layout",
  "08-motion",
  "09-copy",
  "10-imagery",
  "11-runtime",
  "13-native",
  "14-web-gate",
  "15-native-pencil-provider",
  "16-governance",
  "17-integration",
]) {
  const contents = await readFile(
    new URL(`docs/anti-slop/todo/${id}.md`, root),
    "utf8",
  );
  assert.equal(
    /- \[ \]/u.test(contents),
    false,
    `${id} has an incomplete required item`,
  );
}

process.stdout.write(
  "Anti-slop contract: PASS (Impeccable 59, Hallmark 8, canonical 63)\n",
);
