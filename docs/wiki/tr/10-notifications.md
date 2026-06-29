# Bildirimler

!!! info "Çok Kanallı"
    VortexUI webhook, Telegram ve uygulama içi bildirimleri destekler. Tüm kanallar bağımsız olarak yapılandırılabilir ve farklı olay alt kümeleri alabilir.

---

## Webhook (HMAC-SHA256)

**Ayarlar → Bildirimler → Webhook**

Panel olayları için uç noktanızda yapılandırılmış JSON yükleri alın.

### Yapılandırma

| Ayar | Açıklama |
|------|----------|
| URL | Webhook uç noktanız |
| Gizli anahtar | HMAC-SHA256 imzalama anahtarı |
| Olaylar | Hangi olayların gönderileceğini seçin |
| Etkin | Aç/kapat |

### Doğrulama

Her istek bir `X-Signature-256` başlığı içerir:

```
X-Signature-256: sha256=<hex_hmac>
```

`HMAC-SHA256(request_body, secret)` hesaplayıp karşılaştırarak doğrulayın.

### Olay Yükü Yapısı

```json
{
  "event": "user.limited",
  "timestamp": "2025-01-15T10:30:00Z",
  "data": {
    "user_id": "uuid",
    "username": "testuser",
    "data_limit": 53687091200,
    "used_traffic": 53687091200
  }
}
```

### Mevcut Olaylar

| Olay | Tetikleyici |
|------|-------------|
| `user.created` | Yeni kullanıcı oluşturuldu |
| `user.deleted` | Kullanıcı silindi |
| `user.limited` | Kullanıcı veri limitine ulaştı |
| `user.expired` | Kullanıcı abonelik süresi doldu |
| `user.expiry_warning` | Sona ermeden 3 gün önce |
| `user.enabled` | Kullanıcı yeniden etkinleştirildi |
| `user.disabled` | Kullanıcı manuel devre dışı bırakıldı |
| `node.offline` | Düğüm bağlantısı kesildi |
| `node.online` | Düğüm yeniden bağlandı |
| `node.unhealthy` | Düğüm sağlık kontrolü başarısız |
| `order.created` | Yeni satın alma siparişi |
| `order.paid` | Sipariş ödemesi onaylandı |
| `backup.completed` | Yedekleme tamamlandı |
| `admin.login` | Yönetici giriş olayı |

---

## Telegram Bot

**Ayarlar → Bildirimler → Telegram**

### Yönetici Botu

Yöneticinin Telegram sohbetine bildirimler gönderir:

| Ayar | Açıklama |
|------|----------|
| Bot tokeni | [@BotFather](https://t.me/BotFather)'dan |
| Yönetici sohbet ID'si | Kişisel veya grup sohbet ID'niz |
| Olaylar | Bildirim olaylarını seçin |

### Kullanıcıya Yönelik Bot

Kullanıcılar abonelik tokenleri ile botla etkileşime geçebilir:

| Komut | Eylem |
|-------|-------|
| `/start` | Abonelik tokeni ile hesabı bağla |
| `/usage` | Mevcut veri kullanımını görüntüle |
| `/renew` | Yenileme bağlantısını al |
| `/status` | Hesap durumunu kontrol et |
| `/help` | Mevcut komutları listele |

### Bildirim Şablonları

Şablon değişkenleri ile mesaj biçimini özelleştirin:

```
🔔 Kullanıcı Limite Ulaştı
Kullanıcı adı: {username}
Kullanılan: {used_traffic}
Limit: {data_limit}
```

---

## Kota Bildirimleri

**Ayarlar → Bildirimler → Kota Uyarıları**

Kullanıcılar veri limitlerine yaklaştığında uyarı:

| Ayar | Açıklama |
|------|----------|
| Etkin | Kota bildirimlerini etkinleştir |
| Eşikler | Tetikleme yüzdeleri (örn. %80, %90, %100) |
| Telegram | Bot üzerinden gönder |
| Webhook | Webhook URL'sine gönder |
| Bekleme süresi | Tekrarlanan uyarılar arası dakika |

---

## Bayi Kota Uyarıları

**Kenar çubuğu → Bayi Kota Uyarıları** (yalnızca sudo yönetici)

Bayiler trafik/kullanıcı havuz limitlerine yaklaştığında izleme:

| Ayar | Açıklama |
|------|----------|
| Etkin | Global anahtar |
| Telegram | Panel botuna gönder |
| Webhook URL | İsteğe bağlı harici uç nokta |
| Eşikler | Yüzdeler (örn. 80, 90, 100) |
| Bekleme süresi | Uyarılar arası dakika |
| Son uyarılar | Tetiklenen uyarılar tablosu |

---

## Bildirim Merkezi (Zil Açılır Menüsü)

Panel başlığındaki zil simgesi son bildirimleri gösterir:

- Okunmamış sayı rozeti
- Genişletmek için tıklayın
- Her bildirim gösterir: olay türü, açıklama, zaman damgası
- İlgili kaynağa gitmek için tıklayın
- Okundu olarak işaretle / tümünü okundu olarak işaretle

Yönetici hesabı başına kalıcıdır.

---

## SSE Canlı Olaylar

Panel, gerçek zamanlı arayüz güncellemeleri için Server-Sent Events kullanır:

| Akış | İçerik |
|------|--------|
| `/api/sse/events` | Sistem olayları (düğüm durumu, kullanıcı limitleri, vb.) |
| `/api/sse/stats` | Canlı istatistikler (bağlantılar, trafik sayaçları) |
| `/api/sse/monitor` | Aktif bağlantı güncellemeleri |

Ön yüz bileşenleri otomatik abone olur. Yapılandırma gerekmez — SSE her zaman aktiftir.

!!! tip
    SSE olayları gerçek zamanlı gösterge paneli göstergelerini, monitör sayfasını ve bildirim zilini besler.
    Ters proxy'niz yanıtları arabelleğe alıyorsa, SSE akışına izin verildiğinden emin olun (Caddy bunu varsayılan olarak yönetir).
