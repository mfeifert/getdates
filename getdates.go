package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type dateSeries struct {
	start    time.Time
	end      time.Time
	n        int
	day      int
	weekday  int
	weekdayn int
	months   int
}

// ================================END=OF=MONTH================================

func endOfMonth(d time.Time) time.Time {
	year := d.Year()
	month := d.Month() + 1
	date := time.Date(year, month, 0, 0, 0, 0, 0, time.Local)
	return date
}

// ===============================DATE=OF=WEEKDAY==============================

func dateOfWeekday(date time.Time, weekday int, weekdayn int) time.Time {

	start := int(date.Weekday())

	weekday += 7
	diff := weekday - start
	diff %= 7
	if weekdayn > 0 {
		weekdayn--
	}
	diff += 7 * weekdayn

	date = date.AddDate(0, 0, diff)

	return date
}

// =================================MONTHLY=DATE===============================

func monthlyDate(date time.Time, s dateSeries) time.Time {

	if s.day != 0 {
		if s.day > 0 {
			year := date.Year()
			month := date.Month()
			date = time.Date(year, month, s.day, 0, 0, 0, 0, time.Local)
		} else {
			date = endOfMonth(date).AddDate(0, 0, s.day+1)
		}
	} else {
		if s.weekdayn > 0 {
			year := date.Year()
			month := date.Month()
			date = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		} else {
			date = endOfMonth(date)
		}
		date = dateOfWeekday(date, s.weekday, s.weekdayn)
	}

	return date
}

// ============================REFERENCE=DATE=MODE=============================

func (s dateSeries) referenceDateMode() []time.Time {

	date := s.start
	var dates []time.Time

	if s.n != 0 {
		// -n
		for range s.n {
			dates = append(dates, date)
			date = date.AddDate(0, 0, s.day)
		}
	} else {
		// -e
		for date.Compare(s.end) < 1 {
			dates = append(dates, date)
			date = date.AddDate(0, 0, s.day)
		}
	}

	return dates
}

// ===============================MONTHLY=MODE=================================

func (s dateSeries) monthlyMode() []time.Time {

	date := s.start
	day := s.day
	var dates []time.Time
	var mn int

	if s.n != 0 {
		// -n
		if day > 0 {
			if date.Day() > day {
				// -d positive
				date = date.AddDate(0, 1, 0)
				mn = 1
			}
		} else if day < 0 {
			if date.Day() > endOfMonth(date).AddDate(0, 0, day).Day() {
				// -d negative
				date = date.AddDate(0, 1, 0)
				mn = -1
			}
		} else if date.Day() > monthlyDate(date, s).Day() {
			// -k
			date = date.AddDate(0, 1, 0)
			mn = 1
		}
		for range s.n {
			month := date.Month()
			date = monthlyDate(date, s)

			if date.Month() == month {
				dates = append(dates, date)
			} else {
				date = date.AddDate(0, mn, 0)
			}

			for range s.months {
				date = endOfMonth(date).AddDate(0, 0, 1)
			}
		}
	} else {
		// -e
		for date.Compare(s.end) <= 0 {
			date = monthlyDate(date, s)
			dates = append(dates, date)
			for range s.months {
				date = endOfMonth(date).AddDate(0, 0, 1)
			}
		}
	}

	return dates
}

// ==================================MAIN======================================

func main() {

	// Parse flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	start := flag.String("s", string(time.Now().Format(time.DateOnly)), "Start date")
	end := flag.String("e", *start, "End date")
	n := flag.Int("n", 0, "Number of repetitions")
	day := flag.Int("d", 0, "Days")
	weeks := flag.Int("w", 0, "Weeks")
	weekday := flag.String("k", "", "Weekday")
	weekdayn := flag.Int("kn", 0, "Weekday number")
	months := flag.Int("i", 1, "Months per repetition (monthly mode only)")
	h := flag.Bool("h", false, "Human readable output")
	flag.CommandLine.Parse(os.Args[2:])

	// Parse start and end dates from command line flags
	startTime, _ := time.Parse(time.DateOnly, *start)
	endTime, _ := time.Parse(time.DateOnly, *end)

	// Input validation
	unitFlags := 0
	endFlags := 0
	flag.CommandLine.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "d", "w", "k":
			unitFlags++
		case "e", "n":
			endFlags++
		}
	})
	if unitFlags != 1 {
		fmt.Println("One of -d, -w, or -k must be used.")
		os.Exit(1)
	}
	if endFlags != 1 {
		fmt.Println("One of -e or -n must be used.")
		os.Exit(1)
	}
	if *weeks != 0 {
		*day = *weeks * 7
	}

	*weekday = strings.ToLower(*weekday)
	weekdays := map[string]time.Weekday{
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}

	// Assign data to dateSeries type
	s := dateSeries{
		start:    startTime,
		end:      endTime,
		n:        *n,
		day:      *day,
		weekday:  int(weekdays[*weekday]),
		weekdayn: *weekdayn,
		months:   *months,
	}

	// Select reference date or monthly mode
	mode := os.Args[1]
	var dates []time.Time
	if mode == "r" {
		dates = s.referenceDateMode()
	} else if mode == "m" {
		dates = s.monthlyMode()
	}

	// Select output format
	if *h == true {
		for _, date := range dates {
			fmt.Println(date.Format("2006-01-02 Mon"))
		}
	} else {
		for _, date := range dates {
			fmt.Println(date.Unix())
		}
	}
}
