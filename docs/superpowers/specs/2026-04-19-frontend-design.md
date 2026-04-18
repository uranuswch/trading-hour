# Frontend Design — Trading Hour Web UI

**Date:** 2026-04-19  
**Status:** Approved

## Overview

A dark, gold-themed web dashboard that visualises live trading-hour status for all 12 supported markets. Backed by a Go HTTP server that wraps the existing `tradinghour` library. No JS framework, no build step.

## Architecture

```
cmd/server/main.go      ← Go HTTP server binary
web/static/index.html   ← single-page UI
web/static/style.css    ← gold/dark theme
web/static/app.js       ← fetch + DOM manipulation
```

The binary in `cmd/server/` imports the `tradinghour` package directly and serves two things:

1. **Static files** — `GET /` and `GET /static/*` serve from `web/static/`
2. **JSON API** — three endpoints:

| Endpoint | Handler | Returns |
|---|---|---|
| `GET /api/status` | `handleStatus` | `IsOpen` result for all 12 markets |
| `GET /api/timeline/{market}?date=YYYY-MM-DD` | `handleTimeline` | Full `DaySchedule` (phases, holiday flag + name, market tz) |
| `GET /api/nextopen/{market}` | `handleNextOpen` | Next open instant as RFC3339 string |

All API responses are JSON. Date parameter defaults to today in the market's local timezone when omitted. Unknown market returns `404`. CORS is not needed (same origin).

## Visual Design

**Theme:** Near-black background (`#0c0c0e` → `#12100a` gradient), amber/gold accents (`#f59e0b`, `#fcd34d`), muted red for closed (`#f87171`). System sans-serif font. No external font CDN.

**Color mapping:**
- Open session: gold border + soft glow, `●` prefix
- Closed: dim red border, `○` prefix  
- Premarket / postmarket / overnight: amber dim, `◑` prefix

## UI Components

### Hero Section
- Full-width centered banner
- Gradient gold headline: `"N Markets Open"` (live count of currently-open markets)
- Subtitle: `"of 12 tracked · refreshes every 30s"`
- Thin amber separator line below

### Market Pills Grid
- Wrapping flex row of 12 pill chips, one per market
- Each pill: status dot + market name (e.g. `● NASDAQ`, `○ HKEX`)
- Active (open) pills: gold border + glow; closed: red-tinted; partial: amber-dim
- Clicking a pill opens the side drawer for that market

### Next Open Spotlight
- Small card below the pills
- Format: `NEXT OPEN · HKEX · in 6h 22m · 09:30 HKT`
- Shows the soonest upcoming market open across all currently-closed markets
- Updates on each 30-second status refresh

### Side Drawer
- Slides in from the right (~360px wide); rest of page dims with a semi-transparent overlay
- `×` button or click-outside closes it
- **Header:** market name + live status badge (`● OPEN · Regular`)
- **Session list:** one row per phase for the selected date, active phase highlighted in gold
- **Timeline bar:** proportional visual bar across 00:00–24:00, coloured segments per session, vertical cursor at current time (today only)
- **Date picker:** `<input type="date">` defaulting to today in market's local tz; changing date re-fetches `/api/timeline/{market}?date=…` and re-renders sessions + bar; status badge becomes "Holiday — {name}" or "Half Day — {name}" when applicable

## Data Flow

### Page load
1. Fetch `GET /api/status` → render all 12 pills + hero count
2. For each closed market, fetch `GET /api/nextopen/{market}` → find soonest → render spotlight card
3. Start `setInterval(refresh, 30_000)`:
   - Re-fetch `/api/status`
   - Update pill states, hero count, spotlight card
   - If drawer is open and viewing today, also re-fetch timeline to update active-phase highlight + cursor

### Pill click
1. Drawer slides in via CSS `transform: translateX(0)`
2. Fetch `GET /api/timeline/{market}?date=today`
3. Render session list + timeline bar; set date picker to today

### Date change in drawer
1. Fetch `GET /api/timeline/{market}?date=YYYY-MM-DD`
2. Re-render session list + timeline bar
3. Status badge reflects holiday/half-day if applicable; timeline cursor hidden for non-today dates

## File Details

### `cmd/server/main.go`
- `http.FileServer` for `web/static/`
- `http.HandleFunc` for the three API routes
- JSON marshalling of `tradinghour.Status`, `tradinghour.DaySchedule`, and a simple `NextOpenResponse{Market, Time, Local string}`
- Reads `PORT` env var, defaults to `8080`
- Single `main()` — no framework

### `web/static/index.html`
- Minimal boilerplate, links `style.css` and `app.js`
- Empty shells for `#hero`, `#pills`, `#spotlight`, `#drawer`

### `web/static/style.css`
- CSS custom properties for gold palette
- Pill, drawer, timeline bar, session row styles
- Drawer slide-in animation (`transform` + `transition`)
- Responsive: pills wrap naturally; drawer goes full-width on narrow screens

### `web/static/app.js`
- `fetchStatus()`, `fetchTimeline(market, date)`, `fetchNextOpen(market)`
- `renderPills(statuses)`, `renderDrawer(schedule)`, `renderTimeline(phases, date)`
- `setInterval` refresh loop
- No dependencies — plain `fetch` + DOM APIs

## Out of Scope
- Authentication
- Historical data beyond what `Timeline` returns
- WebSocket / server-sent events (30s poll is sufficient)
- Mobile app or PWA manifest
- Dark/light mode toggle (always dark)
