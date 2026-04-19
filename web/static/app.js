// web/static/app.js
'use strict';

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

// ── Drawer (stub — will be replaced in Task 9) ─────────────────
function openDrawer(market) { console.log('openDrawer', market); }
function closeDrawer() {}

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

// ── Init ──────────────────────────────────────────────────────
async function init() {
  const statuses = await fetchStatus();
  renderPills(statuses);
  await refreshSpotlight(statuses);
}

init().catch(err => {
  console.error(err);
  document.getElementById('hero-count').textContent = 'Unable to load';
});
