# First Steps

!!! info "Estimated Time"
    From login to a working subscription link: **5 minutes**.

---

## 1. Login & Secure Your Account

1. Navigate to your panel URL (e.g. `https://panel.example.com`)
2. Log in with the admin credentials you created during installation
3. **Change your password** immediately: Settings → Profile → Change Password

### Enable TOTP 2FA

1. Go to **Settings → Two-Factor Authentication**
2. Click **Enable** — scan the QR code with Google Authenticator or any TOTP app
3. Enter the 6-digit code to confirm
4. Store your recovery codes securely

!!! warning
    Without 2FA, anyone with your password has full panel access. Enable it before exposing the panel to the internet.

---

## 2. Add Your First Node

=== "Local Node (Single Server)"

    If your panel and proxy core run on the same machine:

    1. Go to **Nodes → Add Node**
    2. Select **Local Node**
    3. Choose a core: **Xray-core** or **sing-box**
    4. Click **Create**
    5. The node starts automatically and shows "Online"

=== "Remote Node (Enrollment Wizard)"

    For a separate server:

    1. Go to **Nodes → Add Node**
    2. Select **Remote Node**
    3. The wizard shows a one-line install command — copy it
    4. SSH into your remote server and paste the command
    5. Wait for the agent to connect — the node appears as "Online"

!!! tip
    The enrollment wizard handles certificate exchange, core installation, and service registration. No manual certificate management needed.

---

## 3. Create Your First Inbound

A quick VLESS + REALITY setup (recommended for censorship-resistant connections):

1. Go to **Inbounds → Add Inbound**
2. Configure:

    | Field | Value |
    |-------|-------|
    | Protocol | VLESS |
    | Node | Your node |
    | Port | `443` |
    | Transport | TCP |
    | Security | REALITY |
    | Dest (SNI) | `www.google.com:443` (or use Reality Scanner) |
    | Server Names | `www.google.com` |
    | Short IDs | Auto-generated (or enter your own) |

3. Click **Create**

!!! tip "Reality Scanner"
    Use **Security → Reality Scanner** to find the best SNI domains for your server location. Pick the highest-scoring result for low latency and stable connections.

---

## 4. Create Your First User

1. Go to **Users → New User**
2. Fill in:

    | Field | Value |
    |-------|-------|
    | Username | `testuser` |
    | Data limit | `50 GB` (or `0` for unlimited) |
    | Expire | 30 days from now |
    | Device limit | `3` |
    | Inbounds | Select your VLESS inbound |

3. Click **Create**
4. The user's subscription token is generated automatically

---

## 5. Test the Subscription Link

1. From the user list, click the user's name to open their detail page
2. Copy the **Subscription Link** (or scan the QR code)
3. Import into your client app:

=== "v2rayNG (Android)"

    1. Open v2rayNG → tap **+** → **Import config from clipboard**
    2. Or scan the QR code directly

=== "Clash Meta (Desktop/Mobile)"

    ```
    https://panel.example.com/sub/<token>?format=clash
    ```
    Add as a subscription URL in Clash.

=== "sing-box (iOS/Android)"

    ```
    https://panel.example.com/sub/<token>?format=singbox
    ```
    Add as a remote profile.

4. Connect and verify traffic flows — check the user's usage in the panel

---

## 6. Enable Notifications

Set up Telegram notifications to stay informed:

1. Go to **Settings → Notifications → Telegram**
2. Enter your **Bot Token** (create one via [@BotFather](https://t.me/BotFather))
3. Enter your **Admin Chat ID** (get it from [@userinfobot](https://t.me/userinfobot))
4. Click **Test** to verify delivery
5. Enable desired events:
    - User created/expired/limited
    - Node offline
    - Quota threshold reached
    - Backup completed

---

## What's Next?

You now have a working VortexUI setup with one node, one inbound, and one user. From here:

- **[Dashboard](04-dashboard.md)** — explore real-time monitoring
- **[User Management](05-user-management.md)** — bulk operations, quotas, portal
- **[Node Management](06-node-management.md)** — add more nodes, configure auto-migration
- **[Plans & Payments](09-plans-payments.md)** — set up self-service plan purchases
- **[Security](08-security-administration.md)** — configure anti-censorship features
