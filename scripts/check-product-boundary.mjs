import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const genericRoots = ["packages", "docs", "deslint"];
const genericFiles = ["README.md", "AGENTS.md", "TODO.md"];
const skippedDirectories = new Set(["bin", "dist", "node_modules"]);
const allowedTopLevelDirectories = new Set([
  ".git",
  ".github",
  "deslint",
  "docs",
  "node_modules",
  "packages",
  "release",
  "scripts",
]);
const forbidden = [
  {
    label: "consumer workspace path",
    pattern: /\/Users\/|[A-Za-z]:\\Users\\/u,
  },
  {
    label: "embedded product snapshot",
    pattern: /(?:^|[/"'`])legacy\/(?:snapshot|product)(?:[/"'`]|$)/mu,
  },
];

async function collectFiles(root) {
  const entries = await readdir(root, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    if (entry.isDirectory() && skippedDirectories.has(entry.name)) continue;
    const target = path.join(root, entry.name);
    if (entry.isDirectory()) files.push(...(await collectFiles(target)));
    else files.push(target);
  }
  return files;
}

const violations = [];
for (const entry of await readdir(".", { withFileTypes: true })) {
  if (entry.isDirectory() && !allowedTopLevelDirectories.has(entry.name)) {
    violations.push(`${entry.name}: unapproved top-level source tree`);
  }
}
for (const file of genericFiles) {
  const contents = await readFile(file, "utf8");
  for (const rule of forbidden) {
    if (rule.pattern.test(contents))
      violations.push(`${file}: forbidden ${rule.label}`);
  }
}
for (const root of genericRoots) {
  for (const file of await collectFiles(root)) {
    if (file.split(path.sep).includes("legacy")) {
      violations.push(`${file}: embedded product snapshot directory`);
    }
    const contents = await readFile(file, "utf8");
    for (const rule of forbidden) {
      if (rule.pattern.test(contents))
        violations.push(`${file}: forbidden ${rule.label}`);
    }
  }
}

if (violations.length > 0) {
  process.stderr.write(`${violations.sort().join("\n")}\n`);
  process.exitCode = 1;
} else {
  process.stdout.write("product boundary: PASS\n");
}
