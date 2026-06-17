# ۲. نصب و راه‌اندازی

!!! important
    قبل نصب، پورت‌های **80 و 443** (برای HTTPS) و DNS دامنه را آماده کنید.

---

## پیش‌نیازها

| مورد | Docker (پیشنهادی) | Native |
|------|:-----------------:|:------:|
| سیستم‌عامل | Linux (Ubuntu 22.04+) | Linux |
| RAM | حداقل ۲ GB | ۲ GB+ |
| CPU | ۱ vCPU | ۱+ |
| دیسک | ۱۰ GB | ۱۰ GB |
| Docker + Compose v2 | ✅ | فقط برای DB/Redis |
| Go 1.26 | — | ✅ (نصب خودکار) |
| پورت‌ها | 80, 443 (+ inboundها) | همان |

---

## روش ۱: نصب یک‌خطی (پیشنهادی)

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

اسکript نصب **تعاملی** است و دو سوال اصلی می‌پرسد:

### ۱. روش نصب

| گزینه | توضیح |
|-------|-------|
| **Docker Compose** *(پیشنهادی)* | کل stack در container: web · panel · node · PostgreSQL · Redis |
| **Native (systemd)** | باینری Go به‌صورت سرویس؛ DB/Redis در Docker؛ SPA با Caddy |

### ۲. دسترسی به پنل

| گزینه | توضیح |
|-------|-------|
| **دامنه + HTTPS** | Caddy گواهی Let's Encrypt می‌گیرد — پورت 80 و 443 باید باز باشد |
| **IP + HTTP** | پورت دلخواه (مثلاً 8080) |

### نصب غیرتعاملی (اسکریپت/CI)

```bash
VORTEXUI_METHOD=docker \
VORTEXUI_NONINTERACTIVE=1 \
VORTEXUI_ADMIN_USER=admin \
VORTEXUI_ADMIN_PASS='رمز-قوی-شما' \
bash install.sh
```

### خروجی نصب

- مسیر نصب: `/opt/vortexui` (قابل تغییر با `VORTEXUI_DIR`)
- دستور `vortexui` در `/usr/local/bin`
- فایل env: `deploy/.env` (JWT، DB password، domain)
- گواهی mTLS: `deploy/certs/`
- URL پنل + اطلاعات ادمین اولیه در ترمینال چاپ می‌شود

---

## روش ۲: Docker Compose دستی

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

# تولید secret
echo "JWT_SECRET=$(openssl rand -hex 32)" >> deploy/.env
echo "DB_PASSWORD=$(openssl rand -hex 16)" >> deploy/.env
echo "SITE_ADDRESS=panel.example.com" >> deploy/.env
echo "ACME_EMAIL=admin@example.com" >> deploy/.env

make certs
docker compose --env-file deploy/.env -f deploy/compose.yml up -d --build

# ساخت ادمین
docker compose -f deploy/compose.yml exec panel \
  /usr/local/bin/panel admin create --username admin --password 'change-me' --sudo
```

### سرویس‌های stack

| سرویس | نقش |
|-------|-----|
| `db` | PostgreSQL 16 + TimescaleDB |
| `redis` | Redis 7 |
| `panel` | API + local node (host network) |
| `web` | Caddy + SPA (HTTPS) |

---

## روش ۳: نصب Native (توسعه/پیشرفته)

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

docker compose up -d          # PostgreSQL + Redis
cp .env.example .env
# VORTEX_JWT_SECRET را با openssl rand -hex 32 پر کنید

make build
make certs
make run-panel

# ترمینال دیگر — ساخت ادمین
./bin/panel admin create --username admin --password 'your-password' --sudo
```

فرانت‌اند (توسعه):

```bash
cd web && npm install && npm run dev
```

---

## نصب Node Agent (چند سرور)

برای fleet چند‌نودی، روی هر سرور جداگانه:

```bash
VORTEX_NODE_LISTEN=:50051 \
VORTEX_CORE=xray \
VORTEX_CORE_BIN=/usr/local/bin/xray \
VORTEX_TLS_CERT=node.crt \
VORTEX_TLS_KEY=node.key \
VORTEX_TLS_CA=ca.crt \
./bin/node
```

سپس در پنل: **Nodes → Add Node** — آدرس و گواهی mTLS را ثبت کنید.

---

## نود محلی (Local Node)

برای سرور تک‌نودی، بدون agent جدا:

```env
VORTEX_LOCAL_NODE=true
VORTEX_LOCAL_NODE_NAME=local
VORTEX_LOCAL_NODE_HOST=your-public-ip-or-domain
VORTEX_CORE=xray
VORTEX_CORE_BIN=/usr/local/bin/xray
```

در Docker Compose این به‌صورت پیش‌فرض فعال است.

---

## متغیرهای محیطی مهم

| متغیر | پیش‌فرض | توضیح |
|-------|---------|-------|
| `VORTEX_HTTP_ADDR` | `:8080` | آدرس HTTP پنل |
| `VORTEX_DATABASE_URL` | — | **الزامی** — PostgreSQL |
| `VORTEX_JWT_SECRET` | — | **الزامی** — حداقل ۳۲ بایت |
| `VORTEX_REDIS_URL` | `redis://localhost:6379/0` | Redis |
| `VORTEX_LOCAL_NODE` | `false` | نود in-process |
| `VORTEX_SHARE_AUTOLIMIT` | `false` | محدودیت خودکار در اشتراک‌گذاری |
| `VORTEX_WEBHOOK_URL` | — | Webhook اعلان‌ها |
| `VORTEX_TELEGRAM_TOKEN` | — | توکن ربات تلگرام |
| `VORTEX_CF_API_TOKEN` | — | Cloudflare DNS automation |

لیست کامل: [`.env.example`](../../../.env.example)

---

## بررسی سلامت پس از نصب

```bash
vortexui status
curl -s http://127.0.0.1:8080/api/health
```

پاسخ مورد انتظار: `{"status":"ok"}`

---

## به‌روزرسانی

```bash
vortexui update
# یا
cd /opt/vortexui && git pull && docker compose -f deploy/compose.yml up -d --build
```

نصب مجدد اسکript **ایمن** است — secretها و داده DB حفظ می‌شوند.
