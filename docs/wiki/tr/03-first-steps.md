# İlk Adımlar

!!! info "Tahmini Süre"
    Girişten çalışan bir abonelik bağlantısına: **5 dakika**.

---

## 1. Giriş Yapın ve Hesabınızı Güvenceye Alın

1. Panel URL'nize gidin (örn. `https://panel.example.com`)
2. Kurulum sırasında oluşturduğunuz yönetici kimlik bilgileriyle giriş yapın
3. **Şifrenizi hemen değiştirin**: Ayarlar → Profil → Şifre Değiştir

### TOTP 2FA'yı Etkinleştirin

1. **Ayarlar → İki Faktörlü Doğrulama** bölümüne gidin
2. **Etkinleştir**'e tıklayın — QR kodunu Google Authenticator veya herhangi bir TOTP uygulamasıyla tarayın
3. Onaylamak için 6 haneli kodu girin
4. Kurtarma kodlarınızı güvenli bir yerde saklayın

!!! warning
    2FA olmadan, şifrenizi bilen herkes panele tam erişime sahip olur. Paneli internete açmadan önce etkinleştirin.

---

## 2. İlk Düğümünüzü Ekleyin

=== "Yerel Düğüm (Tek Sunucu)"

    Panel ve proxy çekirdeğiniz aynı makinede çalışıyorsa:

    1. **Düğümler → Düğüm Ekle** bölümüne gidin
    2. **Yerel Düğüm** seçin
    3. Çekirdek seçin: **Xray-core** veya **sing-box**
    4. **Oluştur**'a tıklayın
    5. Düğüm otomatik başlar ve "Çevrimiçi" olarak görünür

=== "Uzak Düğüm (Kayıt Sihirbazı)"

    Ayrı bir sunucu için:

    1. **Düğümler → Düğüm Ekle** bölümüne gidin
    2. **Uzak Düğüm** seçin
    3. Sihirbaz tek satırlık bir kurulum komutu gösterir — kopyalayın
    4. Uzak sunucunuza SSH ile bağlanıp komutu yapıştırın
    5. Ajanın bağlanmasını bekleyin — düğüm "Çevrimiçi" olarak görünür

!!! tip
    Kayıt sihirbazı sertifika değişimini, çekirdek kurulumunu ve servis kaydını yönetir. Manuel sertifika yönetimi gerekmez.

---

## 3. İlk Gelen Bağlantınızı Oluşturun

Hızlı bir VLESS + REALITY kurulumu (sansür dayanıklı bağlantılar için önerilir):

1. **Gelen Bağlantılar → Gelen Bağlantı Ekle** bölümüne gidin
2. Yapılandırın:

    | Alan | Değer |
    |------|-------|
    | Protokol | VLESS |
    | Düğüm | Düğümünüz |
    | Port | `443` |
    | Taşıma | TCP |
    | Güvenlik | REALITY |
    | Hedef (SNI) | `www.google.com:443` (veya Reality Tarayıcı kullanın) |
    | Sunucu Adları | `www.google.com` |
    | Kısa ID'ler | Otomatik oluşturulur (veya kendiniz girin) |

3. **Oluştur**'a tıklayın

!!! tip "Reality Tarayıcı"
    Sunucu konumunuz için en iyi SNI alan adlarını bulmak için **Güvenlik → Reality Tarayıcı**'yı kullanın. Düşük gecikme ve kararlı bağlantılar için en yüksek puanlı sonucu seçin.

---

## 4. İlk Kullanıcınızı Oluşturun

1. **Kullanıcılar → Yeni Kullanıcı** bölümüne gidin
2. Doldurun:

    | Alan | Değer |
    |------|-------|
    | Kullanıcı adı | `testuser` |
    | Veri limiti | `50 GB` (veya sınırsız için `0`) |
    | Süre | Şu andan 30 gün sonra |
    | Cihaz limiti | `3` |
    | Gelen bağlantılar | VLESS gelen bağlantınızı seçin |

3. **Oluştur**'a tıklayın
4. Kullanıcının abonelik tokeni otomatik olarak oluşturulur

---

## 5. Abonelik Bağlantısını Test Edin

1. Kullanıcı listesinden kullanıcı adına tıklayarak detay sayfasını açın
2. **Abonelik Bağlantısı**'nı kopyalayın (veya QR kodunu tarayın)
3. İstemci uygulamanıza aktarın:

=== "v2rayNG (Android)"

    1. v2rayNG'yi açın → **+** dokunun → **Panodan yapılandırma içe aktar**
    2. Veya QR kodunu doğrudan tarayın

=== "Clash Meta (Masaüstü/Mobil)"

    ```
    https://panel.example.com/sub/<token>?format=clash
    ```
    Clash'te abonelik URL'si olarak ekleyin.

=== "sing-box (iOS/Android)"

    ```
    https://panel.example.com/sub/<token>?format=singbox
    ```
    Uzak profil olarak ekleyin.

4. Bağlanın ve trafiğin aktığını doğrulayın — panelde kullanıcının kullanımını kontrol edin

---

## 6. Bildirimleri Etkinleştirin

Bilgilendirilmek için Telegram bildirimlerini ayarlayın:

1. **Ayarlar → Bildirimler → Telegram** bölümüne gidin
2. **Bot Tokeni**'nizi girin ([@BotFather](https://t.me/BotFather) ile oluşturun)
3. **Yönetici Sohbet ID**'nizi girin ([@userinfobot](https://t.me/userinfobot) ile alın)
4. Teslimayı doğrulamak için **Test** butonuna tıklayın
5. İstediğiniz olayları etkinleştirin:
    - Kullanıcı oluşturuldu/süresi doldu/limite ulaştı
    - Düğüm çevrimdışı
    - Kota eşiğine ulaşıldı
    - Yedekleme tamamlandı

---

## Sırada Ne Var?

Artık bir düğüm, bir gelen bağlantı ve bir kullanıcı ile çalışan bir VortexUI kurulumunuz var. Buradan devam edin:

- **[Gösterge Paneli](04-dashboard.md)** — gerçek zamanlı izlemeyi keşfedin
- **[Kullanıcı Yönetimi](05-user-management.md)** — toplu işlemler, kotalar, portal
- **[Düğüm Yönetimi](06-node-management.md)** — daha fazla düğüm ekleyin, otomatik göçü yapılandırın
- **[Planlar ve Ödemeler](09-plans-payments.md)** — self-servis plan satın almalarını ayarlayın
- **[Güvenlik](08-security-administration.md)** — sansür karşıtı özellikleri yapılandırın
