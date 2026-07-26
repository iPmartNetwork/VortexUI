#!/usr/bin/env node

/**
 * Locale Validation Script — zero external dependencies
 *
 * Validates that all locale JSON files in src/i18n/locale/ are
 * consistent with the English source keys from dict.ts.
 *
 * Checks:
 *   1. Missing keys — keys present in `en` but missing from locale
 *   2. Extra keys — keys present in locale but missing from `en`
 *   3. Template variable consistency — {name}, {count}, etc.
 *      must be preserved exactly in all translations
 *
 * Usage:
 *   node scripts/validate-locales.mjs         # normal report
 *   node scripts/validate-locales.mjs --json  # JSON output (for CI)
 *   node scripts/validate-locales.mjs --strict  # fail on any issue
 */

import { readFileSync } from "fs";
import { join, dirname } from "path";
import { fileURLToPath } from "url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = join(__dirname, "..");
const DICT_PATH = join(ROOT, "src/i18n/dict.ts");

const LOCALES = ["fa", "tr", "ar", "ru", "zh", "ja", "es"];

// ─── Parse dict.ts to get en keys + values ───────────────────

function parseEnglishKeys() {
  const content = readFileSync(DICT_PATH, "utf-8");

  const enMatch = content.match(/const en = \{([\s\S]*?)\};/);
  if (!enMatch) throw new Error("Could not find 'const en = { ... };' block");

  const block = enMatch[1];
  const keys = {};
  const regex = /["']([a-zA-Z0-9_.-]+)["']\s*:\s*((["'`])(?:(?!\3)[\s\S])*\3)/g;
  let m;
  while ((m = regex.exec(block)) !== null) {
    let val = m[2].slice(1, -1);
    try { val = JSON.parse(`"${val.replace(/"/g, '\\"')}"`); } catch {}
    keys[m[1]] = val;
  }

  return keys;
}

// ─── Parse locale keys from dict.ts blocks ───────────────────
// Locale data lives in dict.ts as `const lang: Record<string, string> = { ... }`
// The locale/*.json files are supplements merged by apply-i18n-locales.mjs

function parseLocaleKeys(lang) {
  const content = readFileSync(DICT_PATH, "utf-8");
  const blockRegex = new RegExp(`const ${lang}: Record<string, string> = \\{([\\s\\S]*?)\\};`);
  const match = content.match(blockRegex);
  if (!match) return null;

  const block = match[1];
  const keys = {};
  const regex = /["']([a-zA-Z0-9_.-]+)["']\s*:\s*((["'`])(?:(?!\3)[\s\S])*\3)/g;
  let m;
  while ((m = regex.exec(block)) !== null) {
    let val = m[2].slice(1, -1);
    try { val = JSON.parse(`"${val.replace(/"/g, '\\"')}"`); } catch {}
    keys[m[1]] = val;
  }
  return keys;
}

// ─── Extract template variables from a string ────────────────

function extractTemplateVars(str) {
  const vars = [];
  const regex = /\{([a-zA-Z_][a-zA-Z0-9_]*)\}/g;
  let m;
  while ((m = regex.exec(str)) !== null) {
    vars.push(m[1]);
  }
  return vars.sort();
}

// ─── Main ────────────────────────────────────────────────────

function main() {
  const args = process.argv.slice(2);
  const outputJson = args.includes("--json");
  const strict = args.includes("--strict");

  const enKeys = parseEnglishKeys();
  const enVarMap = {};
  for (const [key, value] of Object.entries(enKeys)) {
    enVarMap[key] = extractTemplateVars(value);
  }

  let totalIssues = 0;
  let hasErrors = false;
  const report = {};

  console.log(
    outputJson
      ? ""
      : "\n✅ Locale Validation\n" + "═".repeat(60),
  );

  for (const lang of LOCALES) {
    const localeKeys = parseLocaleKeys(lang);
    if (lang === "en") continue; // en is the source
    if (!localeKeys) {
      if (!outputJson) console.log(`\n❌ ${lang.toUpperCase()} — locale file not found`);
      totalIssues++;
      continue;
    }

    const localeVars = new Set(Object.keys(localeKeys));
    const enKeySet = new Set(Object.keys(enKeys));
    const localeKeySet = new Set(Object.keys(localeKeys));
    const issues = [];

    // ── Check 1: Missing keys ──
    const missing = Object.keys(enKeys).filter((k) => !localeKeySet.has(k));
    for (const key of missing) {
      issues.push({ type: "missing", key, severity: "error" });
    }

    // ── Check 2: Extra keys ──
    const extra = Object.keys(localeKeys).filter((k) => !enKeySet.has(k));
    for (const key of extra) {
      issues.push({ type: "extra", key, severity: "warning" });
    }

    // ── Check 3: Template variable mismatch ──
    for (const key of Object.keys(localeKeys)) {
      if (!(key in enKeys)) continue;
      const enVars = enVarMap[key] || [];
      if (enVars.length === 0) continue;

      const localeVal = localeKeys[key];
      const localeVars = extractTemplateVars(localeVal);

      // Check for missing variables
      const missingVars = enVars.filter((v) => !localeVars.includes(v));
      for (const v of missingVars) {
        issues.push({
          type: "template_missing",
          key,
          detail: `{${v}}`,
          severity: "error",
        });
      }

      // Check for extra variables
      const extraVars = localeVars.filter((v) => !enVars.includes(v));
      for (const v of extraVars) {
        issues.push({
          type: "template_extra",
          key,
          detail: `{${v}}`,
          severity: "warning",
        });
      }
    }

    const errorCount = issues.filter((i) => i.severity === "error").length;
    const warningCount = issues.filter((i) => i.severity === "warning").length;
    totalIssues += issues.length;
    if (errorCount > 0) hasErrors = true;

    report[lang] = {
      total: Object.keys(localeKeys).length,
      missing: missing.length,
      extra: extra.length,
      templateIssues: issues.filter((i) => i.type.startsWith("template_")).length,
      errors: errorCount,
      warnings: warningCount,
      issues,
    };

    if (!outputJson) {
      const status = errorCount === 0 ? "✅" : warningCount > 0 ? "⚠️" : "❌";
      console.log(`\n${status} ${lang.toUpperCase()} — ${Object.keys(localeKeys).length} keys`);
      if (missing.length > 0) console.log(`   ❌ Missing keys:     ${missing.length}`);
      if (extra.length > 0) console.log(`   ⚠️ Extra keys:       ${extra.length}`);
      const tIssues = issues.filter((i) => i.type.startsWith("template_"));
      if (tIssues.length > 0) console.log(`   ⚠️ Template issues:  ${tIssues.length}`);

      if (issues.length > 0 && issues.length <= 15) {
        for (const issue of issues) {
          const icon = issue.severity === "error" ? "❌" : "⚠️";
          const detail = issue.detail ? ` (${issue.detail})` : "";
          console.log(`     ${icon} ${issue.type}: ${issue.key}${detail}`);
        }
      } else if (issues.length > 15) {
        // Show top issues
        const errors = issues.filter((i) => i.severity === "error");
        const warnings = issues.filter((i) => i.severity === "warning");
        for (const issue of errors.slice(0, 5)) {
          const detail = issue.detail ? ` (${issue.detail})` : "";
          console.log(`     ❌ ${issue.type}: ${issue.key}${detail}`);
        }
        if (errors.length > 5) console.log(`     ... and ${errors.length - 5} more errors`);
        for (const issue of warnings.slice(0, 3)) {
          const detail = issue.detail ? ` (${issue.detail})` : "";
          console.log(`     ⚠️ ${issue.type}: ${issue.key}${detail}`);
        }
        if (warnings.length > 3) console.log(`     ... and ${warnings.length - 3} more warnings`);
      }
    }
  }

  // ── Summary ──
  if (!outputJson) {
    const totalLocales = LOCALES.length;
    const okLocales = Object.values(report).filter((r) => r.errors === 0).length;

    console.log("\n" + "═".repeat(60));
    console.log(`📊 Summary:`);
    console.log(`   Languages:      ${okLocales}/${totalLocales} passed`);
    console.log(`   Total issues:   ${totalIssues}`);
    console.log(`   Has errors:     ${hasErrors ? "❌ Yes" : "✅ No"}`);

    if (hasErrors) {
      console.log(`\n❌ Validation FAILED — fix errors above and re-run.`);
      if (strict) process.exit(1);
    } else {
      console.log(`\n✅ All locales valid!`);
    }
  } else {
    // JSON output
    const summary = {
      passed: !hasErrors,
      totalIssues,
      locales: report,
      timestamp: new Date().toISOString(),
    };
    console.log(JSON.stringify(summary, null, 2));
    if (strict && hasErrors) process.exit(1);
  }
}

main();
