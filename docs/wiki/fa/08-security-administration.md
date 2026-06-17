<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [EN](../en/08-security-administration.md) · [AR](../ar/08-security-administration.md) · [TR](../tr/08-security-administration.md)

</div>

<div dir="rtl">

# ۸. امنیت و مدیریت ادمین‌ها

[← سیاست شبکه](./07-network-policy.md) · [فهرست](./README.md) · [بعدی: پلن‌ها →](./09-plans-payments.md)

> [!IMPORTANT]
> برای ادمین اصلی حتماً **2FA** و رمز JWT قوی (≥32 بایت) فعال کنید.

---

## احراز هویت

| لایه | مکانیزم |
|------|---------|
| ورود پنل | JWT (Bearer) — TTL پیش‌فرض ۱h |
| 2FA | TOTP (Google Authenticator و …) |
| API automation | Personal Access Token (PAT) |
| Panel ↔ Node | mTLS (گواهی متقابل) |

### فعال‌سازی 2FA

**Settings → Two-Factor Authentication**

1. Start setup → QR scan
2. کد ۶ رقمی → Confirm
3. برای disable: کد فعلی + Disable

---

## RBAC (Role-Based Access Control)

**Admins → Roles**

| مفهوم | توضیح |
|-------|-------|
| **Role** | مجموعه permission |
| **Permission** | `users.read`, `nodes.write`, … |
| **Reseller quota** | سقف کاربر/ترافیک برای sub-admin |
| **Sub-panel** | ادمین فقط کاربران scope خود را می‌بیند |

### ساخت reseller

1. Role با permission محدود بسازید
2. Admin جدید + quota
3. reseller فقط users خود را مدیریت می‌کند

---

## API Tokens (PAT)

**Settings → API Tokens**

```bash
curl -H "Authorization: Bearer <PAT>" \
  https://panel.example.com/api/users
```

- هر token قابل revoke جداگانه
- permission از role ادمین سازنده به ارث می‌رسد

---

## Account-Sharing Guard

حلقه پس‌زمینه IPهای آنلاین (`GetStatsOnlineIpList`) را با device limit مقایسه می‌کند.

| حالت | رفتار |
|------|-------|
| تشخیص (پیش‌فرض) | رویداد `user.ip_limit` + webhook/TG |
| `VORTEX_SHARE_AUTOLIMIT=true` | محدودیت خودکار user (قابل بازگشت) |

---

## IP Guard (Whitelist/Blacklist)

**Settings → IP Guard**

- محدودیت دسترسی API/subscription بر اساس IP
- مفید برای محدود کردن پنل به IP ادmin

---

## Brute-force Protection

- محدودیت تلاش login ناموفق
- lockout موقت

---

## Audit Log

**Audit** — ثبت همه mutationهای admin:

| فیلد | مثال |
|------|------|
| Actor | admin username |
| Action | `user.create`, `inbound.update` |
| Target | user/node id |
| Timestamp | ISO8601 |
| Diff | before/after |

---

## Bandwidth Limit per Inbound

سقف سرعت روی inbound — جلوگیری از اشباع لینک توسط یک سرویس.

---

## Geo-blocking per Inbound

محدودیت کشور/منطقه برای اتصال به inbound خاص.

---

## Security Checklist

- [ ] JWT secret قوی (≥32 byte random)
- [ ] HTTPS فعال (Let's Encrypt)
- [ ] 2FA برای sudo admin
- [ ] PAT با least privilege
- [ ] Backup رمزنگاری‌شده off-site
- [ ] Webhook secret برای HMAC
- [ ] پورت panel از اینترنت عمومی بسته (فقط Caddy 443)

</div>
