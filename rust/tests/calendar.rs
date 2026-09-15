use chrono::{DateTime, Datelike, NaiveDate, TimeZone, Timelike, Weekday};
use trading_hour::{
    Error, Market, Session, holiday_years, is_open, next_close, next_open, timeline,
};

fn at(value: &str) -> i64 {
    DateTime::parse_from_rfc3339(value).unwrap().timestamp()
}
fn date(value: &str) -> NaiveDate {
    NaiveDate::parse_from_str(value, "%Y-%m-%d").unwrap()
}

#[test]
fn reviewed_market_boundaries_and_holidays() {
    for (market, when, session) in [
        (
            Market::Nasdaq,
            "2026-03-02T09:29:59-05:00",
            Session::PreMarket,
        ),
        (
            Market::Nasdaq,
            "2026-03-02T09:30:00-05:00",
            Session::Regular,
        ),
        (
            Market::Nasdaq,
            "2026-03-02T16:00:00-05:00",
            Session::PostMarket,
        ),
        (
            Market::Nasdaq,
            "2026-03-02T20:00:00-05:00",
            Session::Overnight,
        ),
        (
            Market::Nasdaq,
            "2026-03-03T03:59:59-05:00",
            Session::Overnight,
        ),
        (
            Market::Nasdaq,
            "2026-03-03T04:00:00-05:00",
            Session::PreMarket,
        ),
        (Market::Nasdaq, "2026-03-06T20:00:00-05:00", Session::Closed),
        (
            Market::Nasdaq,
            "2026-03-08T20:00:00-04:00",
            Session::Overnight,
        ),
        (
            Market::Nasdaq,
            "2026-11-01T20:00:00-05:00",
            Session::Overnight,
        ),
        (Market::Nasdaq, "2026-12-25T12:00:00-05:00", Session::Closed),
        (Market::Hkex, "2026-03-02T12:00:00+08:00", Session::Closed),
        (Market::Hkex, "2026-03-02T13:00:00+08:00", Session::Regular),
        (Market::Hkex, "2026-03-02T16:09:59+08:00", Session::Regular),
        (Market::Hkex, "2026-03-02T16:10:00+08:00", Session::Closed),
        (
            Market::ChinaAShare,
            "2026-03-02T11:30:00+08:00",
            Session::Closed,
        ),
        (
            Market::ChinaAShare,
            "2026-10-01T10:00:00+08:00",
            Session::Closed,
        ),
        (Market::Tse, "2026-03-02T12:30:00+09:00", Session::Regular),
        (Market::Tse, "2026-03-02T15:30:00+09:00", Session::Closed),
        (Market::Twse, "2026-09-15T13:30:00+08:00", Session::Closed),
        (
            Market::Twse,
            "2026-09-15T14:00:00+08:00",
            Session::PostMarket,
        ),
        (Market::Twse, "2026-09-15T14:30:00+08:00", Session::Closed),
        (Market::Twse, "2026-02-12T10:00:00+08:00", Session::Closed),
        (Market::Krx, "2026-09-11T18:00:00+09:00", Session::Closed),
        (
            Market::Krx,
            "2026-09-14T18:00:00+09:00",
            Session::PostMarket,
        ),
        (Market::Krx, "2026-09-15T15:30:00+09:00", Session::Closed),
        (
            Market::Krx,
            "2026-09-15T15:40:00+09:00",
            Session::PostMarket,
        ),
        (
            Market::Krx,
            "2026-09-15T19:59:59+09:00",
            Session::PostMarket,
        ),
        (Market::Krx, "2026-09-15T20:00:00+09:00", Session::Closed),
        (Market::Krx, "2026-09-24T10:00:00+09:00", Session::Closed),
    ] {
        let status = is_open(at(when), market).unwrap();
        assert_eq!(status.session, session, "{market} {when}");
        assert_eq!(status.open, session != Session::Closed);
    }
}

#[test]
fn overnight_follows_the_next_trading_day_even_on_holidays() {
    assert!(
        !is_open(at("2026-01-18T20:00:00-05:00"), Market::Nasdaq)
            .unwrap()
            .open
    );
    let holiday = timeline(date("2026-01-19"), Market::Nasdaq).unwrap();
    assert!(holiday.is_holiday);
    assert_eq!(holiday.phases.len(), 1);
    assert_eq!(holiday.phases[0].session, Session::Overnight);
    assert!(
        is_open(at("2026-01-20T02:00:00-05:00"), Market::Nasdaq)
            .unwrap()
            .open
    );
    assert!(
        !is_open(at("2026-01-19T02:00:00-05:00"), Market::Nasdaq)
            .unwrap()
            .open
    );
}

#[test]
fn timeline_and_next_boundaries_retain_local_time_and_half_days() {
    let half = timeline(date("2026-11-27"), Market::Nasdaq).unwrap();
    assert!(half.is_half_day);
    assert!(!half.is_holiday);
    assert_eq!(half.phases[0].end.hour(), 13);
    assert_eq!(half.phases.len(), 2);
    assert_eq!(
        next_close(at("2026-11-27T12:00:00-05:00"), Market::Nasdaq)
            .unwrap()
            .timestamp(),
        at("2026-11-27T13:00:00-05:00")
    );
    assert_eq!(
        next_open(at("2026-03-07T10:00:00-05:00"), Market::Nasdaq)
            .unwrap()
            .timestamp(),
        at("2026-03-08T20:00:00-04:00")
    );
    assert_eq!(
        next_close(at("2026-03-07T10:00:00-05:00"), Market::Nasdaq)
            .unwrap()
            .timestamp(),
        at("2026-03-09T04:00:00-04:00")
    );
    assert_eq!(
        next_open(at("2026-03-02T11:00:00-05:00"), Market::Nasdaq)
            .unwrap()
            .timestamp(),
        at("2026-03-02T16:00:00-05:00")
    );
    assert_eq!(
        next_open(at("2026-12-25T10:00:00-05:00"), Market::Nasdaq)
            .unwrap()
            .timestamp(),
        at("2026-12-27T20:00:00-05:00")
    );
}

#[test]
fn every_market_is_available_with_explicit_coverage_and_bounded_inputs() {
    for market in Market::ALL {
        assert_eq!(market.as_str().parse::<Market>(), Ok(market));
        assert!(holiday_years(market).unwrap().contains(&2026));
        assert!(timeline(date("2026-03-02"), market).is_ok());
        assert_eq!(is_open(i64::MAX, market), Err(Error::InvalidDateTime));
        assert_eq!(next_open(i64::MIN, market), Err(Error::InvalidDateTime));
    }
    assert!(matches!(
        "missing".parse::<Market>(),
        Err(Error::UnknownMarket(_))
    ));
    // Use the next uncovered year so adding an annual calendar does not break
    // this test of the documented weekly fallback.
    let year = holiday_years(Market::Krx).unwrap().last().unwrap() + 1;
    let mut day = NaiveDate::from_ymd_opt(year, 1, 1).unwrap();
    while day.weekday() != Weekday::Mon {
        day = day.succ_opt().unwrap();
    }
    let instant = chrono_tz::Asia::Seoul
        .from_local_datetime(&day.and_hms_opt(10, 0, 0).unwrap())
        .single()
        .unwrap();
    assert!(is_open(instant.timestamp(), Market::Krx).unwrap().open);
}
