#!/usr/bin/env node

/**
 * Auto-Translate Locales
 *
 * Finds all en keys from dict.ts, identifies missing translations
 * for each locale, and translates them via Google Translate's free API.
 *
 * Usage:
 *   node scripts/auto-translate-locales.mjs              # dry-run (shows what would be translated)
 *   node scripts/auto-translate-locales.mjs --apply       # actually translate and save
 *   node scripts/auto-translate-locales.mjs --lang fa,tr  # specific languages only
 *   node scripts/auto-translate-locales.mjs --limit 10    # translate only 10 keys per language
 *
 * After translation, run:
 *   node scripts/apply-i18n-locales.mjs   # merge locale JSON into dict.ts
 */

import { readFileSync, writeFileSync, existsSync, mkdirSync } from "fs";
import { join, dirname } from "path";
import { fileURLToPath } from "url";

// ─── Config ──────────────────────────────────────────────────

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = join(__dirname, "..");
const DICT_PATH = join(ROOT, "src/i18n/dict.ts");
const LOCALE_DIR = join(ROOT, "src/i18n/locale");

const LOCALES = ["fa", "tr", "ar", "ru", "zh", "ja", "es"];

const LOCALE_NAMES = {
  fa: "Persian",
  tr: "Turkish",
  ar: "Arabic",
  ru: "Russian",
  zh: "Chinese (Simplified)",
  ja: "Japanese",
  es: "Spanish",
};

const GLOSSARY = {
  "VortexUI": "VortexUI",
  "Xray": "Xray",
  "sing-box": "sing-box",
  "Clash": "Clash",
  "CDN": "CDN",
  "DPI": "DPI",
  "TLS": "TLS",
  "SSL": "SSL",
  "SNI": "SNI",
  "DNS": "DNS",
  "DoH": "DoH",
  "DoT": "DoT",
  "IP": "IP",
  "QR": "QR",
  "CSV": "CSV",
  "SSO": "SSO",
  "2FA": "2FA",
  "TOTP": "TOTP",
  "UUID": "UUID",
  "VPS": "VPS",
  "VMess": "VMess",
  "VLESS": "VLESS",
  "Trojan": "Trojan",
  "Shadowsocks": "Shadowsocks",
  "WebSocket": "WebSocket",
  "gRPC": "gRPC",
  "REALITY": "REALITY",
  "ZarinPal": "ZarinPal",
  "Telegram": "Telegram",
  "GitHub": "GitHub",
  "Cloudflare": "Cloudflare",
  "Hamrah Aval": "Hamrah Aval",
  "Irancell": "Irancell",
  "Shatel": "Shatel",
  "Asiatech": "Asiatech",
  "Mokhaberat": "Mokhaberat",
  "Backhaul": "Backhaul",
  "Rathole": "Rathole",
  "Wstunnel": "Wstunnel",
  "Kharej": "Kharej",
};

// ─── Parse dict.ts ───────────────────────────────────────────

function parseDict() {
  const content = readFileSync(DICT_PATH, "utf-8");

  // Extract en keys and values
  const enKeys = {};
  const enMatch = content.match(/const en = \{([\s\S]*?)\};/);
  if (!enMatch) throw new Error("Could not find 'const en = { ... };' block");

  const enBlock = enMatch[1];
  const keyRegex = /["']([a-zA-Z0-9_.-]+)["']\s*:\s*((["'`])(?:(?!\3)[\s\S])*\3)/g;
  let m;
  while ((m = keyRegex.exec(enBlock)) !== null) {
    const key = m[1];
    let val = m[2];
    // Remove surrounding quotes
    val = val.slice(1, -1);
    // Unescape JSON
    try { val = JSON.parse(`"${val.replace(/"/g, '\\"')}"`); } catch {}
    enKeys[key] = val;
  }

  // Extract locale keys for each language
  const localeKeys = {};
  for (const lang of LOCALES) {
    localeKeys[lang] = {};
    // Find the block: const lang: Record<string, string> = { ... };
    // The block may have different formatting; look for the key pattern within lang's section
    const blockRegex = new RegExp(`const ${lang}: Record<string, string> = \\{([\\s\\S]*?)\\};`);
    const blockMatch = content.match(blockRegex);
    if (!blockMatch) {
      console.warn(`  ⚠️  Block for '${lang}' not found in dict.ts`);
      continue;
    }
    const block = blockMatch[1];
    const lr = /["']([a-zA-Z0-9_.-]+)["']\s*:\s*((["'`])(?:(?!\3)[\s\S])*\3)/g;
    let lm;
    while ((lm = lr.exec(block)) !== null) {
      let val = lm[2].slice(1, -1);
      try { val = JSON.parse(`"${val.replace(/"/g, '\\"')}"`); } catch {}
      localeKeys[lang][lm[1]] = val;
    }
  }

  return { enKeys, localeKeys };
}

// ─── Template variable protection ──────────────────────────
// Protect {variable} placeholders from being translated
function protectVars(text) {
  const vars = [];
  let i = 0;
  const result = text.replace(/\{[a-zA-Z_][a-zA-Z0-9_]*\}/g, (match) => {
    const sentinel = `__V${i}__`;
    vars.push({ sentinel, original: match });
    i++;
    return sentinel;
  });
  return { text: result, vars };
}

function restoreVars(text, vars) {
  return vars.reduce((acc, { sentinel, original }) => {
    return acc.replace(sentinel, original);
  }, text);
}

// ─── Google Translate via free API (zero deps) ──────────────

async function translateText(rawText, targetLang) {
  const sourceLang = "en";

  // Protect template variables before translation
  const { text, vars } = protectVars(rawText);

  const url = `https://translate.googleapis.com/translate_a/single?client=gtx&sl=${sourceLang}&tl=${targetLang}&dt=t&q=${encodeURIComponent(text)}`;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 10000);

  try {
    const res = await fetch(url, { signal: controller.signal });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    // Response format: [[["translation","original",...],...],...]
    let translated = data[0]?.[0]?.[0] || text;
    // Restore template variables
    translated = restoreVars(translated, vars);
    return translated;
  } catch (err) {
    console.error(`    ❌ Translate failed for "${rawText.slice(0, 40)}...": ${err.message}`);
    return null;
  } finally {
    clearTimeout(timeout);
  }
}

// ─── Apply glossary replacements after translation ──────────

function applyGlossary(text) {
  let result = text;
  for (const [term, replacement] of Object.entries(GLOSSARY)) {
    // Case-insensitive replace, but keep the glossary term's casing
    const regex = new RegExp(term.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), "gi");
    result = result.replace(regex, replacement);
  }
  return result;
}

// ─── Generate locale JSON file ──────────────────────────────

function writeLocaleJson(lang, translations) {
  // Ensure locale directory exists
  if (!existsSync(LOCALE_DIR)) {
    mkdirSync(LOCALE_DIR, { recursive: true });
  }

  const filePath = join(LOCALE_DIR, `${lang}.json`);
  
  // Load existing translations if file exists
  let existing = {};
  if (existsSync(filePath)) {
    try {
      existing = JSON.parse(readFileSync(filePath, "utf-8"));
    } catch {}
  }

  // Merge: new translations override existing ones
  const merged = { ...existing, ...translations };

  // Sort keys
  const sorted = Object.keys(merged).sort().reduce((acc, key) => {
    acc[key] = merged[key];
    return acc;
  }, {});

  writeFileSync(filePath, JSON.stringify(sorted, null, 2) + "\n", "utf-8");
  return filePath;
}

// ─── Rate limiter — prevents 429 Too Many Requests ──────────

function createRateLimiter(requestsPerSecond = 3) {
  const queue = [];
  let running = false;

  async function processQueue() {
    if (running) return;
    running = true;
    try {
      while (queue.length > 0) {
        const { fn, resolve, reject } = queue.shift();
        try {
          const result = await fn();
          resolve(result);
        } catch (e) {
          reject(e);
        }
        // Wait between requests
        if (queue.length > 0) {
          await new Promise((r) => setTimeout(r, 1000 / requestsPerSecond));
        }
      }
    } finally {
      running = false;
    }
  }

  return function enqueue(fn) {
    return new Promise((resolve, reject) => {
      queue.push({ fn, resolve, reject });
      processQueue();
    });
  };
}

// ─── Main ────────────────────────────────────────────────────

async function main() {
  const args = process.argv.slice(2);
  const shouldApply = args.includes("--apply");
  const limitArg = args.find((a) => a.startsWith("--limit="));
  const limit = limitArg ? parseInt(limitArg.split("=")[1], 10) : Infinity;
  const langArg = args.find((a) => a.startsWith("--lang="));
  const targetLocales = langArg ? langArg.split("=")[1].split(",") : LOCALES;

  console.log("\n🌍 Auto-Translate Locales");
  console.log("═".repeat(60));

  const { enKeys, localeKeys } = parseDict();
  const totalEn = Object.keys(enKeys).length;
  console.log(`\n📊 English keys: ${totalEn}`);

  let totalTranslated = 0;
  let totalMissing = 0;

  // Rate limiter: 3 requests/sec to avoid 429
  const enqueue = createRateLimiter(3);

  for (const lang of targetLocales) {
    if (!LOCALES.includes(lang)) {
      console.warn(`  ⚠️  Unknown locale: ${lang}. Skipping.`);
      continue;
    }

    const localeName = LOCALE_NAMES[lang] || lang;
    const existing = localeKeys[lang] || {};
    const existingCount = Object.keys(existing).length;

    // Find missing keys
    const missing = Object.keys(enKeys).filter((key) => !(key in existing));
    totalMissing += missing.length;

    if (missing.length === 0) {
      console.log(`\n✅ ${localeName} (${lang}) — ${existingCount} keys, all up to date.`);
      continue;
    }

    const toTranslate = missing.slice(0, limit);
    console.log(`\n🌐 ${localeName} (${lang}) — ${existingCount} existing, ${missing.length} missing (translating ${toTranslate.length})`);

    if (!shouldApply) {
      // Dry-run: show sample keys
      for (const key of toTranslate.slice(0, 5)) {
        const val = enKeys[key];
        console.log(`  📝 ${key}: "${val.slice(0, 60)}${val.length > 60 ? "…" : ""}"`);
      }
      if (toTranslate.length > 5) {
        console.log(`  ... and ${toTranslate.length - 5} more`);
      }
      continue;
    }

    // Translate
    const translations = {};
    let count = 0;

    for (const key of toTranslate) {
      const value = enKeys[key];
      if (!value || value.length === 0) continue;

      const translated = await enqueue(() => translateText(value, lang));
      if (translated === null) continue;

      // Apply glossary fixes
      const fixed = applyGlossary(translated);
      translations[key] = fixed;
      count++;

      if (count % 5 === 0 || count === toTranslate.length) {
        process.stdout.write(`\r    Progress: ${count}/${toTranslate.length}`);
      }
    }
    console.log(); // newline

    if (Object.keys(translations).length > 0) {
      const filePath = writeLocaleJson(lang, translations);
      console.log(`  ✅ Wrote ${Object.keys(translations).length} translations to ${filePath.replace(ROOT + "/", "")}`);
      totalTranslated += Object.keys(translations).length;
    }
  }

  // Summary
  console.log("\n" + "═".repeat(60));
  if (shouldApply) {
    console.log(`✅ Translated ${totalTranslated} keys across ${targetLocales.length} locales.`);
    console.log(`📁 Locale files updated in ${LOCALE_DIR.replace(ROOT + "/", "")}/`);
    console.log(`\n👉 Run 'node scripts/apply-i18n-locales.mjs' to merge into dict.ts`);
  } else {
    console.log(`📊 Would translate approximately ${totalMissing} missing keys.`);
    console.log(`\n👉 Run with --apply to actually translate and save.`);
    console.log(`   Use --lang=fa,tr to limit to specific languages.`);
    console.log(`   Use --limit=10 to translate only 10 keys per language (test mode).`);
  }
  console.log();
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
