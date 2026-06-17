<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](../README.md) · [FA](../fa/02-installation.md) · [EN](../en/02-installation.md) · [TR](../tr/02-installation.md)

</div>

<div dir="rtl">

# ٢. التثبيت

[← المقدمة](./01-introduction.md) · [الفهرس](./README.md) · [التالي: الخطوات الأولى →](./03-first-steps.md)

> [!IMPORTANT]
> قبل التثبيت، جهّز **المنافذ 80 و 443** (لـ HTTPS) وDNS النطاق.

---

## المتطلبات الأساسية

| العنصر | Docker (موصى به) | Native |
|------|:--------------------:|:------:|
| نظام التشغيل | Linux (Ubuntu 22.04+) | Linux |
| RAM | 2 GB كحد أدنى | 2 GB+ |
| CPU | 1 vCPU | 1+ |
| القرص | 10 GB | 10 GB |
| Docker + Compose v2 | ✅ | DB/Redis فقط |
| Go 1.26 | — | ✅ (تثبيت تلقائي) |
| المنافذ | 80, 443 (+ inbounds) | نفسها |

---

## الطريقة 1: تثبيت بسطر واحد (موصى به)

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

سكربت التثبيت **تفاعلي** ويسأل سؤالين رئيسيين:

### 1. طريقة التثبيت

| الخيار | الوصف |
|--------|-------------|
| **Docker Compose** *(موصى به)* | المكدس الكامل في حاويات: web · panel · node · PostgreSQL · Redis |
| **Native (systemd)** | ثنائي Go كخدمة؛ DB/Redis في Docker؛ SPA عبر Caddy |

### 2. الوصول إلى اللوحة

| الخيار | الوصف |
|--------|-------------|
| **نطاق + HTTPS** | Caddy يحصل على شهادة Let's Encrypt — يجب فتح المنافذ 80 و 443 |
| **IP + HTTP** | منفذ مخصص (مثل 8080) |

### تثبيت غير تفاعلي (سكربت/CI)

```bash
VORTEXUI_METHOD=docker \
VORTEXUI_NONINTERACTIVE=1 \
VORTEXUI_ADMIN_USER=admin \
VORTEXUI_ADMIN_PASS='your-strong-password' \
bash install.sh
```

### مخرجات التثبيت

- مسار التثبيت: `/opt/vortexui` (تجاوز بـ `VORTEXUI_DIR`)
- أمر `vortexui` في `/usr/local/bin`
- ملف البيئة: `deploy/.env` (JWT، كلمة مرور DB، النطاق)
- شهادات mTLS: `deploy/certs/`
- عنوان URL للوحة + بيانات اعتماد المسؤول الأولية تُطبع في الطرفية

---

## الطريقة 2: Docker Compose يدوياً

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

# Generate secrets
echo "JWT_SECRET=$(openssl rand -hex 32)" >> deploy/.env
echo "DB_PASSWORD=$(openssl rand -hex 16)" >> deploy/.env
echo "SITE_ADDRESS=panel.example.com" >> deploy/.env
echo "ACME_EMAIL=admin@example.com" >> deploy/.env

make certs
docker compose --env-file deploy/.env -f deploy/compose.yml up -d --build

# Create admin
docker compose -f deploy/compose.yml exec panel \
  /usr/local/bin/panel admin create --username admin --password 'change-me' --sudo
```

### خدمات المكدس

| الخدمة | الدور |
|---------|------|
| `db` | PostgreSQL 16 + TimescaleDB |
| `redis` | Redis 7 |
| `panel` | API + local node (host network) |
| `web` | Caddy + SPA (HTTPS) |

---

## الطريقة 3: تثبيت Native (تطوير/متقدم)

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

docker compose up -d          # PostgreSQL + Redis
cp .env.example .env
# Fill VORTEX_JWT_SECRET with: openssl rand -hex 32

make build
make certs
make run-panel

# Another terminal — create admin
./bin/panel admin create --username admin --password 'your-password' --sudo
```

الواجهة الأمامية (التطوير):

```bash
cd web && npm install && npm run dev
```

---

## تثبيت Node Agent (خوادم متعددة)

لأسطول متعدد العقد، على كل خادم منفصل:

```bash
VORTEX_NODE_LISTEN=:50051 \
VORTEX_CORE=xray \
VORTEX_CORE_BIN=/usr/local/bin/xray \
VORTEX_TLS_CERT=node.crt \
VORTEX_TLS_KEY=node.key \
VORTEX_TLS_CA=ca.crt \
./bin/node
```

ثم في اللوحة: **Nodes → Add Node** — سجّل العنوان وشهادة mTLS.

---

## Local Node

لإعداد خادم واحد بدون agent منفصل:

```env
VORTEX_LOCAL_NODE=true
VORTEX_LOCAL_NODE_NAME=local
VORTEX_LOCAL_NODE_HOST=your-public-ip-or-domain
VORTEX_CORE=xray
VORTEX_CORE_BIN=/usr/local/bin/xray
```

في Docker Compose هذا مفعّل افتراضياً.

---

## متغيرات البيئة المهمة

| المتغير | الافتراضي | الوصف |
|----------|---------|-------------|
| `VORTEX_HTTP_ADDR` | `:8080` | عنوان HTTP للوحة |
| `VORTEX_DATABASE_URL` | — | **مطلوب** — PostgreSQL |
| `VORTEX_JWT_SECRET` | — | **مطلوب** — 32 بايت كحد أدنى |
| `VORTEX_REDIS_URL` | `redis://localhost:6379/0` | Redis |
| `VORTEX_LOCAL_NODE` | `false` | عقدة in-process |
| `VORTEX_SHARE_AUTOLIMIT` | `false` | حد تلقائي عند مشاركة الحساب |
| `VORTEX_WEBHOOK_URL` | — | webhook للإشعارات |
| `VORTEX_TELEGRAM_TOKEN` | — | رمز بوت Telegram |
| `VORTEX_CF_API_TOKEN` | — | أتمتة DNS Cloudflare |

القائمة الكاملة: [`.env.example`](../../../.env.example)

---

## فحص الصحة بعد التثبيت

```bash
vortexui status
curl -s http://127.0.0.1:8080/api/health
```

الاستجابة المتوقعة: `{"status":"ok"}`

---

## التحديث

```bash
vortexui update
# or
cd /opt/vortexui && git pull && docker compose -f deploy/compose.yml up -d --build
```

إعادة تشغيل سكربت التثبيت **آمنة** — الأسرار وبيانات DB محفوظة.

</div>
