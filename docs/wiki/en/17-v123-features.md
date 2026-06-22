# 17. VortexUI v1.2.3 — New Features Guide

!!! info "Version 1.2.3"
    This page documents all features introduced in VortexUI v1.2.3. Each section explains what the feature does, where it lives in the UI, and how to configure it.

---

## Subscription Hosts

**Location:** Network & Nodes → Subscription Hosts (also reachable from each inbound's **Hosts** tab)

Per-inbound host overrides, in the style of Marzban. A subscription host changes what
the link rendered into a user's subscription advertises — address, SNI, Host header,
path, transport security, and more — **without** touching the live core config on the
node. This lets you front a single inbound behind several CDN domains, SNIs, or ports.

### What you can override

| Field | Description |
|-------|-------------|
| Remark | Display name for the entry (supports template variables) |
| Address | Host/IP advertised to clients (e.g. a CDN domain) |
| Port | Override port (`0` = inherit the inbound's port) |
| SNI | TLS server name |
| Host | HTTP `Host` header |
| Path | WS/HTTPUpgrade/gRPC path |
| ALPN | Comma-separated list (e.g. `h2,http/1.1`) |
| Fingerprint | uTLS fingerprint (e.g. `chrome`) |
| Security | `inbound_default`, `none`, `tls`, or `reality` |
| Allow insecure | Skip certificate verification |
| Mux | Enable multiplexing |
| Fragment | TLS fragment spec (e.g. `1,40-60,30-50`) |
| Priority | Ascending order within the inbound |
| Enabled | Toggle the host on/off |

### Template variables

String fields are rendered per user, so one host definition works for everyone:

| Variable | Expands to |
|----------|-----------|
| `{USERNAME}` | The user's username |
| `{SERVER_IP}` | The node's address |
| `{SERVER_PORT}` | The advertised port |
| `{PROTOCOL}` | The inbound protocol (vless, vmess, …) |
| `{NETWORK}` | The transport (tcp, ws, grpc, …) |
| `{SECURITY}` | The transport security (tls, reality, none) |
| `{REMARK}` | The configured remark |

### Configure

1. Open an inbound and switch to the **Hosts** tab (or go to **Subscription Hosts**)
2. Click **Add Host**
3. Fill the override fields you need — leave the rest blank to inherit the inbound
4. Drag rows (or use **Reorder**) to set priority — lower priority renders first
5. Save. The next subscription fetch reflects the new hosts immediately.

!!! tip
    Set **Security** to `inbound_default` when you only want to change the address/SNI
    but keep whatever the inbound already negotiates.

---

## New Subscription Output Formats

**Location:** Users → user detail → **Subscription** (per-format links), and the public `/sub/{token}` endpoint

In addition to the existing `base64`, `clash`, and `singbox` outputs, v1.2.3 adds three
formats, selectable with the `?format=` query parameter:

| Format | `?format=` | Output |
|--------|-----------|--------|
| Xray / V2Ray JSON | `xray` | Raw Xray/V2Ray client JSON |
| Outline | `outline` | `ss://` Shadowsocks links for Outline |
| Plain links | `links` | V2rayN-style share links, one per line |

The full set is now: `base64`, `clash`, `singbox`, `xray`, `outline`, `links`. When no
format is given, the response is still auto-detected from the client's User-Agent.

```text
https://panel.example.com/sub/<token>?format=xray
https://panel.example.com/sub/<token>?format=outline
https://panel.example.com/sub/<token>?format=links
```

---

## Smart Routing Rule Packs

**Location:** Network & Nodes → Routing Packs

A **routing pack** is a reusable, named collection of routing rules. Build it once, then
apply it to any node or embed it into Clash/sing-box subscriptions — no more recreating
the same block-ads / Iran-direct / streaming rules per node.

### What you can do

| Action | Description |
|--------|-------------|
| Create / edit | Build a pack from ordered routing rules (same fields as node routing) |
| Apply to node | Replace a node's routing rules with the pack and resync its core |
| Set global default | One pack applies fleet-wide unless overridden |
| Assign per user | Give a specific user their own pack (embedded in their subscription) |

### Configure

1. Click **New Pack**, name it (e.g. "Iran — block ads + direct")
2. Add rules in priority order (domains, IPs, ports, inbound tags → outbound/balancer)
3. Save, then:
   - **Apply** the pack to a node to push it live, or
   - mark it **Default** for the whole fleet, or
   - assign it to a user from **Users → user → Routing Pack**

!!! note
    Per-user assignment takes precedence over the global default. A user with no
    assignment falls back to the default pack.

---

## Clean-IP Scanner

**Location:** Network & Nodes → Clean IP

Find well-performing Cloudflare/CDN edge IPs by scanning candidates and scoring them on
latency and packet loss. Use the best ones as the advertised address in a Subscription
Host or CDN chain.

### Results

| Field | Description |
|-------|-------------|
| IP | The candidate address |
| Latency (ms) | Round-trip latency |
| Loss % | Measured packet loss |
| Score | Combined ranking (higher is better) |
| Reachable | Whether the probe connected |
| Scanned at | Timestamp of the measurement |

### Configure

1. Paste candidate IPs into the scan box (one per line)
2. Optionally set a **port** (defaults to `443`)
3. Click **Scan** — results are scored and sorted best-first
4. Copy the top IPs into a Subscription Host address or a CDN/Relay chain

!!! warning "SSRF protection"
    Scan targets are validated before probing. Private, loopback, and link-local
    ranges are rejected so the scanner cannot be used to reach internal services.

---

## IP-Limit Enforcement

**Location:** Security → IP Limit

Enforce per-user concurrent IP/device caps to curb account sharing. When a user exceeds
their limit, the configured action fires.

### Policy

| Setting | Description |
|---------|-------------|
| Enabled | Turn enforcement on/off |
| Action | `warn`, `disable_temporarily`, or `kill_connections` |
| Alert cooldown | Seconds between repeated alerts for the same user |
| Restore after | Seconds before a temporarily disabled user is restored |

### Actions

| Action | Behavior |
|--------|----------|
| **warn** | Record an event and alert; take no connectivity action |
| **disable_temporarily** | Disable the user for `restore_after` seconds, then restore |
| **kill_connections** | Drop the offending connections immediately |

!!! note "Core differences"
    `kill_connections` is **Xray-only**. On sing-box nodes it automatically degrades to
    `disable_temporarily`, since sing-box cannot drop individual live connections.

### Events

The **Events** table lists each enforcement action: user, observed IP count, the IPs,
the action taken, and the timestamp.

---

## New Protocols, Transports & the Capability Matrix

**Location:** Inbounds → Add/Edit (protocol, transport, and security selectors)

v1.2.3 broadens protocol and transport coverage, and the inbound editor now constrains
your choices to what the selected node's core actually supports.

### Xray-core

- **Inbound protocols:** `vless`, `vmess`, `trojan`, `shadowsocks` (+ SS-2022 multi-user), `socks`, `http`, `dokodemo`
- **Transports:** `tcp`, `ws`, `grpc`, `httpupgrade`, `http`/`h2`, `xhttp`, `mkcp` (mKCP)
- **Security:** `none`, `tls`, `reality` — TLS takes an `alpn` list; TCP supports a header type of `none`/`http`; xHTTP supports a `mode` selector

### sing-box

- **Inbound protocols:** `vless`, `vmess`, `trojan`, `shadowsocks`, `hysteria2`, `tuic`, `wireguard`, `hysteria` (v1), `shadowtls`, `anytls`, `socks`, `http`, `naive`
- **Transports:** `tcp`, `ws`, `grpc`, `httpupgrade`, `http`/`h2`, `quic`

### Per-protocol capability matrix

The panel exposes a live **per-protocol capability matrix** (`GET /api/capabilities`),
and it is the **single source of truth** — the inbound editor only offers the
protocol/transport/security combinations the chosen core supports.

A few protocols carry **no stream transport**:

| Protocol | Core | Transport | Security |
|----------|------|-----------|----------|
| `socks` | both | none (raw TCP) | plaintext |
| `http` | both | none (raw TCP) | plaintext |
| `naive` | sing-box | none | **TLS mandatory** |
| `dokodemo` | xray | none (raw TCP/UDP) | plaintext |

!!! warning
    `socks` and `http` inbounds are **plaintext** — only expose them on trusted
    networks or behind a local relay. `naive` **mandates TLS**.

See [Protocols](13-protocols-config.md) and `docs/protocols.md` for full configuration
examples.
