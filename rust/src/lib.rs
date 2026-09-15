//! Global market calendars shared with the Go `tradinghour` package.
//!
//! The repository's YAML files are embedded at compile time and parsed once.
//! Queries perform no network or filesystem I/O. Times returned by this crate
//! retain the market's local timezone. Starts are inclusive; ends are exclusive.
//!
//! ```
//! use trading_hour::{is_open, timeline, Market, Session};
//! let status = is_open(1789455600, Market::Krx)?; // 2026-09-15 16:00 KST
//! assert!(status.open);
//! assert_eq!(status.session, Session::PostMarket);
//! let date = chrono::NaiveDate::from_ymd_opt(2026, 9, 15).unwrap();
//! assert_eq!(timeline(date, Market::Krx)?.phases.len(), 3);
//! # Ok::<(), trading_hour::Error>(())
//! ```

mod calendar;
mod loader;

use std::fmt;
use std::str::FromStr;

use chrono::{DateTime, NaiveDate, TimeZone, Utc};
use chrono_tz::Tz;

/// A supported market, matching the Go `MarketType` identifiers.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash)]
pub enum Market {
    Nasdaq,
    Hkex,
    ChinaAShare,
    Tse,
    Twse,
    Krx,
    Fx,
    Cme,
    Ice,
    FxcmUkOil,
    FxcmUsOil,
    Rates,
    Metals,
}

impl Market {
    pub const ALL: [Self; 13] = [
        Self::Nasdaq,
        Self::Hkex,
        Self::ChinaAShare,
        Self::Tse,
        Self::Twse,
        Self::Krx,
        Self::Fx,
        Self::Cme,
        Self::Ice,
        Self::FxcmUkOil,
        Self::FxcmUsOil,
        Self::Rates,
        Self::Metals,
    ];

    pub const fn as_str(self) -> &'static str {
        match self {
            Self::Nasdaq => "NASDAQ",
            Self::Hkex => "HKEX",
            Self::ChinaAShare => "ChinaAShare",
            Self::Tse => "TSE",
            Self::Twse => "TWSE",
            Self::Krx => "KRX",
            Self::Fx => "FX",
            Self::Cme => "CME",
            Self::Ice => "ICE",
            Self::FxcmUkOil => "FXCMUKOil",
            Self::FxcmUsOil => "FXCMUSOil",
            Self::Rates => "Rates",
            Self::Metals => "Metals",
        }
    }
}

impl fmt::Display for Market {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(self.as_str())
    }
}

impl FromStr for Market {
    type Err = Error;
    fn from_str(value: &str) -> Result<Self, Self::Err> {
        Self::ALL
            .into_iter()
            .find(|market| market.as_str() == value)
            .ok_or_else(|| Error::UnknownMarket(value.to_owned()))
    }
}

/// The market phase, with the same spellings as the shared YAML and Go API.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, serde::Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Session {
    Closed,
    PreMarket,
    Regular,
    PostMarket,
    Overnight,
    Continuous,
}

impl Session {
    pub const fn as_str(self) -> &'static str {
        match self {
            Self::Closed => "closed",
            Self::PreMarket => "premarket",
            Self::Regular => "regular",
            Self::PostMarket => "postmarket",
            Self::Overnight => "overnight",
            Self::Continuous => "continuous",
        }
    }
}

impl fmt::Display for Session {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(self.as_str())
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Status {
    pub open: bool,
    pub session: Session,
    pub market: Market,
}

/// An open interval. An overnight end may fall on the following local date.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Phase {
    pub session: Session,
    pub start: DateTime<Tz>,
    pub end: DateTime<Tz>,
}

/// Phases starting on one market-local date, including holiday metadata.
/// A closed holiday can still have a NASDAQ overnight phase for the next day.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct DaySchedule {
    pub date: DateTime<Tz>,
    pub market: Market,
    pub phases: Vec<Phase>,
    pub is_holiday: bool,
    pub is_half_day: bool,
    pub holiday_name: String,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Error {
    UnknownMarket(String),
    InvalidDateTime,
    InvalidData(String),
    NoOpenFound,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::UnknownMarket(market) => write!(f, "trading-hour: unknown market {market:?}"),
            Self::InvalidDateTime => f.write_str(
                "trading-hour: date/time is out of range or not unique in the market timezone",
            ),
            Self::InvalidData(detail) => write!(f, "trading-hour: invalid calendar data: {detail}"),
            Self::NoOpenFound => {
                f.write_str("trading-hour: no open phase found within search horizon")
            }
        }
    }
}
impl std::error::Error for Error {}

/// Query an absolute Unix timestamp, in seconds, including overnight spillover.
pub fn is_open(unix_sec: i64, market: Market) -> Result<Status, Error> {
    let calendar = loader::lookup(market)?;
    let now = instant(unix_sec, calendar.timezone)?;
    let date = now.date_naive();
    for day in [date, date.pred_opt().ok_or(Error::InvalidDateTime)?] {
        for phase in calendar.materialize(day)?.phases {
            if phase.start <= now && now < phase.end {
                return Ok(Status {
                    open: true,
                    session: phase.session,
                    market,
                });
            }
        }
    }
    Ok(Status {
        open: false,
        session: Session::Closed,
        market,
    })
}

/// Query a civil date in the market's timezone, not a UTC date or instant.
pub fn timeline(date: NaiveDate, market: Market) -> Result<DaySchedule, Error> {
    loader::lookup(market)?.materialize(date)
}

/// Return the next strictly future phase start within 15 calendar days.
/// This includes adjacent session boundaries, matching Go `NextOpen`.
pub fn next_open(unix_sec: i64, market: Market) -> Result<DateTime<Tz>, Error> {
    next_boundary(unix_sec, market, false)
}

/// Return the current phase's end, or the next phase's end when closed.
pub fn next_close(unix_sec: i64, market: Market) -> Result<DateTime<Tz>, Error> {
    next_boundary(unix_sec, market, true)
}

fn next_boundary(unix_sec: i64, market: Market, close: bool) -> Result<DateTime<Tz>, Error> {
    let calendar = loader::lookup(market)?;
    let now = instant(unix_sec, calendar.timezone)?;
    for offset in if close { -1..15 } else { 0..15 } {
        let date = now
            .date_naive()
            .checked_add_signed(chrono::Duration::days(offset))
            .ok_or(Error::InvalidDateTime)?;
        for phase in calendar.materialize(date)?.phases {
            let boundary = if close { phase.end } else { phase.start };
            if boundary > now {
                return Ok(boundary);
            }
        }
    }
    Err(Error::NoOpenFound)
}

pub fn market_location(market: Market) -> Result<Tz, Error> {
    Ok(loader::lookup(market)?.timezone)
}

/// Years with an embedded holiday file, including intentionally empty calendars.
/// Outside these years weekly schedules still apply, matching Go. Consumers can
/// use this metadata to report missing calendar coverage explicitly.
pub fn holiday_years(market: Market) -> Result<&'static [i32], Error> {
    Ok(&loader::lookup(market)?.holiday_years)
}

fn instant(unix_sec: i64, timezone: Tz) -> Result<DateTime<Tz>, Error> {
    Utc.timestamp_opt(unix_sec, 0)
        .single()
        .map(|time| time.with_timezone(&timezone))
        .ok_or(Error::InvalidDateTime)
}
