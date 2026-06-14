// outbound-uri.ts — converts proxy share links (vmess/vless/trojan/ss/hysteria2/
// wireguard) into 3x-ui-style outbound JSON objects, and exposes per-protocol
// default templates for the JSON editor.

export type OutboundJSON = Record<string, unknown>;

// Default editor template (vless), matching 3x-ui's outbound schema.
export const DEFAULT_OUTBOUND_TEMPLATE: OutboundJSON = {
  protocol: "vless",
  settings: { address: "", port: 443, id: "", flow: "", encryption: "none" },
  streamSettings: {
    network: "tcp",
    tcpSettings: { header: { type: "none" } },
    security: "none",
  },
};

function b64decode(s: string): string {
  // tolerate url-safe base64 and missing padding
  let t = s.replace(/-/g, "+").replace(/_/g, "/");
  while (t.length % 4) t += "=";
  return decodeURIComponent(escape(atob(t)));
}

// buildStream maps query params from a vless/trojan share link into a
// streamSettings object (network + transport + security sub-objects).
function buildStream(p: URLSearchParams): OutboundJSON {
  const net = p.get("type") || "tcp";
  const ss: OutboundJSON = { network: net };
  switch (net) {
    case "tcp":
      ss.tcpSettings = { header: { type: p.get("headerType") || "none" } };
      break;
    case "ws":
      ss.wsSettings = { path: p.get("path") || "/", headers: p.get("host") ? { Host: p.get("host") } : {} };
      break;
    case "grpc":
      ss.grpcSettings = { serviceName: p.get("serviceName") || "", multiMode: p.get("mode") === "multi" };
      break;
    case "http":
    case "h2":
      ss.httpSettings = { path: p.get("path") || "/", host: p.get("host") ? [p.get("host")] : [] };
      break;
    case "xhttp":
    case "splithttp":
      ss.xhttpSettings = { path: p.get("path") || "/", host: p.get("host") || "" };
      break;
  }
  const security = p.get("security") || "none";
  ss.security = security;
  if (security === "tls") {
    ss.tlsSettings = {
      serverName: p.get("sni") || "",
      fingerprint: p.get("fp") || "",
      alpn: p.get("alpn") ? p.get("alpn")!.split(",") : [],
    };
  } else if (security === "reality") {
    ss.realitySettings = {
      serverName: p.get("sni") || "",
      fingerprint: p.get("fp") || "",
      publicKey: p.get("pbk") || "",
      shortId: p.get("sid") || "",
      spiderX: p.get("spx") || "",
    };
  }
  return ss;
}

function parseVless(uri: string): OutboundJSON {
  const u = new URL(uri);
  const p = u.searchParams;
  return {
    tag: decodeURIComponent(u.hash.slice(1)) || `vless-${u.hostname}`,
    protocol: "vless",
    settings: {
      address: u.hostname,
      port: Number(u.port) || 443,
      id: decodeURIComponent(u.username),
      flow: p.get("flow") || "",
      encryption: p.get("encryption") || "none",
    },
    streamSettings: buildStream(p),
  };
}

function parseTrojan(uri: string): OutboundJSON {
  const u = new URL(uri);
  const p = u.searchParams;
  return {
    tag: decodeURIComponent(u.hash.slice(1)) || `trojan-${u.hostname}`,
    protocol: "trojan",
    settings: { address: u.hostname, port: Number(u.port) || 443, password: decodeURIComponent(u.username) },
    streamSettings: buildStream(p),
  };
}

function parseVmess(uri: string): OutboundJSON {
  const raw = JSON.parse(b64decode(uri.slice("vmess://".length)));
  const net = raw.net || "tcp";
  const ss: OutboundJSON = { network: net };
  if (net === "ws") ss.wsSettings = { path: raw.path || "/", headers: raw.host ? { Host: raw.host } : {} };
  else if (net === "grpc") ss.grpcSettings = { serviceName: raw.path || "" };
  else if (net === "h2" || net === "http") ss.httpSettings = { path: raw.path || "/", host: raw.host ? [raw.host] : [] };
  else ss.tcpSettings = { header: { type: raw.type || "none" } };
  ss.security = raw.tls === "tls" ? "tls" : "none";
  if (raw.tls === "tls") ss.tlsSettings = { serverName: raw.sni || raw.host || "", alpn: raw.alpn ? String(raw.alpn).split(",") : [] };
  return {
    tag: raw.ps || `vmess-${raw.add}`,
    protocol: "vmess",
    settings: { address: raw.add, port: Number(raw.port) || 443, id: raw.id, alterId: Number(raw.aid) || 0, security: raw.scy || "auto" },
    streamSettings: ss,
  };
}

function parseShadowsocks(uri: string): OutboundJSON {
  const hashIdx = uri.indexOf("#");
  const tag = hashIdx >= 0 ? decodeURIComponent(uri.slice(hashIdx + 1)) : "";
  const body = (hashIdx >= 0 ? uri.slice(0, hashIdx) : uri).slice("ss://".length);
  let method = "", password = "", address = "", port = 0;
  if (body.includes("@")) {
    // SIP002: ss://base64(method:password)@host:port  (or plain userinfo)
    const at = body.lastIndexOf("@");
    const userinfo = body.slice(0, at);
    const hostport = body.slice(at + 1).split("?")[0];
    let creds = userinfo;
    try { creds = b64decode(userinfo); } catch { /* already plain */ }
    [method, password] = creds.split(":");
    const colon = hostport.lastIndexOf(":");
    address = hostport.slice(0, colon);
    port = Number(hostport.slice(colon + 1)) || 0;
  } else {
    // legacy: ss://base64(method:password@host:port)
    const dec = b64decode(body.split("?")[0]);
    const at = dec.lastIndexOf("@");
    [method, password] = dec.slice(0, at).split(":");
    const hostport = dec.slice(at + 1);
    const colon = hostport.lastIndexOf(":");
    address = hostport.slice(0, colon);
    port = Number(hostport.slice(colon + 1)) || 0;
  }
  return {
    tag: tag || `ss-${address}`,
    protocol: "shadowsocks",
    settings: { address, port, password, method },
    streamSettings: { network: "tcp", security: "none" },
  };
}

function parseHysteria2(uri: string): OutboundJSON {
  const u = new URL(uri);
  const p = u.searchParams;
  return {
    tag: decodeURIComponent(u.hash.slice(1)) || `hy2-${u.hostname}`,
    protocol: "hysteria2",
    settings: {
      address: u.hostname,
      port: Number(u.port) || 443,
      password: decodeURIComponent(u.username),
      obfs: p.get("obfs") || "",
      obfsPassword: p.get("obfs-password") || "",
    },
    streamSettings: { security: "tls", tlsSettings: { serverName: p.get("sni") || "", insecure: p.get("insecure") === "1" } },
  };
}

function parseWireguard(uri: string): OutboundJSON {
  const u = new URL(uri);
  const p = u.searchParams;
  return {
    tag: decodeURIComponent(u.hash.slice(1)) || `wg-${u.hostname}`,
    protocol: "wireguard",
    settings: {
      secretKey: decodeURIComponent(u.username),
      address: p.get("address") ? p.get("address")!.split(",") : [],
      peers: [{ publicKey: p.get("publickey") || "", endpoint: `${u.hostname}:${u.port || 51820}`, preSharedKey: p.get("presharedkey") || "" }],
      mtu: Number(p.get("mtu")) || 1420,
    },
    streamSettings: {},
  };
}

// parseShareLink converts a proxy share URI into an outbound JSON object.
// Throws if the scheme is unrecognized or the link is malformed.
export function parseShareLink(uri: string): OutboundJSON {
  const trimmed = uri.trim();
  const scheme = trimmed.slice(0, trimmed.indexOf("://")).toLowerCase();
  switch (scheme) {
    case "vless": return parseVless(trimmed);
    case "vmess": return parseVmess(trimmed);
    case "trojan": return parseTrojan(trimmed);
    case "ss": return parseShadowsocks(trimmed);
    case "hysteria2":
    case "hy2": return parseHysteria2(trimmed);
    case "wireguard":
    case "wg": return parseWireguard(trimmed);
    default: throw new Error(`unsupported scheme: ${scheme || "?"}`);
  }
}
