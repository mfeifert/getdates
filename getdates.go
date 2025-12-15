package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type dateSeries struct {
	start    time.Time
	end      time.Time
	n        int
	interval int
	unit     string
	mn       int
	weekday  string
}

// ================================END=OF=MONTH================================

func endOfMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month()+1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
}

// ===============================DATE=OF=WEEKDAY==============================

func dateOfWeekday(date time.Time, weekday string, direction int) time.Time {

	weekdays := map[string]time.Weekday{
		"Sun": time.Sunday,
		"Mon": time.Monday,
		"Tue": time.Tuesday,
		"Wed": time.Wednesday,
		"Thu": time.Thursday,
		"Fri": time.Friday,
		"Sat": time.Saturday,
	}

	if direction >= 0 {
		diff := (int(weekdays[weekday]) - int(date.Weekday()) + 7) % 7
		date = date.AddDate(0, 0, diff)
	} else {
		diff := (int(date.Weekday()) - int(weekdays[weekday]) + 7) % 7
		date = date.AddDate(0, 0, -diff)
	}

	return date
}

// =================================MONTHLY=DATE===============================

func monthlyDate(date time.Time, unit string, mn int, weekday string) time.Time {

	if unit == "d" {

		if mn > 0 {
			date = time.Date(date.Year(), date.Month(), mn, 0, 0, 0, 0, time.Local)
		} else {
			date = endOfMonth(date).AddDate(0, 0, mn+1)
		}

	} else if unit == "k" {

		if mn > 0 {
			date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
			date = dateOfWeekday(date, weekday, mn).AddDate(0, 0, (mn-1)*7)
		} else {
			date = endOfMonth(date)
			date = dateOfWeekday(date, weekday, mn).AddDate(0, 0, (mn+1)*7)
		}
	}

	return date
}

// ============================REFERENCE=DATE=MODE=============================

func (s dateSeries) referenceDateMode() []time.Time {

	date := s.start
	var dates []time.Time

	// Adjust interval if unit is weeks
	if s.unit == "w" {
		s.interval = s.interval * 7
	}

	// Use -n or -e flag
	if s.n != 0 {
		// -n
		for range s.n {
			dates = append(dates, date)
			date = date.AddDate(0, 0, s.interval)
		}
	} else if s.end != s.start {
		// -e
		for date.Compare(s.end) < 1 {
			dates = append(dates, date)
			date = date.AddDate(0, 0, s.interval)
		}
	}

	return dates
}

// ===============================MONTHLY=MODE=================================

func (s dateSeries) monthlyMode() []time.Time {

	date := s.start
	var dates []time.Time

	// Use -n or -e flag
	if s.n != 0 {

		// -n -d
		if s.unit == "d" {
			if s.mn > 0 && date.Day() > s.mn {
				// -mn is positive
				date = date.AddDate(0, 1, 0)

			} else if s.mn < 0 && date.Day() > endOfMonth(date).AddDate(0, 0, s.mn).Day() {
				// -mn is negative
				date = date.AddDate(0, 1, 0)
			}
		}

		// -n -k
		if s.unit == "k" {
			if date.Day() > monthlyDate(date, s.unit, s.mn, s.weekday).Day() {
				date = date.AddDate(0, 1, 0)
			}
		}

		// -n
		for range s.n {

			month := date.Month()
			date = monthlyDate(date, s.unit, s.mn, s.weekday)

			if date.Month() == month {
				dates = append(dates, date)
			} else {
				if s.mn > 0 {
					date = date.AddDate(0, -1, 0)
				} else if s.mn < 0 {
					date = date.AddDate(0, 1, 0)
				}
			}

			for range s.interval {
				date = endOfMonth(date).AddDate(0, 0, 1)
			}
		}

	} else if s.end != s.start {

		// -e
		for date.Compare(s.end) <= 0 {
			date = monthlyDate(date, s.unit, s.mn, s.weekday)
			dates = append(dates, date)
			for range s.interval {
				date = endOfMonth(date).AddDate(0, 0, 1)
			}
		}
	}

	return dates
}

// ==================================MAIN======================================

func main() {

	// mode can be "r" or "m"
	mode := os.Args[1]

	// Parse flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	start := flag.String("s", string(time.Now().Format(time.DateOnly)), "Start date")
	end := flag.String("e", *start, "End date")
	n := flag.Int("n", 0, "Number of repetitions")
	interval := flag.Int("i", 1, "Interval per repetition")
	days := flag.Bool("d", false, "Days")
	weeks := flag.Bool("w", false, "Weeks")
	weekday := flag.String("k", "", "Which weekday")
	mn := flag.Int("mn", 1, "Number of day or weekday in monthly mode")
	h := flag.Bool("h", false, "Human readable output")
	flag.CommandLine.Parse(os.Args[2:])

	// Parse start and end dates from command line flags
	startTime, _ := time.Parse(time.DateOnly, *start)
	endTime, _ := time.Parse(time.DateOnly, *end)

	// Select unit (days, weeks, weekday)
	var unit string
	if *days == true {
		unit = "d"
	} else if *weeks == true {
		unit = "w"
	} else if *weekday != "" {
		unit = "k"
	}

	// Store data in dateSeries type
	s := dateSeries{
		start:    startTime,
		end:      endTime,
		n:        *n,
		interval: *interval,
		unit:     unit,
		weekday:  *weekday,
		mn:       *mn,
	}

	// Select reference date or monthly mode, output format
	if mode == "r" {
		if *h == true {
			for _, value := range s.referenceDateMode() {
				fmt.Println(value.Format("2006-01-02 Mon"))
			}
		} else {
			for _, value := range s.referenceDateMode() {
				fmt.Println(value.Unix())
			}
		}
	} else if mode == "m" {
		if *h == true {
			for _, value := range s.monthlyMode() {
				fmt.Println(value.Format("2006-01-02 Mon"))
			}
		} else {
			for _, value := range s.monthlyMode() {
				fmt.Println(value.Unix())
			}
		}
	}
}
