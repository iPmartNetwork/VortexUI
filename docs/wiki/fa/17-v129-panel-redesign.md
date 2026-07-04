# Command Tower UI (نسخه ۱.۲.۹)

<div style="text-align: center; margin: 1rem 0;">
  <strong style="font-size: 1.25rem;">VortexUI نسخه ۱.۲.۹</strong><br/>
  <em style="font-size: 1rem;">ادغام صفحات، مرکز تنظیمات، پروفایل ریسلر و تله‌متری ناوگان</em>
</div>

!!! info "انتشار بک‌اند + فرانت‌اند"
    قبل از deploy باینری پنل، migration **`0030_node_location_ping.sql`** را اجرا کنید.
    فرانت‌اند را rebuild کنید (`npm run build`) و سرویس پنل را restart کنید.

---

## خلاصه

<div class="grid cards" markdown>

- :material-view-dashboard: **Overview پیشرفته**

    ویجت‌های زنده با بازه ترافیک، کاربران پرمصرف (با پروتکل) و geo/ping نودها.

- :material-cog: **مرکز تنظیمات**

    یک صفحه برای General، Security، Notifications، Appearance، API، Backup و Admins.

- :material-shield-account: **پروفایل ریسلر**

    کلیک روی نام → شارژ، سهمیه، مصرف، ledger و محدودیت‌های سیاست.

</div>

---

## صفحات ادغام‌شده

| مسیر قدیم | مسیر جدید |
|-----------|-----------|
| `/routing-packs`, `/balancers`, `/outbounds` | `/routing?tab=packs\|balancers\|outbounds` |
| `/reality-scanner`, `/clean-ip`, `/tls-tricks`, `/decoy-website` | `/evasion?tab=reality\|cleanip\|tls\|decoy` |
| `/plans`, `/pending-orders` | `/wallet-billing?tab=plans\|orders\|wallet` |
| `/admins` | `/settings?tab=admins` |
| `/audit` | `/overview` |

!!! tip "Command palette"
    **Ctrl+K** / **⌘K** — مقصدها برای مسیرهای جدید به‌روز شده‌اند.

---

## تنظیمات → Admins

| زیرتب | پارامتر | کاربرد |
|-------|---------|--------|
| **ادمین‌ها** | `section=list` | جدول سهمیه + لیست؛ نام‌ها لینک به پروفایل |
| **نقش‌ها** | `section=roles` | مدیریت Role و permission |
| **دسترسی ریسلر** | `section=access` | مرجع permission + ماتریس تنظیمات پنل |

### پروفایل ریسلر

مسیر: **`/settings/admins/{uuid}`**

شارژ کیف پول، نوار سهمیه، مصرف ترافیک، ledger، محدودیت سیاست، دسترسی تنظیمات پنل، و عملیات (ویرایش، شارژ، ورود به‌عنوان، رفع تعلیق).

---

## Inbounds و ناوگان

- **`/inbounds`** — مدیریت inbound جدا از جدول نودها
- **`/nodes`** — سلامت ناوگان + فیلدهای region، country code، ping (migration 0030)

---

## API جدید

| متد | مسیر | دسترسی |
|-----|------|--------|
| `GET` | `/api/admins/:id/quota` | sudo یا خود ادمین |
| `GET` | `/api/admins/:id/wallet` | sudo یا خود ادمین |

---

## ارتقا

```bash
vortexui migrate
go build -o vortexui-panel ./cmd/panel
cd web && npm run build
systemctl restart vortexui-panel
```

---

## مستندات مرتبط

| سند | لینک |
|-----|------|
| Veltrix UI (۱.۲.۸) | [16-v128-ui-refresh.md](16-v128-ui-refresh.md) |
| امنیت و ادمین | [08-security-administration.md](08-security-administration.md) |
| تنظیمات | [11-settings-backup.md](11-settings-backup.md) |
| Changelog | [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md) |
