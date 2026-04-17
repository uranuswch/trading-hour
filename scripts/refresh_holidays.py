#!/usr/bin/env python3
"""Generate data/holidays/<market>/<year>.yaml from exchange_calendars.

Usage: python scripts/refresh_holidays.py --year 2027 [--out-dir data/holidays]

For each supported market, emits a YAML file listing closed days and
(where detectable) half-days. Half-days require human verification.
"""

import argparse
import datetime as dt
import os
from pathlib import Path

import exchange_calendars as ec
import yaml

# Map our MarketType -> (ec code, data dir)
MARKETS = {
    "NASDAQ":      ("XNAS", "nasdaq"),
    "HKEX":        ("XHKG", "hkex"),
    "ChinaAShare": ("XSHG", "china-ashare"),
    "TSE":         ("XTKS", "tse"),
    "KRX":         ("XKRX", "krx"),
}

# Known half-day markets and their fallback "early close" detection.
# exchange_calendars exposes .early_closes for NYSE/NASDAQ; for HKEX we flag
# manually below in MANUAL_HALF_DAYS (script reviewer verifies).
MANUAL_HALF_DAYS = {
    "HKEX": {
        # date -> name. Lunar New Year Eve and Christmas Eve are traditional.
        # The reviewer must update this set each year.
    },
}

def holidays_for(market: str, year: int) -> list[dict]:
    ec_code, _ = MARKETS[market]
    cal = ec.get_calendar(ec_code)
    start = dt.date(year, 1, 1)
    end = dt.date(year, 12, 31)
    all_dates = {d.date() for d in cal.schedule.loc[str(start):str(end)].index}
    calendar_days = {start + dt.timedelta(days=i) for i in range((end - start).days + 1)}
    closed_weekdays = {d for d in calendar_days
                       if d.weekday() < 5 and d not in all_dates}

    early = set()
    if hasattr(cal, "early_closes"):
        ec_early = cal.early_closes
        if ec_early is not None:
            early = {ts.date() for ts in ec_early
                     if start <= ts.date() <= end}
    for d in MANUAL_HALF_DAYS.get(market, {}):
        if start <= d <= end:
            early.add(d)

    entries = []
    for d in sorted(closed_weekdays):
        entries.append({
            "date": d.isoformat(),
            "name": holiday_name(cal, d),
            "type": "closed",
        })
    for d in sorted(early):
        # Override a closed entry if this date appeared there (unlikely).
        entries = [e for e in entries if e["date"] != d.isoformat()]
        entries.append({
            "date": d.isoformat(),
            "name": MANUAL_HALF_DAYS.get(market, {}).get(d) or "Early close",
            "type": "half_day",
        })
    entries.sort(key=lambda e: e["date"])
    return entries

def holiday_name(cal, date: dt.date) -> str:
    # exchange_calendars doesn't expose names directly; fall back to generic.
    return "Market holiday"

def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--year", type=int, required=True)
    ap.add_argument("--out-dir", type=Path, default=Path("data/holidays"))
    args = ap.parse_args()

    for market, (_, subdir) in MARKETS.items():
        out = args.out_dir / subdir / f"{args.year}.yaml"
        out.parent.mkdir(parents=True, exist_ok=True)
        payload = {
            "market": market,
            "year": args.year,
            "holidays": holidays_for(market, args.year),
        }
        with out.open("w") as f:
            yaml.safe_dump(payload, f, sort_keys=False, allow_unicode=True)
        print(f"wrote {out} ({len(payload['holidays'])} entries)")

if __name__ == "__main__":
    main()
