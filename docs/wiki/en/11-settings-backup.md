<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/11-settings-backup.md) · [AR](../ar/11-settings-backup.md) · [TR](../tr/11-settings-backup.md)

</div>

<div>

# 11. Settings & Backup

[← Notifications](./10-notifications.md) · [Index](./README.md) · [Next: API →](./12-api-reference.md)

> [!TIP]
> Always take a current backup before **Restore**.

---

## Appearance

**Settings → Appearance**

| Option | Values |
|--------|--------|
| Theme | Light / Dark / System |
| Language | EN, FA, TR, AR, RU, ZH, JA, ES |

---

## Change Password

Current password + new password — current JWT session is preserved.

---

## API Tokens

Create, one-time copy, list, revoke — [Chapter 8](./08-security-administration.md)

---

## Backup & Restore

### Export

**Settings → Backup → Download**

- Transactional snapshot of the full DB (users, nodes, inbounds, routing, …)
- JSON format

### Restore

**Upload JSON** — merge or replace (depends on API)

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @backup.json \
  https://panel.example.com/api/backup/restore
```

> Always take a backup before restoring.

---

## Subscription Config Template

**Settings → Config Template**

- Override Clash/sing-box template
- Default rules, DNS, proxy-groups
- Placeholders: `{{USER}}`, `{{NODES}}`, …

---

## Custom Branding

**Settings → Branding**

| Field | Effect |
|-------|--------|
| Panel title | UI title |
| Logo URL | Custom logo |
| Sub page title | `/sub/info` page |

---

## Auto Backup

- Interval
- Telegram / S3
- Retention policy

---

## Update Checker

**Settings → Updates**

- Check GitHub release version
- Auto-update panel + core binaries (optional)

---

## PWA

The panel is a **Progressive Web App** — from mobile browser use "Add to Home Screen" for an app-like experience.

`web/public/manifest.json` — name, icons, theme color.

---

## Logs

**Logs** — panel-level logs (not core):

- Level filter
- Search
- Real-time tail

For core logs: **Nodes → Logs**

</div>
