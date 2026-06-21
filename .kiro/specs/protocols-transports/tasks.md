# Implementation Plan

> All work is LOCAL — do not push. Local commits per task are allowed.

## Phase 1 — P0 correctness (stop silent-broken configs)
- [x] 1. Add transport/security fields to the inbound form so configs build usably
  - NodeInboundsModal.tsx: add conditional inputs for path, host, gRPC serviceName, flow; wire into submit + startEdit
  - hooks.ts: extend Inbound/Create/Update inputs with path, host, flow (sni already present)
  - _Requirements: 2.2, 2.3, 2.4_
- [x] 2. Fix VLESS flow emission (xray + singbox): only for vless + tcp + (tls|reality); never on ws/grpc/non-TLS
  - _Requirements: 5.3_
- [x] 3. sing-box: skip/se reject REALITY inbound with empty private key (mirror xray inboundUsable)
  - _Requirements: 5.2_
- [x] 4. xray Shadowsocks: support multi-user + ss2022 methods, or surface a clear error instead of silent single-user
  - _Requirements: 3.2_

## Phase 2 — P1 single source of truth
- [x] 5. Add core capability matrix (internal/core/capabilities.go) and derive coreSupports from it
  - _Requirements: 1.1, 1.2_
- [x] 6. Expose GET /api/capabilities and consume it in the frontend for per-core option filtering
  - _Requirements: 1.3, 2.1_
- [x] 7. Reconcile quic / xhttp / udp across guard + renderer + UI per core
  - _Requirements: 4.2, 6.1, 6.2_

## Phase 3 — P2 completeness + hardening
- [x] 8. TLS: model + render alpn and fingerprint/utls and sni on both cores
  - _Requirements: 5.1_
- [x] 9. Auto/validated xtls-rprx-vision for vless+reality/tls+tcp
  - _Requirements: 5.3_
- [x] 10. sing-box hysteria2/tuic protocol-specific fields (bandwidth, obfs, congestion)
  - _Requirements: 3.3_
- [x] 11. Add remaining protocols per matrix: ss2022 (both), sing-box shadowtls/anytls/naive/hysteria1/socks/http, xray socks/http/dokodemo/mkcp
  - (shipped: ss2022 both cores; sing-box shadowtls/anytls/hysteria1, socks/http both, naive sing-box; xray dokodemo)
  - _Requirements: 3.4, 4.1_
- [x] 12. Wire REALITY keygen output into the saved inbound (no manual paste)
  - _Requirements: 6.3_
- [x] 13. Complete transport rendering: tcp header(none/http), xhttp mode (xray), correct grpc/httpupgrade/http fields
  - (shipped: tcp header none/http + xhttp mode (xray) + mKCP (kcp) transport (xray); grpc/httpupgrade/http fields complete)
  - _Requirements: 4.1_

## Phase 4 — verify
- [x] 14. Table-driven renderer tests for every matrix combo (both cores) + guard rejection tests
  - _Requirements: 7.1, 7.2_
- [x] 15. Full verify: go build/vet/test, golangci-lint, web npm run build — all green (no push)
  - _Requirements: 7.1, 7.3_
