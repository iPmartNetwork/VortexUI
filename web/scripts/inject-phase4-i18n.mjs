// Injects phase4-i18n.json keys before each language block's closing brace.
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const dictPath = join(__dirname, "..", "src", "i18n", "dict.ts");
const keys = JSON.parse(readFileSync(join(__dirname, "phase4-i18n.json"), "utf8"));

const langs = [
  { name: "en", anchor: "\n} as const;\n\nconst fa" },
  { name: "fa", anchor: "\n};\n\nconst tr" },
  { name: "tr", anchor: "\n};\n\nconst ar" },
  { name: "ar", anchor: "\n};\n\nconst ru" },
  { name: "ru", anchor: "\n};\n\nconst zh" },
  { name: "zh", anchor: "\n};\n\nconst ja" },
  { name: "ja", anchor: "\n};\n\nconst es" },
  { name: "es", anchor: "\n};\n\nexport const dict" },
];

function buildInsertion(lang) {
  const lines = ["  /* phase4 i18n */"];
  for (const [key, translations] of Object.entries(keys)) {
    const value = translations[lang] ?? translations.en;
    const escaped = value.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
    lines.push(`  "${key}": "${escaped}",`);
  }
  return "\n" + lines.join("\n") + "\n";
}

let text = readFileSync(dictPath, "utf8");

for (const { name, anchor } of langs) {
  const idx = text.indexOf(anchor);
  if (idx < 0) {
    console.error(`Anchor not found for ${name}: ${JSON.stringify(anchor)}`);
    process.exit(1);
  }
  let body = text.slice(0, idx);
  body = body.replace(/\n  \/\* phase4 i18n \*\/\n(?:  "[^"]+": ".+",\n)+$/, "");
  text = body + buildInsertion(name) + text.slice(idx);
}

writeFileSync(dictPath, text);
console.log(`Injected ${Object.keys(keys).length} keys × ${langs.length} langs.`);
