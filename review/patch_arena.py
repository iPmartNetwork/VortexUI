#!/usr/bin/env python3
"""Build Arena-style docs site for GitHub Pages (v1.2.9)."""
from __future__ import annotations

import shutil
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
SRC = ROOT / "arena-export" / "index.html"
OUT_DIR = ROOT / "site"
BASE = "/VortexUI/"

# Exact bundle string replacements (en + fa)
REPLACEMENTS = [
    ('"hero.badge.version":"Version 1.2.8', '"hero.badge.version":"Version 1.2.9'),
    ('"hero.badge.refresh":"Veltrix UI Refresh"', '"hero.badge.refresh":"Command Tower UI"'),
    ('"hero.badge.refresh":"بازطراحی Veltrix UI"', '"hero.badge.refresh":"Command Tower UI"'),
    ("Version 1.2.8", "Version 1.2.9"),
    ("نسخه ۱.۲.۸", "نسخه ۱.۲.۹"),
    ("v1.2.8", "v1.2.9"),
    ("1.2.8", "1.2.9"),
    (
        "Veltrix Glass UI, anti-censorship suite, reseller platform, and 14+ protocol support.",
        "Command Tower UI, Settings hub, reseller profiles, anti-censorship suite, and 14+ protocols.",
    ),
    (
        "Glass design system, collapsible sidebar, command palette, and 8-language i18n.",
        "Merged admin pages, Settings hub, reseller profiles, fleet telemetry, and 8-language i18n.",
    ),
    ('href="/favicon-light.svg"', f'href="{BASE}favicon-light.svg"'),
    ('href="/favicon-dark.svg"', f'href="{BASE}favicon-dark.svg"'),
]


def main() -> int:
    if not SRC.is_file():
        print(f"missing {SRC}", file=sys.stderr)
        return 1

    text = SRC.read_text(encoding="utf-8")
    if "<base " not in text:
        text = text.replace("<head>", f'<head>\n    <base href="{BASE}" />', 1)

    for old, new in REPLACEMENTS:
        text = text.replace(old, new)

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    index = OUT_DIR / "index.html"
    index.write_text(text, encoding="utf-8")
    shutil.copy2(index, OUT_DIR / "404.html")

    for name in ("favicon-light.svg", "favicon-dark.svg"):
        src = ROOT / "arena-export" / name
        if not src.is_file():
            src = ROOT / name
        if src.is_file():
            shutil.copy2(src, OUT_DIR / name)

    # touch for GitHub Pages
    (OUT_DIR / ".nojekyll").write_text("", encoding="utf-8")
    print(f"built {OUT_DIR} ({index.stat().st_size} bytes)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
