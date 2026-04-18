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
  return r.json();
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
    btn.innerHTML = `<span>${pillDot(cls)}</span>${label}`;
    btn.addEventListener('click', () => openDrawer(s.market));
    container.appendChild(btn);
  }

  document.getElementById('hero-count').textContent = `${openCount} Markets Open`;
}

// ── Drawer (stub — will be replaced in Task 9) ─────────────────
function openDrawer(market) { console.log('openDrawer', market); }
function closeDrawer() {}

// ── Init ──────────────────────────────────────────────────────
fetchStatus().then(renderPills).catch(console.error);
