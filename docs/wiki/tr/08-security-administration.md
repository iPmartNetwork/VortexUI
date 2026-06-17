<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/08-security-administration.md) · [EN](../en/08-security-administration.md) · [AR](../ar/08-security-administration.md)

</div>

<div>

# 8. Güvenlik ve Yönetim

[← Ağ politikası](./07-network-policy.md) · [Dizin](./README.md) · [Sonraki: Planlar →](./09-plans-payments.md)

> [!IMPORTANT]
> Ana admin için **2FA** ve güçlü JWT secret (≥32 bayt) etkinleştirin.

---

## Kimlik Doğrulama

| Katman | Mekanizma |
|-------|-----------|
| Panel girişi | JWT (Bearer) — varsayılan TTL 1h |
| 2FA | TOTP (Google Authenticator vb.) |
| API otomasyonu | Personal Access Token (PAT) |
| Panel ↔ Node | mTLS (karşılıklı sertifikalar) |

### 2FA Etkinleştirme

**Settings → Two-Factor Authentication**

1. Kurulumu başlat → QR tara
2. 6 haneli kodu gir → Confirm
3. Devre dışı bırakmak için: mevcut kod + Disable

---

## RBAC (Rol Tabanlı Erişim Kontrolü)

**Admins → Roles**

| Kavram | Açıklama |
|---------|-------------|
| **Role** | İzin kümesi |
| **Permission** | `users.read`, `nodes.write`, … |
| **Reseller quota** | Alt admin için kullanıcı/trafik limiti |
| **Sub-panel** | Admin yalnızca kendi kapsamındaki kullanıcıları görür |

### Bayi oluşturma

1. Sınırlı izinlerle rol oluştur
2. Yeni admin + kota
3. Bayi yalnızca kendi kullanıcılarını yönetir

---

## API Token'ları (PAT)

**Settings → API Tokens**

```bash
curl -H "Authorization: Bearer <PAT>" \
  https://panel.example.com/api/users
```

- Her token ayrı ayrı iptal edilebilir
- İzinler oluşturan admin'in rolünden devralınır

---

## Hesap Paylaşım Koruması

Arka plan döngüsü çevrimiçi IP'leri (`GetStatsOnlineIpList`) cihaz limiti ile karşılaştırır.

| Mod | Davranış |
|------|----------|
| Detection (varsayılan) | `user.ip_limit` olayı + webhook/TG |
| `VORTEX_SHARE_AUTOLIMIT=true` | Kullanıcıyı otomatik sınırla (geri alınabilir) |

---

## IP Guard (Whitelist/Blacklist)

**Settings → IP Guard**

- IP'ye göre API/abonelik erişimini kısıtla
- Panel erişimini admin IP'lerine sınırlamak için kullanışlı

---

## Brute-Force Koruması

- Başarısız giriş denemelerinde limit
- Geçici kilitleme

---

## Denetim Günlüğü

**Audit** — tüm admin değişikliklerini kaydeder:

| Alan | Örnek |
|-------|---------|
| Actor | admin kullanıcı adı |
| Action | `user.create`, `inbound.update` |
| Target | user/node id |
| Timestamp | ISO8601 |
| Diff | önce/sonra |

---

## Inbound Başına Bant Genişliği Limiti

Inbound'ta hız limiti — tek bir servisin bağlantıyı doyurmasını önler.

---

## Inbound Başına Geo-Blocking

Belirli bir inbound'a bağlanmak için ülke/bölge kısıtlaması.

---

## Güvenlik Kontrol Listesi

- [ ] Güçlü JWT secret (≥32 bayt rastgele)
- [ ] HTTPS etkin (Let's Encrypt)
- [ ] Sudo admin için 2FA
- [ ] En az ayrıcalıkla PAT
- [ ] Şifreli off-site yedekleme
- [ ] HMAC için webhook secret
- [ ] Panel portu genel internetten kapalı (yalnızca Caddy 443)

</div>
