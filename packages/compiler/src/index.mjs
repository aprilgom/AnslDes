import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";

import Ajv2020 from "ajv/dist/2020.js";

const schema = JSON.parse(
  await readFile(
    new URL(
      "../../schema/design-system-definition.schema.json",
      import.meta.url,
    ),
    "utf8",
  ),
);
const ajv = new Ajv2020({ allErrors: true, strict: true });
const validateSchema = ajv.compile(schema);
const referencePattern = /^\{([a-z]+)\.([a-z]+)\.([a-z0-9][a-z0-9._-]*)\}$/u;

export class DesignSystemCompileError extends Error {
  constructor(diagnostics) {
    super(diagnostics.join("\n"));
    this.name = "DesignSystemCompileError";
    this.diagnostics = diagnostics;
  }
}

export async function readDesignSystemDefinition(path) {
  return JSON.parse(await readFile(path, "utf8"));
}

export function canonicalize(value) {
  if (Array.isArray(value)) return value.map(canonicalize);
  if (!isRecord(value)) return value;
  return Object.fromEntries(
    Object.keys(value)
      .sort()
      .map((key) => [key, canonicalize(value[key])]),
  );
}

export function canonicalStringify(value) {
  return `${JSON.stringify(canonicalize(value), null, 2)}\n`;
}

export function fingerprint(value) {
  return createHash("sha256").update(canonicalStringify(value)).digest("hex");
}

export function compileDesignSystem(definition) {
  validateDefinition(definition);
  const themes = Object.fromEntries(
    [...definition.themes.names]
      .sort()
      .map((theme) => [theme, compileTheme(definition, theme)]),
  );
  const content = {
    version: 2,
    definition: {
      schemaVersion: definition.schemaVersion,
      id: definition.id,
      version: definition.version,
      defaultTheme: definition.themes.default,
    },
    themes,
  };
  return {
    ...content,
    fingerprintSha256: fingerprint(content),
  };
}

export function validateDefinition(definition) {
  const diagnostics = [];
  if (!validateSchema(definition)) {
    diagnostics.push(
      ...validateSchema.errors.map(
        (error) => `schema${error.instancePath || "/"}: ${error.message}`,
      ),
    );
  }
  if (!isRecord(definition)) throwCompile(diagnostics);

  validateThemes(definition, diagnostics);
  validateTypography(definition, diagnostics);
  validateIcons(definition, diagnostics);
  validateFoundationReferenceKinds(definition, diagnostics);
  validateComponentRecipeReferenceKinds(definition, diagnostics);
  throwCompile(diagnostics);
}

function compileTheme(definition, theme) {
  const resolver = createResolver(definition, theme);
  const foundations = definition.foundations;
  const compileNumeric = (name) => ({
    primitive: canonicalize(foundations[name].primitive),
    semantic: resolveMap(foundations[name].semantic, resolver),
    component: resolveMap(foundations[name].component, resolver),
  });
  const typography = {
    familyName: foundations.typography.familyName,
    weights: canonicalize(foundations.typography.weights),
    tokens: canonicalize(foundations.typography.tokens),
    roles: Object.fromEntries(
      Object.keys(foundations.typography.roles)
        .sort()
        .map((name) => [name, resolveTypographyRole(definition, name)]),
    ),
  };

  return {
    foundations: {
      color: {
        primitive: canonicalize(foundations.color.primitive),
        semantic: Object.fromEntries(
          Object.keys(foundations.color.semantic)
            .sort()
            .map((name) => [
              name,
              resolver.resolve(foundations.color.semantic[name][theme]),
            ]),
        ),
        component: resolveMap(foundations.color.component, resolver),
        asset: canonicalize(foundations.color.asset),
      },
      spacing: compileNumeric("spacing"),
      radius: compileNumeric("radius"),
      size: compileNumeric("size"),
      typography,
      motion: {
        durations: canonicalize(foundations.motion.durations),
        easings: canonicalize(foundations.motion.easings),
        transitions: resolver.resolve(foundations.motion.transitions),
      },
      elevation: resolver.resolve(foundations.elevation),
      layer: canonicalize(foundations.layer),
      ...(foundations.icon ? { icon: resolver.resolve(foundations.icon) } : {}),
    },
    components: resolver.resolve(definition.components),
  };
}

function createResolver(definition, theme) {
  const active = [];
  const cache = new Map();

  function resolve(value) {
    if (typeof value === "string") {
      const match = referencePattern.exec(value);
      return match ? resolveReference(value, match) : value;
    }
    if (Array.isArray(value)) return value.map(resolve);
    if (!isRecord(value)) return value;
    return Object.fromEntries(
      Object.keys(value)
        .sort()
        .map((key) => [key, resolve(value[key])]),
    );
  }

  function resolveReference(reference, match) {
    if (cache.has(reference)) return cache.get(reference);
    if (active.includes(reference)) {
      const cycle = [...active.slice(active.indexOf(reference)), reference];
      throw new DesignSystemCompileError([
        `reference cycle: ${cycle.join(" -> ")}`,
      ]);
    }
    active.push(reference);
    const [, collection, layer, name] = match;
    const target = lookupReference(definition, theme, collection, layer, name);
    const result = resolve(target);
    active.pop();
    cache.set(reference, result);
    return result;
  }

  return { resolve };
}

function lookupReference(definition, theme, collection, layer, name) {
  const foundations = definition.foundations;
  let target;
  if (["color", "spacing", "radius", "size"].includes(collection)) {
    target = foundations[collection]?.[layer]?.[name];
    if (
      collection === "color" &&
      layer === "semantic" &&
      target !== undefined
    ) {
      target = target[theme];
    }
  } else if (collection === "typography") {
    if (layer === "role") return resolveTypographyRole(definition, name);
    if (layer === "token") target = foundations.typography.tokens[name];
    if (layer === "weight") target = foundations.typography.weights[name];
  } else if (collection === "motion") {
    const motionLayers = {
      duration: "durations",
      easing: "easings",
      transition: "transitions",
    };
    target = foundations.motion[motionLayers[layer]]?.[name];
  } else if (collection === "elevation" && layer === "recipe") {
    target = foundations.elevation[name];
  } else if (collection === "layer") {
    target = foundations.layer[layer]?.[name];
  }
  if (target === undefined) {
    throw new DesignSystemCompileError([
      `unknown reference {${collection}.${layer}.${name}} for theme ${theme}`,
    ]);
  }
  return target;
}

function resolveTypographyRole(definition, name) {
  const typography = definition.foundations.typography;
  const role = typography.roles[name];
  if (!role)
    throw new DesignSystemCompileError([`unknown typography role ${name}`]);
  const token = typography.tokens[role.visualToken];
  const weight = typography.weights[role.weight];
  if (!token || !weight) {
    throw new DesignSystemCompileError([`invalid typography role ${name}`]);
  }
  return {
    semanticRole: name,
    visualToken: role.visualToken,
    logicalWeight: role.weight,
    fontFamily: weight.fontFamily,
    fontWeight: weight.fontWeight,
    fontSize: token.fontSize,
    lineHeight: token.lineHeight,
    maxFontSizeMultiplier: token.maxFontSizeMultiplier,
  };
}

function resolveMap(values, resolver) {
  return Object.fromEntries(
    Object.keys(values)
      .sort()
      .map((name) => [name, resolver.resolve(values[name])]),
  );
}

function validateThemes(definition, diagnostics) {
  const names = definition.themes?.names ?? [];
  if (!names.includes(definition.themes?.default)) {
    diagnostics.push("themes.default must be present in themes.names");
  }
  const semantic = definition.foundations?.color?.semantic ?? {};
  for (const [name, mapping] of Object.entries(semantic)) {
    const actual = Object.keys(mapping).sort();
    const expected = [...names].sort();
    if (JSON.stringify(actual) !== JSON.stringify(expected)) {
      diagnostics.push(
        `color semantic ${name} themes must be exact: ${expected.join(", ")}`,
      );
    }
  }
}

function validateTypography(definition, diagnostics) {
  const typography = definition.foundations?.typography;
  if (!typography) return;
  for (const [tokenName, token] of Object.entries(typography.tokens ?? {})) {
    for (const weight of token.allowedWeights ?? []) {
      if (!(weight in typography.weights)) {
        diagnostics.push(
          `typography token ${tokenName} has unknown weight ${weight}`,
        );
      }
    }
  }
  for (const [roleName, role] of Object.entries(typography.roles ?? {})) {
    const token = typography.tokens[role.visualToken];
    if (!token)
      diagnostics.push(
        `typography role ${roleName} has unknown token ${role.visualToken}`,
      );
    if (!(role.weight in typography.weights)) {
      diagnostics.push(
        `typography role ${roleName} has unknown weight ${role.weight}`,
      );
    } else if (token && !token.allowedWeights.includes(role.weight)) {
      diagnostics.push(
        `typography role ${roleName} disallows weight ${role.weight}`,
      );
    }
  }
}

function validateIcons(definition, diagnostics) {
  const icon = definition.foundations?.icon;
  if (!icon) return;
  const sizes = icon.sizes ?? {};
  const strokes = icon.strokes ?? {};
  const alignments = icon.opticalAlignments ?? {};
  const icons = icon.icons ?? {};

  for (const [name, recipe] of Object.entries(icons)) {
    if (!(recipe.defaultSize in sizes)) {
      diagnostics.push(
        `icon ${name} has unknown default size ${recipe.defaultSize}`,
      );
    }
    if (!recipe.allowedSizes?.includes(recipe.defaultSize)) {
      diagnostics.push(`icon ${name} default size must be allowed`);
    }
    for (const size of recipe.allowedSizes ?? []) {
      if (!(size in sizes)) {
        diagnostics.push(`icon ${name} has unknown allowed size ${size}`);
      }
    }
    if (!(recipe.opticalAlignment in alignments)) {
      diagnostics.push(
        `icon ${name} has unknown optical alignment ${recipe.opticalAlignment}`,
      );
    }
    for (const part of recipe.geometry ?? []) {
      if (
        typeof part.strokeWidth === "string" &&
        !(part.strokeWidth in strokes)
      ) {
        diagnostics.push(`icon ${name} has unknown stroke ${part.strokeWidth}`);
      }
    }
  }

  for (const [kind, recipes] of [
    ["usage", icon.usages ?? {}],
    ["action", icon.actions ?? {}],
  ]) {
    for (const [name, recipe] of Object.entries(recipes)) {
      if (typeof recipe.icon !== "string" || !(recipe.icon in icons)) {
        diagnostics.push(`${kind} ${name} has unknown icon ${recipe.icon}`);
      }
      if (typeof recipe.size !== "string" || !(recipe.size in sizes)) {
        diagnostics.push(`${kind} ${name} has unknown size ${recipe.size}`);
      }
    }
  }
}

function validateFoundationReferenceKinds(definition, diagnostics) {
  const foundations = definition.foundations ?? {};
  validateReferenceMap(
    foundations.color?.semantic,
    ["color.primitive"],
    diagnostics,
    true,
  );
  validateReferenceMap(
    foundations.color?.component,
    ["color.semantic"],
    diagnostics,
  );
  for (const collection of ["spacing", "radius", "size"]) {
    validateReferenceMap(
      foundations[collection]?.semantic,
      [`${collection}.primitive`],
      diagnostics,
    );
    validateReferenceMap(
      foundations[collection]?.component,
      [`${collection}.semantic`],
      diagnostics,
    );
  }
}

function validateComponentRecipeReferenceKinds(definition, diagnostics) {
  visitRecipeValues(
    definition.components ?? {},
    "components",
    (value, path) => {
      if (typeof value !== "string") return;
      const match = referencePattern.exec(value);
      if (!match || !["color", "spacing", "radius", "size"].includes(match[1]))
        return;
      if (match[2] !== "component") {
        diagnostics.push(
          `${path} must reference ${match[1]}.component instead of ${match[1]}.${match[2]}`,
        );
      }
    },
  );
}

function visitRecipeValues(value, path, visitor) {
  visitor(value, path);
  if (Array.isArray(value)) {
    value.forEach((item, index) => {
      visitRecipeValues(item, `${path}[${index}]`, visitor);
    });
    return;
  }
  if (!isRecord(value)) return;
  for (const [name, item] of Object.entries(value)) {
    visitRecipeValues(item, `${path}.${name}`, visitor);
  }
}

function validateReferenceMap(
  values,
  allowedPrefixes,
  diagnostics,
  themed = false,
) {
  for (const [name, raw] of Object.entries(values ?? {})) {
    const candidates = themed ? Object.values(raw) : [raw];
    for (const value of candidates) {
      const match = referencePattern.exec(value);
      const prefix = match ? `${match[1]}.${match[2]}` : "";
      if (!match || !allowedPrefixes.includes(prefix)) {
        diagnostics.push(
          `${name} must reference ${allowedPrefixes.join(" or ")}`,
        );
      }
    }
  }
}

function throwCompile(diagnostics) {
  if (diagnostics.length > 0) {
    throw new DesignSystemCompileError([...new Set(diagnostics)].sort());
  }
}

function isRecord(value) {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
