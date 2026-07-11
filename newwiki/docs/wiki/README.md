# VortexUI Wiki

Complete documentation for VortexUI **v1.3.1**, built with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/).

## 📁 Structure

```
docs/wiki/
├── mkdocs.yml              ← MkDocs configuration
├── en/                     ← 🇬🇧 English (complete)
│   ├── index.md
│   ├── 01-introduction.md
│   ├── 02-installation.md
│   ├── 03-first-steps.md
│   ├── 04-dashboard.md
│   ├── 05-user-management.md
│   ├── 06-node-management.md
│   ├── 07-network-policy.md
│   ├── 08-security-administration.md
│   ├── 09-plans-payments.md
│   ├── 10-notifications.md
│   ├── 11-settings-backup.md
│   ├── 12-api-reference.md
│   ├── 13-protocols-config.md
│   ├── 14-operations-maintenance.md
│   ├── 15-troubleshooting-faq.md
│   ├── 16-changelog.md
│   └── 17-menu-usage-guide.md
├── fa/                     ← 🇮🇷 فارسی
├── ar/                     ← 🇸🇦 العربية
└── tr/                     ← 🇹🇷 Türkçe
```

## 🚀 Local Development

### Install MkDocs

```bash
pip install mkdocs-material mkdocs-static-i18n
```

### Serve Locally

```bash
cd docs/wiki
mkdocs serve
```

Open http://localhost:8000

### Build Static Site

```bash
mkdocs build
```

Output goes to `site/`.

## 📦 Deploy to GitHub Pages

### Option 1: Manual

```bash
cd docs/wiki
mkdocs gh-deploy
```

### Option 2: GitHub Actions

Create `.github/workflows/docs.yml`:

```yaml
name: Deploy Docs
on:
  push:
    branches: [master]
    paths: ['docs/wiki/**']
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.12'
      - run: pip install mkdocs-material mkdocs-static-i18n
      - run: cd docs/wiki && mkdocs gh-deploy --force
```

## 🌍 Languages

| Language | Code | Status | Direction |
|----------|------|--------|-----------|
| English | en | ✅ Complete (17 pages) | LTR |
| فارسی | fa | 🟡 Core pages | RTL |
| العربية | ar | 🟡 Index | RTL |
| Türkçe | tr | 🟡 Index | LTR |

## ✍️ Contributing

To add or update documentation:

1. Edit the relevant `.md` file
2. For new pages, add to `nav:` in `mkdocs.yml`
3. Test locally with `mkdocs serve`
4. Submit a PR

### Translation Guidelines

- Keep the same file structure across languages
- Preserve code blocks and commands in English
- Translate prose, tables headers, and descriptions
- Maintain RTL formatting for FA/AR

## 📝 Version

This documentation covers **VortexUI v1.3.1**.

See [Changelog](en/16-changelog.md) for version history.
