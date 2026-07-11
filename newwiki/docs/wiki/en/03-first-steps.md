# First Steps

This guide walks you through your first 10 minutes with VortexUI — from login to a working user subscription.

---

## Step 1: Login

1. Open `https://your-domain.com` (or `http://server-ip:8080` if no domain)
2. Enter the admin credentials you created during installation
3. You'll land on the **Dashboard**

> **First-time?** The onboarding tour will guide you through the main features. Click "Start Tour" or dismiss to explore on your own.

---

## Step 2: Add a Node

Nodes are servers running proxy cores (Xray or sing-box) that serve your users.

### Local Node (Single Server)

If panel and proxy run on the same machine:

1. Go to **Nodes** → **Add Node**
2. Select **Local Node**
3. Choose core: **Xray** (recommended) or **sing-box**
4. Click **Create**
5. Status should show 🟢 Online within seconds

### Remote Node

If proxy runs on a separate server:

1. Go to **Nodes** → **Add Node**
2. Select **Remote Node**
3. Fill in:
   - **Name**: e.g., "Frankfurt-01"
   - **Address**: server IP or hostname
   - **Port**: API port (default 62050)
4. Copy the generated install command
5. SSH into your node server and run:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install-node.sh)
```

6. Enter the panel address and enrollment token when prompted
7. Back in the panel, status should show 🟢 Online

---

## Step 3: Create an Inbound

Inbounds are protocol configurations that listen for client connections.

1. Go to **Nodes** → select your node → **Inbounds** tab
2. Click **Add Inbound**
3. Choose a protocol:

### Recommended: VLESS + Reality

| Field | Value |
|-------|-------|
| Protocol | VLESS |
| Port | 443 |
| Transport | TCP |
| Security | Reality |
| Dest | www.google.com:443 |
| Server Names | www.google.com |

4. Click **Create**

> **Tip:** Use the **Reality Scanner** (Security → Reality) to find optimal SNI domains for your server location.

### Alternative: VMess + WebSocket + TLS

Good for CDN fronting (Cloudflare):

| Field | Value |
|-------|-------|
| Protocol | VMess |
| Port | 443 |
| Transport | WebSocket |
| Path | /vmws |
| Security | TLS |

---

## Step 4: Create a User

1. Go to **Users** → **Add User**
2. Fill in:

| Field | Value |
|-------|-------|
| Username | your-username |
| Data Limit | 50 GB (or unlimited: 0) |
| Expire | 30 days from now |
| Device Limit | 2 (optional) |

3. In **Inbounds** section, check the inbound(s) you created
4. Click **Create**

---

## Step 5: Get Subscription Link

1. On the user detail page, find **Subscription**
2. Copy the subscription URL:
   ```
   https://your-domain.com/sub/abc123xyz
   ```
3. Or scan the **QR Code** directly with a mobile app

### Import into Client Apps

| App | Platform | How to Import |
|-----|----------|---------------|
| v2rayNG | Android | Scan QR or paste URL |
| v2rayN | Windows | Subscription → Add → Paste URL |
| Nekoray | Windows/Linux | Subscription → New → Paste URL |
| Clash Meta | All | Profiles → Import from URL |
| Shadowrocket | iOS | Scan QR or Add Subscription |
| Streisand | iOS | Settings → Subscriptions → Add |

---

## Step 6: Verify Connection

1. In your client app, select the imported server
2. Connect
3. Visit [https://ipinfo.io](https://ipinfo.io) — you should see your node's IP

In VortexUI:
- **Dashboard** → check real-time traffic
- **Monitor** → see active connections
- **Users** → verify traffic usage updating

---

## Step 7: Harden Your Panel

### Enable 2FA

1. Click your avatar (top-right) → **Profile**
2. Enable **Two-Factor Authentication**
3. Scan QR with Google Authenticator / Authy
4. Save backup codes securely

### Set Strong JWT Secret

Ensure `VORTEX_JWT_SECRET` in your `.env` is at least 32 random bytes:

```bash
openssl rand -base64 32
```

### Configure IP Whitelist

1. **Settings** → **Security** → **IP Guard**
2. Add your management IPs to the whitelist
3. Enable enforcement

### Set Up Backups

1. **Settings** → **Backup**
2. Configure Telegram or S3 destination
3. Set schedule (e.g., daily at 3 AM)

---

## What's Next?

| Task | Guide |
|------|-------|
| Customize dashboard | [Dashboard](04-dashboard.md) |
| Bulk import users | [User Management](05-user-management.md) |
| Add more nodes | [Node Management](06-node-management.md) |
| Set up routing rules | [Network Policy](07-network-policy.md) |
| Configure anti-censorship | [Security](08-security-administration.md) |
| Create reseller accounts | [Plans & Payments](09-plans-payments.md) |
| Enable notifications | [Notifications](10-notifications.md) |

---

## Troubleshooting First Setup

### Node shows "Offline"

1. Check node server is reachable: `ping node-ip`
2. Verify port is open: `telnet node-ip 62050`
3. Check node agent logs: `journalctl -u vortex-node -f`
4. Ensure panel address in node config is correct

### Subscription URL returns 404

1. Verify the user has at least one inbound assigned
2. Check user hasn't expired or been suspended
3. Verify the domain resolves to your panel server

### Client can't connect

1. Check inbound port is open: `telnet node-ip 443`
2. Verify protocol settings match
3. Check firewall rules on node server
4. Test with `curl -v https://node-ip:port`

### Login fails after panel restart

1. Check database is running: `systemctl status postgresql`
2. Verify JWT secret hasn't changed
3. Check panel logs: `journalctl -u vortexui -f`
