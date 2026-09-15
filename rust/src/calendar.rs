use std::collections::BTreeMap;

use chrono::{DateTime, Datelike, NaiveDate, TimeZone};
use chrono_tz::Tz;

use crate::{DaySchedule, Error, Market, Phase, Session};

pub(crate) type WeeklySchedule = [Vec<CompiledPhase>; 7];

pub(crate) struct Calendar {
    pub market: Market,
    pub timezone: Tz,
    pub weekly: WeeklySchedule,
    pub overrides: BTreeMap<NaiveDate, WeeklySchedule>,
    pub half_day: Vec<CompiledPhase>,
    pub holidays: BTreeMap<NaiveDate, Holiday>,
    pub holiday_years: Vec<i32>,
}

#[derive(Clone)]
pub(crate) struct CompiledPhase {
    pub session: Session,
    pub start: TimeOfDay,
    pub end: TimeOfDay,
}

#[derive(Clone, Copy)]
pub(crate) struct TimeOfDay {
    pub hour: u32,
    pub minute: u32,
    pub next_day: bool,
}

pub(crate) struct Holiday {
    pub name: String,
    pub half_day: bool,
}

impl TimeOfDay {
    pub fn parse(value: &str) -> Result<Self, Error> {
        let (clock, next_day) = value
            .strip_suffix("+1")
            .map_or((value, false), |s| (s, true));
        if !clock.is_ascii() || clock.len() != 5 || &clock[2..3] != ":" {
            return Err(Error::InvalidData(format!("invalid time {value:?}")));
        }
        let parse = |s: &str| {
            s.parse::<u32>()
                .map_err(|_| Error::InvalidData(format!("invalid time {value:?}")))
        };
        let hour = parse(&clock[..2])?;
        let minute = parse(&clock[3..])?;
        if hour > 23 || minute > 59 {
            return Err(Error::InvalidData(format!("invalid time {value:?}")));
        }
        Ok(Self {
            hour,
            minute,
            next_day,
        })
    }

    fn on(self, date: NaiveDate, timezone: Tz) -> Result<DateTime<Tz>, Error> {
        let day = if self.next_day {
            date.succ_opt().ok_or(Error::InvalidDateTime)?
        } else {
            date
        };
        let local = day
            .and_hms_opt(self.hour, self.minute, 0)
            .ok_or(Error::InvalidDateTime)?;
        timezone
            .from_local_datetime(&local)
            .single()
            .ok_or(Error::InvalidDateTime)
    }
}

impl Calendar {
    pub fn materialize(&self, date: NaiveDate) -> Result<DaySchedule, Error> {
        let weekly = self
            .overrides
            .range(..=date)
            .next_back()
            .map_or(&self.weekly, |(_, schedule)| schedule);
        let weekly = &weekly[date.weekday().num_days_from_sunday() as usize];
        let holiday = self.holidays.get(&date);
        let next_day = date.succ_opt().ok_or(Error::InvalidDateTime)?;
        let next_day_closed = self.holidays.get(&next_day).is_some_and(|h| !h.half_day);
        let daytime = match holiday {
            Some(h) if h.half_day => self.half_day.as_slice(),
            Some(_) => &[],
            None => weekly.as_slice(),
        };
        let mut phases = Vec::new();
        // Overnight belongs to its next trading day. Preserve a holiday evening
        // before a normal day, and suppress the evening before a closed holiday.
        for phase in daytime
            .iter()
            .filter(|p| p.session != Session::Overnight)
            .chain(
                weekly
                    .iter()
                    .filter(|p| p.session == Session::Overnight && !next_day_closed),
            )
        {
            phases.push(Phase {
                session: phase.session,
                start: phase.start.on(date, self.timezone)?,
                end: phase.end.on(date, self.timezone)?,
            });
        }
        Ok(DaySchedule {
            date: TimeOfDay {
                hour: 0,
                minute: 0,
                next_day: false,
            }
            .on(date, self.timezone)?,
            market: self.market,
            phases,
            is_holiday: holiday.is_some_and(|h| !h.half_day),
            is_half_day: holiday.is_some_and(|h| h.half_day),
            holiday_name: holiday.map_or_else(String::new, |h| h.name.clone()),
        })
    }
}
