package tradinghour

// HolidayType distinguishes full-closure days from early-close days.
type HolidayType string

const (
	HolidayClosed  HolidayType = "closed"
	HolidayHalfDay HolidayType = "half_day"
)

// civilDate is a timezone-agnostic Y/M/D used as a map key for holiday lookup.
type civilDate struct {
	year  int
	month int // 1..12
	day   int // 1..31
}

// holidayEntry describes a single holiday date.
type holidayEntry struct {
	Name string
	Type HolidayType
}
