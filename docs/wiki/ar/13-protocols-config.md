<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](../README.md) · [FA](../fa/13-protocols-config.md) · [EN](../en/13-protocols-config.md) · [TR](../tr/13-protocols-config.md)

</div>

<div dir="rtl">

# ١٣. البروتوكولات والتكوين

[← API](./12-api-reference.md) · [الفهرس](./README.md) · [التالي: العمليات →](./14-operations-maintenance.md)

> أمثلة JSON كاملة بالإنجليزية: [`docs/protocols.md`](../../protocols.md)

> [!TIP]
> تحت رقابة شديدة: **VLESS + REALITY + Vision** — أنشئ المفاتيح من الواجهة.

---

## VLESS

### VLESS + TCP + REALITY (موصى به للرقابة)

| الحقل | القيمة |
|-------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `tcp` |
| Security | `reality` |
| Flow | `xtls-rprx-vision` |
| SNI | `www.microsoft.com` |

**UI:** Nodes → Inbounds → Add → Generate REALITY keys

### VLESS + WebSocket + TLS (CDN)

| الحقل | القيمة |
|-------|-------|
| Network | `ws` |
| Security | `tls` |
| Path | `/vless-ws` |
| SNI | `your-domain.com` |

Cloudflare: orange cloud proxy + WebSocket enabled.

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

| Type | Cipher |
|------|--------|
| SS2022 | `2022-blake3-aes-128-gcm` |
| Classic | `aes-256-gcm`, `chacha20-ietf-poly1305` |

---

## Hysteria2 (sing-box only)

| الحقل | القيمة |
|-------|-------|
| Core | `singbox` |
| Protocol | `hysteria2` |
| Port | UDP (e.g. 443) |
| TLS | required |

---

## TUIC (sing-box only)

| الحقل | القيمة |
|-------|-------|
| Protocol | `tuic` |
| Congestion | `bbr` |

---

## WireGuard

WireGuard inbound على sing-box — peer لكل مستخدم أو مشترك.

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

### Iran direct, rest via proxy

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

أو من الواجهة: زر **Generate** في نموذج inbound.

---

## Practical Tips

| Scenario | Recommendation |
|----------|----------------|
| Heavy censorship | VLESS+REALITY+Vision |
| CDN | VLESS/VMess+WS+TLS |
| UDP blocked | TCP-based protocols |
| High speed | REALITY or XTLS-Vision |
| Mobile in Iran | Fragment evasion profile |

</div>
