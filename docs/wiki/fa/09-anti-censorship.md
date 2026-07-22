# ضد سانسور

!!! abstract "موتور هوشمند ضد سانسور"
    VortexUI v1.4.0 یک لایه تطبیقی ارائه می‌دهد که تنظیمات پروکسی را بر اساس شدت فیلترینگ ISP،
    ساعت روز و داده‌های تاریخی مسدودسازی تنظیم می‌کند. به‌جای کانفیگ‌های ثابت، کاربران اتصالات
    بهینه‌ای دریافت می‌کنند که با الگوهای سانسور تکامل می‌یابند.

!!! warning
    ویژگی‌های ضد سانسور برای زیرساخت فیلترینگ ایران تنظیم شده‌اند. مناطق دیگر ممکن است به پروفایل‌های سفارشی نیاز داشته باشند.

---

## طبقه‌بندی شدت ISP

موتور، ISP‌ها را بر اساس تهاجمی بودن فیلترینگ طبقه‌بندی کرده و استراتژی‌های پروتکل متناسب اعمال می‌کند:

| ISP | شدت | رفتار | پشته پیشنهادی |
|-----|------|-------|---------------|
| MCI (همراه اول) | تهاجمی | بازرسی عمیق بسته، probing فعال، تحلیل اثرانگشت TLS | VLESS+Reality, gRPC, h2mux |
| ایرانسل (MTN) | متوسط | DPI ساده، مسدودسازی گاه‌به‌گاه پورت، probing کمتر | Trojan+WS+TLS, yamux |
| TCI (مخابرات) | همیشه تهاجمی | فیلتر سنگین، مسدودسازی SNI، ممنوعیت مکرر پروتکل | VLESS+Reality+gRPC, مبهم‌سازی کامل |
| شاتل | متوسط | بازرسی HTTPS استاندارد، به‌ندرت پروتکل جدید مسدود می‌کند | VMess+WS+TLS, mux استاندارد |
| رایتل | کم-متوسط | موبایل‌محور، ظرفیت DPI محدود | اکثر پروتکل‌ها کار می‌کنند |

---

## تطبیق بر اساس ساعت روز

شدت فیلترینگ در ساعات اوج افزایش می‌یابد. موتور شدت را به‌صورت پویا تنظیم می‌کند:

| بازه زمانی (IRST) | تعدیل‌کننده | اقدام |
|--------------------|-------------|-------|
| ۰۰:۰۰–۰۸:۰۰ | -۱ شدت | مبهم‌سازی سبک‌تر قابل قبول |
| ۰۸:۰۰–۱۸:۰۰ | عادی | پروفایل استاندارد اعمال می‌شود |
| ۱۸:۰۰–۲۳:۰۰ (اوج) | +۱ شدت | حداکثر مبهم‌سازی فعال |
| ۲۳:۰۰–۰۰:۰۰ | عادی | پروفایل استاندارد اعمال می‌شود |

!!! note
    در رویدادهای ملی یا قطع اینترنت، شدت بدون توجه به زمان به حداکثر اجبار می‌شود.

---

## چرخش پویای SNI

SNI‌های ثابت ظرف چند روز مسدود می‌شوند. VortexUI روزانه SNI هر پروکسی را از مخزن‌های ISP-خاص چرخش می‌دهد:

```json
{
  "sni_pools": {
    "MCI": ["dl.google.com", "update.microsoft.com", "cdn.cloudflare.com"],
    "IRANCELL": ["fonts.googleapis.com", "ajax.googleapis.com"],
    "TCI": ["www.google.com", "accounts.google.com", "play.google.com"]
  },
  "rotation_interval": "24h",
  "validation": "tls_handshake_check"
}
```

پنل هر SNI کاندید را قبل از استقرار از طریق TLS handshake اعتبارسنجی می‌کند. SNI‌های ناموفق خودکار از مخزن حذف می‌شوند.

### مدیریت مخزن‌های SNI

```bash
# افزودن SNI به مخزن
curl -X POST https://panel.example.com/api/anti-censorship/sni-pools \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "isp": "MCI",
    "snis": ["new.domain.com", "another.domain.com"],
    "validate": true
  }'

# دریافت SNI‌های فعال فعلی
curl https://panel.example.com/api/anti-censorship/active-snis \
  -H "Authorization: Bearer $TOKEN"
```

---

## مبهم‌سازی ترانسپورت

هر ترانسپورت به‌شکل ترافیک مشروع تغییر ظاهر می‌دهد:

| ترانسپورت | ظاهر | هدر/مسیر |
|-----------|------|----------|
| gRPC | Google Pub/Sub API | `google.pubsub.v1.Publisher/Publish` |
| WebSocket | REST API polling | `/api/v2/ws`, `/api/v2/events` |
| HTTPUpgrade | بررسی بروزرسانی نرم‌افزار | `/update/check`, `/api/version` |
| xHTTP | دریافت asset از CDN | `/assets/chunk-{random}.js` |

تنظیم به ازای هر اینباند:

```bash
curl -X PATCH https://panel.example.com/api/inbounds/12/transport \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "grpc",
    "service_name": "google.pubsub.v1.Publisher",
    "authority": "pubsub.googleapis.com"
  }'
```

---

## پیکربندی هوشمند Mux

استراتژی مالتی‌پلکسینگ بر اساس ISP متفاوت است تا از الگوهای تشخیصی جلوگیری شود:

| ISP | پروتکل Mux | Padding | حداکثر استریم | XUDP |
|-----|------------|---------|---------------|------|
| MCI | h2mux | فعال (۱۰۰-۳۰۰ بایت) | ۸ | فعال |
| ایرانسل | yamux | فعال (۵۰-۱۵۰ بایت) | ۱۲ | فعال |
| TCI | h2mux | فعال (۲۰۰-۵۰۰ بایت) | ۴ | فعال |
| شاتل | smux | غیرفعال | ۱۶ | اختیاری |
| پیش‌فرض | yamux | غیرفعال | ۸ | فعال |

نمونه کانفیگ sing-box تولیدشده:

```json
{
  "multiplex": {
    "enabled": true,
    "protocol": "h2mux",
    "max_streams": 8,
    "padding": true,
    "brutal": {
      "enabled": false
    }
  }
}
```

!!! note
    XUDP برای UDP-over-TCP (بازی، VoIP) الزامی است. برای ISP‌های تهاجمی همیشه فعال است.

---

## مسیریابی CDN

برای ISP‌هایی که اتصال مستقیم را مسدود می‌کنند، مسیریابی CDN یک مسیر جایگزین فراهم می‌کند:

### جایگزین Clean-IP

```json
{
  "cdn_fallback": {
    "enabled": true,
    "provider": "cloudflare",
    "clean_ips": ["172.67.x.x", "104.21.x.x"],
    "alpn": ["h2", "http/1.1"],
    "sni": "your-worker.domain.workers.dev"
  }
}
```

**اسکنر Clean-IP** پنل به‌صورت دوره‌ای بهترین IP‌های لبه CDN را کشف می‌کند:

```bash
curl -X POST https://panel.example.com/api/scanner/clean-ip \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "provider": "cloudflare",
    "count": 20,
    "timeout_ms": 3000,
    "target_latency_ms": 200
  }'
```

### پیکربندی ALPN

ALPN صحیح از خرابی اتصال CDN جلوگیری می‌کند:

| سناریو | ALPN | توضیحات |
|---------|------|---------|
| WS روی CDN | `http/1.1` | CDN برای ارتقای WebSocket به HTTP/1.1 نیاز دارد |
| gRPC روی CDN | `h2` | gRPC الزاماً HTTP/2 است |
| HTTPUpgrade | `http/1.1` | از HTTP/1.1 101 Switching Protocols استفاده می‌کند |
| مستقیم (بدون CDN) | `h2, http/1.1` | اجازه ترجیح سرور |

---

## جلوگیری از نشت DNS

جلوگیری از عبور کوئری‌های DNS از تانل:

| روش | پیکربندی | هدف |
|-----|----------|-----|
| DoH (DNS-over-HTTPS) | `https://1.1.1.1/dns-query` | DNS رمزنگاری‌شده برای تمام کوئری‌ها |
| مستقیم IR | مسیریابی مستقیم دامنه‌های `.ir` | سایت‌های ایرانی نیاز به پروکسی ندارند |
| مسدودسازی پورت ۵۳ | قانون فایروال روی کلاینت | جلوگیری از نشت DNS ساده |
| Fake DNS | حالت `fakeip` در sing-box | حذف رفت‌وبرگشت DNS |

کانفیگ تولیدشده کلاینت شامل:

```json
{
  "dns": {
    "servers": [
      { "tag": "doh", "address": "https://1.1.1.1/dns-query", "detour": "proxy" },
      { "tag": "direct-dns", "address": "https://dns.403.online/dns-query", "detour": "direct" }
    ],
    "rules": [
      { "domain_suffix": [".ir"], "server": "direct-dns" },
      { "clash_mode": "direct", "server": "direct-dns" }
    ]
  }
}
```

---

## پرش پورت (Port Hopping)

چرخش پورت‌های شنود برای فرار از مسدودسازی مبتنی بر پورت:

```json
{
  "port_hopping": {
    "enabled": true,
    "port_range": "20000-40000",
    "hop_interval": "30s",
    "protocol": "hysteria2"
  }
}
```

| پارامتر | توضیحات | پیش‌فرض |
|---------|---------|---------|
| `port_range` | محدوده پورت UDP برای پرش | `20000-40000` |
| `hop_interval` | زمان بین تعویض پورت‌ها | `30s` |
| `protocol` | باید UDP-based باشد (Hysteria2, TUIC) | — |

!!! warning
    پرش پورت نیاز دارد محدوده کامل پورت در فایروال باز باشد (`ufw allow 20000:40000/udp`).

---

## اوتباند اضطراری

اوتباند `🆘 Emergency` زمانی فعال می‌شود که تمام اتصالات اصلی ناموفق باشند:

1. کلاینت تشخیص می‌دهد تمام اوتباندهای `urltest` غیرقابل دسترس هستند
2. به اوتباند اضطراری (WARP پیش‌پیکربندی‌شده یا سرور پشتیبان) سوئیچ می‌کند
3. هشدار از طریق اندپوینت fallback به پنل ارسال می‌شود
4. پنل رویداد را ثبت و ادمین را مطلع می‌کند

پیکربندی اوتباند اضطراری:

```bash
curl -X PUT https://panel.example.com/api/anti-censorship/emergency \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "wireguard",
    "server": "engage.cloudflareclient.com",
    "port": 2408,
    "private_key": "...",
    "peer_public_key": "...",
    "reserved": [0, 0, 0]
  }'
```

---

## چرخش گواهی

چرخش خودکار گواهی TLS برای کاهش اثرانگشت:

| تنظیم | مقدار | هدف |
|-------|-------|-----|
| بازه چرخش | ۱۵ روز | تعویض قبل از شکل‌گیری الگو |
| ارائه‌دهندگان | Let's Encrypt ↔ ZeroSSL | تناوب برای جلوگیری از اثرانگشت CA |
| روش | DNS-01 (Cloudflare) | بدون نیاز به پورت ۸۰ |
| همپوشانی | ۲ روز | گواهی قدیمی در انتقال معتبر است |

پنل چرخش را خودکار مدیریت می‌کند. مانیتور از طریق:

```bash
curl https://panel.example.com/api/anti-censorship/certs \
  -H "Authorization: Bearer $TOKEN"
```

---

## خلاصه پیکربندی

تمام تنظیمات ضد سانسور در **تنظیمات → ضد سانسور** در UI پنل یا از طریق خانواده اندپوینت `/api/anti-censorship/*` قابل دسترسی هستند.

| ویژگی | مکان UI | اندپوینت API |
|-------|---------|-------------|
| مخزن‌های SNI | ضد سانسور → SNI | `/api/anti-censorship/sni-pools` |
| پروفایل‌های Mux | ضد سانسور → Mux | `/api/anti-censorship/mux-profiles` |
| جایگزین CDN | ضد سانسور → CDN | `/api/anti-censorship/cdn` |
| پرش پورت | اینباندها → پیشرفته | `/api/inbounds/{id}/port-hopping` |
| اضطراری | ضد سانسور → اضطراری | `/api/anti-censorship/emergency` |
| چرخش گواهی | تنظیمات → TLS | `/api/anti-censorship/certs` |

---

## صفحات مرتبط

- [گروه‌های پروتکل](08-protocol-groups.md)
- [خط لوله سابسکریپشن](12-subscription.md)
- [سیاست شبکه](07-network-policy.md)
- [مرجع API](17-api-reference.md)
