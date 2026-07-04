<div align="center" dir="rtl">
  <img src="img/Logo.svg" alt="VortexUI" width="300" />
  <p><strong>پنل مدیریت پروکسی نسل جدید</strong></p>
  <p>مستقل از هسته · کاربر‌محور · بلادرنگ</p>

  [![Release](https://img.shields.io/github/v/release/iPmartNetwork/VortexUI?style=flat-square&color=blue)](https://github.com/iPmartNetwork/VortexUI/releases)
  [![License](https://img.shields.io/github/license/iPmartNetwork/VortexUI?style=flat-square)](LICENSE)
  [![CI](https://img.shields.io/github/actions/workflow/status/iPmartNetwork/VortexUI/ci.yml?style=flat-square&label=CI)](https://github.com/iPmartNetwork/VortexUI/actions)
  [![GHCR](https://img.shields.io/badge/ghcr.io-images-2496ED?style=flat-square&logo=docker&logoColor=white)](https://github.com/iPmartNetwork/VortexUI/pkgs/container/vortexui-panel)
  [![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](https://go.dev)
  [![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org)
  [![Telegram](https://img.shields.io/badge/Telegram-@vortex__ui-26A5E4?style=flat-square&logo=telegram&logoColor=white)](https://t.me/vortex_ui)

  <br />

  <a href="README.md">English</a> · <strong>فارسی</strong>

  <br />

  [امکانات](#-امکانات) · [تازه‌های نسخه ۱.۲.۸](#-تازههای-نسخه-۱۲۸) · [تازه‌های نسخه ۱.۲.۷](#-تازههای-نسخه-۱۲۷) · [تازه‌های نسخه ۱.۲.۶](#-تازههای-نسخه-۱۲۶) · [تازه‌های نسخه ۱.۲.۵](#-تازههای-نسخه-۱۲۵) · [تازه‌های نسخه ۱.۲.۳](#-تازههای-نسخه-۱۲۳) · [مقایسه](#-مقایسه) · [نصب سریع](#-نصب-سریع) · [پروتکل‌ها](#-پروتکل‌های-پشتیبانی‌شده) · [عملیات](#-عملیات) · [نقشه‌راه](#-نقشه‌راه)
</div>

---

<div dir="rtl">

## ✨ امکانات

### 🔧 هسته‌ی پروکسی
- **Xray-core** و **sing-box** — برای هر نود جداگانه قابل انتخاب
- نود محلی درون‌پردازه‌ای (بدون نیاز به ایجنت جدا)
- بارگذاری مجدد پیکربندی بدون قطعی، افزودن/حذف کاربر در لحظه
- تولید کلید REALITY به‌صورت داخلی

### 👥 مدیریت کاربران
- مدل کاربر‌محور (یک هویت → چند پروتکل)
- اشتراک (Subscription): تشخیص خودکار Clash/sing-box/base64
- کد QR و لینک برای هر فرمت
- حساب‌داری ترافیک: مبتنی بر دلتا و مقاوم در برابر ری‌استارت
- اعمال محدودیت سهمیه + ریست زمان‌بندی‌شده
- محدودیت تعداد دستگاه + لیست مجاز HWID
- عملیات گروهی (انتخاب چندتایی) و **Add Bulk**
- ورود کاربران از **3x-ui / Marzban**

### 🌐 سیاست شبکه
- خروجی‌ها (Outbounds): freedom/blackhole/dns + زنجیره‌ی پروکسی
- ویرایشگر JSON خروجی/ورودی + ایمپورت لینک اشتراک
- قوانین مسیریابی: تطبیق دامنه/IP/پورت/پروتکل
- به‌روزرسان **GeoIP/Geosite** با قوانین ایران
- متعادل‌کننده‌ی بار: random/roundRobin/leastPing/leastLoad
- Observatory با بررسی سلامت

### 🖥 ناوگان نودها
- اتصال mTLS بین پنل و نود
- فِیل‌اوور خودکار + بازگشت پس از بازیابی
- پایش زنده‌ی منابع (CPU/RAM/Disk)
- ری‌استارت/توقف هسته از راه دور
- به‌روزرسانی GeoIP/Geosite با یک کلیک (قوانین ایران)
- استریم زنده‌ی لاگ هر نود

### 🔔 اعلان‌ها
- گذرگاه رویداد: user.limited، user.expired، node.down و …
- Webhook (امضای HMAC-SHA256)
- اعلان تلگرام

### 🔐 امنیت
- JWT + احراز هویت دومرحله‌ای TOTP
- توکن‌های API (Personal Access Tokens)
- RBAC با دسترسی‌های دانه‌ریز
- محافظت در برابر حمله‌ی brute-force روی ورود
- کنترل اشتراک‌گذاری اکانت (تشخیص IP آنلاین)
- لاگ ممیزی (همه‌ی تغییرات ادمین‌ها)

### 🎨 رابط کاربری
- React 18 + TypeScript + Tailwind + **framer-motion**
- **Veltrix UI** — کارت‌های شیشه‌ای، کاشی آمار، نشان وضعیت، پالت cyan/sky
- **۸ زبان** با **۶۳۹ کلید ترجمه** و پشتیبانی RTL (EN/FA/TR/AR/RU/ZH/JA/ES)
- سایدبار تاشو + هدر ثابت با تغییر تم و زبان
- تم تیره و روشن (انتقال نرم)
- **Command palette** (Ctrl+K)
- واکنش‌گرا (کشوی موبایل + پورتال موبایل‌اول)
- داشبورد بلادرنگ + به‌روزرسانی زنده (SSE)
- HTTPS خودکار با Caddy / Let's Encrypt

---

## 🆕 تازه‌های نسخه ۱.۲.۸

<div align="center">

**Veltrix UI · i18n کامل · بازطراحی پنل ادمین و پورتال**

</div>

| ویژگی | توضیحات |
|-------|---------|
| **سیستم طراحی Veltrix** | کارت شیشه‌ای، آمار زنده، نشان وضعیت، انیمیشن ورود صفحه، پالت cyan/sky |
| **Shell جدید** | سایدبار تاشو + هدر با حالت mini، کشوی موبایل، تغییر تم و زبان |
| **Command palette** | جستجوی صفحات با Ctrl+K / ⌘K |
| **صفحات اصلی** | Overview، Users و Nodes با کارت آمار زنده از API |
| **پورتال کاربری** | ورود، داشبورد، سایدبار دسکتاپ و ناوبری پایین موبایل بازطراحی شد |
| **i18n کامل** | ۶۳۹ کلید در ۸ زبان — billing، پرداخت ریسلر، سفارش‌های در انتظار، shell، portal |
| **ابزار locale** | فایل‌های JSON + اسکریپت apply/check برای نگهداری ترجمه‌ها |

---

## 🆕 تازه‌های نسخه ۱.۲.۷

<div align="center">

**تجارت ریسلر · پلن اختصاصی · آپلود فیش · تمدید سلف‌سرویس**

</div>

| ویژگی | توضیحات |
|-------|---------|
| **تمدید سلف‌سرویس** | خرید پلن از `/sub/:token/shop`؛ ترافیک و مدت به موجودی فعلی اضافه می‌شود |
| **تنظیمات پرداخت ریسلر** | شماره کارت، آدرس رمزارز و مرچنت زرین‌پال برای هر ریسلر |
| **پلن اختصاصی ریسلر** | هر ریسلر پلن و قیمت خود را می‌سازد؛ کاربران فقط پلن ریسلر خود را می‌بینند |
| **آپلود فیش پرداخت** | کارت‌به‌کارت با تصویر فیش؛ رمزارز با TX hash و اسکرین‌شات |
| **بررسی سفارش در انتظار** | تأیید/رد پرداخت دستی با نمایش thumbnail فیش |
| **روش‌های پرداخت دستی** | کارت‌به‌کارت و رمزارز در کنار زرین‌پال در shop و پورتال React |

---

## 🆕 تازه‌های نسخه ۱.۲.۶

<div align="center">

**ویزارد ثبت نود · billing کیف پول · doctor CLI · تشخیص سلامت**

</div>

| ویژگی | توضیحات |
|-------|---------|
| **ویزارد ثبت نود** | چهار مرحله: کپی bundle mTLS → نصب → ثبت → تست اتصال |
| **تشخیص سلامت نود** | طبقه‌بندی قطعی (mTLS / unreachable / core down) + debug bundle |
| **`vortexui doctor`** | بررسی cert، سرویس، پورت و `/health` برای panel/node/docker |
| **billing کیف پول** | بسته چندارزی، زرین‌پال + NowPayments، کارت‌به‌کارت و رمزارز |
| **UI کیف پول** | شارژ از Admins، خروجی CSV، شارژ زیرریسلر توسط والد |

---

## 🆕 تازه‌های نسخه ۱.۲.۵

<div align="center">

**پلتفرم ریسلر · کیف پول و زیرریسلر · برند سفید · وب‌هوک · محدودیت سیاست · تعلیق خودکار · i18n کامل**

</div>

| ویژگی | توضیحات |
|-------|---------|
| **لیست مجاز** | انتخاب پلن، نود و inbound برای هر ریسلر — فقط همان‌ها را می‌بیند |
| **حالت سهمیه ترافیک** | `allocated` در برابر `consumed` برای استخر ترافیک |
| **داشبورد ریسلر** | اکانت‌ها، استخر ترافیک، پرمصرف‌ها، انقضای نزدیک، خروجی CSV |
| **هشدار سهمیه** | تلگرام + وب‌هوک هنگام نزدیک شدن به سقف |
| **کیف پول** | اعتبار ترافیک/کاربر با تاریخچه و شارژ |
| **زیرریسلر** | سلسله‌مراتب ریسلر فرزند با نقش و سهمیه |
| **برند سفید** | عنوان، لوگو، رنگ، اسلاگ و فوتر اختصاصی |
| **وب‌هوک خروجی** | رویدادهای امضاشده `user.created` / `user.deleted` |
| **ورود به‌عنوان** | sudo با **Login as** برای پشتیبانی |
| **ممیزی محدود** | ریسلر فقط لاگ خودش را می‌بیند |
| **محدودیت سیاست** | سقف data/expire و کنترل bulk |
| **تعلیق خودکار** | worker برای تخلف IP و عبور از سهمیه |
| **تنظیم سریع سهمیه** | +۵۰ اکانت / +۱۰ / +۵۰ گیگ از جدول Admins |
| **چندزبانه** | همه صفحات ریسلر در ۸ زبان |

راهنمای کامل: [`docs/wiki/fa/18-v125-features.md`](docs/wiki/fa/18-v125-features.md)

---

## 🆕 تازه‌های نسخه ۱.۲.۳

<div align="center">

**میزبان‌های اشتراک · فرمت‌های خروجی جدید · بسته‌های مسیریابی هوشمند · اسکنر IP تمیز · اعمال محدودیت IP · پروتکل‌های بیشتر**

</div>

| ویژگی | توضیحات |
|-------|---------|
| **میزبان‌های اشتراک (Subscription Hosts)** | بازنویسی میزبان به‌ازای هر inbound به سبک Marzban (آدرس/SNI/هدر Host/مسیر/ALPN/اثرانگشت/امنیت/fragment/mux) که در لینک‌های اشتراک اعمال می‌شود، همراه با متغیرهای قالب (`{USERNAME}`، `{SERVER_IP}`، …) |
| **فرمت‌های جدید خروجی اشتراک** | JSON خام Xray/V2Ray، Outline `ss://`، و لینک‌های ساده V2rayN (`?format=xray\|outline\|links`) |
| **بسته‌های قانون مسیریابی هوشمند** | مجموعه‌قوانین مسیریابی قابل‌استفاده‌ی مجدد که روی نودها اعمال یا در اشتراک Clash/sing-box تعبیه می‌شوند؛ انتخاب سراسری + به‌ازای کاربر |
| **اسکنر IP تمیز (Cloudflare)** | اسکن و امتیازدهی IPهای کاندید CDN بر اساس تأخیر + اتلاف بسته، با محافظت SSRF |
| **اعمال محدودیت IP** | هشدار / غیرفعال‌سازی موقت / قطع اتصال‌ها وقتی کاربر از سقف IP/دستگاهش عبور می‌کند (قطع اتصال برای Xray؛ sing-box به غیرفعال‌سازی موقت تنزل می‌یابد) |
| **پروتکل‌های جدید** | `socks`، `http`، `naive` (sing-box)، `dokodemo` (xray)؛ hysteria v1، shadowtls، anytls در sing-box؛ ترنسپورت mKCP؛ ماتریس قابلیت به‌ازای پروتکل |

---

## 📸 تصاویر

<details open>
<summary><strong>☀️ حالت روشن</strong></summary>
<br />

| داشبورد | نودها | کاربران |
|:-------:|:-----:|:-------:|
| ![Overview Light](img/panel/overview_light.png) | ![Nodes Light](img/panel/Node_light.png) | ![Users Light](img/panel/User_light.png) |

</details>

<details>
<summary><strong>🌙 حالت تیره</strong></summary>
<br />

| داشبورد | نودها | کاربران |
|:-------:|:-----:|:-------:|
| ![Overview Dark](img/panel/overview_dark.png) | ![Nodes Dark](img/panel/Node_dark.png) | ![Users Dark](img/panel/User_dark.png) |

</details>

---

## ⚖️ مقایسه

| | VortexUI 1.2.8 | 3x-ui | Marzban | Hiddify |
|:--|:--:|:--:|:--:|:--:|
| **هسته‌ی پروکسی** | Xray + sing-box | Xray | Xray | Xray + sing-box |
| **مدل داده** | کاربر‌محور | inbound‌محور | کاربر‌محور | کاربر‌محور |
| **روش ترافیک** | دلتای push ✨ | polling | polling | polling |
| **چند‌نودی** | ✅ mTLS + failover | ✅ | ✅ | ✅ |
| **متعادل‌کننده** | ✅ | ❌ | ❌ | ❌ |
| **CRUD خروجی** | ✅ | جزئی | ❌ | ❌ |
| **قوانین مسیریابی** | ✅ | ❌ | ❌ | ❌ |
| **توکن API** | ✅ | ❌ | ❌ | ❌ |
| **لاگ ممیزی** | ✅ | ❌ | ❌ | ❌ |
| **ضد‌اشتراک‌گذاری** | ✅ اعمال IP | محدودیت IP | ❌ | ❌ |
| **HTTPS خودکار** | ✅ Caddy/ACME | ❌ | ❌ | ✅ |
| **به‌روزرسان Geo (ایران)** | ✅ | ❌ | ❌ | جزئی |
| **بک‌اند** | Go | Go | Python | Python |
| **پایگاه‌داده** | PG + TimescaleDB | SQLite/PG | SQLite/Maria | SQLite |

---

## 📡 پروتکل‌های پشتیبانی‌شده

| پروتکل | ورودی | خروجی | ترنسپورت |
|--------|:-----:|:-----:|:---------:|
| **VLESS** | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, xHTTP, mKCP |
| **VMess** | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, mKCP |
| **Trojan** | ✅ | ✅ | TCP, WS, gRPC, mKCP |
| **Shadowsocks** | ✅ | ✅ | TCP (+ SS-2022 چندکاربره) |
| **SOCKS** | ✅ | ✅ | TCP |
| **HTTP** | ✅ | ✅ | TCP |
| **Naive** | ✅ (sing-box) | — | TCP/TLS |
| **Dokodemo** | ✅ (xray) | — | TCP/UDP |
| **Hysteria2** | ✅ (sing-box) | — | UDP |
| **Hysteria (v1)** | ✅ (sing-box) | — | UDP |
| **TUIC** | ✅ (sing-box) | — | UDP |
| **ShadowTLS** | ✅ (sing-box) | — | TCP |
| **AnyTLS** | ✅ (sing-box) | — | TCP |
| **WireGuard** | ✅ | — | UDP |

**خروجی اشتراک:** base64 · Clash/Clash.Meta · sing-box · Xray JSON · Outline · لینک‌های ساده (تشخیص خودکار توسط کلاینت).

**لایه‌های امنیت:** None، TLS، REALITY

---

## 🚀 نصب سریع

### نصب یک‌خطی

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

نصب‌کننده تعاملی است و دو چیز می‌پرسد:

۱. **روش نصب**
   - **Docker Compose** *(پیشنهادی)* — کل استک (وب · پنل · نود · PostgreSQL/TimescaleDB · Redis) در کانتینر اجرا می‌شود.
   - **Native (systemd)** — باینری‌های Go به‌صورت سرویس سیستمی اجرا می‌شوند، SPA با Caddy سرو می‌شود و فقط PostgreSQL + Redis در داکر هستند.

۲. **روش دسترسی به پنل**
   - **دامنه + HTTPS خودکار** — یک دامنه (و ایمیل اختیاری) وارد می‌کنی؛ Caddy یک گواهی **Let's Encrypt** می‌گیرد و **خودکار تمدید** می‌کند (نیازمند باز بودن پورت‌های **۸۰ و ۴۴۳** و اشاره‌ی رکورد DNS به سرور).
   - **IP + HTTP** — یک پورت ساده‌ی HTTP انتخاب می‌کنی.

سپس کلیدها و گواهی‌های mTLS را می‌سازد، استک را بالا می‌آورد، اولین ادمین را می‌سازد، آدرس/اطلاعات ورود را چاپ می‌کند و دستور `vortexui` را نصب می‌کند.

اجرای غیرتعاملی (برای اسکریپت):

```bash
VORTEXUI_METHOD=docker VORTEXUI_NONINTERACTIVE=1 \
  VORTEXUI_ADMIN_USER=admin VORTEXUI_ADMIN_PASS='s3cret' \
  bash install.sh
```

### کنسول مدیریت (`vortexui`)

بعد از نصب، کافیست **`vortexui`** را تایپ کنی تا منوی تعاملی (مثل x-ui) بیاید:

```text
   1) Start            2) Stop
   3) Restart          4) Status
   5) Logs             6) Update
   7) Create admin     8) Change web port
   9) Domain / SSL     10) Settings / URL
   11) Uninstall       0) Exit
```

یا به‌صورت ساب‌کامند: `vortexui start|stop|restart|status|logs|update|admin|settings|uninstall`.

### نصب دستی

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI
docker compose up -d           # PostgreSQL + TimescaleDB + Redis
cp .env.example .env           # VORTEX_JWT_SECRET را با openssl rand -hex 32 ست کن
make build && make certs && make run-panel
./bin/panel admin create --username admin --password 'your-password' --sudo
```

---

## 🔒 عملیات

### HTTPS خودکار
لایه‌ی وب **Caddy** است. کافیست یک دامنه ست کنی (`SITE_ADDRESS` در `deploy/.env` یا هنگام نصب) تا Caddy خودکار گواهی **Let's Encrypt** بگیرد و تمدید کند — بدون certbot و بدون cron. برای HTTP ساده مقدار را `:80` بگذار. هر زمان با `vortexui` → *Domain / SSL* قابل تغییر است. گواهی‌ها در volume به نام `caddy-data` پایدار می‌مانند. نیازمند باز بودن پورت‌های **۸۰** و **۴۴۳**.

### به‌روزرسانی زنده (SSE)
پنل رویدادهای دامنه را با **Server-Sent Events** به مرورگر می‌فرستد (`GET /api/events/stream`). رابط کاربری یک‌بار subscribe می‌کند و به‌محض هر تغییر (قطع شدن نود، رسیدن کاربر به سقف، تشخیص اشتراک‌گذاری) همان بخش را تازه می‌کند — بدون polling. توکن از `?access_token=` می‌رود چون `EventSource` نمی‌تواند هدر بفرستد.

### GeoIP / Geosite (قوانین ایران)
هر نود با `geoip.dat` / `geosite.dat` می‌آید. از مسیر **Nodes → Update Geo** آخرین دیتابیس **[Iran-v2ray-rules](https://github.com/chocolate4u/Iran-v2ray-rules)** دانلود می‌شود (شامل `geoip:ir`، `geosite:ir`، `category-ir`، دسته‌های تبلیغات/بدافزار و …) و هسته reload می‌شود؛ پس قوانین مسیریابی می‌توانند IP و دامنه‌های ایرانی را هدف بگیرند (مثلاً *ایران مستقیم، بقیه از پروکسی*). URL سفارشی هم از طریق API (`POST /api/nodes/:id/geo-update`) پذیرفته می‌شود.

### محافظ اشتراک‌گذاری اکانت
یک حلقه‌ی پس‌زمینه تعداد **IPهای آنلاین متمایز** هر کاربر (Xray `GetStatsOnlineIpList`) را با محدودیت دستگاهش مقایسه می‌کند. تخلف یک هشدار `user.ip_limit` (webhook/تلگرام) می‌سازد و با `VORTEX_SHARE_AUTOLIMIT=true` کاربر متخلف را خودکار limit می‌کند (قابل بازگشت).

---

## 📖 مستندات

| موضوع | لینک |
|-------|------|
| **سایت مستندات (MkDocs)** | **[ipmartnetwork.github.io/VortexUI](https://ipmartnetwork.github.io/VortexUI/)** — FA · EN · AR · TR |
| **کانال تلگرام** | [**@vortex_ui**](https://t.me/vortex_ui) — اخبار، release، جامعه کاربران |
| **GitHub Discussions** | [سؤال و پشتیبانی](https://github.com/iPmartNetwork/VortexUI/discussions) |
| **ویکی (منبع Markdown)** | [`docs/WIKI-HUB.md`](docs/WIKI-HUB.md) · [فارسی](docs/wiki/fa/README.md) · [English](docs/wiki/en/README.md) · [العربية](docs/wiki/ar/README.md) · [Türkçe](docs/wiki/tr/README.md) |
| مرجع API (OpenAPI 3.0) | [`docs/openapi.yaml`](docs/openapi.yaml) |
| پروتکل‌ها و مثال پیکربندی | [`docs/protocols.md`](docs/protocols.md) |
| متغیرهای محیطی | [`.env.example`](.env.example) |
| استقرار با Docker | [`deploy/`](deploy/) |
| CI/CD | [`.github/workflows/`](.github/workflows/) |
| تغییرات نسخه‌ها | [`CHANGELOG.md`](CHANGELOG.md) |
| راهنمای مشارکت | [`CONTRIBUTING.md`](CONTRIBUTING.md) |

---

## 🗺 نقشه‌راه

- [x] هسته‌ی مستقل (Xray + sing-box)
- [x] حساب‌داری ترافیک مبتنی بر دلتا
- [x] فِیل‌اوور خودکار + بازگشت
- [x] مدیریت Outbound/Routing/Balancer
- [x] لاگ ممیزی · توکن API · محافظ اشتراک‌گذاری
- [x] ورود از 3x-ui / Marzban
- [x] رابط ۸ زبانه + RTL
- [x] داشبورد بلادرنگ + به‌روزرسانی زنده (SSE)
- [x] HTTPS خودکار (Caddy + Let's Encrypt)
- [x] به‌روزرسان GeoIP/Geosite با قوانین ایران
- [x] نصب‌کننده‌ی یک‌خطی + کنسول مدیریت `vortexui`
- [ ] پروتکل‌های Hysteria2 / TUIC / WireGuard
- [ ] مدیریت DNS
- [ ] بات تلگرام کاربران
- [ ] زیرپنل‌های نمایندگی (Reseller)

---

## 💝 حمایت

اگر VortexUI برایت مفید بود:

- ⭐ به مخزن **ستاره** بده
- 🍴 **Fork** کن و مشارکت کن
- 📢 با دیگران **به اشتراک** بگذار
- 💬 در **کانال تلگرام** عضو شو: [@vortex_ui](https://t.me/vortex_ui)
- 💰 برای حمایت از توسعه، **کریپتو اهدا** کن

| شبکه | آدرس |
|:----:|------|
| **USDT (TRC20)** | `TRLnjZ7YDSwjh3oay28qigEYNieGPMs6ew` |
| **BTC** | `bc1qszt4g7jdv7ev2t3pexctc07ults8nfflht3nj5` |
| **TON** | `UQAYSSSirtQ9_67ZHYUgLVLMx9Ir9vvh3vpcq2qbpit_8-Db` |

---

## 🤝 مشارکت

مشارکت‌ها استقبال می‌شوند! لطفاً:

۱. مخزن را Fork کن
۲. یک شاخه‌ی ویژگی بساز (`git checkout -b feat/amazing`)
۳. کامیت کن (`git commit -m 'feat: add amazing feature'`)
۴. push کن (`git push origin feat/amazing`)
۵. یک Pull Request باز کن

برای راهنمای کامل [CONTRIBUTING.md](CONTRIBUTING.md) را ببین.

---

## 🌐 چندزبانگی

| 🇺🇸 English | 🇮🇷 فارسی | 🇹🇷 Türkçe | 🇸🇦 العربية |
|:-----------:|:---------:|:----------:|:-----------:|
| 🇷🇺 Русский | 🇨🇳 中文 | 🇯🇵 日本語 | 🇪🇸 Español |

پشتیبانی کامل RTL برای فارسی و عربی.

---

## 📄 لایسنس

این پروژه تحت لایسنس **GPL-3.0** منتشر شده است — فایل [LICENSE](LICENSE) را ببین.

---

<div align="center" dir="ltr">
  <sub>© 2026 iPmart Network. All rights reserved.</sub>
  <br />
  <strong>ساخته‌شده با ❤️ توسط <a href="https://github.com/iPmartNetwork">iPmart Network</a></strong>
  <br />
  <a href="https://t.me/vortex_ui">تلگرام @vortex_ui</a> · <a href="https://ipmartnetwork.github.io/VortexUI/fa/">مستندات</a>
</div>

</div>
