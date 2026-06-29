# Gösterge Paneli

!!! info "Gerçek Zamanlı"
    Tüm gösterge paneli verileri Server-Sent Events (SSE) ile güncellenir — manuel yenileme gerekmez.

---

## Genel Bakış Widget'ları

Ana gösterge paneli **özelleştirilebilir bir widget ızgarasıdır**. Widget'ları sürükleyip bırakarak ve yeniden boyutlandırarak tercih ettiğiniz düzeni oluşturun.

### Mevcut Widget'lar

| Widget | İçerik |
|--------|--------|
| Kullanıcı özeti | Toplam, aktif, limitli, süresi dolmuş sayıları |
| Trafik özeti | Bugünün yükleme/indirme, aylık toplam |
| Düğüm durumu | Çevrimiçi/çevrimdışı düğüm sayıları, sağlık göstergeleri |
| Sistem göstergeleri | CPU, RAM, disk kullanımı (animasyonlu halkalar) |
| Aktif bağlantılar | Mevcut canlı tünel sayısı |
| Son olaylar | En son sistem olayları (kullanıcı oluşturuldu, düğüm çevrimdışı, vb.) |
| Hızlı eylemler | Kullanıcı oluştur, düğüm ekle, tarama çalıştır kısayolları |
| Bayi havuzu | Hesap/trafik kota bakiyeniz (bayi görünümü) |

### Özelleştirme

1. **Izgara simgesine** tıklayın (gösterge paneli sağ üst)
2. Widget'ları yeniden düzenlemek için sürükleyin
3. Köşe tutamağını sürükleyerek yeniden boyutlandırın
4. Onay kutularıyla widget görünürlüğünü değiştirin
5. **Düzeni Kaydet**'e tıklayın — yönetici hesabı başına kalıcıdır

---

## Gerçek Zamanlı Göstergeler

Animasyonlu göstergeler her düğüm için canlı sistem metriklerini gösterir:

- **CPU** — mevcut kullanım yüzdesi
- **RAM** — kullanılan / toplam, yüzde ile
- **Bant Genişliği** — mevcut aktarım hızı (yükleme + indirme)
- **Bağlantılar** — aktif tünel sayısı

Göstergeler SSE push ile her 3 saniyede güncellenir.

---

## Grafikler

| Grafik | Açıklama |
|--------|----------|
| Trafik (çizgi) | Zamana göre yükleme/indirme (saatlik/günlük/haftalık) |
| Bağlantılar (alan) | Zamana göre aktif bağlantılar |
| Kullanıcılar (çubuk) | Gün/hafta bazında yeni kullanıcılar |
| Düğüm yükü (yığın) | Düğüm başına CPU dağılımı |

Tüm grafikler zaman aralığı seçimini destekler: **24s**, **7g**, **30g** veya özel tarih aralığı.

---

## Dünya Haritası Coğrafi Görselleştirme

Düğüm ajanlarından gelen GeoIP verilerine dayalı olarak kullanıcılarınızın nereden bağlandığını gösteren bir ısı haritası.

- Bağlantı sayısı ve trafik hacmi için bir ülkenin üzerine gelin
- Ayrıntılı dağılım için tıklayın (en aktif şehirler, o ülkeden en aktif kullanıcılar)
- Renk yoğunluğu trafik hacmini yansıtır

!!! note
    Coğrafi veriler, düğüm ajanlarının bağlantı meta verilerini raporlamasını gerektirir. Harita boşsa, düğümlerinizin en son ajan sürümünü çalıştırdığını doğrulayın.

---

## Monitör Sayfası

**Gösterge Paneli → Monitör** — tüm aktif bağlantıların gerçek zamanlı görünümü.

| Sütun | Açıklama |
|-------|----------|
| Kullanıcı | Kullanıcı adı |
| Düğüm | Bağlı olduğu sunucu |
| IP | İstemci kaynak IP'si |
| Protokol | VLESS, VMess, Trojan, vb. |
| Süre | Bağlantının ne kadar süredir aktif olduğu |
| Trafik | Bu oturum için yükleme/indirme |

Özellikler:

- Her 3 saniyede otomatik yenileme
- Düğüm, protokol veya kullanıcıya göre filtreleme
- Herhangi bir sütuna göre sıralama
- Detay sayfasına gitmek için kullanıcıya tıklayın

---

## Analitik Sayfası

**Gösterge Paneli → Analitik** — ülke, kullanıcı ve zamana göre toplu trafik bilgileri.

### Bölümler

| Bölüm | Gösterir |
|-------|----------|
| Özet kartları | Toplam yükleme, indirme, ülke sayısı |
| Ülkeye Göre Trafik | Coğrafi dağılım — ülke, bağlantılar, baytlar |
| En Aktif Kullanıcılar | Toplam tüketilen trafiğe göre sıralı |
| Yoğun Saatler | Saatlik trafik hacmi çubuk grafiği |

### Zaman Aralıkları

**Son 24 saat**, **Son 7 gün**, **Son 30 gün** veya özel tarih aralığı seçin.

### Dışa Aktarma

Seçilen zaman aralığı için coğrafi + kullanıcı verilerini tablo olarak indirmek için **CSV Dışa Aktar** butonuna tıklayın.

!!! tip
    En yoğun saatlerinizi belirlemek ve düğüm kapasitesini buna göre planlamak için analitik kullanın.

---

## Komut Paleti

Tüm panelde bulanık arama yapmak için ++ctrl+k++ (veya macOS'ta ++cmd+k++) tuşlarına basın.

### Arama Yapılabilecekler

- **Kullanıcılar** — ada göre herhangi bir kullanıcıya atlayın
- **Düğümler** — düğüm detayına gidin
- **Gelen Bağlantılar** — gelen bağlantı yapılandırmasını açın
- **Sayfalar** — herhangi bir panel sayfasına atlayın
- **Eylemler** — kullanıcı oluştur, düğüm ekle, yedekleme çalıştır, vb.

### Klavye Kısayolları

| Kısayol | Eylem |
|---------|-------|
| ++ctrl+k++ | Komut paletini aç |
| ++ctrl+shift+n++ | Yeni kullanıcı |
| ++ctrl+shift+d++ | Koyu/açık modu değiştir |
| ++ctrl+period++ | Bildirimleri aç |
| ++escape++ | Modal/paleti kapat |
| ++g++ ardından ++d++ | Gösterge paneline git |
| ++g++ ardından ++u++ | Kullanıcılara git |
| ++g++ ardından ++n++ | Düğümlere git |
