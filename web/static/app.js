// web/static/app.js
'use strict';

// ── HTML helpers ───────────────────────────────────────────────

/** Escape a string for safe use inside innerHTML. */
function esc(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

const MARKET_LABELS = {
  NASDAQ: 'NASDAQ', HKEX: 'HKEX', ChinaAShare: 'SSE/SZSE', TSE: 'TSE',
  KRX: 'KRX', FX: 'FX', CME: 'CME', ICE: 'ICE',
  FXCMUKOil: 'UK Oil', FXCMUSOil: 'US Oil', Rates: 'Rates', Metals: 'Metals',
};

// ── Fetch helpers ──────────────────────────────────────────────

async function fetchStatus() {
  const r = await fetch('/api/status');
  if (!r.ok) throw new Error('status fetch failed');
  return r.json();
}

async function fetchTimeline(market, date) {
  const url = date ? `/api/timeline/${market}?date=${date}` : `/api/timeline/${market}`;
  const r = await fetch(url);
  if (!r.ok) throw new Error(`timeline fetch failed: ${await r.text()}`);
  return r.json();
}

async function fetchNextOpen(market) {
  const r = await fetch(`/api/nextopen/${market}`);
  if (!r.ok) return null;
  try { return await r.json(); } catch { return null; }
}

// ── Pill helpers ───────────────────────────────────────────────

function pillClass(session) {
  if (session === 'regular' || session === 'continuous') return 'open';
  if (session === 'premarket' || session === 'postmarket' || session === 'overnight') return 'partial';
  return 'closed';
}

function pillDot(cls) {
  if (cls === 'open') return '●';
  if (cls === 'partial') return '◑';
  return '○';
}

// ── Renderers ─────────────────────────────────────────────────

function renderPills(statuses) {
  const container = document.getElementById('pills');
  container.innerHTML = '';
  let openCount = 0;

  for (const s of statuses) {
    const cls = pillClass(s.session);
    if (cls === 'open') openCount++;
    const label = MARKET_LABELS[s.market] || s.market;
    const btn = document.createElement('button');
    btn.className = `pill ${cls}`;
    btn.dataset.market = s.market;
    btn.innerHTML = `<span aria-hidden="true">${pillDot(cls)}</span>${label}`;
    btn.addEventListener('click', () => openDrawer(s.market));
    container.appendChild(btn);
  }

  document.getElementById('hero-count').textContent = `${openCount} Markets Open`;
}

// ── Timezone helpers ───────────────────────────────────────────

function marketHourMin(isoStr, tz) {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: tz, hour: 'numeric', minute: 'numeric', hour12: false,
  }).formatToParts(new Date(isoStr));
  return {
    h: parseInt(parts.find(p => p.type === 'hour').value, 10),
    m: parseInt(parts.find(p => p.type === 'minute').value, 10),
  };
}

function fmtMarketTime(isoStr, tz) {
  return new Date(isoStr).toLocaleTimeString('en-US', {
    timeZone: tz, hour: '2-digit', minute: '2-digit', hour12: false,
  });
}

// ── Timeline bar ───────────────────────────────────────────────

function buildTimelineBar(phases, tz, showCursor) {
  const TOTAL = 24 * 60;
  let segs = '';
  for (const p of phases) {
    const s = marketHourMin(p.start, tz);
    const startM = s.h * 60 + s.m;
    const durationMs = new Date(p.end).getTime() - new Date(p.start).getTime();
    const durationM  = Math.min(durationMs / 60_000, TOTAL - startM); // cap at bar end
    const left  = (startM / TOTAL * 100).toFixed(2);
    const width = (durationM / TOTAL * 100).toFixed(2);
    segs += `<div class="tl-seg ${p.session}" style="left:${left}%;width:${width}%;"></div>`;
  }

  let cursor = '';
  if (showCursor) {
    const { h, m } = marketHourMin(new Date().toISOString(), tz);
    const pct = ((h * 60 + m) / TOTAL * 100).toFixed(2);
    cursor = `<div class="tl-cursor" style="left:${pct}%;"></div>`;
  }

  return `
    <div class="timeline-bar">${segs}${cursor}</div>
    <div class="timeline-labels">
      <span>00:00</span><span>06:00</span><span>12:00</span><span>18:00</span><span>24:00</span>
    </div>`;
}

// ── Session rows ───────────────────────────────────────────────

function buildSessionRows(phases, tz, showActive) {
  if (!phases.length) return '<p class="no-sessions">No sessions today</p>';
  const nowMs = Date.now();
  return phases.map(p => {
    const active = showActive &&
      new Date(p.start).getTime() <= nowMs && nowMs < new Date(p.end).getTime();
    const name = p.session.charAt(0).toUpperCase() + p.session.slice(1);
    const start = fmtMarketTime(p.start, tz);
    const end   = fmtMarketTime(p.end,   tz);
    return `
      <div class="session-row${active ? ' active' : ''}">
        <span class="session-name">${active ? '● ' : ''}${name}</span>
        <span class="session-time">${start} – ${end}</span>
      </div>`;
  }).join('');
}

// ── Drawer ─────────────────────────────────────────────────────

let activeMarket = null;
let marketToday  = null; // YYYY-MM-DD in market's local tz, set when drawer first opens

function drawerBadge(sched) {
  if (sched.isHoliday) return { cls: 'holiday', text: `Holiday — ${esc(sched.holidayName)}` };
  if (sched.isHalfDay) return { cls: 'partial', text: `Half Day — ${esc(sched.holidayName)}` };
  const nowMs = Date.now();
  for (const p of sched.phases || []) {
    if (new Date(p.start).getTime() <= nowMs && nowMs < new Date(p.end).getTime()) {
      const cls = pillClass(p.session);
      const dot = pillDot(cls);
      return { cls, text: `${dot} ${p.session.toUpperCase()}` };
    }
  }
  return { cls: 'closed', text: '○ CLOSED' };
}

function renderDrawerContent(sched) {
  const tz = sched.timezone;
  const isToday = sched.date === marketToday;
  const { cls, text } = drawerBadge(sched);

  document.getElementById('drawer-content').innerHTML = `
    <div class="drawer-mkt-name">${MARKET_LABELS[sched.market] || esc(sched.market)}</div>
    <span class="drawer-badge ${cls}">${text}</span>
    <div class="drawer-tz">${tz}</div>
    <div class="drawer-section-label">DATE</div>
    <input class="drawer-date-input" type="date" id="date-picker" value="${sched.date}" />
    <div class="drawer-section-label">TIMELINE</div>
    ${buildTimelineBar(sched.phases || [], tz, isToday)}
    <div class="drawer-section-label">SESSIONS</div>
    <div class="session-list">${buildSessionRows(sched.phases || [], tz, isToday)}</div>
  `;

  document.getElementById('date-picker').addEventListener('change', async e => {
    const val = e.target.value;
    if (!val) return;
    try {
      const s = await fetchTimeline(activeMarket, val);
      renderDrawerContent(s);
    } catch (err) { console.error(err); }
  });
}

function openDrawer(market) {
  activeMarket = market;
  document.getElementById('overlay').classList.remove('hidden');
  document.getElementById('drawer').classList.add('open');
  document.getElementById('drawer-content').innerHTML =
    '<p style="color:var(--muted);font-size:13px;padding-top:8px;">Loading…</p>';

  fetchTimeline(market).then(sched => {
    if (activeMarket !== market) return; // drawer was closed before response arrived
    marketToday = sched.date;
    renderDrawerContent(sched);
  }).catch(console.error);
}

function closeDrawer() {
  activeMarket = null;
  marketToday  = null;
  document.getElementById('overlay').classList.add('hidden');
  document.getElementById('drawer').classList.remove('open');
  document.getElementById('drawer-content').innerHTML = '';
}

function renderSpotlight(data) {
  const el = document.getElementById('spotlight');
  if (!data) { el.classList.add('hidden'); return; }

  const diffMs = new Date(data.time).getTime() - Date.now();
  if (diffMs <= 0) { el.classList.add('hidden'); return; }
  const h = Math.floor(diffMs / 3_600_000);
  const m = Math.floor((diffMs % 3_600_000) / 60_000);
  const countdown = h > 0 ? `in ${h}h ${m}m` : `in ${m}m`;

  el.classList.remove('hidden');
  el.innerHTML = `
    <div class="spotlight-label">NEXT OPEN</div>
    <div class="spotlight-market">${MARKET_LABELS[data.market] || data.market}</div>
    <div class="spotlight-time">${countdown} · ${data.local}</div>
  `;
}

async function refreshSpotlight(statuses) {
  const closed = statuses.filter(s => pillClass(s.session) === 'closed');
  if (!closed.length) { renderSpotlight(null); return; }
  const results = await Promise.all(closed.map(s => fetchNextOpen(s.market)));
  const valid = results.filter(Boolean);
  if (!valid.length) { renderSpotlight(null); return; }
  const soonest = valid.reduce((a, b) => (a.time < b.time ? a : b));
  renderSpotlight(soonest);
}

// ── Refresh loop ───────────────────────────────────────────────

let initialised = false;

async function refresh() {
  try {
    const statuses = await fetchStatus();
    renderPills(statuses);
    await refreshSpotlight(statuses);
    initialised = true;

    // If drawer is open and showing today, refresh it too
    if (activeMarket && marketToday) {
      const picker = document.getElementById('date-picker');
      const displayedDate = picker ? picker.value : null;
      if (displayedDate !== marketToday) return; // user navigated to a non-today date
      const marketAtLaunch = activeMarket;
      const todayAtLaunch  = marketToday;
      const sched = await fetchTimeline(marketAtLaunch);
      // Re-check: market or date may have changed while fetch was in-flight
      if (activeMarket === marketAtLaunch && sched.date === todayAtLaunch) {
        renderDrawerContent(sched);
      }
    }
  } catch (err) {
    console.error('refresh error:', err);
    if (!initialised) {
      document.getElementById('hero-count').textContent = 'Unable to load';
    }
  }
}

// ── Init ──────────────────────────────────────────────────────
refresh();
setInterval(refresh, 30_000);
