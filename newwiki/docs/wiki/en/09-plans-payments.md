# Plans & Payments

VortexUI includes a complete commerce system: plans, multiple payment gateways, reseller wallets, and order management.

---

## Overview

```
User → Shop → Select Plan → Pay → Order → Auto-Provision
                                     ↓
                          Reseller Wallet (debited)
```

---

## Plans

Plans define what a user gets: traffic, duration, device limits, and price.

**Plans & Payments → Plans**

### Creating a Plan

1. Click **Add Plan**
2. Configure:

| Field | Description |
|-------|-------------|
| Name | Display name (e.g., "Premium 100GB") |
| Data Limit | Traffic in GB (0 = unlimited) |
| Duration | Days until expiry |
| Device Limit | Max concurrent connections |
| Price (Toman) | Price for ZarinPal/card |
| Price (USD) | Price for crypto |
| Inbounds | Which inbounds this plan grants |

### Plan Visibility

| Type | Who Sees It |
|------|-------------|
| Admin plans | All users (created by sudo) |
| Reseller plans | Only that reseller's users |

> **Per-reseller plans:** Each reseller creates their own plans with their own pricing. These appear only in that reseller's shop at `/sub/{token}/shop`.

---

## Payment Gateways

VortexUI supports three payment methods, configurable per reseller.

### ZarinPal (Online — Iran)

**Plans & Payments → Payment Methods → ZarinPal**

| Field | Description |
|-------|-------------|
| Merchant ID | Your ZarinPal merchant ID |
| Enabled | Toggle on/off |
| Sandbox | Test mode |

Flow:
1. User selects plan → clicks Pay
2. Redirected to ZarinPal
3. Completes payment
4. Redirected back → order auto-approved
5. Plan applied instantly

### Card-to-Card (Manual)

**Plans & Payments → Payment Methods → Card-to-Card**

| Field | Description |
|-------|-------------|
| Card Number | Destination card |
| Holder Name | Card holder name |
| Enabled | Toggle on/off |

Flow:
1. User selects plan → sees card details
2. Transfers money manually
3. Uploads payment proof (screenshot)
4. Admin/reseller reviews → approves
5. Plan applied on approval

### Crypto (NowPayments)

**Plans & Payments → Payment Methods → Crypto**

| Field | Description |
|-------|-------------|
| API Key | NowPayments API key |
| IPN Secret | HMAC secret for webhooks |
| Wallet addresses | BTC, USDT, etc. |

Flow:
1. User selects plan → chooses crypto
2. Sends payment to wallet
3. Submits TX hash
4. IPN webhook confirms → order approved
5. Plan applied automatically

---

## Orders

**Plans & Payments → Orders**

Track all purchases and their status.

| Status | Meaning |
|--------|---------|
| Pending | Awaiting payment |
| Review | Proof uploaded, awaiting approval |
| Approved | Payment confirmed, plan applied |
| Rejected | Payment rejected |
| Expired | Order timed out |

### Order Actions

- **Approve**: Confirm payment, apply plan
- **Reject**: Decline with reason
- **View Proof**: See uploaded screenshot
- **Export PDF**: Generate invoice

---

## Reseller Wallet

> **Reseller Platform (1.2.9+)**

Resellers operate on a credit system managed by their wallet.

**Plans & Payments → Wallet** (reseller view)

### Credit Types

| Credit Type | Consumed When |
|-------------|---------------|
| Traffic credits | Users consume data (consumed mode) or assigned limits (allocated mode) |
| User credits | Creating new users |

### Wallet Operations

| Operation | Description |
|-----------|-------------|
| Top-up request | Reseller requests credits; sudo approves |
| Auto-deduct | System deducts as users consume/created |
| Ledger | Full history of all credit changes |
| Balance | Current available credits |

### Traffic Quota Modes

| Mode | Counts Against Pool When... |
|------|------------------------------|
| **Allocated** | Reseller assigns data limits (sum of limits) |
| **Consumed** | Users actually consume traffic |

---

## Reseller Wallet Ledger

Every credit change is recorded:

| Field | Content |
|-------|---------|
| Timestamp | When the change occurred |
| Type | Top-up, deduct, adjust |
| Amount | Credit delta (+/-) |
| Balance | Balance after change |
| Reason | Description (e.g., "User alice created") |
| Actor | Who made the change |

---

## Top-Up Workflow

```
Reseller requests top-up
         ↓
Sudo admin reviews queue
         ↓
   Approve / Reject
         ↓
Credits added to wallet
         ↓
Ledger entry created
```

### For Resellers: Request Top-Up

1. **Wallet → Request Top-Up**
2. Enter amount and note
3. Submit → status "Pending"
4. Wait for sudo approval

### For Sudo: Approve Top-Ups

1. **Plans & Payments → Top-Up Queue**
2. Review pending requests
3. Approve or reject
4. Credits applied instantly

### Quick Quota Adjust

Sudo admins can quickly adjust reseller quotas:

- **+50 accounts** button
- **+10 GB** button
- **+50 GB** button

---

## Self-Service Shop

The shop lets users purchase plans directly from their portal.

### Shop URL

```
https://your-domain.com/sub/{token}/shop
```

### Shop Features

- Browse available plans (filtered by reseller)
- See only their reseller's payment methods
- Purchase renewals or upgrades
- View order history
- Download invoices

### Enabling the Shop

1. Create at least one plan
2. Configure at least one payment method
3. Shop auto-appears in user portal

---

## Referral Program

> **New in 1.3.0**

Users earn rewards for inviting others.

**Settings → Features → Referrals**

| Setting | Description |
|---------|-------------|
| Enabled | Turn referrals on/off |
| Reward (data) | GB added per referral |
| Reward (days) | Days added per referral |
| Referrer bonus | Reward for the inviter |
| Referee bonus | Reward for the new user |

### User Flow

1. User finds their referral link in portal: `/invite/{code}`
2. Shares with friends
3. New user signs up through link
4. Both receive configured bonus

---

## Pricing Strategy Tips

### For Resellers
- Set competitive plan prices
- Offer volume discounts (larger plans, better GB/price ratio)
- Use referral rewards to grow user base

### For Sudo Admins
- Set reasonable reseller quotas
- Monitor wallet balances to prevent service interruption
- Use "consumed" mode for pay-as-you-go resellers
- Use "allocated" mode for prepaid resellers

---

## Financial Reporting

### Export Options

- **Orders CSV**: All transactions in date range
- **Invoice PDF**: Per-order invoices
- **Wallet Ledger CSV**: Credit history

### Analytics

- Total revenue by period
- Revenue by reseller
- Popular plans
- Payment method breakdown
