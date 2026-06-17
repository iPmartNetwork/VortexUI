#!/usr/bin/env python3
"""Sanitize wiki markdown for MkDocs Material (strip HTML wrappers, fix admonitions)."""

from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent
LANGS = ("en", "fa", "ar", "tr")
GITHUB = "https://github.com/iPmartNetwork/VortexUI/blob/master"

ALERT_MAP = {
    "NOTE": "note",
    "TIP": "tip",
    "WARNING": "warning",
    "IMPORTANT": "important",
}

ADMONITION_TITLES = {
    "en": {"note": "Note", "tip": "Tip", "warning": "Warning", "important": "Important"},
    "fa": {"note": "نکته", "tip": "راهنما", "warning": "هشدار", "important": "مهم"},
    "ar": {"note": "ملاحظة", "tip": "نصيحة", "warning": "تحذير", "important": "مهم"},
    "tr": {"note": "Not", "tip": "İpucu", "warning": "Uyarı", "important": "Önemli"},
}

HOME_COPY = {
    "en": {
        "title": "VortexUI Documentation",
        "subtitle": "Welcome to the official VortexUI guide.",
        "intro": "Install, configure, and operate the next-generation proxy panel (Xray + sing-box). Use the **language selector** in the header to switch between English, Persian, Arabic, and Turkish.",
        "tip": "Quick install",
        "map": "Documentation map",
        "map_section": "Section",
        "map_chapters": "Chapters",
        "architecture": "Architecture",
        "links": "Useful links",
        "resource": "Resource",
        "link": "Link",
        "rows": [
            ("Getting started", "[Introduction](01-introduction.md) · [Installation](02-installation.md) · [First steps](03-first-steps.md)"),
            ("Panel guide", "[Dashboard](04-dashboard.md) · [Users](05-user-management.md) · [Nodes](06-node-management.md) · [Network](07-network-policy.md)"),
            ("Administration", "[Security](08-security-administration.md) · [Plans](09-plans-payments.md) · [Notifications](10-notifications.md) · [Settings](11-settings-backup.md)"),
            ("Technical reference", "[API](12-api-reference.md) · [Protocols](13-protocols-config.md) · [Operations](14-operations-maintenance.md) · [FAQ](15-troubleshooting-faq.md)"),
        ],
    },
    "fa": {
        "title": "مستندات VortexUI",
        "subtitle": "به راهنمای رسمی VortexUI خوش آمدید.",
        "intro": "نصب، پیکربندی و مدیریت پنل پروکسی نسل جدید (Xray + sing-box). از **انتخابگر زبان** در بالای صفحه برای جابه‌جایی بین ۴ زبان استفاده کنید.",
        "tip": "نصب سریع",
        "map": "نقشه مستندات",
        "map_section": "بخش",
        "map_chapters": "فصل‌ها",
        "architecture": "معماری",
        "links": "لینک‌های مفید",
        "resource": "منبع",
        "link": "لینک",
        "rows": [
            ("شروع کار", "[معرفی](01-introduction.md) · [نصب](02-installation.md) · [اولین قدم‌ها](03-first-steps.md)"),
            ("راهنمای پنل", "[داشبورد](04-dashboard.md) · [کاربران](05-user-management.md) · [نودها](06-node-management.md) · [شبکه](07-network-policy.md)"),
            ("مدیریت", "[امنیت](08-security-administration.md) · [پلن‌ها](09-plans-payments.md) · [اعلان‌ها](10-notifications.md) · [تنظیمات](11-settings-backup.md)"),
            ("مرجع فنی", "[API](12-api-reference.md) · [پروتکل‌ها](13-protocols-config.md) · [عملیات](14-operations-maintenance.md) · [FAQ](15-troubleshooting-faq.md)"),
        ],
    },
    "ar": {
        "title": "وثائق VortexUI",
        "subtitle": "مرحباً بكم في دليل VortexUI الرسمي.",
        "intro": "ثبّت وأدر لوحة البروكسي (Xray + sing-box). استخدم **مبدّل اللغة** في الشريط العلوي للتنقل بين اللغات الأربع.",
        "tip": "تثبيت سريع",
        "map": "خريطة الوثائق",
        "map_section": "القسم",
        "map_chapters": "الفصول",
        "architecture": "البنية",
        "links": "روابط مفيدة",
        "resource": "المورد",
        "link": "الرابط",
        "rows": [
            ("البدء", "[المقدمة](01-introduction.md) · [التثبيت](02-installation.md) · [الخطوات الأولى](03-first-steps.md)"),
            ("دليل اللوحة", "[لوحة المعلومات](04-dashboard.md) · [المستخدمون](05-user-management.md) · [العقد](06-node-management.md) · [الشبكة](07-network-policy.md)"),
            ("الإدارة", "[الأمان](08-security-administration.md) · [الخطط](09-plans-payments.md) · [الإشعارات](10-notifications.md) · [الإعدادات](11-settings-backup.md)"),
            ("مرجع تقني", "[API](12-api-reference.md) · [البروتوكولات](13-protocols-config.md) · [العمليات](14-operations-maintenance.md) · [FAQ](15-troubleshooting-faq.md)"),
        ],
    },
    "tr": {
        "title": "VortexUI Dokümantasyon",
        "subtitle": "Resmi VortexUI kılavuzuna hoş geldiniz.",
        "intro": "Yeni nesil proxy panelini kurun ve yönetin (Xray + sing-box). **Dil seçiciyi** üst menüden kullanın.",
        "tip": "Hızlı kurulum",
        "map": "Dokümantasyon haritası",
        "map_section": "Bölüm",
        "map_chapters": "Bölümler",
        "architecture": "Mimari",
        "links": "Faydalı bağlantılar",
        "resource": "Kaynak",
        "link": "Bağlantı",
        "rows": [
            ("Başlangıç", "[Giriş](01-introduction.md) · [Kurulum](02-installation.md) · [İlk adımlar](03-first-steps.md)"),
            ("Panel rehberi", "[Pano](04-dashboard.md) · [Kullanıcılar](05-user-management.md) · [Node'lar](06-node-management.md) · [Ağ](07-network-policy.md)"),
            ("Yönetim", "[Güvenlik](08-security-administration.md) · [Planlar](09-plans-payments.md) · [Bildirimler](10-notifications.md) · [Ayarlar](11-settings-backup.md)"),
            ("Teknik referans", "[API](12-api-reference.md) · [Protokoller](13-protocols-config.md) · [Operasyonlar](14-operations-maintenance.md) · [SSS](15-troubleshooting-faq.md)"),
        ],
    },
}


def localize_admonitions(text: str, lang: str) -> str:
    titles = ADMONITION_TITLES.get(lang, ADMONITION_TITLES["en"])

    def repl(match: re.Match[str]) -> str:
        kind = match.group(1).lower()
        if match.group(2):
            return match.group(0)
        return f'!!! {kind} "{titles[kind]}"'

    return re.sub(
        r"^!!! (note|tip|warning|important)(\s+\"[^\"]+\")?\s*$",
        repl,
        text,
        flags=re.MULTILINE | re.IGNORECASE,
    )


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


def fix_external_links(text: str) -> str:
    """Point repo-root references at GitHub (outside docs_dir)."""
    replacements = {
        "../../openapi.yaml": f"{GITHUB}/docs/openapi.yaml",
        "../../protocols.md": f"{GITHUB}/docs/protocols.md",
        "../../../.env.example": f"{GITHUB}/.env.example",
        "../../../SECURITY.md": f"{GITHUB}/SECURITY.md",
        "../../../CONTRIBUTING.md": f"{GITHUB}/CONTRIBUTING.md",
        "../../../CHANGELOG.md": f"{GITHUB}/CHANGELOG.md",
    }
    for old, new in replacements.items():
        text = text.replace(old, new)
    return text


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
    c = HOME_COPY.get(lang, HOME_COPY["en"])
    map_rows = "\n".join(f"| {section} | {chapters} |" for section, chapters in c["rows"])
    return f"""# {c["title"]}

{c["subtitle"]}

{c["intro"]}

!!! tip "{c["tip"]}"
    ```bash
    bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
    ```

## {c["map"]}

| {c["map_section"]} | {c["map_chapters"]} |
|---------|----------|
{map_rows}

## {c["architecture"]}

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

## {c["links"]}

| {c["resource"]} | {c["link"]} |
|----------|------|
| OpenAPI | [openapi.yaml on GitHub](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) |
| Protocol examples | [protocols.md](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/protocols.md) |
| Repository | [github.com/iPmartNetwork/VortexUI](https://github.com/iPmartNetwork/VortexUI) |
| Telegram | [@vortex_ui](https://t.me/vortex_ui) |
"""


def sanitize_chapter(path: Path, lang: str) -> None:
    text = path.read_text(encoding="utf-8")
    text = strip_html_wrappers(text)
    text = github_alerts_to_material(text)
    text = fix_admonition_quotes(text)
    text = localize_admonitions(text, lang)
    text = simplify_nav_line(text)
    text = fix_external_links(text)
    text = ensure_table_spacing(text)
    text = re.sub(r"\n{3,}", "\n\n", text).strip() + "\n"
    path.write_text(text, encoding="utf-8")


def sanitize_readme(path: Path, lang: str) -> None:
    path.write_text(fix_readme_home("", lang), encoding="utf-8")


def main() -> None:
    for lang in LANGS:
        d = ROOT / lang
        for f in sorted(d.glob("[0-9][0-9]-*.md")):
            sanitize_chapter(f, lang)
            print("fixed", f.relative_to(ROOT))
        readme = d / "README.md"
        if readme.exists():
            sanitize_readme(readme, lang)
            print("home", readme.relative_to(ROOT))


if __name__ == "__main__":
    main()
