# العمليات والصيانة

---

## HTTPS مع Caddy

تستخدم VortexUI خادم Caddy لـ HTTPS تلقائي:

- **Let's Encrypt تلقائي** — تُصدر الشهادات وتُجدد بدون إعداد
- **إعادة توجيه HTTP → HTTPS** — جميع الحركة تُجبر على HTTPS
- **توجيه SPA** — واجهة React تُقدَّم بشكل صحيح
- **وكيل عكسي** — طلبات API تُوجَّه لخلفية Go
- **نقطة نهاية DoH** — `/dns-query` تُقدَّم إذا كان DoH مفعّلاً

### بنية Caddyfile

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
    إذا كنت بحاجة لإعدادات TLS مخصصة (مثل شهادة عميل، مجموعات تشفير محددة)، عدّل `deploy/Caddyfile`.

---

## مقاييس Prometheus

**التفعيل:** عيّن `VORTEX_METRICS_ENABLED=true` و `VORTEX_METRICS_LISTEN=:9090`.

### المقاييس المتاحة

| المقياس | النوع | الوصف |
|---------|-------|-------|
| `vortex_users_total` | Gauge | إجمالي عدد المستخدمين حسب الحالة |
| `vortex_traffic_bytes` | Counter | بايتات الحركة (رفع/تنزيل) |
| `vortex_connections_active` | Gauge | الاتصالات النشطة الحالية |
| `vortex_nodes_status` | Gauge | حالة العقدة (1=متصلة، 0=غير متصلة) |
| `vortex_node_cpu_percent` | Gauge | استخدام المعالج للعقدة |
| `vortex_node_memory_percent` | Gauge | استخدام الذاكرة للعقدة |
| `vortex_orders_total` | Counter | الطلبات حسب الحالة |
| `vortex_api_requests_total` | Counter | طلبات API حسب نقطة النهاية/الحالة |
| `vortex_api_latency_seconds` | Histogram | زمن استجابة API |

### لوحة معلومات Grafana

استورد لوحة المعلومات الجاهزة من `deploy/grafana/vortexui-dashboard.json`:

1. افتح Grafana → Dashboards → Import
2. ارفع ملف JSON أو الصق محتواه
3. اختر مصدر بيانات Prometheus الخاص بك
4. تعرض لوحة المعلومات: اتجاهات الحركة، سلامة العقد، نشاط المستخدمين، أداء API

---

## التوسّع والتوفر العالي

### التوسّع الأفقي

للنشرات الكبيرة:

| المكوّن | استراتيجية التوسّع |
|---------|-------------------|
| Panel API | تشغيل عدة نسخ خلف موازن حمل |
| قاعدة البيانات | PostgreSQL مع نسخ قراءة |
| Redis | Redis Cluster أو Sentinel |
| العقد | إضافة بقدر الحاجة (تُدار تلقائياً) |
| الاتحاد | ربط عدة لوحات للتوزيع الإقليمي |

### وضع العنقود

تشغيل عدة نسخ من لوحة التحكم:

1. جميع النسخ تتشارك نفس PostgreSQL و Redis
2. استخدم موازن حمل (Caddy، nginx، HAProxy) أمامها
3. اتصالات SSE ملتصقة (يُوصى بتقارب الجلسة)
4. المهام الخلفية تستخدم أقفال موزّعة مبنية على Redis

### اعتبارات قاعدة البيانات

| المشكلة | الحل |
|---------|------|
| تجمّع الاتصالات | pgBouncer أمام PostgreSQL |
| توسّع القراءة | نسخ قراءة لاستعلامات التحليلات |
| إنتاجية الكتابة | جداول TimescaleDB الفائقة لبيانات الحركة |
| النسخ الاحتياطي | pg_dump + أرشفة WAL لاسترداد نقطة زمنية |

---

## صيانة قاعدة البيانات

### نصائح TimescaleDB

بيانات الحركة تُخزَّن في جداول TimescaleDB الفائقة لاستعلامات سلاسل زمنية فعّالة:

```sql
-- فحص أحجام الأجزاء
SELECT show_chunks('traffic_data');

-- تعيين سياسة الاحتفاظ (الاحتفاظ 90 يوماً)
SELECT add_retention_policy('traffic_data', INTERVAL '90 days');

-- ضغط الأجزاء القديمة
SELECT add_compression_policy('traffic_data', INTERVAL '7 days');
```

### صيانة روتينية

```bash
# Vacuum والتحليل
psql -U vortex -d vortex -c "VACUUM ANALYZE;"

# فحص حجم قاعدة البيانات
psql -U vortex -d vortex -c "SELECT pg_size_pretty(pg_database_size('vortex'));"

# قائمة أكبر الجداول
psql -U vortex -d vortex -c "
  SELECT relname, pg_size_pretty(pg_relation_size(oid))
  FROM pg_class
  ORDER BY pg_relation_size(oid) DESC
  LIMIT 10;
"
```

### الترحيلات

تعمل الترحيلات تلقائياً عند بدء لوحة التحكم. للتشغيل يدوياً:

```bash
vortexui migrate
```

فحص حالة الترحيل:

```bash
vortexui migrate status
```

---

## إدارة السجلات

### سجلات لوحة التحكم

```bash
# سجلات حيّة
journalctl -u vortexui -f

# آخر 100 سطر
journalctl -u vortexui -n 100

# تصفية حسب المستوى
journalctl -u vortexui | grep ERROR
```

### سجلات عميل العقدة

```bash
journalctl -u vortex-node -f
```

### مستويات السجل

تُعدّ عبر البيئة:

```
VORTEX_LOG_LEVEL=info   # debug, info, warn, error
VORTEX_LOG_FORMAT=json  # json أو text
```

### تدوير السجلات

journald في systemd يتولى التدوير افتراضياً. اضبط الحدود:

```ini
# /etc/systemd/journald.conf
SystemMaxUse=500M
MaxRetentionSec=30d
```

---

## تحسين الأداء

### حدود النظام

```bash
# زيادة حدّ واصفات الملفات
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf

# الزيادة لخدمة systemd
# في /etc/systemd/system/vortexui.service
[Service]
LimitNOFILE=65535
```

### تحسين الشبكة

```bash
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
```

تطبيق: `sysctl -p`

### تحسين Redis

```
# /etc/redis/redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
```

### تحسين PostgreSQL

الإعدادات الرئيسية لأعباء عمل VortexUI:

```
# postgresql.conf
shared_buffers = 512MB          # 25% من الذاكرة
effective_cache_size = 1536MB   # 75% من الذاكرة
work_mem = 16MB
maintenance_work_mem = 128MB
random_page_cost = 1.1          # SSD
```

---

## خدمات Systemd

### خدمة لوحة التحكم

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

### خدمة عميل العقدة

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

### الأوامر الشائعة

```bash
sudo systemctl start vortexui
sudo systemctl stop vortexui
sudo systemctl restart vortexui
sudo systemctl status vortexui
sudo systemctl enable vortexui   # البدء عند الإقلاع
```
