import { createHash } from "node:crypto";
import { readdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const referenceRoot = path.join(
  root,
  "docs/anti-slop/references/impeccable-detector-2026",
);
const upstreamRoot = path.join(referenceRoot, "upstream");
const registryPath = path.join(
  upstreamRoot,
  "cli/engine/registry/antipatterns.mjs",
);
const outputPath = path.join(
  root,
  "deslint/internal/rules/anti_slop_catalog.json",
);
const migrationNote = {
  path: "docs/anti-slop/references/impeccable-detector-2026/MIGRATION.md",
  sha256: "f2573b1d0776901b291a7c72e3c74c4259493ba57d4e7058310c63eb5fe10c16",
};
const check = process.argv.includes("--check");
const write = process.argv.includes("--write");

const impeccable = {
  name: "impeccable",
  version: "3.6.0",
  integrity:
    "sha512-nysc6/2OHTWqLrcSxTxZk4r4QMufhU8NTIuG2ic6p5zzyZe45AWBX3/18OA5S88pCWq+4z8pKsjUxhAM990RKg==",
  commit: "7b646bafd60b9dd9828ce5c4c1a25691702c9e92",
  sourceTreeSha256:
    "84d3d3a66be62ea4c361420a6662f42010b123b680e00d8306e6a627ab5f5954",
  treeSha256:
    "a4dfe2060edf918b1241afac131ed9fbdf2d9eecaa448a6f57365e895da1908c",
};
const hallmark = {
  commit: "13ac0ec7e148655948100b6396439e481361d690",
  sourceSha256:
    "83083b5e37b99cb8211268b778aa0cb3f677f8fe53f8d817f54fcad799882e0e",
};
const hallmarkMappings = [
  ["hallmark-eight-01", "ai-color-palette"],
  ["hallmark-eight-02", "overused-font"],
  ["hallmark-eight-03", "equal-icon-feature-columns"],
  ["hallmark-eight-04", "full-viewport-centered-hero"],
  ["hallmark-eight-05", "nested-cards"],
  ["hallmark-eight-06", "gradient-text"],
  ["hallmark-eight-07", "pure-extreme-surface"],
  ["hallmark-eight-08", "unverified-social-proof"],
];
const categories = {
  visual: new Set([
    "side-tab",
    "border-accent-on-rounded",
    "gpt-thin-border-wide-shadow",
    "repeating-stripes-gradient",
    "codex-grid-background",
  ]),
  typography: new Set([
    "overused-font",
    "flat-type-hierarchy",
    "icon-tile-stack",
    "italic-serif-display",
    "hero-eyebrow-chip",
    "kicker-above-heading",
    "oversized-h1",
    "extreme-negative-tracking",
    "tight-leading",
    "skipped-heading",
    "tiny-text",
    "undersized-ui-text",
    "all-caps-body",
    "wide-tracking",
  ]),
  color: new Set([
    "gradient-text",
    "ai-color-palette",
    "cream-palette",
    "dark-glow",
    "radial-halo",
    "radial-spotlight-glow",
    "gray-on-color",
    "low-contrast",
  ]),
  layout: new Set([
    "nested-cards",
    "monotonous-spacing",
    "numbered-section-labels",
    "edge-flush-cards",
    "text-occlusion",
    "first-viewport-column-overflow",
    "heading-rhythm",
    "line-length",
    "cramped-padding",
    "body-text-viewport-edge",
    "text-overflow",
    "clipped-overflow-container",
  ]),
  motion: new Set([
    "bounce-easing",
    "pulsing-dot",
    "blinking-cursor",
    "marquee",
    "layout-transition",
    "image-hover-transform",
  ]),
  copy: new Set([
    "em-dash-overuse",
    "marketing-buzzword",
    "aphoristic-cadence",
    "repeated-container-text",
    "theater-slop-phrase",
  ]),
  imagery: new Set(["shape-assembled-illustration", "broken-image"]),
  runtime: new Set([
    "script-error",
    "content-hidden-at-rest",
    "justified-text",
  ]),
  "design-system": new Set([
    "design-system-font",
    "design-system-color",
    "design-system-radius",
    "design-system-font-size",
  ]),
};
const supplementCategories = new Map([
  ["equal-icon-feature-columns", "layout"],
  ["full-viewport-centered-hero", "layout"],
  ["pure-extreme-surface", "color"],
  ["unverified-social-proof", "copy"],
]);
const browserOnly = new Set(["script-error", "content-hidden-at-rest"]);
const sourceUnsupported = new Set([
  ...categories.layout,
  "low-contrast",
  "flat-type-hierarchy",
  "content-hidden-at-rest",
]);

const module = await import(pathToFileURL(registryPath));
const upstreamRules = module.ANTIPATTERNS;
if (upstreamRules.length !== 59) {
  throw new Error(`upstream rule count ${upstreamRules.length}, want 59`);
}
const packageJson = JSON.parse(
  await readFile(path.join(upstreamRoot, "package.json"), "utf8"),
);
if (
  packageJson.name !== impeccable.name ||
  packageJson.version !== impeccable.version
) {
  throw new Error("pinned detector package identity drifted");
}
const treeSha256 = await fingerprintTree(upstreamRoot);
if (treeSha256 !== impeccable.treeSha256) {
  throw new Error(
    `snapshot tree SHA-256 ${treeSha256}, want ${impeccable.treeSha256}`,
  );
}
const migrationNoteContents = await readFile(
  path.join(root, migrationNote.path),
);
const migrationNoteSha256 = createHash("sha256")
  .update(migrationNoteContents)
  .digest("hex");
if (migrationNoteSha256 !== migrationNote.sha256) {
  throw new Error(
    `migration note SHA-256 ${migrationNoteSha256}, want ${migrationNote.sha256}`,
  );
}

const mappingByCanonicalSource = new Map(
  hallmarkMappings.map(([sourceId, id]) => [id, sourceId]),
);
const rules = upstreamRules.map((rule) => {
  const namespace = namespaceFor(rule.id);
  const hallmarkSource = mappingByCanonicalSource.get(rule.id);
  return {
    id: `${namespace}/${rule.id.replace(/^design-system-/u, "")}`,
    implementationVersion: `${impeccable.name}@${impeccable.version}`,
    category: rule.category,
    scopes: [...(rule.scopes ?? [])].sort(),
    platforms: ["web"],
    evidenceKinds: evidenceKindsFor(rule.id),
    defaultSeverity:
      rule.advisory === true || rule.severity === "advisory"
        ? "advisory"
        : "error",
    upstreamAdvisory: rule.advisory === true,
    upstreamSeverity: rule.severity ?? "error",
    provenance: [
      `impeccable/${rule.id}@${impeccable.version}#${impeccable.commit}`,
      ...(hallmarkSource
        ? [`hallmark/${hallmarkSource}#${hallmark.commit}`]
        : []),
    ].sort(),
    sourceRuleIds: [
      `impeccable/${rule.id}`,
      ...(hallmarkSource ? [`hallmark/${hallmarkSource}`] : []),
    ].sort(),
    dependencies: dependenciesFor(rule.id),
    providers: providersFor(rule.id),
  };
});
for (const [sourceId, id] of hallmarkMappings
  .slice(2, 4)
  .concat(hallmarkMappings.slice(6))) {
  const namespace = supplementCategories.get(id);
  rules.push({
    id: `${namespace}/${id}`,
    implementationVersion: "hallmark-eight@2026-08-06",
    category: "supplement",
    scopes: namespace === "layout" ? ["layout"] : [],
    platforms: ["web"],
    evidenceKinds:
      namespace === "copy"
        ? ["consumer-content-registry", "web-source"]
        : ["web-rendered"],
    defaultSeverity: "error",
    upstreamAdvisory: false,
    upstreamSeverity: "error",
    provenance: [`hallmark/${sourceId}#${hallmark.commit}`],
    sourceRuleIds: [`hallmark/${sourceId}`],
    dependencies: dependenciesFor(id),
    providers: providersFor(id),
  });
}
rules.sort((left, right) => left.id.localeCompare(right.id));
assertUnique(
  rules.map((rule) => rule.id),
  "canonical rule",
);
if (rules.length !== 63) {
  throw new Error(`canonical rule count ${rules.length}, want 63`);
}
assertUnique(
  hallmarkMappings.map(([sourceId]) => sourceId),
  "Hallmark source",
);
if (hallmarkMappings.length !== 8) {
  throw new Error("Hallmark mapping must contain exactly eight source rules");
}

const pack = {
  id: "ansldes-anti-slop",
  version: "1.1.0",
  fingerprintSha256: packFingerprint(
    "ansldes-anti-slop",
    "1.1.0",
    rules.map((rule) => rule.id),
  ),
};
const manifest = {
  schemaVersion: 1,
  migrationNote,
  impeccable: { ...impeccable, computedTreeSha256: treeSha256 },
  hallmark: {
    ...hallmark,
    mappings: hallmarkMappings.map(([sourceId, canonicalSourceId]) => ({
      sourceId,
      canonicalSourceId,
    })),
  },
  pack,
  rules,
};
const serialized = `${JSON.stringify(manifest, null, 2)}\n`;
if (write) {
  await writeFile(outputPath, serialized);
} else if (check) {
  const actual = await readFile(outputPath, "utf8").catch(() => "");
  if (actual !== serialized) {
    throw new Error(
      "anti-slop catalog is stale; run npm run generate:anti-slop-registry",
    );
  }
} else {
  process.stdout.write(serialized);
}
console.log(
  `anti-slop registry: PASS ${rules.length} rules ${pack.fingerprintSha256}`,
);

function namespaceFor(id) {
  for (const [namespace, ids] of Object.entries(categories)) {
    if (ids.has(id)) return namespace;
  }
  throw new Error(`upstream rule ${id} has no canonical namespace mapping`);
}

function providersFor(id) {
  const result = ["browser"];
  if (!browserOnly.has(id)) result.push("static-html");
  if (!sourceUnsupported.has(id)) result.push("regex-source");
  if (id === "low-contrast") result.push("visual-contrast");
  return result.sort();
}

function evidenceKindsFor(id) {
  const providers = providersFor(id);
  const kinds = new Set();
  if (providers.includes("regex-source") || providers.includes("static-html")) {
    kinds.add("web-source");
  }
  if (providers.includes("browser") || providers.includes("visual-contrast")) {
    kinds.add("web-rendered");
  }
  return [...kinds].sort();
}

function dependenciesFor(id) {
  if (id.startsWith("design-system-")) return ["design-context"];
  if (id === "unverified-social-proof") return ["consumer-content-registry"];
  if (id === "low-contrast") return ["computed-color"];
  return [];
}

function packFingerprint(id, version, members) {
  return createHash("sha256")
    .update(JSON.stringify({ id, version, members: [...members].sort() }))
    .digest("hex");
}

async function fingerprintTree(directory) {
  const files = await collectFiles(directory, directory);
  const hash = createHash("sha256");
  for (const relativePath of files.sort()) {
    const contents = await readFile(path.join(directory, relativePath));
    const fileHash = createHash("sha256").update(contents).digest("hex");
    hash.update(relativePath.replaceAll("\\", "/"));
    hash.update("\0");
    hash.update(fileHash);
    hash.update("\n");
  }
  return hash.digest("hex");
}

async function collectFiles(directory, rootDirectory) {
  const result = [];
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const absolutePath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      result.push(...(await collectFiles(absolutePath, rootDirectory)));
    } else {
      result.push(path.relative(rootDirectory, absolutePath));
    }
  }
  return result;
}

function assertUnique(values, label) {
  if (new Set(values).size !== values.length) {
    throw new Error(`${label} identity is duplicated`);
  }
}
