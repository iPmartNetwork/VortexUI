<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/05-user-management.md) · [EN](../en/05-user-management.md) · [AR](../ar/05-user-management.md)

</div>

<div>

# 5. Kullanıcı Yönetimi

[← Gösterge paneli](./04-dashboard.md) · [Dizin](./README.md) · [Sonraki: Node'lar →](./06-node-management.md)

> [!WARNING]
> **Revoke Sub** eski bağlantıyı geçersiz kılar — yalnızca token sızdıysa kullanın.

<div align="center">

| Light | Dark |
|:-----:|:----:|
| ![Users sayfası — kullanıcı yönetimi ve abonelik](../assets/panel/User_light.png) | ![Users sayfası — kullanıcı yönetimi ve abonelik](../assets/panel/User_dark.png) |

*Users sayfası — kullanıcı yönetimi ve abonelik*

</div>

---

## Felsefe: Tek Kullanıcı, Birden Fazla Protokol

Her **User** tek bir kayıttır. Tek bir **subscription token** ile izin verilen tüm inbound'lara erişir — VLESS ve VMess için ayrı girişler oluşturmaya gerek yoktur.

---

## Kullanıcı Oluşturma

**Users → New User**

| Alan | Açıklama |
|-------|-------------|
| **Username** | Benzersiz tanımlayıcı |
| **Data limit** | Trafik limiti (bayt) — 0 = sınırsız |
| **Expire at** | Bitiş tarihi |
| **Device limit** | Maksimum eşzamanlı cihaz (IP/HWID) |
| **Reset strategy** | `none` / `daily` / `weekly` / `monthly` |
| **Status** | `active` / `disabled` / `limited` |
| **Inbounds** | İzin verilen inbound'lar |
| **Note** | Admin notu |

---

## Toplu İşlemler

| Eylem | Yol |
|--------|------|
| **Bulk Create** | Users → Add Bulk — CSV/count |
| **Multi-select** | Birden fazla kullanıcı seç → eylem |
| **Import** | Users → Import — 3x-ui / Marzban |

---

## Abonelik

### Bağlantılar

| Endpoint | Çıktı |
|----------|--------|
| `GET /sub/{token}` | base64 (varsayılan) |
| `GET /sub/{token}?format=clash` | Clash YAML |
| `GET /sub/{token}?format=singbox` | sing-box JSON |
| `GET /sub/info/{token}` | Kullanıcı HTML sayfası |

### İptal

**Users → Revoke Sub** — yeni token verilir; önceki bağlantı geçersiz olur.

---

## Trafik Muhasebesi

- **Delta push** yöntemi: çekirdek deltaları push eder (polling değil)
- **Yeniden başlatmaya dayanıklı**: sayaçlar DB'de saklanır
- **Reset**: manuel veya zamanlanmış (aylık vb.)
- **Kota uygulama**: limit aşımı → `limited` durumu + `user.limited` olayı

---

## Cihaz Limiti ve HWID

| Mekanizma | Açıklama |
|-----------|-------------|
| **Device limit** | Eşzamanlı IP/ayrı cihaz sayısı |
| **HWID allowlist** | Yalnızca kayıtlı cihazlar |
| **Online IP guard** | Çevrimiçi IP'yi limit ile karşılaştır — [Bölüm 8](./08-security-administration.md) |

---

## Kullanıcı Detay Sayfası

**Users → kullanıcı adına tıkla** → `/users/:id`

- Kullanım grafiği
- Çevrimiçi IP'ler
- Sıfırlama geçmişi
- Satır içi düzenleme

---

## En İyi Node'u Otomatik Seçme

Aboneliklerde url-test etkinleştirilerek en düşük gecikmeli aktif node otomatik seçilebilir (yapılandırma şablonunda).

---

## Kullanıcı Bildirimleri

- **Telegram kullanıcı botu**: kullanıcı token ile giriş yapar — `/usage`, `/renew`
- **Süre dolumu uyarısı**: 3 gün önce — `user.expiry_warning` olayı

</div>
