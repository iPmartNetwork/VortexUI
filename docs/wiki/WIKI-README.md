# VortexUI Wiki

Complete documentation for VortexUI **v1.3.1**, built with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/).

**Author:** ali (zeus)

## Structure

```
docs/wiki/
├── en/                     ← English (complete — 17 pages + README)
├── fa/                     ← فارسی (core pages + full panel guide where translated)
├── ar/                     ← العربية (home + existing translated pages)
└── tr/                     ← Türkçe (home + existing translated pages)
```

MkDocs config: `mkdocs.yml` at repository root (`docs_dir: docs/wiki`).

## Local development

```bash
pip install -r docs/requirements.txt
mkdocs serve
```

Open http://127.0.0.1:8000

## Build

```bash
mkdocs build
```

Output: `site/`

## Landing page

Marketing/landing site (React + Vite) lives in `docs/landing/`:

```bash
cd docs/landing
npm install
npm run dev
```

## Languages

| Language | Code | Status |
|----------|------|--------|
| English | en | Complete (17 pages) |
| فارسی | fa | Core + extended pages |
| العربية | ar | Home + partial |
| Türkçe | tr | Home + partial |

## Version

This documentation covers **VortexUI v1.3.1**.

See [Changelog](en/16-changelog.md) for version history.
