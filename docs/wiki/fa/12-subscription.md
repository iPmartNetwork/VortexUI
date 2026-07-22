# خط لوله سابسکریپشن

!!! abstract "پردازش هوشمند سابسکریپشن"
    هر درخواست سابسکریپشن از یک خط لوله ۹ مرحله‌ای عبور می‌کند که درخواست خام `/sub/{token}` را
    به یک کانفیگ بهینه، آگاه از ISP و مختص فرمت تبدیل می‌کند.

---

## مرور کلی

VortexUI v1.4.0 هر درخواست سابسکریپشن را از یک **خط لوله ۹ مرحله‌ای** عبور می‌دهد. هر مرحله هوشمندی به خروجی نهایی اضافه می‌کند.

---

## مراحل خط لوله

```
┌─────────────────────────────────────────────────────────┐
│  مرحله ۱: تشخیص خودکار ISP                             │
│  مرحله ۲: تشخیص فرمت (بر اساس UA)                      │
│  مرحله ۳: حل گروه پروتکل                               │
│  مرحله ۴: اعمال پروفایل‌های هوشمند                      │
│  مرحله ۵: محاسبه امتیاز کیفیت                          │
│  مرحله ۶: انتخاب اوتباند چندمسیره                      │
│  مرحله ۷: پیکربندی Probing پیشرفته                     │
│  مرحله ۸: تزریق پارامترهای Share Link                   │
│  مرحله ۹: تولید پروکسی جایگزین CDN                     │
└─────────────────────────────────────────────────────────┘
```

---

## مرحله ۱: تشخیص خودکار ISP

خط لوله ISP کاربر را از IP درخواست سابسکریپشن شناسایی می‌کند:

| روش | اولویت | منبع |
|-----|--------|------|
| MaxMind GeoIP2 ASN | اصلی | جستجوی دیتابیس ASN |
| محدوده‌های IP ثابت | جایگزین | بلوک‌های CIDR ایرانی هاردکد |
| X-Forwarded-For | حالت CDN | هنگام قرارگیری پشت Cloudflare/CDN |
| بازنویسی دستی | بالاترین | پارامتر `?isp=MCI` |

نتیجه تشخیص تعیین می‌کند کدام گروه پروتکل، پروفایل mux و مخزن SNI اعمال شود.

```bash
# بررسی ISP تشخیص‌داده‌شده برای یک سابسکریپشن
curl https://panel.example.com/sub/{token}?format=json&debug=true \
  -H "X-Debug-Key: $DEBUG_KEY"

# پاسخ شامل:
# { "detected_isp": "MCI", "asn": 197207, "severity": "aggressive" }
```

---

## مرحله ۲: تشخیص فرمت

User-Agent فرمت خروجی را تعیین می‌کند:

| کلاینت | الگوی UA | فرمت |
|--------|---------|------|
| sing-box | `sing-box` | sing-box JSON |
| Clash Meta | `clash-meta`, `mihomo` | Clash YAML |
| Clash Verge | `clash-verge` | Clash YAML |
| V2rayN | `v2rayn` | Base64 share links |
| V2rayNG | `v2rayng` | Base64 share links |
| Nekoray | `nekoray` | Base64 share links |
| Hiddify | `hiddify` | sing-box JSON |
| Streisand | `streisand` | sing-box JSON |
| ناشناخته | — | Base64 share links (پیش‌فرض) |

بازنویسی با پارامتر:

```
/sub/{token}?format=sing-box
/sub/{token}?format=clash
/sub/{token}?format=v2ray
```

---

## مرحله ۳: حل گروه پروتکل

بر اساس تشخیص ISP، خط لوله [گروه پروتکل](08-protocol-groups.md) مناسب را انتخاب می‌کند:

1. یافتن گروه‌های مطابق با پروفایل ISP تشخیص‌داده‌شده
2. اگر چند مطابقت وجود دارد، ترجیح گروه با بالاترین امتیاز کیفیت
3. اگر گروه ISP-خاص نیست، استفاده از گروه `DEFAULT`
4. حل لیست اینباندها با ترتیب اولویت

---

## مرحله ۴: اعمال پروفایل‌های هوشمند

پروفایل‌های هوشمند تنظیمات ضد سانسور بر اساس ISP حل‌شده تزریق می‌کنند:

| تنظیم | منبع | اعمال به |
|-------|------|----------|
| پروتکل Mux | پروفایل mux اپراتور | تمام اوتباندها |
| SNI | مخزن چرخشی روزانه | اوتباندهای TLS |
| هدرهای ترانسپورت | قوانین مبهم‌سازی | gRPC, WS, HTTPUpgrade |
| تنظیمات Fragment | شدت ISP | TLS handshake |
| Padding | پروفایل mux اپراتور | اوتباندهای Mux-فعال |

جزئیات پروفایل‌ها در [ضد سانسور](09-anti-censorship.md).

---

## مرحله ۵: امتیاز کیفیت

هر اینباند امتیاز کیفیت (۰–۱۰۰) بر اساس تلمتری بلادرنگ دریافت می‌کند:

### فرمول امتیازدهی

```
quality_score = (
    success_rate × 40 +
    latency_score × 25 +
    uptime_7d × 20 +
    switch_stability × 15
)
```

| مؤلفه | وزن | اندازه‌گیری |
|-------|-----|-------------|
| نرخ موفقیت | ۴۰٪ | اتصال‌های موفق / کل تلاش‌ها (۲۴ ساعت) |
| امتیاز تأخیر | ۲۵٪ | `100 - min(latency_ms / 20, 100)` |
| آپتایم (۷ روزه) | ۲۰٪ | درصد آپتایم نود در ۷ روز |
| ثبات سوئیچ | ۱۵٪ | معکوس رویدادهای سوئیچ (۲۴ ساعت) |

### آستانه‌های امتیاز

| امتیاز | رتبه | اقدام |
|--------|------|-------|
| ۸۰–۱۰۰ | عالی | اولویت‌دار در گروه |
| ۶۰–۷۹ | خوب | شامل به‌صورت عادی |
| ۴۰–۵۹ | متوسط | کاهش اولویت، افزایش مانیتورینگ |
| ۰–۳۹ | ضعیف | حذف از سابسکریپشن، ارسال هشدار |

```bash
# دریافت امتیازهای کیفیت
curl https://panel.example.com/api/subscription/quality-scores \
  -H "Authorization: Bearer $TOKEN"
```

---

## مرحله ۶: اوتباند چندمسیره

خط لوله **۴ اینباند برتر** بر اساس امتیاز کیفیت انتخاب و کانفیگ اوتباند چندمسیره تولید می‌کند:

```json
{
  "outbounds": [
    {
      "type": "urltest",
      "tag": "auto-select",
      "outbounds": ["proxy-1", "proxy-2", "proxy-3", "proxy-4"],
      "url": "https://cp.cloudflare.com/generate_204",
      "interval": "3m",
      "tolerance": 100
    },
    { "type": "vless", "tag": "proxy-1", "...": "..." },
    { "type": "trojan", "tag": "proxy-2", "...": "..." },
    { "type": "vmess", "tag": "proxy-3", "...": "..." },
    { "type": "hysteria2", "tag": "proxy-4", "...": "..." }
  ]
}
```

!!! note
    تعداد اوتباندها به ازای هر گروه قابل تنظیم است (پیش‌فرض: ۴، حداکثر: ۸).

---

## مرحله ۷: Probing پیشرفته

بررسی سلامت سمت کلاینت بر اساس ISP تنظیم شده:

| پارامتر | MCI | ایرانسل | TCI | پیش‌فرض |
|---------|-----|---------|-----|---------|
| URL‌های Probe | ۳ | ۲ | ۳ | ۲ |
| بازه | ۹۰ ثانیه | ۱۸۰ ثانیه | ۶۰ ثانیه | ۱۸۰ ثانیه |
| تلرانس | ۱۵۰ms | ۲۰۰ms | ۱۰۰ms | ۲۰۰ms |
| زمان‌محدودیت | ۵ ثانیه | ۵ ثانیه | ۳ ثانیه | ۵ ثانیه |

URL‌های Probe (چرخشی برای جلوگیری از اثرانگشت):

```json
{
  "probe_urls": [
    "https://cp.cloudflare.com/generate_204",
    "https://www.gstatic.com/generate_204",
    "https://detectportal.firefox.com/success.txt"
  ]
}
```

---

## مرحله ۸: پارامترهای Share Link

برای فرمت‌های Base64 / share link، پارامترهای اضافی تزریق می‌شوند:

| پارامتر | مقدار | هدف |
|---------|-------|-----|
| `port_range` | `20000-40000` | محدوده پرش پورت (Hysteria2/TUIC) |
| `hop_interval` | `30` | ثانیه بین پرش پورت‌ها |
| `fragment` | `1-3,100-200,tlshello` | تنظیمات fragment TLS |
| `mux` | `h2mux,8,padding` | پیکربندی پروتکل mux |

نمونه share link تولیدشده:

```
vless://uuid@server:443?type=grpc&security=reality&sni=dl.google.com
  &serviceName=google.pubsub.v1.Publisher
  &pbk=...&sid=...&fp=chrome
  &fragment=1-3,100-200,tlshello
  &mux=h2mux,8,padding
```

---

## مرحله ۹: پروکسی جایگزین CDN

مرحله نهایی یک پروکسی CDN-مسیر برای هر اتصال اصلی تولید می‌کند:

```json
{
  "outbounds": [
    { "tag": "primary-vless", "server": "direct-ip", "...": "..." },
    {
      "tag": "cdn-fallback-vless",
      "type": "vless",
      "server": "172.67.x.x",
      "server_port": 443,
      "tls": {
        "enabled": true,
        "server_name": "your-worker.domain.workers.dev",
        "alpn": ["h2", "http/1.1"]
      },
      "transport": {
        "type": "ws",
        "path": "/sub-ws",
        "headers": { "Host": "your-worker.domain.workers.dev" }
      }
    }
  ]
}
```

جایگزین CDN هنگام ناموفق بودن اتصال مستقیم فعال شده و مسیر ثانویه از طریق شبکه Cloudflare فراهم می‌کند.

---

## تاب‌آوری اتصال

سابسکریپشن شامل تنظیمات تاب‌آوری برای اتصال مقاوم است:

| تنظیم | مقدار | توضیحات |
|-------|-------|---------|
| `idle_timeout` | `15m` | بستن اتصالات بیکار بعد از ۱۵ دقیقه |
| `connect_timeout` | `5s` | زمان‌محدودیت برقراری اتصال |
| `tcp_fast_open` | `true` | کاهش تأخیر handshake |
| `tcp_multi_path` | `true` | MPTCP در صورت موجود بودن |
| `udp_timeout` | `5m` | زمان‌محدودیت سشن UDP |
| `domain_strategy` | `prefer_ipv4` | اول IPv4 برای ثبات |

---

## مانیتورینگ و دیباگ

### تحلیل سابسکریپشن

```bash
# دریافت آمار اجرای خط لوله
curl https://panel.example.com/api/subscription/stats \
  -H "Authorization: Bearer $TOKEN"
```

پاسخ:

```json
{
  "total_requests_24h": 12450,
  "by_format": { "sing-box": 6200, "clash": 3100, "v2ray": 3150 },
  "by_isp": { "MCI": 5400, "IRANCELL": 3800, "TCI": 2100, "other": 1150 },
  "avg_quality_score": 78,
  "cache_hit_rate": 0.92
}
```

### حالت دیباگ

`?debug=true` با کلید دیباگ معتبر اضافه کنید تا تصمیمات خط لوله را ببینید:

```bash
curl "https://panel.example.com/sub/{token}?debug=true" \
  -H "X-Debug-Key: $DEBUG_KEY"
```

---

## پیکربندی

تنظیمات خط لوله در **تنظیمات → سابسکریپشن** یا از طریق API مدیریت می‌شوند:

```bash
curl -X PATCH https://panel.example.com/api/settings/subscription \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "max_outbounds": 4,
    "probe_interval": "3m",
    "quality_threshold": 40,
    "cdn_fallback_enabled": true,
    "cache_ttl": "5m"
  }'
```

| تنظیم | توضیحات | پیش‌فرض |
|-------|---------|---------|
| `max_outbounds` | حداکثر اوتباند در هر بلوک urltest | ۴ |
| `probe_interval` | بازه بررسی سلامت urltest | `3m` |
| `quality_threshold` | حداقل امتیاز برای شامل شدن | ۴۰ |
| `cdn_fallback_enabled` | تولید پروکسی جایگزین CDN | `true` |
| `cache_ttl` | مدت کش سابسکریپشن | `5m` |

---

## صفحات مرتبط

- [گروه‌های پروتکل](08-protocol-groups.md)
- [ضد سانسور](09-anti-censorship.md)
- [مدیریت کاربران](05-user-management.md)
- [مرجع API](17-api-reference.md)
