#!/usr/bin/env python3
"""Enhance wiki chapters: unified header, callouts, screenshots."""
from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent
LOGO = "../assets/Logo.svg"
IMG = "../assets/panel"

LANGS = ("fa", "en", "ar", "tr")
RTL = {"fa", "ar"}

# per-chapter: (screenshot_file or None, callout_key)
CHAPTER_META: dict[str, tuple[str | None, str]] = {
    "01": (None, "intro"),
    "02": (None, "install"),
    "03": ("overview_light.png", "first_steps"),
    "04": ("overview_light.png", "dashboard"),
    "05": ("User_light.png", "users"),
    "06": ("Node_light.png", "nodes"),
    "07": (None, "network"),
    "08": (None, "security"),
    "09": (None, "plans"),
    "10": (None, "notifications"),
    "11": (None, "settings"),
    "12": (None, "api"),
    "13": (None, "protocols"),
    "14": (None, "ops"),
    "15": (None, "troubleshoot"),
}

CALLOUTS: dict[str, dict[str, tuple[str, str]]] = {
    "intro": {
        "fa": ("NOTE", "VortexUI مدل **کاربر‌محور** دارد — یک subscription برای همه inboundها."),
        "en": ("NOTE", "VortexUI is **user-centric** — one subscription covers all assigned inbounds."),
        "ar": ("NOTE", "VortexUI **يركز على المستخدم** — اشتراك واحد يغطي جميع inbounds المعيّنة."),
        "tr": ("NOTE", "VortexUI **kullanıcı merkezlidir** — tek abonelik tüm atanmış inbound'ları kapsar."),
    },
    "install": {
        "fa": ("IMPORTANT", "قبل نصب، پورت‌های **80 و 443** (برای HTTPS) و DNS دامنه را آماده کنید."),
        "en": ("IMPORTANT", "Before installing, prepare **ports 80 and 443** (for HTTPS) and your domain DNS."),
        "ar": ("IMPORTANT", "قبل التثبيت، جهّز **المنافذ 80 و 443** (لـ HTTPS) وDNS النطاق."),
        "tr": ("IMPORTANT", "Kurulumdan önce **80 ve 443 portlarını** (HTTPS) ve alan adı DNS'ini hazırlayın."),
    },
    "first_steps": {
        "fa": ("TIP", "کل این گردش کار را در **۵ دقیقه** انجام دهید: نود → inbound → user → subscription → تست."),
        "en": ("TIP", "Complete this workflow in **5 minutes**: node → inbound → user → subscription → test."),
        "ar": ("TIP", "أكمل هذا المسار في **5 دقائق**: عقدة → inbound → مستخدم → اشتراك → اختبار."),
        "tr": ("TIP", "Bu akışı **5 dakikada** tamamlayın: node → inbound → kullanıcı → abonelik → test."),
    },
    "dashboard": {
        "fa": ("NOTE", "داشبورد با **SSE** بدون refresh به‌روز می‌شود — polling لازم نیست."),
        "en": ("NOTE", "The dashboard updates via **SSE** without refresh — no polling needed."),
        "ar": ("NOTE", "لوحة المعلومات تتحدّث عبر **SSE** دون refresh — لا حاجة لـ polling."),
        "tr": ("NOTE", "Pano **SSE** ile yenilenir — polling gerekmez."),
    },
    "users": {
        "fa": ("WARNING", "**Revoke Sub** لینک قبلی را باطل می‌کند — فقط در صورت نشت token استفاده کنید."),
        "en": ("WARNING", "**Revoke Sub** invalidates the old link — use only if the token was leaked."),
        "ar": ("WARNING", "**Revoke Sub** يُبطل الرابط السابق — استخدمه فقط عند تسريب الرمز."),
        "tr": ("WARNING", "**Revoke Sub** eski bağlantıyı geçersiz kılar — yalnızca token sızdıysa kullanın."),
    },
    "nodes": {
        "fa": ("TIP", "برای مسیریابی ایران از **Nodes → Update Geo** استفاده کنید."),
        "en": ("TIP", "For Iran routing rules, use **Nodes → Update Geo**."),
        "ar": ("TIP", "لقواعد التوجيه الإيرانية استخدم **Nodes → Update Geo**."),
        "tr": ("TIP", "İran yönlendirme kuralları için **Nodes → Update Geo** kullanın."),
    },
    "network": {
        "fa": ("TIP", "الگوی رایج: `geosite:ir` و `geoip:ir` → direct، بقیه → proxy."),
        "en": ("TIP", "Common pattern: `geosite:ir` and `geoip:ir` → direct, everything else → proxy."),
        "ar": ("TIP", "النمط الشائع: `geosite:ir` و `geoip:ir` → direct، الباقي → proxy."),
        "tr": ("TIP", "Yaygın kalıp: `geosite:ir` ve `geoip:ir` → direct, geri kalan → proxy."),
    },
    "security": {
        "fa": ("IMPORTANT", "برای ادمین اصلی حتماً **2FA** و رمز JWT قوی (≥32 بایت) فعال کنید."),
        "en": ("IMPORTANT", "Enable **2FA** and a strong JWT secret (≥32 bytes) for the primary admin."),
        "ar": ("IMPORTANT", "فعّل **2FA** وJWT secret قوي (≥32 بايت) للمسؤول الرئيسي."),
        "tr": ("IMPORTANT", "Ana admin için **2FA** ve güçlü JWT secret (≥32 bayt) etkinleştirin."),
    },
    "plans": {
        "fa": ("NOTE", "پس از پرداخت موفق، user به‌صورت خودکار با پارامترهای plan ساخته/تمدید می‌شود."),
        "en": ("NOTE", "After successful payment, a user is auto-created/renewed with the plan parameters."),
        "ar": ("NOTE", "بعد الدفع الناجح يُنشأ/يُجدّد المستخدم تلقائياً بمعاملات الخطة."),
        "tr": ("NOTE", "Başarılı ödemeden sonra kullanıcı plan parametreleriyle otomatik oluşturulur/yenilenir."),
    },
    "notifications": {
        "fa": ("TIP", "Webhook با `X-Vortex-Signature: sha256=...` امضا می‌شود — secret را در env تنظیم کنید."),
        "en": ("TIP", "Webhooks are signed with `X-Vortex-Signature: sha256=...` — set the secret in env."),
        "ar": ("TIP", "Webhooks موقّعة بـ `X-Vortex-Signature: sha256=...` — اضبط secret في env."),
        "tr": ("TIP", "Webhook'lar `X-Vortex-Signature: sha256=...` ile imzalanır — secret'ı env'de ayarlayın."),
    },
    "settings": {
        "fa": ("TIP", "قبل از **Restore** حتماً backup فعلی بگیرید."),
        "en": ("TIP", "Always take a current backup before **Restore**."),
        "ar": ("TIP", "خذ نسخة احتياطية حالية دائماً قبل **Restore**."),
        "tr": ("TIP", "**Restore** öncesinde mutlaka güncel yedek alın."),
    },
    "api": {
        "fa": ("NOTE", "مرجع کامل: [`docs/openapi.yaml`](../../openapi.yaml) — همه endpointها و RBAC."),
        "en": ("NOTE", "Full spec: [`docs/openapi.yaml`](../../openapi.yaml) — all endpoints and RBAC."),
        "ar": ("NOTE", "المواصفات الكاملة: [`docs/openapi.yaml`](../../openapi.yaml) — جميع endpoints وRBAC."),
        "tr": ("NOTE", "Tam şema: [`docs/openapi.yaml`](../../openapi.yaml) — tüm endpoint'ler ve RBAC."),
    },
    "protocols": {
        "fa": ("TIP", "برای سانسور شدید: **VLESS + REALITY + Vision** — کلیدها را از UI Generate کنید."),
        "en": ("TIP", "Under heavy censorship: **VLESS + REALITY + Vision** — generate keys from the UI."),
        "ar": ("TIP", "تحت رقابة شديدة: **VLESS + REALITY + Vision** — أنشئ المفاتيح من الواجهة."),
        "tr": ("TIP", "Ağır sansür altında: **VLESS + REALITY + Vision** — anahtarları UI'dan oluşturun."),
    },
    "ops": {
        "fa": ("TIP", "بعد از نصب `vortexui status` و `curl .../api/health` را برای sanity check اجرا کنید."),
        "en": ("TIP", "After install run `vortexui status` and `curl .../api/health` as a sanity check."),
        "ar": ("TIP", "بعد التثبيت نفّذ `vortexui status` و`curl .../api/health` للتحقق السريع."),
        "tr": ("TIP", "Kurulumdan sonra `vortexui status` ve `curl .../api/health` ile kontrol edin."),
    },
    "troubleshoot": {
        "fa": ("TIP", "اول **`vortexui logs`** و **`/api/health`** — بیشتر مشکلات از JWT، DB یا firewall است."),
        "en": ("TIP", "Start with **`vortexui logs`** and **`/api/health`** — most issues are JWT, DB, or firewall."),
        "ar": ("TIP", "ابدأ بـ **`vortexui logs`** و**`/api/health`** — أغلب المشاكل JWT أو DB أو firewall."),
        "tr": ("TIP", "Önce **`vortexui logs`** ve **`/api/health`** — çoğu sorun JWT, DB veya firewall."),
    },
}

SCREENSHOT_CAPTION: dict[str, dict[str, str]] = {
    "03": {
        "fa": "نمای کلی پنل — حالت روشن",
        "en": "Panel overview — light mode",
        "ar": "نظرة عامة على اللوحة — الوضع الفاتح",
        "tr": "Panel genel görünüm — açık tema",
    },
    "04": {
        "fa": "داشبورد Overview — آمار زنده و نمودار ترافیک",
        "en": "Overview dashboard — live stats and traffic chart",
        "ar": "لوحة Overview — إحصائيات مباشرة ومخطط الحركة",
        "tr": "Overview panosu — canlı istatistikler ve trafik grafiği",
    },
    "05": {
        "fa": "صفحه Users — مدیریت کاربران و subscription",
        "en": "Users page — user management and subscriptions",
        "ar": "صفحة Users — إدارة المستخدمين والاشتراك",
        "tr": "Users sayfası — kullanıcı yönetimi ve abonelik",
    },
    "06": {
        "fa": "صفحه Nodes — مانیتورینگ CPU/RAM و عملیات نود",
        "en": "Nodes page — CPU/RAM monitoring and node actions",
        "ar": "صفحة Nodes — مراقبة CPU/RAM وإجراءات العقد",
        "tr": "Nodes sayfası — CPU/RAM izleme ve node işlemleri",
    },
}

LANG_LABELS = {
    "fa": ("Wiki", "EN", "AR", "TR"),
    "en": ("Wiki", "FA", "AR", "TR"),
    "ar": ("Wiki", "EN", "FA", "TR"),
    "tr": ("Wiki", "EN", "FA", "AR"),
}


def lang_links(lang: str, num: str, name: str) -> str:
    others = [l for l in LANGS if l != lang]
    parts = [f"[{LANG_LABELS[lang][0]}](../README.md)"]
    labels = {"en": "EN", "fa": "FA", "ar": "AR", "tr": "TR"}
    for i, o in enumerate(others):
        lbl = labels[o]
        parts.append(f"[{lbl}](../{o}/{name})")
    return " · ".join(parts)


def build_header(lang: str, num: str, name: str) -> str:
    rtl = lang in RTL
    dir_open = ' dir="rtl"' if rtl else ""
    links = lang_links(lang, num, name)
    return f"""<div align="center"{dir_open}>

<img src="{LOGO}" alt="VortexUI" width="120" />

**VortexUI Wiki**

{links}

</div>

"""


def build_callout(lang: str, key: str) -> str:
    kind, text = CALLOUTS[key][lang]
    return f"> [!{kind}]\n> {text}\n\n"


def build_screenshot(lang: str, num: str, filename: str) -> str:
    cap = SCREENSHOT_CAPTION.get(num, {}).get(lang, "")
    dark = filename.replace("_light", "_dark")
    return f"""<div align="center">

| Light | Dark |
|:-----:|:----:|
| ![{cap}]({IMG}/{filename}) | ![{cap}]({IMG}/{dark}) |

*{cap}*

</div>

"""


def enhance_file(path: Path, lang: str) -> None:
    text = path.read_text(encoding="utf-8")
    m = re.match(r"(\d{2})-", path.name)
    if not m:
        return
    num = m.group(1)
    screenshot, callout_key = CHAPTER_META[num]

    # strip old opening: first div through first --- after h1 breadcrumb
    body = text
    if body.startswith("<div"):
        # remove first wrapper div block (nav only)
        body = re.sub(r"^<div(?: dir=\"rtl\")?>\s*\n\n?\[← Wiki\][^\n]+\n\n?", "", body, count=1)

    # extract inner content div if present
    inner_match = re.match(r"(<div dir=\"rtl\">|<div>)\s*\n", body)
    if inner_match:
        body = body[len(inner_match.group(0)) :]
    if body.rstrip().endswith("</div>"):
        body = body.rstrip()[:-6].rstrip()

    # find h1 line and breadcrumb after it
    lines = body.split("\n")
    h1_idx = next(i for i, l in enumerate(lines) if l.startswith("# "))
    # keep from h1 onwards (includes breadcrumb lines until ---)
    rest = "\n".join(lines[h1_idx:])

    header = build_header(lang, num, path.name)
    callout = build_callout(lang, callout_key)
    shot = build_screenshot(lang, num, screenshot) if screenshot else ""

    rtl = lang in RTL
    wrap_open = '<div dir="rtl">\n\n' if rtl else "<div>\n\n"
    new_text = header + wrap_open + rest.split("---", 1)[0].rstrip() + "\n\n" + callout
    if shot:
        new_text += shot
    remainder = rest.split("---", 1)
    if len(remainder) > 1:
        new_text += "---\n\n" + remainder[1].lstrip()
    new_text = new_text.rstrip() + "\n\n</div>\n"

    path.write_text(new_text, encoding="utf-8")


def enhance_readme(path: Path, lang: str) -> None:
    text = path.read_text(encoding="utf-8")
    if "wiki-hero" in text:
        return
    rtl = lang in RTL
    dir_attr = ' dir="rtl"' if rtl else ""
    hero = f"""<div align="center"{dir_attr} class="wiki-hero">

<img src="../assets/Logo.svg" alt="VortexUI" width="160" />

"""
    text = re.sub(
        r"<div align=\"center\"(?: dir=\"rtl\")?>\s*\n\n# ",
        hero + "# ",
        text,
        count=1,
    )
    path.write_text(text, encoding="utf-8")


def main() -> None:
    for lang in LANGS:
        d = ROOT / lang
        for f in sorted(d.glob("[0-9][0-9]-*.md")):
            enhance_file(f, lang)
            print("enhanced", f.relative_to(ROOT))
        readme = d / "README.md"
        if readme.exists():
            enhance_readme(readme, lang)
            print("readme", readme.relative_to(ROOT))


if __name__ == "__main__":
    main()
