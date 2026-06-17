<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/07-network-policy.md) · [EN](../en/07-network-policy.md) · [TR](../tr/07-network-policy.md)

</div>

<div dir="rtl">

# ٧. سياسة الشبكة

[← العقد](./06-node-management.md) · [الفهرس](./README.md) · [التالي: الأمان →](./08-security-administration.md)

> [!TIP]
> النمط الشائع: `geosite:ir` و `geoip:ir` → direct، الباقي → proxy.

---

## Outbounds

**Outbounds** تحدد مسار الخروج للحركة بعد inbound.

### الأنواع

| Tag | الدور |
|-----|------|
| `freedom` | مباشر — بدون proxy |
| `blackhole` | إسقاط |
| `dns` | محلل داخلي |
| `proxy` | سلسلة إلى upstream آخر |
| `warp` | Cloudflare WARP+ |

### CRUD

- **Outbounds → Add** — محرر JSON + استيراد share link
- ربط بـ inbound/قاعدة توجيه

### مثال: سلسلة proxy

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

## Routing

**Routing → Add Rule**

| Matcher | مثال |
|---------|---------|
| Domain | `geosite:ir`, `domain:google.com` |
| IP | `geoip:ir`, `192.168.0.0/16` |
| Port | `80,443` |
| Protocol | `tcp,udp` |
| Inbound tag | `vless-in` |

### نمط إيران الشائع

```
geosite:ir  → outbound: direct (freedom)
geoip:ir    → outbound: direct
default     → outbound: proxy
```

---

## Balancers

**Balancers → Add**

| Strategy | السلوك |
|----------|----------|
| `random` | اختيار عشوائي |
| `roundRobin` | Round-robin |
| `leastPing` | أقل ping |
| `leastLoad` | أقل حمل |

### Observatory

- فحوصات صحة على outbounds
- إزالة غير السليم تلقائياً من المجموعة

---

## Evasion Profiles

**Evasion** — إعدادات مسبقة anti-DPI:

| Preset | المحتوى |
|--------|---------|
| Iran (Fragment + Chrome) | TLS fragment + Chrome fingerprint |
| China (Mux + Random) | h2mux + randomized FP |
| Russia (Fragment + Firefox) | Fragment قصير |

مرتبط بـ inbound — fragment، mux، fingerprint في مكان واحد.

---

## تكامل WARP+

Outbound من نوع WARP لنفق Cloudflare — مفيد للتجاوز أو طبقة الخصوصية.

---

## قوالب التكوين

**Settings → Subscription Template**

- تخصيص مخرجات Clash/sing-box
- قواعد افتراضية، proxy-groups، DNS

</div>
