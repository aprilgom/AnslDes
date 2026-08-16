import { readFile } from "node:fs/promises";

const tag = process.argv[2];
if (!tag) throw new Error("usage: check-release-tag.mjs <tag>");

const manifest = JSON.parse(
  await readFile(
    new URL("../release/ansldes-release.json", import.meta.url),
    "utf8",
  ),
);
const packageJson = JSON.parse(
  await readFile(new URL("../package.json", import.meta.url), "utf8"),
);

if (manifest.release?.tag !== tag) {
  throw new Error(
    `release tag ${tag} does not match manifest ${manifest.release?.tag}`,
  );
}
if (manifest.release?.version !== packageJson.version) {
  throw new Error("release manifest and root package version differ");
}
process.stdout.write(`AnslDes release tag: PASS ${tag}\n`);
