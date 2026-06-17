<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/04-dashboard.md) · [EN](../en/04-dashboard.md) · [AR](../ar/04-dashboard.md)

</div>

<div>

# 4. Gösterge Paneli (Overview)

[← İlk adımlar](./03-first-steps.md) · [Dizin](./README.md) · [Sonraki: Kullanıcılar →](./05-user-management.md)

> [!NOTE]
> Pano **SSE** ile yenilenir — polling gerekmez.

<div align="center">

| Light | Dark |
|:-----:|:----:|
| ![Overview panosu — canlı istatistikler ve trafik grafiği](../assets/panel/overview_light.png) | ![Overview panosu — canlı istatistikler ve trafik grafiği](../assets/panel/overview_dark.png) |

*Overview panosu — canlı istatistikler ve trafik grafiği*

</div>

---

## Genel Bakış

**Overview** sayfası merkezi operasyon görünümüdür: filo durumu, trafik, aktif kullanıcılar ve son olaylar — hepsi **canlı güncellemeler (SSE)** ile.

---

## İstatistik Kartları

| Kart | İçerik |
|------|---------|
| **Users** | Toplam, aktif, sınırlı, süresi dolmuş |
| **Traffic** | Toplam yükleme/indirme, zaman serisi |
| **Nodes** | Çevrimiçi/çevrimdışı sayısı |
| **Connections** | Aktif proxy bağlantıları |

---

## Trafik Grafiği

- **TimescaleDB** destekli zaman serisi
- Seçilebilir aralıklar (24h, 7d, 30d)
- Yükleme/indirme dökümü

---

## Canlı Güncellemeler (SSE)

Panel **Server-Sent Events** kullanır:

```
GET /api/events/stream?access_token=<JWT>
```

Bir olay gerçekleştiğinde (node kapalı, kullanıcı sınırlı vb.) UI sayfa yenilemeden güncellenir.

| Olay | UI etkisi |
|-------|-----------|
| `node.down` | Kırmızı node rozeti + toast |
| `user.limited` | Kullanıcı durumu güncellendi |
| `user.ip_limit` | Hesap paylaşım uyarısı |
| `user.expiry_warning` | 3 gün kala süre dolumu bildirimi |

> Caddy bu akışı şeffaf şekilde proxy'ler. Token query string'den gelir çünkü `EventSource` özel başlık gönderemez.

---

## Prometheus / Grafana

Metrikler Prometheus endpoint'inde mevcuttur (harici izleme için). Ayrıntılar: [Bölüm 14 — İşletim](./14-operations-maintenance.md).

</div>
