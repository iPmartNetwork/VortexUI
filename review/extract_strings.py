#!/usr/bin/env python3
"""Extract readable strings from Arena docs bundle."""
from __future__ import annotations

import re
from pathlib import Path

text = Path(__file__).resolve().parent.joinpath("arena-export", "index.html").read_text(encoding="utf-8")

# i18n map entries like key:"value"
pairs = re.findall(r'(hero\.[^"]+|inst\.[^"]+|docs\.[^"]+|nav\.[^"]+|feat\.[^"]+):"((?:\\.|[^"\\])*)"', text)
seen = set()
for k, v in pairs:
    if k in seen:
        continue
    seen.add(k)
    v = v.encode().decode("unicode_escape") if "\\" in v else v
    print(f"{k} = {v[:120]}")

print("\n--- version mentions ---")
for m in re.finditer(r'.{0,40}1\.2\.8.{0,40}', text):
    s = m.group().replace("\n", " ")
    if "256" not in s and "131072" not in s:
        print(s[:100])
