# 11. Ayarlar ve Yedekleme

!!! tip
    **Restore** öncesinde mutlaka güncel yedek alın.

---

## Görünüm

**Settings → Appearance**

| Seçenek | Değerler |
|--------|--------|
| Theme | Light / Dark / System |
| Language | EN, FA, TR, AR, RU, ZH, JA, ES |

---

## Şifre Değiştirme

Mevcut şifre + yeni şifre — mevcut JWT oturumu korunur.

---

## API Token'ları

Oluştur, tek seferlik kopyala, listele, iptal et — [Bölüm 8](./08-security-administration.md)

---

## Yedekleme ve Geri Yükleme

### Dışa Aktarma

**Settings → Backup → Download**

- Tam DB'nin transactional anlık görüntüsü (users, nodes, inbounds, routing, …)
- JSON formatı

### Geri Yükleme

**Upload JSON** — birleştir veya değiştir (API'ye bağlı)

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @backup.json \
  https://panel.example.com/api/backup/restore
```

> Geri yüklemeden önce her zaman yedek alın.

---

## Abonelik Yapılandırma Şablonu

**Settings → Config Template**

- Clash/sing-box şablonunu geçersiz kıl
- Varsayılan kurallar, DNS, proxy-groups
- Placeholder'lar: `{{USER}}`, `{{NODES}}`, …

---

## Özel Markalama

**Settings → Branding**

| Alan | Etki |
|-------|--------|
| Panel title | UI başlığı |
| Logo URL | Özel logo |
| Sub page title | `/sub/info` sayfası |

---

## Otomatik Yedekleme

- Aralık
- Telegram / S3
- Saklama politikası

---

## Güncelleme Denetleyici

**Settings → Updates**

- GitHub release sürümünü kontrol et
- Panel + çekirdek ikililerini otomatik güncelle (isteğe bağlı)

---

## PWA

Panel bir **Progressive Web App**'tir — mobil tarayıcıdan uygulama benzeri deneyim için "Add to Home Screen" kullanın.

`web/public/manifest.json` — ad, simgeler, tema rengi.

---

## Loglar

**Logs** — panel düzeyi loglar (çekirdek değil):

- Seviye filtresi
- Arama
- Gerçek zamanlı tail

Çekirdek logları için: **Nodes → Logs**
