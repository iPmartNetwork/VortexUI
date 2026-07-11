# نصب

> **توصیه:** از **نصب‌کننده یک‌خطی** برای سریع‌ترین راه به یک پنل کارا استفاده کنید. این نصب‌کننده وابستگی‌ها، تنظیم دیتابیس، HTTPS و سرویس‌های systemd را به صورت خودکار مدیریت می‌کند.

---

## پیش‌نیازها

| نیازمندی | حداقل | توصیه‌شده |
|----------|-------|-----------|
| سیستم‌عامل | Ubuntu 20.04 / Debian 11 | Ubuntu 22.04+ / Debian 12 |
| RAM | 1 GB | 2 GB+ |
| دیسک | 10 GB | 20 GB+ (TimescaleDB با داده ترافیک رشد می‌کند) |
| CPU | 1 vCPU | 2+ vCPU |
| Go (فقط برای بیلد بومی) | 1.26 | 1.26 |
| Docker (نصب کانتینری) | 24.0+ | آخرین نسخه پایدار |
| دامنه | اختیاری | توصیه‌شده (برای HTTPS + اشتراک‌ها) |

---

## نصب یک‌خطی

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

نصب‌کننده:

1. سیستم‌عامل و معماری شما را تشخیص می‌دهد
2. وابستگی‌ها را نصب می‌کند (PostgreSQL, Redis, Caddy)
3. VortexUI را دانلود و بیلد می‌کند
4. migration‌های دیتابیس را اجرا می‌کند
5. یک اکانت ادمین sudo ایجاد می‌کند (prompt تعاملی)
6. سرویس‌های systemd را پیکربندی می‌کند
7. HTTPS را از طریق Caddy تنظیم می‌کند (اگر دامنه ارائه شود)

پس از تکمیل، به پنل در `https://your-domain.com` یا `http://server-ip:8080` دسترسی پیدا کنید.

---

## داکر کامپوز

### شروع سریع

```bash
git clone https://github.com/iPmartNetwork/VortexUI.git
cd VortexUI/deploy
cp ../.env.example .env
# فایل .env را با تنظیمات خود ویرایش کنید
docker compose up -d
```

### پروداکشن (با Caddy HTTPS)

```bash
git clone https://github.com/iPmartNetwork/VortexUI.git
cd VortexUI/deploy
cp ../.env.example .env
```

فایل `.env` را ویرایش کنید:

```env
VORTEX_DOMAIN=panel.example.com
VORTEX_ADMIN_USER=admin
VORTEX_ADMIN_PASS=your-secure-password
VORTEX_JWT_SECRET=random-32-byte-string
VORTEX_DB_URL=postgres://vortex:pass@db:5432/vortex?sslmode=disable
VORTEX_REDIS_URL=redis://redis:6379/0
```

سپس:

```bash
docker compose up -d
```

فایل `deploy/compose.yml` شامل: پنل، فرانت‌اند وب، PostgreSQL + TimescaleDB، Redis و Caddy است.

---

## بیلد بومی

### اوبونتو/دبیان

```bash
# نصب Go 1.26
sudo snap install go --classic
go version  # باید go1.26.x را نشان دهد

# نصب وابستگی‌ها
sudo apt update && sudo apt install -y postgresql redis-server

# کلون و بیلد
git clone https://github.com/iPmartNetwork/VortexUI.git
cd VortexUI
go build -o vortexui ./cmd/panel

# اجرای migration‌ها
./vortexui migrate

# ایجاد ادمین
./vortexui admin create --username admin --password your-password --sudo

# شروع
./vortexui serve
```

### سایر لینوکس‌ها

```bash
# نصب Go 1.26 از tarball رسمی
wget https://go.dev/dl/go1.26.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# سپس همان مراحل کلون/بیلد اوبونتو را دنبال کنید
```

> **توجه:** VortexUI به **Go 1.26** یا بالاتر نیاز دارد. نسخه‌های قبلی کامپایل نخواهند شد.

---

## راه‌اندازی Node Agent

Node agent روی سرورهای ریموت اجرا می‌شود و از طریق gRPC + mTLS با پنل ارتباط برقرار می‌کند.

### ویزارد ثبت‌نام (توصیه‌شده)

1. در رابط پنل، به **Nodes → Add Node** بروید
2. ویزارد ثبت‌نام یک دستور نصب یک‌خطی تولید می‌کند
3. SSH به سرور ریموت خود زده و دستور را paste کنید
4. Agent به صورت خودکار ثبت‌نام می‌شود، گواهی‌ها را تبادل می‌کند و شروع به گزارش می‌کند

### نصب دستی

```bash
# روی سرور ریموت
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install-node.sh)
```

از شما خواسته می‌شود:
- آدرس پنل (مثلاً `https://panel.example.com`)
- توکن ثبت‌نام نود (در رابط پنل تولید می‌شود)

### Docker Node

```bash
docker run -d --name vortex-node \
  -e PANEL_ADDR=https://panel.example.com \
  -e NODE_TOKEN=your-enrollment-token \
  --network host \
  ghcr.io/ipmartnetwork/vortexui-node:latest
```

---

## نود محلی (تک سرور)

اگر فقط به یک سرور نیاز دارید، از **نود محلی** استفاده کنید — هسته پروکسی در پروسه کنار پنل اجرا می‌شود. نیازی به agent جداگانه نیست.

1. در حین نصب، وقتی درباره نود محلی سوال شد "Yes" را انتخاب کنید
2. یا بعداً: **Nodes → Add Node → Local**
3. هسته را انتخاب کنید (Xray یا sing-box)
4. پنل مستقیماً پروسه هسته را مدیریت می‌کند

> **نکته:** نود محلی برای تنظیمات تک‌سرور عالی است. برای استقرارهای چندسرور، از نودهای ریموت با ویزارد ثبت‌نام استفاده کنید.

---

## متغیرهای محیطی

| متغیر | توضیحات | پیش‌فرض |
|-------|---------|---------|
| VORTEX_DOMAIN | دامنه پنل (برای HTTPS) | — |
| VORTEX_LISTEN | آدرس شنود API | :8080 |
| VORTEX_DB_URL | رشته اتصال PostgreSQL | postgres://localhost/vortex |
| VORTEX_REDIS_URL | رشته اتصال Redis | redis://localhost:6379/0 |
| VORTEX_JWT_SECRET | کلید امضای JWT (≥32 بایت) | — (الزامی) |
| VORTEX_ADMIN_USER | نام‌کاربری ادمین اولیه | — |
| VORTEX_ADMIN_PASS | رمزعبور ادمین اولیه | — |
| VORTEX_TELEGRAM_TOKEN | توکن ربات تلگرام | — |
| VORTEX_TELEGRAM_ADMIN | شناسه چت ادمین برای اعلان‌ها | — |
| VORTEX_ZARINPAL_MERCHANT | شناسه مرچنت زرین‌پال | — |
| VORTEX_BACKUP_CRON | زمان‌بندی پشتیبان‌گیری (cron expression) | — |
| VORTEX_BACKUP_TELEGRAM | ارسال پشتیبان به تلگرام | false |
| VORTEX_METRICS_ENABLED | فعال‌سازی متریک‌های Prometheus | false |

---

## مدیریت CLI

باینری `vortexui` یک منوی تعاملی فراهم می‌کند:

```bash
vortexui
```

```
╔══════════════════════════════════════╗
║          مدیریت VortexUI             ║
╠══════════════════════════════════════╣
║  1) شروع پنل                         ║
║  2) توقف پنل                         ║
║  3) ری‌استارت پنل                     ║
║  4) وضعیت                            ║
║  5) لاگ‌ها (زنده)                     ║
║  6) به‌روزرسانی                       ║
║  7) مدیریت ادمین                     ║
║  8) پشتیبان‌گیری                      ║
║  9) دکتر (تشخیص)                     ║
║  0) خروج                             ║
╚══════════════════════════════════════╝
```

دستورات کلیدی:

| دستور | عملکرد |
|-------|--------|
| vortexui update | دریافت آخرین نسخه و ری‌استارت |
| vortexui admin create | ایجاد ادمین جدید |
| vortexui admin reset-password | بازنشانی رمزعبور ادمین |
| vortexui backup | ایجاد پشتیبان فوری |
| vortexui doctor | اجرای تشخیص (DB, Redis, nodes, ports) |
| vortexui migrate | اجرای migration‌های معلق دیتابیس |

---

## به‌روزرسانی

### به‌روزرسانی خودکار (توصیه‌شده)

```bash
vortexui update
```

این آخرین نسخه را دریافت، بازسازی، migration‌ها را اجرا و ری‌استارت می‌کند.

### به‌روزرسانی دستی (سرور پنل)

```bash
cd /opt/VortexUI  # یا هرجا که کلون کردید
git pull origin master
go build -o vortexui ./cmd/panel
./vortexui migrate
sudo systemctl restart vortexui
```

### به‌روزرسانی داکر

```bash
cd /opt/VortexUI/deploy
docker compose pull
docker compose up -d
```

---

## تأیید پس از نصب

پس از نصب، تأیید کنید که همه چیز کار می‌کند:

1. **دسترسی به پنل** — `https://your-domain.com` را در مرورگر باز کنید
2. **ورود کار می‌کند** — با اعتبارنامه ادمین خود وارد شوید
3. **دیتابیس متصل** — Settings → System Info را بررسی کنید
4. **نود آنلاین** — اگر از نود محلی استفاده می‌کنید، تأیید کنید که در صفحه Nodes "Online" نشان می‌دهد
5. **اجرای تشخیص** — `vortexui doctor` تمام اجزا را بررسی می‌کند

### اندپوینت سلامت

پنل `GET /api/health` را ارائه می‌دهد — با `200 OK` و وضعیت اجزا پاسخ می‌دهد. از این برای مانیتورینگ خارجی (UptimeRobot, Prometheus blackbox و غیره) استفاده کنید.
