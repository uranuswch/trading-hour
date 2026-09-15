use std::collections::BTreeMap;
use std::sync::OnceLock;

use chrono::NaiveDate;
use serde::Deserialize;

use crate::calendar::{Calendar, CompiledPhase, Holiday, TimeOfDay, WeeklySchedule};
use crate::{Error, Market, Session};

include!(concat!(env!("OUT_DIR"), "/calendars.rs"));

#[derive(Deserialize)]
struct MarketYaml {
    market: String,
    timezone: String,
    weekly_schedule: BTreeMap<String, Vec<PhaseYaml>>,
    #[serde(default)]
    weekly_schedule_overrides: Vec<OverrideYaml>,
    #[serde(default)]
    half_day_schedule: Vec<PhaseYaml>,
}

#[derive(Deserialize)]
struct OverrideYaml {
    effective_from: String,
    weekly_schedule: BTreeMap<String, Vec<PhaseYaml>>,
}

#[derive(Deserialize)]
struct PhaseYaml {
    session: Session,
    start: String,
    end: String,
}

#[derive(Deserialize)]
struct HolidaysYaml {
    market: String,
    year: i32,
    holidays: Vec<HolidayYaml>,
}

#[derive(Deserialize)]
struct HolidayYaml {
    date: String,
    name: String,
    #[serde(rename = "type")]
    kind: HolidayKind,
}

#[derive(Deserialize)]
#[serde(rename_all = "snake_case")]
enum HolidayKind {
    Closed,
    HalfDay,
}

pub(crate) fn lookup(market: Market) -> Result<&'static Calendar, Error> {
    static REGISTRY: OnceLock<Result<BTreeMap<Market, Calendar>, Error>> = OnceLock::new();
    REGISTRY
        .get_or_init(load)
        .as_ref()
        .map_err(Clone::clone)?
        .get(&market)
        .ok_or_else(|| Error::UnknownMarket(market.to_string()))
}

fn load() -> Result<BTreeMap<Market, Calendar>, Error> {
    let mut calendars = BTreeMap::new();
    for (path, source) in FILES.iter().filter(|(p, _)| p.starts_with("data/markets/")) {
        let calendar =
            parse_market(source).map_err(|e| Error::InvalidData(format!("{path}: {e}")))?;
        if calendars.insert(calendar.market, calendar).is_some() {
            return Err(Error::InvalidData(format!("duplicate market in {path}")));
        }
    }
    for market in Market::ALL {
        if !calendars.contains_key(&market) {
            return Err(Error::InvalidData(format!("missing market {market}")));
        }
    }
    for (path, source) in FILES
        .iter()
        .filter(|(p, _)| p.starts_with("data/holidays/"))
    {
        let file: HolidaysYaml =
            serde_yaml::from_str(source).map_err(|e| Error::InvalidData(format!("{path}: {e}")))?;
        let market: Market = file.market.parse()?;
        let calendar = calendars
            .get_mut(&market)
            .ok_or_else(|| Error::UnknownMarket(file.market.clone()))?;
        let directory = match market {
            Market::ChinaAShare => "china-ashare".to_owned(),
            _ => market.as_str().to_ascii_lowercase(),
        };
        if !path.starts_with(&format!("data/holidays/{directory}/")) {
            return Err(Error::InvalidData(format!("{path}: market mismatch")));
        }
        calendar.holiday_years.push(file.year);
        for holiday in file.holidays {
            let half_day = matches!(holiday.kind, HolidayKind::HalfDay);
            if half_day && calendar.half_day.is_empty() {
                return Err(Error::InvalidData(format!(
                    "{path}: half-day without a half-day schedule"
                )));
            }
            calendar.holidays.insert(
                parse_date(&holiday.date)?,
                Holiday {
                    name: holiday.name,
                    half_day,
                },
            );
        }
    }
    for calendar in calendars.values_mut() {
        calendar.holiday_years.sort_unstable();
        calendar.holiday_years.dedup();
    }
    Ok(calendars)
}

fn parse_date(value: &str) -> Result<NaiveDate, Error> {
    NaiveDate::parse_from_str(value, "%Y-%m-%d")
        .map_err(|e| Error::InvalidData(format!("date {value:?}: {e}")))
}

fn parse_market(source: &str) -> Result<Calendar, Error> {
    let file: MarketYaml =
        serde_yaml::from_str(source).map_err(|e| Error::InvalidData(e.to_string()))?;
    let mut overrides = BTreeMap::new();
    for row in file.weekly_schedule_overrides {
        let date = parse_date(&row.effective_from)?;
        if overrides
            .insert(date, compile_week(row.weekly_schedule)?)
            .is_some()
        {
            return Err(Error::InvalidData(format!(
                "duplicate effective_from {date}"
            )));
        }
    }
    Ok(Calendar {
        market: file.market.parse()?,
        timezone: file
            .timezone
            .parse()
            .map_err(|_| Error::InvalidData(format!("timezone {:?}", file.timezone)))?,
        weekly: compile_week(file.weekly_schedule)?,
        overrides,
        half_day: compile_phases(file.half_day_schedule)?,
        holidays: BTreeMap::new(),
        holiday_years: Vec::new(),
    })
}

fn compile_week(source: BTreeMap<String, Vec<PhaseYaml>>) -> Result<WeeklySchedule, Error> {
    let mut weekly = std::array::from_fn(|_| Vec::new());
    for (day, phases) in source {
        let index = [
            "sunday",
            "monday",
            "tuesday",
            "wednesday",
            "thursday",
            "friday",
            "saturday",
        ]
        .iter()
        .position(|name| day.eq_ignore_ascii_case(name))
        .ok_or_else(|| Error::InvalidData(format!("weekday {day:?}")))?;
        weekly[index] = compile_phases(phases)?;
    }
    Ok(weekly)
}

fn compile_phases(source: Vec<PhaseYaml>) -> Result<Vec<CompiledPhase>, Error> {
    source
        .into_iter()
        .map(|p| {
            Ok(CompiledPhase {
                session: p.session,
                start: TimeOfDay::parse(&p.start)?,
                end: TimeOfDay::parse(&p.end)?,
            })
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn malformed_schedules_fail_closed() {
        for bad in ["25:00", "09:60", "9:00", "09:00+2", "é8:00"] {
            assert!(TimeOfDay::parse(bad).is_err(), "{bad}");
        }
        let schedule = "market: KRX\ntimezone: Asia/Seoul\nweekly_schedule: {}\n";
        assert!(parse_market(&schedule.replace("Asia/Seoul", "Bad/Zone")).is_err());
        assert!(parse_market(&schedule.replace("{}", "{funday: []}")).is_err());
        assert!(parse_market(&format!("{schedule}weekly_schedule_overrides:\n  - effective_from: 2026-09-14\n    weekly_schedule: {{}}\n  - effective_from: 2026-09-14\n    weekly_schedule: {{}}\n")).is_err());
        assert!(
            parse_market(&format!(
                "{schedule}weekly_schedule_overrides:\n  - effective_from: 2026-09-14\n"
            ))
            .is_err()
        );
    }
}
