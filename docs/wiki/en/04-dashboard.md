# Dashboard

!!! info "Real-Time"
    All dashboard data updates via Server-Sent Events (SSE) — no manual refresh needed.

!!! tip "Command Tower UI (v1.2.9)"
    The **Overview** page includes live Command Tower widgets — traffic range tabs,
    top users with protocol badges, and node geo/ping. See
    [Command Tower (v1.2.9)](17-v129-panel-redesign.md).

!!! tip "Veltrix UI (v1.2.8)"
    The **Overview** page uses the new Veltrix design — live stat cards, node fleet
    health, and top traffic users. See [Veltrix UI (v1.2.8)](16-v128-ui-refresh.md)
    for navigation, command palette, and language switcher details.

---

## Overview Widgets

The main dashboard is a **customizable widget grid**. Drag, drop, and resize widgets to build your preferred layout.

### Available Widgets

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

### Customization

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

!!! note
    Geo data requires node agents to report connection metadata. If the map is empty, verify your nodes are running the latest agent version.

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

!!! tip
    Use analytics to identify your busiest hours and plan node capacity accordingly.

---

## Command Palette

Press ++ctrl+k++ (or ++cmd+k++ on macOS) to open the command palette — a fuzzy search across the entire panel.

### What you can search

- **Users** — jump to any user by name
- **Nodes** — navigate to node detail
- **Inbounds** — open inbound configuration
- **Pages** — jump to any panel page
- **Actions** — create user, add node, run backup, etc.

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| ++ctrl+k++ | Open command palette |
| ++ctrl+shift+n++ | New user |
| ++ctrl+shift+d++ | Toggle dark/light mode |
| ++ctrl+period++ | Open notifications |
| ++escape++ | Close modal/palette |
| ++g++ then ++d++ | Go to dashboard |
| ++g++ then ++u++ | Go to users |
| ++g++ then ++n++ | Go to nodes |
