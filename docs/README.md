# VortexUI Documentation

## Documentation site (MkDocs Material)

**Live site:** [https://ipmartnetwork.github.io/VortexUI/](https://ipmartnetwork.github.io/VortexUI/)

Four languages: **English · فارسی · العربية · Türkçe** — search, dark/light theme, sidebar navigation, Mermaid diagrams.

Wiki pages follow the design patterns in [`review/DOC-TEMPLATE.md`](../review/DOC-TEMPLATE.md) (hero block, card grid, admonitions, Veltrix cyan styling).

### Local preview

```bash
pip install -r docs/requirements.txt
mkdocs serve
# open http://127.0.0.1:8000
```

Or:

```bash
make docs-serve
```

### Build

```bash
make docs
# output in site/
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
