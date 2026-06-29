# عملیات و نگهداری

---

## HTTPS با Caddy

VortexUI از Caddy به‌عنوان وب‌سرور برای HTTPS خودکار استفاده می‌کند:

- **Let's Encrypt خودکار** — گواهی‌ها بدون تنظیم صادر و تمدید می‌شوند
- **ریدایرکت HTTP → HTTPS** — تمام ترافیک به HTTPS اجبار می‌شود
- **مسیریابی SPA** — فرانتند React به‌درستی سرو می‌شود
- **پروکسی معکوس** — درخواست‌های API به بکند Go فوروارد می‌شوند
- **اندپوینت DoH** — `/dns-query` در صورت فعال بودن DoH سرو می‌شود

### ساختار Caddyfile

```
{$VORTEX_DOMAIN} {
    encode gzip
    handle /api/* {
        reverse_proxy localhost:8080
    }
    handle /sub/* {
        reverse_proxy localhost:8080
    }
    handle /dns-query {
        reverse_proxy localhost:8053
    }
    handle {
        root * /opt/vortexui/web/dist
        try_files {path} /index.html
        file_server
    }
}
```

!!! tip
    اگر تنظیمات TLS سفارشی نیاز دارید (مثلاً گواهی کلاینت یا cipher suite خاص)، فایل `deploy/Caddyfile` را ویرایش کنید.

---

## متریک‌های Prometheus

**فعال‌سازی:** متغیر `VORTEX_METRICS_ENABLED=true` و `VORTEX_METRICS_LISTEN=:9090` را تنظیم کنید.

### متریک‌های موجود

| متریک | نوع | توضیحات |
|--------|-----|---------|
| `vortex_users_total` | Gauge | تعداد کل کاربران به تفکیک وضعیت |
| `vortex_traffic_bytes` | Counter | بایت‌های ترافیک (آپلود/دانلود) |
| `vortex_connections_active` | Gauge | اتصالات فعال فعلی |
| `vortex_nodes_status` | Gauge | وضعیت نود (1=آنلاین، 0=آفلاین) |
| `vortex_node_cpu_percent` | Gauge | درصد استفاده CPU نود |
| `vortex_node_memory_percent` | Gauge | درصد استفاده حافظه نود |
| `vortex_orders_total` | Counter | سفارشات به تفکیک وضعیت |
| `vortex_api_requests_total` | Counter | درخواست‌های API به تفکیک اندپوینت/وضعیت |
| `vortex_api_latency_seconds` | Histogram | تأخیر پاسخ API |

### داشبورد Grafana

داشبورد آماده را از `deploy/grafana/vortexui-dashboard.json` ایمپورت کنید:

1. Grafana → Dashboards → Import را باز کنید
2. فایل JSON را آپلود کنید یا محتوای آن را paste کنید
3. منبع داده Prometheus خود را انتخاب کنید
4. داشبورد نمایش می‌دهد: روند ترافیک، سلامت نود، فعالیت کاربران، عملکرد API

---

## مقیاس‌پذیری و دسترسی‌پذیری بالا

### مقیاس‌پذیری افقی

برای استقرارهای بزرگ:

| کامپوننت | استراتژی مقیاس‌پذیری |
|----------|---------------------|
| API پنل | اجرای چند نمونه پشت بالانسر |
| دیتابیس | PostgreSQL با رپلیکاهای خوانش |
| Redis | Redis Cluster یا Sentinel |
| نودها | هر تعداد لازم اضافه کنید (مدیریت خودکار) |
| فدراسیون | اتصال چند پنل برای توزیع منطقه‌ای |

### حالت کلاستر

اجرای چند نمونه پنل:

1. همه نمونه‌ها PostgreSQL و Redis مشترک دارند
2. از بالانسر (Caddy، nginx، HAProxy) در جلو استفاده کنید
3. اتصالات SSE چسبنده هستند (session affinity پیشنهادی)
4. جاب‌های پس‌زمینه از قفل‌های توزیع‌شده مبتنی بر Redis استفاده می‌کنند

### ملاحظات دیتابیس

| نگرانی | راه‌حل |
|--------|--------|
| استخر اتصال | pgBouncer جلوی PostgreSQL |
| مقیاس خوانش | رپلیکاهای خوانش برای کوئری‌های تحلیلی |
| توان نوشتن | هایپرتیبل‌های TimescaleDB برای داده ترافیک |
| بکاپ | pg_dump + آرشیو WAL برای بازیابی نقطه‌ای |

---

## نگهداری دیتابیس

### نکات TimescaleDB

داده ترافیک در هایپرتیبل‌های TimescaleDB برای کوئری‌های سری‌زمانی کارا ذخیره می‌شود:

```sql
-- بررسی اندازه chunk‌ها
SELECT show_chunks('traffic_data');

-- تنظیم سیاست نگهداری (نگه‌داشتن ۹۰ روز)
SELECT add_retention_policy('traffic_data', INTERVAL '90 days');

-- فشرده‌سازی chunk‌های قدیمی
SELECT add_compression_policy('traffic_data', INTERVAL '7 days');
```

### نگهداری روتین

```bash
# Vacuum و Analyze
psql -U vortex -d vortex -c "VACUUM ANALYZE;"

# بررسی حجم دیتابیس
psql -U vortex -d vortex -c "SELECT pg_size_pretty(pg_database_size('vortex'));"

# لیست بزرگ‌ترین جداول
psql -U vortex -d vortex -c "
  SELECT relname, pg_size_pretty(pg_relation_size(oid))
  FROM pg_class
  ORDER BY pg_relation_size(oid) DESC
  LIMIT 10;
"
```

### مهاجرت‌ها

مهاجرت‌ها هنگام شروع پنل به‌صورت خودکار اجرا می‌شوند. برای اجرای دستی:

```bash
vortexui migrate
```

بررسی وضعیت مهاجرت:

```bash
vortexui migrate status
```

---

## مدیریت لاگ

### لاگ‌های پنل

```bash
# لاگ زنده
journalctl -u vortexui -f

# ۱۰۰ خط آخر
journalctl -u vortexui -n 100

# فیلتر بر اساس سطح
journalctl -u vortexui | grep ERROR
```

### لاگ‌های ایجنت نود

```bash
journalctl -u vortex-node -f
```

### سطوح لاگ

از طریق متغیر محیطی تنظیم کنید:

```
VORTEX_LOG_LEVEL=info   # debug, info, warn, error
VORTEX_LOG_FORMAT=json  # json یا text
```

### چرخش لاگ

ژورنال systemd به‌صورت پیش‌فرض چرخش را مدیریت می‌کند. محدودیت‌ها را تنظیم کنید:

```ini
# /etc/systemd/journald.conf
SystemMaxUse=500M
MaxRetentionSec=30d
```

---

## تنظیم عملکرد

### محدودیت‌های سیستم

```bash
# افزایش محدودیت file descriptor
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf

# افزایش برای سرویس systemd
# در /etc/systemd/system/vortexui.service
[Service]
LimitNOFILE=65535
```

### تنظیم شبکه

```bash
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
```

اعمال: `sysctl -p`

### تنظیم Redis

```
# /etc/redis/redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
```

### تنظیم PostgreSQL

تنظیمات کلیدی برای بار کاری VortexUI:

```
# postgresql.conf
shared_buffers = 512MB          # ۲۵% رم
effective_cache_size = 1536MB   # ۷۵% رم
work_mem = 16MB
maintenance_work_mem = 128MB
random_page_cost = 1.1          # SSD
```

---

## سرویس‌های Systemd

### سرویس پنل

```ini
# /etc/systemd/system/vortexui.service
[Unit]
Description=VortexUI Panel
After=postgresql.service redis.service

[Service]
Type=simple
User=vortex
WorkingDirectory=/opt/vortexui
ExecStart=/opt/vortexui/vortexui serve
Restart=always
RestartSec=5
LimitNOFILE=65535
EnvironmentFile=/opt/vortexui/.env

[Install]
WantedBy=multi-user.target
```

### سرویس ایجنت نود

```ini
# /etc/systemd/system/vortex-node.service
[Unit]
Description=VortexUI Node Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vortex-node
ExecStart=/opt/vortex-node/vortex-node
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### دستورات رایج

```bash
sudo systemctl start vortexui
sudo systemctl stop vortexui
sudo systemctl restart vortexui
sudo systemctl status vortexui
sudo systemctl enable vortexui   # شروع هنگام بوت
```
