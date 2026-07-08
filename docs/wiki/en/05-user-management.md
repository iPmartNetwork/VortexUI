# User Management

VortexUI uses a **user-centric model**: each user has a single identity that works across all assigned protocols and nodes.

---

## User List

**Sidebar → Users**

The user list shows all users with key information:

| Column | Description |
|--------|-------------|
| Username | Unique identifier |
| Status | Active, Limited, Expired, Suspended |
| Traffic | Used / Limit |
| Expire | Days remaining or expiry date |
| Connections | Current active connections |
| Actions | Edit, copy link, suspend, delete |

### Filtering & Search

- **Search**: Type username to filter
- **Status filter**: Active, Limited, Expired, All
- **Reseller filter**: Show only users owned by a specific reseller

### Bulk Actions

Select multiple users, then:
- **Bulk Edit** — change expire, data limit, inbounds
- **Bulk Delete** — remove selected users
- **Bulk Export** — download as CSV

---

## Creating a User

**Users → Add User**

### Basic Info

| Field | Description |
|-------|-------------|
| Username | Unique identifier (letters, numbers, underscore) |
| Data Limit | Traffic cap in GB (0 = unlimited) |
| Expire | Expiration date or duration |
| Device Limit | Max concurrent connections (0 = unlimited) |
| Note | Admin note (not visible to user) |

### Inbound Assignment

Check which inbounds this user can access:

- **Select All** — access to all current and future inbounds
- **Select Specific** — only checked inbounds

### Smart Quota

Enable progressive speed reduction:

| Tier | Usage | Speed |
|------|-------|-------|
| Tier 1 | 0-50% | 100% speed |
| Tier 2 | 50-80% | 50% speed |
| Tier 3 | 80-100% | 25% speed |

### Advanced Options

- **Auto Reset**: Reset traffic on day/week/month
- **Family Group**: Link to a shared data pool
- **Referral Code**: Pre-assigned invite code

---

## User Detail Page

Click any user to see full details:

### Overview Tab

- Traffic usage gauge
- Expire countdown
- Active connections
- Quick actions (copy link, suspend, delete)

### Subscription Tab

- **Subscription URL**: `https://domain.com/sub/{token}`
- **QR Code**: Scannable by mobile apps
- **Deep Links**: One-tap import for specific apps
- **Config Preview**: See actual proxy configuration

### Connections Tab

- List of active connections
- IP address, start time, traffic
- Force disconnect option

### History Tab

- Traffic log by day
- Connection history
- Admin action log

---

## Subscription Formats

VortexUI auto-detects client User-Agent and serves the appropriate format:

| Format | Content-Type | Clients |
|--------|--------------|---------|
| base64 | text/plain | v2rayNG, V2RayN, Nekoray |
| clash | text/yaml | Clash Meta, ClashX |
| singbox | application/json | sing-box |
| xray | application/json | Raw Xray/V2Ray |
| outline | text/plain | Outline |
| links | text/plain | One link per line |

Force specific format with `?format=clash` query parameter.

### Subscription Hosts

Override addresses per inbound for CDN fronting:

| Field | Purpose |
|-------|---------|
| Address | Override server address |
| Port | Override port |
| SNI | Override TLS server name |
| Host | Override HTTP host header |

Template variables:
- `{node.address}` — node IP
- `{node.name}` — node name
- `{user.username}` — username

---

## Self-Service Portal

End-users can access their own portal at:

```
https://your-domain.com/sub/{token}
```

### Portal Features

| Feature | Description |
|---------|-------------|
| Traffic Dashboard | Usage gauge, history chart |
| Connection List | Active connections |
| Subscription | Download configs, QR codes |
| Plan Purchase | Buy extensions (if shop enabled) |
| Support Tickets | Contact admin |
| Referral | Share invite link |

### Whitelabel Branding

Each reseller can customize their users' portal:

- Logo
- Brand colors
- Title
- Footer text
- Custom CSS

Configure at **Settings → Branding** (reseller view).

---

## Self-Service Shop

**Per-reseller shops** allow users to purchase plans directly.

### For Admins: Enable Shop

1. **Plans & Payments → Plans → Add Plan**
2. Set name, duration, traffic, price
3. **Plans & Payments → Payment Methods**
4. Configure ZarinPal, card-to-card, or crypto

### For Users: Purchase

1. Visit `/sub/{token}/shop`
2. Browse available plans
3. Select payment method
4. Complete payment
5. Plan applied automatically

---

## Family Groups

Share a data pool across multiple users.

### Creating a Family

1. **Users → Families → Add Family**
2. Set name, total data limit
3. Add member users

### How It Works

- All members share the family's traffic limit
- Individual users show "Family: {name}" instead of their own limit
- When family limit is reached, all members are limited

---

## Referral System

Users can invite others and earn rewards.

### Enable Referrals

1. **Settings → Features → Referrals**
2. Set reward per referral (e.g., +5 GB, +7 days)
3. Enable signup via referral link

### User Flow

1. User shares their referral link: `/invite/{code}`
2. New user signs up through link
3. Both users receive bonus

---

## Import & Export

### Import Users

**Users → Import**

Supported formats:
- CSV with columns: username, data_limit, expire, inbounds
- JSON array of user objects
- From 3x-ui database
- From Marzban export

### Export Users

**Users → Export**

- CSV with all user data
- JSON for API integration
- Filter by status, reseller, date range

---

## User Lifecycle

```
Created → Active → (Limited/Expired/Suspended) → Deleted
              ↑           ↓
              └───── Renewed ─────┘
```

| Status | Cause | User Can Connect? |
|--------|-------|-------------------|
| Active | Normal state | ✅ Yes |
| Limited | Data limit reached | ❌ No |
| Expired | Expire date passed | ❌ No |
| Suspended | Admin action | ❌ No |
| Deleted | Removed | ❌ No |

### Notifications

Configure alerts at **Settings → Notifications**:

- **Quota 80%**: User approaching limit
- **Quota 100%**: User reached limit
- **Expire 3d**: User expires soon
- **Expired**: User has expired
