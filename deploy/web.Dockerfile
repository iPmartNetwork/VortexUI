# Builds the VortexUI web SPA and serves it with nginx, reverse-proxying the API
# and subscription endpoints to the panel. Build from the repo root:
#   docker build -f deploy/web.Dockerfile -t vortexui/web .

# ---- build the Vite bundle ----
FROM node:22-alpine AS build
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ---- serve with nginx ----
FROM nginx:1.27-alpine
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /web/dist /usr/share/nginx/html
EXPOSE 80
