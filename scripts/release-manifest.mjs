import { createHash } from "node:crypto";
import { readdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const manifestPath = path.join(
  repositoryRoot,
  "release",
  "ansldes-release.json",
);

const artifactRoots = {
  schema: ["packages/schema"],
  compiler: ["packages/compiler"],
  core: ["packages/core"],
  reactNative: ["packages/react-native"],
  deslint: [
    "deslint/go.mod",
    "deslint/go.sum",
    "deslint/cmd",
    "deslint/internal",
  ],
};

const ignoredDirectoryNames = new Set(["dist", "node_modules"]);
const includedExtensions = new Set([
  ".go",
  ".json",
  ".md",
  ".mjs",
  ".sum",
  ".mod",
  ".ts",
  ".tsx",
]);

async function readJson(relativePath) {
  return JSON.parse(
    await readFile(path.join(repositoryRoot, relativePath), "utf8"),
  );
}

async function fileSha256(relativePath) {
  const contents = await readFile(path.join(repositoryRoot, relativePath));
  return createHash("sha256").update(contents).digest("hex");
}

export async function collectFiles(relativePath, root = repositoryRoot) {
  const absolutePath = path.join(root, relativePath);
  const entries = await readdir(absolutePath, { withFileTypes: true }).catch(
    () => null,
  );
  if (entries === null) return [relativePath];

  const files = [];
  for (const entry of entries) {
    if (entry.isDirectory() && ignoredDirectoryNames.has(entry.name)) continue;
    const child = path.posix.join(relativePath, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await collectFiles(child, root)));
    } else if (includedExtensions.has(path.extname(entry.name))) {
      files.push(child);
    }
  }
  return files;
}

export async function fingerprint(paths, root = repositoryRoot) {
  const files = (
    await Promise.all(
      paths.map((relativePath) => collectFiles(relativePath, root)),
    )
  )
    .flat()
    .sort();
  const hash = createHash("sha256");
  for (const file of files) {
    const contents = await readFile(path.join(root, file));
    const fileHash = createHash("sha256").update(contents).digest("hex");
    hash.update(file);
    hash.update("\0");
    hash.update(fileHash);
    hash.update("\n");
  }
  return { fileCount: files.length, sha256: hash.digest("hex") };
}

export async function createManifest() {
  const rootPackage = await readJson("package.json");
  const packagePaths = {
    schema: "packages/schema/package.json",
    compiler: "packages/compiler/package.json",
    core: "packages/core/package.json",
    reactNative: "packages/react-native/package.json",
  };
  const packages = {};
  for (const [id, packagePath] of Object.entries(packagePaths)) {
    const packageJson = await readJson(packagePath);
    if (packageJson.version !== rootPackage.version) {
      throw new Error(
        `${packagePath} version ${packageJson.version} does not match release ${rootPackage.version}`,
      );
    }
    packages[id] = { name: packageJson.name, version: packageJson.version };
  }

  const artifacts = {};
  for (const [id, paths] of Object.entries(artifactRoots)) {
    artifacts[id] = await fingerprint(paths);
  }

  const antiSlopCatalog = await readJson(
    "deslint/internal/rules/anti_slop_catalog.json",
  );
  const antiSlopCatalogSha256 = await fileSha256(
    "deslint/internal/rules/anti_slop_catalog.json",
  );
  const antiSlopContract = await fingerprint([
    "packages/schema/design-system-policy.schema.json",
    "packages/schema/deslint-report.schema.json",
    "packages/schema/web-provider-evidence.schema.json",
    "packages/schema/native-source-evidence.schema.json",
    "packages/schema/native-runtime-evidence.schema.json",
    "packages/schema/layout-evidence.schema.json",
    "packages/schema/stage-execution-evidence.schema.json",
  ]);

  return {
    schemaVersion: 1,
    release: {
      version: rootPackage.version,
      tag: `v${rootPackage.version}`,
    },
    compatibility: {
      definitionSchemaVersion: 2,
      policySchemaVersion: 1,
      reportSchemaVersion: 1,
      consumerConformanceSchemaVersion: 1,
      designContextSchemaVersion: 1,
      visualDetailEvidenceSchemaVersion: 1,
      typographyEvidenceSchemaVersion: 1,
      colorEvidenceSchemaVersion: 1,
      layoutEvidenceSchemaVersion: 1,
      motionEvidenceSchemaVersion: 1,
      copyEvidenceSchemaVersion: 1,
      imageryEvidenceSchemaVersion: 1,
      runtimeEvidenceSchemaVersion: 1,
      nativeSourceEvidenceSchemaVersion: 1,
      nativeRuntimeEvidenceSchemaVersion: 1,
      webProviderEvidenceSchemaVersion: 1,
      consumerReleaseLockSchemaVersion: 1,
      stageExecutionEvidenceSchemaVersion: 1,
      compiledBundleVersion: 2,
      node: ">=22",
      go: ">=1.26.0",
      react: ">=19.0.0 <20",
      reactNative: ">=0.85.0 <0.86",
    },
    packages,
    deslint: {
      module: "github.com/aprilgom/AnslDes/deslint",
      version: rootPackage.version,
      releaseVersionLinkerFlag: `-X main.version=${rootPackage.version}`,
    },
    dependencies: {
      antiSlopCatalog: {
        sha256: antiSlopCatalogSha256,
        pack: antiSlopCatalog.pack,
      },
      impeccable: antiSlopCatalog.impeccable,
      hallmark: {
        commit: antiSlopCatalog.hallmark.commit,
        sourceSha256: antiSlopCatalog.hallmark.sourceSha256,
      },
      migrationNote: antiSlopCatalog.migrationNote,
      antiSlopContract,
    },
    artifacts,
  };
}

export function serialize(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

async function main() {
  const mode = process.argv[2];
  if (mode !== "--write" && mode !== "--check") {
    throw new Error("usage: release-manifest.mjs <--write|--check>");
  }
  const expected = serialize(await createManifest());
  if (mode === "--write") {
    await writeFile(manifestPath, expected, "utf8");
  } else {
    const actual = await readFile(manifestPath, "utf8");
    if (actual !== expected) {
      throw new Error(
        "release manifest is stale; run npm run release:manifest",
      );
    }
  }
  const manifestSha256 = createHash("sha256").update(expected).digest("hex");
  process.stdout.write(`AnslDes release manifest: PASS ${manifestSha256}\n`);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();
