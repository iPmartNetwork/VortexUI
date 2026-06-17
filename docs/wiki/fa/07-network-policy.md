# ۷. سیاست شبکه

!!! tip "راهنما"
    الگوی رایج: `geosite:ir` و `geoip:ir` → direct، بقیه → proxy.

---

## Outbounds (خروجی‌ها)

**Outbounds** مسیر خروج ترافیک پس از inbound را تعریف می‌کنند.

### انواع

| Tag | نقش |
|-----|------|
| `freedom` | مستقیم — بدون proxy |
| `blackhole` | drop |
| `dns` | resolver داخلی |
| `proxy` | زنجیره به upstream دیگر |
| `warp` | Cloudflare WARP+ |

### CRUD

- **Outbounds → Add** — JSON editor + share link import
- اتصال به inbound/routing rule

### مثال: زنجیره proxy

```json
{
  "tag": "chain-de",
  "protocol": "vless",
  "settings": {
    "vnext": [{
      "address": "upstream.example.com",
      "port": 443,
      "users": [{"id": "uuid", "encryption": "none", "flow": "xtls-rprx-vision"}]
    }]
  }
}
```

---

## Routing (مسیریابی)

**Routing → Add Rule**

| Matcher | مثال |
|---------|------|
| Domain | `geosite:ir`, `domain:google.com` |
| IP | `geoip:ir`, `192.168.0.0/16` |
| Port | `80,443` |
| Protocol | `tcp,udp` |
| Inbound tag | `vless-in` |

### الگوی رایج ایران

```
geosite:ir  → outbound: direct (freedom)
geoip:ir    → outbound: direct
default     → outbound: proxy
```

---

## Balancers (متعادل‌کننده)

**Balancers → Add**

| Strategy | رفتار |
|----------|-------|
| `random` | تصادفی |
| `roundRobin` | نوبتی |
| `leastPing` | کمترین ping |
| `leastLoad` | کمترین بار |

### Observatory

- Health probe روی outboundها
- حذف خودکار unhealthy از pool

---

## Evasion Profiles

**Evasion** — preset ضد DPI:

| Preset | محتوا |
|--------|-------|
| Iran (Fragment + Chrome) | TLS fragment + fingerprint chrome |
| China (Mux + Random) | h2mux + randomized FP |
| Russia (Fragment + Firefox) | fragment کوتاه |

به inbound link می‌شود — fragment، mux، fingerprint یک‌جا.

---

## WARP+ Integration

Outbound نوع WARP برای عبور از Cloudflare — مفید برای bypass یا privacy layer.

---

## Config Templates

**Settings → Subscription Template**

- سفارشی‌سازی خروجی Clash/sing-box
- ruleهای پیش‌فرض، proxy-groups، dns
