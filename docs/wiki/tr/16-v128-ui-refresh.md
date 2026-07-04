# Veltrix UI (v1.2.8)

!!! info "Yalnızca arayüz sürümü"
    v1.2.8 bir **UI ve i18n sürümüdür**. Veritabanı migration'ı veya API değişikliği
    gerekmez — yalnızca panel arayüzünü yeniden derleyin veya dağıtın.

---

## Yenilikler

| Alan | Öne çıkanlar |
|------|--------------|
| **Tasarım sistemi** | Cam kartlar, istatistik kutuları, durum rozetleri, cyan/sky paleti, sayfa giriş animasyonları |
| **Uygulama kabuğu** | Katlanabilir kenar çubuğu, sabit üst çubuk, mini mod, mobil çekmece |
| **Komut paleti** | **Ctrl+K** / **⌘K** ile herhangi bir yönetim sayfasına hızlı geçiş |
| **Temel sayfalar** | Overview, Users ve Nodes — canlı API istatistik kartları |
| **Kullanıcı portalı** | Yeniden tasarlanmış giriş, pano, masaüstü kenar çubuğu, mobil alt menü |
| **i18n** | EN / FA / TR / AR / RU / ZH / JA / ES dillerinde **639 anahtar** |

---

## Gezinme

### Kenar çubuğu

Sayfalar bölümlere ayrılır (Overview, Users, Nodes, Network, Security, Reseller, …).
Alttaki **ok** ile **mini moda** (yalnızca simgeler) geçin — simgenin üzerine gelince etiket görünür.

Mobilde **hamburger** menüsünü kullanın.

### Komut paleti

**Ctrl+K** (Windows/Linux) veya **⌘K** (macOS), veya üst çubuktaki arama alanına tıklayın.
Sayfa adını yazıp Enter'a basın.

### Dil değiştirici

Üst çubuk veya giriş sayfasındaki **globe** simgesi. Sekiz dil aynı anahtar kümesini paylaşır.

### Kullanıcı portalı kısayolu

Kenar çubuğu veya üst çubukta **User Portal** → `/portal/login` (kullanım, planlar, destek).

---

## Tema

Üst çubuktaki **güneş/ay** düğmesi — tercih yerel olarak saklanır.

---

## Geliştiriciler için

Çeviri dosyaları: `web/src/i18n/locale/*.json`. Düzenlemeden sonra:

```bash
cd web
node scripts/apply-i18n-locales.mjs
node scripts/check-i18n.mjs
npm run build
```

Tam notlar: [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md).
