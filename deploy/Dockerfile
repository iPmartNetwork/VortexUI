# Multi-stage build for both VortexUI binaries. Build a specific image with:
#   docker build -f deploy/Dockerfile --target panel -t vortexui/panel .
#   docker build -f deploy/Dockerfile --target node  -t vortexui/node  .
# The binaries are pure Go (CGO disabled), so the panel ships on distroless/static.

# ---- builder ----
FROM golang:1.26-alpine AS builder
WORKDIR /src
# Cache modules first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=1.0.0
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
        -ldflags "-s -w -X main.version=${VERSION}" \
        -o /out/panel ./cmd/panel && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath \
        -ldflags "-s -w -X main.version=${VERSION}" \
        -o /out/node ./cmd/node

# ---- xray fetch (for the node image) ----
FROM alpine:3.20 AS xray
ARG XRAY_VERSION=v1.8.24
ARG TARGETARCH=64
RUN apk add --no-cache curl unzip && \
    curl -fsSL -o /tmp/xray.zip \
      "https://github.com/XTLS/Xray-core/releases/download/${XRAY_VERSION}/Xray-linux-${TARGETARCH}.zip" && \
    unzip /tmp/xray.zip -d /xray && chmod +x /xray/xray

# ---- sing-box fetch (for the node image) ----
FROM alpine:3.20 AS singbox
ARG SINGBOX_VERSION=1.9.3
ARG SB_ARCH=amd64
RUN apk add --no-cache curl tar && \
    curl -fsSL -o /tmp/sb.tgz \
      "https://github.com/SagerNet/sing-box/releases/download/v${SINGBOX_VERSION}/sing-box-${SINGBOX_VERSION}-linux-${SB_ARCH}.tar.gz" && \
    tar -xzf /tmp/sb.tgz -C /tmp && \
    mv /tmp/sing-box-*/sing-box /usr/local/bin/sing-box && chmod +x /usr/local/bin/sing-box

# ---- panel (control plane) ----
FROM gcr.io/distroless/static-debian12 AS panel
COPY --from=builder /out/panel /usr/local/bin/panel
COPY --from=builder /src/migrations /migrations
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/panel"]

# ---- node (agent + xray-core) ----
FROM alpine:3.20 AS node
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 vortex
COPY --from=builder /out/node /usr/local/bin/node
COPY --from=xray /xray/xray /usr/local/bin/xray
COPY --from=singbox /usr/local/bin/sing-box /usr/local/bin/sing-box
# Geo databases live in a writable dir so the panel can refresh them at runtime.
COPY --from=xray /xray/*.dat /etc/vortex/assets/
RUN chown -R vortex:vortex /etc/vortex
# Defaults to xray; set VORTEX_CORE=singbox + VORTEX_CORE_BIN=/usr/local/bin/sing-box
# on a node to run sing-box instead.
ENV VORTEX_CORE=xray \
    VORTEX_CORE_BIN=/usr/local/bin/xray \
    VORTEX_CORE_CONFIG=/etc/vortex/core.json \
    XRAY_LOCATION_ASSET=/etc/vortex/assets
EXPOSE 50051
USER vortex
ENTRYPOINT ["/usr/local/bin/node"]
