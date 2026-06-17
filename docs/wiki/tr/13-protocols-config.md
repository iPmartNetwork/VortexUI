# 13. Protokoller ve Yapılandırma

> Tam İngilizce JSON örnekleri: [`docs/protocols.md`](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/protocols.md)

!!! tip "İpucu"
    Ağır sansür altında: **VLESS + REALITY + Vision** — anahtarları UI'dan oluşturun.

---

## VLESS

### VLESS + TCP + REALITY (sansür için önerilen)

| Alan | Değer |
|-------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `tcp` |
| Security | `reality` |
| Flow | `xtls-rprx-vision` |
| SNI | `www.microsoft.com` |

**UI:** Nodes → Inbounds → Add → Generate REALITY keys

### VLESS + WebSocket + TLS (CDN)

| Alan | Değer |
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

| Alan | Değer |
|-------|-------|
| Core | `singbox` |
| Protocol | `hysteria2` |
| Port | UDP (e.g. 443) |
| TLS | required |

---

## TUIC (sing-box only)

| Alan | Değer |
|-------|-------|
| Protocol | `tuic` |
| Congestion | `bbr` |

---

## WireGuard

sing-box üzerinde WireGuard inbound — kullanıcı başına veya paylaşımlı peer.

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

Veya UI'dan: inbound formundaki **Generate** düğmesi.

---

## Practical Tips

| Scenario | Recommendation |
|----------|----------------|
| Heavy censorship | VLESS+REALITY+Vision |
| CDN | VLESS/VMess+WS+TLS |
| UDP blocked | TCP-based protocols |
| High speed | REALITY or XTLS-Vision |
| Mobile in Iran | Fragment evasion profile |
