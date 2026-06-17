<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](../README.md) · [EN](../en/13-protocols-config.md) · [AR](../ar/13-protocols-config.md) · [TR](../tr/13-protocols-config.md)

</div>

<div dir="rtl">

# ۱۳. پروتکل‌ها و پیکربندی

[← API](./12-api-reference.md) · [فهرست](./README.md) · [بعدی: عملیات →](./14-operations-maintenance.md)

> مثال‌های JSON کامل انگلیسی: [`docs/protocols.md`](../../protocols.md)

> [!TIP]
> برای سانسور شدید: **VLESS + REALITY + Vision** — کلیدها را از UI Generate کنید.

---

## VLESS

### VLESS + TCP + REALITY (پیشنهادی برای censorship)

| فیلد | مقدار |
|------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `tcp` |
| Security | `reality` |
| Flow | `xtls-rprx-vision` |
| SNI | `www.microsoft.com` |

**UI:** Nodes → Inbounds → Add → Generate REALITY keys

### VLESS + WebSocket + TLS (CDN)

| فیلد | مقدار |
|------|-------|
| Network | `ws` |
| Security | `tls` |
| Path | `/vless-ws` |
| SNI | `your-domain.com` |

Cloudflare: proxy orange cloud + WebSocket enabled.

### VLESS + gRPC + TLS

| Network | `grpc` |
| Path | service name (e.g. `vless-grpc`) |

### VLESS + HTTPUpgrade + TLS

| Network | `httpupgrade` |
| Path | `/vless-hu` |

---

## VMess

| Setup | Network | Security |
|-------|---------|----------|
| CDN | `ws` | `tls` |
| Simple | `tcp` | `none` |
| Multiplex | `grpc` | `tls` |

---

## Trojan

| Setup | Network |
|-------|---------|
| Direct TLS | `tcp` |
| CDN | `ws` |
| gRPC | `grpc` |

---

## Shadowsocks

| نوع | Cipher |
|-----|--------|
| SS2022 | `2022-blake3-aes-128-gcm` |
| Classic | `aes-256-gcm`, `chacha20-ietf-poly1305` |

---

## Hysteria2 (sing-box only)

| فیلد | مقدار |
|------|-------|
| Core | `singbox` |
| Protocol | `hysteria2` |
| Port | UDP (e.g. 443) |
| TLS | required |

---

## TUIC (sing-box only)

| فیلد | مقدار |
|------|-------|
| Protocol | `tuic` |
| Congestion | `bbr` |

---

## WireGuard

Inbound WireGuard روی sing-box — peer per user یا shared.

---

## Outbound Examples

### Direct

```json
{ "tag": "direct", "protocol": "freedom" }
```

### Block

```json
{ "tag": "block", "protocol": "blackhole" }
```

---

## Routing Examples

### ایران مستقیم، بقیه proxy

```json
{
  "rules": [
    { "type": "field", "domain": ["geosite:ir"], "outboundTag": "direct" },
    { "type": "field", "ip": ["geoip:ir"], "outboundTag": "direct" },
    { "type": "field", "network": "tcp,udp", "outboundTag": "proxy" }
  ]
}
```

---

## REALITY Key Generation

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  https://panel.example.com/api/reality/keypair
```

یا از UI: دکمه **Generate** در فرم inbound.

---

## نکات عملی

| سناریو | پیشنهاد |
|--------|---------|
| فیلترینگ شدید | VLESS+REALITY+Vision |
| CDN | VLESS/VMess+WS+TLS |
| UDP blocked | TCP-based protocols |
| سرعت بالا | REALITY یا XTLS-Vision |
| موبایل ایران | Fragment evasion profile |

</div>
