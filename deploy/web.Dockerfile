# Builds the VortexUI web SPA and serves it with Caddy, which reverse-proxies the
# API + subscription endpoints to the panel and handles automatic HTTPS
# (Let's Encrypt) when a domain is configured. Build from the repo root:
#   docker build -f deploy/web.Dockerfile -t vortexui/web .

# ---- build the Vite bundle ----
FROM node:22-alpine AS build
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ---- serve with Caddy ----
FROM caddy:2-alpine
COPY deploy/Caddyfile /etc/caddy/Caddyfile
COPY --from=build /web/dist /usr/share/caddy
EXPOSE 80 443
