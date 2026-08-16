import { createHash } from "node:crypto";
import { readdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";

const directory = process.argv[2];
const outputName = process.argv[3];
if (!directory || !outputName) {
  throw new Error("usage: hash-release-assets.mjs <directory> <output-name>");
}

const files = (await readdir(directory, { withFileTypes: true }))
  .filter((entry) => entry.isFile() && entry.name !== outputName)
  .map((entry) => entry.name)
  .sort();
if (files.length === 0) throw new Error("release asset directory is empty");

const lines = [];
for (const file of files) {
  const contents = await readFile(path.join(directory, file));
  const hash = createHash("sha256").update(contents).digest("hex");
  lines.push(`${hash}  ${file}`);
}
await writeFile(
  path.join(directory, outputName),
  `${lines.join("\n")}\n`,
  "utf8",
);
