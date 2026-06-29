# Güvenlik ve Yönetim

!!! important "Birinci Öncelik"
    Paneli internete açmadan önce birincil yönetici için **2FA** etkinleştirin ve güçlü bir JWT gizli anahtarı (≥32 byte) ayarlayın.

---

## Kimlik Doğrulama

| Katman | Mekanizma |
|--------|-----------|
| Panel girişi | JWT (Bearer) — yapılandırılabilir TTL |
| 2FA | TOTP (Google Authenticator, Authy, vb.) |
| API otomasyonu | Kişisel Erişim Tokeni (PAT) |
| Panel ↔ Düğüm | mTLS (karşılıklı sertifikalar) |
| Portal girişi | Abonelik tokeni |

### 2FA (TOTP) Etkinleştirme

1. **Ayarlar → İki Faktörlü Doğrulama**
2. **Etkinleştir**'e tıklayın → QR kodunu TOTP uygulamanızla tarayın
3. Onaylamak için 6 haneli kodu girin
4. Kurtarma kodlarını güvenli bir yerde saklayın
5. Devre dışı bırakmak için: mevcut kodu girin → **Devre Dışı Bırak**'a tıklayın

---

## RBAC ve Roller

**Yöneticiler → Roller**

| Kavram | Açıklama |
|--------|----------|
| Rol | İsimlendirilmiş izin seti |
| İzin | Ayrıntılı erişim: `users.read`, `nodes.write`, `plans.manage`, vb. |
| Bayi kotası | Sudo olmayan yöneticiler için kullanıcı sayısı + trafik sınırı |
| Kapsamlı görünüm | Yönetici yalnızca kendi kullanıcılarını/verilerini görür |

### Dahili Roller

| Rol | Erişim |
|-----|--------|
| **Sudo** | Her şeye tam erişim |
| **Admin** | Kullanıcıları, düğümleri, gelen bağlantıları yönet (sistem ayarları hariç) |
| **Bayi** | Kendi kullanıcılarını, planlarını, cüzdanını, markalamasını yönet |

---

## Bayi Platformu

VortexUI'nın bayi sistemi, her sudo olmayan yöneticiyi eksiksiz bir alt-panel operatörüne dönüştürür.

### Bayi Özellikleri

| Özellik | Açıklama |
|---------|----------|
| **Sahipli kullanıcılar** | Bayi yalnızca oluşturduğu kullanıcıları yönetir |
| **Sahipli planlar** | Bayi kendi fiyatlandırmasıyla kendi planlarını oluşturur |
| **Ödeme yapılandırması** | Bayi kendi ödeme yöntemlerini yapılandırır |
| **Cüzdan** | Trafik + kullanıcı kredileri, defter, yükleme talepleri |
| **Alt bayiler** | Miras alınan kapsamla alt bayiler oluşturma |
| **Beyaz etiket** | Özel markalama (logo, başlık, vurgu rengi, alt bilgi) |
| **Webhooklar** | Otomasyon için giden olaylar (user.created, user.deleted) |
| **Gösterge paneli** | Bayi'ye özel istatistikler (hesap sayısı, trafik, en aktif kullanıcılar) |
| **CSV dışa aktarma** | Sahipli kullanıcı listesini dışa aktar |

### Kapsamlı İzin Listeleri

**Yöneticiler → Yönetici düzenle** — bayinin neye erişebileceğini kısıtlayın:

| İzin Listesi | Etki |
|--------------|------|
| Planlar | Yalnızca listelenen planlar mağazasında görünür |
| Düğümler | Bayi yalnızca bu düğümlere kullanıcı bağlayabilir |
| Gelen bağlantılar | Bayi yalnızca bu gelen bağlantılara kullanıcı ekleyebilir |

Boş izin listesi = kısıtlama yok (tüm öğeler görünür). Sudo yöneticiler asla kapsamlanmaz.

### Trafik Kotası Modları

| Mod | Havuz ne zaman azalır... |
|-----|--------------------------|
| **Atanmış** | Bayi kullanıcılara veri limiti atar (limitlerin toplamı) |
| **Tüketilen** | Kullanıcılar gerçekten trafik tüketir |

### Bayi Bazlı Ödeme Yapılandırması

Her bayi kendi ödeme yöntemlerini yapılandırır:

| Yöntem | Yapılandırma |
|--------|-------------|
| ZarinPal | Satıcı ID |
| Karttan karta | Kart numarası + kart sahibi adı |
| Kripto | Cüzdan adresleri (BTC, USDT, vb.) |

O bayinin mağazasındaki kullanıcılar yalnızca bayilerinin ödeme seçeneklerini görür.

### Bayi Bazlı Sahipli Planlar

Her bayi yalnızca kendi kullanıcılarına görünen planlar oluşturur:

- Bayi fiyat, veri limiti, süre, cihaz limitini belirler
- Planlar o bayinin kullanıcıları için `/sub/{token}/shop` sayfasında görünür
- Yönetici planları bayi planlarından ayrıdır

### Bayi Cüzdanı

| Kredi Türü | Amaç |
|------------|------|
| Trafik kredileri | Kullanıcılar veri kullandığında tüketilir |
| Kullanıcı kredileri | Yeni kullanıcılar oluşturulduğunda tüketilir |

Cüzdan işlemleri:

- **Yükleme talebi** — bayi kredi talep eder, sudo onaylar
- **Otomatik düşüm** — kullanıcılar tükettikçe/oluşturuldukça sistem düşer
- **Defter** — nedenlerle birlikte tüm kredi değişikliklerinin tam geçmişi

### Alt Bayiler

Bir bayi alt bayiler oluşturabilir:

- Alt bayi, ebeveynin kapsamını miras alır (ebeveynin izin listelerini aşamaz)
- Ebeveyn, alt bayinin cüzdanını yönetir
- Ebeveyn, alt bayinin kullanıcılarını ve istatistiklerini görür

### Politika Limitleri (bayi başına)

**Yöneticiler → Yönetici düzenle → Politika** bölümünde ayarlanır:

| Ayar | Etki |
|------|------|
| Maks. veri limiti (GB) | Bayinin ayarlayabileceği kullanıcı başına veri limiti tavanı |
| Maks. süre (gün) | Abonelik süresi tavanı |
| Toplu oluştur/içe aktar izin ver | Toplu kullanıcı oluşturmayı kontrol et |
| Toplu silme izin ver | Çoklu silmeyi kontrol et |

### Otomatik Askıya Alma

| Ayar | Etki |
|------|------|
| Otomatik askıya almayı etkinleştir | Ana anahtar |
| IP ihlalleri (7g) | N paylaşım-tespit olayından sonra askıya al |
| Kota toleransı (dakika) | Kota aşıldıktan sonra tolerans süresi |

Askıya alınan bayiler, bir sudo yönetici **Askıyı Kaldır**'a tıklayana kadar giriş yapamaz.

### Sudo Yönetici Araçları

| Eylem | Açıklama |
|-------|----------|
| Kota kullanım tablosu | Bayi bazında hesaplar, trafik, kalan havuz |
| Hızlı kota ayarlama | +50 hesap / +10 GB / +50 GB butonları |
| Olarak giriş yap (taklit) | Bayi JWT oturumu düzenle |
| Askıyı kaldır | Otomatik askıya alma bayrağını temizle |
| Kota uyarıları | Tüm bayiler için eşik bildirimlerini yapılandır |

---

## TLS Hileleri Yöneticisi

**Güvenlik → TLS Hileleri**

Birden fazla DPI bypass tekniğini birleştiren ISP'ye özel profiller:

| Teknik | Açıklama |
|--------|----------|
| Fragment | TLS ClientHello'yu küçük paketlere böl |
| Mux | Bağlantıları tek bir akışta çoğulla |
| Padding | Paketlere rastgele dolgulama ekle |

### Profil Oluşturma

1. **Yeni Profil**'e tıklayın
2. Adlandırın (örn. "İran — Fragment + Chrome")
3. Yapılandırın:
    - Parmak izi: Chrome, Firefox, Safari, Random, Randomized
    - Fragment: Etkinleştir + uzunluk aralığı (örn. `10-30`) + aralık
    - Mux: Etkinleştir + protokol (smux, yamux, h2mux)
4. Kaydet → gelen bağlantılara ata

!!! example "Ülke Önayarları"
    - **İran**: Fragment `10-30` + Chrome parmak izi
    - **Çin**: Mux h2mux + Randomized parmak izi
    - **Rusya**: Fragment `1-3` + Firefox parmak izi

---

## Problama Koruması

**Güvenlik → Problama Koruması**

Sansürcülerden (GFW, TSPU) gelen aktif problama girişimlerini tespit etme ve engelleme.

| Eylem | Davranış |
|-------|----------|
| **Engelle** | Bağlantıyı düşür + IP'yi yapılandırılan süre boyunca yasakla |
| **Bal küpü** | Problayıcıyı kandırmak için sahte bir web sitesi döndür |
| **Yalnızca günlük** | Eylem almadan kaydet (izleme modu) |

Yapılandırma:

- Engel süresi (varsayılan: 3600sn)
- Maks. prob/dakika eşiği (varsayılan: 5)
- Güvenilir IP beyaz listesi
- Tespit durumunda Telegram uyarısı

---

## İstemci Parmak İzi Doğrulama (JA3)

**Güvenlik → Parmak İzi**

TLS ClientHello parmak izlerine göre bağlantıları engelleme. Bilinen tarayıcı araçları (curl, Go HTTP, Python requests) ayırt edici parmak izleri üretir.

| Ayar | Açıklama |
|------|----------|
| Etkin | Parmak izi kontrolünü etkinleştir |
| Varsayılan eylem | Bilinmeyen parmak izleri: İzin Ver / Engelle / Günlükle |
| Kurallar | Parmak izi veya JA3 hash başına açık izin ver/engelle |

!!! example
    Tarayıcı araçlarını engelle:
    - Kural 1: `fingerprint=curl`, eylem=engelle
    - Kural 2: `fingerprint=python`, eylem=engelle

---

## Sahte Web Sitesi

**Güvenlik → Sahte Web Sitesi**

Birisi sunucunuzun IP'sini doğrudan ziyaret ettiğinde sahte bir web sitesi gösterin:

| Mod | Davranış |
|-----|----------|
| **Proxy** | Mevcut bir web sitesini ters-proxy ile aynala |
| **Statik** | Özel HTML sun |

Sunucunuzun sansürcülere ve rastgele ziyaretçilere normal bir web sitesi gibi görünmesini sağlar.

---

## Gizlenme Profilleri

**Güvenlik → Gizlenme Profilleri**

Tek tıkla sansür kaçınma için gelen bağlantılara atanan önceden yapılandırılmış anti-DPI teknik paketleri. Fragment, mux ve parmak izi ayarlarını isimlendirilmiş bir profilde birleştirir.

---

## WARP+ Entegrasyonu

**Ağ → Giden Bağlantılar → WARP+**

Temiz IP almak için trafiği Cloudflare WARP üzerinden yönlendirin:

- Ücretsiz katman veya WARP+ lisans anahtarı ile
- Belirli alan adları için yönlendirme kurallarına atayın
- Düğüm IP'si servisler tarafından işaretlendiğinde kullanışlıdır

---

## DNS-over-HTTPS (DoH)

**Güvenlik → DNS-over-HTTPS**

DNS sızıntılarını önleyen dahili DoH sunucusu:

| Ayar | Açıklama | Varsayılan |
|------|----------|------------|
| Etkin | DoH'u aç/kapat | Kapalı |
| Dinleme adresi | Bağlama adresi | `:8053` |
| Üst DNS | Yönlendirilecek çözümleyiciler | `1.1.1.1`, `8.8.8.8` |
| Reklamları engelle | Reklam alan adlarını filtrele | Kapalı |
| Kötü yazılımları engelle | Kötü amaçlı yazılım alan adlarını filtrele | Açık |
| Özel engel listesi | Kendi engellenen alan adlarınız | Boş |
| Önbellek TTL | Önbellek süresi (saniye) | 300 |

---

## Temiz-IP Tarayıcı

**Ağ → Temiz IP**

Gecikme ve paket kaybına göre tarayarak ve puanlayarak iyi performans gösteren CDN kenar IP'leri bulun.

1. Aday IP'leri yapıştırın (satır başına bir)
2. Port ayarlayın (varsayılan: 443)
3. **Tara**'ya tıklayın — sonuçlar en iyiden sıralanır
4. En iyi IP'leri abonelik sunucularında veya CDN zincirlerinde kullanın

!!! warning "SSRF Koruması"
    Dahili ağ taramasını önlemek için özel, geri döngü ve bağlantı-yerel aralıklar reddedilir.

---

## IP-Limit Uygulaması

**Güvenlik → IP Limiti**

Hesap paylaşımını önlemek için kullanıcı başına eş zamanlı IP/cihaz sınırları.

| Ayar | Açıklama |
|------|----------|
| Etkin | Uygulamayı aç/kapat |
| Eylem | `warn`, `disable_temporarily` veya `kill_connections` |
| Uyarı bekleme süresi | Tekrarlanan uyarılar arası saniye |
| Sonra geri yükle | Geçici devre dışı bırakılan kullanıcının geri yüklenmesine kadar saniye |

!!! note
    `kill_connections` yalnızca Xray'e özgüdür. sing-box düğümlerinde `disable_temporarily`'ye düşer.

---

## IP Beyaz Liste/Kara Liste

**Ayarlar → IP Koruması**

Panel API ve abonelik erişimini IP'ye göre kısıtlayın:

- **Beyaz liste modu** — yalnızca listelenen IP'ler erişebilir
- **Kara liste modu** — listelenen IP'ler engellenir
- CIDR aralıklarını destekler

---

## Gelen Bağlantı Başına Coğrafi Engelleme

Belirli bir gelen bağlantıya hangi ülkelerin bağlanabileceğini kısıtlayın:

- Gelen bağlantı düzenleme sayfasında yapılandırın
- Virgülle ayrılmış ISO 3166-1 alpha-2 kodları
- Boş = tüm ülkelere izin verilir

---

## Hesap Paylaşım Koruması

Çevrimiçi IP'leri cihaz limitleriyle karşılaştıran arka plan döngüsü:

| Mod | Davranış |
|-----|----------|
| Tespit (varsayılan) | `user.ip_limit` olayı + webhook/Telegram tetikle |
| Otomatik limit (`VORTEX_SHARE_AUTOLIMIT=true`) | Kullanıcıyı otomatik olarak limitle |

---

## Denetim Günlüğü

**Denetim** — tüm yönetici değişikliklerini kaydeder:

| Alan | İçerik |
|------|--------|
| Aktör | Yönetici kullanıcı adı |
| Eylem | `user.create`, `inbound.update`, `admin.login`, vb. |
| Hedef | Kaynak ID'si |
| Zaman damgası | ISO 8601 |
| Fark | Önce/sonra JSON |

Bayiler yalnızca kendi eylemlerini görür. Sudo yöneticiler tümünü görür.

---

## API Tokenleri (PAT)

**Ayarlar → API Tokenleri**

Otomasyon için kişisel erişim tokenleri oluşturun:

```bash
curl -H "Authorization: Bearer <PAT>" \
  https://panel.example.com/api/users
```

- Her token ayrı ayrı iptal edilebilir
- İzinler oluşturan yöneticinin rolünden miras alınır
- İsteğe bağlı son kullanma tarihi belirleyin

---

## Güvenlik Kontrol Listesi

- [ ] Güçlü JWT gizli anahtarı (≥32 byte rastgele)
- [ ] HTTPS etkin (Caddy ile Let's Encrypt)
- [ ] Sudo yönetici için TOTP 2FA
- [ ] En az ayrıcalıklı API tokenleri
- [ ] Şifreli site dışı yedekleme
- [ ] HMAC doğrulaması için webhook gizli anahtarı
- [ ] Panel portu herkese kapalı (yalnızca Caddy 443)
- [ ] Problama koruması etkin
- [ ] IP-limit uygulaması yapılandırılmış
- [ ] Sahte web sitesi aktif
