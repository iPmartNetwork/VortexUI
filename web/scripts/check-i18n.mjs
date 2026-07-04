import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const src = fs.readFileSync(path.join(__dirname, "../src/i18n/dict.ts"), "utf8");

function extractBlock(lang) {
  if (lang === "en") {
    const m = src.match(/const en = \{([\s\S]*?)\n\} as const;/);
    return m ? m[1] : null;
  }
  const re = new RegExp(
    `const ${lang}: Record<string, string> = \\{([\\s\\S]*?)\\n\\};`,
    "m",
  );
  const m = src.match(re);
  return m ? m[1] : null;
}

function extractKeys(block) {
  return [...block.matchAll(/"([^"]+)":/g)].map((x) => x[1]);
}

function extractPairs(block) {
  const pairs = {};
  for (const m of block.matchAll(/"([^"]+)":\s*"((?:\\.|[^"\\])*)"/g)) {
    pairs[m[1]] = m[2];
  }
  return pairs;
}

const enBlock = extractBlock("en");
const enKeys = extractKeys(enBlock);
const en = extractPairs(enBlock);

const langs = ["fa", "tr", "ar", "ru", "zh", "ja", "es"];
for (const l of langs) {
  const block = extractBlock(l);
  const keys = new Set(extractKeys(block));
  const missing = enKeys.filter((k) => !keys.has(k));
  const lines = missing.map((k) => `${k}\t${en[k] ?? ""}`);
  fs.writeFileSync(path.join(__dirname, `missing-${l}.tsv`), lines.join("\n"), "utf8");
  console.log(l, keys.size, "missing", missing.length);
}

console.log("EN total:", enKeys.length);
