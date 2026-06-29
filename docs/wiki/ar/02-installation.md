# التثبيت

!!! success "موصى به"
    استخدم **المثبّت بأمر واحد** للحصول على أسرع مسار نحو لوحة تحكم عاملة. يتولى تلقائياً
    التبعيات، إعداد قاعدة البيانات، HTTPS، وخدمات systemd.

---

## المتطلبات الأساسية

| المتطلب | الحد الأدنى | الموصى به |
|---------|------------|-----------|
| نظام التشغيل | Ubuntu 20.04 / Debian 11 | Ubuntu 22.04+ / Debian 12 |
| الذاكرة | 1 جيجابايت | 2 جيجابايت أو أكثر |
| القرص | 10 جيجابايت | 20 جيجابايت أو أكثر (TimescaleDB ينمو مع بيانات الحركة) |
| المعالج | 1 vCPU | 2+ vCPU |
| Go (للبناء المحلي فقط) | 1.26 | 1.26 |
| Docker (للتثبيت بالحاويات) | 24.0+ | أحدث إصدار مستقر |
| نطاق | اختياري | موصى به (لـ HTTPS + الاشتراكات) |

---

## التثبيت بأمر واحد

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

سيقوم المثبّت بـ:

1. اكتشاف نظام التشغيل والبنية المعمارية
2. تثبيت التبعيات (PostgreSQL، Redis، Caddy)
3. تنزيل وبناء VortexUI
4. تشغيل ترحيلات قاعدة البيانات
5. إنشاء حساب مسؤول أعلى (طلب تفاعلي)
6. تكوين خدمات systemd
7. إعداد HTTPS عبر Caddy (في حال توفير نطاق)

بعد الاكتمال، يمكنك الوصول للوحة التحكم عبر `https://your-domain.com` أو `http://server-ip:8080`.

---

## Docker Compose

=== "بداية سريعة"

    ```bash
    git clone https://github.com/iPmartNetwork/VortexUI.git
    cd VortexUI/deploy
    cp ../.env.example .env
    # عدّل .env بإعداداتك
    docker compose up -d
    ```

=== "إنتاج (مع Caddy HTTPS)"

    ```bash
    git clone https://github.com/iPmartNetwork/VortexUI.git
    cd VortexUI/deploy
    cp ../.env.example .env
    ```

    عدّل `.env`:
    ```env
    VORTEX_DOMAIN=panel.example.com
    VORTEX_ADMIN_USER=admin
    VORTEX_ADMIN_PASS=your-secure-password
    VORTEX_JWT_SECRET=random-32-byte-string
    VORTEX_DB_URL=postgres://vortex:pass@db:5432/vortex?sslmode=disable
    VORTEX_REDIS_URL=redis://redis:6379/0
    ```

    ثم:
    ```bash
    docker compose up -d
    ```

يتضمن `deploy/compose.yml`: لوحة التحكم، واجهة الويب، PostgreSQL + TimescaleDB، Redis، و Caddy.

---

## البناء المحلي

=== "Ubuntu/Debian"

    ```bash
    # تثبيت Go 1.26
    sudo snap install go --classic
    go version  # يجب أن يُظهر go1.26.x

    # تثبيت التبعيات
    sudo apt update && sudo apt install -y postgresql redis-server

    # استنساخ وبناء
    git clone https://github.com/iPmartNetwork/VortexUI.git
    cd VortexUI
    go build -o vortexui ./cmd/panel

    # تشغيل الترحيلات
    ./vortexui migrate

    # إنشاء المسؤول
    ./vortexui admin create --username admin --password your-password --sudo

    # التشغيل
    ./vortexui serve
    ```

=== "توزيعات Linux الأخرى"

    ```bash
    # تثبيت Go 1.26 من الحزمة الرسمية
    wget https://go.dev/dl/go1.26.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.26.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin

    # ثم اتبع نفس خطوات الاستنساخ/البناء كـ Ubuntu
    ```

!!! warning "إصدار Go"
    VortexUI يتطلب **Go 1.26** أو أحدث. الإصدارات السابقة ستفشل في الترجمة.

---

## إعداد عميل العقدة

يعمل عميل العقدة على الخوادم البعيدة ويتواصل مع لوحة التحكم عبر gRPC + mTLS.

=== "معالج التسجيل (موصى به)"

    1. في واجهة لوحة التحكم، اذهب إلى **العقد → إضافة عقدة**
    2. يُنشئ معالج التسجيل أمر تثبيت بسطر واحد
    3. ادخل عبر SSH إلى خادمك البعيد والصق الأمر
    4. يسجّل العميل نفسه تلقائياً، يتبادل الشهادات، ويبدأ الإبلاغ

=== "تثبيت يدوي"

    ```bash
    # على الخادم البعيد
    bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install-node.sh)
    ```

    سيُطلب منك:
    - عنوان لوحة التحكم (مثل `https://panel.example.com`)
    - توكن تسجيل العقدة (يُنشأ في واجهة لوحة التحكم)

=== "عقدة Docker"

    ```bash
    docker run -d --name vortex-node \
      -e PANEL_ADDR=https://panel.example.com \
      -e NODE_TOKEN=your-enrollment-token \
      --network host \
      ghcr.io/ipmartnetwork/vortexui-node:latest
    ```

---

## العقدة المحلية (خادم واحد)

إذا كنت تحتاج خادماً واحداً فقط، استخدم **العقدة المحلية** — تعمل نواة البروكسي داخل العملية بجانب لوحة التحكم. لا حاجة لعميل منفصل.

1. أثناء التثبيت، اختر "نعم" عند السؤال عن العقدة المحلية
2. أو لاحقاً: **العقد → إضافة عقدة → محلية**
3. اختر النواة (Xray أو sing-box)
4. تدير لوحة التحكم عملية النواة مباشرة

!!! tip
    العقدة المحلية مثالية لإعدادات الخادم الواحد. لنشرات متعددة الخوادم، استخدم العقد البعيدة مع معالج التسجيل.

---

## متغيرات البيئة

| المتغير | الوصف | القيمة الافتراضية |
|---------|-------|-----------------|
| `VORTEX_DOMAIN` | نطاق لوحة التحكم (لـ HTTPS) | — |
| `VORTEX_LISTEN` | عنوان استماع API | `:8080` |
| `VORTEX_DB_URL` | سلسلة اتصال PostgreSQL | `postgres://localhost/vortex` |
| `VORTEX_REDIS_URL` | سلسلة اتصال Redis | `redis://localhost:6379/0` |
| `VORTEX_JWT_SECRET` | مفتاح توقيع JWT (≥32 بايت) | — (مطلوب) |
| `VORTEX_ADMIN_USER` | اسم المسؤول الأولي | — |
| `VORTEX_ADMIN_PASS` | كلمة مرور المسؤول الأولية | — |
| `VORTEX_TELEGRAM_TOKEN` | توكن بوت تيليجرام | — |
| `VORTEX_TELEGRAM_ADMIN` | معرّف محادثة المسؤول للإشعارات | — |
| `VORTEX_ZARINPAL_MERCHANT` | معرّف تاجر ZarinPal | — |
| `VORTEX_NOWPAYMENTS_KEY` | مفتاح API لـ NowPayments | — |
| `VORTEX_NOWPAYMENTS_IPN_SECRET` | سر HMAC لـ NowPayments IPN | — |
| `VORTEX_BACKUP_CRON` | جدول النسخ الاحتياطي (تعبير cron) | — |
| `VORTEX_BACKUP_TELEGRAM` | إرسال النسخ الاحتياطية إلى تيليجرام | `false` |
| `VORTEX_BACKUP_S3_BUCKET` | حاوية S3 للنسخ الاحتياطي | — |
| `VORTEX_METRICS_ENABLED` | تفعيل مقاييس Prometheus | `false` |
| `VORTEX_METRICS_LISTEN` | عنوان نقطة نهاية المقاييس | `:9090` |
| `VORTEX_SHARE_AUTOLIMIT` | تقييد تلقائي عند كشف مشاركة الحساب | `false` |

---

## إدارة CLI

يوفر ملف `vortexui` التنفيذي قائمة تفاعلية:

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

الأوامر الرئيسية:

| الأمر | الإجراء |
|-------|---------|
| `vortexui update` | سحب أحدث إصدار وإعادة التشغيل |
| `vortexui admin create` | إنشاء مسؤول جديد |
| `vortexui admin reset-password` | إعادة تعيين كلمة مرور المسؤول |
| `vortexui backup` | إنشاء نسخة احتياطية فورية |
| `vortexui doctor` | تشغيل التشخيصات (قاعدة بيانات، Redis، العقد، المنافذ) |
| `vortexui migrate` | تشغيل ترحيلات قاعدة البيانات المعلّقة |

---

## التحديث

=== "تحديث تلقائي (موصى به)"

    ```bash
    vortexui update
    ```

    يسحب أحدث إصدار، يعيد البناء، يشغّل الترحيلات، ويعيد التشغيل.

=== "تحديث يدوي (خادم لوحة التحكم)"

    ```bash
    cd /opt/VortexUI  # أو أينما استنسخت
    git pull origin master
    go build -o vortexui ./cmd/panel
    ./vortexui migrate
    sudo systemctl restart vortexui
    ```

=== "تحديث يدوي (خوادم العقد)"

    ```bash
    cd /opt/VortexUI-node
    git pull origin master
    go build -o vortex-node ./cmd/node
    sudo systemctl restart vortex-node
    ```

=== "تحديث Docker"

    ```bash
    cd /opt/VortexUI/deploy
    docker compose pull
    docker compose up -d
    ```

---

## التحقق بعد التثبيت

بعد التثبيت، تحقق من أن كل شيء يعمل:

1. **لوحة التحكم متاحة** — افتح `https://your-domain.com` في المتصفح
2. **تسجيل الدخول يعمل** — سجّل الدخول ببيانات المسؤول
3. **قاعدة البيانات متصلة** — تحقق من الإعدادات → معلومات النظام
4. **العقدة متصلة** — إذا كنت تستخدم العقدة المحلية، تحقق أنها تظهر "متصلة" في صفحة العقد
5. **تشغيل التشخيصات** — `vortexui doctor` يفحص جميع المكونات

!!! tip "نقطة نهاية الصحة"
    تكشف لوحة التحكم `GET /api/health` — تُرجع `200 OK` مع حالة المكونات.
    استخدمها للمراقبة الخارجية (UptimeRobot، Prometheus blackbox، إلخ).
