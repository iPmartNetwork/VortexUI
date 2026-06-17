# 8. Security & Administration

!!! important
    Enable **2FA** and a strong JWT secret (≥32 bytes) for the primary admin.

---

## Authentication

| Layer | Mechanism |
|-------|-----------|
| Panel login | JWT (Bearer) — default TTL 1h |
| 2FA | TOTP (Google Authenticator, etc.) |
| API automation | Personal Access Token (PAT) |
| Panel ↔ Node | mTLS (mutual certificates) |

### Enabling 2FA

**Settings → Two-Factor Authentication**

1. Start setup → scan QR
2. Enter 6-digit code → Confirm
3. To disable: current code + Disable

---

## RBAC (Role-Based Access Control)

**Admins → Roles**

| Concept | Description |
|---------|-------------|
| **Role** | Set of permissions |
| **Permission** | `users.read`, `nodes.write`, … |
| **Reseller quota** | User/traffic cap for sub-admin |
| **Sub-panel** | Admin sees only users in their scope |

### Creating a reseller

1. Create a role with limited permissions
2. New admin + quota
3. Reseller manages only their own users

---

## API Tokens (PAT)

**Settings → API Tokens**

```bash
curl -H "Authorization: Bearer <PAT>" \
  https://panel.example.com/api/users
```

- Each token can be revoked individually
- Permissions inherit from the creating admin's role

---

## Account-Sharing Guard

A background loop compares online IPs (`GetStatsOnlineIpList`) with the device limit.

| Mode | Behavior |
|------|----------|
| Detection (default) | `user.ip_limit` event + webhook/TG |
| `VORTEX_SHARE_AUTOLIMIT=true` | Auto-limit user (reversible) |

---

## IP Guard (Whitelist/Blacklist)

**Settings → IP Guard**

- Restrict API/subscription access by IP
- Useful for limiting panel access to admin IPs

---

## Brute-Force Protection

- Limit on failed login attempts
- Temporary lockout

---

## Audit Log

**Audit** — records all admin mutations:

| Field | Example |
|-------|---------|
| Actor | admin username |
| Action | `user.create`, `inbound.update` |
| Target | user/node id |
| Timestamp | ISO8601 |
| Diff | before/after |

---

## Bandwidth Limit per Inbound

Speed cap on inbound — prevents one service from saturating the link.

---

## Geo-Blocking per Inbound

Country/region restriction for connecting to a specific inbound.

---

## Security Checklist

- [ ] Strong JWT secret (≥32 byte random)
- [ ] HTTPS enabled (Let's Encrypt)
- [ ] 2FA for sudo admin
- [ ] PAT with least privilege
- [ ] Encrypted off-site backup
- [ ] Webhook secret for HMAC
- [ ] Panel port closed from public internet (Caddy 443 only)
