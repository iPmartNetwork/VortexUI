<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/14-operations-maintenance.md) · [EN](../en/14-operations-maintenance.md) · [AR](../ar/14-operations-maintenance.md)

</div>

<div>

# 14. İşletim ve Bakım

[← Protokoller](./13-protocols-config.md) · [Dizin](./README.md) · [Sonraki: Sorun giderme →](./15-troubleshooting-faq.md)

> [!TIP]
> Kurulumdan sonra `vortexui status` ve `curl .../api/health` ile kontrol edin.

---

## `vortexui` Konsolu

```bash
vortexui              # interactive menu
vortexui start        # start stack
vortexui stop         # stop
vortexui restart      # restart
vortexui status       # status
vortexui logs         # tail logs
vortexui update       # git pull + rebuild
vortexui admin        # create admin
vortexui settings     # URL and settings
vortexui uninstall    # remove (with confirmation)
```

Kurulum yolu: `VORTEXUI_DIR` (varsayılan `/opt/vortexui`)

---

## HTTPS / SSL

### Docker (Caddy)

`deploy/.env`:

```env
SITE_ADDRESS=panel.example.com
ACME_EMAIL=admin@example.com
```

- 80 + 443 portları açık
- DNS A kaydı sunucuya
- Sertifika `caddy-data` volume'ünde

Alan adı değiştirme:

```bash
vortexui   # → option 9) Domain / SSL
```

### HTTP only

```env
SITE_ADDRESS=:8080
```

---

## Cluster Mode (HA)

Paylaşılan DB ile birden fazla panel instance — yüksek erişilebilirlik için. Ayrıntılar env ve deploy docs'ta.

---

## Prometheus / Grafana

Prometheus scraping için panel ve node metrikleri. Sürümlerde örnek Grafana dashboard.

---

## Auto-Update

- Panel ikili dosyası GitHub releases'ten
- Çekirdek ikilileri (xray/sing-box) upstream'den
- `vortexui update` veya Settings → Updates

---

## Migration

```bash
# with goose
export VORTEX_DATABASE_URL=postgres://...
make migrate-up
```

Docker: migration panel başlangıcında çalışır.

---

## Makefile (Development)

| Target | Action |
|--------|--------|
| `make build` | panel + node binaries |
| `make test` | tests with race detector |
| `make certs` | mTLS dev certs |
| `make stack-up` | full docker stack |
| `make proto` | regenerate gRPC |
| `make sqlc` | regenerate DB code |

---

## systemd (Native)

| Service | Role |
|---------|------|
| `vortexui-panel` | panel API |
| `vortex-node` | node agent (optional) |
| `caddy` | web + HTTPS |

```bash
systemctl status vortexui-panel caddy
journalctl -u vortexui-panel -f
```

---

## Recommended Backup Strategy

| Layer | Method | Frequency |
|-------|--------|-----------|
| DB | `GET /api/backup` | daily |
| Auto | Telegram/S3 | daily |
| Config | `deploy/.env` + certs | after changes |
| Off-site | copy to separate storage | weekly |

---

## Monitoring Checklist

- [ ] Uptime monitor'da `/api/health`
- [ ] `node.down` için uyarı
- [ ] Disk >85% uyarısı
- [ ] DB connection pool
- [ ] Sertifika süresi (Caddy auto — port 80'i doğrula)

</div>
