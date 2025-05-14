package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {

	// Flags for Reference date mode
	refCmd := flag.NewFlagSet("ref", flag.ExitOnError)
	refStart := refCmd.String("s", string(time.Now().Format("2006-01-02")), "Start date")
	refEnd := refCmd.String("e", *refStart, "End date")
	refN := refCmd.Int("n", 0, "Number of repetitions")
	refDays := refCmd.Bool("d", false, "Days")
	refWeeks := refCmd.Bool("w", false, "Weeks")
	refWeekday := refCmd.String("k", "", "Which weekday")
	refInterval := refCmd.Int("i", 1, "Interval per repetition")

	// Flags for Monthly mode
	monCmd := flag.NewFlagSet("monthly", flag.ExitOnError)
	monStart := monCmd.String("s", string(time.Now().Format("2006-01-02")), "Start date")
	monEnd := monCmd.String("e", *monStart, "End date")
	monN := monCmd.Int("n", 0, "Number of repetitions")
	monDay := monCmd.Int("d", 0, "Day of month")
	monWeekday := monCmd.String("k", "", "Weekday")
	monWeekdayN := monCmd.Int("kn", 1, "Weekday number")
	monInterval := monCmd.Int("i", 1, "Number of months per repetition")

	flag.Parse()

	// Check whether Reference date mode or Monthly mode is invoked
	switch os.Args[1] {
	case "ref":

		refCmd.Parse(os.Args[2:])

		// Handle mutually exclusive flags
		endFlags := 0
		if *refEnd != *refStart {
			endFlags++
		}
		if *refN != 0 {
			endFlags++
		}
		if endFlags != 1 {
			fmt.Println("Error: Choose one of -e or -n.")
			return
		}

		var unit string
		unitFlags := 0
		if *refDays != false {
			unitFlags++
			unit = "d"
		}
		if *refWeeks != false {
			unitFlags++
			unit = "w"
		}
		if *refWeekday != "" {
			unitFlags++
			unit = "k"
		}
		if unitFlags != 1 {
			fmt.Println("Error: Choose one of -d, -w, or -k.")
			return
		}

		start, err := time.Parse("2006-01-02", *refStart)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}

		end, err := time.Parse("2006-01-02", *refEnd)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}

		referenceDateMode(start, end, unit, *refWeekday, *refN, *refInterval)

	case "monthly":

		monCmd.Parse(os.Args[2:])

		// Handle mutually exclusive flags
		endFlags := 0
		if *monEnd != *monStart {
			endFlags++
		}
		if *monN != 0 {
			endFlags++
		}
		if endFlags != 1 {
			fmt.Println("Error: Choose one of -e or -n.")
			return
		}

		var unit string
		unitFlags := 0
		if *monDay != 0 {
			unitFlags++
			unit = "d"
		}
		if *monWeekday != "" {
			unitFlags++
			unit = "k"
		}
		if unitFlags != 1 {
			fmt.Println("Error: Choose one of -d or -k.")
			return
		}

		start, err := time.Parse("2006-01-02", *monStart)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}

		end, err := time.Parse("2006-01-02", *monEnd)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}

		monthlyMode(start, end, unit, *monN, *monDay, *monWeekday, *monWeekdayN, *monInterval)
	default:
		fmt.Println("Use subcommands \"ref\" or \"monthly\"")
	}
}

func referenceDateMode(date time.Time, end time.Time, unit string, weekday string, n int, interval int) {

	// Set day value of interval
	days := 0
	direction := ""
	if unit == "d" {
		days = interval
	} else if unit == "w" {
		days = interval * 7
	} else if unit == "k" {
		if interval > 0 && date.Compare(end) <= 0 {
			direction = "next"
		} else {
			direction = "previous"
		}
		days = interval * 7
		date = dateOfWeekday(date, weekday, direction)
	}

	// Go forward or backward from reference date
	if n != 0 {
		for range n {
			// fmt.Println(date.Format("2006-01-02 Mon"))
			fmt.Println(date.Unix())
			date = date.AddDate(0, 0, days)
		}
	} else {
		if date.Compare(end) < 1 {
			for date.Compare(end) < 1 {
				// fmt.Println(date.Format("2006-01-02 Mon"))
				fmt.Println(date.Unix())
				date = date.AddDate(0, 0, days)
			}
		} else if date.Compare(end) >= 0 {
			if interval < 1 {
				fmt.Println("Error: When using -e, -i must be positive.")
				return
			}
			for date.Compare(end) >= 0 {
				// fmt.Println(date.Format("2006-01-02 Mon"))
				fmt.Println(date.Unix())
				date = date.AddDate(0, 0, -days)
			}
		}

	}
}

func monthlyMode(date time.Time, end time.Time, unit string, n int, day int, weekday string, weekdayN int, interval int) {

	if n != 0 {

		// monthly -n -d
		if day > 0 && date.Day() > day {
			// monthly -n -d, d is positive
			date = date.AddDate(0, 1, 0)
		} else if day < 0 && date.Day() > lastDayOfMonth(date).AddDate(0, 0, day).Day() {
			// monthly -n -d, d is negative
			// ISSUE: problems occur if -d is less than -28
			date = date.AddDate(0, 1, 0)
		}
		// monthly -n -k
		// ISSUE: problems occur if -kn is less than -4
		if day == 0 && date.Day() > monthlyDate(date, unit, day, weekday, weekdayN).Day() {
			date = date.AddDate(0, 1, 0)
		}
		// monthly -n
		for range n {
			month := date.Month()
			date = monthlyDate(date, unit, day, weekday, weekdayN)
			if date.Month() == month {
				// fmt.Println(date.Format("2006-01-02 Mon"))
				fmt.Println(date.Unix())
			} else {
				// ISSUE: related to the two above issues
				if day > 0 {
					date = date.AddDate(0, -1, 0)
				} else if day < 0 {
					date = date.AddDate(0, 1, 0)
				}
			}
			for range interval {
				date = lastDayOfMonth(date).AddDate(0, 0, 1)
			}
		}
	} else {
		// monthly -e
		// ISSUE: problems occur if -d is greater than 28 or less than -28
		// ISSUE: problems occur if -kn is greater than 4 or less than -4
		for date.Compare(end) <= 0 {
			date = monthlyDate(date, unit, day, weekday, weekdayN)
			// fmt.Println(date.Format("2006-01-02 Mon"))
			fmt.Println(date.Unix())
			for range interval {
				date = lastDayOfMonth(date).AddDate(0, 0, 1)
			}
		}
	}
}

// Return the date of the next or previous specified weekday
// from the provided date
func dateOfWeekday(date time.Time, weekday string, direction string) time.Time {

	weekdays := map[string]time.Weekday{
		"Sun": time.Sunday,
		"Mon": time.Monday,
		"Tue": time.Tuesday,
		"Wed": time.Wednesday,
		"Thu": time.Thursday,
		"Fri": time.Friday,
		"Sat": time.Saturday,
	}

	if direction == "next" {
		diff := (int(weekdays[weekday]) - int(date.Weekday()) + 7) % 7
		date = date.AddDate(0, 0, diff)
	} else if direction == "previous" {
		diff := (int(date.Weekday()) - int(weekdays[weekday]) + 7) % 7
		date = date.AddDate(0, 0, -diff)
	}

	return date
}

func monthlyDate(date time.Time, unit string, day int, weekday string, weekdayN int) time.Time {

	if unit == "d" {

		if day > 0 {
			date = time.Date(date.Year(), date.Month(), day, 0, 0, 0, 0, time.Local)
		} else if day < 0 {
			date = lastDayOfMonth(date).AddDate(0, 0, day+1)
		}

	} else if unit == "k" {

		direction := ""

		if weekdayN > 0 {
			direction = "next"
			date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
			date = dateOfWeekday(date, weekday, direction).AddDate(0, 0, (weekdayN-1)*7)
		} else {
			direction = "previous"
			date = lastDayOfMonth(date)
			date = dateOfWeekday(date, weekday, direction).AddDate(0, 0, (weekdayN+1)*7)
		}
	}

	return date
}

func lastDayOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
}
