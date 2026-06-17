<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/15-troubleshooting-faq.md) · [EN](../en/15-troubleshooting-faq.md) · [AR](../ar/15-troubleshooting-faq.md)

</div>

<div>

# 15. Sorun Giderme ve SSS

[← İşletim](./14-operations-maintenance.md) · [Dizin](./README.md)

> [!TIP]
> Önce **`vortexui logs`** ve **`/api/health`** — çoğu sorun JWT, DB veya firewall.

---

## Yaygın Sorunlar

### Panel başlamıyor

```bash
vortexui status
vortexui logs
curl http://127.0.0.1:8080/api/health
```

| Neden | Çözüm |
|-------|----------|
| Empty JWT secret | `deploy/.env` → `JWT_SECRET=$(openssl rand -hex 32)` |
| DB down | `docker compose ps` — restart `db` |
| Port in use | `ss -tlnp \| grep 8080` |

---

### HTTPS / Let's Encrypt hatası

| Neden | Çözüm |
|-------|----------|
| Wrong DNS | A record to server IP |
| Port 80 closed | firewall: `ufw allow 80,443` |
| LE rate limit | wait 1h or staging test |

---

### Node offline / kırmızı

| Neden | Çözüm |
|-------|----------|
| Agent down | `systemctl status vortex-node` |
| mTLS mismatch | regenerate certs, SAN includes IP |
| Firewall | port 50051 gRPC open |
| Core crash | Nodes → Logs |

---

### Kullanıcı bağlanamıyor

| Kontrol | |
|-------|---|
| Inbound active? | Nodes → Inbounds |
| User status `active`? | Users |
| Expired / limited? | User detail |
| Inbound port open? | `ufw` / cloud security group |
| REALITY keys match? | regenerate + new sub |

---

### Trafik kaydedilmiyor

| Neden | Çözüm |
|-------|----------|
| Core API port | `VORTEX_CORE_API_PORT=10085` |
| Stats disabled in core | panel config renders stats |
| Redis down | restart redis |

---

### Boş abonelik

| Neden | Çözüm |
|-------|----------|
| No inbound assigned | Edit user → select inbounds |
| Node down | fix node first |
| Wrong endpoint | set Custom Endpoint |

---

### SSE / canlı güncelleme çalışmıyor

| Neden | Çözüm |
|-------|----------|
| Caddy buffering | default OK — check proxy timeout |
| Token expired | re-login |
| Ad blocker | disable for panel domain |

---

## SSS

### VortexUI 3x-ui'den nasıl farklı?

**Kullanıcı merkezli** model, push delta trafik, tam outbound/routing/balancer, audit, API token, gelişmiş failover.

### SQLite destekliyor mu?

Hayır — **PostgreSQL + TimescaleDB** (üretim sınıfı, trafik zaman serileri).

### Kaç node destekleniyor?

Sınırsız — her node'un ayrı agent'ı veya bir local node'u vardır.

### sing-box mu xray mi?

Node başına — Hysteria2/TUIC yalnızca sing-box'ta; REALITY her ikisinde.

### Marzban'dan içe aktarma?

Evet — Users → Import.

### Hesap paylaşımı?

Cihaz limiti + online IP guard + isteğe bağlı autolimit.

### ZarinPal ile satış?

Plans + ZarinPal gateway — [Bölüm 9](./09-plans-payments.md).

### Güncellemeden önce yedek?

**Her zaman** — `vortexui update` güvenli ama yedekleme önerilir.

### Lisans?

GPL-3.0 — türevler açık kaynak olmalıdır.

---

## Hata Bildirimi

1. [GitHub Issues](https://github.com/iPmartNetwork/VortexUI/issues)
2. Sürüm: `vortexui settings` veya kenar çubuğu
3. Loglar: `vortexui logs` (sırlar olmadan)
4. Güvenlik açıkları için [SECURITY.md](../../../SECURITY.md)

---

## Topluluk

- ⭐ Star on GitHub
- [Contributing](../../../CONTRIBUTING.md)
- [Changelog](../../../CHANGELOG.md)

</div>
