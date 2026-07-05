// One-off audit: verify every language block in src/i18n/dict.ts defines
// exactly the same set of keys as `en` (the source of truth). Run with:
//   node scripts/i18n-audit.mjs
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const __dirname = dirname(fileURLToPath(import.meta.url));
const filePath = join(__dirname, "..", "src", "i18n", "dict.ts");
const text = readFileSync(filePath, "utf8");
const lines = text.split("\n");

const blockStart = /^const (\w+)(?::[^=]+)? = \{$/;
const blocks = [];
let current = null;
for (let i = 0; i < lines.length; i++) {
  const m = lines[i].match(blockStart);
  if (m) {
    if (current) current.end = i;
    current = { lang: m[1], start: i + 1 };
    blocks.push(current);
    continue;
  }
  if (current && lines[i] === "};") {
    current.end = i;
    current = null;
  }
}

const keyLine = /^\s*"([^"]+)":/;
const keysByLang = {};
for (const b of blocks) {
  const keys = new Set();
  for (let i = b.start; i < b.end; i++) {
    const m = lines[i].match(keyLine);
    if (m) keys.add(m[1]);
  }
  keysByLang[b.lang] = keys;
}

const langs = Object.keys(keysByLang);
console.log("Languages found:", langs.join(", "));
for (const l of langs) console.log(`  ${l}: ${keysByLang[l].size} keys`);

const source = keysByLang["en"];
let problems = 0;
for (const lang of langs) {
  if (lang === "en") continue;
  const keys = keysByLang[lang];
  const missing = [...source].filter((k) => !keys.has(k));
  const extra = [...keys].filter((k) => !source.has(k));
  if (missing.length || extra.length) {
    problems++;
    console.log(`\n=== ${lang} ===`);
    if (missing.length) {
      console.log(`  MISSING (${missing.length}):`);
      for (const k of missing) console.log(`    ${k}`);
    }
    if (extra.length) {
      console.log(`  EXTRA (${extra.length}):`);
      for (const k of extra) console.log(`    ${k}`);
    }
  }
}

// Also check for duplicate keys within a single block (last one silently wins).
for (const b of blocks) {
  const seen = new Map();
  for (let i = b.start; i < b.end; i++) {
    const m = lines[i].match(keyLine);
    if (!m) continue;
    const key = m[1];
    if (seen.has(key)) {
      problems++;
      console.log(`\nDUPLICATE key "${key}" in ${b.lang} at lines ${seen.get(key) + 1} and ${i + 1}`);
    }
    seen.set(key, i);
  }
}

if (problems === 0) {
  console.log("\nAll language blocks are key-complete and consistent with `en`. No duplicates found.");
  process.exit(0);
} else {
  console.log(`\n${problems} block(s) had issues.`);
  process.exit(1);
}
