package tradinghour

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed data/markets/*.yaml data/holidays/*/*.yaml
var dataFS embed.FS

type scheduleFile struct {
	Market          string                 `yaml:"market"`
	Timezone        string                 `yaml:"timezone"`
	WeeklySchedule  map[string][]phaseYAML `yaml:"weekly_schedule"`
	HalfDaySchedule []phaseYAML            `yaml:"half_day_schedule"`
}

type phaseYAML struct {
	Session string `yaml:"session"`
	Start   string `yaml:"start"`
	End     string `yaml:"end"`
}

type holidayFile struct {
	Market   string        `yaml:"market"`
	Year     int           `yaml:"year"`
	Holidays []holidayYAML `yaml:"holidays"`
}

type holidayYAML struct {
	Date string `yaml:"date"`
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

var weekdayByName = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
	"wednesday": time.Wednesday, "thursday": time.Thursday,
	"friday": time.Friday, "saturday": time.Saturday,
}

func init() {
	entries, err := fs.ReadDir(dataFS, "data/markets")
	if err != nil {
		panic(fmt.Errorf("tradinghour: read markets dir: %w", err))
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		if err := loadMarket("data/markets/" + e.Name()); err != nil {
			panic(err)
		}
	}
	if err := loadHolidays(); err != nil {
		panic(err)
	}
}

func loadMarket(p string) error {
	raw, err := dataFS.ReadFile(p)
	if err != nil {
		return fmt.Errorf("tradinghour: read %s: %w", p, err)
	}
	var sf scheduleFile
	if err := yaml.Unmarshal(raw, &sf); err != nil {
		return fmt.Errorf("tradinghour: parse %s: %w", p, err)
	}
	loc, err := time.LoadLocation(sf.Timezone)
	if err != nil {
		return fmt.Errorf("tradinghour: load tz %q: %w", sf.Timezone, err)
	}
	m := &Market{
		Type:     MarketType(sf.Market),
		Location: loc,
		Holidays: map[civilDate]holidayEntry{},
	}
	for dayName, phases := range sf.WeeklySchedule {
		wd, ok := weekdayByName[strings.ToLower(dayName)]
		if !ok {
			return fmt.Errorf("tradinghour: unknown weekday %q in %s", dayName, p)
		}
		cps, err := compilePhases(phases, p)
		if err != nil {
			return err
		}
		m.WeeklyPhases[int(wd)] = cps
	}
	if len(sf.HalfDaySchedule) > 0 {
		cps, err := compilePhases(sf.HalfDaySchedule, p)
		if err != nil {
			return err
		}
		m.HalfDayPhases = cps
	}
	registry[m.Type] = m
	return nil
}

func compilePhases(ps []phaseYAML, srcPath string) ([]compiledPhase, error) {
	out := make([]compiledPhase, 0, len(ps))
	for _, p := range ps {
		start, err := parseTimeOfDay(p.Start)
		if err != nil {
			return nil, fmt.Errorf("tradinghour: %s: start: %w", srcPath, err)
		}
		end, err := parseTimeOfDay(p.End)
		if err != nil {
			return nil, fmt.Errorf("tradinghour: %s: end: %w", srcPath, err)
		}
		out = append(out, compiledPhase{
			Session: Session(p.Session),
			Start:   start,
			End:     end,
		})
	}
	return out, nil
}

func loadHolidays() error {
	for mtype, mkt := range registry {
		dir := "data/holidays/" + marketDataDir(mtype)
		entries, err := fs.ReadDir(dataFS, dir)
		if err != nil {
			// Market with no holiday files is allowed (empty calendar).
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			p := path.Join(dir, e.Name())
			if err := loadHolidayFile(mkt, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadHolidayFile(mkt *Market, p string) error {
	raw, err := dataFS.ReadFile(p)
	if err != nil {
		return fmt.Errorf("tradinghour: read %s: %w", p, err)
	}
	var hf holidayFile
	if err := yaml.Unmarshal(raw, &hf); err != nil {
		return fmt.Errorf("tradinghour: parse %s: %w", p, err)
	}
	if MarketType(hf.Market) != mkt.Type {
		return fmt.Errorf("tradinghour: %s: market mismatch (%q vs %q)", p, hf.Market, mkt.Type)
	}
	for _, h := range hf.Holidays {
		t, err := time.Parse("2006-01-02", h.Date)
		if err != nil {
			return fmt.Errorf("tradinghour: %s: bad date %q: %w", p, h.Date, err)
		}
		cd := civilDate{year: t.Year(), month: int(t.Month()), day: t.Day()}
		mkt.Holidays[cd] = holidayEntry{
			Name: h.Name,
			Type: HolidayType(h.Type),
		}
	}
	return nil
}

// marketDataDir maps MarketType to the directory name under data/holidays/.
func marketDataDir(m MarketType) string {
	switch m {
	case MarketNASDAQ:
		return "nasdaq"
	case MarketHKEX:
		return "hkex"
	case MarketChinaAShare:
		return "china-ashare"
	case MarketTSE:
		return "tse"
	case MarketKRX:
		return "krx"
	default:
		return strings.ToLower(string(m))
	}
}
