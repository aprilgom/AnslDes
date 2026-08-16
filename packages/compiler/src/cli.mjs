#!/usr/bin/env node

import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

import {
  canonicalStringify,
  compileDesignSystem,
  readDesignSystemDefinition,
} from "./index.mjs";

async function main(args) {
  if (args.length !== 4 || args[0] !== "compile" || args[2] !== "--out") {
    throw new Error(
      "usage: ansldes-compile compile <definition.json> --out <bundle.json>",
    );
  }
  const [, input, , output] = args;
  const definition = await readDesignSystemDefinition(input);
  const bundle = compileDesignSystem(definition);
  await mkdir(path.dirname(output), { recursive: true });
  await writeFile(output, canonicalStringify(bundle), "utf8");
  process.stdout.write(
    `compiled ${bundle.definition.id}@${bundle.definition.version} ${bundle.fingerprintSha256}\n`,
  );
}

main(process.argv.slice(2)).catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
});
