<div align="center" dir="rtl">
  <img src="img/Logo.svg" alt="VortexUI" width="300" />
  <p><strong>پنل مدیریت پروکسی نسل جدید</strong></p>
  <p>مستقل از هسته · کاربر‌محور · بلادرنگ</p>

  [![Release](https://img.shields.io/github/v/release/iPmartNetwork/VortexUI?style=flat-square&color=blue)](https://github.com/iPmartNetwork/VortexUI/releases)
  [![License](https://img.shields.io/github/license/iPmartNetwork/VortexUI?style=flat-square)](LICENSE)
  [![CI](https://img.shields.io/github/actions/workflow/status/iPmartNetwork/VortexUI/ci.yml?style=flat-square&label=CI)](https://github.com/iPmartNetwork/VortexUI/actions)
  [![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](https://go.dev)
  [![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org)

  <br />

  <a href="README.md">English</a> · <strong>فارسی</strong>

  <br />

  [امکانات](#-امکانات) · [مقایسه](#-مقایسه) · [نصب سریع](#-نصب-سریع) · [پروتکل‌ها](#-پروتکل‌های-پشتیبانی‌شده) · [عملیات](#-عملیات) · [نقشه‌راه](#-نقشه‌راه)
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
- React 18 + TypeScript + Tailwind
- ۸ زبان (EN/FA/TR/AR/RU/ZH/JA/ES) + پشتیبانی RTL
- تم تیره (Navy Blue) و روشن (Pastel)
- واکنش‌گرا (کشوی موبایل)
- داشبورد بلادرنگ + به‌روزرسانی زنده (SSE)
- HTTPS خودکار با Caddy / Let's Encrypt

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

| | VortexUI | 3x-ui | Marzban | Hiddify |
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
| **VLESS** | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade |
| **VMess** | ✅ | ✅ | TCP, WS, gRPC |
| **Trojan** | ✅ | ✅ | TCP, WS, gRPC |
| **Shadowsocks** | ✅ | ✅ | TCP |
| **SOCKS** | — | ✅ | TCP |
| **HTTP** | — | ✅ | TCP |
| **Hysteria2 / TUIC / WireGuard** | 🔜 | — | UDP |

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
| مرجع API (OpenAPI 3.0) | [`docs/openapi.yaml`](docs/openapi.yaml) |
| متغیرهای محیطی | [`.env.example`](.env.example) |
| استقرار با Docker | [`deploy/`](deploy/) |
| CI/CD | [`.github/workflows/`](.github/workflows/) |

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

این پروژه تحت لایسنس **AGPL-3.0** منتشر شده است — فایل [LICENSE](LICENSE) را ببین.

---

<div align="center" dir="ltr">
  <sub>© 2026 iPmart Network. All rights reserved.</sub>
  <br />
  <strong>ساخته‌شده با ❤️ توسط <a href="https://github.com/iPmartNetwork">iPmart Network</a></strong>
</div>

</div>
