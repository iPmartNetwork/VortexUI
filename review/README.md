# VortexUI Docs — Review Template

Design reference for pages published at **[ipmartnetwork.github.io/VortexUI](https://ipmartnetwork.github.io/VortexUI/)**.

Wiki sources live in `docs/wiki/{en,fa,ar,tr}/`. When adding or updating documentation, copy patterns from **[DOC-TEMPLATE.md](./DOC-TEMPLATE.md)** — hero block, card grid, admonitions, tables, and Veltrix cyan accent styling via `docs/wiki/stylesheets/extra.css`.

## Quick checklist

1. Start from `DOC-TEMPLATE.md` — replace placeholders, keep structure.
2. Add the page to `mkdocs.yml` under **Panel guide** (all four locales).
3. Cross-link from related chapters (Dashboard, Settings, Security).
4. Mention migration steps if the release touches the database.
5. Link to `CHANGELOG.md` for the full release notes.

## Build locally

```bash
pip install -r docs/requirements.txt
mkdocs serve
```
