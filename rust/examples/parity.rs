//! Deterministic output compared with the existing Go implementation in CI.
use std::collections::BTreeSet;
use std::io::{self, Write};

use chrono::{DateTime, NaiveDate, Offset};
use chrono_tz::Tz;
use serde::Deserialize;
use trading_hour::{Error, Market, is_open, market_location, next_close, next_open, timeline};

#[derive(Deserialize)]
struct Input {
    markets: Vec<String>,
    start: String,
    end: String,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let input: Input = serde_json::from_str(include_str!("../../testdata/parity.json"))?;
    let start = NaiveDate::parse_from_str(&input.start, "%Y-%m-%d")?;
    let end = NaiveDate::parse_from_str(&input.end, "%Y-%m-%d")?;
    let mut out = io::BufWriter::new(io::stdout().lock());
    assert_eq!(input.markets.len(), Market::ALL.len());
    for name in input.markets {
        let market: Market = name.parse()?;
        writeln!(out, "market|{market}|{}", market_location(market)?)?;
        let mut day = start;
        while day <= end {
            let schedule = timeline(day, market)?;
            writeln!(
                out,
                "day|{market}|{day}|{}|{}|{}|{}",
                schedule.date.timestamp(),
                schedule.is_holiday,
                schedule.is_half_day,
                schedule.holiday_name
            )?;
            let mut points = BTreeSet::new();
            for hour in (0..24).step_by(3) {
                points.insert(day.and_hms_opt(hour, 0, 0).unwrap().and_utc().timestamp());
            }
            for phase in schedule.phases {
                writeln!(
                    out,
                    "phase|{}|{}|{}|{}|{}",
                    phase.session,
                    phase.start.timestamp(),
                    phase.end.timestamp(),
                    phase.start.offset().fix().local_minus_utc(),
                    phase.end.offset().fix().local_minus_utc()
                )?;
                for edge in [phase.start.timestamp(), phase.end.timestamp()] {
                    for delta in [-1, 0, 1] {
                        points.insert(edge + delta);
                    }
                }
            }
            for point in points {
                let status = is_open(point, market)?;
                writeln!(
                    out,
                    "at|{point}|{}|{}|{}|{}",
                    status.open,
                    status.session,
                    boundary(next_open(point, market))?,
                    boundary(next_close(point, market))?
                )?;
            }
            day = day.succ_opt().unwrap();
        }
    }
    out.flush()?;
    Ok(())
}

fn boundary(value: Result<DateTime<Tz>, Error>) -> Result<String, Error> {
    match value {
        Ok(at) => Ok(at.timestamp().to_string()),
        Err(Error::NoOpenFound) => Ok("none".to_owned()),
        Err(error) => Err(error),
    }
}
