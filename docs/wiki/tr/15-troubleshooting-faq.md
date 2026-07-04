# Sorun Giderme ve SSS

---

## Yaygın Sorunlar

### Bağlantı Reddedildi

**Belirti:** İstemciler proxy'ye bağlanamıyor.

**Kontrol:**

1. Düğüm çevrimiçi mi? **Düğümler** sayfasında durumu kontrol edin
2. Gelen bağlantı portu açık mı? `ss -tlnp | grep <port>`
3. Güvenlik duvarı trafiğe izin veriyor mu? `ufw status` veya `iptables -L`
4. Çekirdek çalışıyor mu? Düğüm günlüklerinde hata kontrol edin
5. İstemci yapılandırmasında protokol/taşıma doğru mu?

!!! tip
    Tüm bileşenleri tek seferde kontrol etmek için `vortexui doctor` çalıştırın.

### TLS Hataları

**Belirti:** `tls: handshake failure` veya `certificate verify failed`

**Kontrol:**

1. **REALITY:** Hedef/SNI alan adı düğümden erişilebilir mi? `curl -I https://hedef-alan` deneyin
2. **TLS:** Sertifika geçerli mi? Sona ermeyi kontrol edin: `openssl s_client -connect host:443`
3. **CDN:** Cloudflare SSL modu "Full (Strict)" olarak ayarlanmış mı?
4. **İstemci:** SNI alanı sunucu yapılandırmasıyla eşleşiyor mu?
5. **Fragment:** TLS fragment kullanıyorsanız geçici olarak devre dışı bırakmayı deneyin

### Düğüm Bağlantısı Kesildi

**Belirti:** Düğüm panelde "Çevrimdışı" gösteriyor.

**Kontrol:**

1. Düğüm sunucusu çalışıyor mu? SSH ile kontrol edin: `systemctl status vortex-node`
2. Ağ bağlantısı: düğümden `ping <panel-ip>`
3. mTLS sertifikaları: sona erme veya uyumsuzluk kontrol edin
4. Güvenlik duvarı: gRPC portu (varsayılan 9090) panel ve düğüm arasında açık mı?
5. Düğüm ajan günlüklerini kontrol edin: `journalctl -u vortex-node -n 50`

### Abonelik Boş

**Belirti:** İstemci boş abonelik alıyor (yapılandırma yok).

**Kontrol:**

1. Kullanıcının atanmış gelen bağlantıları var mı? Kullanıcı detayı → Gelen Bağlantılar kontrol edin
2. Atanmış gelen bağlantılar çevrimiçi düğümlerde mi?
3. Abonelik sunucuları doğru yapılandırılmış mı (kullanılıyorsa)?
4. Abonelik tokeni geçerli mi (iptal edilmemiş)?
5. Ham yanıtı kontrol edin: `curl https://panel.example.com/sub/<token>`

### Düğümde Yüksek CPU

**Belirti:** Düğüm CPU'su %90'ın üzerinde kalıyor.

**Kontrol:**

1. Çok fazla kullanıcı mı? Aktif bağlantı sayısını kontrol edin
2. Otomatik göç yapılandırılmış mı? Kullanıcıları uzaklaştırmalı
3. Çekirdek süreci: `top -p $(pgrep xray)` veya `pgrep sing-box`
4. Daha fazla düğüm eklemeyi ve yük dengelemeyi etkinleştirmeyi düşünün

### Veritabanı Bağlantı Sorunları

**Belirti:** Panel 500 hataları döndürüyor, günlükler PostgreSQL bağlantı hataları gösteriyor.

**Kontrol:**

1. PostgreSQL çalışıyor mu? `systemctl status postgresql`
2. `.env` dosyasında bağlantı dizesi doğru mu?
3. Maksimum bağlantı tükendi mi? `SELECT count(*) FROM pg_stat_activity;`
4. Bağlantı havuzlama için pgBouncer eklemeyi düşünün

---

## Hata Ayıklama İpuçları

### `vortexui doctor`

Kapsamlı tanılama çalıştırın:

```bash
vortexui doctor
```

Kontroller:

- ✅ PostgreSQL bağlantısı + şema sürümü
- ✅ Redis bağlantısı + gecikme
- ✅ Düğüm gRPC bağlantısı (düğüm başına)
- ✅ Sertifika geçerliliği
- ✅ Port erişilebilirliği
- ✅ DNS çözümleme
- ✅ Disk alanı
- ✅ Çekirdek ikili dosyası varlığı ve sürümü

### Sağlık Uç Noktası

```bash
curl https://panel.example.com/api/health
```

Bileşen durumunu döndürür:

```json
{
  "status": "healthy",
  "components": {
    "database": "ok",
    "redis": "ok",
    "nodes": { "online": 3, "offline": 0 }
  },
  "version": "1.2.8"
}
```

### Hata Ayıklama Günlüğünü Etkinleştirme

```bash
VORTEX_LOG_LEVEL=debug systemctl restart vortexui
```

Ardından günlükleri izleyin:

```bash
journalctl -u vortexui -f
```

!!! warning
    Hata ayıklama günlüğü ayrıntılıdır. Disk dolmasını önlemek için sorun giderdikten sonra devre dışı bırakın.

### Aboneliği Manuel Test Etme

```bash
# Base64 biçimi
curl -s https://panel.example.com/sub/<token>

# Clash biçimi
curl -s "https://panel.example.com/sub/<token>?format=clash"

# User-Agent algılama ile
curl -s -A "clash-meta" https://panel.example.com/sub/<token>
```

### API ile Düğüm Sağlığını Kontrol Etme

```bash
curl -H "Authorization: Bearer <token>" \
  https://panel.example.com/api/nodes/<id>/health
```

---

## SSS

### Yönetici şifresini nasıl sıfırlarım?

```bash
vortexui admin reset-password --username admin
```

Veya etkileşimli menü ile:

```bash
vortexui
# Seçenek 7 → Şifre sıfırla
```

### 3x-ui'den nasıl göç ederim?

1. 3x-ui veritabanınızı dışa aktarın (`x-ui.db`)
2. VortexUI'da: **Kullanıcılar → İçe Aktar → 3x-ui**
3. Veritabanı dosyasını yükleyin
4. Gelen bağlantıları eşleyin (VortexUI kullanıcıları eşleşen gelen bağlantılara atar)
5. İnceleyin ve onaylayın

### Marzban'dan nasıl göç ederim?

1. VortexUI'da: **Kullanıcılar → İçe Aktar → Marzban**
2. Veritabanı bağlantı dizesi veya dışa aktarma dosyası sağlayın
3. Kullanıcılar, trafik verileri ve bitiş tarihleri korunur
4. Gelen bağlantı eşlemesi mümkün olduğunda otomatik yapılır

### Panel ve düğümü aynı sunucuda çalıştırabilir miyim?

Evet — **Yerel Düğüm** özelliğini kullanın. Proxy çekirdeği panel ile birlikte süreç içinde çalışır. Ajan gerekmez.

### Abonelik otomatik algılama nasıl çalışır?

Panel istemcinin `User-Agent` başlığını inceler:

- "clash" içeriyorsa → Clash YAML
- "sing-box" içeriyorsa → sing-box JSON
- "outline" içeriyorsa → Outline `ss://` bağlantıları
- Diğer → base64 kodlanmış paylaşım bağlantıları

`?format=` sorgu parametresi ile geçersiz kılın.

### HTTPS için alan adı nasıl eklerim?

1. Alan adınızın DNS A kaydını sunucu IP'nize yönlendirin
2. `.env` dosyasında `VORTEX_DOMAIN=alan-adiniz.com` ayarlayın
3. Paneli yeniden başlatın — Caddy otomatik olarak sertifika düzenler

### Yedekleme ve geri yükleme nasıl yapılır?

**Yedekleme:**
```bash
vortexui backup
# veya otomatik: VORTEX_BACKUP_CRON="0 3 * * *" ayarlayın
```

**Geri yükleme:**
```bash
vortexui restore /path/to/backup.tar.gz
```

### Atanmış ve tüketilen kota modu arasındaki fark nedir?

- **Atanmış:** Kullanıcılara veri limiti atadığınızda havuz azalır (tüm kullanıcı limitlerinin toplamı sayılır)
- **Tüketilen:** Yalnızca kullanıcılar gerçekten trafik kullandığında havuz azalır

Önceden satılmış paketler için atanmış kullanın. Kullanım başına ödeme için tüketilen kullanın.

### Bayi bazlı ödemeleri nasıl yapılandırırım?

1. Uygun rolle bir bayi yöneticisi oluşturun
2. Bayi giriş yapar → **Bayi Hesabı → Ödeme Yapılandırması**
3. Kendi kart numarasını, kripto adreslerini veya ZarinPal satıcısını ayarlar
4. Kullanıcıları mağazada bu ödeme seçeneklerini görür

### Düğümüm neden "Sağlıksız" gösteriyor?

Düğüm sağlık kontrollerini geçemiyor. Yaygın nedenler:

- Yüksek CPU (>%90) veya RAM (>%90)
- Paket kaybı >%10
- Çekirdek süreci çöktü (otomatik yeniden başlatma bunu yönetmelidir)
- Sertifika sorunu

Kontrol: Belirli başarısızlık nedenleri için **Düğümler → düğüm → Sağlık**.

### VortexUI ile Cloudflare'ı nasıl kullanırım?

1. Alan adını Cloudflare üzerinden sunucunuza yönlendirin (turuncu bulut)
2. Cloudflare SSL modunu **Full (Strict)** olarak ayarlayın
3. WebSocket taşıması kullanın (Cloudflare proxy'leme için gerekli)
4. CDN alan adını tanıtmak için abonelik sunucularını yapılandırın
5. Kullanıcılar Cloudflare'a bağlanır → Cloudflare düğümünüze yönlendirir

### Self-servis mağazayı nasıl etkinleştiririm?

1. Planlar oluşturun (**Planlar → Yeni Plan**)
2. Ödeme yöntemlerini yapılandırın (**Ayarlar → Ödeme Yapılandırması**)
3. Portal bağlantısını kullanıcılarla paylaşın: `/portal/login`
4. Kullanıcılar mağazaya `/sub/{token}/shop` üzerinden de erişebilir

### Bir bayi kotası bittiğinde ne olur?

- **Kullanıcı kredileri tükendi:** Yeni kullanıcı oluşturulamaz
- **Trafik kredileri tükendi:** Mevcut kullanıcılar bireysel olarak sınırlanana kadar devam eder (tüketilen mod) veya daha fazla veri atanamaz (atanmış mod)
- Otomatik askıya alma, bayiyi tamamen devre dışı bırakacak şekilde yapılandırılabilir
