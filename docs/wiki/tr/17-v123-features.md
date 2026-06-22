# 17. VortexUI v1.2.3 — Yeni Özellikler Rehberi

!!! info "Version 1.2.3"
    Bu sayfa VortexUI v1.2.3 ile tanıtılan tüm özellikleri belgelemektedir. Her bölüm özelliğin ne yaptığını, arayüzde nerede bulunduğunu ve nasıl yapılandırılacağını açıklar.

---

## Abonelik Sunucuları (Subscription Hosts)

**Konum:** Ağ ve Düğümler → Abonelik Sunucuları (her inbound'un **Hosts** sekmesinden de erişilebilir)

Marzban tarzında, inbound başına sunucu geçersiz kılmaları. Bir abonelik sunucusu, kullanıcının aboneliğinde oluşturulan bağlantının ne duyurduğunu değiştirir — adres, SNI, Host başlığı, yol, aktarım güvenliği ve daha fazlası — düğümdeki canlı çekirdek yapılandırmasına **dokunmadan**. Bu, tek bir inbound'u birden fazla CDN alan adı, SNI veya portun arkasına yerleştirmenizi sağlar.

### Geçersiz kılabilecekleriniz

| Field | Açıklama |
|-------|-------------|
| Remark | Girdinin görünen adı (şablon değişkenlerini destekler) |
| Address | İstemcilere duyurulan sunucu/IP (örn. bir CDN alan adı) |
| Port | Portu geçersiz kıl (`0` = inbound'un portunu devral) |
| SNI | TLS sunucu adı |
| Host | HTTP `Host` başlığı |
| Path | WS/HTTPUpgrade/gRPC yolu |
| ALPN | Virgülle ayrılmış liste (örn. `h2,http/1.1`) |
| Fingerprint | uTLS parmak izi (örn. `chrome`) |
| Security | `inbound_default`, `none`, `tls` veya `reality` |
| Allow insecure | Sertifika doğrulamasını atla |
| Mux | Çoğullamayı etkinleştir |
| Fragment | TLS fragment spesifikasyonu (örn. `1,40-60,30-50`) |
| Priority | Inbound içinde artan sıralama |
| Enabled | Sunucuyu aç/kapat |

### Şablon değişkenleri

Dize alanları kullanıcı başına oluşturulur, böylece tek bir sunucu tanımı herkes için çalışır:

| Variable | Genişler |
|----------|-----------|
| `{USERNAME}` | Kullanıcının kullanıcı adı |
| `{SERVER_IP}` | Düğümün adresi |
| `{SERVER_PORT}` | Duyurulan port |
| `{PROTOCOL}` | Inbound protokolü (vless, vmess, …) |
| `{NETWORK}` | Aktarım (tcp, ws, grpc, …) |
| `{SECURITY}` | Aktarım güvenliği (tls, reality, none) |
| `{REMARK}` | Yapılandırılmış remark |

### Yapılandırma

1. Bir inbound açın ve **Hosts** sekmesine geçin (veya **Abonelik Sunucuları**'na gidin)
2. **Add Host** düğmesine tıklayın
3. İhtiyacınız olan geçersiz kılma alanlarını doldurun — inbound'dan devralmak için diğerlerini boş bırakın
4. Öncelik belirlemek için satırları sürükleyin (veya **Reorder** kullanın) — düşük öncelik önce oluşturulur
5. Kaydedin. Bir sonraki abonelik çekme işlemi yeni sunucuları hemen yansıtır.

!!! tip
    Yalnızca adresi/SNI'yi değiştirmek ancak inbound'un zaten müzakere ettiği şeyi korumak istediğinizde **Security**'yi `inbound_default` olarak ayarlayın.

---

## Yeni Abonelik Çıktı Formatları

**Konum:** Kullanıcılar → kullanıcı detayı → **Subscription** (format başına bağlantılar) ve genel `/sub/{token}` uç noktası

Mevcut `base64`, `clash` ve `singbox` çıktılarına ek olarak v1.2.3, `?format=` sorgu parametresiyle seçilebilen üç format ekler:

| Format | `?format=` | Çıktı |
|--------|-----------|--------|
| Xray / V2Ray JSON | `xray` | Ham Xray/V2Ray istemci JSON'u |
| Outline | `outline` | Outline için `ss://` Shadowsocks bağlantıları |
| Düz bağlantılar | `links` | V2rayN tarzı paylaşım bağlantıları, satır başına bir tane |

Tam set artık şudur: `base64`, `clash`, `singbox`, `xray`, `outline`, `links`. Bir format verilmediğinde, yanıt yine istemcinin User-Agent'ından otomatik olarak algılanır.

```text
https://panel.example.com/sub/<token>?format=xray
https://panel.example.com/sub/<token>?format=outline
https://panel.example.com/sub/<token>?format=links
```

---

## Akıllı Yönlendirme Kural Paketleri

**Konum:** Ağ ve Düğümler → Yönlendirme Paketleri (Routing Packs)

Bir **yönlendirme paketi**, yeniden kullanılabilir, adlandırılmış bir yönlendirme kuralları koleksiyonudur. Bir kez oluşturun, ardından herhangi bir düğüme uygulayın veya Clash/sing-box aboneliklerine gömün — aynı reklam engelleme / İran-doğrudan / yayın kurallarını düğüm başına yeniden oluşturmaya artık gerek yok.

### Yapabilecekleriniz

| Action | Açıklama |
|--------|-------------|
| Oluştur / düzenle | Sıralı yönlendirme kurallarından bir paket oluştur (düğüm yönlendirmesiyle aynı alanlar) |
| Düğüme uygula | Bir düğümün yönlendirme kurallarını paketle değiştir ve çekirdeğini yeniden senkronize et |
| Genel varsayılan ayarla | Geçersiz kılınmadıkça tek bir paket tüm filo genelinde uygulanır |
| Kullanıcı başına ata | Belirli bir kullanıcıya kendi paketini ver (aboneliğine gömülür) |

### Yapılandırma

1. **New Pack** düğmesine tıklayın ve adlandırın (örn. "İran — reklam engelle + doğrudan")
2. Kuralları öncelik sırasına göre ekleyin (alan adları, IP'ler, portlar, inbound etiketleri → outbound/balancer)
3. Kaydedin, ardından:
   - Paketi canlıya almak için bir düğüme **Apply** edin, veya
   - tüm filo için **Default** olarak işaretleyin, veya
   - **Kullanıcılar → kullanıcı → Routing Pack** üzerinden bir kullanıcıya atayın

!!! note
    Kullanıcı başına atama, genel varsayılana göre önceliklidir. Ataması olmayan bir kullanıcı varsayılan pakete geri döner.

---

## Temiz-IP Tarayıcısı (Clean-IP)

**Konum:** Ağ ve Düğümler → Clean IP

Adayları tarayarak ve gecikme ile paket kaybına göre puanlayarak iyi performans gösteren Cloudflare/CDN uç IP'lerini bulun. En iyilerini bir Abonelik Sunucusunda veya CDN zincirinde duyurulan adres olarak kullanın.

### Sonuçlar

| Field | Açıklama |
|-------|-------------|
| IP | Aday adres |
| Latency (ms) | Gidiş-dönüş gecikmesi |
| Loss % | Ölçülen paket kaybı |
| Score | Birleşik sıralama (yüksek olan daha iyidir) |
| Reachable | Sondanın bağlanıp bağlanmadığı |
| Scanned at | Ölçümün zaman damgası |

### Yapılandırma

1. Aday IP'leri tarama kutusuna yapıştırın (satır başına bir tane)
2. İsteğe bağlı olarak bir **port** ayarlayın (varsayılan `443`)
3. **Scan** düğmesine tıklayın — sonuçlar puanlanır ve en iyiden en kötüye sıralanır
4. En iyi IP'leri bir Abonelik Sunucusu adresine veya bir CDN/Relay zincirine kopyalayın

!!! warning "SSRF koruması"
    Tarama hedefleri sondalanmadan önce doğrulanır. Özel, loopback ve link-local aralıkları reddedilir, böylece tarayıcı dahili hizmetlere erişmek için kullanılamaz.

---

## IP-Limit Uygulaması

**Konum:** Güvenlik → IP Limit

Hesap paylaşımını engellemek için kullanıcı başına eşzamanlı IP/cihaz üst sınırlarını uygulayın. Bir kullanıcı limitini aştığında, yapılandırılan eylem tetiklenir.

### Politika

| Setting | Açıklama |
|---------|-------------|
| Enabled | Uygulamayı aç/kapat |
| Action | `warn`, `disable_temporarily` veya `kill_connections` |
| Alert cooldown | Aynı kullanıcı için tekrarlanan uyarılar arasındaki saniye |
| Restore after | Geçici olarak devre dışı bırakılan bir kullanıcının geri yüklenmesinden önceki saniye |

### Eylemler

| Action | Davranış |
|--------|----------|
| **warn** | Bir olay kaydet ve uyar; bağlantıya yönelik eylem alma |
| **disable_temporarily** | Kullanıcıyı `restore_after` saniye boyunca devre dışı bırak, ardından geri yükle |
| **kill_connections** | İhlal eden bağlantıları hemen düşür |

!!! note "Çekirdek farkları"
    `kill_connections` yalnızca **Xray**'e özgüdür. sing-box düğümlerinde otomatik olarak `disable_temporarily`'ye düşer, çünkü sing-box tek tek canlı bağlantıları düşüremez.

### Olaylar

**Events** tablosu her uygulama eylemini listeler: kullanıcı, gözlemlenen IP sayısı, IP'ler, alınan eylem ve zaman damgası.

---

## Yeni Protokoller, Aktarımlar ve Yetenek Matrisi

**Konum:** Inbounds → Ekle/Düzenle (protokol, aktarım ve güvenlik seçicileri)

v1.2.3, protokol ve aktarım kapsamını genişletir ve inbound düzenleyici artık seçimlerinizi, seçili düğümün çekirdeğinin gerçekten desteklediği şeylerle sınırlar.

### Xray-core

- **Gelen protokoller:** `vless`, `vmess`, `trojan`, `shadowsocks` (+ SS-2022 çoklu kullanıcı), `socks`, `http`, `dokodemo`
- **Aktarımlar:** `tcp`, `ws`, `grpc`, `httpupgrade`, `http`/`h2`, `xhttp`, `mkcp` (mKCP)
- **Güvenlik:** `none`, `tls`, `reality` — TLS bir `alpn` listesi alır; TCP, `none`/`http` başlık türünü destekler; xHTTP bir `mode` seçici destekler

### sing-box

- **Gelen protokoller:** `vless`, `vmess`, `trojan`, `shadowsocks`, `hysteria2`, `tuic`, `wireguard`, `hysteria` (v1), `shadowtls`, `anytls`, `socks`, `http`, `naive`
- **Aktarımlar:** `tcp`, `ws`, `grpc`, `httpupgrade`, `http`/`h2`, `quic`

### Protokol başına yetenek matrisi

Panel, canlı bir **protokol başına yetenek matrisi** sunar (`GET /api/capabilities`) ve bu, **tek doğruluk kaynağıdır** — inbound düzenleyici yalnızca seçilen çekirdeğin desteklediği protokol/aktarım/güvenlik kombinasyonlarını sunar.

Birkaç protokol **akış aktarımı taşımaz**:

| Protokol | Çekirdek | Aktarım | Güvenlik |
|----------|------|-----------|----------|
| `socks` | her ikisi | yok (ham TCP) | plaintext |
| `http` | her ikisi | yok (ham TCP) | plaintext |
| `naive` | sing-box | yok | **TLS zorunlu** |
| `dokodemo` | xray | yok (ham TCP/UDP) | plaintext |

!!! warning
    `socks` ve `http` inbound'ları **plaintext**'tir — bunları yalnızca güvenilir ağlarda veya yerel bir relay arkasında açın. `naive` **TLS'i zorunlu kılar**.

Tam yapılandırma örnekleri için [Protokoller](13-protocols-config.md) ve `docs/protocols.md` dosyasına bakın.
