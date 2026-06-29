# API Referansı

!!! info "OpenAPI Spesifikasyonu"
    Tam OpenAPI 3.0 spesifikasyonu [`docs/openapi.yaml`](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) adresinde mevcuttur.

---

## Kimlik Doğrulama

### JWT Girişi

```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "sifreniz",
  "totp_code": "123456"  // isteğe bağlı, 2FA etkinse
}
```

Yanıt:

```json
{
  "access_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

Sonraki isteklerde tokeni kullanın:

```
Authorization: Bearer <access_token>
```

### API Tokenleri (PAT)

Otomasyon için Kişisel Erişim Tokeni oluşturun:

1. **Ayarlar → API Tokenleri → Oluştur**
2. JWT gibi kullanın: `Authorization: Bearer <PAT>`

PAT'ler yapılandırılmadıkça sona ermez ve tek tek iptal edilebilir.

---

## Temel URL ve Sürümleme

```
https://panel.example.com/api/
```

Tüm uç noktalar `/api/` altındadır. Sürüm öneki yok — API ileriye uyumludur.

---

## Temel Uç Noktalar

### Kimlik Doğrulama

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| POST | `/api/auth/login` | Giriş yap, JWT al |
| POST | `/api/auth/refresh` | Token yenile |
| GET | `/api/auth/me` | Mevcut yönetici bilgisi |

### Kullanıcılar

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/users` | Kullanıcıları listele (sayfalama, filtreler) |
| POST | `/api/users` | Kullanıcı oluştur |
| GET | `/api/users/:id` | Kullanıcı detayını al |
| PUT | `/api/users/:id` | Kullanıcıyı güncelle |
| DELETE | `/api/users/:id` | Kullanıcıyı sil |
| POST | `/api/users/:id/reset-traffic` | Trafik sayacını sıfırla |
| POST | `/api/users/:id/revoke-sub` | Abonelik tokenini iptal et |
| GET | `/api/users/:id/usage` | Kullanım geçmişi |

### Düğümler

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/nodes` | Düğümleri listele |
| POST | `/api/nodes` | Düğüm oluştur |
| GET | `/api/nodes/:id` | Düğüm detayını al |
| PUT | `/api/nodes/:id` | Düğümü güncelle |
| DELETE | `/api/nodes/:id` | Düğümü sil |
| POST | `/api/nodes/:id/restart` | Düğüm çekirdeğini yeniden başlat |
| GET | `/api/nodes/:id/health` | Sağlık durumu |
| GET | `/api/nodes/:id/stats` | Canlı istatistikler |

### Gelen Bağlantılar

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/inbounds` | Gelen bağlantıları listele |
| POST | `/api/inbounds` | Gelen bağlantı oluştur |
| PUT | `/api/inbounds/:id` | Gelen bağlantıyı güncelle |
| DELETE | `/api/inbounds/:id` | Gelen bağlantıyı sil |
| GET | `/api/capabilities` | Protokol başına yetenek matrisi |

### Planlar

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/plans` | Planları listele (yöneticiye göre kapsamlı) |
| POST | `/api/plans` | Plan oluştur |
| PUT | `/api/plans/:id` | Planı güncelle |
| DELETE | `/api/plans/:id` | Planı sil |

### Siparişler

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/orders` | Siparişleri listele |
| GET | `/api/orders/:id` | Sipariş detayını al |
| POST | `/api/orders/:id/approve` | Bekleyen siparişi onayla |
| POST | `/api/orders/:id/reject` | Bekleyen siparişi reddet |

### Ödeme Yapılandırması

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/payment-config` | Ödeme yapılandırmasını al |
| PUT | `/api/payment-config` | Ödeme yapılandırmasını güncelle |

### Yöneticiler ve Bayi

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/admins` | Yöneticileri listele |
| POST | `/api/admins` | Yönetici oluştur |
| PUT | `/api/admins/:id` | Yöneticiyi güncelle |
| POST | `/api/admins/:id/quota-adjust` | Bayi kotasını ayarla |
| POST | `/api/admins/:id/unsuspend` | Askıyı kaldır |
| POST | `/api/admins/:id/impersonate` | Bayi tokeni düzenle |

### Bayi Hesabı

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/account/dashboard` | Bayi gösterge paneli istatistikleri |
| GET | `/api/account/export/users` | Sahipli kullanıcıların CSV dışa aktarması |
| GET | `/api/account/wallet` | Cüzdan + defter |
| GET/PUT | `/api/account/branding` | Beyaz etiket ayarları |
| GET/PUT | `/api/account/webhook` | Giden webhook yapılandırması |
| GET/POST | `/api/account/sub-admins` | Alt bayi yönetimi |

### Sistem

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/api/health` | Sağlık kontrolü |
| GET | `/api/stats` | Sistem istatistikleri |
| GET | `/api/audit` | Denetim günlüğü |
| POST | `/api/backup` | Yedekleme tetikle |
| GET | `/api/settings` | Panel ayarları |
| PUT | `/api/settings` | Ayarları güncelle |

### Abonelikler (Genel)

| Yöntem | Uç Nokta | Açıklama |
|--------|----------|----------|
| GET | `/sub/{token}` | Kullanıcı aboneliği (otomatik biçim) |
| GET | `/sub/{token}?format=clash` | Clash YAML |
| GET | `/sub/{token}?format=singbox` | sing-box JSON |
| GET | `/sub/{token}?format=xray` | Xray JSON |
| GET | `/sub/{token}?format=outline` | Outline bağlantıları |
| GET | `/sub/{token}?format=links` | Düz paylaşım bağlantıları |
| GET | `/sub/info/{token}` | Kullanıcı bilgi HTML sayfası |
| GET | `/sub/{token}/shop` | Self-servis mağaza |

---

## Örnek İstekler

### Giriş

```bash
curl -X POST https://panel.example.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"secret"}'
```

### Kullanıcı Oluşturma

```bash
curl -X POST https://panel.example.com/api/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "data_limit": 53687091200,
    "expire_at": "2025-03-01T00:00:00Z",
    "device_limit": 3,
    "inbound_ids": ["uuid-1", "uuid-2"]
  }'
```

### Plan Listeleme

```bash
curl https://panel.example.com/api/plans \
  -H "Authorization: Bearer <token>"
```

### Satın Alma (Portal)

```bash
curl -X POST https://panel.example.com/api/portal/purchase \
  -H "Authorization: Bearer <sub_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "plan_id": "uuid",
    "payment_method": "card",
    "proof_image": "base64_encoded_image",
    "reference_number": "123456789"
  }'
```

---

## Sayfalama

Listeleme uç noktaları sayfalamayı destekler:

| Parametre | Açıklama | Varsayılan |
|-----------|----------|------------|
| `page` | Sayfa numarası (1'den başlar) | 1 |
| `per_page` | Sayfa başına öğe | 20 |
| `sort` | Sıralama alanı | `created_at` |
| `order` | `asc` veya `desc` | `desc` |

Yanıt sayfalama meta verisi içerir:

```json
{
  "data": [...],
  "total": 150,
  "page": 1,
  "per_page": 20,
  "total_pages": 8
}
```

---

## Hata Yanıtları

Tüm hatalar tutarlı bir biçim izler:

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "Belirtilen ID ile kullanıcı bulunamadı",
    "status": 404
  }
}
```

Yaygın HTTP durum kodları:

| Kod | Anlamı |
|-----|--------|
| 400 | Hatalı istek (doğrulama hatası) |
| 401 | Yetkisiz (geçersiz/süresi dolmuş token) |
| 403 | Yasak (yetersiz izinler) |
| 404 | Kaynak bulunamadı |
| 409 | Çakışma (yinelenen kullanıcı adı, vb.) |
| 429 | Hız sınırı aşıldı |
| 500 | Dahili sunucu hatası |
