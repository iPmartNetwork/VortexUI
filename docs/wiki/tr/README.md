# VortexUI Dokümantasyonu

**Sürüm: 1.4.0** — Otomatik Protokol Geçişi ve Sansür Karşıtı Zeka

VortexUI resmi dokümantasyonuna hoş geldiniz. VortexUI, akıllı sansür karşıtı yeteneklerle Xray ve sing-box destekleyen yeni nesil, çekirdek bağımsız bir proxy yönetim panelidir.

## Hızlı Navigasyon

| Bölüm | Açıklama |
|-------|----------|
| [Giriş](01-introduction.md) | VortexUI nedir, tasarım ilkeleri |
| [Kurulum](02-installation.md) | Ön koşullar, tek satır kurulum, Docker |
| [İlk Adımlar](03-first-steps.md) | İlk kurulum, ilk düğüm, ilk kullanıcı |
| [Kontrol Paneli](04-dashboard.md) | Genel bakış, istatistikler |
| [Kullanıcı Yönetimi](05-user-management.md) | Kullanıcılar, kotalar, abonelikler |
| [Düğüm Yönetimi](06-node-management.md) | Filo yönetimi, sağlık |
| [Ağ Politikası](07-network-policy.md) | Yönlendirme, dengeleyiciler |
| [Güvenlik](08-security-administration.md) | TLS hileleri, prob koruması |
| [Planlar ve Ödemeler](09-plans-payments.md) | Ticaret, bayi, cüzdan |
| [Bildirimler](10-notifications.md) | Webhook, Telegram |
| [Ayarlar](11-settings-backup.md) | Yapılandırma, yedekleme |
| [API Referansı](12-api-reference.md) | REST API dokümantasyonu |
| [Protokoller](13-protocols-config.md) | Desteklenen protokoller |
| [Operasyonlar](14-operations-maintenance.md) | Bakım, yükseltmeler |
| [Sorun Giderme](15-troubleshooting-faq.md) | SSS, yaygın sorunlar |
| [Değişiklik Günlüğü](16-changelog.md) | Sürüm geçmişi |
| [Menü Rehberi](17-menu-usage-guide.md) | UI navigasyon rehberi |

## v1.4.0'daki Yenilikler

- **Otomatik Protokol Geçişi** — kendini onaran protokol yük devretme
- **Akıllı Yapılandırma Motoru** — ISP bazlı anti-DPI optimizasyonu
- **Dinamik SNI Rotasyonu** — ISP'ye özel havuzlardan günlük rotasyon
- **Multi-CDN Yönlendirme** — Cloudflare/ArvanCloud/Gcore temiz IP
- **Akıllı Mux** — ISP optimize çoğullama (h2mux/yamux)
- **Kalite Puanı** — proxy başına puanlama ve otomatik sıralama
- **DNS Sızıntı Önleme** — DoH + düz DNS engelleme
- **Acil Durum Yedek** — son çare çıkışı

## Bağlantılar

- [GitHub Deposu](https://github.com/iPmartNetwork/VortexUI)
- [Telegram Kanalı](https://t.me/vortex_ui)
