# 18. VortexUI v1.2.5 — Bayi Platformu Rehberi

!!! info "Sürüm 1.2.5"
    Bu sayfa VortexUI v1.2.5 ile tanıtılan bayi platformunu belgeler: kapsamlı izin
    listeleri, kota uygulaması, cüzdan hiyerarşisi, marka özelleştirme, otomasyon
    webhook'ları, politika limitleri, otomatik askıya alma ve tam arayüz çevirileri.

---

## Genel bakış

v1.2.5, temel bayi rolünü **tam bir alt panele** dönüştürür:

| Kitle | Ne alır |
|-------|---------|
| **Sudo yönetici** | Kota kullanım tablosu, hızlı ayar, kimliğe bürünme, cüzdan yükleme |
| **Bayi** | Pano, kullanıcılar, cüzdan, alt bayiler, marka, webhook |
| **Son kullanıcı** | Değişmedi — abonelik linkleri ve portal |

**Kenar çubuğu → Bayi:** Bayi paneli · Bayi hesabı · Kota uyarıları · Kotam (panele yönlendirir).

---

## İzin listeleri (plan, node, inbound)

**Konum:** Yöneticiler → Düzenle (sudo olmayan)

| Seçici | Etki |
|--------|------|
| **Planlar** | Yalnızca listelenen planlar satılabilir |
| **Node'lar** | Kullanıcılar yalnızca bu node'lara bağlanır |
| **Inbound'lar** | Yalnızca bu inbound'lar atanabilir |

Boş liste = kısıtlama yok.

---

## Trafik kota modları

| Mod | Havuzdan düşer… |
|-----|-----------------|
| **Allocated** | Kullanıcılara data limit atadığınızda |
| **Consumed** | Kullanıcılar gerçekten trafik tükettiğinde |

---

## Bayi paneli

**Konum:** `/reseller-dashboard` — özet kartlar, trafik, duruma göre kullanıcılar, en çok tüketenler, **CSV dışa aktar**.

---

## Kota uyarıları (sudo)

**Konum:** `/reseller-quota-alerts` — Telegram, webhook, eşik yüzdeleri, cooldown.

---

## Bayi hesabı

**Konum:** `/reseller-account`

- **Cüzdan** ve defter
- **Alt bayiler** oluşturma
- **Portal markası** (başlık, logo, renk, slug, alt bilgi)
- **Giden webhook** — `user.created` / `user.deleted` (HMAC-SHA256)

---

## Sudo araçları

**Konum:** `/admins` — kota tablosu, +50 hesap / +10 / +50 GB, Olarak giriş, Askıyı kaldır, düzenleme.

- **Kapsamlı denetim:** bayi yalnızca kendi kayıtlarını görür
- **Politika limitleri:** maks. veri/süre, toplu oluşturma/silme
- **Otomatik askıya alma:** IP ihlali veya kota aşımı

---

## Overview entegrasyonu

Bayiler için **Kotanız** kartı ve panele bağlantı.

---

## Uluslararasılaştırma

8 dil: EN, FA, TR, AR, RU, ZH, JA, ES.

---

## Yükseltme

1. v1.2.5'e geçin ve paneli yeniden başlatın.
2. Migrasyonlar: `0021`, `0022`, `0023`
3. Bayi rolü ve izin listelerini yapılandırın.

[← İçindekiler](./README.md) · [v1.2.3 Özellikleri](17-v123-features.md)
