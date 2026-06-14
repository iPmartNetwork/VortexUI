# VortexUI API docs

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
