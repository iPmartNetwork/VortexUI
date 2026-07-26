#!/usr/bin/env node

/**
 * i18n Audit Script — zero external dependencies
 *
 * 1. Recursively scans all .tsx files in src/ for t("...") calls
 * 2. Extracts defined keys from src/i18n/dict.ts (the `en` object)
 * 3. Reports keys used in pages but missing from dict
 * 4. With --fix, auto-generates missing entries into dict.ts
 */

import { readFileSync, writeFileSync, statSync, readdirSync } from "fs";
import { join, extname, relative } from "path";
import { fileURLToPath } from "url";

const DIR = join(fileURLToPath(import.meta.url), "../..");
const SRC_DIR = join(DIR, "src");
const DICT_PATH = join(DIR, "src/i18n/dict.ts");

// ─── Zero-dep recursive file finder ─────────────────────────

function findTsxFiles(dir, results = []) {
  const entries = readdirSync(dir, { withFileTypes: true });
  for (const entry of entries) {
    const full = join(dir, entry.name);
    if (entry.isDirectory() && !entry.name.startsWith(".") && entry.name !== "node_modules") {
      findTsxFiles(full, results);
    } else if (entry.isFile() && extname(entry.name) === ".tsx") {
      results.push(full);
    }
  }
  return results;
}

// ─── Step 1: Extract t() keys from .tsx files ───────────────

function extractUsedKeys() {
  const files = findTsxFiles(SRC_DIR);
  const usedKeys = new Set();
  const keyFileMap = {};

  for (const file of files) {
    const content = readFileSync(file, "utf-8");
    // Match t("..."), t('...'), t(`...`)
    const regex = /t\(["'`]([a-zA-Z0-9_.-]+)["'`]\)/g;
    let match;
    while ((match = regex.exec(content)) !== null) {
      const key = match[1];
      usedKeys.add(key);
      if (!keyFileMap[key]) keyFileMap[key] = [];
      const relPath = relative(DIR, file).replace(/\\/g, "/");
      if (!keyFileMap[key].includes(relPath)) keyFileMap[key].push(relPath);
    }
  }

  return { usedKeys: [...usedKeys].sort(), keyFileMap };
}

// ─── Step 2: Extract defined keys from dict.ts ──────────────

function extractDefinedKeys() {
  const content = readFileSync(DICT_PATH, "utf-8");
  const definedKeys = new Set();
  // Match dot-notation keys inside the `en` object and the type union
  // e.g. "app.tagline": "..." or `app.tagline`: "..."
  const regex = /["'`]([a-zA-Z0-9_.-]+)["'`]\s*:\s*["'`]/g;
  let match;
  while ((match = regex.exec(content)) !== null) {
    definedKeys.add(match[1]);
  }
  return [...definedKeys].sort();
}

// ─── Step 3: Generate missing key entries ───────────────────

function generateMissingEntries(missingKeys) {
  const lines = [];
  for (const key of missingKeys) {
    const parts = key.split(".");
    // Generate a readable English default from the last segment
    const last = parts[parts.length - 1]
      .replace(/-/g, " ")
      .replace(/([A-Z])/g, " $1")
      .trim();
    const defaultValue = last.charAt(0).toUpperCase() + last.slice(1);
    lines.push(`  "${key}": "${defaultValue}",`);
  }
  return lines;
}

// ─── Step 4: Insert missing keys into dict.ts ───────────────

function insertIntoDict(missingKeys) {
  const content = readFileSync(DICT_PATH, "utf-8");
  const entries = generateMissingEntries(missingKeys);
  if (entries.length === 0) return;

  // Find the position of the closing `};` of the `en` object
  // Strategy: find `const en = {` and track brace depth
  const enStart = content.indexOf("const en = {");
  if (enStart === -1) {
    console.error("  ❌ Could not find 'const en = {' in dict.ts");
    return;
  }

  let depth = 0;
  let insertPos = -1;
  for (let i = enStart; i < content.length; i++) {
    if (content[i] === "{") depth++;
    else if (content[i] === "}") {
      depth--;
      if (depth === 0) {
        // Found the closing brace of `en` — insert before it
        insertPos = i;
        break;
      }
    }
  }

  if (insertPos === -1) {
    console.error("  ❌ Could not find closing brace of 'en' object");
    return;
  }

  const insertText = `\n  /* i18n auto-fill — ${new Date().toISOString().slice(0, 10)} */\n${entries.join("\n")}\n`;
  const newContent =
    content.slice(0, insertPos) + insertText + content.slice(insertPos);
  writeFileSync(DICT_PATH, newContent, "utf-8");
  console.log(`  ✅ Inserted ${entries.length} missing keys into src/i18n/dict.ts`);
}

// ─── Main ────────────────────────────────────────────────────

function main() {
  const args = process.argv.slice(2);
  const autoFix = args.includes("--fix");

  console.log("\n🔍 i18n Audit");
  console.log("═".repeat(60));

  console.log(`\nScanning ${SRC_DIR} for t() usage...`);

  const { usedKeys, keyFileMap } = extractUsedKeys();
  const definedKeys = extractDefinedKeys();

  const usedSet = new Set(usedKeys);
  const definedSet = new Set(definedKeys);
  const missing = usedKeys.filter((k) => !definedSet.has(k));
  const unused = definedKeys.filter((k) => !usedSet.has(k));

  console.log(`\n📊 Summary:`);
  console.log(`  Keys used in pages:   ${usedKeys.length}`);
  console.log(`  Keys defined in dict: ${definedKeys.length}`);
  console.log(`  Missing keys:         ${missing.length}`);
  console.log(`  Unused keys:          ${unused.length}`);

  // Report missing with file locations
  if (missing.length > 0) {
    console.log(`\n❌ Missing keys (used in pages but NOT in dict):`);
    for (const key of missing) {
      const files = keyFileMap[key] || [];
      console.log(`  • ${key}`);
      for (const f of files.slice(0, 3)) {
        console.log(`    └─ ${f}`);
      }
      if (files.length > 3) console.log(`    └─ ... and ${files.length - 3} more`);
    }
  }

  // Report unused
  if (unused.length > 0) {
    console.log(`\n🟡 Unused keys (defined in dict but never used in pages):`);
    for (const key of unused.slice(0, 30)) {
      console.log(`  • ${key}`);
    }
    if (unused.length > 30) console.log(`  ... and ${unused.length - 30} more`);
  }

  // Auto-fix
  if (autoFix && missing.length > 0) {
    console.log(`\n🔧 Auto-fixing ${missing.length} missing keys...`);
    insertIntoDict(missing);
  }

  console.log("\n" + "═".repeat(60));
  console.log(`📁 ${usedKeys.length} keys used • ${definedKeys.length} defined • ${missing.length} missing • ${unused.length} unused`);
}

main();
