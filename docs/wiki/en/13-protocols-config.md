<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/13-protocols-config.md) · [AR](../ar/13-protocols-config.md) · [TR](../tr/13-protocols-config.md)

</div>

<div>

# 13. Protocols & Configuration

[← API](./12-api-reference.md) · [Index](./README.md) · [Next: Operations →](./14-operations-maintenance.md)

> Full English JSON examples: [`docs/protocols.md`](../../protocols.md)

> [!TIP]
> Under heavy censorship: **VLESS + REALITY + Vision** — generate keys from the UI.

---

## VLESS

### VLESS + TCP + REALITY (recommended for censorship)

| Field | Value |
|-------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `tcp` |
| Security | `reality` |
| Flow | `xtls-rprx-vision` |
| SNI | `www.microsoft.com` |

**UI:** Nodes → Inbounds → Add → Generate REALITY keys

### VLESS + WebSocket + TLS (CDN)

| Field | Value |
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

| Field | Value |
|-------|-------|
| Core | `singbox` |
| Protocol | `hysteria2` |
| Port | UDP (e.g. 443) |
| TLS | required |

---

## TUIC (sing-box only)

| Field | Value |
|-------|-------|
| Protocol | `tuic` |
| Congestion | `bbr` |

---

## WireGuard

WireGuard inbound on sing-box — peer per user or shared.

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

Or from the UI: **Generate** button in the inbound form.

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
