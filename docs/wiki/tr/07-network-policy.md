# 7. Ağ Politikası

!!! tip
    Yaygın kalıp: `geosite:ir` ve `geoip:ir` → direct, geri kalan → proxy.

---

## Outbound'lar

**Outbound'lar**, inbound sonrası trafik için çıkış yolunu tanımlar.

### Türler

| Tag | Rol |
|-----|------|
| `freedom` | Doğrudan — proxy yok |
| `blackhole` | Düşür |
| `dns` | Dahili çözücü |
| `proxy` | Başka bir upstream'e zincir |
| `warp` | Cloudflare WARP+ |

### CRUD

- **Outbounds → Add** — JSON editörü + share link içe aktarma
- Inbound/yönlendirme kuralına bağlantı

### Örnek: Proxy zinciri

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

## Yönlendirme

**Routing → Add Rule**

| Matcher | Örnek |
|---------|---------|
| Domain | `geosite:ir`, `domain:google.com` |
| IP | `geoip:ir`, `192.168.0.0/16` |
| Port | `80,443` |
| Protocol | `tcp,udp` |
| Inbound tag | `vless-in` |

### Yaygın İran deseni

```
geosite:ir  → outbound: direct (freedom)
geoip:ir    → outbound: direct
default     → outbound: proxy
```

---

## Dengeleyiciler

**Balancers → Add**

| Strategy | Davranış |
|----------|----------|
| `random` | Rastgele seçim |
| `roundRobin` | Round-robin |
| `leastPing` | En düşük ping |
| `leastLoad` | En düşük yük |

### Observatory

- Outbound'larda sağlık prob'ları
- Sağlıksız olanları havuzdan otomatik kaldırır

---

## Evasion Profilleri

**Evasion** — anti-DPI ön ayarları:

| Preset | İçerik |
|--------|---------|
| Iran (Fragment + Chrome) | TLS fragment + Chrome fingerprint |
| China (Mux + Random) | h2mux + randomized FP |
| Russia (Fragment + Firefox) | Kısa fragment |

Inbound'a bağlı — fragment, mux, fingerprint tek yerde.

---

## WARP+ Entegrasyonu

Cloudflare tünellemesi için WARP tipi outbound — bypass veya gizlilik katmanı için kullanışlı.

---

## Yapılandırma Şablonları

**Settings → Subscription Template**

- Clash/sing-box çıktısını özelleştir
- Varsayılan kurallar, proxy-groups, DNS
