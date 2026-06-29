# نصب

!!! success "پیشنهادی"
    از **نصب‌کننده تک‌خطی** برای سریع‌ترین مسیر به یک پنل کارا استفاده کنید. این اسکریپت
    وابستگی‌ها، دیتابیس، HTTPS و سرویس‌های systemd را به‌صورت خودکار مدیریت می‌کند.

---

## پیش‌نیازها

| نیازمندی | حداقل | پیشنهادی |
|----------|-------|----------|
| سیستم‌عامل | Ubuntu 20.04 / Debian 11 | Ubuntu 22.04+ / Debian 12 |
| RAM | ۱ گیگابایت | ۲+ گیگابایت |
| دیسک | ۱۰ گیگابایت | ۲۰+ گیگابایت (TimescaleDB با داده ترافیک رشد می‌کند) |
| CPU | ۱ vCPU | ۲+ vCPU |
| Go (فقط بیلد بومی) | 1.26 | 1.26 |
| Docker (نصب کانتینری) | 24.0+ | آخرین نسخه پایدار |
| دامنه | اختیاری | پیشنهادی (برای HTTPS + سابسکریپشن‌ها) |

---

## نصب تک‌خطی

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

نصب‌کننده این کارها را انجام می‌دهد:

1. شناسایی سیستم‌عامل و معماری
2. نصب وابستگی‌ها (PostgreSQL, Redis, Caddy)
3. دانلود و بیلد VortexUI
4. اجرای مهاجرت‌های دیتابیس
5. ایجاد اکانت ادمین sudo (پرامپت تعاملی)
6. پیکربندی سرویس‌های systemd
7. راه‌اندازی HTTPS از طریق Caddy (در صورت ارائه دامنه)

پس از اتمام، از طریق `https://your-domain.com` یا `http://server-ip:8080` به پنل دسترسی پیدا کنید.

---

## Docker Compose

=== "شروع سریع"

    ```bash
    git clone https://github.com/iPmartNetwork/VortexUI.git
    cd VortexUI/deploy
    cp ../.env.example .env
    # فایل .env را با تنظیمات خود ویرایش کنید
    docker compose up -d
    ```

=== "پروداکشن (با HTTPS توسط Caddy)"

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

فایل `deploy/compose.yml` شامل: پنل، فرانتند وب، PostgreSQL + TimescaleDB، Redis و Caddy است.

---

## بیلد بومی

=== "Ubuntu/Debian"

    ```bash
    # نصب Go 1.26
    sudo snap install go --classic
    go version  # باید go1.26.x نمایش دهد

    # نصب وابستگی‌ها
    sudo apt update && sudo apt install -y postgresql redis-server

    # کلون و بیلد
    git clone https://github.com/iPmartNetwork/VortexUI.git
    cd VortexUI
    go build -o vortexui ./cmd/panel

    # اجرای مهاجرت‌ها
    ./vortexui migrate

    # ایجاد ادمین
    ./vortexui admin create --username admin --password your-password --sudo

    # اجرا
    ./vortexui serve
    ```

=== "سایر توزیع‌های لینوکس"

    ```bash
    # نصب Go 1.26 از تاربال رسمی
    wget https://go.dev/dl/go1.26.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.26.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin

    # سپس همان مراحل کلون/بیلد اوبونتو را دنبال کنید
    ```

!!! warning "نسخه Go"
    VortexUI به **Go 1.26** یا بالاتر نیاز دارد. نسخه‌های قدیمی‌تر قادر به کامپایل نخواهند بود.

---

## راه‌اندازی ایجنت نود

ایجنت نود روی سرورهای ریموت اجرا شده و از طریق gRPC + mTLS با پنل ارتباط برقرار می‌کند.

=== "ویزارد ثبت‌نام (پیشنهادی)"

    1. در رابط پنل، به **نودها → افزودن نود** بروید
    2. ویزارد ثبت‌نام یک دستور نصب تک‌خطی تولید می‌کند
    3. به سرور ریموت SSH کرده و دستور را paste کنید
    4. ایجنت به‌صورت خودکار ثبت‌نام می‌شود، گواهی‌ها را تبادل کرده و شروع به گزارش‌دهی می‌کند

=== "نصب دستی"

    ```bash
    # روی سرور ریموت
    bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install-node.sh)
    ```

    از شما پرسیده می‌شود:
    - آدرس پنل (مثلاً `https://panel.example.com`)
    - توکن ثبت‌نام نود (تولید شده در UI پنل)

=== "نود داکر"

    ```bash
    docker run -d --name vortex-node \
      -e PANEL_ADDR=https://panel.example.com \
      -e NODE_TOKEN=your-enrollment-token \
      --network host \
      ghcr.io/ipmartnetwork/vortexui-node:latest
    ```

---

## نود محلی (تک سرور)

اگر فقط به یک سرور نیاز دارید، از **نود محلی** استفاده کنید — هسته پروکسی درون‌پروسسی در کنار پنل اجرا می‌شود. نیازی به ایجنت جداگانه نیست.

1. در حین نصب، هنگام سؤال درباره نود محلی "بله" را انتخاب کنید
2. یا بعداً: **نودها → افزودن نود → محلی**
3. هسته را انتخاب کنید (Xray یا sing-box)
4. پنل پروسس هسته را مستقیماً مدیریت می‌کند

!!! tip
    نود محلی برای راه‌اندازی‌های تک‌سرور ایده‌آل است. برای استقرارهای چندسرور، از نودهای ریموت با ویزارد ثبت‌نام استفاده کنید.

---

## متغیرهای محیطی

| متغیر | توضیحات | مقدار پیش‌فرض |
|--------|---------|---------------|
| `VORTEX_DOMAIN` | دامنه پنل (برای HTTPS) | — |
| `VORTEX_LISTEN` | آدرس شنود API | `:8080` |
| `VORTEX_DB_URL` | رشته اتصال PostgreSQL | `postgres://localhost/vortex` |
| `VORTEX_REDIS_URL` | رشته اتصال Redis | `redis://localhost:6379/0` |
| `VORTEX_JWT_SECRET` | کلید امضای JWT (حداقل ۳۲ بایت) | — (ضروری) |
| `VORTEX_ADMIN_USER` | نام کاربری ادمین اولیه | — |
| `VORTEX_ADMIN_PASS` | رمز عبور ادمین اولیه | — |
| `VORTEX_TELEGRAM_TOKEN` | توکن بات تلگرام | — |
| `VORTEX_TELEGRAM_ADMIN` | شناسه چت ادمین برای اعلان‌ها | — |
| `VORTEX_ZARINPAL_MERCHANT` | شناسه مرچنت زرین‌پال | — |
| `VORTEX_NOWPAYMENTS_KEY` | کلید API نودپیمنتس | — |
| `VORTEX_NOWPAYMENTS_IPN_SECRET` | رمز HMAC IPN نودپیمنتس | — |
| `VORTEX_BACKUP_CRON` | زمان‌بندی بکاپ (عبارت cron) | — |
| `VORTEX_BACKUP_TELEGRAM` | ارسال بکاپ به تلگرام | `false` |
| `VORTEX_BACKUP_S3_BUCKET` | باکت S3 برای بکاپ‌ها | — |
| `VORTEX_METRICS_ENABLED` | فعال‌سازی متریک‌های Prometheus | `false` |
| `VORTEX_METRICS_LISTEN` | آدرس اندپوینت متریک‌ها | `:9090` |
| `VORTEX_SHARE_AUTOLIMIT` | محدودسازی خودکار در تشخیص اشتراک‌گذاری | `false` |

---

## مدیریت از طریق CLI

باینری `vortexui` یک منوی تعاملی ارائه می‌دهد:

```bash
vortexui
```

```
╔══════════════════════════════════════╗
║          VortexUI Management         ║
╠══════════════════════════════════════╣
║  1) Start panel                      ║
║  2) Stop panel                       ║
║  3) Restart panel                    ║
║  4) Status                           ║
║  5) Logs (live)                      ║
║  6) Update                           ║
║  7) Admin management                 ║
║  8) Backup                           ║
║  9) Doctor (diagnostics)             ║
║  0) Exit                             ║
╚══════════════════════════════════════╝
```

دستورات کلیدی:

| دستور | عملکرد |
|--------|--------|
| `vortexui update` | دریافت آخرین نسخه و ریستارت |
| `vortexui admin create` | ایجاد ادمین جدید |
| `vortexui admin reset-password` | بازنشانی رمز عبور ادمین |
| `vortexui backup` | ایجاد فوری بکاپ |
| `vortexui doctor` | اجرای تشخیص (دیتابیس، Redis، نودها، پورت‌ها) |
| `vortexui migrate` | اجرای مهاجرت‌های معلق دیتابیس |

---

## بروزرسانی

=== "بروزرسانی خودکار (پیشنهادی)"

    ```bash
    vortexui update
    ```

    آخرین نسخه را دریافت، بیلد، مهاجرت و ریستارت می‌کند.

=== "بروزرسانی دستی (سرور پنل)"

    ```bash
    cd /opt/VortexUI  # یا هر مسیری که clone کرده‌اید
    git pull origin master
    go build -o vortexui ./cmd/panel
    ./vortexui migrate
    sudo systemctl restart vortexui
    ```

=== "بروزرسانی دستی (سرورهای نود)"

    ```bash
    cd /opt/VortexUI-node
    git pull origin master
    go build -o vortex-node ./cmd/node
    sudo systemctl restart vortex-node
    ```

=== "بروزرسانی داکر"

    ```bash
    cd /opt/VortexUI/deploy
    docker compose pull
    docker compose up -d
    ```

---

## تأیید پس از نصب

پس از نصب، صحت عملکرد را بررسی کنید:

1. **دسترسی به پنل** — آدرس `https://your-domain.com` را در مرورگر باز کنید
2. **ورود** — با اعتبارنامه ادمین وارد شوید
3. **اتصال دیتابیس** — تنظیمات → اطلاعات سیستم را بررسی کنید
4. **نود آنلاین** — اگر نود محلی استفاده می‌کنید، وضعیت "آنلاین" را در صفحه نودها تأیید کنید
5. **اجرای تشخیص** — دستور `vortexui doctor` تمام اجزا را بررسی می‌کند

!!! tip "اندپوینت سلامت"
    پنل اندپوینت `GET /api/health` را ارائه می‌دهد — با `200 OK` و وضعیت اجزا پاسخ می‌دهد.
    از آن برای مانیتورینگ خارجی استفاده کنید (UptimeRobot، Prometheus blackbox و غیره).
