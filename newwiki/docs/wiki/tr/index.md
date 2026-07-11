# 🌀 VortexUI Dokümantasyonu

<div align="center">

**Yeni Nesil Proxy Yönetim Paneli**

*Kullanıcı Odaklı · Çekirdek Bağımsız · Kurumsal Hazır*

[![Sürüm](https://img.shields.io/badge/sürüm-1.3.1-7c3aed?style=for-the-badge)](https://github.com/iPmartNetwork/VortexUI/releases)
[![Lisans](https://img.shields.io/badge/lisans-MIT-green?style=for-the-badge)](https://github.com/iPmartNetwork/VortexUI/blob/master/LICENSE)
[![Docker](https://img.shields.io/badge/docker-hazır-blue?style=for-the-badge)](https://hub.docker.com/r/ipmartnetwork/vortexui)

</div>

---

## 🚀 Hızlı Kurulum

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

Tek komut. Etkileşimli kurulum. HTTPS dahil.

---

## 📖 Dokümantasyon Haritası

| Bölüm | Açıklama |
|-------|----------|
| [Giriş](01-introduction.md) | Mimari, özellik genel bakışı, karşılaştırma |
| [Kurulum](02-installation.md) | Tek satır kurulum, Docker, yerel derleme |
| [İlk Adımlar](03-first-steps.md) | Giriş, düğüm ekleme, inbound oluşturma, kullanıcı ekleme |
| [Kontrol Paneli](04-dashboard.md) | Widget'lar, analizler, monitör, komut paleti |
| [Kullanıcılar](05-user-management.md) | Yönetim, kotalar, abonelikler, portal, mağaza |
| [Düğümler](06-node-management.md) | Kayıt, sağlık, otomatik geçiş, izleme |
| [Ağ](07-network-policy.md) | Çıkışlar, yönlendirme paketleri, CDN zincirleri, yük dengeleyici |
| [Güvenlik](08-security-administration.md) | RBAC, TLS hileleri, prob koruması, IP limiti |
| [Planlar ve Ödemeler](09-plans-payments.md) | Bayi planları, ödeme yapılandırması, cüzdan |
| [Bildirimler](10-notifications.md) | Webhook'lar, Telegram, kota uyarıları, SSE |
| [Ayarlar](11-settings-backup.md) | Markalama, beyaz etiket, yedekleme, güncellemeler |
| [API Referansı](12-api-reference.md) | Kimlik doğrulama, uç noktalar, OpenAPI spec |
| [Protokoller](13-protocols-config.md) | 14 protokol, taşımalar, güvenlik katmanları |
| [Operasyonlar](14-operations-maintenance.md) | HTTPS, Prometheus, ölçekleme, performans |
| [Sorun Giderme](15-troubleshooting-faq.md) | Yaygın sorunlar, hata ayıklama, SSS |

---

## ✨ Temel Özellikler

### 🔧 Motor ve Altyapı
- **Çift çekirdek desteği** — Xray-core ve sing-box, düğüm başına seçim
- **Delta trafik** — Yeniden başlatmaya dayanıklı, veri kaybı yok
- **mTLS düğüm filosu** — Şifreli bağlantılar, otomatik yük devretme
- **Otomatik geçiş** — Sağlıksız düğümlerden kullanıcıları taşı
- **Federasyon** — Birden fazla panel arasında kullanıcı/düğüm senkronizasyonu

### 🛡 Güvenlik ve Sansür Karşıtı
- **Reality Tarayıcı** — Gecikme puanlamasıyla optimal SNI keşfi
- **TLS Hileleri Yöneticisi** — ISP profilleri (parçalama, mux, dolgu)
- **Prob Koruması** — GFW problarını algıla ve engelle
- **Sahte Web Sitesi** — Denetleyicilere sahte site göster
- **DNS-over-HTTPS** — Reklam engelleme ile yerleşik DoH

### 👥 Kullanıcı Yönetimi ve Ticaret
- **Self-servis portal** — Token ile giriş, kullanım görüntüleme, biletler
- **Self-servis mağaza** — Çoklu ödeme yöntemleriyle bayi planları
- **Akıllı Kota** — Kademeli hız azaltma
- **Aile grupları** — Paylaşılan veri havuzları
- **Referans sistemi** — Ödüllü davet kodları

### 💼 Bayi Platformu
- **Cüzdan** — Yükleme kuyruğu ile kredi sistemi
- **Alt bayiler** — Devralınan kapsamla alt bayiler oluştur
- **Beyaz etiket** — Özel markalama (logo, renkler, başlık)
- **Webhook'lar** — Otomasyon için giden olaylar
- **Politika limitleri** — Maks veri, maks süre, toplu kısıtlamalar

### 🎨 Ön Uç ve UX
- **Veltrix Cam Arayüzü** — Modern Glass tasarım sistemi
- **Komut paleti** — Her yerde Ctrl+K araması
- **Kontrol paneli widget'ları** — Sürükle, bırak, yeniden boyutlandır
- **8 dil** — Farsça ve Arapça için tam RTL desteği
- **Karanlık ve Aydınlık tema** — Yumuşak animasyonlu geçiş

---

## 🔗 Hızlı Bağlantılar

| Kaynak | Bağlantı |
|--------|----------|
| GitHub Deposu | [github.com/iPmartNetwork/VortexUI](https://github.com/iPmartNetwork/VortexUI) |
| Telegram Kanalı | [@vortex_ui](https://t.me/vortex_ui) |
| OpenAPI Spec | [openapi.yaml](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) |
| Değişiklik Günlüğü | [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md) |
| Hata Bildirimi | [GitHub Issues](https://github.com/iPmartNetwork/VortexUI/issues) |

---

## 🌍 Diller

Bu dokümantasyon şu dillerde mevcuttur:

- 🇹🇷 **Türkçe** (mevcut)
- 🇬🇧 [English](../en/index.md)
- 🇮🇷 [فارسی](../fa/index.md)
- 🇸🇦 [العربية](../ar/index.md)
