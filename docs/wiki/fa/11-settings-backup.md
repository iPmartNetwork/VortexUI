# ۱۱. تنظیمات و پشتیبان‌گیری

!!! tip
    قبل از **Restore** حتماً backup فعلی بگیرید.

---

## Appearance

**Settings → Appearance**

| گزینه | مقادیر |
|-------|--------|
| Theme | Light / Dark / System |
| Language | EN, FA, TR, AR, RU, ZH, JA, ES |

---

## Change Password

رمز فعلی + رمز جدید — JWT session فعلی حفظ می‌شود.

---

## API Tokens

ایجاد، کپی یک‌بار، لیست، revoke — [فصل ۸](./08-security-administration.md)

---

## Backup & Restore

### Export

**Settings → Backup → Download**

- snapshot transactional کل DB (users, nodes, inbounds, routing, …)
- JSON format

### Restore

**Upload JSON** — merge یا replace (بسته به API)

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @backup.json \
  https://panel.example.com/api/backup/restore
```

> قبل از restore حتماً backup فعلی بگیرید.

---

## Subscription Config Template

**Settings → Config Template**

- override template Clash/sing-box
- ruleهای پیش‌فرض، dns، proxy-groups
- placeholder: `{{USER}}`, `{{NODES}}`, …

---

## Custom Branding

**Settings → Branding**

| فیلد | اثر |
|------|-----|
| Panel title | عنوان UI |
| Logo URL | لوگوی سفارشی |
| Sub page title | صفحه `/sub/info` |

---

## Auto Backup

- فاصله زمانی
- Telegram / S3
- retention policy

---

## Update Checker

**Settings → Updates**

- بررسی نسخه GitHub release
- Auto-update panel + core binaries (قابل فعال‌سازی)

---

## PWA

پنل **Progressive Web App** است — از مرورگر موبایل «Add to Home Screen» برای تجربه app-like.

`web/public/manifest.json` — نام، آیکون، theme color.

---

## Logs

**Logs** — لاگ سطح panel (نه هسته):

- فیلتر level
- جستجو
- real-time tail

برای لاگ هسته: **Nodes → Logs**
