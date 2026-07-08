# Dashboard

> **Real-Time:** All dashboard data updates via Server-Sent Events (SSE) — no manual refresh needed.

The main dashboard is a **customizable widget grid**. Drag, drop, and resize widgets to build your preferred layout.

---

## Available Widgets

| Widget | Content |
|--------|---------|
| Users summary | Total, active, limited, expired counts |
| Traffic summary | Today's upload/download, monthly total |
| Node status | Online/offline node counts with health indicators |
| System gauges | CPU, RAM, disk usage (animated rings) |
| Active connections | Current live tunnel count |
| Recent events | Latest system events (user created, node offline, etc.) |
| Quick actions | Create user, add node, run scan shortcuts |
| Reseller pool | Your account/traffic quota remaining (reseller view) |

---

## Customization

1. Click the **grid icon** (top-right of dashboard)
2. Drag widgets to rearrange
3. Resize by dragging the corner handle
4. Toggle widget visibility with checkboxes
5. Click **Save Layout** — persisted per admin account

---

## Real-Time Gauges

Animated gauges show live system metrics for each node:

- **CPU** — current utilization percentage
- **RAM** — used / total with percentage
- **Bandwidth** — current throughput (upload + download)
- **Connections** — active tunnel count

Gauges update every 3 seconds via SSE push.

---

## Charts

| Chart | Description |
|-------|-------------|
| Traffic (line) | Upload/download over time (hourly/daily/weekly) |
| Connections (area) | Active connections over time |
| Users (bar) | New users by day/week |
| Node load (stacked) | Per-node CPU distribution |

All charts support time range selection: **24h**, **7d**, **30d**, or custom date range.

---

## World Map Geo-Visualization

A heatmap showing where your users connect from, based on GeoIP data from node agents.

- Hover over a country for connection count and traffic volume
- Click for detailed breakdown (top cities, top users from that country)
- Color intensity reflects traffic volume

> **Note:** Geo data requires node agents to report connection metadata. If the map is empty, verify your nodes are running the latest agent version.

---

## Monitor Page

**Dashboard → Monitor** — real-time view of all active connections.

| Column | Description |
|--------|-------------|
| User | Username |
| Node | Which server they're connected to |
| IP | Client source IP |
| Protocol | VLESS, VMess, Trojan, etc. |
| Duration | How long the connection has been active |
| Traffic | Upload/download for this session |

Features:
- Auto-refreshes every 3 seconds
- Filter by node, protocol, or user
- Sort by any column
- Click a user to jump to their detail page

---

## Analytics Page

**Dashboard → Analytics** — traffic insights aggregated by country, user, and time.

### Sections

| Section | Shows |
|---------|-------|
| Summary cards | Total upload, download, country count |
| Traffic by Country | Geo breakdown — country, connections, bytes |
| Top Users | Ranked by total traffic consumed |
| Peak Hours | Bar chart of hourly traffic volume |

### Time Ranges

Select **Last 24h**, **Last 7 days**, **Last 30 days**, or a custom date range.

### Export

Click **Export CSV** to download geo + user data as a spreadsheet for the selected time range.

> **Tip:** Use analytics to identify your busiest hours and plan node capacity accordingly.

---

## Command Palette

Press `Ctrl+K` (or `Cmd+K` on macOS) to open the command palette — a fuzzy search across the entire panel.

### What you can search

- **Users** — jump to any user by name
- **Nodes** — navigate to node detail
- **Inbounds** — open inbound configuration
- **Pages** — jump to any panel page
- **Actions** — create user, add node, run backup, etc.

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+K | Open command palette |
| Ctrl+Shift+N | New user |
| Ctrl+Shift+D | Toggle dark/light mode |
| Ctrl+. | Open notifications |
| Esc | Close modal/palette |
| G then D | Go to dashboard |
| G then U | Go to users |
| G then N | Go to nodes |

---

## Veltrix UI Features

### Collapsible Sidebar

- Click the hamburger icon to collapse/expand
- Collapsed mode shows only icons
- State persists across sessions

### Theme Switcher

- Click the sun/moon icon in the header
- Smooth animated transition between dark and light
- Preference saved to your profile

### Language Selector

- 8 languages available: EN, FA, TR, AR, RU, ZH, JA, ES
- Full RTL support for Farsi and Arabic
- Language saved per-admin

### Notification Bell

- Real-time notifications for important events
- Mark as read, clear all
- Click to navigate to related item

---

## Mobile Dashboard

On mobile devices, the dashboard adapts:

- Single-column widget layout
- Bottom navigation bar
- Pull-to-refresh for manual update
- Bottom sheets for quick actions
- Swipe gestures for navigation

---

## Dashboard Tips

### For Operators
- Pin the "Quick Actions" widget at the top for fast access
- Use the world map to identify geographic traffic patterns
- Set up analytics exports for billing reports

### For Resellers
- The "Reseller Pool" widget shows your quota usage at a glance
- Monitor "Recent Events" for user activity
- Use "Top Users" to identify heavy consumers

### For Monitoring
- Keep the Monitor page open for live connection tracking
- Set up browser notifications for offline node alerts
- Export analytics weekly for trend analysis
