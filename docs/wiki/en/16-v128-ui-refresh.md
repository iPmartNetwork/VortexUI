# Veltrix UI (v1.2.8)

!!! info "Frontend-only release"
    v1.2.8 is a **UI and i18n release**. No database migrations or API changes are
    required — rebuild or redeploy the panel frontend to pick it up.

---

## What's new

| Area | Highlights |
|------|------------|
| **Design system** | Glass cards, stat tiles, status badges, cyan/sky palette, page-enter animations |
| **App shell** | Collapsible sidebar, sticky header, mini mode, mobile drawer |
| **Command palette** | **Ctrl+K** / **⌘K** fuzzy navigation to any admin page |
| **Core pages** | Overview, Users, and Nodes rebuilt with live API stat cards |
| **User portal** | Redesigned login, dashboard, desktop sidebar, mobile bottom nav |
| **i18n** | **639 keys** in EN / FA / TR / AR / RU / ZH / JA / ES |

---

## Navigation

### Sidebar

The left sidebar groups pages by section (Overview, Users, Nodes, Network, Security,
Reseller, etc.). Click the **chevron** at the bottom to collapse to icon-only **mini
mode** — hover an icon to see its label.

On mobile, open the menu with the **hamburger** button in the header.

### Command palette

Press **Ctrl+K** (Windows/Linux) or **⌘K** (macOS), or click the search field in the
header. Type a page name and press Enter to jump there instantly.

### Language switcher

Click the **globe** icon in the header or login page. All eight languages share the
same key set — billing, reseller payment, pending orders, shell labels, and portal
strings are fully translated.

### User portal shortcut

Use **User Portal** in the sidebar or header to open `/portal/login` for end-user
self-service (usage, plans, support tickets).

---

## Theming

Toggle **dark / light** from the header sun/moon button. Theme preference is stored
locally and applies to both admin panel and portal login.

---

## For developers

Translation supplements live in `web/src/i18n/locale/*.json`. After editing:

```bash
cd web
node scripts/apply-i18n-locales.mjs   # merge into dict.ts
node scripts/check-i18n.mjs           # verify 0 missing keys per language
npm run build
```

See [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md)
for the full v1.2.8 release notes.
