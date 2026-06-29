# Ayarlar ve Yedekleme

---

## Panel Ayarları

**Ayarlar** sayfası genel panel yapılandırmasını içerir:

| Ayar | Açıklama |
|------|----------|
| Panel başlığı | Tarayıcı sekmesinde ve başlıkta görüntülenir |
| Panel URL'si | Genel URL (abonelik bağlantıları için kullanılır) |
| Dil | Varsayılan arayüz dili |
| Saat dilimi | Görüntüleme için panel saat dilimi |
| JWT TTL | Token sona erme süresi |
| Abonelik güncelleme aralığı | İstemcilerin ne sıklıkla yenilemesi gerektiği |

---

## Özel Markalama

**Ayarlar → Markalama**

Panelin görünümünü özelleştirin:

| Alan | Açıklama |
|------|----------|
| Başlık | Panel başlığı (başlık + tarayıcı sekmesi) |
| Logo | Özel logo yükle (SVG/PNG) |
| Favicon | Özel tarayıcı favicon'u |
| Vurgu rengi | Birincil tema rengi (`#hex`) |
| Alt bilgi metni | Özel alt bilgi satırı |
| Giriş arka planı | Özel giriş sayfası arka planı |

---

## Bayi Beyaz Etiketi

Her bayi kendi portalını bağımsız olarak markalayabilir:

**Bayi Hesabı → Markalama**

| Alan | Açıklama |
|------|----------|
| Panel başlığı | Bayinin panel/portal başlığı |
| Logo URL'si | Bayinin marka görseli |
| Vurgu rengi | Kullanıcıları için `#hex` tema vurgusu |
| Portal slug'ı | İsteğe bağlı özel URL slug'ı |
| Alt bilgi metni | Portal alt bilgi satırı |

Portala erişen kullanıcılar ana panelin değil, bayilerinin markalamasını görür.

---

## Yedekleme ve Geri Yükleme

### Manuel Yedekleme

**Ayarlar → Yedekleme → Yedekleme Oluştur**

Şunları içeren tam bir yedekleme oluşturur:

- Veritabanı dökümü (kullanıcılar, düğümler, gelen bağlantılar, planlar, siparişler, ayarlar)
- Yapılandırma dosyaları
- TLS sertifikaları
- Yüklenen varlıklar (logolar, ödeme kanıtları)

### Otomatik Yedekleme

| Ayar | Açıklama |
|------|----------|
| Zamanlama | Cron ifadesi (örn. `0 3 * * *` = her gün saat 3'te) |
| Tutma | Saklanacak yedek sayısı |
| Telegram | Yedek dosyasını yönetici sohbetine gönder |
| S3 | S3 uyumlu depolamaya yükle |

### Yedekleme Hedefleri

=== "Telegram"

    Yedekler yapılandırılan yönetici sohbet ID'sine dosya olarak gönderilir.
    Maksimum dosya boyutu: 50 MB (Telegram limiti). Daha büyük yedekler bölünür.

=== "S3"

    S3 uyumlu depolamayı yapılandırın:
    ```
    VORTEX_BACKUP_S3_BUCKET=my-backups
    VORTEX_BACKUP_S3_ENDPOINT=s3.amazonaws.com
    VORTEX_BACKUP_S3_ACCESS_KEY=xxxx
    VORTEX_BACKUP_S3_SECRET_KEY=xxxx
    VORTEX_BACKUP_S3_REGION=us-east-1
    ```

=== "Yerel"

    Yedekler `/var/lib/vortexui/backups/` konumunda saklanır (varsayılan).
    Disk dolmasını önlemek için tutma politikasını yapılandırın.

### Geri Yükleme

```bash
vortexui restore /path/to/backup.tar.gz
```

Veya arayüz ile: **Ayarlar → Yedekleme → Geri Yükle → Yedek dosyası yükle**.

!!! warning
    Geri yükleme mevcut veritabanının üzerine yazar. Geri yüklemeden önce taze bir yedek oluşturun.

---

## Derin Bağlantı Yapılandırması

**Ayarlar → Derin Bağlantılar**

Yerel uygulama entegrasyonu için abonelik derin bağlantılarını yapılandırın:

| Ayar | Açıklama |
|------|----------|
| Temel URL | Panelin genel URL'si |
| Uygulama şeması | URL şeması (örn. `vortex://`, `clash://`) |
| Sunucu adını dahil et | Bağlantılara sunucu tanımlayıcısı ekle |
| QR logosu | QR kodu merkezinde gösterilen özel logo |

---

## Güncelleme Denetleyicisi

**Ayarlar → Güncellemeler**

- Yeni VortexUI sürümlerini kontrol edin
- Mevcut güncellemeler için değişiklik günlüğünü görüntüleyin
- Tek tıkla güncelleme (dahili olarak `vortexui update` çalıştırır)
- Otomatik kontrol aralığını yapılandırın

---

## Yapılandırma Şablonları

**Ayarlar → Yapılandırma Şablonları**

Abonelikler için Clash/sing-box yapılandırma çıktısını özelleştirin:

### Clash Şablonu

Varsayılan Clash YAML yapısını geçersiz kılın:

- Özel proxy grupları (url-test, fallback, load-balance)
- Özel kurallar (reklamları engelle, yerel trafiği doğrudan yönlendir)
- DNS yapılandırması
- Özel yönlendirme kuralları

### sing-box Şablonu

Varsayılan sing-box JSON yapısını geçersiz kılın:

- Özel giden bağlantı grupları
- Yönlendirme kuralları
- DNS ayarları
- Deneysel özellikler

Şablonlar değişkenleri destekler:

| Değişken | Karşılığı |
|----------|-----------|
| `{PROXIES}` | Kullanıcının gelen bağlantılarından proxy yapılandırmaları listesi |
| `{USERNAME}` | Kullanıcının kullanıcı adı |
| `{SUB_URL}` | Abonelik URL'si |

---

## Uluslararasılaştırma

Panel tam RTL desteği ile **8 dili** destekler:

- English (EN)
- فارسی (FA)
- Türkçe (TR)
- العربية (AR)
- Русский (RU)
- 中文 (ZH)
- 日本語 (JA)
- Español (ES)

Başlıktaki açılır menüden dil değiştirin. Her yöneticinin tercihi bağımsız olarak kaydedilir.
