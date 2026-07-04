# VortexUI Docs Site (Arena / Veltrix)

Public docs at **[ipmartnetwork.github.io/VortexUI](https://ipmartnetwork.github.io/VortexUI/)** are the same **Arena single-page React site** (glass UI, hero, install tabs, i18n) — not MkDocs Material.

## Source layout

| Path | Role |
|------|------|
| `arena-export/index.html` | Frozen Arena bundle (v1.2.8 baseline) |
| `arena-export/favicon-*.svg` | Favicons |
| `patch_arena.py` | Patches version + v1.2.9 copy, adds GitHub Pages `base href` |
| `site/` | Build output (generated; deployed by CI) |

## Build locally

```bash
python review/patch_arena.py
# open review/site/index.html or serve review/site/
```

## Release updates

1. Edit string replacements in `patch_arena.py` (hero badge, meta description, feature blurbs).
2. Run the script and spot-check `review/site/index.html`.
3. Push — `.github/workflows/docs.yml` builds `review/site` and deploys to GitHub Pages.

## Wiki (reference only)

Long-form panel chapters remain in `docs/wiki/` for maintainers. They are **not** the public landing site. See `DOC-TEMPLATE.md` if you extend the wiki separately.
