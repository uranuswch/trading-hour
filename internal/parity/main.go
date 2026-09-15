// Command parity emits deterministic public-API results for cross-language CI.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	th "github.com/uranuswch/trading-hour"
)

func main() {
	var input struct {
		Markets []th.MarketType `json:"markets"`
		Start   string          `json:"start"`
		End     string          `json:"end"`
	}
	f, err := os.Open("testdata/parity.json")
	check(err)
	defer f.Close()
	check(json.NewDecoder(f).Decode(&input))
	start, err := time.Parse(time.DateOnly, input.Start)
	check(err)
	end, err := time.Parse(time.DateOnly, input.End)
	check(err)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, market := range input.Markets {
		loc, err := th.MarketLocation(market)
		check(err)
		fmt.Fprintf(w, "market|%s|%s\n", market, loc.String())
		for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
			ds, err := th.Timeline(day, market)
			check(err)
			fmt.Fprintf(w, "day|%s|%s|%d|%t|%t|%s\n", market, day.Format(time.DateOnly), ds.Date.Unix(), ds.IsHoliday, ds.IsHalfDay, ds.HolidayName)
			points := map[int64]bool{}
			// Fixed UTC probes remain independent of either engine's phase output.
			for h := 0; h < 24; h += 3 {
				points[day.Add(time.Duration(h)*time.Hour).Unix()] = true
			}
			for _, p := range ds.Phases {
				_, startOffset := p.Start.Zone()
				_, endOffset := p.End.Zone()
				fmt.Fprintf(w, "phase|%s|%d|%d|%d|%d\n", p.Session, p.Start.Unix(), p.End.Unix(), startOffset, endOffset)
				for _, edge := range []int64{p.Start.Unix(), p.End.Unix()} {
					for _, delta := range []int64{-1, 0, 1} {
						points[edge+delta] = true
					}
				}
			}
			ordered := make([]int64, 0, len(points))
			for point := range points {
				ordered = append(ordered, point)
			}
			slices.Sort(ordered)
			for _, point := range ordered {
				status, err := th.IsOpen(point, market)
				check(err)
				fmt.Fprintf(w, "at|%d|%t|%s|%s|%s\n", point, status.Open, status.Session, boundary(th.NextOpen(point, market)), boundary(th.NextClose(point, market)))
			}
		}
	}
}

func boundary(at time.Time, err error) string {
	if err == th.ErrNoOpenFound {
		return "none"
	}
	check(err)
	return fmt.Sprint(at.Unix())
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
