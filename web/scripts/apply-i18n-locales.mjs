/**
 * Merges locale/*.json supplements into dict.ts for each language block.
 * Run: node scripts/apply-i18n-locales.mjs
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dictPath = path.join(__dirname, "../src/i18n/dict.ts");
const localeDir = path.join(__dirname, "../src/i18n/locale");

const langs = ["fa", "tr", "ar", "ru", "zh", "ja", "es"];
let dict = fs.readFileSync(dictPath, "utf8");

for (const lang of langs) {
  const file = path.join(localeDir, `${lang}.json`);
  if (!fs.existsSync(file)) {
    console.warn("skip", lang, "(no locale file)");
    continue;
  }
  const entries = JSON.parse(fs.readFileSync(file, "utf8"));
  const lines = Object.entries(entries).map(
    ([k, v]) => `  "${k}": ${JSON.stringify(v)},`,
  );
  if (!lines.length) continue;

  const marker = `const ${lang}: Record<string, string> = {`;
  const start = dict.indexOf(marker);
  if (start === -1) throw new Error(`block not found: ${lang}`);

  const close = dict.indexOf("\n};", start);
  if (close === -1) throw new Error(`close not found: ${lang}`);

  const block = dict.slice(start, close);
  const toInsert = lines.filter((line) => {
    const key = line.match(/^  "([^"]+)":/)?.[1];
    return key && !block.includes(`"${key}":`);
  });

  if (!toInsert.length) {
    console.log(lang, "up to date");
    continue;
  }

  dict = dict.slice(0, close) + "\n" + toInsert.join("\n") + dict.slice(close);
  console.log(lang, "added", toInsert.length, "keys");
}

fs.writeFileSync(dictPath, dict, "utf8");
console.log("dict.ts updated");
