#!/usr/bin/env python3
"""Sanitize wiki markdown for MkDocs Material (strip HTML wrappers, fix admonitions)."""

from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent
LANGS = ("en", "fa", "ar", "tr")

ALERT_MAP = {
    "NOTE": "note",
    "TIP": "tip",
    "WARNING": "warning",
    "IMPORTANT": "important",
}


def strip_html_wrappers(text: str) -> str:
    # Remove centered logo / language header block
    text = re.sub(
        r"<div align=\"center\"[^>]*>.*?</div>\s*\n+",
        "",
        text,
        count=1,
        flags=re.DOTALL | re.IGNORECASE,
    )
    # Remove wiki-hero wrapper (README)
    text = re.sub(
        r"<div align=\"center\"[^>]*class=\"wiki-hero\"[^>]*>.*?</div>\s*\n+",
        "",
        text,
        count=1,
        flags=re.DOTALL | re.IGNORECASE,
    )
    # Remove screenshot gallery blocks (HTML table inside div)
    text = re.sub(
        r"<div align=\"center\">\s*\n+\| Light \| Dark \|.*?</div>\s*\n+",
        "",
        text,
        flags=re.DOTALL | re.IGNORECASE,
    )
    # Remove outer content wrapper
    text = re.sub(r"^<div dir=\"rtl\">\s*\n+", "", text, flags=re.IGNORECASE)
    text = re.sub(r"^<div>\s*\n+", "", text)
    text = re.sub(r"\n+</div>\s*$", "\n", text)
    return text


def github_alerts_to_material(text: str) -> str:
    def repl(match: re.Match[str]) -> str:
        kind = ALERT_MAP.get(match.group(1).upper(), "note")
        body = match.group(2).strip()
        lines = [f"!!! {kind}"]
        for line in body.split("\n"):
            line = line.strip()
            if line.startswith("> "):
                line = line[2:]
            lines.append(f"    {line}")
        return "\n".join(lines) + "\n"

    return re.sub(
        r"> \[!(NOTE|TIP|WARNING|IMPORTANT)\]\n((?:> .+\n?)+)",
        repl,
        text,
        flags=re.IGNORECASE,
    )


def fix_admonition_quotes(text: str) -> str:
    """Remove leftover blockquote markers inside Material admonitions."""
    lines = text.split("\n")
    out: list[str] = []
    in_admonition = False
    for line in lines:
        if re.match(r"^!!!\s", line):
            in_admonition = True
            out.append(line)
            continue
        if in_admonition:
            if line.startswith("    "):
                body = line[4:]
                if body.startswith("> "):
                    body = body[2:]
                out.append("    " + body)
                continue
            in_admonition = False
        out.append(line)
    return "\n".join(out)


def simplify_nav_line(text: str) -> str:
    # MkDocs has sidebar; drop redundant breadcrumb nav lines
    text = re.sub(
        r"^\[(?:←[^\]]+|Wiki[^\]]*)\]\([^)]+\)(?: · \[[^\]]+\]\([^)]+\))*\s*\n\n",
        "",
        text,
        flags=re.MULTILINE,
    )
    return text


def ensure_table_spacing(text: str) -> str:
    # Blank line before markdown tables
    return re.sub(r"([^\n|])\n(\|)", r"\1\n\n\2", text)


def fix_readme_home(text: str, lang: str) -> str:
    """Produce a clean MkDocs home page."""
    titles = {
        "en": ("VortexUI Documentation", "Welcome to the official VortexUI guide."),
        "fa": ("مستندات VortexUI", "به راهنمای رسمی VortexUI خوش آمدید."),
        "ar": ("وثائق VortexUI", "مرحباً بكم في دليل VortexUI الرسمي."),
        "tr": ("VortexUI Dokümantasyon", "Resmi VortexUI kılavuzuna hoş geldiniz."),
    }
    title, subtitle = titles.get(lang, titles["en"])
    intro = {
        "en": "Install, configure, and operate the next-generation proxy panel (Xray + sing-box). Use the **language selector** in the header to switch between English, Persian, Arabic, and Turkish.",
        "fa": "نصب، پیکربندی و مدیریت پنل پروکسی نسل جدید (Xray + sing-box). از **انتخابگر زبان** در header برای جابه‌جایی بین ۴ زبان استفاده کنید.",
        "ar": "ثبّت وأدر لوحة البروكسي (Xray + sing-box). استخدم **مبدّل اللغة** في الشريط العلوي.",
        "tr": "Yeni nesil proxy panelini kurun ve yönetin (Xray + sing-box). **Dil seçiciyi** üst menüden kullanın.",
    }[lang]
    return f"""# {title}

{subtitle}

{intro}

!!! tip "Quick install"
    ```bash
    bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
    ```

## Architecture

```mermaid
flowchart TB
    subgraph Client["Clients"]
        Browser["Browser / PWA"]
        ProxyApp["Clash / sing-box / v2rayNG"]
    end
    subgraph Web["Web Layer"]
        Caddy["Caddy — HTTPS + SPA"]
    end
    subgraph Panel["Control Plane"]
        API["Panel API — Go"]
        SSE["SSE — Live Events"]
        DB[(PostgreSQL + TimescaleDB)]
        Redis[(Redis)]
    end
    subgraph Nodes["Node Fleet"]
        Local["Local Node"]
        Remote["Remote Nodes — mTLS"]
    end
    Browser --> Caddy
    ProxyApp --> Caddy
    Caddy --> API
    API --> DB
    API --> Redis
    API --> Local
    API --> Remote
```

## Useful links

| Resource | Link |
|----------|------|
| OpenAPI | [openapi.yaml on GitHub](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) |
| Protocol examples | [protocols.md](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/protocols.md) |
| Repository | [github.com/iPmartNetwork/VortexUI](https://github.com/iPmartNetwork/VortexUI) |
"""


def sanitize_chapter(path: Path) -> None:
    text = path.read_text(encoding="utf-8")
    text = strip_html_wrappers(text)
    text = github_alerts_to_material(text)
    text = fix_admonition_quotes(text)
    text = simplify_nav_line(text)
    text = ensure_table_spacing(text)
    text = re.sub(r"\n{3,}", "\n\n", text).strip() + "\n"
    path.write_text(text, encoding="utf-8")


def sanitize_readme(path: Path, lang: str) -> None:
    path.write_text(fix_readme_home("", lang), encoding="utf-8")


def main() -> None:
    for lang in LANGS:
        d = ROOT / lang
        for f in sorted(d.glob("[0-9][0-9]-*.md")):
            sanitize_chapter(f)
            print("fixed", f.relative_to(ROOT))
        readme = d / "README.md"
        if readme.exists():
            sanitize_readme(readme, lang)
            print("home", readme.relative_to(ROOT))


if __name__ == "__main__":
    main()
