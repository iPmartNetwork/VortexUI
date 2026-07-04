# VortexUI Documentation

## Documentation site (Arena / Veltrix UI)

**Live site:** [https://ipmartnetwork.github.io/VortexUI/](https://ipmartnetwork.github.io/VortexUI/)

Single-page React site (same design as the Arena mock): hero, install tabs, protocols, architecture, 8-language i18n, glass UI.

Built from `review/arena-export/` via `review/patch_arena.py` — see [`review/README.md`](../review/README.md).

### Local preview

```bash
python review/patch_arena.py
# serve review/site/ with any static server
```

### Wiki (MkDocs — reference / long-form)

Panel chapters in `docs/wiki/{en,fa,ar,tr}/` can still be built locally:

```bash
pip install -r docs/requirements.txt
mkdocs serve
```

### Deploy

Push to `master` — GitHub Actions workflow [`.github/workflows/docs.yml`](../.github/workflows/docs.yml) builds and deploys to **GitHub Pages**.

**One-time setup:** Repository **Settings → Pages → Build and deployment → Source: GitHub Actions**.

---

## Wiki (Markdown source)

Markdown files live in [`wiki/`](WIKI-HUB.md) (FA / EN / AR / TR).

**[📚 WIKI-HUB.md](WIKI-HUB.md)** · [فارسی](wiki/fa/README.md) · [English](wiki/en/README.md) · [العربية](wiki/ar/README.md) · [Türkçe](wiki/tr/README.md)

Edit wiki Markdown → CI rebuilds the site automatically.

---

## API Reference

[`openapi.yaml`](openapi.yaml) is the full OpenAPI 3.0 specification for the panel
REST API — every endpoint, request/response shape, auth requirement, and RBAC
permission.

View or use it:

```bash
# Interactive docs (any of these)
npx @redocly/cli preview-docs docs/openapi.yaml
docker run -p 8081:8080 -e SWAGGER_JSON=/spec/openapi.yaml -v "$PWD/docs:/spec" swaggerapi/swagger-ui

# Generate a typed client (e.g. for the frontend)
npx openapi-typescript docs/openapi.yaml -o web/src/api/types.ts
```

Auth: obtain a JWT from `POST /api/login`, then send `Authorization: Bearer <token>`.
The public `GET /sub/{token}` endpoint needs no JWT — the token is the credential.
