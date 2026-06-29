# Kullanıcı Yönetimi

!!! abstract "Kullanıcı-Merkezli Model"
    Bir kullanıcı = bir abonelik tokeni = tüm düğümlerdeki atanmış tüm gelen bağlantılara erişim.
    Protokol başına ayrı hesap gerekmez.

---

## Kullanıcı CRUD

### Kullanıcı Oluşturma

**Kullanıcılar → Yeni Kullanıcı**

| Alan | Açıklama |
|------|----------|
| Kullanıcı adı | Benzersiz tanımlayıcı |
| Veri limiti | Trafik sınırı (bayt) — `0` = sınırsız |
| Bitiş tarihi | Abonelik sona erme tarihi |
| Cihaz limiti | Maksimum eş zamanlı cihaz |
| Sıfırlama stratejisi | `none` / `daily` / `weekly` / `monthly` |
| Durum | `active` / `disabled` / `limited` |
| Gelen bağlantılar | İzin verilen gelen bağlantılar (bir veya daha fazla seçin) |
| Not | Yalnızca yönetici notu |

### Toplu İşlemler

| Eylem | Açıklama |
|-------|----------|
| **Toplu Oluştur** | Kullanıcılar → Toplu Ekle — sayı belirtin veya CSV yükleyin |
| **Çoklu seçim** | Birden fazla kullanıcıyı işaretleyin → eylem uygulayın |
| **Toplu eylemler** | Etkinleştir, devre dışı bırak, sil, uzat, trafiği sıfırla |
| **İçe aktar** | Kullanıcılar → İçe Aktar — 3x-ui veya Marzban veritabanından |

---

## Kotalar ve Limitler

### Veri Limiti

Kullanıcı başına trafik sınırı belirleyin. Aşıldığında kullanıcının durumu `limited` olur ve `user.limited` olayı tetiklenir.

### Cihaz Limiti

Maksimum eş zamanlı bağlantılar (farklı IP'ler). Uygulama [IP-limit sistemi](08-security-administration.md) tarafından yönetilir.

### Süre Sonu

Kullanıcılar yapılandırılan tarihte sona erer. Sistem 3 gün önce `user.expiry_warning` ve sona ermede `user.expired` olayını tetikler.

### Sıfırlama Stratejisi

| Strateji | Davranış |
|----------|----------|
| `none` | Asla sıfırlamaz — trafik sonsuza kadar birikir |
| `daily` | Gece yarısı UTC'de sıfırla |
| `weekly` | Her Pazartesi gece yarısı UTC'de sıfırla |
| `monthly` | Her ayın 1'inde gece yarısı UTC'de sıfırla |

---

## Abonelik Teslimi

### Abonelik Bağlantısı

Her kullanıcı benzersiz bir abonelik URL'si alır:

```
https://panel.example.com/sub/{token}
```

Yanıt biçimi istemcinin User-Agent'ından otomatik algılanır veya `?format=` ile zorlanır:

| Biçim | Parametre | Çıktı |
|-------|-----------|-------|
| Base64 | `?format=base64` | V2Ray uyumlu base64 kodlanmış bağlantılar |
| Clash | `?format=clash` | Clash YAML yapılandırması |
| sing-box | `?format=singbox` | sing-box JSON yapılandırması |
| Xray JSON | `?format=xray` | Ham Xray/V2Ray istemci JSON |
| Outline | `?format=outline` | Outline için `ss://` Shadowsocks bağlantıları |
| Düz bağlantılar | `?format=links` | V2rayN tarzı paylaşım bağlantıları, satır başına bir |

### Abonelik Sunucuları

Canlı çekirdek yapılandırmasına dokunmadan aboneliğin ne tanıttığını değiştiren gelen bağlantı başına geçersiz kılmalar.

| Alan | Açıklama |
|------|----------|
| Etiket | Görüntü adı (şablon değişkenlerini destekler) |
| Adres | Tanıtılan sunucu/IP (örn. bir CDN alan adı) |
| Port | Geçersiz kılma portu (`0` = miras al) |
| SNI | TLS sunucu adı |
| Host | HTTP `Host` başlığı |
| Yol | WS/HTTPUpgrade/gRPC yolu |
| ALPN | Virgülle ayrılmış (örn. `h2,http/1.1`) |
| Parmak izi | uTLS parmak izi (örn. `chrome`) |
| Güvenlik | `inbound_default`, `none`, `tls` veya `reality` |
| Güvensiz izin ver | Sertifika doğrulamayı atla |
| Mux | Çoğullamayı etkinleştir |
| Fragment | TLS fragment belirtimi (örn. `1,40-60,30-50`) |
| Öncelik | Gelen bağlantı içindeki sıra (düşük olan önce gösterilir) |

**Şablon Değişkenleri** — kullanıcı başına işlenir:

| Değişken | Karşılığı |
|----------|-----------|
| `{USERNAME}` | Kullanıcının kullanıcı adı |
| `{SERVER_IP}` | Düğümün adresi |
| `{SERVER_PORT}` | Tanıtılan port |
| `{PROTOCOL}` | Gelen bağlantı protokolü |
| `{NETWORK}` | Taşıma türü |
| `{SECURITY}` | Taşıma güvenliği |
| `{REMARK}` | Yapılandırılan etiket |

!!! tip
    Tek bir gelen bağlantıyı birden fazla CDN alan adı arkasına koymak için abonelik sunucularını kullanın — her biri farklı SNI ve yol ile, tümü şablon değişkenleri kullanan tek bir sunucu tanımından.

---

## Self-Servis Portal

**Son kullanıcı URL'si:** `/portal/login`

Kullanıcılar abonelik tokenleriyle giriş yapar ve şunlara erişir:

| Özellik | Açıklama |
|---------|----------|
| Gösterge paneli | Kullanım istatistikleri, kalan veri/süre, aktif cihazlar |
| Planlar | Abonelik planlarını göz atma ve satın alma (bayilerinin mağazasından) |
| Talepler | Destek talebi açma, yönetici mesajlarına yanıt verme |
| Referans | Referans kodunu görüntüleme/paylaşma, kazanılan ödülleri görme |
| QR Kodu | Mobil uygulamalara abonelik aktarmak için tarama |

### Self-Servis Mağaza

**URL:** `/sub/{token}/shop`

Her bayi kendi planlarını ve ödeme yöntemlerini yapılandırır. Son kullanıcılar yalnızca bayilerinin tekliflerini görür:

1. Kullanıcı mevcut planları göz atar
2. Plan ve ödeme yöntemi seçer (ZarinPal, karttan karta veya kripto)
3. Ödemeyi tamamlar (veya dekont yükler)
4. Sipariş otomatik olarak (ZarinPal) veya yönetici onayından sonra (kart/kripto) tamamlanır

Tam detaylar için [Planlar ve Ödemeler](09-plans-payments.md) sayfasına bakın.

---

## Aile/Grup Abonelikleri

**Kullanıcılar → Aile Grupları**

Kullanıcıların bir veri havuzunu paylaşmasına izin verin:

1. Paylaşılan veri limiti ile bir **Aile Grubu** oluşturun
2. Mevcut kullanıcıları üye olarak ekleyin
3. Her üyenin trafiği paylaşılan havuzdan düşer
4. Paylaşılan havuz içinde isteğe bağlı üye başına sınır

| Alan | Açıklama |
|------|----------|
| Ad | Grup adı |
| Sahip | Birincil hesap |
| Veri limiti | Toplam paylaşılan havuz |
| Maksimum üye | Limit (varsayılan: 5) |
| Üye kotası | Havuz içindeki üye başına sınır |

---

## Akıllı Kota

Kullanıcıları limitlerinde sert kesme yerine kademeli hız azaltma.

Katmanları JSON olarak yapılandırın:

```json
[
  { "threshold_pct": 80, "action": "warn", "speed_limit": 0 },
  { "threshold_pct": 95, "action": "throttle", "speed_limit": 524288 },
  { "threshold_pct": 100, "action": "disable" }
]
```

- **%80**'de → kullanıcıyı uyar (bildirim olayı)
- **%95**'te → 512 KB/s'ye kısıtla
- **%100**'de → hesabı devre dışı bırak

!!! info
    Akıllı Kota plan başına veya global olarak yapılandırılır. Plan bazlı ayarlar global varsayılanı geçersiz kılar.

---

## Referans Sistemi

**Kullanıcılar → Referanslar**

| Ayar | Açıklama | Varsayılan |
|------|----------|------------|
| Etkin | Referansları aç/kapat | Kapalı |
| Ödül türü | `data` (ek GB) veya `days` (ek süre) | `data` |
| Ödül miktarı | Referans başına ne kadar | 1 GB |
| Maksimum referans | Kullanıcı başına limit (`0` = sınırsız) | `0` |
| Ücretli gerektir | Yalnızca ödeme yapan referanslar için ödül | Kapalı |
| Her iki taraf ödüllendirilir | Referans veren + yeni kullanıcı | Evet |

Kullanıcılar referans kodlarına portal üzerinden erişir. Bir arkadaş kodu kullanarak kaydolduğunda her iki taraf ödüllendirilir.

---

## Yapılandırma Şablonları

Kullanıcıların aboneliklerinde aldıkları Clash/sing-box çıktısını özelleştirin:

- Özel yönlendirme kuralları ekleyin (reklamları engelle, yerel trafiği doğrudan yönlendir)
- DNS ayarlarını yapılandırın
- Proxy grubu stratejilerini ayarlayın (url-test, fallback, load-balance)
- Kullanıcı başına veya global şablonlar

---

## Derin Bağlantılar ve QR Kodları

**Sistem → Derin Bağlantılar**

Tek dokunuşla uygulama kurulumu için abonelik derin bağlantıları oluşturun:

| Ayar | Açıklama |
|------|----------|
| Temel URL | Panelin genel URL'si |
| Uygulama şeması | Yerel uygulamalar için URL şeması |
| Ad dahil et | Bağlantıya sunucu adını ekle |
| QR logosu | QR merkezi için özel logo |

---

## Diğer Panellerden İçe Aktarma

**Kullanıcılar → İçe Aktar**

Şuralardan kullanıcıları taşıyın:

- **3x-ui** — veritabanı dosya yolunu belirtin
- **Marzban** — veritabanı bağlantısı veya dışa aktarma dosyası sağlayın

İçe aktarıcı kullanıcıları eşler; kullanıcı adlarını, trafik limitlerini ve bitiş tarihlerini korur.
